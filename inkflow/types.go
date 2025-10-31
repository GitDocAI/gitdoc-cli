package inkflow

// Document represents the root document structure
type Document struct {
	Type    string `json:"type"`
	Content []Node `json:"content"`
}

// Node represents a generic content node
type Node struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []Node                 `json:"content,omitempty"`
	Marks   []Mark                 `json:"marks,omitempty"`
	Text    string                 `json:"text,omitempty"`
}

// Mark represents text formatting marks
type Mark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
}

// ComponentDefinition defines how to parse a custom component
type ComponentDefinition struct {
	// Name of the component (e.g., "Tip", "Info", "Card")
	Name string

	// MDXTag is the tag name used in MDX (e.g., "Tip", "Info")
	MDXTag string

	// JSONType is the type used in JSON (e.g., "tipBlock", "infoBlock")
	JSONType string

	// HasContent indicates if the component has nested content
	HasContent bool

	// AttributeMapping maps MDX attributes to JSON attrs
	// Key: MDX attribute name, Value: JSON attribute name
	AttributeMapping map[string]string

	// DefaultAttrs provides default attributes for the component
	DefaultAttrs map[string]interface{}

	// CustomParser allows custom parsing logic for complex components
	CustomParser func(mdx string, attrs map[string]string) (*Node, error)

	// CustomRenderer allows custom rendering logic when converting to MDX
	CustomRenderer func(node *Node) (string, error)
}

// Config holds the parser configuration
type Config struct {
	// Components is a map of component definitions
	// Key: component name/identifier
	Components map[string]*ComponentDefinition

	// StrictMode determines if unknown components should cause errors
	StrictMode bool

	// PreserveWhitespace controls whitespace handling
	PreserveWhitespace bool

	// DefaultLanguage for code blocks
	DefaultLanguage string
}

// Parser handles MDX to JSON conversion
type Parser struct {
	config *Config
}

// Renderer handles JSON to MDX conversion
type Renderer struct {
	config *Config
}
