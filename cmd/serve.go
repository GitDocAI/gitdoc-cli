package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/YHVCorp/gitdocai-cli/config"
	"github.com/YHVCorp/gitdocai-cli/schemas"
	"github.com/YHVCorp/gitdocai-cli/utils"
)

func Serve(serveInput, command string) error {
	utils.PrintBanner()
	utils.PrintHeader(fmt.Sprintf("Starting %s mode", command))

	// Start flusher service for production mode
	if command == "prod" {
		flusher, err := StartFlusher()
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to start flusher service: %v", err))
			return fmt.Errorf("failed to start flusher service: %v", err)
		}
		defer flusher.Stop()
		utils.PrintDivider()
	}

	utils.PrintStep("Checking for theme updates...")
	err := DownloadThemes(true)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to update themes: %v", err))
		return fmt.Errorf("failed to update themes: %v", err)
	}
	utils.PrintDivider()

	err = Deploy(serveInput, "", true)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to deploy: %v", err))
		return fmt.Errorf("failed to deploy: %v", err)
	}

	// Obtener configuración global del proyecto
	cnf := config.GetConfig()

	docs_file := filepath.Join(cnf.OutputPath, "public")
	webhookCfg := schemas.Config{
		Port:                "8082",
		AllowOrigin:         "*",
		DocsPath:            docs_file,
		InputPath:           cnf.InputPath,
		ConfigPath:          filepath.Join(docs_file, config.ConfigFileName),
		RemoteFileSystemUrl: os.Getenv("FILESYSTEM_API_URL"),
		Authentication: schemas.Authentication{
			OrganizationID: os.Getenv("ORGANIZATION_ID"),
			ProjectID:      os.Getenv("ORGANIZATION_ID"),
			SessionToken:   os.Getenv("SESSION_TOKEN"),
		},
	}

	port, err := StartWebhook(webhookCfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to start webhook server: %v", err))
		return fmt.Errorf("failed to start webhook server: %v", err)
	}
	webhookURL := fmt.Sprintf("http://localhost:%d/api", port)

	sourceMode := "dev"
	if command == "preview" {
		sourceMode = "production"
	} else if command == "prod" {
		sourceMode = "production"
	}

	envVars := map[string]any{
		"VITE_MODE":        sourceMode,
		"VITE_BACKEND_URL": webhookURL,
		"VITE_USE_BACKEND": sourceMode == "dev",
		"VITE_SOURCE":      "/docs",
	}

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
