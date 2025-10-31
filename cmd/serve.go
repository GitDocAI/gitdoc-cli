package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/schemas"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

func Serve(serveInput, command string) error {
	utils.PrintBanner()
	utils.PrintHeader(fmt.Sprintf("Starting %s mode", command))

	// Start flusher service for preview mode (Cloud)
	if command == "preview" {
		flusher, err := StartFlusher()
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to start flusher service: %v", err))
			return fmt.Errorf("failed to start flusher service: %v", err)
		}
		defer flusher.Stop()
		utils.PrintDivider()
	}

	// In preview mode, use prebuilt themes. In dev mode, use normal flow
	var err error
	if command == "preview" {
		utils.PrintStep("Using prebuilt themes...")
		err = DeployPrebuilt(serveInput, "")
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to deploy prebuilt theme: %v", err))
			return fmt.Errorf("failed to deploy prebuilt theme: %v", err)
		}
	} else {
		utils.PrintStep("Checking for theme updates...")
		err = DownloadThemes(true)
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
	}

	// Obtener configuración global del proyecto
	cnf := config.GetConfig()

	docs_file := filepath.Join(cnf.OutputPath, "public")
	build_path := filepath.Join(cnf.OutputPath, "dist") // Vite builds to dist/ by default

	// Configure authentication only for preview mode (Cloud)
	var auth schemas.Authentication
	if command == "preview" {
		auth = schemas.Authentication{
			OrganizationID: os.Getenv("ORGANIZATION_ID"),
			ProjectID:      os.Getenv("PROJECT_ID"),
			SessionToken:   os.Getenv("SESSION_TOKEN"),
		}
	}

	// Determine port: use PORT env var in preview (for Cloud Run), otherwise default to 8082
	port := "8082"
	if command == "preview" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			port = envPort
		}
	}

	sourceMode := "dev"
	if command == "preview" {
		sourceMode = "production"
	}

	// Prepare webhook URL for Vite env vars
	// In preview (Cloud), use relative path since frontend and backend are on same domain
	// In dev, use localhost URL
	var webhookURL string
	if command == "dev" {
		webhookURL = fmt.Sprintf("http://localhost:%s/api", port)
	} else {
		webhookURL = "/api" // Relative path for preview (Cloud)
	}

	envVars := map[string]any{
		"VITE_MODE":        sourceMode,
		"VITE_BACKEND_URL": webhookURL,
		"VITE_USE_BACKEND": sourceMode == "dev",
		"VITE_SOURCE":      "/docs",
	}

	// In preview mode, skip npm install and build (already done in prebuild)
	if command == "preview" {
		utils.PrintSuccess("Using prebuilt dependencies and assets")
	} else {
		// Dev mode: install dependencies and build
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

		utils.PrintDivider()
		utils.PrintSection(fmt.Sprintf("Running npm %s", command))
		utils.PrintDivider()

		err = utils.RunBlockingWithTerminalIO("npm", cnf.OutputPath, envVars, "run", "build")
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to run npm %s: %v", command, err))
			return fmt.Errorf("failed to run npm %s: %v", command, err)
		}
	}

	go func() {
		InitFolderListen(*cnf)
	}()
	utils.PrintSuccess("File watcher started")

	utils.PrintDivider()
	utils.PrintStep("Starting webhook server...")

	// Now that build is complete, start the webhook server
	webhookCfg := schemas.Config{
		Port:                port,
		AllowOrigin:         "*",
		BuildPath:           build_path,
		DocsPath:            docs_file,
		InputPath:           cnf.InputPath,
		ConfigPath:          filepath.Join(docs_file, config.ConfigFileName),
		RemoteFileSystemUrl: os.Getenv("FILESYSTEM_API_URL"),
		Authentication:      auth,
		Mode:                command, // "dev" or "preview"
	}

	actualPort, err := StartWebhook(webhookCfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to start webhook server: %v", err))
		return fmt.Errorf("failed to start webhook server: %v", err)
	}

	utils.PrintDivider()
	serverURL := fmt.Sprintf("http://localhost:%d", actualPort)
	utils.PrintSuccess(fmt.Sprintf("Server running at %s", serverURL))
	utils.PrintSuccess("Press Ctrl+C to stop")
	utils.PrintDivider()

	// Open browser automatically in dev mode
	if command == "dev" {
		go func() {
			// Wait a bit for the server to be fully ready
			time.Sleep(500 * time.Millisecond)
			if err := openBrowser(serverURL); err != nil {
				utils.PrintError(fmt.Sprintf("Failed to open browser: %v", err))
			}
		}()
	}

	select {}

}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}
