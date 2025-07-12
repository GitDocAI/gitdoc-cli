package config

import (
	"fmt"
	"os"
)

const VERSION = ""

var (
	ThemesRepositoryURL = "https://github.com/YHVCorp/themes.git"
	DefaultDockitAIPath = fmt.Sprintf("%s/.dockitai", os.Getenv("HOME"))
	DefaultThemesPath   = fmt.Sprintf("%s/themes", DefaultDockitAIPath)
	ConfigFileName      = "dockitai.config.json"
	ClearExceptionDirs  = []string{"startswith:.", "equals:node_modules"}
)

func GetThemesPath() string {
	pathEnv := os.Getenv("DOCKITAI_THEMES_PATH")
	if pathEnv != "" {
		return pathEnv
	}
	return DefaultThemesPath
}
