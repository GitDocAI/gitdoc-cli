

package main

import (
	// Import the flag package
	"fmt"
	"os"
	"path/filepath"

	"github.com/GitDocAI/gitdocai-cli/cmd"
	"github.com/GitDocAI/gitdocai-cli/config"
	"github.com/GitDocAI/gitdocai-cli/utils"
	"github.com/spf13/cobra"
)

func main() {
	// Define the command-line flags and their default values
	var serveInput string

	// ################### Root Command ###################
	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "gitdocai",
		Short: "GitDocAI CLI to serve and manage documentation projects",
		Long: `GitDocAI is a CLI tool for working with documentation sites.

Available commands:

- dev: Starts a local server in development mode
- init-sample: Initializes a sample project structure
- update-themes: Downloads or updates the default themes
- version: Displays the current version of GitDocAI CLI`,
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			return cmd.ValidateSystemRequirements()
		},
	}

	// #################### Dev Command ###################
	// Create the dev command
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Serve the documentation locally in development mode",
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.Serve(serveInput, "dev")
		},
	}
	devCmd.Flags().StringVar(&serveInput, "input", utils.GetMyPath(), "Path to the source folder containing MDX and gitdocai.config.json")
	rootCmd.AddCommand(devCmd)

	// #################### Production Command ###################
	// Create the prod command
	prodCmd := &cobra.Command{
		Use:   "prod",
		Short: "Serve the documentation in production mode",
		RunE: func(command *cobra.Command, args []string) error {
			// In prod mode, always use the flusher's docs directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %v", err)
			}
			prodInput := filepath.Join(homeDir, ".gitdocai", "docs")
			return cmd.Serve(prodInput, "prod")
		},
	}
	// No input flag for prod - it always uses ~/.gitdocai/docs
	rootCmd.AddCommand(prodCmd)
	// ################### Init Sample Command ###################
	// Add the init-sample command to the root command
	initCmd := &cobra.Command{
		Use:   "init-sample [theme]",
		Short: "Initialize a sample GitDocAI project with the specified theme",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.InitSample(args[0])
		},
	}
	rootCmd.AddCommand(initCmd)

	// #################### Update Themes Command ###################
	// Add the update-themes command to the root command
	updateThemesCmd := &cobra.Command{
		Use:   "update-themes",
		Short: "Download or update the default themes",
		RunE: func(command *cobra.Command, args []string) error {
			utils.PrintBanner()
			utils.PrintHeader("Updating Themes")
			err := cmd.DownloadThemes(true)
			if err != nil {
				utils.PrintError(fmt.Sprintf("Failed to update themes: %v", err))
				return err
			}
			utils.PrintDivider()
			utils.PrintInfo("Themes are now up to date!")
			return nil
		},
	}
	rootCmd.AddCommand(updateThemesCmd)

	// ################### Version Command ###################
	// Add the version command to the root command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display the current version of GitDocAI CLI",
		Run: func(command *cobra.Command, args []string) {
			utils.PrintBanner()
			utils.PrintLogo(config.VERSION)
			utils.PrintDivider()
			utils.PrintInfo("A CLI tool for building beautiful documentation sites")
			utils.PrintURL("Documentation", "https://docs.gitdoc.ai")
			utils.PrintURL("GitHub", "https://github.com/GitDocAI/gitdocai-cli")
		},
	}
	rootCmd.AddCommand(versionCmd)

	// ################### Execute the Root Command ###################
	// Execute the root command, which will parse the flags and run the appropriate command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

