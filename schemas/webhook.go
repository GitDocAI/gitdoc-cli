package schemas
// Config holds server configuration passed to StartWebhook.
type Config struct {
	Port           		string         `json:"port"`
	DocsPath       		string         `json:"docs_path"`
	InputPath      		string         `json:"input_path"`
	ConfigPath     		string         `json:"config_path"`
	AllowOrigin    		string 		   `json:"allow_origin"`
	RemoteFileSystemUrl string 		   `json:"remote_fielsystem_url"`
	Authentication 		Authentication `json:"authentication"`
}


type Authentication struct{
	OrganizationID string
	SessionToken   string
	ProjectID      string
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
