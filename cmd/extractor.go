package cmd

import (
	"strings"

	"github.com/YHVCorp/DockitAICLI/schemas"
)

func ExtractUsedPaths(config schemas.SiteConfig) ([]string, []string) {
	publicSet := make(map[string]struct{})
	contentSet := make(map[string]struct{})

	extractAndAddPath := func(path string, target map[string]struct{}) {
		if path == "" || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			return
		}

		normalized := strings.TrimPrefix(path, "./")
		normalized = strings.TrimPrefix(normalized, "/")

		parts := strings.Split(normalized, "/")
		if len(parts) == 0 || parts[0] == "" {
			return
		}

		target[parts[0]] = struct{}{}
	}

	extractAndAddPath(config.Logo.Light, publicSet)
	extractAndAddPath(config.Logo.Dark, publicSet)
	extractAndAddPath(config.Favicon, publicSet)
	extractAndAddPath(config.Background.Image.Light, publicSet)
	extractAndAddPath(config.Background.Image.Dark, publicSet)

	var extractPagesFromItems func([]schemas.Item)

	extractPagesFromItems = func(items []schemas.Item) {
		for i := range items {
			if items[i].Type == "page" && items[i].Page != "" {
				extractAndAddPath(items[i].Page, contentSet)
			}
			if len(items[i].Children) > 0 {
				extractPagesFromItems(items[i].Children)
			}
		}
	}

	if config.Navigation != nil {
		for _, version := range config.Navigation.Versions {
			for _, tab := range version.Tabs {
				extractPagesFromItems(tab.Items)
			}
			extractPagesFromItems(version.Items)
		}
		for _, tab := range config.Navigation.Tabs {
			extractPagesFromItems(tab.Items)
		}
		extractPagesFromItems(config.Navigation.Items)
	}

	publicList := make([]string, 0, len(publicSet))
	for path := range publicSet {
		publicList = append(publicList, path)
	}

	contentList := make([]string, 0, len(contentSet))
	for path := range contentSet {
		contentList = append(contentList, path)
	}

	return publicList, contentList
}
