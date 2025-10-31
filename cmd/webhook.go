package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GitDocAI/gitdocai-cli/inkflow"
	"github.com/GitDocAI/gitdocai-cli/schemas"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

// Config holds server configuration passed to StartWebhook.
type Config = schemas.Config

// Request payloads
type SaveContentRequest = schemas.SaveContentRequest
type UploadFileRequest = schemas.UploadFileRequest
type RenameFileRequest = schemas.RenameFileRequest

var webhookConfig Config
var requestClient *utils.RemoteClient

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

	if webhookConfig.RemoteFileSystemUrl != "" {
		requestClient = utils.NewRemoteClient(webhookConfig.RemoteFileSystemUrl, cfg.Authentication)
	}

	// Build the HTTP handlers (same handlers rely on webhookConfig)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/docs/list", corsMiddleware(getDocListHandler))
	mux.HandleFunc("/api/docs/", corsMiddleware(docHandler))
	mux.HandleFunc("/api/config", corsMiddleware(configHandler))
	mux.HandleFunc("/api/files/upload", corsMiddleware(uploadFileHandler))
	mux.HandleFunc("/api/files/rename", corsMiddleware(renameFileHandler))
	mux.HandleFunc("/api/health", corsMiddleware(healthHandler))

	// SPA handler: serve index.html for all routes except static files
	mux.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Ignore api endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Build the full file path
		var path string
		// Serve gitdocai.config.json from ConfigPath (where it's actually saved)
		if r.URL.Path == "/gitdocai.config.json" {
			path = webhookConfig.ConfigPath
		} else {
			path = filepath.Join(webhookConfig.BuildPath, r.URL.Path)
		}

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
	}))

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

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", webhookConfig.AllowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
		"config": map[string]string{
			"docs_path":   webhookConfig.DocsPath,
			"config_path": webhookConfig.ConfigPath,
		},
	}
	jsonResponse(w, http.StatusOK, response)
}

