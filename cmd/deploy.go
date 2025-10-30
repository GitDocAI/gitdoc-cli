package cmd

import (
	"fmt"

	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

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
