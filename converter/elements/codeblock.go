package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// CodeBlockConverter handles conversion of ADF codeBlock nodes to/from markdown
type CodeBlockConverter struct{}

func NewCodeBlockConverter() *CodeBlockConverter {
	return &CodeBlockConverter{}
}

func (c *CodeBlockConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := converter.NewEnhancedConversionResultBuilder(converter.MarkdownCodeBlock)

	// Extract code text
	var text string
	if len(node.Content) > 0 {
		text = node.Content[0].Text
	}

	// Extract language
	var language string
	if node.Attrs != nil {
		if lang, ok := node.Attrs["language"].(string); ok && lang != "" {
			language = lang
		}
	}

	// Warn about extra attrs that won't be preserved (v2 feature)
	if node.Attrs != nil {
		for key := range node.Attrs {
			if key != "language" {
				builder.AddWarningf("codeBlock attr not preserved: %s", key)
			}
		}
	}

	// Build fenced code block
	fence := strings.Repeat("`", computeFenceLength(text))
	builder.AppendContent(fence + language + "\n" + text + "\n" + fence + "\n\n")

	return builder.Build(), nil
}

func (c *CodeBlockConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("startIndex out of range")
	}

	firstLine := lines[startIndex]
	trimmed := strings.TrimSpace(firstLine)

	// Count opening fence length and extract language
	fenceLen := 0
	for _, ch := range trimmed {
		if ch == '`' {
			fenceLen++
		} else {
			break
		}
	}
	if fenceLen < 3 {
		return adf_types.ADFNode{}, 0, fmt.Errorf("not a valid code fence: %s", firstLine)
	}

	language := strings.TrimSpace(trimmed[fenceLen:])

	// Collect content lines until closing fence
	var contentLines []string
	closingFound := false
	i := startIndex + 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		// Closing fence: at least fenceLen backticks and nothing else
		if len(line) >= fenceLen && strings.Count(line, "`") == len(line) && len(line) >= fenceLen {
			closingFound = true
			i++
			break
		}
		contentLines = append(contentLines, lines[i])
		i++
	}

	if !closingFound {
		return adf_types.ADFNode{}, 0, fmt.Errorf("unclosed code fence starting at line %d", startIndex)
	}

	consumed := i - startIndex

	// Build ADF node
	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeCodeBlock,
	}

	// Set language attr only if non-empty
	if language != "" {
		node.Attrs = map[string]interface{}{"language": language}
	}

	// Content is always a single text node
	node.Content = []adf_types.ADFNode{
		{Type: adf_types.NodeTypeText, Text: strings.Join(contentLines, "\n")},
	}

	return node, consumed, nil
}

func (c *CodeBlockConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "```")
}

func (c *CodeBlockConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeCodeBlock)
}

func (c *CodeBlockConverter) GetStrategy() converter.ConversionStrategy {
	return converter.MarkdownCodeBlock
}

func (c *CodeBlockConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypeCodeBlock {
		return fmt.Errorf("node type must be codeBlock, got: %s", node.Type)
	}
	return nil
}

// computeFenceLength returns the minimum fence length needed to safely wrap content.
// Scans for the longest consecutive backtick run and returns max(3, longest+1).
func computeFenceLength(content string) int {
	maxRun := 0
	currentRun := 0
	for _, ch := range content {
		if ch == '`' {
			currentRun++
			if currentRun > maxRun {
				maxRun = currentRun
			}
		} else {
			currentRun = 0
		}
	}
	if maxRun >= 3 {
		return maxRun + 1
	}
	return 3
}
