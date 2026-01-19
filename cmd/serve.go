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
	config.LoadConfig(serveInput,"")

	utils.PrintStep("Checking for theme updates...")
	err := DownloadThemes(false)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to update themes: %v", err))
		return fmt.Errorf("failed to update themes: %v", err)
	}

	utils.PrintDivider()
	cnf := config.GetConfig()


	webhookCfg := schemas.Config{
		Port:                "8082",
		AllowOrigin:         "*",
		BuildPath:           filepath.Join(config.GetThemesPath(),"dist"),
		DocsPath:            cnf.InputPath,
		InputPath:           cnf.InputPath,
		ConfigPath:          filepath.Join(cnf.InputPath, config.ConfigFileName),
		RemoteFileSystemUrl: os.Getenv("FILESYSTEM_API_URL"),
	}

	actualPort, err := StartWebhook(webhookCfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to start webhook server: %v", err))
		return fmt.Errorf("failed to start webhook server: %v", err)
	}
	webhookURL := fmt.Sprintf("http://localhost:%d/api", actualPort)

	sourceMode := "visual-editor"
	if command == "preview" {
		sourceMode = "production"
	} else if command == "prod" {
		sourceMode = "production"
	}

	envVars := map[string]any{
		"VITE_MODE":        sourceMode,
		"VITE_BACKEND_URL": webhookURL,
	}

	err = utils.RunWithSpinner("Installing dependencies...", func() error {
		errStr, isError := utils.ExecuteCmd("npm", config.GetThemesPath(), "install", "--force")
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

	err = utils.RunBlockingWithTerminalIO("npm", config.GetThemesPath(), envVars, "run", "build")
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to run npm %s: %v", command, err))
		return fmt.Errorf("failed to run npm %s: %v", command, err)
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

