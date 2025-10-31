package inkflow

import (
	"fmt"
	"regexp"
	"strings"
)

// NewParser creates a new MDX parser with the given configuration
func NewParser(config *Config) *Parser {
	if config == nil {
		config = DefaultConfig()
	}
	return &Parser{config: config}
}

// Parse converts MDX string to JSON document structure
func (p *Parser) Parse(mdx string) (*Document, error) {
	doc := &Document{
		Type:    "doc",
		Content: []Node{},
	}

	lines := strings.Split(mdx, "\n")
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Check for custom JSX components
		if strings.HasPrefix(trimmed, "<") && len(trimmed) > 1 {
			// Try uppercase first (e.g., <Tip>, <Card>)
			tagMatch := regexp.MustCompile(`^<([A-Z][a-zA-Z0-9]*)`).FindStringSubmatch(trimmed)
			if len(tagMatch) > 0 {
				tagName := tagMatch[1]
				var componentDef *ComponentDefinition
				for _, comp := range p.config.Components {
					if comp.MDXTag == tagName {
						componentDef = comp
						break
					}
				}

				if componentDef != nil {
					node, consumed, err := p.parseComponent(lines, i)
					if err != nil {
						return nil, fmt.Errorf("error parsing component at line %d: %w", i+1, err)
					}
					if node != nil {
						doc.Content = append(doc.Content, *node)
					}
					i += consumed
					continue
				}
			}

			// Try lowercase (e.g., <taskList>, <taskItem>)
			tagMatchLower := regexp.MustCompile(`^<([a-z][a-zA-Z0-9]*)`).FindStringSubmatch(trimmed)
			if len(tagMatchLower) > 0 {
				tagName := tagMatchLower[1]
				var componentDef *ComponentDefinition
				for _, comp := range p.config.Components {
					// Case-insensitive comparison
					if strings.EqualFold(comp.MDXTag, tagName) {
						componentDef = comp
						break
					}
				}

				if componentDef != nil {
					node, consumed, err := p.parseComponent(lines, i)
					if err != nil {
						return nil, fmt.Errorf("error parsing component at line %d: %w", i+1, err)
					}
					if node != nil {
						doc.Content = append(doc.Content, *node)
					}
					i += consumed
					continue
				}
			}
		}

		// Parse markdown elements
		node, consumed, err := p.parseMarkdownElement(lines, i)
		if err != nil {
			return nil, fmt.Errorf("error parsing markdown at line %d: %w", i+1, err)
		}

		if node != nil {
			doc.Content = append(doc.Content, *node)
		}

		i += consumed
	}

	return doc, nil
}

// parseMarkdownElement parses a single markdown element starting at the given line
func (p *Parser) parseMarkdownElement(lines []string, start int) (*Node, int, error) {
	line := lines[start]
	trimmed := strings.TrimSpace(line)

	// Heading
	if strings.HasPrefix(trimmed, "#") {
		return p.parseHeading(line)
	}

	// Horizontal rule
	if regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`).MatchString(trimmed) {
		return &Node{Type: "horizontalRule"}, 1, nil
	}

	// Fenced code block
	if strings.HasPrefix(trimmed, "```") {
		return p.parseCodeBlock(lines, start)
	}

	// Blockquote
	if strings.HasPrefix(trimmed, ">") {
		return p.parseBlockquote(lines, start)
	}

	// Table (check if this line and next lines form a table)
	if strings.HasPrefix(trimmed, "|") && start+1 < len(lines) {
		nextLine := strings.TrimSpace(lines[start+1])
		if strings.Contains(nextLine, "---") && strings.HasPrefix(nextLine, "|") {
			return p.parseTable(lines, start)
		}
	}

	// Bullet list
	if regexp.MustCompile(`^[\*\-\+]\s+`).MatchString(trimmed) {
		return p.parseList(lines, start, false)
	}

	// Ordered list
	if regexp.MustCompile(`^\d+\.\s+`).MatchString(trimmed) {
		return p.parseList(lines, start, true)
	}

	// HTML Image tag
	if strings.HasPrefix(trimmed, "<img") {
		return p.parseHTMLImage(line)
	}

	// Default: paragraph
	return p.parseParagraph(lines, start)
}

