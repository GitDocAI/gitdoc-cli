package inkflow

import (
	"encoding/json"
)

// ToJSON converts a Document to JSON string
func (d *Document) ToJSON() (string, error) {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates a Document from JSON string
func FromJSON(jsonStr string) (*Document, error) {
	// First try to unmarshal as-is
	var doc Document
	err := json.Unmarshal([]byte(jsonStr), &doc)
	if err == nil {
		return &doc, nil
	}

	// If that fails, try a more flexible approach
	// Parse into a generic map first to handle different formats
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, err
	}

	// Build the document
	doc.Type = "doc"
	if typ, ok := raw["type"].(string); ok {
		doc.Type = typ
	}

	// Handle content - could be array or object
	if content, ok := raw["content"]; ok {
		switch v := content.(type) {
		case []interface{}:
			// Already an array, convert to Nodes
			doc.Content = make([]Node, len(v))
			for i, item := range v {
				if nodeData, err := json.Marshal(item); err == nil {
					var node Node
					if err := json.Unmarshal(nodeData, &node); err == nil {
						doc.Content[i] = node
					}
				}
			}
		case map[string]interface{}:
			// Object format - convert to array
			// Try to convert object values to nodes
			for _, item := range v {
				if nodeData, err := json.Marshal(item); err == nil {
					var node Node
					if err := json.Unmarshal(nodeData, &node); err == nil {
						doc.Content = append(doc.Content, node)
					}
				}
			}
		}
	}

	return &doc, nil
}

// ParseMDXToJSON is a convenience function to parse MDX and return JSON
func ParseMDXToJSON(mdx string, config *Config) (string, error) {
	parser := NewParser(config)
	doc, err := parser.Parse(mdx)
	if err != nil {
		return "", err
	}
	return doc.ToJSON()
}

// RenderJSONToMDX is a convenience function to parse JSON and render to MDX
func RenderJSONToMDX(jsonStr string, config *Config) (string, error) {
	doc, err := FromJSON(jsonStr)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer(config)
	return renderer.Render(doc)
}
