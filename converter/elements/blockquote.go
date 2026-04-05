package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter/elements/inline"
	"adf-converter/converter/internal"
)

// BlockquoteConverter implements markdown blockquote conversion for ADF blockquote nodes
type BlockquoteConverter struct{}

func NewBlockquoteConverter() *BlockquoteConverter {
	return &BlockquoteConverter{}
}

func (bc *BlockquoteConverter) ToMarkdown(node adf_types.ADFNode, context ConversionContext) (EnhancedConversionResult, error) {
	if node.Type != "blockquote" {
		return EnhancedConversionResult{}, fmt.Errorf("blockquote converter can only handle blockquote nodes, got: %s", node.Type)
	}

	builder := NewEnhancedConversionResultBuilder(MarkdownBlockquote)

	if context.PreserveAttrs && node.Attrs != nil {
		builder.PreserveAttributes(node.Attrs)
	}

	if len(node.Content) == 0 {
		quotePrefix := bc.createQuotePrefix(context.NestedLevel)
		builder.AppendLine(quotePrefix + " ")
		builder.IncrementConverted()
		return builder.Build(), nil
	}

	for i, contentNode := range node.Content {
		if i > 0 {
			quotePrefix := bc.createQuotePrefix(context.NestedLevel)
			builder.AppendLine(quotePrefix + " ")
		}

		switch contentNode.Type {
		case "paragraph":
			text := bc.convertParagraphToMarkdown(contentNode)
			quotePrefix := bc.createQuotePrefix(context.NestedLevel)

			if strings.TrimSpace(text) == "" {
				builder.AppendLine(quotePrefix + " ")
			} else {
				builder.AppendLine(fmt.Sprintf("%s %s", quotePrefix, text))
			}

		case "blockquote":
			nestedContext := context
			nestedContext.NestedLevel++

			nestedResult, err := bc.ToMarkdown(contentNode, nestedContext)
			if err != nil {
				builder.AddWarningf("Failed to convert nested blockquote: %v", err)
				continue
			}

			builder.AppendContent(nestedResult.Content)
			builder.AddConverted(nestedResult.ElementsConverted)

		default:
			text := bc.extractTextContent(contentNode)
			quotePrefix := bc.createQuotePrefix(context.NestedLevel)

			if strings.TrimSpace(text) == "" {
				builder.AppendLine(quotePrefix + " ")
			} else {
				builder.AppendLine(fmt.Sprintf("%s %s", quotePrefix, text))
			}
		}

		builder.IncrementConverted()
	}

	result := builder.Build()

	if context.PreserveAttrs && node.Attrs != nil && len(node.Attrs) > 0 {
		wrappedMarkdown, err := bc.wrapBlockquoteWithXML(result.Content, node.Attrs, context.NestedLevel)
		if err != nil {
			return CreateErrorResult(err.Error(), MarkdownBlockquote), err
		}
		result.Content = wrappedMarkdown
	} else {
		// For plain markdown blockquotes, trim trailing newline for cleaner output
		result.Content = strings.TrimSuffix(result.Content, "\n")
	}

	return result, nil
}

func (bc *BlockquoteConverter) extractTextContent(node adf_types.ADFNode) string {
	var content strings.Builder

	switch node.Type {
	case "text":
		content.WriteString(node.Text)
	case "paragraph":
		for _, child := range node.Content {
			childText := bc.extractTextContent(child)
			content.WriteString(childText)
		}
	default:
		for _, child := range node.Content {
			childText := bc.extractTextContent(child)
			content.WriteString(childText)
		}
	}

	return content.String()
}

func (bc *BlockquoteConverter) convertParagraphToMarkdown(paragraph adf_types.ADFNode) string {
	var result strings.Builder

	for _, textNode := range paragraph.Content {
		if textNode.Type == "text" {
			text := textNode.Text

			for _, mark := range textNode.Marks {
				switch mark.Type {
				case "strong":
					text = "**" + text + "**"
				case "em":
					text = "*" + text + "*"
				case "code":
					text = "`" + text + "`"
				}
			}

			result.WriteString(text)
		}
	}

	return result.String()
}

func (bc *BlockquoteConverter) createQuotePrefix(nestedLevel int) string {
	if nestedLevel <= 0 {
		nestedLevel = 1
	}

	var prefix strings.Builder
	for i := 0; i < nestedLevel; i++ {
		if i > 0 {
			prefix.WriteString(" ")
		}
		prefix.WriteString(">")
	}

	return prefix.String()
}

