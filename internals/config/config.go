package config

import (
	"fmt"
	"os"
)

type Config struct {
	ProjectPath string
}

func NewConfig() Config {
	projectPath := os.Getenv("NEXTRAPATH")
	if projectPath == "" {
		projectPath = fmt.Sprintf("%s/.dockitai/template", os.Getenv("HOME"))
	}

	return Config{
		ProjectPath: projectPath,
	}
}
