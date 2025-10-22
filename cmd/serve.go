package cmd

import (
	"fmt"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/utils"
)

func Serve(serveInput, webhook_url, authentication, command string) error {
	utils.PrintBanner()
	utils.PrintHeader(fmt.Sprintf("Starting %s mode", command))

	err := Deploy(serveInput, "", true)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to deploy: %v", err))
		return fmt.Errorf("failed to deploy: %v", err)
	}

	if webhook_url == "" {
		webhook_url = fmt.Sprintf(config.WebhookRoute, "8082")
		webhook := NewWebhookServer(serveInput, "8082")
		go func() {
			webhook.Start()
		}()
		utils.PrintSuccess("Webhook server started on port 8082")
	}

	//this is for now is just a test mode for local dev
	envVars := map[string]string{
		"WEBHOOK_URL":    webhook_url,
		"AUTHENTICATION": authentication,
	}
	cnf := config.GetConfig()

	err = utils.RunWithSpinner("Installing dependencies...", func() error {
		errStr, isError := utils.ExecuteCmd("npm", cnf.OutputPath, "install", "--force")
		if isError {
			return fmt.Errorf("%v", errStr)
		}
		return nil
	})
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to install dependencies: %v", err))
		return fmt.Errorf("failed to execute npm install: %v", err)
	}
	utils.PrintSuccess("Dependencies installed")

	go func() {
		InitFolderListen(*cnf)
	}()
	utils.PrintSuccess("File watcher started")

	utils.PrintDivider()
	utils.PrintSection(fmt.Sprintf("Running npm %s", command))
	utils.PrintDivider()

	err = utils.RunBlockingWithTerminalIO("npm", cnf.OutputPath, envVars, "run", command)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to run npm %s: %v", command, err))
		return fmt.Errorf("failed to run npm %s: %v", command, err)
	}
	return nil
}
