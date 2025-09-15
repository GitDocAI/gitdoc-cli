package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/YHVCorp/gitdocai-cli/schemas"
	"encoding/json"
)

// CreateFileWithApiReference creates a file (and its directories) at the given path
// and writes the <ApiReference> component inside, using the provided ApiReference data.
func CreateFileWithApiReference(apiRef schemas.OpenapiData, path string) error {
	// Ensure directories exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	var tryitbaseurl string = (func() string{ if apiRef.TryItBaseUrl ==nil{ return ""} else{return *apiRef.TryItBaseUrl}})();
	// Convert parameters and responses to JSON format
		parametersJSON, err := json.Marshal(apiRef.Parameters)
		if err != nil {
			return fmt.Errorf("failed to convert parameters to JSON: %w", err)
		}

		responsesJSON, err := json.Marshal(apiRef.Responses)
		if err != nil {
			return fmt.Errorf("failed to convert responses to JSON: %w", err)
		}

		// Generate the <ApiReference> component as a string
		component := fmt.Sprintf(`<ApiReference
			  title="%s"
			  method="%s"
			  path="%s"
			  parameters='%s'
			  responses='%s'
			  tryItBaseUrl="%s"
			/>`,
			apiRef.Title,
			apiRef.Method,
			apiRef.Path,
			parametersJSON,
			responsesJSON,
		tryitbaseurl,
		)

	// Create and write to the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(component); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

