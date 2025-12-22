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
		cfg.AllowOrigin = "*"
	}

	absInputPath, err := filepath.Abs(cfg.InputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve input path: %w", err)
	}
	absConfigPath, err := filepath.Abs(cfg.ConfigPath)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve config path: %w", err)
	}
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
	mux.HandleFunc("/api/content/api/v1/filesystem/entry",corsMiddleware(saveDocumentHandler))
	mux.HandleFunc("/api/content/api/v1/filesystem/file",corsMiddleware(getDocumentHandler))
	mux.HandleFunc("/api/content/api/v1/filesystem/move",corsMiddleware(moveDocumentHandler))
	mux.HandleFunc("/api/content/api/v1/filesystem/download",corsMiddleware(downLoadHandler))
	mux.HandleFunc("/api/content/api/v1/filesystem/upload",corsMiddleware(upLoadHandler))
	mux.HandleFunc("/api/theme",corsMiddleware(themeHandler))

	// SPA handler: serve index.html for all routes except static files
	mux.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Build the full file path
		path := filepath.Join(webhookConfig.BuildPath, r.URL.Path)

		// Check if the path is a file that exists
		fileInfo, err := os.Stat(path)
		if err == nil && !fileInfo.IsDir() {
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}


//==================== Generals================
type EntryType string
const (
	EntryTypeFile EntryType ="file"
	EntryTypeFolder EntryType ="folder"
)


// Helper: JSON response
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}


//==================== Get Document Controller============

type GetDocumentBody struct{
	Path string `json:"path"`
}


func getDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method!= http.MethodPost{
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}
    body,err:=io.ReadAll(r.Body)
	if err!=nil{
		http.Error(w, "Body is required", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var entry_body GetDocumentBody

	err = json.Unmarshal(body,&entry_body)
	if err!=nil{
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}


	docPath := filepath.Join(webhookConfig.DocsPath,entry_body.Path )

	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	file,err:=os.Open(docPath)
	if err!=nil{
		http.Error(w, "could not open file", http.StatusInternalServerError)
		return
	}

	content,err:= io.ReadAll(file)
	defer file.Close()

	response :=map[string]any{
		"content":string(content),
    }

	jsonResponse(w, http.StatusOK,response)
}

//==================== Save Document Controller============

type SaveDocumentBody struct{
	Path string `json:"path"`
	Type EntryType `json:"type"`
	Content string `json:"content"`
}

// Save document
func saveDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var entry_body SaveDocumentBody

	err = json.Unmarshal(body,&entry_body)
	if err!=nil{
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}

	// Build path to source file (InputPath)
	mdxPath := filepath.Join(webhookConfig.InputPath, entry_body.Path)
	mdxDir := filepath.Dir(mdxPath)

	// Ensure directory exists
	if err := os.MkdirAll(mdxDir, 0755); err != nil {
		log.Printf("Error creating directory %s: %v", mdxDir, err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	if entry_body.Type == EntryTypeFile{
		// Write the MDX file to InputPath
		if err := os.WriteFile(mdxPath, []byte(entry_body.Content), 0644); err != nil {
			log.Printf("Error writing MDX file %s: %v", mdxPath, err)
			http.Error(w, "Failed to save document", http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Document %s saved successfully", entry_body.Path),
		"docId":   entry_body.Path,
	}
	jsonResponse(w, http.StatusOK, response)
}

//===================Move Document=============================
type MoveDocumentBody struct{
	From string `json:"from"`
	To string `json:"to"`
}

func moveDocumentHandler(w http.ResponseWriter, r *http.Request){
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var move_body MoveDocumentBody

	err = json.Unmarshal(body,&move_body)
	if err!=nil{
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}


	fromPath := filepath.Join(webhookConfig.InputPath, move_body.From)
	toPath := filepath.Join(webhookConfig.InputPath, move_body.To)

	fromFile, err := os.Open(fromPath)
    if err != nil {
		http.Error(w, "Could not open file", http.StatusInternalServerError)
		return
    }
    defer fromFile.Close()

    outputFile, err := os.Create(toPath)
    if err != nil {
		http.Error(w, "Could not create new file", http.StatusInternalServerError)
		return
    }
    defer outputFile.Close()

    _, err = io.Copy(outputFile, fromFile)
    if err != nil {
		http.Error(w, "Couldn't copy to dest from source:", http.StatusInternalServerError)
		return
    }

    err = os.Remove(fromPath)
    if err != nil {
		http.Error(w, "Couldn't remove source file after movement", http.StatusInternalServerError)
		return
    }
	w.WriteHeader(http.StatusOK)
}
//========================Download Handler=========================

type DownloadBody struct{
	Path string `json:"path"`
}


func downLoadHandler(w http.ResponseWriter, r *http.Request){

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var move_body DownloadBody

	err = json.Unmarshal(body,&move_body)
	if err!=nil{
		http.Error(w, "wrong body", http.StatusBadRequest)
		return
	}

	file := filepath.Join(webhookConfig.InputPath, move_body.Path)
	http.ServeFile(w,r,file)

}

//========================Upload Handler=========================

func upLoadHandler(w http.ResponseWriter, r *http.Request){
	r.ParseMultipartForm(32 << 26)
	file,handle,err:= r.FormFile("file")
	if err!=nil{
		http.Error(w, "wrong file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	assets_folder := filepath.Join(webhookConfig.InputPath,"assets")
	if err := os.MkdirAll(assets_folder, 0755); err != nil {
		http.Error(w, "Failed to create assets directory", http.StatusInternalServerError)
		return
	}

	asset_path := filepath.Join(assets_folder,handle.Filename)
	destiny,err:=os.Create(asset_path)
	if err!=nil{
		http.Error(w, "Failed to create asset", http.StatusInternalServerError)
		return
	}
	defer destiny.Close()

	_,err=io.Copy(destiny,file)
	if err!=nil{
		http.Error(w, "Failed to copy asset", http.StatusInternalServerError)
		return
	}

	jsonResponse(w,http.StatusOK,map[string]any{
		"file_path":asset_path,
	})
}
//===========================Theme Handler ===============

func themeHandler(w http.ResponseWriter, r *http.Request){
	file,err:=os.Open(webhookConfig.ConfigPath)
	if err!=nil{
		http.Error(w, "Failed to open config file", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	file_content,err:= io.ReadAll(file)
	if err!=nil{
		http.Error(w, "Failed to read config file", http.StatusInternalServerError)
		return
	}

	var config schemas.SiteConfig
	err= json.Unmarshal(file_content,&config)
	if err!=nil{
		http.Error(w, "wrong configuration", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text")
	io.WriteString(w,config.Theme)
}
