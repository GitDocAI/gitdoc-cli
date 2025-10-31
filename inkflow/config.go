package inkflow

import (
	"fmt"
	"strings"
)

// DefaultConfig returns a default configuration with common components
func DefaultConfig() *Config {
	return &Config{
		Components:         DefaultComponents(),
		StrictMode:         false,
		PreserveWhitespace: false,
		DefaultLanguage:    "text",
	}
}

// DefaultComponents returns a map of commonly used components
func DefaultComponents() map[string]*ComponentDefinition {
	return map[string]*ComponentDefinition{
		"tip": {
			Name:       "tip",
			MDXTag:     "Tip",
			JSONType:   "tipBlock",
			HasContent: true,
			AttributeMapping: map[string]string{
				"type": "type",
			},
			DefaultAttrs: map[string]interface{}{
				"type": "tip",
			},
		},
		"info": {
			Name:       "info",
			MDXTag:     "Info",
			JSONType:   "infoBlock",
			HasContent: true,
			AttributeMapping: map[string]string{
				"type": "type",
			},
			DefaultAttrs: map[string]interface{}{
				"type": "info",
			},
		},
		"warning": {
			Name:       "warning",
			MDXTag:     "Warning",
			JSONType:   "warningBlock",
			HasContent: true,
			AttributeMapping: map[string]string{
				"type": "type",
			},
			DefaultAttrs: map[string]interface{}{
				"type": "warning",
			},
		},
		"danger": {
			Name:       "danger",
			MDXTag:     "Danger",
			JSONType:   "dangerBlock",
			HasContent: true,
			AttributeMapping: map[string]string{
				"type": "type",
			},
			DefaultAttrs: map[string]interface{}{
				"type": "danger",
			},
		},
		"note": {
			Name:       "note",
			MDXTag:     "Note",
			JSONType:   "noteBlock",
			HasContent: true,
			AttributeMapping: map[string]string{
				"type": "type",
			},
			DefaultAttrs: map[string]interface{}{
				"type": "note",
			},
		},
		"card": {
			Name:             "card",
			MDXTag:           "Card",
			JSONType:         "cardBlock",
			HasContent:       false,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseCard,
			CustomRenderer:   renderCard,
		},
		"accordion": {
			Name:             "accordion",
			MDXTag:           "Accordion",
			JSONType:         "accordionBlock",
			HasContent:       true,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseAccordion,
			CustomRenderer:   renderAccordion,
		},
		"tabs": {
			Name:             "tabs",
			MDXTag:           "Tabs",
			JSONType:         "tabsBlock",
			HasContent:       true,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseTabs,
			CustomRenderer:   renderTabs,
		},
		"table": {
			Name:             "table",
			MDXTag:           "Table",
			JSONType:         "tableBlock",
			HasContent:       false,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseTable,
			CustomRenderer:   renderTable,
		},
		"codeGroup": {
			Name:             "codeGroup",
			MDXTag:           "CodeGroup",
			JSONType:         "codeGroup",
			HasContent:       false,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseCodeGroup,
			CustomRenderer:   renderCodeGroup,
		},
		"codeGroupBlock": {
			Name:             "codeGroupBlock",
			MDXTag:           "CodeGroup",
			JSONType:         "codeGroupBlock",
			HasContent:       false,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseCodeGroup,
			CustomRenderer:   renderCodeGroup,
		},
		"codeBlock": {
			Name:       "codeBlock",
			MDXTag:     "CodeBlock",
			JSONType:   "codeBlockItem",
			HasContent: true,
			AttributeMapping: map[string]string{
				"language": "language",
				"title":    "title",
				"filename": "filename",
			},
			DefaultAttrs: map[string]interface{}{},
		},
		"endpoint": {
			Name:             "endpoint",
			MDXTag:           "Endpoint",
			JSONType:         "endpointBlock",
			HasContent:       false,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
		},
		"label": {
			Name:             "label",
			MDXTag:           "Label",
			JSONType:         "labelBlock",
			HasContent:       false,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
		},
		"columnGroup": {
			Name:             "columnGroup",
			MDXTag:           "ColumnGroup",
			JSONType:         "columnGroup",
			HasContent:       true,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseColumnGroup,
			CustomRenderer:   renderColumnGroup,
		},
		"rightPanel": {
			Name:             "rightPanel",
			MDXTag:           "RightPanel",
			JSONType:         "rightPanel",
			HasContent:       true,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
		},
		"taskList": {
			Name:             "taskList",
			MDXTag:           "taskList",
			JSONType:         "taskList",
			HasContent:       true,
			AttributeMapping: map[string]string{},
			DefaultAttrs:     map[string]interface{}{},
			CustomParser:     parseTaskList,
			CustomRenderer:   renderTaskList,
		},
	}
}

