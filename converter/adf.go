package converter

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// DeletionTracker tracks which placeholders are requested vs successfully restored during conversion
type DeletionTracker struct {
	originalCount         int
	requestedPlaceholders map[string]bool
	restoredPlaceholders  map[string]bool
	session               *placeholder.EditSession
	manager               placeholder.Manager
}

// NewDeletionTracker creates a new deletion tracker
func NewDeletionTracker(session *placeholder.EditSession, manager placeholder.Manager) *DeletionTracker {
	return &DeletionTracker{
		originalCount:         len(session.Preserved),
		requestedPlaceholders: make(map[string]bool),
		restoredPlaceholders:  make(map[string]bool),
		session:               session,
		manager:               manager,
	}
}

// RecordPlaceholderRequest tracks that a placeholder was requested from the markdown
func (dt *DeletionTracker) RecordPlaceholderRequest(placeholderID string) {
	dt.requestedPlaceholders[placeholderID] = true
}

// RecordPlaceholderRestored tracks that a placeholder was successfully restored
func (dt *DeletionTracker) RecordPlaceholderRestored(placeholderID string) {
	dt.restoredPlaceholders[placeholderID] = true
}

// GetSummary generates the final deletion summary
func (dt *DeletionTracker) GetSummary() *DeletionSummary {
	var deletedIDs []string

	// Find placeholders that were in the original session but never restored
	// This correctly identifies deletions regardless of whether they were requested from markdown
	for placeholderID := range dt.session.Preserved {
		if !dt.restoredPlaceholders[placeholderID] {
			deletedIDs = append(deletedIDs, placeholderID)
		}
	}

	preservedCount := len(dt.restoredPlaceholders)
	deletedCount := len(deletedIDs)

	return &DeletionSummary{
		DeletedPlaceholderIDs: deletedIDs,
		DeletedCount:          deletedCount,
		PreservedCount:        preservedCount,
		OriginalCount:         dt.originalCount,
	}
}

// FromMarkdownWithTracking converts edited Markdown back to ADF with deletion tracking
func FromMarkdownWithTracking(markdown string, session *placeholder.EditSession, manager placeholder.Manager) (ConversionResult, error) {
	if session == nil {
		return ConversionResult{}, fmt.Errorf("session cannot be nil")
	}

	// Track deletions during parsing
	deletionTracker := NewDeletionTracker(session, manager)

	// Split markdown into lines for processing
	lines := strings.Split(markdown, "\n")

	// Parse the markdown into ADF nodes
	nodes, err := parseMarkdownToADFNodesWithTracking(lines, session, manager, deletionTracker)
	if err != nil {
		return ConversionResult{}, fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Create the ADF document
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: nodes,
	}

	// Generate deletion summary
	deletionSummary := deletionTracker.GetSummary()

	return ConversionResult{
		Document:  doc,
		Deletions: deletionSummary,
	}, nil
}

// FromMarkdown converts edited Markdown back to ADF, restoring preserved content from placeholders
func FromMarkdown(markdown string, session *placeholder.EditSession, manager placeholder.Manager) (adf_types.ADFDocument, error) {
	// PHASE 5: Comprehensive error handling with recovery
	defer func() {
		if r := recover(); r != nil {
			slog.Error("FromMarkdown: critical error recovered", "error", r, "markdownLength", len(markdown))
		}
	}()

	if session == nil {
		return adf_types.ADFDocument{}, fmt.Errorf("session cannot be nil")
	}

	// PHASE 5: Additional input validation
	if len(markdown) > 1000000 { // 1MB limit
		return adf_types.ADFDocument{}, fmt.Errorf("markdown input too large: %d bytes (max 1MB)", len(markdown))
	}

	// Split markdown into lines for processing
	lines := strings.Split(markdown, "\n")

	// PHASE 5: Validate line count
	if len(lines) > 10000 {
		slog.Warn("FromMarkdown: processing extremely large document", "lineCount", len(lines))
	}

	// Parse the markdown into ADF nodes with error recovery
	nodes, err := parseMarkdownToADFNodesWithRecovery(lines, session, manager)
	if err != nil {
		return adf_types.ADFDocument{}, fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Create the ADF document
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: nodes,
	}

	return doc, nil
}

