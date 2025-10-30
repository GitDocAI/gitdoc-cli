package cmd

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GitDocAI/gitdocai-cli/utils"
)

// FlusherConfig holds configuration for the flusher service
type FlusherConfig struct {
	OrganizationID   string
	SessionToken     string
	ProjectID        string
	FilesystemAPIURL string
	FlushInterval    time.Duration
	LocalDocsPath    string
}

// FileState represents the state of a file for change detection
type FileState struct {
	Path         string
	Hash         string
	LastModified time.Time
}

// API Request/Response types
type listDirectoryRequest struct {
	Path       string   `json:"path"`
	Branch     string   `json:"branch"`
	Extensions []string `json:"extensions"`
}

type fileEntry struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

type directoryListResponse struct {
	Entries []fileEntry `json:"entries"`
}

type upsertEntryRequest struct {
	Path    string `json:"path"`
	Type    string `json:"type"` // "FILE" or "FOLDER"
	Content string `json:"content"`
}

type deleteEntryRequest struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

type genericHTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Flusher manages file synchronization with remote filesystem
type Flusher struct {
	config     FlusherConfig
	fileStates map[string]FileState // path -> FileState
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// NewFlusher creates a new Flusher instance
func NewFlusher() (*Flusher, error) {
	// Load environment variables
	orgID := os.Getenv("ORGANIZATION_ID")
	sessionToken := os.Getenv("SESSION_TOKEN")
	projectID := os.Getenv("PROJECT_ID")
	filesystemURL := os.Getenv("FILESYSTEM_API_URL")
	flushIntervalStr := os.Getenv("FLUSH_INTERVAL")

	// Validate required env vars
	if orgID == "" || sessionToken == "" || projectID == "" || filesystemURL == "" {
		return nil, fmt.Errorf("missing required environment variables: ORGANIZATION_ID, SESSION_TOKEN, PROJECT_ID, FILESYSTEM_API_URL")
	}

	// Parse flush interval (default to 5 minutes)
	flushInterval := 5 * time.Minute
	if flushIntervalStr != "" {
		duration, err := time.ParseDuration(flushIntervalStr)
		if err == nil {
			flushInterval = duration
		}
	}

	// Get local docs path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	localDocsPath := filepath.Join(homeDir, ".gitdocai", "docs")

	return &Flusher{
		config: FlusherConfig{
			OrganizationID:   orgID,
			SessionToken:     sessionToken,
			ProjectID:        projectID,
			FilesystemAPIURL: strings.TrimSuffix(filesystemURL, "/"),
			FlushInterval:    flushInterval,
			LocalDocsPath:    localDocsPath,
		},
		fileStates: make(map[string]FileState),
		stopChan:   make(chan struct{}),
	}, nil
}

// Start begins the flusher service
func (f *Flusher) Start() error {
	utils.PrintStep("Starting flusher service...")

	// Create local docs directory if it doesn't exist
	if err := os.MkdirAll(f.config.LocalDocsPath, 0755); err != nil {
		return fmt.Errorf("failed to create local docs path: %v", err)
	}

	// Initial download of all files from remote filesystem
	if err := f.downloadAllFiles(); err != nil {
		return fmt.Errorf("failed to download initial files: %v", err)
	}

	// Build initial file state
	if err := f.buildFileState(); err != nil {
		return fmt.Errorf("failed to build initial file state: %v", err)
	}

	utils.PrintSuccess("Flusher service started")
	utils.PrintInfo(fmt.Sprintf("Flush interval: %v", f.config.FlushInterval))
	utils.PrintInfo(fmt.Sprintf("Local docs path: %s", f.config.LocalDocsPath))

	// Start periodic flush in background
	go f.periodicFlush()

	return nil
}

// Stop stops the flusher service
func (f *Flusher) Stop() {
	close(f.stopChan)
}

// periodicFlush runs periodic synchronization
func (f *Flusher) periodicFlush() {
	ticker := time.NewTicker(f.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := f.flush(); err != nil {
				utils.PrintError(fmt.Sprintf("Flush error: %v", err))
			}
		case <-f.stopChan:
			return
		}
	}
}

// flush performs a synchronization of files
func (f *Flusher) flush() error {
	utils.PrintStep("Flushing changes to remote filesystem...")

	changes, err := f.detectChanges()
	if err != nil {
		return fmt.Errorf("failed to detect changes: %v", err)
	}

	if len(changes.New) == 0 && len(changes.Modified) == 0 && len(changes.Deleted) == 0 {
		utils.PrintInfo("No changes detected")
		return nil
	}

	// Upload new files
	for _, path := range changes.New {
		if err := f.uploadFile(path); err != nil {
			utils.PrintError(fmt.Sprintf("Failed to upload new file %s: %v", path, err))
		} else {
			utils.PrintInfo(fmt.Sprintf("Uploaded new file: %s", path))
		}
	}

	// Upload modified files
	for _, path := range changes.Modified {
		if err := f.uploadFile(path); err != nil {
			utils.PrintError(fmt.Sprintf("Failed to upload modified file %s: %v", path, err))
		} else {
			utils.PrintInfo(fmt.Sprintf("Uploaded modified file: %s", path))
		}
	}

	// Delete removed files
	for _, path := range changes.Deleted {
		if err := f.deleteFile(path); err != nil {
			utils.PrintError(fmt.Sprintf("Failed to delete file %s: %v", path, err))
		} else {
			utils.PrintInfo(fmt.Sprintf("Deleted file: %s", path))
		}
	}

	// Update file state after successful flush
	if err := f.buildFileState(); err != nil {
		return fmt.Errorf("failed to rebuild file state: %v", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Flush completed: %d new, %d modified, %d deleted",
		len(changes.New), len(changes.Modified), len(changes.Deleted)))

	return nil
}

