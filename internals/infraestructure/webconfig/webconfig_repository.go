package webconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	config "github.com/YHVCorp/DockitAi/cli/internals/domain/models"
)

type WebConfigRepository struct {
	webPath     string
	client_path string
}

func NewWebConfigRepository(webPath, client_path string) *WebConfigRepository {
	return &WebConfigRepository{webPath: webPath, client_path: client_path}
}

func (wcr *WebConfigRepository) SaveConfig(webconfig config.WebConfig, jsonconfig string) error {
	// 1. Saves the site.json file
	siteConfigPath := filepath.Join(wcr.webPath, "site.json")
	if err := os.WriteFile(siteConfigPath, []byte(jsonconfig), 0o644); err != nil {
		return fmt.Errorf("error writing site.json: %w", err)
	}

	contentBase := filepath.Join(wcr.webPath, "content")

	meta_content := "export default {\n"

	for versionName, version := range webconfig.Versions {
		meta_content += fmt.Sprintf("\t'%s': { title:' ',type:'page'},\n", versionName)
		versionPath := filepath.Join(contentBase, versionName)
		if err := os.MkdirAll(versionPath, 0o755); err != nil {
			return fmt.Errorf("error creating version directory: %w", err)
		}

		// copyng index path
		srcPath := filepath.Join(wcr.client_path, version.MainView)
		dstPath := filepath.Join(versionPath, "index.mdx")
		copyFile(srcPath, dstPath)

		for _, tab := range version.Tabs {
			groupPath := filepath.Join(versionPath, tab.Group)
			if err := os.MkdirAll(groupPath, 0o755); err != nil {
				return fmt.Errorf("error creating group directory: %w", err)
			}

			for _, pagePath := range tab.Pages {
				srcPath := filepath.Join(wcr.client_path, pagePath)
				dstPath := filepath.Join(groupPath, filepath.Base(pagePath)) // mantiene el nombre del archivo

				if err := copyFile(srcPath, dstPath); err != nil {
					return fmt.Errorf("error copying page '%s' to '%s': %w", srcPath, dstPath, err)
				}
			}
		}
	}
	// Writing meta file
	meta_content += "}"
	meta_file, err := os.Create(filepath.Join(contentBase, "_meta.js"))
	if err != nil {
		return err
	}
	_, err = io.WriteString(meta_file, meta_content)
	if err != nil {
		return errors.New(fmt.Sprint("Error writing _meta.js: %s", err.Error()))
	}
	meta_file.Close()
	return nil
}

// Copia un archivo de src a dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying file content: %w", err)
	}

	return nil
}