// parseHeading parses a heading line
func (p *Parser) parseHeading(line string) (*Node, int, error) {
	trimmed := strings.TrimSpace(line)
	level := 0
	for i := 0; i < len(trimmed) && trimmed[i] == '#'; i++ {
		level++
	}

	if level == 0 || level > 6 {
		return nil, 0, fmt.Errorf("invalid heading level")
	}

	text := strings.TrimSpace(trimmed[level:])
	content := p.parseInlineContent(text)

	return &Node{
		Type: "heading",
		Attrs: map[string]interface{}{
			"level": level,
		},
		Content: content,
	}, 1, nil
}

// parseParagraph parses a paragraph (one or more consecutive non-empty lines)
func (p *Parser) parseParagraph(lines []string, start int) (*Node, int, error) {
	var paragraphLines []string
	i := start

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Stop at empty line
		if trimmed == "" {
			break
		}

		// Stop at special markdown elements
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "```") ||
			strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "|") ||
			regexp.MustCompile(`^[\*\-\+]\s+`).MatchString(trimmed) ||
			regexp.MustCompile(`^\d+\.\s+`).MatchString(trimmed) ||
			regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`).MatchString(trimmed) {
			break
		}

		// Stop at JSX components
		if strings.HasPrefix(trimmed, "<") && len(trimmed) > 1 {
			tagMatch := regexp.MustCompile(`^<([A-Z][a-zA-Z0-9]*)`).FindStringSubmatch(trimmed)
			if len(tagMatch) > 0 {
				break
			}
		}

		paragraphLines = append(paragraphLines, trimmed)
		i++
	}

	if len(paragraphLines) == 0 {
		return nil, 1, nil
	}

	text := strings.Join(paragraphLines, " ")
	content := p.parseInlineContent(text)

	return &Node{
		Type:    "paragraph",
		Content: content,
	}, i - start, nil
}

// parseInlineContent parses inline formatting (bold, italic, code, links, images)
func (p *Parser) parseInlineContent(text string) []Node {
	var nodes []Node
	i := 0

	for i < len(text) {
		// Track if any special parser consumed characters
		consumed := false

		// Check for image: ![alt](url)
		if i < len(text)-1 && text[i] == '!' && text[i+1] == '[' {
			if node, consumedChars := p.parseImage(text[i:]); node != nil && consumedChars > 0 {
				nodes = append(nodes, *node)
				i += consumedChars
				consumed = true
			}
		}

		// Check for link: [text](url)
		if !consumed && text[i] == '[' {
			if node, consumedChars := p.parseLink(text[i:]); node != nil && consumedChars > 0 {
				nodes = append(nodes, node...)
				i += consumedChars
				consumed = true
			}
		}

		// Check for inline code: `code`
		if !consumed && text[i] == '`' {
			if node, consumedChars := p.parseInlineCode(text[i:]); node != nil && consumedChars > 0 {
				nodes = append(nodes, *node)
				i += consumedChars
				consumed = true
			}
		}

		// Check for bold: **text** or __text__
		if !consumed && i < len(text)-1 && (text[i:i+2] == "**" || text[i:i+2] == "__") {
			if node, consumedChars := p.parseBold(text[i:]); node != nil && consumedChars > 0 {
				nodes = append(nodes, node...)
				i += consumedChars
				consumed = true
			}
		}

		// Check for italic: *text* or _text_
		if !consumed && (text[i] == '*' || text[i] == '_') {
			if node, consumedChars := p.parseItalic(text[i:]); node != nil && consumedChars > 0 {
				nodes = append(nodes, node...)
				i += consumedChars
				consumed = true
			}
		}

		// Check for strikethrough: ~~text~~
		if !consumed && i < len(text)-1 && text[i:i+2] == "~~" {
			if node, consumedChars := p.parseStrikethrough(text[i:]); node != nil && consumedChars > 0 {
				nodes = append(nodes, node...)
				i += consumedChars
				consumed = true
			}
		}

		// If special character wasn't consumed, or regular text, consume character by character
		if consumed {
			continue
		}

		// Regular text - consume until next special character
		textStart := i
		for i < len(text) && text[i] != '!' && text[i] != '[' && text[i] != '`' && text[i] != '*' && text[i] != '_' && text[i] != '~' {
			i++
		}

		// If we consumed regular text, add it to nodes
		if i > textStart {
			nodes = append(nodes, Node{
				Type: "text",
				Text: text[textStart:i],
			})
		} else {
			// We're at a special character that couldn't be parsed
			// Treat it as literal text and advance by 1 to avoid infinite loop
			if i < len(text) {
				nodes = append(nodes, Node{
					Type: "text",
					Text: string(text[i]),
				})
				i++
			}
		}
	}

	// If no nodes were created, create a single text node
	if len(nodes) == 0 && text != "" {
		nodes = append(nodes, Node{
			Type: "text",
			Text: text,
		})
	}

	return nodes
}