// NewConfig creates a new empty configuration
func NewConfig() *Config {
	return &Config{
		Components:         make(map[string]*ComponentDefinition),
		StrictMode:         false,
		PreserveWhitespace: false,
		DefaultLanguage:    "text",
	}
}

// AddComponent adds a component definition to the configuration
func (c *Config) AddComponent(comp *ComponentDefinition) {
	if c.Components == nil {
		c.Components = make(map[string]*ComponentDefinition)
	}
	c.Components[comp.Name] = comp
}

// RemoveComponent removes a component definition from the configuration
func (c *Config) RemoveComponent(name string) {
	delete(c.Components, name)
}

// GetComponent retrieves a component definition by name
func (c *Config) GetComponent(name string) (*ComponentDefinition, bool) {
	comp, ok := c.Components[name]
	return comp, ok
}

// parseAccordion is a custom parser for accordion components
func parseAccordion(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:    "accordionBlock",
		Attrs:   make(map[string]interface{}),
		Content: []Node{},
	}

	// Parse accordion-level attributes
	if multiple, ok := attrs["multiple"]; ok {
		node.Attrs["multiple"] = (multiple == "true")
	}

	// Parse AccordionTab components
	lines := strings.Split(mdx, "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Look for AccordionTab opening tag
		if strings.HasPrefix(line, "<AccordionTab") {
			// Extract header attribute
			headerStart := strings.Index(line, `header="`)
			header := ""
			if headerStart != -1 {
				headerStart += 8 // len(`header="`)
				headerEnd := strings.Index(line[headerStart:], `"`)
				if headerEnd != -1 {
					header = line[headerStart : headerStart+headerEnd]
				}
			}

			// Check for disabled attribute
			disabled := strings.Contains(line, "disabled={true}")

			// Check for isActive attribute
			isActive := strings.Contains(line, "isActive={true}")

			// Find closing tag
			closingTag := "</AccordionTab>"
			endLine := i + 1
			var contentLines []string

			for endLine < len(lines) {
				if strings.TrimSpace(lines[endLine]) == closingTag {
					break
				}
				contentLines = append(contentLines, lines[endLine])
				endLine++
			}

			// Parse content
			var content []Node
			if len(contentLines) > 0 {
				contentMDX := strings.Join(contentLines, "\n")
				// Remove leading whitespace from each line
				contentMDX = strings.TrimSpace(contentMDX)

				subParser := NewParser(DefaultConfig())
				subDoc, err := subParser.Parse(contentMDX)
				if err == nil {
					content = subDoc.Content
				}
			}

			// Add AccordionTab node
			tabAttrs := map[string]interface{}{
				"header":   header,
				"disabled": disabled,
				"isActive": isActive,
			}

			tabNode := Node{
				Type:    "accordionTab",
				Attrs:   tabAttrs,
				Content: content,
			}
			node.Content = append(node.Content, tabNode)

			i = endLine + 1
		} else {
			i++
		}
	}

	return node, nil
}

// renderAccordion is a custom renderer for accordion components
func renderAccordion(node *Node) (string, error) {
	var result strings.Builder

	// Build opening tag with attributes
	result.WriteString("<Accordion")
	if multiple, ok := node.Attrs["multiple"].(bool); ok && multiple {
		result.WriteString(" multiple={true}")
	}
	result.WriteString(">\n")

	// Render each accordion tab
	for _, tab := range node.Content {
		if tab.Type == "accordionTab" {
			// Get tab attributes
			header := ""
			disabled := false
			isActive := false

			if h, ok := tab.Attrs["header"].(string); ok {
				header = h
			}
			if d, ok := tab.Attrs["disabled"].(bool); ok {
				disabled = d
			}
			if a, ok := tab.Attrs["isActive"].(bool); ok {
				isActive = a
			}

			// Build opening tag
			result.WriteString(fmt.Sprintf("  <AccordionTab header=\"%s\"", header))
			if disabled {
				result.WriteString(" disabled={true}")
			}
			if isActive {
				result.WriteString(" isActive={true}")
			}
			result.WriteString(">\n")

			// Render tab content
			renderer := NewRenderer(DefaultConfig())
			for _, content := range tab.Content {
				mdx, err := renderer.renderNode(&content, 2)
				if err != nil {
					return "", err
				}
				// Indent content
				lines := strings.Split(mdx, "\n")
				for _, line := range lines {
					if line != "" {
						result.WriteString("    " + line + "\n")
					}
				}
			}

			result.WriteString("  </AccordionTab>\n")
		}
	}

	result.WriteString("</Accordion>")
	return result.String(), nil
}

