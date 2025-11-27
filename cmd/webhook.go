package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GitDocAI/gitdocai-cli/schemas"
)

// Config holds server configuration passed to StartWebhook.
type Config = schemas.Config

// Request payloads
type SaveContentRequest = schemas.SaveContentRequest
type UploadFileRequest = schemas.UploadFileRequest
type RenameFileRequest = schemas.RenameFileRequest

var webhookConfig Config

// StartWebhook initializes the server using the provided config, finds an available port
// starting from cfg.Port (or 8080 if empty) and starts the HTTP server in a goroutine.
// It returns the actual port that the server is listening on, or an error.
func StartWebhook(cfg Config) (int, error) {
	// Set defaults if necessary
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.AllowOrigin == "" {
		cfg.AllowOrigin = "http://localhost:8080"
	}

	// Resolve absolute paths for docs, input, and config
	absDocsPath, err := filepath.Abs(cfg.DocsPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve docs path: %w", err)
	}
	absInputPath, err := filepath.Abs(cfg.InputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve input path: %w", err)
	}
	absConfigPath, err := filepath.Abs(cfg.ConfigPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve config path: %w", err)
	}
	cfg.DocsPath = absDocsPath
	cfg.InputPath = absInputPath
	cfg.ConfigPath = absConfigPath

	if _, err := os.Stat(cfg.DocsPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("docs path does not exist: %s", cfg.DocsPath)
	}
	if _, err := os.Stat(cfg.InputPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("input path does not exist: %s", cfg.InputPath)
	}

	webhookConfig = cfg

	// Build the HTTP handlers (same handlers rely on webhookConfig)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/docs/content/*", corsMiddleware(corsMiddleware(docHandler)))

	// SPA handler: serve index.html for all routes except static files
	mux.HandleFunc("/", corsMiddleware(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Ignore api endpoints
		if strings.HasPrefix(r.URL.Path, "/api/docs/content") {
			docHandler(w, r)
			return
		}

		// Build the full file path
		path := filepath.Join(webhookConfig.BuildPath, r.URL.Path)

		// Check if the path is a file that exists
		fileInfo, err := os.Stat(path)
		if err == nil && !fileInfo.IsDir() {
			// File exists, serve it
			http.ServeFile(w, r, path)
			return
		}

		// If it's a directory or file doesn't exist, serve index.html (SPA fallback)
		indexPath := filepath.Join(webhookConfig.BuildPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
		} else {
			http.Error(w, "Site not built yet. Run 'npm run build' first.", http.StatusNotFound)
		}
	})))

	// Port scanning: try ports starting at cfg.Port and increment until free
	startPort, err := parsePort(cfg.Port)
	if err != nil {
		return 0, fmt.Errorf("invalid start port: %w", err)
	}

	const maxAttempts = 1000
	var listener net.Listener
	var actualPort int
	for i := 0; i < maxAttempts; i++ {
		tryPort := startPort + i
		addr := fmt.Sprintf(":%d", tryPort)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			// port busy/failed, continue to next
			continue
		}
		// success: we have reserved the port
		listener = l
		actualPort = tryPort
		break
	}

	if listener == nil {
		return 0, fmt.Errorf("no available port found starting from %d after %d attempts", startPort, maxAttempts)
	}

	// Start server in a goroutine so StartWebhook returns immediately.
	server := &http.Server{
		Handler: mux,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed: %v", err)
		}
	}()

	return actualPort, nil
}

// parsePort converts a string port to int and validates range.
func parsePort(p string) (int, error) {
	var port int
	_, err := fmt.Sscanf(p, "%d", &port)
	if err != nil {
		return 0, err
	}
	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("port out of range")
	}
	return port, nil
}

// Authentication middleware - only validates in production mode
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth in dev and preview modes
		if webhookConfig.Mode != "prod" {
			next(w, r)
			return
		}

		// Only validate auth in production mode
		orgId := r.Header.Get("X-Organization-ID")
		sessionToken := r.Header.Get("X-Session-Token")
		projectId := r.Header.Get("X-Project-ID")

		// Check if all auth values are provided and match
		if webhookConfig.Authentication.OrganizationID != "" &&
			webhookConfig.Authentication.ProjectID != "" &&
			webhookConfig.Authentication.SessionToken != "" &&
			orgId == webhookConfig.Authentication.OrganizationID &&
			projectId == webhookConfig.Authentication.ProjectID &&
			sessionToken == webhookConfig.Authentication.SessionToken {
			next(w, r)
			return
		}

		http.Error(w, "Access denied", http.StatusUnauthorized)
	}
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", webhookConfig.AllowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Document handler (GET and PUT)
func docHandler(w http.ResponseWriter, r *http.Request) {
	// Extract docId from path: /api/docs/content/{docId}
	docId := strings.TrimPrefix(r.URL.Path, "/api/docs/content")
	if docId == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	// Sanitize docId to prevent directory traversal
	if strings.Contains(docId, "..") {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	// Clean the path to remove any double slashes or other issues
	docId = filepath.Clean(docId)

	switch r.Method {
	case http.MethodGet:
		getDocumentHandler(w, r, docId)
	case http.MethodPut:
		saveDocumentHandler(w, r, docId)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getDocumentHandler(w http.ResponseWriter, r *http.Request, docId string) {
	docPath := filepath.Join(webhookConfig.DocsPath, docId)

	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, docPath)
}

// Save document
func saveDocumentHandler(w http.ResponseWriter, r *http.Request, docId string) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Build path to source file (InputPath)
	mdxPath := filepath.Join(webhookConfig.InputPath, docId)
	mdxDir := filepath.Dir(mdxPath)

	// Ensure directory exists
	if err := os.MkdirAll(mdxDir, 0755); err != nil {
		log.Printf("Error creating directory %s: %v", mdxDir, err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Write the MDX file to InputPath
	if err := os.WriteFile(mdxPath, []byte(body), 0644); err != nil {
		log.Printf("Error writing MDX file %s: %v", mdxPath, err)
		http.Error(w, "Failed to save document", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Document %s saved successfully", docId),
		"docId":   docId,
	}
	jsonResponse(w, http.StatusOK, response)
}

// Helper: JSON response
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
