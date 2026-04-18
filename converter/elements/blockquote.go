package elements

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/elements/internal/inline"
	"github.com/seflue/adf-converter/converter/elements/internal/lists"
	"github.com/seflue/adf-converter/converter/internal"
)

// blockquoteConverter implements markdown blockquote conversion for ADF blockquote nodes
type blockquoteConverter struct{}

func NewBlockquoteConverter() converter.ElementConverter {
	return &blockquoteConverter{}
}

func (bc *blockquoteConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (EnhancedConversionResult, error) {
	if node.Type != "blockquote" {
		return EnhancedConversionResult{}, fmt.Errorf("blockquote converter can only handle blockquote nodes, got: %s", node.Type)
	}

	builder := NewEnhancedConversionResultBuilder(MarkdownBlockquote)

	if bc.shouldPreserveAttrs(context, node) {
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

		case "bulletList":
			listResult, err := NewBulletListConverter().ToMarkdown(contentNode, context)
			if err != nil {
				builder.AddWarningf("Failed to convert bulletList: %v", err)
				continue
			}
			builder.AppendContent(bc.prefixLines(listResult.Content, context.NestedLevel) + "\n")

		case "orderedList":
			listResult, err := NewOrderedListConverter().ToMarkdown(contentNode, context)
			if err != nil {
				builder.AddWarningf("Failed to convert orderedList: %v", err)
				continue
			}
			builder.AppendContent(bc.prefixLines(listResult.Content, context.NestedLevel) + "\n")

		case "codeBlock":
			codeResult, err := NewCodeBlockConverter().ToMarkdown(contentNode, context)
			if err != nil {
				builder.AddWarningf("Failed to convert codeBlock: %v", err)
				continue
			}
			builder.AppendContent(bc.prefixLines(codeResult.Content, context.NestedLevel) + "\n")

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

	if bc.shouldPreserveAttrs(context, node) {
		result.Content = bc.wrapBlockquoteWithXML(result.Content, node.Attrs, context.NestedLevel)
	} else {
		// Trim trailing newline, then add block-level spacing
		result.Content = strings.TrimSuffix(result.Content, "\n") + "\n\n"
	}

	return result, nil
}

func (bc *blockquoteConverter) extractTextContent(node adf_types.ADFNode) string {
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

func (bc *blockquoteConverter) convertParagraphToMarkdown(paragraph adf_types.ADFNode) string {
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

// prefixLines adds a blockquote prefix ("> ") to each line of multi-line content.
// Empty lines get only the bare prefix without trailing space.
func (bc *blockquoteConverter) prefixLines(content string, nestedLevel int) string {
	prefix := bc.createQuotePrefix(nestedLevel) + " "
	trimmed := strings.TrimRight(content, "\n")
	lines := strings.Split(trimmed, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, strings.TrimRight(prefix, " "))
		} else {
			result = append(result, prefix+line)
		}
	}
	return strings.Join(result, "\n")
}

func (bc *blockquoteConverter) createQuotePrefix(nestedLevel int) string {
	depth := nestedLevel + 1
	var prefix strings.Builder
	for i := 0; i < depth; i++ {
		if i > 0 {
			prefix.WriteString(" ")
		}
		prefix.WriteString(">")
	}
	return prefix.String()
}

func (bc *blockquoteConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
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

func (bc *blockquoteConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "<blockquote") || strings.HasPrefix(line, ">")
}

func (bc *blockquoteConverter) CanHandle(nodeType ADFNodeType) bool {
	return nodeType == NodeBlockquote
}

func (bc *blockquoteConverter) GetStrategy() converter.ConversionStrategy {
	return MarkdownBlockquote
}

func (bc *blockquoteConverter) ValidateInput(input interface{}) error {
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

func (bc *blockquoteConverter) shouldPreserveAttrs(context converter.ConversionContext, node adf_types.ADFNode) bool {
	return context.PreserveAttrs && len(node.Attrs) > 0
}

func (bc *blockquoteConverter) wrapBlockquoteWithXML(markdownBlockquote string, attrs map[string]interface{}, nestedLevel int) string {
	var xmlBuilder strings.Builder

	xmlBuilder.WriteString("<blockquote")

	for key, value := range attrs {
		switch key {
		case "localId":
			if localIdStr, ok := value.(string); ok {
				fmt.Fprintf(&xmlBuilder, ` localId="%s"`, localIdStr)
			}
		case "level":
			if levelInt, ok := value.(int); ok {
				fmt.Fprintf(&xmlBuilder, ` level="%d"`, levelInt)
			}
		}
	}

	xmlBuilder.WriteString(">\n")

	xmlBuilder.WriteString(markdownBlockquote)

	if !strings.HasSuffix(markdownBlockquote, "\n") {
		xmlBuilder.WriteString("\n")
	}
	xmlBuilder.WriteString("</blockquote>")

	return xmlBuilder.String()
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

// parseMarkdownBlockquote parses standard markdown blockquote (> prefix) into ADF using goldmark.
// Only strips one level of > prefix. Remaining > characters (nested blockquotes) stay as literal
// text because ADF does not allow nested blockquote nodes.
//
// Note: goldmark splits a blank-line-separated blockquote ("> a\n\n> b") into multiple sibling
// Blockquote nodes at the document level. We collect paragraphs from all of them.
func parseMarkdownBlockquote(lines []string) (adf_types.ADFNode, error) {
	source := []byte(strings.Join(lines, "\n"))
	parser := goldmark.New()
	doc := parser.Parser().Parse(text.NewReader(source))

	paragraphs := []adf_types.ADFNode{}
	for topLevel := doc.FirstChild(); topLevel != nil; topLevel = topLevel.NextSibling() {
		bq, ok := topLevel.(*ast.Blockquote)
		if !ok {
			continue
		}
		for child := bq.FirstChild(); child != nil; child = child.NextSibling() {
			switch n := child.(type) {
			case *ast.Paragraph:
				para, err := convertBlockquoteParagraph(n, source)
				if err != nil {
					return adf_types.ADFNode{}, err
				}
				paragraphs = append(paragraphs, para)
			case *ast.Blockquote:
				// Flatten nested blockquote — ADF forbids nesting.
				// The stripped > becomes literal text in a paragraph.
				para := flattenNestedBlockquote(n, source)
				paragraphs = append(paragraphs, para)
			case *ast.List:
				listNode, err := lists.ConvertListNode(n, source, nil)
				if err != nil {
					return adf_types.ADFNode{}, fmt.Errorf("converting list in blockquote: %w", err)
				}
				paragraphs = append(paragraphs, listNode)
			case *ast.FencedCodeBlock:
				paragraphs = append(paragraphs, convertBlockquoteCodeBlock(n, source))
			}
		}
	}

	return adf_types.ADFNode{Type: "blockquote", Content: paragraphs}, nil
}

// convertBlockquoteCodeBlock converts a goldmark FencedCodeBlock inside a blockquote to an ADF codeBlock node.
func convertBlockquoteCodeBlock(n *ast.FencedCodeBlock, source []byte) adf_types.ADFNode {
	language := strings.TrimSpace(string(n.Language(source)))

	var lines []string
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		lines = append(lines, string(source[seg.Start:seg.Stop]))
	}
	content := strings.TrimRight(strings.Join(lines, ""), "\n")

	node := adf_types.ADFNode{Type: adf_types.NodeTypeCodeBlock}
	if language != "" {
		node.Attrs = map[string]interface{}{"language": language}
	}
	node.Content = []adf_types.ADFNode{{Type: adf_types.NodeTypeText, Text: content}}
	return node
}

// convertBlockquoteParagraph converts a goldmark Paragraph node inside a blockquote to an ADF paragraph.
// Multiple lines are joined with a space to match the previous behaviour.
func convertBlockquoteParagraph(para *ast.Paragraph, source []byte) (adf_types.ADFNode, error) {
	lineTexts := make([]string, 0, para.Lines().Len())
	for i := 0; i < para.Lines().Len(); i++ {
		seg := para.Lines().At(i)
		lineTexts = append(lineTexts, strings.TrimSpace(string(source[seg.Start:seg.Stop])))
	}
	rawText := strings.Join(lineTexts, " ")

	inlineContent, err := inline.ParseContent(rawText)
	if err != nil || len(inlineContent) == 0 {
		inlineContent = []adf_types.ADFNode{{Type: "text", Text: rawText}}
	}

	return adf_types.ADFNode{Type: "paragraph", Content: inlineContent}, nil
}

// flattenNestedBlockquote converts a goldmark nested Blockquote to an ADF paragraph with
// literal > text — ADF does not allow blockquote nodes nested inside blockquote nodes.
func flattenNestedBlockquote(bq *ast.Blockquote, source []byte) adf_types.ADFNode {
	var parts []string
	for child := bq.FirstChild(); child != nil; child = child.NextSibling() {
		if para, ok := child.(*ast.Paragraph); ok {
			for i := 0; i < para.Lines().Len(); i++ {
				seg := para.Lines().At(i)
				parts = append(parts, "> "+strings.TrimSpace(string(source[seg.Start:seg.Stop])))
			}
		}
	}
	rawText := strings.Join(parts, " ")
	return adf_types.ADFNode{
		Type:    "paragraph",
		Content: []adf_types.ADFNode{{Type: "text", Text: rawText}},
	}
}
