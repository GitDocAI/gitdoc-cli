package utils
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)



func FindOpenAPIJSONInDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("could'nt read dir: '%s': %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".json") {
			full := filepath.Join(dir, e.Name())
			ok, err := isOpenAPISpec(full)
			if err != nil {
				continue
			}
			if ok {
				return full, nil
			}
		}
	}
	return "", nil
}

func FindOpenAPIAsDictInCurrentDir(filePath string) map[string]interface{} {
	content, err := loadJSONFileToMap(filePath)
	if err != nil {
		return nil
	}
	return content
}


func isOpenAPISpec(path string) (bool, error) {
	loader := &openapi3.Loader{Context: context.Background()}
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return false, fmt.Errorf("not a valid OpenAPI spec: %w", err)
	}

	// Verify the presence of essential OpenAPI fields
	if doc.OpenAPI == "" || doc.Info == nil || doc.Paths == nil {
		return false, fmt.Errorf("file is missing essential OpenAPI fields")
	}

	return true, nil
}


func loadJSONFileToMap(path string) (map[string]interface{}, error) {
	loader := &openapi3.Loader{Context: context.Background()}
	openapiDoc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI file: %w", err)
	}
	openapiMap, err := json.Marshal(openapiDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert OpenAPI document to map: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(openapiMap, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAPI document: %w", err)
	}
	return result, nil
}
