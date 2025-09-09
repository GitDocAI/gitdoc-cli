package utils

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// ExtractFrontMatter extracts the content within the front matter delimiters (e.g., ---)
// and parses it into a map.
func ExtractFrontMatter(content string) (map[string]interface{}, bool) {
	// Regular expression to match front matter delimiters.
	reDelim := regexp.MustCompile(`(?m)^-{3,}\s*$`)
	positions := reDelim.FindAllStringIndex(content, 2)

	if len(positions) < 2 {
		return nil, false
	}

	start := positions[0][1]
	end := positions[1][0]
	frontMatter := content[start:end]

	// Remove any unnecessary whitespace.
	trimmed := strings.TrimSpace(frontMatter)

	// Parse the YAML content.
	parsed := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, false
	}

	return parsed, true
}