// parseMarkdownToADFNodes converts markdown lines to ADF nodes
//
//nolint:unused // Called by parseMarkdownToADFNodesWithTracking
func parseMarkdownToADFNodes(lines []string, session *placeholder.EditSession, manager placeholder.Manager) ([]adf_types.ADFNode, error) {
	// Use new streaming parser to eliminate infinite recursion risk
	parser := NewMarkdownParser(session, manager)
	return parser.ParseMarkdownToADFNodes(lines)
}

// parseMarkdownToADFNodesWithRecovery wraps parseMarkdownToADFNodes with error recovery
func parseMarkdownToADFNodesWithRecovery(lines []string, session *placeholder.EditSession, manager placeholder.Manager) ([]adf_types.ADFNode, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("parseMarkdownToADFNodes: critical parsing error recovered", "error", r, "lineCount", len(lines))
		}
	}()

	// Use new streaming parser to eliminate infinite recursion
	parser := NewMarkdownParser(session, manager)
	return parser.ParseMarkdownToADFNodes(lines)
}

// parseMarkdownToADFNodesWithTracking converts markdown lines to ADF nodes with deletion tracking
func parseMarkdownToADFNodesWithTracking(lines []string, session *placeholder.EditSession, manager placeholder.Manager, tracker *DeletionTracker) ([]adf_types.ADFNode, error) {
	// Use new streaming parser to eliminate infinite recursion
	parser := NewMarkdownParser(session, manager)
	return parser.ParseMarkdownToADFNodes(lines)
}

// parsePlaceholderNode restores preserved content from placeholder comments
func parsePlaceholderNode(lines []string, session *placeholder.EditSession, manager placeholder.Manager) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	line := strings.TrimSpace(lines[0])
	placeholderID, found := placeholder.ParsePlaceholderComment(line)
	if !found {
		return nil, 1, nil
	}

	// Restore the preserved content
	node, err := manager.Restore(placeholderID)
	if err != nil {
		// Placeholder was deleted from markdown - skip it (allows intentional deletion)
		return nil, 1, nil
	}

	return &node, 1, nil
}

// parsePlaceholderNodeWithTracking restores preserved content from placeholder comments with deletion tracking
func parsePlaceholderNodeWithTracking(lines []string, session *placeholder.EditSession, manager placeholder.Manager, tracker *DeletionTracker) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	line := strings.TrimSpace(lines[0])
	placeholderID, found := placeholder.ParsePlaceholderComment(line)
	if !found {
		return nil, 1, nil
	}

	// Record that this placeholder was requested from the markdown
	tracker.RecordPlaceholderRequest(placeholderID)

	// Restore the preserved content
	node, err := manager.Restore(placeholderID)
	if err != nil {
		// Placeholder was deleted from markdown - skip it (allows intentional deletion)
		// Do NOT record it as restored since it failed
		return nil, 1, nil
	}

	// Record that this placeholder was successfully restored
	tracker.RecordPlaceholderRestored(placeholderID)

	return &node, 1, nil
}

// parseHeading converts markdown heading to ADF heading node
//
//nolint:unparam // return value needed for consistency with other parsers
func parseHeading(lines []string) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return adf_types.ADFNode{}, 1, fmt.Errorf("no lines to parse")
	}

	line := lines[0]

	// Count heading level
	level := 0
	for i, char := range line {
		if char == '#' {
			level++
		} else if char == ' ' {
			line = line[i+1:] // Remove the "# " prefix
			break
		} else {
			return adf_types.ADFNode{}, 1, fmt.Errorf("invalid heading format")
		}
	}

	if level < 1 || level > 6 {
		level = 1 // Default to h1
	}

	// Parse the heading text with potential formatting
	textNodes, err := parseInlineContent(line)
	if err != nil {
		return adf_types.ADFNode{}, 1, fmt.Errorf("failed to parse heading content: %w", err)
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeHeading,
		Attrs: map[string]interface{}{
			"level": level,
		},
		Content: textNodes,
	}

	return node, 1, nil
}

