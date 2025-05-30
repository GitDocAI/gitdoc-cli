package initializer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/YHVCorp/DockitAi/cli/internals/config"
)

type Initializer struct {
	config *config.Config
}

func NewInitializer(config *config.Config) *Initializer {
	return &Initializer{
		config: config,
	}
}

func (i *Initializer) CheckEnvironment() error {
	// 1 check if folder exists
	if _, err := os.Stat(i.config.ProjectPath); os.IsNotExist(err) {
		return fmt.Errorf("Project path does not exist: %s", err.Error())
	}
	// 2 check npm is installed
	if _, err := exec.LookPath("npm"); err != nil {
		return fmt.Errorf("npm is not installed")
	}
	// 3 check git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed")
	}
	// 3 check package json dependencies
	if _, err := os.Stat(filepath.Join(i.config.ProjectPath, "node_modules")); os.IsNotExist(err) {
		return fmt.Errorf("modules not installed: %s", err.Error())
	}

	return nil
}

func (i *Initializer) SetEnvironment() error {
	// 1 create project file
	fmt.Println("Creating file path")
	err := os.MkdirAll(i.config.ProjectPath, os.FileMode(0o755))
	if err != nil {
		return err
	}
	// set project file content
	fmt.Println("Downloading Theme")
	cloneCmd := exec.Command("git", "clone", "https://github.com/YHVCorp/DockitAIThemes.git", i.config.ProjectPath)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("error downloading theme: %s", err.Error())
	}
	// install dependencies
	fmt.Println("Installing dependencies")
	npmCommand := exec.Command("npm", "install")
	npmCommand.Dir = i.config.ProjectPath
	if err := npmCommand.Run(); err != nil {
		return fmt.Errorf("error installing dependencies: %s", err.Error())
	}
	return nil
}
