package converter

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// MarkdownParser processes markdown using stack-based iterative parsing
type MarkdownParser struct {
	stack   []*parseFrame
	session *placeholder.EditSession
	manager placeholder.Manager
}

// parseFrame represents a parsing context on the stack
type parseFrame struct {
	elementType string                 // "details", "expand", "paragraph", etc.
	attributes  map[string]interface{} // Extracted attributes
	state       parseState             // Current parsing state
}

type parseState int

const (
	stateSearchingTag parseState = iota
	stateParsingAttributes
	stateParsingContent
	stateComplete
)

// NewMarkdownParser creates a new parser instance
func NewMarkdownParser(session *placeholder.EditSession, manager placeholder.Manager) *MarkdownParser {
	return &MarkdownParser{
		stack:   make([]*parseFrame, 0, 100), // Pre-allocate for performance
		session: session,
		manager: manager,
	}
}

// ParseMarkdownToADFNodes converts markdown lines to ADF nodes using iterative stack processing
func (p *MarkdownParser) ParseMarkdownToADFNodes(lines []string) ([]adf_types.ADFNode, error) {
	var result []adf_types.ADFNode

	// Reset parser state
	p.stack = p.stack[:0]

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Skip empty lines at document level
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Try to parse the current line
		node, consumed, err := p.parseNext(lines[i:])
		if err != nil {
			return nil, fmt.Errorf("parsing failed at line %d: %w", i+1, err)
		}

		if node != nil {
			result = append(result, *node)
		}

		i += consumed
	}

	// Verify stack is empty (all elements properly closed)
	if len(p.stack) > 0 {
		return nil, fmt.Errorf("unclosed elements detected: %d frames remain on stack", len(p.stack))
	}

	return result, nil
}

// parseNext processes the next markdown content using strategy-based parsing
func (p *MarkdownParser) parseNext(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	line := strings.TrimSpace(lines[0])

	// Strategy-based parsing for special elements
	switch {
	case strings.HasPrefix(line, "<details"):
		return p.parseDetailsElement(lines)
	case strings.HasPrefix(line, "<table"), strings.HasPrefix(line, "<taskList"), strings.HasPrefix(line, "<blockquote"):
		return p.parseXMLWrapper(lines)
	case strings.HasPrefix(line, "<!--"):
		return p.parsePlaceholder(lines)
	case strings.HasPrefix(line, ":::"):
		return p.parsePanel(lines)
	case isGitHubAdmonitionLine(line):
		return p.parsePanel(lines)
	case strings.HasPrefix(line, "```"):
		return p.parseCodeBlock(lines)
	case strings.HasPrefix(line, "#"):
		return p.parseHeading(lines)
	case isThematicBreak(line):
		return p.parseRule(lines)
	case strings.HasPrefix(line, "|"):
		return p.parseMarkdownTable(lines)
	case strings.HasPrefix(strings.TrimSpace(line), "- "):
		return p.parseBulletList(lines)
	default:
		if matched, _ := regexp.MatchString(`^\s*\d+\.\s`, line); matched {
			return p.parseOrderedList(lines)
		}
		return p.parseParagraph(lines)
	}
}

// parseDetailsElement parses HTML details elements
func (p *MarkdownParser) parseDetailsElement(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	firstLine := strings.TrimSpace(lines[0])

	// Match opening tag: <details> or <details open> or <details id="...">
	detailsRegex := regexp.MustCompile(`^<details(\s+[^>]*)?>`)
	if !detailsRegex.MatchString(firstLine) {
		return nil, 1, nil
	}

	// Extract attributes using stack-based parsing
	frame := &parseFrame{
		elementType: "details",
		attributes:  make(map[string]interface{}),
		state:       stateParsingAttributes,
	}

	// Parse attributes from opening tag
	if err := p.parseDetailsAttributes(firstLine, frame); err != nil {
		return nil, 1, err
	}

	// Push frame onto stack
	p.stack = append(p.stack, frame)

	// Find content boundaries iteratively
	contentStart, contentEnd, title, err := p.findDetailsContent(lines)
	if err != nil {
		p.stack = p.stack[:len(p.stack)-1] // Pop frame
		return nil, 1, err
	}

	// Determine node type from parsed attributes
	nodeType := adf_types.NodeTypeExpand // Default to expand
	if detectedType, exists := frame.attributes["__nodeType"]; exists {
		if detectedType == "nestedExpand" {
			nodeType = adf_types.NodeTypeNestedExpand
		}
	}

	// Create ADF node with correct type
	node := &adf_types.ADFNode{
		Type:  nodeType,
		Attrs: map[string]interface{}{"title": title},
	}

	// Apply parsed attributes (excluding internal __nodeType)
	for key, value := range frame.attributes {
		if key != "__nodeType" { // Skip internal attribute
			node.Attrs[key] = value
		}
	}

	// Parse content using stack (iterative, not recursive)
	if contentEnd > contentStart {
		contentLines := lines[contentStart:contentEnd]
		contentNodes, err := p.parseContentIteratively(contentLines)
		if err != nil {
			p.stack = p.stack[:len(p.stack)-1] // Pop frame
			return nil, contentEnd, err
		}
		node.Content = contentNodes
	}

	// Pop frame from stack
	p.stack = p.stack[:len(p.stack)-1]

	return node, contentEnd + 1, nil
}

