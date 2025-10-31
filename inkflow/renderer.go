package inkflow

import (
	"fmt"
	"strings"
)

// NewRenderer creates a new JSON to MDX renderer with the given configuration
func NewRenderer(config *Config) *Renderer {
	if config == nil {
		config = DefaultConfig()
	}
	return &Renderer{config: config}
}

// Render converts a JSON document structure to MDX string
func (r *Renderer) Render(doc *Document) (string, error) {
	var builder strings.Builder

	for i, node := range doc.Content {
		mdx, err := r.renderNode(&node, 0)
		if err != nil {
			return "", err
		}

		builder.WriteString(mdx)

		// Add spacing between blocks (except for last element)
		if i < len(doc.Content)-1 {
			builder.WriteString("\n\n")
		}
	}

	return builder.String(), nil
}

// renderNode renders a single node to MDX
func (r *Renderer) renderNode(node *Node, indent int) (string, error) {
	switch node.Type {
	case "heading":
		return r.renderHeading(node)
	case "paragraph":
		return r.renderParagraph(node, indent)
	case "codeBlock":
		return r.renderCodeBlock(node)
	case "bulletList":
		return r.renderBulletList(node, indent)
	case "orderedList":
		return r.renderOrderedList(node, indent)
	case "imageBlock":
		return r.renderImageBlock(node)
	case "horizontalRule":
		return "---", nil
	case "endpointBlock":
		return r.renderEndpointBlock(node)
	case "labelBlock":
		return r.renderLabelBlock(node)
	case "rightPanel":
		return r.renderRightPanel(node, indent)
	case "blockquote":
		return r.renderBlockquote(node, indent)
	default:
		// Check if it's a custom component
		return r.renderComponent(node, indent)
	}
}

// renderHeading renders a heading node
func (r *Renderer) renderHeading(node *Node) (string, error) {
	level := 1
	if l, ok := node.Attrs["level"].(float64); ok {
		level = int(l)
	} else if l, ok := node.Attrs["level"].(int); ok {
		level = l
	}

	prefix := strings.Repeat("#", level)
	text := r.renderInlineContent(node.Content)

	return fmt.Sprintf("%s %s", prefix, text), nil
}

// renderParagraph renders a paragraph node
func (r *Renderer) renderParagraph(node *Node, _ int) (string, error) {
	text := r.renderInlineContent(node.Content)
	return text, nil
}

// renderCodeBlock renders a code block
func (r *Renderer) renderCodeBlock(node *Node) (string, error) {
	language := ""
	if lang, ok := node.Attrs["language"].(string); ok {
		language = lang
	}

	var code string
	if len(node.Content) > 0 && node.Content[0].Type == "text" {
		code = node.Content[0].Text
	}

	return fmt.Sprintf("```%s\n%s\n```", language, code), nil
}