// Get documentation list handler
func getDocListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if requestClient != nil {
		requestClient.Get("")
		return
	}

	indexPath := filepath.Join(webhookConfig.DocsPath, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		log.Printf("Error reading index.json: %v", err)
		http.Error(w, "Failed to read documentation list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Document handler (GET and PUT)
func docHandler(w http.ResponseWriter, r *http.Request) {
	// Extract docId from path: /api/docs/{docId}
	docId := strings.TrimPrefix(r.URL.Path, "/api/docs/")
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

// Get single document
func getDocumentHandler(w http.ResponseWriter, r *http.Request, docId string) {
	// Add .json extension if not present
	if !strings.HasSuffix(docId, ".json") {
		docId = docId + ".json"
	}
	docPath := filepath.Join(webhookConfig.DocsPath, docId)

	data, err := os.ReadFile(docPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Document not found", http.StatusNotFound)
		} else {
			log.Printf("Error reading document %s: %v", docId, err)
			http.Error(w, "Failed to read document", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
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

	var req SaveContentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	// Validate that content is valid JSON
	var contentData interface{}
	if err := json.Unmarshal([]byte(req.Content), &contentData); err != nil {
		http.Error(w, "Content must be valid JSON", http.StatusBadRequest)
		return
	}

	// Convert JSON to MDX using inkflow
	mdxContent, err := inkflow.RenderJSONToMDX(req.Content, nil)
	if err != nil {
		log.Printf("❌ Error converting JSON to MDX: %v", err)
		log.Printf("📄 Full JSON that failed: %s", req.Content)
		http.Error(w, fmt.Sprintf("Failed to convert JSON to MDX: %v", err), http.StatusInternalServerError)
		return
	}

	// Remove .json extension and add .mdx if needed
	docId = strings.TrimSuffix(docId, ".json")
	if !strings.HasSuffix(docId, ".mdx") {
		docId = docId + ".mdx"
	}

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
	if err := os.WriteFile(mdxPath, []byte(mdxContent), 0644); err != nil {
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

// Config handler (GET and PUT)
func configHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getConfigHandler(w, r)
	case http.MethodPut:
		saveConfigHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get configuration
func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(webhookConfig.ConfigPath)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		http.Error(w, "Failed to read configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Save configuration
func saveConfigHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate JSON
	var configData interface{}
	if err := json.Unmarshal(body, &configData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Pretty print
	prettyJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		http.Error(w, "Failed to format JSON", http.StatusInternalServerError)
		return
	}

	// Determine where to save based on mode
	if webhookConfig.Mode == "preview" {
		// In preview mode (Cloud), save to BOTH locations:
		// 1. InputPath - so flusher can sync it to filesystem API
		inputConfigPath := filepath.Join(webhookConfig.InputPath, "gitdocai.config.json")
		inputConfigDir := filepath.Dir(inputConfigPath)
		if err := os.MkdirAll(inputConfigDir, 0755); err != nil {
			log.Printf("Error creating config directory %s: %v", inputConfigDir, err)
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(inputConfigPath, prettyJSON, 0644); err != nil {
			log.Printf("Error writing config to %s: %v", inputConfigPath, err)
			http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Saved configuration to InputPath: %s", inputConfigPath)

		// 2. OutputPath/public - so GET requests immediately see the new config
		outputConfigPath := webhookConfig.ConfigPath
		outputConfigDir := filepath.Dir(outputConfigPath)
		if err := os.MkdirAll(outputConfigDir, 0755); err != nil {
			log.Printf("Error creating config directory %s: %v", outputConfigDir, err)
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(outputConfigPath, prettyJSON, 0644); err != nil {
			log.Printf("Error writing config to %s: %v", outputConfigPath, err)
			http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Saved configuration to OutputPath: %s", outputConfigPath)
	} else {
		// In dev mode, save only to ConfigPath (output/public)
		savePath := webhookConfig.ConfigPath
		configDir := filepath.Dir(savePath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Printf("Error creating config directory %s: %v", configDir, err)
			http.Error(w, "Failed to create directory", http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(savePath, prettyJSON, 0644); err != nil {
			log.Printf("Error writing config to %s: %v", savePath, err)
			http.Error(w, "Failed to save configuration", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Successfully saved configuration to %s", savePath)
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Configuration saved successfully",
	}
	jsonResponse(w, http.StatusOK, response)
}

// Upload file handler
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req UploadFileRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FilePath == "" {
		http.Error(w, "file_path is required", http.StatusBadRequest)
		return
	}
	if req.FileData == "" {
		http.Error(w, "file_data is required", http.StatusBadRequest)
		return
	}

	// Sanitize file path to prevent directory traversal
	if strings.Contains(req.FilePath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Clean the path
	cleanPath := filepath.Clean(req.FilePath)

	// Decode base64 data
	fileBytes, err := base64.StdEncoding.DecodeString(req.FileData)
	if err != nil {
		http.Error(w, "Invalid base64 data", http.StatusBadRequest)
		return
	}

	// Build full file path
	fullPath := filepath.Join(webhookConfig.DocsPath, cleanPath)

	// Ensure directory exists
	fileDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		log.Printf("Error creating directory %s: %v", fileDir, err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Write file
	if err := os.WriteFile(fullPath, fileBytes, 0644); err != nil {
		log.Printf("Error writing file %s: %v", fullPath, err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Successfully uploaded file: %s", cleanPath)

	response := map[string]interface{}{
		"success":   true,
		"message":   "File uploaded successfully",
		"file_path": cleanPath,
	}
	jsonResponse(w, http.StatusOK, response)
}

// Rename file handler
func renameFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req RenameFileRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.OldPath == "" {
		http.Error(w, "old_path is required", http.StatusBadRequest)
		return
	}
	if req.NewPath == "" {
		http.Error(w, "new_path is required", http.StatusBadRequest)
		return
	}

	// Sanitize paths to prevent directory traversal
	if strings.Contains(req.OldPath, "..") || strings.Contains(req.NewPath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Clean the paths
	oldPath := filepath.Clean(req.OldPath)
	newPath := filepath.Clean(req.NewPath)

	// Add .json extension if not present (convert .mdx to .json if necessary)
	if !strings.HasSuffix(oldPath, ".json") {
		oldPath = strings.TrimSuffix(oldPath, ".mdx") + ".json"
	}
	if !strings.HasSuffix(newPath, ".json") {
		newPath = strings.TrimSuffix(newPath, ".mdx") + ".json"
	}

	// Build full file paths
	fullOldPath := filepath.Join(webhookConfig.DocsPath, oldPath)
	fullNewPath := filepath.Join(webhookConfig.DocsPath, newPath)

	// Check if old file exists
	if _, err := os.Stat(fullOldPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("File not found: %s", oldPath), http.StatusNotFound)
		return
	}

	// Check if new file already exists
	if _, err := os.Stat(fullNewPath); err == nil {
		http.Error(w, fmt.Sprintf("File already exists: %s", newPath), http.StatusConflict)
		return
	}

	// Ensure new directory exists
	newDir := filepath.Dir(fullNewPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		log.Printf("Error creating directory %s: %v", newDir, err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Rename/move the file
	if err := os.Rename(fullOldPath, fullNewPath); err != nil {
		log.Printf("Error renaming file from %s to %s: %v", fullOldPath, fullNewPath, err)
		http.Error(w, "Failed to rename file", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "File renamed successfully",
		"old_path": oldPath,
		"new_path": newPath,
	}
	jsonResponse(w, http.StatusOK, response)
}

// Helper: JSON response
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper: Copy file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
