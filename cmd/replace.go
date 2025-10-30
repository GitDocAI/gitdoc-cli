package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

func ReplaceComponents(cnf config.Config) error {
	configFile := filepath.Join(cnf.InputPath, config.ConfigFileName)
	if !utils.CheckIfPathExist(configFile) {
		return fmt.Errorf("%s not found in input path: %s", config.ConfigFileName, cnf.InputPath)
	}
	if _, err := utils.CreatePathIfNotExist(filepath.Join(cnf.OutputPath, "public")); err != nil {
		return fmt.Errorf("could not create generation file : %v ", err)
	}
	if err := utils.CopyFile(configFile, filepath.Join(cnf.OutputPath, "public", config.ConfigFileName), os.FileMode(0644), false); err != nil {
		return fmt.Errorf("error copying file '%s': %v", configFile, err)
	}

	publicPaths, contentPaths := ExtractUsedPaths(cnf.SiteConfig)
	apirefMap := ExtractApirefPaths(cnf.SiteConfig)

	for _, path := range publicPaths {
		srcPath := filepath.Join(cnf.InputPath, path)
		destPath := filepath.Join(cnf.OutputPath, "public", path)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("Warning: Public path '%s' does not exist, skipping.\n", srcPath)
			continue
		} else if err != nil {
			return fmt.Errorf("error accessing '%s': %v", srcPath, err)
		}

		if info.IsDir() {
			if err := utils.CopyDir(srcPath, destPath, true); err != nil {
				return fmt.Errorf("error copying directory '%s': %v", srcPath, err)
			}
		} else {
			if err := utils.CopyFile(srcPath, destPath, info.Mode(), true); err != nil {
				return fmt.Errorf("error copying file '%s': %v", destPath, err)
			}
		}
	}

	for _, path := range contentPaths {
		fmt.Printf("Processing content path: %s\n", path)
		srcPath := filepath.Join(cnf.InputPath, path)
		destPath := filepath.Join(cnf.OutputPath, "public", path)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("Warning: Content path '%s' does not exist, skipping.\n", srcPath)
			continue
		} else if err != nil {
			return fmt.Errorf("error accessing '%s': %v", srcPath, err)
		}

		if info.IsDir() {
			if err := utils.CopyDir(srcPath, destPath, true); err != nil {
				return fmt.Errorf("error copying directory '%s': %v", srcPath, err)
			}
		} else {
			// Use CopyFile with parsing enabled - it will only convert .mdx/.md files
			if err := utils.CopyFile(srcPath, destPath, info.Mode(), true); err != nil {
				return fmt.Errorf("error copying file '%s': %v", srcPath, err)
			}
		}
	}

	for path, item := range apirefMap {
		if item.Reference == "" || item.Method == "" {
			fmt.Printf("Warning: Route not configured for '%s' skipping.\n", item.Title)
			continue //ignore file if theres not route configured
		}

		openapi_path := filepath.Join(cnf.InputPath, item.Path)
		apiref, err := utils.ExtractRequestInfoFromOpenapi(item.Title, openapi_path, item.Method, item.Reference)
		if err != nil {
			fmt.Printf("Warning: Api Reference path '%s' does not exist, skipping.\n", item.Path)
			continue //ignore file if theres not apiref
		}

		jsonData, err := json.Marshal(apiref)
		if err != nil {
			return fmt.Errorf("error marshaling JSON from '%s': %v", path, err)
		}

		destPath := filepath.Join(cnf.OutputPath, "public", path)
		destPath = strings.Replace(destPath, filepath.Ext(destPath), ".json", 1)

		err = os.WriteFile(destPath, jsonData, os.FileMode(0644))
		if err != nil {
			fmt.Printf("Warning: Api Reference path '%s' does not exist, skipping.\n", item.Reference)
			return fmt.Errorf("error creating apireference file: '%s' : %v", destPath, err)
		}
	}

	return nil
}
