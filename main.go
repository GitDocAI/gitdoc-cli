package main

import (
	// Import the flag package
	"fmt"
	"os"

	"github.com/YHVCorp/DockitAICLI/cmd"
	"github.com/YHVCorp/DockitAICLI/config"
	"github.com/YHVCorp/DockitAICLI/utils"
	"github.com/spf13/cobra"
)

func main() {
	// Define the command-line flags and their default values
	var serveInput, deployInput, deployOutput string

	// ################### Root Command ###################
	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "dockitai",
		Short: "DockitAI CLI to serve and manage documentation projects",
		Long: `DockitAI is a CLI tool for working with documentation sites.

Available commands:

- serve: Starts a local server to preview the documentation site
- init-sample: Initializes a sample project structure
- update-themes: Downloads or updates the default themes
- version: Displays the current version of DockitAI CLI`,
		PersistentPreRunE: func(command *cobra.Command, args []string) error {
			return cmd.ValidateSystemRequirements()
		},
	}

	// #################### Serve Command ###################
	// Create the serve command
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the documentation locally",
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.Serve(serveInput)
		},
	}
	serveCmd.Flags().StringVar(&serveInput, "input", utils.GetMyPath(), "Path to the source folder containing MDX and dockitai.config.json")
	rootCmd.AddCommand(serveCmd)

	// ################### Deploy Command ###################
	// Create the deploy command
	deployCmd := &cobra.Command{
		Use:    "deploy",
		Short:  "[Internal] Deploy a DockitAI project",
		Hidden: true,
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.Deploy(deployInput, deployOutput)
		},
	}

	deployCmd.Flags().StringVar(&deployInput, "input", "", "Path to the source folder")
	deployCmd.Flags().StringVar(&deployOutput, "output", "", "Path to the output folder")
	rootCmd.AddCommand(deployCmd)

	// ################### Init Sample Command ###################
	// Add the init-sample command to the root command
	initCmd := &cobra.Command{
		Use:   "init-sample",
		Short: "Initialize a sample DockitAI project",
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.InitSample()
		},
	}
	rootCmd.AddCommand(initCmd)

	// #################### Update Themes Command ###################
	// Add the update-themes command to the root command
	updateThemesCmd := &cobra.Command{
		Use:   "update-themes",
		Short: "Download or update the default themes",
		RunE: func(command *cobra.Command, args []string) error {
			return cmd.DownloadThemes(true)
		},
	}
	rootCmd.AddCommand(updateThemesCmd)

	// ################### Version Command ###################
	// Add the version command to the root command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display the current version of DockitAI CLI",
		Run: func(command *cobra.Command, args []string) {
			fmt.Printf("DockitAI CLI version: %s\n", config.VERSION)
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