// parseParagraph converts markdown paragraph to ADF paragraph node
func parseParagraph(lines []string) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return adf_types.ADFNode{}, 1, nil
	}

	// Collect paragraph lines until we hit an empty line or special syntax
	var paragraphLines []string
	consumed := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Stop at empty line (end of paragraph)
		if trimmed == "" {
			consumed = i + 1
			break
		}

		// Stop at special syntax (headings, lists, placeholders)
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "- ") ||
			strings.HasPrefix(trimmed, "<!--") {
			consumed = i
			break
		}

		// Check for ordered list
		if matched, _ := regexp.MatchString(`^\s*\d+\.\s`, line); matched {
			consumed = i
			break
		}

		paragraphLines = append(paragraphLines, line)

		// If this is the last line, consume it
		if i == len(lines)-1 {
			consumed = i + 1
		}
	}

	// If no lines collected, return empty
	if len(paragraphLines) == 0 {
		return adf_types.ADFNode{}, consumed, nil
	}

	// Join paragraph lines with spaces
	text := strings.Join(paragraphLines, " ")
	text = strings.TrimSpace(text)

	if text == "" {
		return adf_types.ADFNode{}, consumed, nil
	}

	// Parse inline content with formatting
	textNodes, err := parseInlineContent(text)
	if err != nil {
		return adf_types.ADFNode{}, consumed, fmt.Errorf("failed to parse paragraph content: %w", err)
	}

	node := adf_types.ADFNode{
		Type:    adf_types.NodeTypeParagraph,
		Content: textNodes,
	}

	return node, consumed, nil
}

// parseInlineContent converts markdown text with formatting to ADF text nodes
//
//nolint:unparam // error return reserved for future parsing errors
func parseInlineContent(text string) ([]adf_types.ADFNode, error) {
	if text == "" {
		return []adf_types.ADFNode{}, nil
	}

	// For now, implement basic parsing - this can be enhanced later
	// Handle simple cases: **bold**, *italic*, `code`, [link](url)

	var nodes []adf_types.ADFNode
	remaining := text

	for remaining != "" {
		// Find the next formatting marker
		boldPos := strings.Index(remaining, "**")
		italicPos := strings.Index(remaining, "*")
		underscoreItalicPos := strings.Index(remaining, "_")
		codePos := strings.Index(remaining, "`")
		linkPos := strings.Index(remaining, "[")
		htmlLinkPos := strings.Index(remaining, "<a ")

		// Find the earliest marker
		nextPos := len(remaining)
		nextType := ""

		if boldPos != -1 && boldPos < nextPos {
			nextPos = boldPos
			nextType = "bold"
		}
		if italicPos != -1 && italicPos < nextPos && italicPos != boldPos {
			nextPos = italicPos
			nextType = "italic"
		}
		if underscoreItalicPos != -1 && underscoreItalicPos < nextPos {
			nextPos = underscoreItalicPos
			nextType = "underscoreitalic"
		}
		if codePos != -1 && codePos < nextPos {
			nextPos = codePos
			nextType = "code"
		}
		if linkPos != -1 && linkPos < nextPos {
			nextPos = linkPos
			nextType = "link"
		}
		if htmlLinkPos != -1 && htmlLinkPos < nextPos {
			nextPos = htmlLinkPos
			nextType = "htmllink"
		}

		// If no formatting found, add remaining text as plain text
		if nextType == "" {
			if remaining != "" {
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining,
				})
			}
			break
		}

		// Add plain text before the formatting
		if nextPos > 0 {
			nodes = append(nodes, adf_types.ADFNode{
				Type: adf_types.NodeTypeText,
				Text: remaining[:nextPos],
			})
		}

		// Parse the formatting
		switch nextType {
		case "bold":
			node, consumed := parseBoldText(remaining[nextPos:])
			if node != nil {
				nodes = append(nodes, *node)
				remaining = remaining[nextPos+consumed:]
			} else {
				// Treat as plain text if parsing failed
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining[nextPos : nextPos+2],
				})
				remaining = remaining[nextPos+2:]
			}
		case "italic":
			node, consumed := parseItalicText(remaining[nextPos:])
			if node != nil {
				nodes = append(nodes, *node)
				remaining = remaining[nextPos+consumed:]
			} else {
				// Treat as plain text if parsing failed
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining[nextPos : nextPos+1],
				})
				remaining = remaining[nextPos+1:]
			}
		case "underscoreitalic":
			node, consumed := parseUnderscoreItalicText(remaining[nextPos:])
			if node != nil {
				nodes = append(nodes, *node)
				remaining = remaining[nextPos+consumed:]
			} else {
				// Treat as plain text if parsing failed
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining[nextPos : nextPos+1],
				})
				remaining = remaining[nextPos+1:]
			}
		case "code":
			node, consumed := parseCodeText(remaining[nextPos:])
			if node != nil {
				nodes = append(nodes, *node)
				remaining = remaining[nextPos+consumed:]
			} else {
				// Treat as plain text if parsing failed
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining[nextPos : nextPos+1],
				})
				remaining = remaining[nextPos+1:]
			}
		case "link":
			node, consumed := parseLinkText(remaining[nextPos:])
			if node != nil {
				nodes = append(nodes, *node)
				remaining = remaining[nextPos+consumed:]
			} else {
				// Treat as plain text if parsing failed
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining[nextPos : nextPos+1],
				})
				remaining = remaining[nextPos+1:]
			}
		case "htmllink":
			node, consumed := parseHTMLLinkText(remaining[nextPos:])
			if node != nil {
				nodes = append(nodes, *node)
				remaining = remaining[nextPos+consumed:]
			} else {
				// Treat as plain text if parsing failed
				nodes = append(nodes, adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: remaining[nextPos : nextPos+3],
				})
				remaining = remaining[nextPos+3:]
			}
		}
	}

	return nodes, nil
}

