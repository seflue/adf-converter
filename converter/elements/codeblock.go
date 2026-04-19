package elements

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/converter/internal/convresult"
)

// codeBlockConverter handles conversion of ADF codeBlock nodes to/from markdown
type codeBlockConverter struct{}

func NewCodeBlockConverter() element.Converter {
	return &codeBlockConverter{}
}

func (c *codeBlockConverter) ToMarkdown(node adf_types.ADFNode, context element.ConversionContext) (element.EnhancedConversionResult, error) {
	builder := convresult.NewEnhancedConversionResultBuilder(element.MarkdownCodeBlock)

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

func (c *codeBlockConverter) FromMarkdown(lines []string, startIndex int, context element.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("startIndex out of range")
	}

	remainingLines := lines[startIndex:]
	source := []byte(strings.Join(remainingLines, "\n"))

	parser := goldmark.New()
	doc := parser.Parser().Parse(text.NewReader(source))

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		fcb, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			continue
		}

		// Extract language from info string
		var language string
		if lang := fcb.Language(source); len(lang) > 0 {
			language = strings.TrimSpace(string(lang))
		}

		// Extract content — each segment is one line including its trailing newline
		var contentParts []string
		for i := 0; i < fcb.Lines().Len(); i++ {
			seg := fcb.Lines().At(i)
			contentParts = append(contentParts, string(source[seg.Start:seg.Stop]))
		}
		content := strings.TrimSuffix(strings.Join(contentParts, ""), "\n")

		// consumed = opening fence (1) + content lines + closing fence (1)
		consumed := 1 + fcb.Lines().Len() + 1
		if consumed > len(remainingLines) {
			return adf_types.ADFNode{}, 0, fmt.Errorf("unclosed code fence starting at line %d", startIndex)
		}

		node := adf_types.ADFNode{Type: adf_types.NodeTypeCodeBlock}
		if language != "" {
			node.Attrs = map[string]any{"language": language}
		}
		node.Content = []adf_types.ADFNode{
			{Type: adf_types.NodeTypeText, Text: content},
		}

		return node, consumed, nil
	}

	return adf_types.ADFNode{}, 0, fmt.Errorf("not a valid code fence: %s", lines[startIndex])
}

func (c *codeBlockConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "```")
}

func (c *codeBlockConverter) CanHandle(nodeType element.ADFNodeType) bool {
	return nodeType == element.ADFNodeType(adf_types.NodeTypeCodeBlock)
}

func (c *codeBlockConverter) GetStrategy() element.ConversionStrategy {
	return element.MarkdownCodeBlock
}

func (c *codeBlockConverter) ValidateInput(input any) error {
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
