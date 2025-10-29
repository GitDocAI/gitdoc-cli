package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GitDocAI/inkflow"
)

func GetMyPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func CreatePathIfNotExist(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return false, fmt.Errorf("error creating path: %v", err)
		}
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking path: %v", err)
	}
	return true, nil
}

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadJson(fileName string, data interface{}) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, data)
	if err != nil {
		return err
	}

	return nil
}

func ClearFolderExcept(dir string, filters []string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		name := entry.Name()

		if matchesAnyFilter(name, filters) {
			continue
		}

		fullPath := filepath.Join(dir, name)

		if entry.IsDir() {
			err = os.RemoveAll(fullPath)
		} else {
			err = os.Remove(fullPath)
		}

		if err != nil {
			return fmt.Errorf("failed to delete %s: %w", fullPath, err)
		}
	}

	return nil
}

func matchesAnyFilter(name string, filters []string) bool {
	for _, filter := range filters {
		if strings.HasPrefix(filter, "startswith:") {
			value := strings.TrimPrefix(filter, "startswith:")
			if strings.HasPrefix(name, value) {
				return true
			}
		} else if strings.HasPrefix(filter, "contains:") {
			value := strings.TrimPrefix(filter, "contains:")
			if strings.Contains(name, value) {
				return true
			}
		} else if strings.HasPrefix(filter, "equals:") {
			value := strings.TrimPrefix(filter, "equals:")
			if name == value {
				return true
			}
		}
	}
	return false
}

func ShouldExclude(name string, excludeFilters []string) bool {
	return matchesAnyFilter(name, excludeFilters)
}

func CopyFile(src, dst string, mode os.FileMode,with_parsing bool) error {

	if with_parsing && (strings.HasSuffix(src,".mdx") || strings.HasSuffix(src,".md")){
		mdxToJsonParser := inkflow.NewParser(nil)
		 content,err:= os.ReadFile(src)
		if err!=nil{
			return err
		}

		result, err := mdxToJsonParser.Parse(string(content))
		if err != nil {
			return fmt.Errorf("error parsing MDX content from '%s': %v", src, err)
		}

		// Marshal the parsed content to JSON
		jsonData, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("error marshaling JSON from '%s': %v", src, err)
		}

		// Change the file extension to .json
		destPath := strings.Replace(dst, filepath.Ext(dst), ".json", 1)

		// Write the JSON data to the destination file
		err = os.WriteFile(destPath, jsonData, mode)
		if err != nil {
			return fmt.Errorf("error writing JSON file '%s': %v", destPath, err)
		}

		return nil
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func CopyDir(src string, dst string,with_parsing bool) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath,with_parsing); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath, info.Mode(),with_parsing); err != nil {
				return err
			}
		}
	}

	return nil
}