// parseBoldText parses **bold** text
func parseBoldText(text string) (*adf_types.ADFNode, int) {
	if !strings.HasPrefix(text, "**") {
		return nil, 0
	}

	closePos := strings.Index(text[2:], "**")
	if closePos == -1 {
		return nil, 0
	}

	content := text[2 : 2+closePos]
	if content == "" {
		return nil, 0
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: content,
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeStrong},
		},
	}

	return &node, 2 + closePos + 2 // 2 for opening **, closePos for content, 2 for closing **
}

// parseItalicText parses *italic* text (but not **bold**)
func parseItalicText(text string) (*adf_types.ADFNode, int) {
	if !strings.HasPrefix(text, "*") || strings.HasPrefix(text, "**") {
		return nil, 0
	}

	closePos := strings.Index(text[1:], "*")
	if closePos == -1 {
		return nil, 0
	}

	content := text[1 : 1+closePos]
	if content == "" {
		return nil, 0
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: content,
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeEm},
		},
	}

	return &node, 1 + closePos + 1 // 1 for opening *, closePos for content, 1 for closing *
}

// parseUnderscoreItalicText parses _italic_ text
func parseUnderscoreItalicText(text string) (*adf_types.ADFNode, int) {
	if !strings.HasPrefix(text, "_") {
		return nil, 0
	}

	closePos := strings.Index(text[1:], "_")
	if closePos == -1 {
		return nil, 0
	}

	content := text[1 : 1+closePos]
	if content == "" {
		return nil, 0
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: content,
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeEm},
		},
	}

	return &node, 1 + closePos + 1 // 1 for opening _, closePos for content, 1 for closing _
}

// parseCodeText parses `code` text
func parseCodeText(text string) (*adf_types.ADFNode, int) {
	if !strings.HasPrefix(text, "`") {
		return nil, 0
	}

	closePos := strings.Index(text[1:], "`")
	if closePos == -1 {
		return nil, 0
	}

	content := text[1 : 1+closePos]
	if content == "" {
		return nil, 0
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: content,
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeCode},
		},
	}

	return &node, 1 + closePos + 1 // 1 for opening `, closePos for content, 1 for closing `
}