func (bc *BlockquoteConverter) FromMarkdown(lines []string, startIndex int, context ConversionContext) (adf_types.ADFNode, int, error) {
	emptyNode := adf_types.ADFNode{Type: "blockquote", Content: []adf_types.ADFNode{}}

	if startIndex >= len(lines) {
		return emptyNode, 0, nil
	}

	firstLine := strings.TrimSpace(lines[startIndex])

	// XML-wrapped blockquote
	if strings.HasPrefix(firstLine, "<blockquote") {
		node, consumed, err := parseXMLBlockquote(lines[startIndex:])
		if err != nil {
			return adf_types.ADFNode{}, 0, fmt.Errorf("parsing XML-wrapped blockquote: %w", err)
		}
		if node == nil {
			return emptyNode, consumed, nil
		}
		return *node, consumed, nil
	}

	// Plain markdown blockquote
	consumed := countBlockquoteLines(lines, startIndex)
	if consumed == 0 {
		return emptyNode, 0, nil
	}

	node, err := parseMarkdownBlockquote(lines[startIndex : startIndex+consumed])
	if err != nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("parsing markdown blockquote: %w", err)
	}
	return node, consumed, nil
}

// countBlockquoteLines counts consecutive blockquote lines starting from startIndex.
// Empty lines between > lines are included; trailing empty lines are not.
func countBlockquoteLines(lines []string, startIndex int) int {
	lastQuoteLine := -1
	for i := startIndex; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, ">") {
			lastQuoteLine = i - startIndex + 1
		} else if trimmed == "" && lastQuoteLine > 0 {
			// Empty line — might be between blockquote paragraphs, keep scanning
			continue
		} else {
			break
		}
	}
	if lastQuoteLine < 0 {
		return 0
	}
	return lastQuoteLine
}

func (bc *BlockquoteConverter) CanHandle(nodeType ADFNodeType) bool {
	return nodeType == NodeBlockquote
}

func (bc *BlockquoteConverter) GetStrategy() ConversionStrategy {
	return MarkdownBlockquote
}

func (bc *BlockquoteConverter) ValidateInput(input interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	switch v := input.(type) {
	case adf_types.ADFNode:
		if v.Type != "blockquote" {
			return fmt.Errorf("ADF node must be of type 'blockquote', got: %s", v.Type)
		}
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("markdown input cannot be empty")
		}
		return nil
	default:
		return fmt.Errorf("input must be adf_types.ADFNode or string, got: %T", input)
	}
}

func (bc *BlockquoteConverter) wrapBlockquoteWithXML(markdownBlockquote string, attrs map[string]interface{}, nestedLevel int) (string, error) {
	var xmlBuilder strings.Builder

	xmlBuilder.WriteString("<blockquote")

	if attrs != nil {
		for key, value := range attrs {
			switch key {
			case "localId":
				if localIdStr, ok := value.(string); ok {
					xmlBuilder.WriteString(fmt.Sprintf(` localId="%s"`, localIdStr))
				}
			case "level":
				if levelInt, ok := value.(int); ok {
					xmlBuilder.WriteString(fmt.Sprintf(` level="%d"`, levelInt))
				}
			}
		}
	}

	xmlBuilder.WriteString(">\n")

	xmlBuilder.WriteString(markdownBlockquote)

	if !strings.HasSuffix(markdownBlockquote, "\n") {
		xmlBuilder.WriteString("\n")
	}
	xmlBuilder.WriteString("</blockquote>")

	return xmlBuilder.String(), nil
}

// parseXMLBlockquote parses XML-formatted blockquote from markdown lines
func parseXMLBlockquote(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, fmt.Errorf("no lines to parse")
	}

	// Find the opening and closing tags
	startIdx := -1
	endIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<blockquote") {
			startIdx = i
		} else if strings.HasPrefix(trimmed, "</blockquote>") {
			endIdx = i
			break
		}
	}

	if startIdx == -1 || endIdx == -1 {
		return nil, 1, fmt.Errorf("malformed XML blockquote: missing opening or closing tag")
	}

	// Parse XML attributes from opening tag
	openTag := strings.TrimSpace(lines[startIdx])
	attrs := internal.ParseXMLAttributes(openTag)

	var markdownLines []string
	for i := startIdx + 1; i < endIdx; i++ {
		markdownLines = append(markdownLines, lines[i])
	}

	// Parse the markdown blockquote content into ADF
	blockquoteNode, err := parseMarkdownBlockquote(markdownLines)
	if err != nil {
		return nil, endIdx - startIdx + 1, fmt.Errorf("failed to parse markdown blockquote content: %w", err)
	}

	// Add the preserved XML attributes to the blockquote node
	if blockquoteNode.Attrs == nil {
		blockquoteNode.Attrs = make(map[string]interface{})
	}
	for key, value := range attrs {
		blockquoteNode.Attrs[key] = value
	}

	return &blockquoteNode, endIdx - startIdx + 1, nil
}

// parseMarkdownBlockquote parses standard markdown blockquote (> prefix) into ADF
func parseMarkdownBlockquote(lines []string) (adf_types.ADFNode, error) {
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

	// Parse inline content (bold, italic, links, etc.)
	inlineContent, err := inline.ParseContent(text)
	if err != nil || len(inlineContent) == 0 {
		// Fallback to plain text on error
		inlineContent = []adf_types.ADFNode{{Type: "text", Text: text}}
	}

	return adf_types.ADFNode{
		Type:    "paragraph",
		Content: inlineContent,
	}
}