// parseImage parses an image: ![alt](url)
func (p *Parser) parseImage(text string) (*Node, int) {
	re := regexp.MustCompile(`^!\[([^\]]*)\]\(([^\)]+)\)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 3 {
		return nil, 0
	}

	return &Node{
		Type: "imageBlock",
		Attrs: map[string]interface{}{
			"src": matches[2],
			"alt": matches[1],
		},
	}, len(matches[0])
}

// parseLink parses a link: [text](url)
func (p *Parser) parseLink(text string) ([]Node, int) {
	re := regexp.MustCompile(`^\[([^\]]+)\]\(([^\)]+)\)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 3 {
		return nil, 0
	}

	linkText := matches[1]
	linkURL := matches[2]

	// Parse inline content within the link text
	linkContent := p.parseInlineContent(linkText)

	// Apply link mark to all text nodes
	for i := range linkContent {
		if linkContent[i].Type == "text" {
			linkContent[i].Marks = append(linkContent[i].Marks, Mark{
				Type: "link",
				Attrs: map[string]interface{}{
					"href":   linkURL,
					"target": "_self",
				},
			})
		}
	}

	return linkContent, len(matches[0])
}

// parseInlineCode parses inline code: `code`
func (p *Parser) parseInlineCode(text string) (*Node, int) {
	if !strings.HasPrefix(text, "`") {
		return nil, 0
	}

	end := strings.Index(text[1:], "`")
	if end == -1 {
		return nil, 0
	}

	code := text[1 : end+1]
	return &Node{
		Type:  "text",
		Text:  code,
		Marks: []Mark{{Type: "code"}},
	}, end + 2
}

// parseBold parses bold text: **text** or __text__
func (p *Parser) parseBold(text string) ([]Node, int) {
	delimiter := ""
	if strings.HasPrefix(text, "**") {
		delimiter = "**"
	} else if strings.HasPrefix(text, "__") {
		delimiter = "__"
	} else {
		return nil, 0
	}

	end := strings.Index(text[2:], delimiter)
	if end == -1 {
		return nil, 0
	}

	boldText := text[2 : end+2]
	content := p.parseInlineContent(boldText)

	// Apply bold mark to all text nodes
	for i := range content {
		if content[i].Type == "text" {
			content[i].Marks = append(content[i].Marks, Mark{Type: "bold"})
		}
	}

	return content, end + 4
}

// parseItalic parses italic text: *text* or _text_
func (p *Parser) parseItalic(text string) ([]Node, int) {
	if len(text) < 2 {
		return nil, 0
	}

	delimiter := string(text[0])
	if delimiter != "*" && delimiter != "_" {
		return nil, 0
	}

	// Make sure it's not bold
	if len(text) > 1 && text[1:2] == delimiter {
		return nil, 0
	}

	end := strings.Index(text[1:], delimiter)
	if end == -1 {
		return nil, 0
	}

	italicText := text[1 : end+1]
	content := p.parseInlineContent(italicText)

	// Apply italic mark to all text nodes
	for i := range content {
		if content[i].Type == "text" {
			content[i].Marks = append(content[i].Marks, Mark{Type: "italic"})
		}
	}

	return content, end + 2
}