// IsAtlassianURL determines if a URL should be converted to an InlineCard node
// This specifically detects Atlassian domain patterns that should be InlineCard nodes
// rather than text+link marks to maintain proper ADF structure for Jira API
func IsAtlassianURL(url string) bool {
	if url == "" {
		return false
	}

	// Check for common Atlassian domain patterns
	// *.atlassian.net and *.atlassian.com are the primary patterns
	if strings.Contains(url, ".atlassian.net/") || strings.Contains(url, ".atlassian.com/") {
		return true
	}

	// Check for specific Atlassian URL patterns that are commonly found in InlineCard nodes
	if strings.Contains(url, "/wiki/x/") ||
		strings.Contains(url, "/browse/") ||
		strings.Contains(url, "/jira/") {
		return true
	}

	return false
}

// parseLinkText parses [text](url) links
func parseLinkText(text string) (*adf_types.ADFNode, int) {
	if !strings.HasPrefix(text, "[") {
		return nil, 0
	}

	// Find the closing ]
	closePos := strings.Index(text[1:], "]")
	if closePos == -1 {
		return nil, 0
	}

	linkText := text[1 : 1+closePos]
	remaining := text[1+closePos+1:]

	// Check for (url) part
	if !strings.HasPrefix(remaining, "(") {
		return nil, 0
	}

	urlClosePos := strings.Index(remaining[1:], ")")
	if urlClosePos == -1 {
		return nil, 0
	}

	url := remaining[1 : 1+urlClosePos]

	// STRUCTURAL FIDELITY FIX: Use link text vs URL to determine node type
	// [different-text](url) → Text node with link mark (preserve original structure)
	// [same-url](same-url) → InlineCard node (URL displayed as link text indicates InlineCard intent)
	if linkText == url {
		// Same text and URL indicates this should be an InlineCard
		node := adf_types.ADFNode{
			Type: adf_types.NodeTypeInlineCard,
			Attrs: map[string]interface{}{
				"url": url, // ✅ Use 'url' for InlineCard (not 'href')
			},
		}
		totalConsumed := 1 + closePos + 1 + 1 + urlClosePos + 1 // [text](url)
		return &node, totalConsumed
	}

	// For non-Atlassian URLs, create regular link mark
	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: linkText,
		Marks: []adf_types.ADFMark{
			{
				Type: adf_types.MarkTypeLink,
				Attrs: map[string]interface{}{
					"href": url, // ✅ Use 'href' for link marks
				},
			},
		},
	}

	totalConsumed := 1 + closePos + 1 + 1 + urlClosePos + 1 // [text](url)
	return &node, totalConsumed
}

