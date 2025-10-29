package schemas

// HTTP request/response types
type ListDirectoryRequest struct {
	Path       string   `json:"path"`
	Branch     string   `json:"branch"`
	Extensions []string `json:"extensions"`
}

type GetFileRequest struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

type UpsertEntryRequest struct {
	Path    string `json:"path"`
	Type    string `json:"type"` // "file" or "folder"
	Content string `json:"content"`
}

type MoveEntryHTTPRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"`   // "file" or "folder"
	Action string `json:"action"` // "move" or "copy"
}

type DeleteEntryRequest struct {
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

type FileEntry struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

type DirectoryListResponse struct {
	Entries []FileEntry `json:"entries"`
}

type GenericHTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