// parseStrikethrough parses strikethrough text: ~~text~~
func (p *Parser) parseStrikethrough(text string) ([]Node, int) {
	if !strings.HasPrefix(text, "~~") {
		return nil, 0
	}

	end := strings.Index(text[2:], "~~")
	if end == -1 {
		return nil, 0
	}

	strikeText := text[2 : end+2]
	content := p.parseInlineContent(strikeText)

	// Apply strikethrough mark to all text nodes
	for i := range content {
		if content[i].Type == "text" {
			content[i].Marks = append(content[i].Marks, Mark{Type: "strikethrough"})
		}
	}

	return content, end + 4
}

// parseCodeBlock parses a fenced code block
func (p *Parser) parseCodeBlock(lines []string, start int) (*Node, int, error) {
	firstLine := strings.TrimSpace(lines[start])
	if !strings.HasPrefix(firstLine, "```") {
		return nil, 0, fmt.Errorf("not a code block")
	}

	language := strings.TrimSpace(strings.TrimPrefix(firstLine, "```"))

	var codeLines []string
	i := start + 1

	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "```" {
			break
		}
		codeLines = append(codeLines, lines[i])
		i++
	}

	if i >= len(lines) {
		return nil, 0, fmt.Errorf("unclosed code block")
	}

	code := strings.Join(codeLines, "\n")

	node := &Node{
		Type:  "codeBlock",
		Attrs: map[string]interface{}{},
		Content: []Node{
			{
				Type: "text",
				Text: code,
			},
		},
	}

	if language != "" {
		node.Attrs["language"] = language
	}

	return node, i - start + 1, nil
}

// parseBlockquote parses a blockquote
func (p *Parser) parseBlockquote(lines []string, start int) (*Node, int, error) {
	var quoteLines []string
	i := start

	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, ">") {
			break
		}

		// Remove the > prefix and trim
		quoteLine := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
		quoteLines = append(quoteLines, quoteLine)
		i++
	}

	if len(quoteLines) == 0 {
		return nil, 1, nil
	}

	// Parse the content of the blockquote recursively
	quoteContent := strings.Join(quoteLines, "\n")
	quoteDoc, err := p.Parse(quoteContent)
	if err != nil {
		return nil, 0, err
	}

	return &Node{
		Type:    "blockquote",
		Content: quoteDoc.Content,
	}, i - start, nil
}

// parseTable parses a GFM-style table
func (p *Parser) parseTable(lines []string, start int) (*Node, int, error) {
	// First line is header
	headerLine := strings.TrimSpace(lines[start])
	if !strings.HasPrefix(headerLine, "|") {
		return nil, 0, fmt.Errorf("invalid table header")
	}

	// Second line is separator
	if start+1 >= len(lines) {
		return nil, 0, fmt.Errorf("incomplete table")
	}
	separatorLine := strings.TrimSpace(lines[start+1])
	if !strings.Contains(separatorLine, "---") {
		return nil, 0, fmt.Errorf("invalid table separator")
	}

	// Parse header
	headerLine = strings.Trim(headerLine, "|")
	headers := strings.Split(headerLine, "|")
	var columns []map[string]interface{}

	for _, header := range headers {
		header = strings.TrimSpace(header)
		columns = append(columns, map[string]interface{}{
			"label": header,
		})
	}

	// Parse alignment from separator (optional enhancement)
	separatorLine = strings.Trim(separatorLine, "|")
	alignments := strings.Split(separatorLine, "|")
	for i, align := range alignments {
		align = strings.TrimSpace(align)
		if i < len(columns) {
			if strings.HasPrefix(align, ":") && strings.HasSuffix(align, ":") {
				columns[i]["align"] = "center"
			} else if strings.HasSuffix(align, ":") {
				columns[i]["align"] = "right"
			} else {
				columns[i]["align"] = "left"
			}
		}
	}

	// Parse data rows
	var rows []map[string]interface{}
	i := start + 2

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "" || !strings.HasPrefix(line, "|") {
			break
		}

		line = strings.Trim(line, "|")
		cells := strings.Split(line, "|")

		row := make(map[string]interface{})
		for j, cell := range cells {
			cell = strings.TrimSpace(cell)
			if j < len(columns) {
				columnLabel := columns[j]["label"].(string)
				// Process inline markdown in cell content
				row[columnLabel] = p.extractPlainTextFromInline(cell)
			}
		}

		rows = append(rows, row)
		i++
	}

	return &Node{
		Type: "tableBlock",
		Attrs: map[string]interface{}{
			"columns": columns,
			"rows":    rows,
		},
	}, i - start, nil
}

