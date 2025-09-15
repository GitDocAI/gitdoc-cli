package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/YHVCorp/gitdocai-cli/schemas"
)

// CreateFileWithApiReference creates a file (and its directories) at the given path
// and writes the <ApiReference> component inside, using the provided ApiReference data.
func CreateFileWithApiReference(apiRef schemas.ApiReference, path string) error {
	// Ensure directories exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	var tryitbaseurl string = (func() string{ if apiRef.TryItBaseUrl ==nil{ return ""} else{return *apiRef.TryItBaseUrl}})();
	// Generate the <ApiReference> component as a string
	component := fmt.Sprintf(`<ApiReference
		  title="%s"
		  method="%s"
		  path="%s"
		  parameters={%v}
		  responses={%v}
		  tryItBaseUrl="%s"
		/>`,
		apiRef.Title,
		apiRef.Method,
		apiRef.Path,
		apiRef.Parameters,
		apiRef.Responses,
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