// parseTabs is a custom parser for tabs components
func parseTabs(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:    "tabsBlock",
		Attrs:   make(map[string]interface{}),
		Content: []Node{},
	}

	// Parse tabs-level attributes
	if alignment, ok := attrs["alignment"]; ok {
		node.Attrs["alignment"] = alignment
	}

	// Parse Tab components
	lines := strings.Split(mdx, "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Look for Tab opening tag (not Tabs)
		if strings.HasPrefix(line, "<Tab ") || strings.HasPrefix(line, "<Tab>") {
			// Extract label attribute
			label := ""
			if labelStart := strings.Index(line, `label="`); labelStart != -1 {
				labelStart += 7 // len(`label="`)
				// Find the closing quote for the label value
				if labelEnd := strings.Index(line[labelStart:], `"`); labelEnd != -1 {
					label = line[labelStart : labelStart+labelEnd]
				}
			}

			// Extract icon attribute
			icon := ""
			if iconStart := strings.Index(line, `icon="`); iconStart != -1 {
				iconStart += 6 // len(`icon="`)
				if iconEnd := strings.Index(line[iconStart:], `"`); iconEnd != -1 {
					icon = line[iconStart : iconStart+iconEnd]
				}
			}

			// Check for isActive attribute
			isActive := strings.Contains(line, "isActive={true}")

			// Find closing tag
			endLine := i + 1
			var contentLines []string

			for endLine < len(lines) {
				if strings.Contains(strings.TrimSpace(lines[endLine]), "</Tab>") {
					break
				}
				contentLines = append(contentLines, lines[endLine])
				endLine++
			}

			// Parse content
			var content []Node
			if len(contentLines) > 0 {
				contentMDX := strings.Join(contentLines, "\n")
				contentMDX = strings.TrimSpace(contentMDX)

				subParser := NewParser(DefaultConfig())
				subDoc, err := subParser.Parse(contentMDX)
				if err == nil {
					content = subDoc.Content
				}
			}

			// Add Tab node
			tabAttrs := map[string]interface{}{
				"label":    label,
				"icon":     icon,
				"isActive": isActive,
			}

			tabNode := Node{
				Type:    "tabBlock",
				Attrs:   tabAttrs,
				Content: content,
			}
			node.Content = append(node.Content, tabNode)

			i = endLine + 1
		} else {
			i++
		}
	}

	return node, nil
}

// renderTabs is a custom renderer for tabs components
func renderTabs(node *Node) (string, error) {
	var result strings.Builder

	// Build opening tag with attributes
	result.WriteString("<Tabs")
	if alignment, ok := node.Attrs["alignment"].(string); ok && alignment != "" {
		result.WriteString(fmt.Sprintf(` alignment="%s"`, alignment))
	}
	result.WriteString(">\n")

	// Render each tab
	for _, tab := range node.Content {
		if tab.Type == "tabBlock" {
			// Get tab attributes
			label := ""
			icon := ""
			isActive := false

			if l, ok := tab.Attrs["label"].(string); ok {
				label = l
			}
			if i, ok := tab.Attrs["icon"].(string); ok {
				icon = i
			}
			if a, ok := tab.Attrs["isActive"].(bool); ok {
				isActive = a
			}

			// Build opening tag with attributes
			result.WriteString(fmt.Sprintf("  <Tab label=\"%s\"", label))
			if icon != "" {
				result.WriteString(fmt.Sprintf(` icon="%s"`, icon))
			}
			if isActive {
				result.WriteString(" isActive={true}")
			}
			result.WriteString(">\n")

			// Render tab content
			renderer := NewRenderer(DefaultConfig())
			for _, content := range tab.Content {
				mdx, err := renderer.renderNode(&content, 2)
				if err != nil {
					return "", err
				}
				// Indent content
				lines := strings.Split(mdx, "\n")
				for _, line := range lines {
					if line != "" {
						result.WriteString("    " + line + "\n")
					}
				}
			}

			result.WriteString("  </Tab>\n")
		}
	}

	result.WriteString("</Tabs>")
	return result.String(), nil
}