// extractPlainTextFromInline extracts plain text from markdown inline formatting
// Removes markdown markers like **, *, `, etc. but keeps the actual text
func (p *Parser) extractPlainTextFromInline(text string) string {
	// Parse inline content to get nodes
	nodes := p.parseInlineContent(text)

	// Extract just the text content
	var result strings.Builder
	for _, node := range nodes {
		if node.Type == "text" {
			result.WriteString(node.Text)
		}
	}

	return result.String()
}

// parseList parses a bullet or ordered list
func (p *Parser) parseList(lines []string, start int, ordered bool) (*Node, int, error) {
	nodeType := "bulletList"
	listStart := 1

	if ordered {
		nodeType = "orderedList"
		// Extract start number from first item
		re := regexp.MustCompile(`^(\d+)\.`)
		if matches := re.FindStringSubmatch(strings.TrimSpace(lines[start])); len(matches) > 1 {
			fmt.Sscanf(matches[1], "%d", &listStart)
		}
	}

	var items []Node
	i := start

	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			break
		}

		// Check if this is a list item
		var isListItem bool
		var itemText string

		if ordered {
			re := regexp.MustCompile(`^\d+\.\s+(.*)$`)
			if matches := re.FindStringSubmatch(trimmed); len(matches) > 1 {
				isListItem = true
				itemText = matches[1]
			}
		} else {
			re := regexp.MustCompile(`^[\*\-\+]\s+(.*)$`)
			if matches := re.FindStringSubmatch(trimmed); len(matches) > 1 {
				isListItem = true
				itemText = matches[1]
			}
		}

		if !isListItem {
			break
		}

		// Parse item content (could contain nested lists, etc.)
		content := p.parseInlineContent(itemText)

		items = append(items, Node{
			Type: "listItem",
			Content: []Node{
				{
					Type:    "paragraph",
					Content: content,
				},
			},
		})

		i++
	}

	node := &Node{
		Type:    nodeType,
		Content: items,
	}

	if ordered {
		node.Attrs = map[string]interface{}{
			"start": listStart,
		}
	}

	return node, i - start, nil
}

// parseComponent parses a custom JSX component (kept from original implementation)
func (p *Parser) parseComponent(lines []string, start int) (*Node, int, error) {
	line := lines[start]
	trimmed := strings.TrimSpace(line)

	// Extract tag name (support both uppercase and lowercase)
	// Support attributes: name="value" or name={value}
	tagMatch := regexp.MustCompile(`^<([A-Za-z][a-zA-Z0-9]*)((?:\s+[a-zA-Z][a-zA-Z0-9-]*(?:="[^"]*"|=\{[^}]*\})?)*)\s*(/?)>`).FindStringSubmatch(trimmed)
	if len(tagMatch) == 0 {
		return nil, 0, fmt.Errorf("invalid component syntax: %s", line)
	}

	tagName := tagMatch[1]
	attrsStr := tagMatch[2]
	isSelfClosing := tagMatch[3] == "/"

	// Find component definition (case-insensitive)
	var componentDef *ComponentDefinition
	for _, comp := range p.config.Components {
		if strings.EqualFold(comp.MDXTag, tagName) {
			componentDef = comp
			break
		}
	}

	if componentDef == nil {
		return nil, 0, fmt.Errorf("unknown component: %s", tagName)
	}

	// Parse attributes
	attrs := p.parseAttributes(attrsStr)

	// Handle self-closing tags
	if isSelfClosing {
		return &Node{
			Type:  componentDef.JSONType,
			Attrs: attrs,
		}, 1, nil
	}

	// Find closing tag
	closingTag := fmt.Sprintf("</%s>", tagName)
	endLine := -1
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], closingTag) {
			endLine = i
			break
		}
	}

	if endLine == -1 {
		return nil, 0, fmt.Errorf("unclosed tag: %s", tagName)
	}

	// Extract content between tags
	var contentLines []string

	// Handle first line (might have content after opening tag)
	firstLine := lines[start]
	openTagEnd := strings.Index(firstLine, ">")
	if openTagEnd != -1 && openTagEnd < len(firstLine)-1 {
		contentLines = append(contentLines, firstLine[openTagEnd+1:])
	}

	// Add middle lines
	if endLine > start+1 {
		contentLines = append(contentLines, lines[start+1:endLine]...)
	}

	// Handle last line (might have content before closing tag)
	if endLine > start {
		lastLine := lines[endLine]
		closeTagStart := strings.Index(lastLine, closingTag)
		if closeTagStart > 0 {
			contentLines = append(contentLines, lastLine[:closeTagStart])
		}
	}

	content := strings.Join(contentLines, "\n")
	fullTag := strings.Join(lines[start:endLine+1], "\n")

	// Check if component has a custom parser
	if componentDef.CustomParser != nil {
		customNode, err := componentDef.CustomParser(fullTag, p.attributesToStringMap(attrs))
		if err != nil {
			return nil, 0, fmt.Errorf("error in custom parser for %s: %w", tagName, err)
		}
		return customNode, endLine - start + 1, nil
	}

	// Default parsing: parse content recursively
	contentDoc, err := p.Parse(content)
	if err != nil {
		return nil, 0, fmt.Errorf("error parsing component content: %w", err)
	}

	node := &Node{
		Type:    componentDef.JSONType,
		Attrs:   attrs,
		Content: contentDoc.Content,
	}

	return node, endLine - start + 1, nil
}

