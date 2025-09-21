package config

import (
	"fmt"
	"os"
)

const VERSION = ""

var (
	ThemesRepositoryURL = "https://github.com/YHVCorp/themes.git"
	DefaultGitDocAIPath = fmt.Sprintf("%s/.gitdocai", os.Getenv("HOME"))
	DefaultThemesPath   = fmt.Sprintf("%s/themes", DefaultGitDocAIPath)
	ConfigFileName      = "gitdocai.config.json"
	ClearExceptionDirs  = []string{"startswith:.", "equals:node_modules"}
	ThemeExcludeDirs    = []string{"equals:content", "equals:public", "equals:gitdocai.config.json"}
)

func GetThemesPath() string {
	pathEnv := os.Getenv("GITDOCAI_THEMES_PATH")
	if pathEnv != "" {
		return pathEnv
	}
	return DefaultThemesPath
}