// parseTable is a custom parser for table components
func parseTable(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:  "tableBlock",
		Attrs: make(map[string]interface{}),
	}

	// Parse table-level attributes from attrs map
	for key, value := range attrs {
		switch key {
		case "scrollable", "pagination":
			node.Attrs[key] = value == "true"
		case "scrollHeight", "rowsPerPage":
			// Convert string to number
			var num float64
			fmt.Sscanf(value, "%f", &num)
			node.Attrs[key] = num
		default:
			node.Attrs[key] = value
		}
	}

	// Parse markdown table from mdx content
	lines := strings.Split(mdx, "\n")
	var columns []map[string]interface{}
	var rows []map[string]interface{}

	headerLine := ""
	dataLines := []string{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check if it's a table line
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}

		// Skip separator line
		if strings.Contains(trimmed, "---") {
			continue
		}

		// First table line is header
		if headerLine == "" {
			headerLine = trimmed
		} else {
			dataLines = append(dataLines, trimmed)
		}

		// Stop at closing tag
		if i < len(lines)-1 && strings.Contains(lines[i+1], "</Table>") {
			break
		}
	}

	// Parse header (columns)
	if headerLine != "" {
		// Remove leading/trailing pipes
		headerLine = strings.Trim(headerLine, "|")
		headers := strings.Split(headerLine, "|")

		for i, header := range headers {
			header = strings.TrimSpace(header)

			// Extract column name and props
			parts := strings.Fields(header)
			columnName := ""
			sortable := false
			filterable := false

			for _, part := range parts {
				if strings.HasPrefix(part, "::") {
					prop := strings.TrimPrefix(part, "::")
					if prop == "sortable" {
						sortable = true
					} else if prop == "filterable" {
						filterable = true
					}
				} else {
					if columnName == "" {
						columnName = part
					} else {
						columnName += " " + part
					}
				}
			}

			column := map[string]interface{}{
				"id":    fmt.Sprintf("col%d", i),
				"label": columnName,
			}

			if sortable {
				column["sortable"] = true
			}
			if filterable {
				column["filterable"] = true
			}

			// Try to derive column ID from label (lowercase, no spaces)
			colID := strings.ToLower(columnName)
			colID = strings.ReplaceAll(colID, " ", "_")
			if colID != "" {
				column["id"] = colID
			}

			columns = append(columns, column)
		}
	}

	// Parse data rows
	for rowIdx, dataLine := range dataLines {
		dataLine = strings.Trim(dataLine, "|")
		cells := strings.Split(dataLine, "|")

		row := map[string]interface{}{
			"id": fmt.Sprintf("row%d", rowIdx+1),
		}

		for i, cell := range cells {
			cell = strings.TrimSpace(cell)
			if i < len(columns) {
				colID := ""
				if id, ok := columns[i]["id"].(string); ok {
					colID = id
				}
				if colID != "" {
					row[colID] = cell
				}
			}
		}

		rows = append(rows, row)
	}

	// Add columns and rows to node
	if len(columns) > 0 {
		node.Attrs["columns"] = columns
	}
	if len(rows) > 0 {
		node.Attrs["rows"] = rows
	}

	return node, nil
}