// parseDetailsAttributes extracts attributes from details opening tag
//
//nolint:unparam // error return reserved for future use
func (p *MarkdownParser) parseDetailsAttributes(line string, frame *parseFrame) error {
	// Check for 'open' attribute
	if strings.Contains(line, " open") || strings.Contains(line, " open>") {
		frame.attributes["expanded"] = true
	}

	idRegex := regexp.MustCompile(`\sid\s*=\s*["|']([^"']*)["|']`)
	if idMatch := idRegex.FindStringSubmatch(line); len(idMatch) > 1 {
		frame.attributes["localId"] = idMatch[1]
	}

	// CRITICAL FIX: Extract data-adf-type attribute to preserve original node type
	adfTypeRegex := regexp.MustCompile(`data-adf-type="([^"]+)"`)
	if matches := adfTypeRegex.FindStringSubmatch(line); len(matches) > 1 {
		frame.attributes["__nodeType"] = matches[1] // Store as internal attribute
	}

	return nil
}

// findDetailsContent locates content boundaries within details element
func (p *MarkdownParser) findDetailsContent(lines []string) (contentStart, contentEnd int, title string, err error) {
	// Look for <summary> tag iteratively
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for summary tag
		summaryRegex := regexp.MustCompile(`<summary>([^<]*)</summary>`)
		if summaryMatch := summaryRegex.FindStringSubmatch(line); len(summaryMatch) > 1 {
			title = strings.TrimSpace(summaryMatch[1])

			// Determine content start
			afterSummary := summaryRegex.ReplaceAllString(line, "")
			if strings.TrimSpace(afterSummary) != "" && !strings.HasPrefix(strings.TrimSpace(afterSummary), "<details") {
				contentStart = i
			} else {
				contentStart = i + 1
			}
			break
		}

		// Limit search to prevent excessive scanning
		if i > 5 {
			return 0, 0, "", fmt.Errorf("summary tag not found within reasonable range")
		}
	}

	// Title is required
	if title == "" {
		return 0, 0, "", fmt.Errorf("details element missing required summary tag")
	}

	// Find closing tag iteratively
	nestingLevel := 0
	for i := contentStart; i < len(lines); i++ {
		line := lines[i]

		// Count nested details tags
		if strings.Contains(line, "<details") {
			nestingLevel++
		}

		// Check for closing tag
		if strings.Contains(line, "</details>") {
			if nestingLevel > 0 {
				nestingLevel--
			} else {
				// Extract content before closing tag if any
				before := strings.Split(line, "</details>")[0]
				if strings.TrimSpace(before) != "" {
					contentEnd = i + 1
				} else {
					contentEnd = i
				}
				return contentStart, contentEnd, title, nil
			}
		}
	}

	return 0, 0, "", fmt.Errorf("unclosed details element")
}

// parseContentIteratively processes content using stack instead of recursion
func (p *MarkdownParser) parseContentIteratively(lines []string) ([]adf_types.ADFNode, error) {
	// Prevent stack overflow even with iterative approach
	if len(p.stack) > 100 {
		return nil, fmt.Errorf("maximum nesting depth exceeded (100 levels) - content too complex")
	}

	var result []adf_types.ADFNode
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Parse next element iteratively
		node, consumed, err := p.parseNext(lines[i:])
		if err != nil {
			return nil, err
		}

		if node != nil {
			result = append(result, *node)
		}

		// CRITICAL: Prevent infinite loops - always advance at least 1 line
		if consumed == 0 {
			consumed = 1
		}
		i += consumed
	}

	return result, nil
}

// Placeholder parsing methods that delegate to existing implementations
func (p *MarkdownParser) parsePlaceholder(lines []string) (*adf_types.ADFNode, int, error) {
	return parsePlaceholderNode(lines, p.session, p.manager)
}

func (p *MarkdownParser) parseXMLWrapper(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, fmt.Errorf("no lines to parse")
	}

	firstLine := strings.TrimSpace(lines[0])

	// Determine element type and use converter via registry
	var nodeType ADFNodeType
	switch {
	case strings.HasPrefix(firstLine, "<table"):
		nodeType = "table"
	case strings.HasPrefix(firstLine, "<taskList"):
		nodeType = "taskList"
	case strings.HasPrefix(firstLine, "<blockquote"):
		nodeType = "blockquote"
	default:
		return nil, 1, fmt.Errorf("unsupported XML wrapper: %s", firstLine)
	}

	// Get converter from registry (same pattern as parseHeading)
	if converter := globalRegistry.GetConverter(nodeType); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, ConversionContext{})
		return &node, consumed, err
	}

	return nil, 1, fmt.Errorf("converter not registered: %s", nodeType)
}

