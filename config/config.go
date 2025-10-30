package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GitDocAI/gitdocai-cli/schemas"
	"github.com/GitDocAI/gitdocai-cli/utils"
)

type Config struct {
	InputPath  string
	OutputPath string
	SiteConfig schemas.SiteConfig
}

var configInstance *Config

func LoadConfig(inputPath, outputPath string) error {
	configInstance = &Config{
		InputPath:  inputPath,
		OutputPath: outputPath,
	}

	if outputPath == "" {
		configInstance.OutputPath = fmt.Sprintf("%s/.gitdocai/generated", os.Getenv("HOME"))
	}

	if inputPath == "" {
		configInstance.InputPath = utils.GetMyPath()
	} else {
		if !utils.CheckIfPathExist(inputPath) {
			return fmt.Errorf("input path does not exist: %s", inputPath)
		}
	}

	configInstance.SiteConfig = schemas.SiteConfig{}
	if err := utils.ReadJson(filepath.Join(configInstance.InputPath, ConfigFileName), &configInstance.SiteConfig); err != nil {
		return fmt.Errorf("error reading gitdocai config: %v", err)
	}

	if err := ValidateSiteConfig(&configInstance.SiteConfig); err != nil {
		return fmt.Errorf("site config validation failed: %v", err)
	}

	return nil
}

func GetConfig() *Config {
	return configInstance
}

func ValidateSiteConfig(cfg *schemas.SiteConfig) error {
	if cfg == nil {
		return errors.New("site config is nil")
	}
	if cfg.Name == "" {
		return errors.New("field 'name' is required")
	}
	if cfg.Theme == "" {
		return errors.New("field 'theme' is required")
	}
	if cfg.DefaultThemeMode == "" {
		return errors.New("field 'defaultThemeMode' is required")
	}
	if cfg.Colors.Dark == "" || cfg.Colors.Light == "" {
		return errors.New("both 'colors.dark' and 'colors.light' are required")
	}
	if cfg.Navigation == nil {
		return errors.New("'navigation' is required")
	}

	for i, nav := range cfg.Navbar {
		if nav.Type == "" || nav.Label == "" || nav.Reference == "" {
			return fmt.Errorf("navbar[%d] must include 'type', 'label', and 'reference'", i)
		}
	}

	validFooterTypes := map[string]bool{
		"x": true, "github": true, "linkedin": true, "facebook": true, "youtube": true, "instagram": true,
	}
	for i, link := range cfg.Footer {
		if link.Type == "" || link.Reference == "" {
			return fmt.Errorf("footer[%d] must include 'type' and 'reference'", i)
		}
		if !validFooterTypes[link.Type] {
			return fmt.Errorf("footer[%d] has invalid type '%s'", i, link.Type)
		}
	}

	if err := validateNavigation(cfg.Navigation); err != nil {
		return err
	}

	return nil
}

func validateNavigation(nav *schemas.Navigation) error {
	if len(nav.Versions) > 0 {
		if len(nav.Tabs) > 0 || len(nav.Items) > 0 {
			return errors.New("if 'versions' is present, 'tabs' and 'items' must be empty")
		}
		for i, v := range nav.Versions {
			if v.Version == "" {
				return fmt.Errorf("version[%d] must have 'version' field", i)
			}
			if len(v.Tabs) > 0 && len(v.Items) > 0 {
				return fmt.Errorf("version[%s] cannot have both 'tabs' and 'items'", v.Version)
			}
			for _, tab := range v.Tabs {
				if err := validateTab(tab); err != nil {
					return err
				}
			}
			for _, item := range v.Items {
				if err := validateItem(item); err != nil {
					return err
				}
			}
		}
	} else if len(nav.Tabs) > 0 {
		if len(nav.Items) > 0 {
			return errors.New("'navigation' cannot have both 'tabs' and 'items'")
		}
		for _, tab := range nav.Tabs {
			if err := validateTab(tab); err != nil {
				return err
			}
		}
	} else {
		for _, item := range nav.Items {
			if err := validateItem(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateTab(tab schemas.Tab) error {
	if tab.Tab == "" {
		return errors.New("tab is missing 'tab' field")
	}
	for _, item := range tab.Items {
		if err := validateItem(item); err != nil {
			return err
		}
	}
	return nil
}

func validateItem(item schemas.Item) error {
	if item.Type == "" {
		return errors.New("item is missing 'type'")
	}
	if item.Title == "" {
		return errors.New("item is missing 'title'")
	}

	switch item.Type {
	case "page":
		if item.Page == "" {
			return fmt.Errorf("page item '%s' must have 'page' field", item.Title)
		}
		if len(item.Children) > 0 {
			return fmt.Errorf("page item '%s' cannot have children", item.Title)
		}
	case "group", "dropdown":
		var predominant_type string = ""
		for _, child := range item.Children {
			if child.Type != "page" && child.Type != "openapi" {
				return fmt.Errorf("'%s' item '%s' can only contain children of type 'page' or 'openapi' found '%s'", item.Type, item.Title, child.Type)
			}
			if predominant_type == "" {
				predominant_type = child.Type
			} else if predominant_type != child.Type {
				return fmt.Errorf("%s must have childrens of the same type, founded %s and %s on %s", item.Type, child.Type, predominant_type, item.Title)
			}
			if err := validateItem(child); err != nil {
				return err
			}
		}
	case "openapi":
		if item.ApiReference.Method == "" || item.ApiReference.Reference == "" || item.ApiReference.Path == "" {
			return fmt.Errorf("%s item '%s' must have 'method', 'reference', and 'path' in 'api_reference'", item.Type, item.Title)
		}
		if len(item.Children) > 0 {
			return fmt.Errorf("%s item '%s' cannot have children", item.Type, item.Title)
		}
	default:
		return fmt.Errorf("item '%s' has unknown type '%s'", item.Title, item.Type)
	}

	return nil
}