// renderTable is a custom renderer for table components
func renderTable(node *Node) (string, error) {
	// Tables use JSON attributes, so we need to serialize them
	if len(node.Attrs) == 0 {
		return "<Table />", nil
	}

	var result strings.Builder

	// Extract table-level props (not column/row specific)
	tableLevelProps := make(map[string]interface{})
	for key, value := range node.Attrs {
		if key != "columns" && key != "rows" {
			tableLevelProps[key] = value
		}
	}

	// Build opening tag with table-level props
	result.WriteString("<Table")
	if scrollable, ok := tableLevelProps["scrollable"].(bool); ok {
		result.WriteString(fmt.Sprintf(" scrollable={%v}", scrollable))
	}
	if pagination, ok := tableLevelProps["pagination"].(bool); ok {
		result.WriteString(fmt.Sprintf(" pagination={%v}", pagination))
	}
	if scrollHeight, ok := tableLevelProps["scrollHeight"].(float64); ok {
		result.WriteString(fmt.Sprintf(" scrollHeight={%d}", int(scrollHeight)))
	}
	if rowsPerPage, ok := tableLevelProps["rowsPerPage"].(float64); ok {
		result.WriteString(fmt.Sprintf(" rowsPerPage={%d}", int(rowsPerPage)))
	}
	result.WriteString(">\n\n")

	// Get columns and rows
	var columns []map[string]interface{}
	var rows []map[string]interface{}

	// Try to get columns - could be []interface{} or []map[string]interface{}
	if cols, ok := node.Attrs["columns"].([]interface{}); ok {
		for _, col := range cols {
			if colMap, ok := col.(map[string]interface{}); ok {
				columns = append(columns, colMap)
			}
		}
	} else if cols, ok := node.Attrs["columns"].([]map[string]interface{}); ok {
		columns = cols
	}

	// Try to get rows - could be []interface{} or []map[string]interface{}
	if rowsData, ok := node.Attrs["rows"].([]interface{}); ok {
		for _, row := range rowsData {
			if rowMap, ok := row.(map[string]interface{}); ok {
				rows = append(rows, rowMap)
			}
		}
	} else if rowsData, ok := node.Attrs["rows"].([]map[string]interface{}); ok {
		rows = rowsData
	}

	if len(columns) == 0 || len(rows) == 0 {
		result.WriteString("</Table>")
		return result.String(), nil
	}

	// Build markdown table header with column props
	result.WriteString("| ")
	for i, col := range columns {
		label := ""
		if l, ok := col["label"].(string); ok {
			label = l
		}

		// Add column props using :: syntax
		props := []string{}
		if sortable, ok := col["sortable"].(bool); ok && sortable {
			props = append(props, "::sortable")
		}
		if filterable, ok := col["filterable"].(bool); ok && filterable {
			props = append(props, "::filterable")
		}

		if len(props) > 0 {
			result.WriteString(fmt.Sprintf("%s %s", label, strings.Join(props, " ")))
		} else {
			result.WriteString(label)
		}

		if i < len(columns)-1 {
			result.WriteString(" | ")
		}
	}
	result.WriteString(" |\n")

	// Build separator
	result.WriteString("|")
	for range columns {
		result.WriteString("---------|")
	}
	result.WriteString("\n")

	// Build rows
	for _, row := range rows {
		result.WriteString("| ")
		for i, col := range columns {
			colID := ""
			if id, ok := col["id"].(string); ok {
				colID = id
			}

			// Get cell value
			cellValue := ""
			if val, ok := row[colID]; ok {
				cellValue = fmt.Sprintf("%v", val)
			}

			result.WriteString(cellValue)

			if i < len(columns)-1 {
				result.WriteString(" | ")
			}
		}
		result.WriteString(" |\n")
	}

	result.WriteString("\n</Table>")
	return result.String(), nil
}

// parseCodeGroup is a custom parser for codeGroup components
func parseCodeGroup(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:  "codeGroup",
		Attrs: make(map[string]interface{}),
	}

	var files []map[string]interface{}

	// Parse CodeBlock components within CodeGroup
	lines := strings.Split(mdx, "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Look for CodeBlock opening tag
		if strings.HasPrefix(line, "<CodeBlock") {
			// Extract language attribute
			language := ""
			if langStart := strings.Index(line, `language="`); langStart != -1 {
				langStart += 10 // len(`language="`)
				if langEnd := strings.Index(line[langStart:], `"`); langEnd != -1 {
					language = line[langStart : langStart+langEnd]
				}
			}

			// Extract filename attribute
			filename := ""
			if fnStart := strings.Index(line, `filename="`); fnStart != -1 {
				fnStart += 10 // len(`filename="`)
				if fnEnd := strings.Index(line[fnStart:], `"`); fnEnd != -1 {
					filename = line[fnStart : fnStart+fnEnd]
				}
			}

			// Find closing tag
			closingTag := "</CodeBlock>"
			endLine := i + 1
			var codeLines []string
			inCodeFence := false

			for endLine < len(lines) {
				if strings.TrimSpace(lines[endLine]) == closingTag {
					break
				}

				trimmed := strings.TrimSpace(lines[endLine])

				// Handle code fence
				if strings.HasPrefix(trimmed, "```") {
					if !inCodeFence {
						inCodeFence = true
						endLine++
						continue
					} else {
						// End of code fence
						break
					}
				}

				if inCodeFence {
					codeLines = append(codeLines, lines[endLine])
				}

				endLine++
			}

			// Create file entry
			code := strings.Join(codeLines, "\n")
			file := map[string]interface{}{
				"language": language,
				"code":     code,
			}
			if filename != "" {
				file["filename"] = filename
			}
			files = append(files, file)

			i = endLine + 1
		} else {
			i++
		}
	}

	node.Attrs["files"] = files
	return node, nil
}

