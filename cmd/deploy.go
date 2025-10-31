package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

// DeployPrebuilt copies a prebuilt theme from prebuilt-themes directory to the output directory
// This is used in preview mode to skip the npm build step
func DeployPrebuilt(deployInput, deployOutput string) error {
	// Load config to get the theme name
	err := config.LoadConfig(deployInput, deployOutput)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cnf := config.GetConfig()
	themeName := cnf.SiteConfig.Theme

	// Determine prebuilt themes location
	// Check env var first, then try Docker location, finally local location
	var prebuiltBasePath string
	if envPath := os.Getenv("PREBUILD_OUTPUT_PATH"); envPath != "" {
		prebuiltBasePath = envPath
	} else if _, err := os.Stat("/prebuilt-themes"); err == nil {
		// Docker: /prebuilt-themes exists
		prebuiltBasePath = "/prebuilt-themes"
	} else {
		// Local: ~/.gitdocai/prebuilt-themes
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		prebuiltBasePath = filepath.Join(homeDir, ".gitdocai", "prebuilt-themes")
	}

	prebuiltPath := filepath.Join(prebuiltBasePath, themeName)

	// Check if prebuilt theme exists
	if !utils.CheckIfPathExist(prebuiltPath) {
		return fmt.Errorf("prebuilt theme '%s' not found at %s. Available themes may not include this one", themeName, prebuiltPath)
	}

	utils.PrintSuccess(fmt.Sprintf("Found prebuilt theme: %s", themeName))

	// Create output path if it doesn't exist
	exists, err := utils.CreatePathIfNotExist(cnf.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output path: %w", err)
	}
	if exists {
		// Clear output path except certain directories
		err := utils.ClearFolderExcept(cnf.OutputPath, config.ClearExceptionDirs)
		if err != nil {
			return fmt.Errorf("failed to clear output path: %w", err)
		}
	}

	// Copy prebuilt theme to output directory
	// Exclude node_modules since it's already in the prebuild
	utils.PrintStep("Copying prebuilt theme...")
	err = copyDirExclude(prebuiltPath, cnf.OutputPath, []string{"node_modules"})
	if err != nil {
		return fmt.Errorf("failed to copy prebuilt theme: %w", err)
	}

	// Copy user's content and config
	utils.PrintStep("Copying user content...")

	// Copy user's gitdocai.config.json to output/public/
	userConfig := filepath.Join(cnf.InputPath, config.ConfigFileName)
	outputConfig := filepath.Join(cnf.OutputPath, "public", config.ConfigFileName)
	if err := utils.CopyFile(userConfig, outputConfig, 0644, true); err != nil {
		return fmt.Errorf("failed to copy config: %w", err)
	}

	// Copy user's content directory to output/public/docs/
	contentSrc := filepath.Join(cnf.InputPath, "content")
	contentDest := filepath.Join(cnf.OutputPath, "public", "docs")
	if utils.CheckIfPathExist(contentSrc) {
		if err := copyDir(contentSrc, contentDest); err != nil {
			return fmt.Errorf("failed to copy content: %w", err)
		}
	}

	// Copy user's public assets if they exist
	publicSrc := filepath.Join(cnf.InputPath, "public")
	publicDest := filepath.Join(cnf.OutputPath, "public")
	if utils.CheckIfPathExist(publicSrc) {
		if err := copyDir(publicSrc, publicDest); err != nil {
			return fmt.Errorf("failed to copy public assets: %w", err)
		}
	}

	// Re-read the config from the new location to update configInstance
	err = config.LoadConfig(deployInput, deployOutput)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Run component replacement (same as regular deploy)
	err = ReplaceComponents(*cnf)
	if err != nil {
		return fmt.Errorf("failed to replace components: %w", err)
	}

	utils.PrintSuccess("Prebuilt theme deployed successfully")
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return copyDirExclude(src, dst, nil)
}

// copyDirExclude recursively copies a directory, excluding specified directories
func copyDirExclude(src, dst string, exclude []string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Check if this path should be excluded
		if info.IsDir() {
			for _, excl := range exclude {
				if info.Name() == excl || relPath == excl {
					return filepath.SkipDir
				}
			}
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return utils.CopyFile(path, destPath, info.Mode(), true)
	})
}

func Deploy(deployInput, deployOutput string, showSpinner bool) error {
	deployFunc := func() error {
		err := config.LoadConfig(deployInput, deployOutput)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		err = DownloadThemes(false)
		if err != nil {
			return fmt.Errorf("failed to download themes: %w", err)
		}

		cnf := config.GetConfig()
		exists, err := utils.CreatePathIfNotExist(cnf.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to create output path: %w", err)
		}
		if exists {
			err := utils.ClearFolderExcept(cnf.OutputPath, config.ClearExceptionDirs)
			if err != nil {
				return fmt.Errorf("failed to clear output path: %w", err)
			}
		}

		err = CopyTheme(cnf.OutputPath, cnf.SiteConfig.Theme)
		if err != nil {
			return fmt.Errorf("failed to copy theme: %w", err)
		}

		err = ReplaceComponents(*cnf)
		if err != nil {
			return fmt.Errorf("failed to replace components: %w", err)
		}

		return nil
	}

	var err error
	if showSpinner {
		err = utils.RunWithSpinner("Deploying project...", deployFunc)
	} else {
		err = deployFunc()
	}

	if err != nil {
		return err
	}

	utils.PrintSuccess("Deployment completed successfully")

	return nil
}
