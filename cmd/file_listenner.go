package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
	"github.com/fsnotify/fsnotify"
)

type PathMapping struct {
	InputPath    string
	OutputPath   string
	PublicPaths  []string
	ContentPaths []string
	ConfigFile   string
}

func setup_watcher(mapping PathMapping, callback func(string, string)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("error creating watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	files_to_listen := []string{}

	files_to_listen = append(files_to_listen, mapping.ConfigFile)

	for _, path := range mapping.PublicPaths {
		srcPath := filepath.Join(mapping.InputPath, path)
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

func handleFileChange(mapping PathMapping, filePath string, operation string) {
	// Determine file type and destination
	relPath, err := filepath.Rel(mapping.InputPath, filePath)
	if err != nil {
		return
	}


	var destPath string

	// Check if it's the config file
	if filePath == mapping.ConfigFile {
		destPath = filepath.Join(mapping.OutputPath, config.ConfigFileName)
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
					// Map to output/content
					destPath = filepath.Join(mapping.OutputPath, "content", relPath)
					break
				}
			}
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

		// Copy file
		if err := utils.CopyFile(filePath, destPath, os.FileMode(0644)); err != nil {
			fmt.Printf("error copying file %s to %s: %v\n", filePath, destPath, err)
		}
	case "remove":
		// Remove file from output
		if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("error removing file %s: %v\n", destPath, err)
		}
	case "rename":
		// For rename, we remove the old file (new file will trigger create event)
		if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("error removing renamed file %s: %v\n", destPath, err)
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
	}

	// Start watching
	setup_watcher(mapping, func(filePath, operation string) {
		handleFileChange(mapping, filePath, operation)
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