// parseAttributes parses HTML-style attributes from a string
func (p *Parser) parseAttributes(attrsStr string) map[string]interface{} {
	attrs := make(map[string]interface{})

	// Regular expression to match attribute="value" or attribute={value} or attribute
	attrPattern := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9-]*)(?:="([^"]*)"|=\{([^}]*)\})?`)
	matches := attrPattern.FindAllStringSubmatch(attrsStr, -1)

	for _, match := range matches {
		key := match[1]
		quotedValue := match[2]   // value from ="..."
		bracedValue := match[3]   // value from ={...}

		if quotedValue != "" {
			attrs[key] = quotedValue
		} else if bracedValue != "" {
			// Parse brace values - could be numbers, booleans, etc.
			switch bracedValue {
			case "true":
				attrs[key] = true
			case "false":
				attrs[key] = false
			default:
				// Try to parse as number
				var num float64
				if _, err := fmt.Sscanf(bracedValue, "%f", &num); err == nil {
					attrs[key] = num
				} else {
					// Keep as string
					attrs[key] = bracedValue
				}
			}
		} else {
			// No value means it's a boolean flag
			attrs[key] = true
		}
	}

	return attrs
}

// attributesToStringMap converts attributes map to string map (for custom parsers)
func (p *Parser) attributesToStringMap(attrs map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range attrs {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// parseHTMLImage parses an HTML <img> tag into an imageBlock
func (p *Parser) parseHTMLImage(line string) (*Node, int, error) {
	trimmed := strings.TrimSpace(line)

	// Extract src attribute
	srcPattern := regexp.MustCompile(`src="([^"]*)"`)
	srcMatch := srcPattern.FindStringSubmatch(trimmed)
	if len(srcMatch) < 2 {
		// No src found, treat as text
		return p.parseParagraph([]string{line}, 0)
	}
	src := srcMatch[1]

	// Extract alt attribute
	alt := ""
	altPattern := regexp.MustCompile(`alt="([^"]*)"`)
	altMatch := altPattern.FindStringSubmatch(trimmed)
	if len(altMatch) >= 2 {
		alt = altMatch[1]
	}

	// Determine image type based on src
	imageType := "local"
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		imageType = "url"
	}

	// Normalize local paths to remove ../../../
	if imageType == "local" {
		// Remove all leading ../ patterns
		for strings.HasPrefix(src, "../") {
			src = strings.TrimPrefix(src, "../")
		}
		// Remove leading ./
		src = strings.TrimPrefix(src, "./")
		// Ensure it starts with /
		if !strings.HasPrefix(src, "/") {
			src = "/" + src
		}
	}

	return &Node{
		Type: "imageBlock",
		Attrs: map[string]interface{}{
			"src":     src,
			"alt":     alt,
			"caption": "",
			"type":    imageType,
		},
	}, 1, nil
}
