package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"os"
	"github.com/YHVCorp/gitdocai-cli/schemas"
)

type ReplaceRequest = schemas.ReplaceRequest
type ReplaceResponse = schemas.ReplaceResponse

type WebhookServer struct {
	basePath string
	port     string
}

func NewWebhookServer(basePath string, port string) *WebhookServer {
	return &WebhookServer{
		basePath: basePath,
		port:     port,
	}
}

func (ws *WebhookServer) handleFileReplace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   "Only POST allowed",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   "wrong request",
		})
		return
	}
	defer r.Body.Close()

	var req ReplaceRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   "invalid JSON",
		})
		return
	}

	if req.FilePath == "" || req.OldSegment == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   "file_path and old_segment are required",
		})
		return
	}

	fullPath := filepath.Join(ws.basePath, req.FilePath)

	fullPath = filepath.Clean(fullPath)

	if !strings.HasPrefix(fullPath, filepath.Clean(ws.basePath)) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   "Wrong filepath",
		})
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   fmt.Sprintf("error reading file: %v", err),
		})
		return
	}

	fileContent := string(content)
	if !strings.Contains(fileContent, req.OldSegment) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   "Pattern not found",
		})
		return
	}

	newContent := strings.Replace(fileContent, req.OldSegment, req.NewText, -1)

	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: false,
			Error:   fmt.Sprintf("Error writting to file: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ReplaceResponse{
		Success: true,
		Message: fmt.Sprintf("file updated: %s", fullPath),
	})
}

func (ws *WebhookServer) Start() error {
	if _, err := os.Stat(ws.basePath); os.IsNotExist(err) {
		return fmt.Errorf("base directory doesnt exist: %s", ws.basePath)
	}

	http.HandleFunc("/replace", ws.handleFileReplace)
	return http.ListenAndServe(":"+ws.port, nil)
}
