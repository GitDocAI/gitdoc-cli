package schemas
// Config holds server configuration passed to StartWebhook.
type Config struct {
	Port        string `json:"port"`        // Starting port (e.g. "8080"). If empty, defaults to "8080".
	DocsPath    string `json:"docs_path"`   // Path where JSON docs are stored
	ConfigPath  string `json:"config_path"` // Path to the server config file
	AllowOrigin string `json:"allow_origin"`
}

// Request payloads
type SaveContentRequest struct {
	Content string `json:"content"`
}

type UploadFileRequest struct {
	FilePath string `json:"file_path"` // Relative path where to save the file
	FileData string `json:"file_data"` // Base64 encoded file data
}

type RenameFileRequest struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}