// renderCodeGroup is a custom renderer for codeGroup components
func renderCodeGroup(node *Node) (string, error) {
	var result strings.Builder
	result.WriteString("<CodeGroup>\n")

	// Try to get files or tabs from attrs (support both)
	var items []interface{}

	// First try "files" (from visual editor)
	if files, ok := node.Attrs["files"].([]interface{}); ok {
		items = files
	} else if files, ok := node.Attrs["files"].([]map[string]interface{}); ok {
		for _, f := range files {
			items = append(items, f)
		}
	} else if tabs, ok := node.Attrs["tabs"].([]interface{}); ok {
		items = tabs
	} else if tabs, ok := node.Attrs["tabs"].([]map[string]interface{}); ok {
		for _, t := range tabs {
			items = append(items, t)
		}
	}

	// Render each item
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			language := ""
			filename := ""
			code := ""

			if lang, ok := itemMap["language"].(string); ok {
				language = lang
			}
			if fname, ok := itemMap["filename"].(string); ok {
				filename = fname
			}
			if c, ok := itemMap["code"].(string); ok {
				code = c
			}

			// Build CodeBlock tag
			result.WriteString("  <CodeBlock")
			if language != "" {
				result.WriteString(fmt.Sprintf(` language="%s"`, language))
			}
			if filename != "" {
				result.WriteString(fmt.Sprintf(` filename="%s"`, filename))
			}
			result.WriteString(">\n")

			// Write code fence
			if language != "" {
				result.WriteString(fmt.Sprintf("  ```%s\n", language))
			} else {
				result.WriteString("  ```\n")
			}

			// Write code
			codeLines := strings.Split(code, "\n")
			for _, line := range codeLines {
				result.WriteString("  " + line + "\n")
			}

			result.WriteString("  ```\n")
			result.WriteString("  </CodeBlock>\n")
		}
	}

	result.WriteString("</CodeGroup>")
	return result.String(), nil
}

// ComponentBuilder helps build component definitions fluently
type ComponentBuilder struct {
	comp *ComponentDefinition
}

// NewComponentBuilder creates a new component builder
func NewComponentBuilder(name string) *ComponentBuilder {
	return &ComponentBuilder{
		comp: &ComponentDefinition{
			Name:             name,
			MDXTag:           name,
			JSONType:         name + "Block",
			HasContent:       false,
			AttributeMapping: make(map[string]string),
			DefaultAttrs:     make(map[string]interface{}),
		},
	}
}

// WithMDXTag sets the MDX tag name
func (b *ComponentBuilder) WithMDXTag(tag string) *ComponentBuilder {
	b.comp.MDXTag = tag
	return b
}

// WithJSONType sets the JSON type name
func (b *ComponentBuilder) WithJSONType(jsonType string) *ComponentBuilder {
	b.comp.JSONType = jsonType
	return b
}

// WithContent specifies if the component has content
func (b *ComponentBuilder) WithContent(hasContent bool) *ComponentBuilder {
	b.comp.HasContent = hasContent
	return b
}

// WithAttribute adds an attribute mapping
func (b *ComponentBuilder) WithAttribute(mdxAttr, jsonAttr string) *ComponentBuilder {
	b.comp.AttributeMapping[mdxAttr] = jsonAttr
	return b
}

// WithDefaultAttr adds a default attribute
func (b *ComponentBuilder) WithDefaultAttr(key string, value interface{}) *ComponentBuilder {
	b.comp.DefaultAttrs[key] = value
	return b
}

// WithCustomParser sets a custom parser function
func (b *ComponentBuilder) WithCustomParser(parser func(mdx string, attrs map[string]string) (*Node, error)) *ComponentBuilder {
	b.comp.CustomParser = parser
	return b
}

// WithCustomRenderer sets a custom renderer function
func (b *ComponentBuilder) WithCustomRenderer(renderer func(node *Node) (string, error)) *ComponentBuilder {
	b.comp.CustomRenderer = renderer
	return b
}

// Build returns the component definition
func (b *ComponentBuilder) Build() *ComponentDefinition {
	return b.comp
}

// parseCard is a custom parser for card components
func parseCard(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:  "cardBlock",
		Attrs: make(map[string]interface{}),
	}

	// Parse attributes from the opening tag
	if title, ok := attrs["title"]; ok {
		node.Attrs["title"] = title
	}
	if icon, ok := attrs["icon"]; ok {
		node.Attrs["icon"] = icon
	}
	if iconAlign, ok := attrs["iconAlign"]; ok {
		node.Attrs["iconAlign"] = iconAlign
	}
	if href, ok := attrs["href"]; ok {
		node.Attrs["href"] = href
	}
	if content, ok := attrs["content"]; ok {
		node.Attrs["content"] = content
	}

	return node, nil
}

