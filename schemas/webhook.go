package schemas


type ReplaceRequest struct {
	FilePath   string `json:"file_path"`
	OldSegment string `json:"old_segment"`
	NewText    string `json:"new_text"`
}

type ReplaceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
