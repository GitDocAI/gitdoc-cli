package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/inkflow"
	"github.com/GitDocAI/gitdocai-cli/schemas"
	"github.com/GitDocAI/gitdocai-cli/utils"
	"github.com/fsnotify/fsnotify"
)

type PathMapping struct {
	InputPath    string
	OutputPath   string
	PublicPaths  []string
	ContentPaths []string
	ConfigFile   string
	SiteConfig   schemas.SiteConfig
}

func setup_watcher(mapping PathMapping, callback func(string, string)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	files_to_listen := []string{}

	files_to_listen = append(files_to_listen, mapping.InputPath)
	files_to_listen = append(files_to_listen, mapping.ConfigFile)

	for _, path := range mapping.PublicPaths {
		srcPath := filepath.Join(mapping.InputPath, path)
		// Check if directory exists before trying to watch it
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		files_to_listen = append(files_to_listen, srcPath)
		inner, err := GetDirectoryContents(srcPath)
		if err != nil {
			fmt.Printf("error reading directory %s: %v\n", srcPath, err)
		} else {
			files_to_listen = append(files_to_listen, inner...)
		}
	}

	for _, path := range mapping.ContentPaths {
		srcPath := filepath.Join(mapping.InputPath, path)
		// Check if directory exists before trying to watch it
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		files_to_listen = append(files_to_listen, srcPath)
		inner, err := GetDirectoryContents(srcPath)
		if err != nil {
			fmt.Printf("error reading directory %s: %v\n", srcPath, err)
		} else {
			files_to_listen = append(files_to_listen, inner...)
		}
	}

	for _, file := range files_to_listen {
		if err := watcher.Add(file); err != nil {
			fmt.Printf("error watching %s: %v\n", file, err)
		}
	}

	go func() {
		debounce := map[string]*time.Timer{}
		const delay = 250 * time.Millisecond

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					if t, ok := debounce[event.Name]; ok {
						t.Stop()
					}
					debounce[event.Name] = time.AfterFunc(delay, func() {
						callback(event.Name, "write")
					})

				case event.Op&fsnotify.Create == fsnotify.Create:
					watcher.Add(event.Name)
					callback(event.Name, "create")

				case event.Op&fsnotify.Remove == fsnotify.Remove:
					delete(debounce, event.Name)
					callback(event.Name, "remove")

				case event.Op&fsnotify.Rename == fsnotify.Rename:
					delete(debounce, event.Name)
					callback(event.Name, "rename")
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("watcher error: %v\n", err)
			}
		}
	}()

	done := make(chan struct{})
	<-done
}

func shouldIgnoreFile(filePath string) bool {
	base := filepath.Base(filePath)

	// Ignore temporary files
	if strings.HasPrefix(base, ".") {
		return true
	}
	if strings.HasSuffix(base, "~") {
		return true
	}
	if strings.Contains(base, ".tmp") {
		return true
	}
	if strings.HasSuffix(base, ".swp") {
		return true
	}
	if strings.HasSuffix(base, ".swx") {
		return true
	}

	return false
}

