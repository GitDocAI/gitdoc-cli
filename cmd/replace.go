package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
	"github.com/GitDocAI/inkflow"
)

func ReplaceComponents(cnf config.Config) error {

	mdxToJsonParser := inkflow.NewParser(nil)

	configFile := filepath.Join(cnf.InputPath, config.ConfigFileName)
	if !utils.CheckIfPathExist(configFile) {
		return fmt.Errorf("%s not found in input path: %s", config.ConfigFileName, cnf.InputPath)
	}
	if err := utils.CopyFile(configFile, filepath.Join(cnf.OutputPath, config.ConfigFileName), os.FileMode(0644),false); err != nil {
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
			if err := utils.CopyDir(srcPath, destPath,true); err != nil {
				return fmt.Errorf("error copying directory '%s': %v", srcPath, err)
			}
		} else {
			if err := utils.CopyFile(srcPath, destPath, info.Mode(),true); err != nil {
				return fmt.Errorf("error copying file '%s': %v", destPath, err)
			}
		}
	}

	for _, path := range contentPaths {
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
			if err := utils.CopyDir(srcPath, destPath,true); err != nil {
				return fmt.Errorf("error copying directory '%s': %v", srcPath, err)
			}
		} else {
			// Read the MDX file
			content, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("error reading file '%s': %v", srcPath, err)
			}

			// Parse the MDX content
			result, err := mdxToJsonParser.Parse(string(content))
			if err != nil {
				return fmt.Errorf("error parsing MDX content from '%s': %v", srcPath, err)
			}

			// Marshal the parsed content to JSON
			jsonData, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("error marshaling JSON from '%s': %v", srcPath, err)
			}

			// Change the file extension to .json
			destPath = strings.Replace(destPath, filepath.Ext(destPath), ".json", 1)

			// Write the JSON data to the destination file
			err = os.WriteFile(destPath, jsonData, info.Mode())
			if err != nil {
				return fmt.Errorf("error writing JSON file '%s': %v", destPath, err)
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
