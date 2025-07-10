package config

import (
	"fmt"
	"os"
)

var (
	ThemesRepositoryURL = "https://github.com/YHVCorp/themes.git"
	DefaultDockitAIPath = fmt.Sprintf("%s/.dockitai", os.Getenv("HOME"))
	DefaultThemesPath   = fmt.Sprintf("%s/themes", DefaultDockitAIPath)
	ConfigFileName      = "dockitai.config.json"
	ClearExceptionDirs  = []string{"startswith:.", "equals:node_modules"}
)
