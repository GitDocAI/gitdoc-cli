package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

func InitSample(themeName string) error {
	if themeName == "" {
		return fmt.Errorf("theme name is required")
	}

	utils.PrintBanner()
	utils.PrintHeader("Initializing Sample Project")

	utils.PrintStep("Downloading themes...")
	err := DownloadThemes(false)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to download themes: %v", err))
		return fmt.Errorf("failed to download themes: %w", err)
	}

	themeDir := filepath.Join(config.GetThemesPath(), themeName)
	if !utils.CheckIfPathExist(themeDir) {
		utils.PrintError(fmt.Sprintf("Theme '%s' does not exist", themeName))
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	currentPath := utils.GetMyPath()

	utils.PrintStep(fmt.Sprintf("Copying theme '%s'...", themeName))
	err = copyThemeToCurrentPath(themeDir, currentPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to copy theme files: %v", err))
		return fmt.Errorf("failed to copy theme files: %w", err)
	}

	utils.PrintDivider()
	utils.PrintSuccess(fmt.Sprintf("Sample project initialized with theme '%s'", themeName))
	utils.PrintInfo("Next steps:")
	utils.PrintList("Edit gitdocai.config.json to configure your site")
	utils.PrintList("Add your content in MDX files")
	utils.PrintList("Run 'gitdocai dev' to start development server")
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
		if err := utils.CopyFile(configFile, destConfig, info.Mode(), false); err != nil {
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
			if err := utils.CopyDir(srcPath, destPath, false); err != nil {
				return fmt.Errorf("error copying directory %s: %w", srcPath, err)
			}
		} else {
			if err := utils.CopyFile(srcPath, destPath, info.Mode(), false); err != nil {
				return fmt.Errorf("error copying file %s: %w", srcPath, err)
			}
		}
	}

	return nil
}