func (p *MarkdownParser) parseHeading(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	// Create cleaned lines with trimmed first line to handle indented headings
	cleanedLines := make([]string, len(lines))
	cleanedLines[0] = strings.TrimSpace(lines[0]) // Remove indentation from heading line
	copy(cleanedLines[1:], lines[1:])

	// Try to get converter from registry, fallback to legacy function
	if converter := globalRegistry.GetConverter("heading"); converter != nil {
		node, consumed, err := converter.FromMarkdown(cleanedLines, 0, ConversionContext{})
		return &node, consumed, err
	}

	// Fallback to legacy function
	node, consumed, err := parseHeading(cleanedLines)
	return &node, consumed, err
}

func (p *MarkdownParser) parseBulletList(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	// Collect list lines while preserving indentation for nesting
	// Collect list lines including:
	// - Lines starting with bullet markers (-, *, +)
	// - Indented continuation lines (for multiline items and nested lists)
	cleanedLines := make([]string, 0, len(lines))
	inList := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Empty line ends the list
		if trimmed == "" {
			break
		}

		// Check if line starts with bullet marker
		isBulletLine := strings.HasPrefix(trimmed, "- ") ||
			strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "+ ")

		if isBulletLine {
			// Bullet line - always include
			inList = true
			cleanedLines = append(cleanedLines, line)
		} else if inList && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// Indented continuation line - include if we're in a list
			cleanedLines = append(cleanedLines, line)
		} else {
			// Non-indented, non-bullet line - end of list
			break
		}
	}

	// Try to get converter from registry
	if converter := globalRegistry.GetConverter("bulletList"); converter != nil {
		// Pass placeholder manager through context for emoji/placeholder restoration
		ctx := ConversionContext{
			PlaceholderManager: p.manager,
		}
		node, consumed, err := converter.FromMarkdown(cleanedLines, 0, ctx)
		return &node, consumed, err
	}

	// Converter not registered (should not happen in production)
	return nil, 0, fmt.Errorf("bulletList converter not registered")
}

func (p *MarkdownParser) parseOrderedList(lines []string) (*adf_types.ADFNode, int, error) {
	// Try to get converter from registry
	if converter := globalRegistry.GetConverter("orderedList"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, ConversionContext{})
		return &node, consumed, err
	}

	// Converter not registered (should not happen in production)
	return nil, 0, fmt.Errorf("orderedList converter not registered")
}

// admonitionDispatchRegex matches GitHub-style admonition headers for panel dispatch
var admonitionDispatchRegex = regexp.MustCompile(`(?i)^>\s*\[!(INFO|WARNING|ERROR|SUCCESS|NOTE|TIP)\]\s*$`)

// isGitHubAdmonitionLine checks if a line matches > [!TYPE] pattern for supported panel types
func isGitHubAdmonitionLine(line string) bool {
	return admonitionDispatchRegex.MatchString(line)
}

func (p *MarkdownParser) parsePanel(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("panel"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, ConversionContext{})
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("panel converter not registered")
}

func (p *MarkdownParser) parseCodeBlock(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("codeBlock"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, ConversionContext{})
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("codeBlock converter not registered")
}

func (p *MarkdownParser) parseParagraph(lines []string) (*adf_types.ADFNode, int, error) {
	// Try to get converter from registry, fallback to legacy function
	if converter := globalRegistry.GetConverter("paragraph"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, ConversionContext{})
		return &node, consumed, err
	}

	// Fallback to legacy function
	node, consumed, err := parseParagraph(lines)
	return &node, consumed, err
}

func (p *MarkdownParser) parseMarkdownTable(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("table"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, ConversionContext{})
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("table converter not registered")
}

func (p *MarkdownParser) parseRule(lines []string) (*adf_types.ADFNode, int, error) {
	node := &adf_types.ADFNode{Type: adf_types.NodeTypeRule}
	return node, 1, nil
}

// isThematicBreak returns true if the line is a Markdown thematic break (---, ***, ___).
func isThematicBreak(line string) bool {
	if len(line) < 3 {
		return false
	}

	ch := line[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}

	for i := 1; i < len(line); i++ {
		if line[i] != ch {
			return false
		}
	}

	return true
}

// GetStackDepth returns current stack depth for debugging/monitoring
func (p *MarkdownParser) GetStackDepth() int {
	return len(p.stack)
}

// IsStackEmpty checks if parser stack is clean
func (p *MarkdownParser) IsStackEmpty() bool {
	return len(p.stack) == 0
}
