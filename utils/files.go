package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GetMyPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exPath := filepath.Dir(ex)
	return exPath
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

func CopyFile(src, dst string, mode os.FileMode) error {
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

func CopyDir(src string, dst string) error {
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
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath, info.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}
