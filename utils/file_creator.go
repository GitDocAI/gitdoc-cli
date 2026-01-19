package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GitDocAI/gitdocai-cli/schemas"
)


func CreateFileWithApiReference(apiRef schemas.OpenapiData, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	jsonData, err := json.MarshalIndent(apiRef, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if filepath.Ext(path) == "" {
		path = path + ".json"
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