// renderCard is a custom renderer for card components
func renderCard(node *Node) (string, error) {
	var attrs []string

	if title, ok := node.Attrs["title"].(string); ok {
		attrs = append(attrs, fmt.Sprintf(`title="%s"`, title))
	}
	if icon, ok := node.Attrs["icon"].(string); ok {
		attrs = append(attrs, fmt.Sprintf(`icon="%s"`, icon))
	}
	if iconAlign, ok := node.Attrs["iconAlign"].(string); ok {
		attrs = append(attrs, fmt.Sprintf(`iconAlign="%s"`, iconAlign))
	}
	if href, ok := node.Attrs["href"].(string); ok {
		attrs = append(attrs, fmt.Sprintf(`href="%s"`, href))
	}
	if content, ok := node.Attrs["content"].(string); ok {
		attrs = append(attrs, fmt.Sprintf(`content="%s"`, content))
	}

	attrString := ""
	if len(attrs) > 0 {
		attrString = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf("<Card%s />", attrString), nil
}

// parseColumnGroup is a custom parser for columnGroup components
func parseColumnGroup(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:    "columnGroup",
		Attrs:   make(map[string]interface{}),
		Content: []Node{},
	}

	// Parse columnCount attribute
	if columnCount, ok := attrs["columnCount"]; ok {
		var count int
		fmt.Sscanf(columnCount, "%d", &count)
		node.Attrs["columnCount"] = count
	}

	// Parse Column components
	lines := strings.Split(mdx, "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Look for Column opening tag
		if strings.HasPrefix(line, "<Column") {
			// Find closing tag
			closingTag := "</Column>"
			endLine := i + 1
			var contentLines []string

			for endLine < len(lines) {
				if strings.TrimSpace(lines[endLine]) == closingTag {
					break
				}
				contentLines = append(contentLines, lines[endLine])
				endLine++
			}

			// Parse content
			var content []Node
			if len(contentLines) > 0 {
				contentMDX := strings.Join(contentLines, "\n")
				contentMDX = strings.TrimSpace(contentMDX)

				subParser := NewParser(DefaultConfig())
				subDoc, err := subParser.Parse(contentMDX)
				if err == nil {
					content = subDoc.Content
				}
			}

			// Add Column node
			columnNode := Node{
				Type:    "column",
				Content: content,
			}
			node.Content = append(node.Content, columnNode)

			i = endLine + 1
		} else {
			i++
		}
	}

	return node, nil
}

// renderColumnGroup is a custom renderer for columnGroup components
func renderColumnGroup(node *Node) (string, error) {
	var result strings.Builder

	// Build opening tag
	result.WriteString("<ColumnGroup")
	if columnCount, ok := node.Attrs["columnCount"].(int); ok {
		result.WriteString(fmt.Sprintf(` columnCount={%d}`, columnCount))
	} else if columnCount, ok := node.Attrs["columnCount"].(float64); ok {
		result.WriteString(fmt.Sprintf(` columnCount={%d}`, int(columnCount)))
	}
	result.WriteString(">\n")

	// Render each column
	for _, column := range node.Content {
		if column.Type == "column" {
			result.WriteString("  <Column>\n")

			// Render column content
			renderer := NewRenderer(DefaultConfig())
			for _, content := range column.Content {
				mdx, err := renderer.renderNode(&content, 2)
				if err != nil {
					return "", err
				}
				// Indent content
				lines := strings.Split(mdx, "\n")
				for _, line := range lines {
					if line != "" {
						result.WriteString("    " + line + "\n")
					}
				}
			}

			result.WriteString("  </Column>\n")
		}
	}

	result.WriteString("</ColumnGroup>")
	return result.String(), nil
}

