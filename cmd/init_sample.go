package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
)

func InitSample(themeName string) error {
	if themeName == "" {
		return fmt.Errorf("theme name is required")
	}

	err := DownloadThemes(false)
	if err != nil {
		return fmt.Errorf("failed to download themes: %w", err)
	}

	themeDir := filepath.Join(config.GetThemesPath(), themeName)
	if !utils.CheckIfPathExist(themeDir) {
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	currentPath := utils.GetMyPath()

	err = copyThemeToCurrentPath(themeDir, currentPath)
	if err != nil {
		return fmt.Errorf("failed to copy theme files: %w", err)
	}

	fmt.Printf("Sample project initialized with theme '%s' successfully.\n", themeName)
	return nil
}

func copyThemeToCurrentPath(themeDir, targetDir string) error {
	configFile := filepath.Join(themeDir, config.ConfigFileName)
	if utils.CheckIfPathExist(configFile) {
		destConfig := filepath.Join(targetDir, config.ConfigFileName)
		info, err := os.Stat(configFile)
		if err != nil {
			return fmt.Errorf("error accessing config file: %w", err)
		}
		if err := utils.CopyFile(configFile, destConfig, info.Mode()); err != nil {
			return fmt.Errorf("error copying config file: %w", err)
		}
	}

	publicDir := filepath.Join(themeDir, "public")
	if utils.CheckIfPathExist(publicDir) {
		err := copyDirectoryContents(publicDir, targetDir)
		if err != nil {
			return fmt.Errorf("error copying public contents: %w", err)
		}
	}

	contentDir := filepath.Join(themeDir, "content")
	if utils.CheckIfPathExist(contentDir) {
		err := copyDirectoryContents(contentDir, targetDir)
		if err != nil {
			return fmt.Errorf("error copying content: %w", err)
		}
	}

	return nil
}

func copyDirectoryContents(sourceDir, targetDir string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("could not read directory %s: %w", sourceDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(targetDir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("error getting file info: %w", err)
		}

		if entry.IsDir() {
			if err := utils.CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("error copying directory %s: %w", srcPath, err)
			}
		} else {
			if err := utils.CopyFile(srcPath, destPath, info.Mode()); err != nil {
				return fmt.Errorf("error copying file %s: %w", srcPath, err)
			}
		}
	}

	return nil
}
