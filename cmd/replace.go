package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/YHVCorp/DockitAICLI/config"
	"github.com/YHVCorp/DockitAICLI/utils"
)

func ReplaceComponents(cnf config.Config) error {
	configFile := filepath.Join(cnf.InputPath, config.ConfigFileName)
	if !utils.CheckIfPathExist(configFile) {
		return fmt.Errorf("%s not found in input path: %s", config.ConfigFileName, cnf.InputPath)
	}
	if err := utils.CopyFile(configFile, filepath.Join(cnf.OutputPath, config.ConfigFileName), os.FileMode(0644)); err != nil {
		return fmt.Errorf("error copying file '%s': %v", configFile, err)
	}

	publicPaths, contentPaths := ExtractUsedPaths(cnf.SiteConfig)

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
			if err := utils.CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("error copying directory '%s': %v", srcPath, err)
			}
		} else {
			if err := utils.CopyFile(srcPath, destPath, info.Mode()); err != nil {
				return fmt.Errorf("error copying file '%s': %v", destPath, err)
			}
		}
	}

	for _, path := range contentPaths {
		srcPath := filepath.Join(cnf.InputPath, path)
		destPath := filepath.Join(cnf.OutputPath, "content", path)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("Warning: Content path '%s' does not exist, skipping.\n", srcPath)
			continue
		} else if err != nil {
			return fmt.Errorf("error accessing '%s': %v", srcPath, err)
		}

		if info.IsDir() {
			if err := utils.CopyDir(srcPath, destPath); err != nil {
				return fmt.Errorf("error copying directory '%s': %v", srcPath, err)
			}
		} else {
			if err := utils.CopyFile(srcPath, destPath, info.Mode()); err != nil {
				return fmt.Errorf("error copying file '%s': %v", destPath, err)
			}
		}
	}

	return nil
}
