package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InjectApiReference generates a React component <ApiReference />
// using OpenAPI data (as map[string]any) and the frontmatter.
func InjectApiReference(openapi map[string]any, frontmatter map[string]any) string {
	// frontmatter["openapi"] -> "get /api/users/{id}"
	splits := strings.SplitN(frontmatter["openapi"].(string), " ", 2)
	if len(splits) != 2 {
		return "<!-- invalid openapi frontmatter -->"
	}
	method := strings.ToLower(splits[0])
	path := splits[1]

	// Navigate: openapi["paths"][path][method]
	paths := openapi["paths"].(map[string]any)
	pathItem, ok := paths[path].(map[string]any)
	if !ok {
		return "<!-- path not found -->"
	}

	op, ok := pathItem[method].(map[string]any)
	if !ok {
		return "<!-- method not found -->"
	}

	title := ""
	if sum, ok := op["summary"].(string); ok {
		title = sum
	}
	description := ""
	if desc, ok := op["description"].(string); ok {
		description = desc
	}

	parameters := op["parameters"]
	responses := op["responses"]

	component := fmt.Sprintf(`<ApiReference
  title="%s"
  method="%s"
  path="%s"
  description="%s"
  parameters={%s}
  responses={%s}
/>`,
		title,
		strings.ToUpper(method),
		path,
		description,
		serializeToJS(parameters),
		serializeToJS(responses),
	)

	return component
}

func serializeToJS(data interface{}) string {
	if data == nil {
		return "null"
	}
	jsonData, err := json.MarshalIndent(data, "    ", "  ")
	if err != nil {
		return "{}"
	}

	str := string(jsonData)
	str = strings.ReplaceAll(str, "\"", "")
	str = strings.ReplaceAll(str, "\n", "\n    ")
	return str
}