func handleFileChange(mapping *PathMapping, filePath string, operation string) {
	// Ignore temporary and hidden files
	if shouldIgnoreFile(filePath) {
		return
	}

	// Determine file type and destination
	relPath, err := filepath.Rel(mapping.InputPath, filePath)
	if err != nil {
		return
	}

	var destPath string

	// Check if it's the config file
	if filePath == mapping.ConfigFile {
		destPath = filepath.Join(mapping.OutputPath, "public", config.ConfigFileName)
		//ignore if there was an error
		if newConfig, err := getConfigFileJson(mapping.ConfigFile); err == nil {
			if hasChangedNavigation(mapping.SiteConfig, *newConfig) {
				mapping.SiteConfig = *newConfig
				config := config.Config{
					OutputPath: mapping.OutputPath,
					InputPath:  mapping.InputPath,
					SiteConfig: *newConfig,
				}
				ReplaceComponents(config)
				return
			}
			// Navigation didn't change, but config was modified - copy it
			mapping.SiteConfig = *newConfig
			if err := utils.CopyFile(filePath, destPath, os.FileMode(0644), false); err != nil {
				fmt.Printf("error copying config file %s to %s: %v\n", filePath, destPath, err)
			}
			return
		} else {
			return
		}

	} else {
		// Check if file is in public paths
		isPublic := false
		for _, pubPath := range mapping.PublicPaths {
			if strings.HasPrefix(relPath, pubPath) {
				// Map to output/public
				destPath = filepath.Join(mapping.OutputPath, "public", relPath)
				isPublic = true
				break
			}
		}

		// Check if file is in content paths
		if !isPublic {
			for _, contentPath := range mapping.ContentPaths {
				if strings.HasPrefix(relPath, contentPath) {
					// Map to output/public
					destPath = filepath.Join(mapping.OutputPath, "public", relPath)
					break
				}
			}
		}

		// If not in any specific path, still copy it to output/public
		// This handles files like assets/, images/, etc. that come from the filesystem API
		if destPath == "" {
			destPath = filepath.Join(mapping.OutputPath, "public", relPath)
		}
	}

	// Handle operation
	switch operation {
	case "write", "create":
		// Ensure destination directory exists
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			fmt.Printf("error creating directory %s: %v\n", destDir, err)
			return
		}

		// Check if this is an MDX file in content paths
		isMDX := strings.HasSuffix(strings.ToLower(filePath), ".mdx")
		isInContentPaths := false
		for _, contentPath := range mapping.ContentPaths {
			if strings.HasPrefix(relPath, contentPath) {
				isInContentPaths = true
				break
			}
		}

		if isMDX && isInContentPaths {
			// Convert MDX to JSON
			if err := convertMDXToJSON(filePath, destPath); err != nil {
				fmt.Printf("error converting MDX to JSON %s: %v\n", filePath, err)
			}
		} else {
			// Copy file as-is
			if err := utils.CopyFile(filePath, destPath, os.FileMode(0644), false); err != nil {
				fmt.Printf("error copying file %s to %s: %v\n", filePath, destPath, err)
			}
		}
	case "remove":
		// For MDX files, remove the .json file instead
		removeDestPath := destPath
		if strings.HasSuffix(strings.ToLower(filePath), ".mdx") {
			removeDestPath = strings.Replace(destPath, filepath.Ext(destPath), ".json", 1)
		}

		// Remove file from output
		if err := os.RemoveAll(removeDestPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("error removing file %s: %v\n", removeDestPath, err)
		}
	case "rename":
		// For rename, we remove the old file (new file will trigger create event)
		removeDestPath := destPath
		if strings.HasSuffix(strings.ToLower(filePath), ".mdx") {
			removeDestPath = strings.Replace(destPath, filepath.Ext(destPath), ".json", 1)
		}

		if err := os.RemoveAll(removeDestPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("error removing renamed file %s: %v\n", removeDestPath, err)
		}
	}
}

func InitFolderListen(cnf config.Config) {
	configFile := filepath.Join(cnf.InputPath, config.ConfigFileName)
	if !utils.CheckIfPathExist(configFile) {
		fmt.Printf("error: %s not found in input path: %s\n", config.ConfigFileName, cnf.InputPath)
		return
	}
	publicPaths, contentPaths := ExtractUsedPaths(cnf.SiteConfig)

	// Create mapping for watcher
	mapping := PathMapping{
		InputPath:    cnf.InputPath,
		OutputPath:   cnf.OutputPath,
		PublicPaths:  publicPaths,
		ContentPaths: contentPaths,
		ConfigFile:   configFile,
		SiteConfig:   cnf.SiteConfig,
	}

	// Start watching
	setup_watcher(mapping, func(filePath, operation string) {
		handleFileChange(&mapping, filePath, operation)
	})
}

func GetDirectoryContents(src string) ([]string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, err
	}

	filepaths := make([]string, 0)
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		if entry.IsDir() {
			paths, err := GetDirectoryContents(srcPath)
			if err != nil {
				return nil, err
			}
			filepaths = append(filepaths, paths...)
		} else {
			filepaths = append(filepaths, srcPath)
		}
	}
	return filepaths, nil
}

func getConfigFileJson(path string) (*schemas.SiteConfig, error) {
	configText, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error readign config File: %v", err)
	}

	var configJson schemas.SiteConfig
	if err := json.Unmarshal(configText, &configJson); err != nil {
		return nil, fmt.Errorf("invalid config file")
	}

	return &configJson, nil
}

func hasChangedNavigation(original schemas.SiteConfig, new_config schemas.SiteConfig) bool {
	// Verify if there's a change on navigation
	originalBytes, _ := json.Marshal(original.Navigation)
	originalHash := sha256.Sum256(originalBytes)
	newBytes, _ := json.Marshal(new_config.Navigation)
	newHash := sha256.Sum256(newBytes)

	return newHash != originalHash

}

// convertMDXToJSON reads an MDX file, parses it, and writes it as JSON
func convertMDXToJSON(mdxFilePath, destPath string) error {
	// Read the MDX file
	content, err := os.ReadFile(mdxFilePath)
	if err != nil {
		return fmt.Errorf("error reading MDX file: %v", err)
	}

	// Create MDX parser
	mdxToJsonParser := inkflow.NewParser(nil)

	// Parse the MDX content
	result, err := mdxToJsonParser.Parse(string(content))
	if err != nil {
		return fmt.Errorf("error parsing MDX content: %v", err)
	}

	// Marshal the parsed content to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Change the file extension to .json
	jsonDestPath := strings.Replace(destPath, filepath.Ext(destPath), ".json", 1)

	// Write the JSON data to the destination file
	if err := os.WriteFile(jsonDestPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing JSON file: %v", err)
	}

	return nil
}
