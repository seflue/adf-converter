package elements

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/converter/elements/inline"
)

// HeadingConverter handles conversion of ADF heading nodes to/from markdown
type HeadingConverter struct{}

func NewHeadingConverter() *HeadingConverter {
	return &HeadingConverter{}
}

func (hc *HeadingConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	// Get heading level (1-6)
	level := node.GetHeadingLevel()
	if level < 1 || level > 6 {
		level = 1 // Default to h1 for invalid levels
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	prefix := strings.Repeat("#", level) + " "
	builder.AppendContent(prefix)

	rendered, err := inline.RenderInlineNodes(node.Content, context)
	if err != nil {
		return converter.EnhancedConversionResult{}, fmt.Errorf("rendering heading content: %w", err)
	}
	// Remove newlines from heading content (headings are single-line in markdown)
	rendered = strings.ReplaceAll(rendered, "\n", " ")
	builder.AppendContent(rendered)

	builder.AppendContent("\n\n")

	return builder.Build(), nil
}

func (hc *HeadingConverter) FromMarkdown(lines []string, startIndex int, _ converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no lines to parse")
	}

	source := []byte(lines[startIndex])
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	headingNode, ok := doc.FirstChild().(*ast.Heading)
	if !ok {
		return adf_types.ADFNode{}, 0, fmt.Errorf("not a valid ATX heading: %q", lines[startIndex])
	}

	level := headingNode.Level

	// Extract inline content: strip leading whitespace + # prefix + separator space.
	trimmed := strings.TrimSpace(lines[startIndex])
	rest := strings.TrimLeft(trimmed[level:], " \t")

	textNodes, err := inline.ParseContent(rest)
	if err != nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("failed to parse heading content: %w", err)
	}

	node := adf_types.ADFNode{
		Type:    adf_types.NodeTypeHeading,
		Attrs:   map[string]interface{}{"level": level},
		Content: textNodes,
	}
	return node, 1, nil
}

func (hc *HeadingConverter) CanParseLine(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level < 1 || level > 6 {
		return false
	}
	// CommonMark: must be followed by space, tab, or end of line.
	return level == len(line) || line[level] == ' ' || line[level] == '\t'
}

func (hc *HeadingConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeHeading)
}

func (hc *HeadingConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (hc *HeadingConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeHeading {
		return fmt.Errorf("node type must be heading, got: %s", node.Type)
	}

	return nil
}