// FileChanges represents detected changes
type FileChanges struct {
	New      []string
	Modified []string
	Deleted  []string
}

// detectChanges detects changes in the local filesystem
func (f *Flusher) detectChanges() (*FileChanges, error) {
	f.mu.RLock()
	oldStates := make(map[string]FileState)
	for k, v := range f.fileStates {
		oldStates[k] = v
	}
	f.mu.RUnlock()

	// Get current file states
	currentStates := make(map[string]FileState)
	err := filepath.Walk(f.config.LocalDocsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(f.config.LocalDocsPath, path)
		if err != nil {
			return err
		}

		hash, err := f.hashFile(path)
		if err != nil {
			return err
		}

		currentStates[relPath] = FileState{
			Path:         relPath,
			Hash:         hash,
			LastModified: info.ModTime(),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	changes := &FileChanges{
		New:      []string{},
		Modified: []string{},
		Deleted:  []string{},
	}

	// Detect new and modified files
	for path, currentState := range currentStates {
		oldState, exists := oldStates[path]
		if !exists {
			changes.New = append(changes.New, path)
		} else if oldState.Hash != currentState.Hash {
			changes.Modified = append(changes.Modified, path)
		}
	}

	// Detect deleted files
	for path := range oldStates {
		if _, exists := currentStates[path]; !exists {
			changes.Deleted = append(changes.Deleted, path)
		}
	}

	return changes, nil
}

// buildFileState builds the current file state map
func (f *Flusher) buildFileState() error {
	newStates := make(map[string]FileState)

	err := filepath.Walk(f.config.LocalDocsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(f.config.LocalDocsPath, path)
		if err != nil {
			return err
		}

		hash, err := f.hashFile(path)
		if err != nil {
			return err
		}

		newStates[relPath] = FileState{
			Path:         relPath,
			Hash:         hash,
			LastModified: info.ModTime(),
		}

		return nil
	})

	if err != nil {
		return err
	}

	f.mu.Lock()
	f.fileStates = newStates
	f.mu.Unlock()

	return nil
}

// hashFile calculates MD5 hash of a file
func (f *Flusher) hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// downloadAllFiles downloads all files from remote filesystem
func (f *Flusher) downloadAllFiles() error {
	utils.PrintStep("Downloading files from remote filesystem...")

	// List all files from root directory
	reqBody := listDirectoryRequest{
		Path:       "/",
		Branch:     "main",
		Extensions: []string{}, // Get all file types
	}

	var listResp directoryListResponse
	if err := f.doRequest("POST", "/list", reqBody, &listResp); err != nil {
		return fmt.Errorf("failed to list remote files: %v", err)
	}

	// Download and save each file
	for _, entry := range listResp.Entries {
		if entry.Type == "FILE" {
			destPath := filepath.Join(f.config.LocalDocsPath, entry.Path)

			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %v", entry.Path, err)
			}

			// Write file content
			if err := os.WriteFile(destPath, []byte(entry.Content), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %v", entry.Path, err)
			}
		}
	}

	utils.PrintSuccess(fmt.Sprintf("Downloaded %d files", len(listResp.Entries)))
	return nil
}

// uploadFile uploads a file to remote filesystem
func (f *Flusher) uploadFile(relativePath string) error {
	fullPath := filepath.Join(f.config.LocalDocsPath, relativePath)

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Prepare request
	reqBody := upsertEntryRequest{
		Path:    relativePath,
		Type:    "FILE",
		Content: string(content),
	}

	var resp genericHTTPResponse
	if err := f.doRequest("POST", "/entry", reqBody, &resp); err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("upload failed: %s", resp.Message)
	}

	return nil
}

// deleteFile deletes a file from remote filesystem
func (f *Flusher) deleteFile(relativePath string) error {
	reqBody := deleteEntryRequest{
		Path:   relativePath,
		Branch: "main",
	}

	var resp genericHTTPResponse
	if err := f.doRequest("DELETE", "/entry", reqBody, &resp); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("delete failed: %s", resp.Message)
	}

	return nil
}

// doRequest performs an HTTP request to the filesystem API
func (f *Flusher) doRequest(method, endpoint string, reqBody interface{}, respBody interface{}) error {
	// Marshal request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create request
	url := f.config.FilesystemAPIURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Organization-ID", f.config.OrganizationID)
	req.Header.Set("X-Session-Token", f.config.SessionToken)
	req.Header.Set("X-Project-ID", f.config.ProjectID)

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Decode response
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}
	}

	return nil
}

// StartFlusher creates and starts the flusher service
func StartFlusher() (*Flusher, error) {
	flusher, err := NewFlusher()
	if err != nil {
		return nil, err
	}

	if err := flusher.Start(); err != nil {
		return nil, err
	}

	return flusher, nil
}