// parseTaskList is a custom parser for taskList components
func parseTaskList(mdx string, attrs map[string]string) (*Node, error) {
	node := &Node{
		Type:    "taskList",
		Attrs:   make(map[string]interface{}),
		Content: []Node{},
	}

	// Parse taskItem components
	lines := strings.Split(mdx, "\n")
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Look for taskItem opening tag (case-insensitive)
		if strings.HasPrefix(strings.ToLower(trimmed), "<taskitem") {
			// Extract checked attribute
			checked := false
			if checkedStart := strings.Index(trimmed, `checked="`); checkedStart != -1 {
				checkedStart += 9 // len(`checked="`)
				checkedEnd := strings.Index(trimmed[checkedStart:], `"`)
				if checkedEnd != -1 {
					checkedValue := trimmed[checkedStart : checkedStart+checkedEnd]
					checked = (checkedValue == "true")
				}
			}

			// Check if it's a single-line taskItem (opening and closing on same line)
			closingTag := "</taskItem>"
			if strings.Contains(strings.ToLower(trimmed), strings.ToLower(closingTag)) {
				// Single-line taskItem - extract content between tags
				openTagEnd := strings.Index(trimmed, ">")
				closeTagStart := strings.Index(strings.ToLower(trimmed), strings.ToLower(closingTag))

				var contentText string
				if openTagEnd != -1 && closeTagStart != -1 && openTagEnd < closeTagStart {
					contentText = strings.TrimSpace(trimmed[openTagEnd+1 : closeTagStart])
				}

				// Create paragraph with the text
				var content []Node
				if contentText != "" {
					content = []Node{
						{
							Type: "paragraph",
							Content: []Node{
								{
									Type: "text",
									Text: contentText,
								},
							},
						},
					}
				}

				// Add taskItem node
				taskItemNode := Node{
					Type: "taskItem",
					Attrs: map[string]interface{}{
						"checked": checked,
					},
					Content: content,
				}
				node.Content = append(node.Content, taskItemNode)
				i++
				continue
			}

			// Multi-line taskItem - find closing tag
			endLine := i + 1
			var contentLines []string

			for endLine < len(lines) {
				if strings.Contains(strings.ToLower(strings.TrimSpace(lines[endLine])), strings.ToLower(closingTag)) {
					break
				}
				contentLines = append(contentLines, lines[endLine])
				endLine++
			}

			// Parse content - the text inside taskItem becomes a paragraph
			var content []Node
			if len(contentLines) > 0 {
				contentMDX := strings.Join(contentLines, "\n")
				contentMDX = strings.TrimSpace(contentMDX)

				// Create paragraph with the text
				if contentMDX != "" {
					content = []Node{
						{
							Type: "paragraph",
							Content: []Node{
								{
									Type: "text",
									Text: contentMDX,
								},
							},
						},
					}
				}
			}

			// Add taskItem node
			taskItemNode := Node{
				Type: "taskItem",
				Attrs: map[string]interface{}{
					"checked": checked,
				},
				Content: content,
			}
			node.Content = append(node.Content, taskItemNode)

			i = endLine + 1
		} else {
			i++
		}
	}

	return node, nil
}

// renderTaskList is a custom renderer for taskList components
func renderTaskList(node *Node) (string, error) {
	var result strings.Builder

	result.WriteString("<taskList>\n")

	// Render each task item
	for _, item := range node.Content {
		if item.Type == "taskItem" {
			// Get checked attribute
			checked := false
			if c, ok := item.Attrs["checked"].(bool); ok {
				checked = c
			}

			// Build opening tag
			result.WriteString(fmt.Sprintf(`  <taskItem checked="%v">`, checked))

			// Render content (extract text from paragraph)
			if len(item.Content) > 0 {
				for _, content := range item.Content {
					if content.Type == "paragraph" && len(content.Content) > 0 {
						for _, textNode := range content.Content {
							if textNode.Type == "text" {
								result.WriteString(" " + textNode.Text + " ")
							}
						}
					}
				}
			}

			result.WriteString("</taskItem>\n")
		}
	}

	result.WriteString("</taskList>")
	return result.String(), nil
}

// Example: Creating a custom component
func ExampleCustomComponent() *ComponentDefinition {
	return NewComponentBuilder("myComponent").
		WithMDXTag("MyComponent").
		WithJSONType("myComponentBlock").
		WithContent(true).
		WithAttribute("title", "title").
		WithAttribute("color", "color").
		WithDefaultAttr("color", "blue").
		WithCustomParser(func(mdx string, attrs map[string]string) (*Node, error) {
			// Custom parsing logic
			return &Node{
				Type: "myComponentBlock",
				Attrs: map[string]interface{}{
					"title": attrs["title"],
					"color": attrs["color"],
				},
			}, nil
		}).
		WithCustomRenderer(func(node *Node) (string, error) {
			title := ""
			if t, ok := node.Attrs["title"].(string); ok {
				title = t
			}
			return fmt.Sprintf("<MyComponent title=\"%s\" />", title), nil
		}).
		Build()
}
