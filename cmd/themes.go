package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
)

func cleanAndCloneThemesRepo() error {
	os.RemoveAll(config.GetThemesPath())
	cloneCmd := exec.Command("git", "clone", config.ThemesRepositoryURL, config.GetThemesPath())
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("error cloning themes repository: %v", err)
	}
	return nil
}

func DownloadThemes(forceUpdate bool) error {
	exists, err := utils.CreatePathIfNotExist(config.GetThemesPath())
	if err != nil {
		return fmt.Errorf("error creating default themes path: %v", err)
	}

	if !exists {
		if forceUpdate {
			utils.PrintStep("Downloading themes...")
		}
		cloneCmd := exec.Command("git", "clone", config.ThemesRepositoryURL, config.GetThemesPath())
		if err := cloneCmd.Run(); err != nil {
			if err := cleanAndCloneThemesRepo(); err != nil {
				return err
			}
		}
		if forceUpdate {
			utils.PrintSuccess("Themes downloaded successfully")
		}
	} else if forceUpdate {
		utils.PrintStep("Updating themes...")
		updateCmd := exec.Command("git", "-C", config.GetThemesPath(), "pull")
		if err := updateCmd.Run(); err != nil {
			if err := cleanAndCloneThemesRepo(); err != nil {
				return err
			}
		}
		utils.PrintSuccess("Themes updated successfully")
	}

	return nil
}

func CopyTheme(targetDir string, themeName string) error {
	sourceDir := filepath.Join(config.GetThemesPath(), themeName)

	if !utils.CheckIfPathExist(sourceDir) {
		return fmt.Errorf("theme '%s' does not exist", themeName)
	}

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip excluded files and directories at the root level
		if relPath != "." {
			name := filepath.Base(relPath)
			// For root level items, check if they should be excluded
			if !strings.Contains(relPath, string(os.PathSeparator)) {
				if utils.ShouldExclude(name, config.ThemeExcludeDirs) {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
		}

		destPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return utils.CopyFile(path, destPath, info.Mode())
	})

	if err != nil {
		return fmt.Errorf("error copying theme: %w", err)
	}

	return nil
}