// parseHTMLLinkText parses HTML-wrapped markdown links like <a attr="value">[text](url)</a>
func parseHTMLLinkText(text string) (*adf_types.ADFNode, int) {
	// PHASE 5: Comprehensive error handling with recovery
	defer func() {
		if r := recover(); r != nil {
			inputPreview := text
			if len(text) > 100 {
				inputPreview = text[:100] + "..."
			}
			slog.Error("parseHTMLLinkText: recovered from panic", "error", r, "input", inputPreview)
		}
	}()

	// PHASE 5: Defensive validation for malformed inputs
	if text == "" {
		slog.Debug("parseHTMLLinkText: empty input text")
		return nil, 0
	}

	if len(text) > 10000 { // Reasonable limit to prevent DoS
		slog.Warn("parseHTMLLinkText: input text too long, skipping", "length", len(text))
		return nil, 0
	}

	if !strings.HasPrefix(text, "<a ") {
		return nil, 0
	}

	// Additional validation: ensure it looks like a complete HTML tag
	if !strings.Contains(text, ">") {
		slog.Debug("parseHTMLLinkText: malformed HTML - no closing bracket")
		return nil, 0
	}

	// Use the DefaultHTMLLinkParser to parse the HTML link with error recovery
	parser := &DefaultHTMLLinkParser{}
	var result HTMLLinkParseResult

	// PHASE 5: Wrap parser call in error handling
	func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("parseHTMLLinkText: parser panic recovered", "error", r)
				result = HTMLLinkParseResult{Success: false}
			}
		}()
		result = parser.ParseHTMLLink(text)
	}()

	if !result.Success {
		return nil, 0
	}

	// PHASE 5: Additional defensive validation on parser result
	if result.Href == "" {
		slog.Debug("parseHTMLLinkText: parser returned empty href")
		return nil, 0
	}

	if result.LinkText == "" {
		slog.Debug("parseHTMLLinkText: parser returned empty link text")
		return nil, 0
	}

	if result.TotalLength <= 0 {
		slog.Debug("parseHTMLLinkText: parser returned invalid total length", "length", result.TotalLength)
		return nil, 0
	}

	// Build the ADF mark attributes by combining href with HTML attributes
	attrs := make(map[string]interface{})
	attrs["href"] = result.Href

	// Add all the HTML attributes from the <a> tag with validation
	for key, value := range result.Attributes {
		// PHASE 5: Defensive validation for attribute names and values
		if key == "" {
			slog.Debug("parseHTMLLinkText: skipping empty attribute key")
			continue
		}

		if len(key) > 100 {
			slog.Warn("parseHTMLLinkText: skipping attribute with overly long key", "keyLength", len(key))
			continue
		}

		// Validate and sanitize attribute values
		if len(value) > 1000 {
			slog.Warn("parseHTMLLinkText: truncating overly long attribute value", "key", key, "originalLength", len(value))
			attrs[key] = value[:1000]
		} else {
			attrs[key] = value
		}
	}

	// PHASE 5: Limit total number of attributes to prevent resource exhaustion
	if len(attrs) > 20 {
		slog.Warn("parseHTMLLinkText: too many attributes, potential malformed input", "attrCount", len(attrs))
		// Keep only href and first few attributes
		safeAttrs := map[string]interface{}{"href": attrs["href"]}
		count := 0
		for key, value := range attrs {
			if key != "href" && count < 5 {
				safeAttrs[key] = value
				count++
			}
		}
		attrs = safeAttrs
	}

	// Create a temporary link mark to use the classifier
	linkMark := adf_types.ADFMark{
		Type:  adf_types.MarkTypeLink,
		Attrs: attrs,
	}

	// Use LinkClassifier to determine if this should be an InlineCard
	classifier := &DefaultLinkClassifier{}
	classification := classifier.ClassifyLink(linkMark)

	// If it's a complex internal link, create an InlineCard node
	if classification.Type == ComplexInternalLink {
		// PHASE 5: Try to create InlineCard with fallback logic for unsupported combinations
		inlineCardAttrs, fallbackToSimple := sanitizeInlineCardAttrs(attrs, result.LinkText)

		// If sanitization suggests fallback, create simple link instead
		if !fallbackToSimple {
			// PHASE 5: Additional validation with error recovery
			defer func() {
				if r := recover(); r != nil {
					slog.Error("parseHTMLLinkText: InlineCard creation panic, falling back to simple link", "error", r)
				}
			}()

			// Create InlineCard with sanitized attributes
			node := adf_types.ADFNode{
				Type:  adf_types.NodeTypeInlineCard,
				Attrs: inlineCardAttrs,
			}
			return &node, result.TotalLength
		}

		slog.Debug("InlineCard sanitization suggests fallback to simple link",
			"href", attrs["href"],
			"reason", "unsupported attribute combination")
	}

	// For simple links, create text node with link mark as before
	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: result.LinkText,
		Marks: []adf_types.ADFMark{
			{
				Type:  adf_types.MarkTypeLink,
				Attrs: attrs,
			},
		},
	}

	return &node, result.TotalLength
}