// renderBulletList renders an unordered list
func (r *Renderer) renderBulletList(node *Node, indent int) (string, error) {
	var builder strings.Builder
	indentStr := strings.Repeat("  ", indent)

	for i, item := range node.Content {
		if item.Type != "listItem" {
			continue
		}

		// Render list item content
		itemText := r.renderListItemContent(&item)
		builder.WriteString(fmt.Sprintf("%s- %s", indentStr, itemText))

		if i < len(node.Content)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// renderOrderedList renders an ordered list
func (r *Renderer) renderOrderedList(node *Node, indent int) (string, error) {
	var builder strings.Builder
	indentStr := strings.Repeat("  ", indent)

	for i, item := range node.Content {
		if item.Type != "listItem" {
			continue
		}

		// Render list item content
		itemText := r.renderListItemContent(&item)
		builder.WriteString(fmt.Sprintf("%s%d. %s", indentStr, i+1, itemText))

		if i < len(node.Content)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// renderListItemContent renders the content of a list item
func (r *Renderer) renderListItemContent(item *Node) string {
	var parts []string

	for _, content := range item.Content {
		if content.Type == "paragraph" {
			parts = append(parts, r.renderInlineContent(content.Content))
		} else {
			// For nested structures, we would recursively render
			text, _ := r.renderNode(&content, 0)
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " ")
}

// renderImageBlock renders an image
func (r *Renderer) renderImageBlock(node *Node) (string, error) {
	src := ""
	alt := ""

	if s, ok := node.Attrs["src"].(string); ok {
		src = s
	}
	if a, ok := node.Attrs["alt"].(string); ok {
		alt = a
	}

	return fmt.Sprintf("![%s](%s)", alt, src), nil
}

// renderComponent renders a custom component
func (r *Renderer) renderComponent(node *Node, indent int) (string, error) {
	// Find component definition by JSON type
	var componentDef *ComponentDefinition
	for _, comp := range r.config.Components {
		if comp.JSONType == node.Type {
			componentDef = comp
			break
		}
	}

	// Use custom renderer if available
	if componentDef != nil && componentDef.CustomRenderer != nil {
		return componentDef.CustomRenderer(node)
	}

	// Default rendering
	tagName := node.Type
	if componentDef != nil {
		tagName = componentDef.MDXTag
	}

	// Build attributes string
	var attrs []string
	if node.Attrs != nil {
		for key, value := range node.Attrs {
			// Skip nil values
			if value == nil {
				continue
			}

			// Reverse attribute mapping if component definition exists
			attrName := key
			if componentDef != nil {
				for mdxAttr, jsonAttr := range componentDef.AttributeMapping {
					if jsonAttr == key {
						attrName = mdxAttr
						break
					}
				}
			}

			// Skip default attributes
			if componentDef != nil {
				if defaultVal, ok := componentDef.DefaultAttrs[key]; ok && defaultVal == value {
					continue
				}
			}

			attrs = append(attrs, fmt.Sprintf("%s=\"%v\"", attrName, value))
		}
	}

	attrString := ""
	if len(attrs) > 0 {
		attrString = " " + strings.Join(attrs, " ")
	}

	// Check if component has content
	if len(node.Content) == 0 {
		// Self-closing tag
		return fmt.Sprintf("<%s%s />", tagName, attrString), nil
	}

	// Component with content
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<%s%s>\n", tagName, attrString))

	// Render content
	for _, child := range node.Content {
		childMDX, err := r.renderNode(&child, indent+1)
		if err != nil {
			return "", err
		}
		builder.WriteString(childMDX)
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("</%s>", tagName))

	return builder.String(), nil
}

// renderInlineContent renders inline content with formatting
func (r *Renderer) renderInlineContent(nodes []Node) string {
	var parts []string

	for _, node := range nodes {
		if node.Type != "text" {
			continue
		}

		text := node.Text

		// Apply marks
		for _, mark := range node.Marks {
			switch mark.Type {
			case "bold":
				text = fmt.Sprintf("**%s**", text)
			case "italic":
				text = fmt.Sprintf("*%s*", text)
			case "code":
				text = fmt.Sprintf("`%s`", text)
			case "link":
				href := ""
				if mark.Attrs != nil {
					if h, ok := mark.Attrs["href"].(string); ok {
						href = h
					}
				}
				text = fmt.Sprintf("[%s](%s)", text, href)
			case "strikethrough":
				text = fmt.Sprintf("~~%s~~", text)
			}
		}

		parts = append(parts, text)
	}

	return strings.Join(parts, "")
}

// renderEndpointBlock renders an endpoint badge
func (r *Renderer) renderEndpointBlock(node *Node) (string, error) {
	method := ""
	path := ""

	if m, ok := node.Attrs["method"].(string); ok {
		method = m
	}
	if p, ok := node.Attrs["path"].(string); ok {
		path = p
	}

	return fmt.Sprintf("<Endpoint method=\"%s\" path=\"%s\" />", method, path), nil
}

// renderLabelBlock renders a label/badge
func (r *Renderer) renderLabelBlock(node *Node) (string, error) {
	label := ""
	color := ""
	size := ""

	if l, ok := node.Attrs["label"].(string); ok {
		label = l
	}
	if c, ok := node.Attrs["color"].(string); ok {
		color = c
	}
	if s, ok := node.Attrs["size"].(string); ok {
		size = s
	}

	var attrs []string
	if label != "" {
		attrs = append(attrs, fmt.Sprintf(`label="%s"`, label))
	}
	if color != "" {
		attrs = append(attrs, fmt.Sprintf(`color="%s"`, color))
	}
	if size != "" {
		attrs = append(attrs, fmt.Sprintf(`size="%s"`, size))
	}

	attrString := ""
	if len(attrs) > 0 {
		attrString = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf("<Label%s />", attrString), nil
}

// renderRightPanel renders a right panel component
func (r *Renderer) renderRightPanel(node *Node, indent int) (string, error) {
	var builder strings.Builder
	builder.WriteString("<RightPanel>\n")

	// Render content
	for _, child := range node.Content {
		childMDX, err := r.renderNode(&child, indent+1)
		if err != nil {
			return "", err
		}
		builder.WriteString(childMDX)
		builder.WriteString("\n")
	}

	builder.WriteString("</RightPanel>")
	return builder.String(), nil
}

// renderBlockquote renders a blockquote
func (r *Renderer) renderBlockquote(node *Node, indent int) (string, error) {
	var builder strings.Builder

	for i, child := range node.Content {
		childMDX, err := r.renderNode(&child, indent)
		if err != nil {
			return "", err
		}

		// Add > prefix to each line
		lines := strings.Split(childMDX, "\n")
		for _, line := range lines {
			if line != "" {
				builder.WriteString("> ")
				builder.WriteString(line)
				builder.WriteString("\n")
			}
		}

		if i < len(node.Content)-1 {
			builder.WriteString(">\n")
		}
	}

	return strings.TrimRight(builder.String(), "\n"), nil
}
