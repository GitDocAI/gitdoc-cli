package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	// Check if the request is multipart/form-data
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		// Parse the multipart form data
		err := r.ParseMultipartForm(10 << 20) // 10 MB limit
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   "Error parsing multipart form",
			})
			return
		}

		// Get the file part
		file, _, err := r.FormFile("binary_content")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   "Error retrieving the file",
			})
			return
		}
		defer file.Close()

		// Get the file path from the form
		filePath := r.FormValue("file_path")
		if filePath == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   "file_path is required in multipart form",
			})
			return
		}

		fullPath := filepath.Join(ws.basePath, filePath)
		fullPath = filepath.Clean(fullPath)

		if !strings.HasPrefix(fullPath, filepath.Clean(ws.basePath)) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   "Wrong filepath",
			})
			return
		}

		// Create the file
		dst, err := os.Create(fullPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   fmt.Sprintf("Error creating file: %v", err),
			})
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the created file
		_, err = io.Copy(dst, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   fmt.Sprintf("Error copying file: %v", err),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ReplaceResponse{
			Success: true,
			Message: fmt.Sprintf("file updated: %s", fullPath),
		})
		return
	} else {
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

		if req.FilePath == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ReplaceResponse{
				Success: false,
				Error:   "file_path is required",
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

		if err := os.WriteFile(fullPath, []byte(req.NewText), 0644); err != nil {
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
}

func (ws *WebhookServer) Start() error {
	if _, err := os.Stat(ws.basePath); os.IsNotExist(err) {
		return fmt.Errorf("base directory doesnt exist: %s", ws.basePath)
	}

	http.HandleFunc("/replace", ws.handleFileReplace)
	return http.ListenAndServe(":"+ws.port, nil)
}