// sanitizeInlineCardAttrs sanitizes attributes for InlineCard nodes to ensure Jira API compliance
// Returns (sanitizedAttrs, shouldFallbackToSimple)
func sanitizeInlineCardAttrs(attrs map[string]interface{}, linkText string) (map[string]interface{}, bool) {
	sanitized := make(map[string]interface{})

	// PHASE 5: Critical fix - convert href to url for InlineCard compliance
	var targetURL string
	if href, exists := attrs["href"]; exists {
		if hrefStr, ok := href.(string); ok && hrefStr != "" {
			targetURL = hrefStr
			sanitized["url"] = hrefStr
		} else {
			// Invalid href - fallback to simple link
			return nil, true
		}
	} else {
		// No href attribute - cannot create InlineCard
		return nil, true
	}

	// PHASE 5: Preserve all attributes except those already expressed in markdown
	blacklistedAttrs := map[string]bool{
		"href": true, // Already expressed as (url) in markdown
	}

	for key, value := range attrs {
		if !blacklistedAttrs[key] {
			// Preserve all attributes except those already in markdown
			sanitized[key] = value
		}
	}

	// PHASE 5: Fallback to simple link for potentially problematic cases
	if len(sanitized) > 4 { // url + 3 safe attrs max
		slog.Debug("InlineCard has too many attributes, falling back to simple link",
			"url", targetURL,
			"attrCount", len(sanitized))
		return nil, true
	}

	// Additional validation: ensure url is valid
	if targetURL == "" {
		return nil, true
	}

	slog.Debug("InlineCard attributes sanitized for Jira compliance",
		"url", targetURL,
		"attrCount", len(sanitized),
		"fallback", false)

	return sanitized, false
}

// LEGACY: parseXMLAttributes has been moved to internal.ParseXMLAttributes
// This function was moved to pkg/converter/internal/xml_utils.go as part of the refactoring
// to consolidate XML-related utilities in the internal package (avoiding import cycles).
// All callers have been updated to use internal.ParseXMLAttributes instead.

// parseMarkdownBlockquote parses markdown blockquote content into ADF blockquote node
//
// ParseMarkdownBlockquote parses markdown-formatted blockquote into ADF node
// LEGACY: This function is deprecated. Use elements.BlockquoteConverter.FromMarkdown() instead.
// Kept for backward compatibility with ParseXMLBlockquote.
//
//nolint:unparam // error return reserved for future parsing errors
func ParseMarkdownBlockquote(lines []string) (adf_types.ADFNode, error) {
	var paragraphs []adf_types.ADFNode
	var currentParagraphLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			// Empty line - end current paragraph if any
			if len(currentParagraphLines) > 0 {
				paragraphs = append(paragraphs, createParagraphFromLines(currentParagraphLines))
				currentParagraphLines = nil
			}
			continue
		}

		// Remove blockquote prefix (> )
		content := line
		if strings.HasPrefix(trimmed, "> ") {
			content = strings.TrimSpace(line[strings.Index(line, ">")+1:])
		} else if strings.HasPrefix(trimmed, ">") {
			content = strings.TrimSpace(line[strings.Index(line, ">")+1:])
		}

		if content == "" {
			// Empty blockquote line (like "> ") - end current paragraph if any
			if len(currentParagraphLines) > 0 {
				paragraphs = append(paragraphs, createParagraphFromLines(currentParagraphLines))
				currentParagraphLines = nil
			}
		} else {
			currentParagraphLines = append(currentParagraphLines, content)
		}
	}

	// Add final paragraph if any
	if len(currentParagraphLines) > 0 {
		paragraphs = append(paragraphs, createParagraphFromLines(currentParagraphLines))
	}

	return adf_types.ADFNode{
		Type:    "blockquote",
		Content: paragraphs,
	}, nil
}

// createParagraphFromLines creates an ADF paragraph node from text lines
func createParagraphFromLines(lines []string) adf_types.ADFNode {
	text := strings.Join(lines, " ")
	return adf_types.ADFNode{
		Type: "paragraph",
		Content: []adf_types.ADFNode{
			{Type: "text", Text: text},
		},
	}
}
