package elements

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/elements/internal/inline"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// headingRenderer handles conversion of ADF heading nodes to/from markdown
type headingRenderer struct{}

func NewHeadingRenderer() adf.Renderer {
	return &headingRenderer{}
}

func (hc *headingRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	// Get heading level (1-6)
	level := node.GetHeadingLevel()
	if level < 1 || level > 6 {
		level = 1 // Default to h1 for invalid levels
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)

	prefix := strings.Repeat("#", level) + " "
	builder.AppendContent(prefix)

	rendered, err := inline.RenderInlineNodes(node.Content, context)
	if err != nil {
		return adf.RenderResult{}, fmt.Errorf("rendering heading content: %w", err)
	}
	// Remove newlines from heading content (headings are single-line in markdown)
	rendered = strings.ReplaceAll(rendered, "\n", " ")
	builder.AppendContent(rendered)

	builder.AppendContent("\n\n")

	return builder.Build(), nil
}

func (hc *headingRenderer) FromMarkdown(lines []string, startIndex int, _ adf.ConversionContext) (adf.Node, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("no lines to parse")
	}

	source := []byte(lines[startIndex])
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	headingNode, ok := doc.FirstChild().(*ast.Heading)
	if !ok {
		return adf.Node{}, 0, fmt.Errorf("not a valid ATX heading: %q", lines[startIndex])
	}

	level := headingNode.Level

	// Extract inline content: strip leading whitespace + # prefix + separator space.
	trimmed := strings.TrimSpace(lines[startIndex])
	rest := strings.TrimLeft(trimmed[level:], " \t")

	textNodes, err := inline.ParseContent(rest)
	if err != nil {
		return adf.Node{}, 0, fmt.Errorf("failed to parse heading content: %w", err)
	}

	node := adf.Node{
		Type:    adf.NodeTypeHeading,
		Attrs:   map[string]any{"level": level},
		Content: textNodes,
	}
	return node, 1, nil
}

func (hc *headingRenderer) CanParseLine(line string) bool {
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

func (hc *headingRenderer) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeTypeHeading
}

func (hc *headingRenderer) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (hc *headingRenderer) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be a Node")
	}

	if node.Type != adf.NodeTypeHeading {
		return fmt.Errorf("node type must be heading, got: %s", node.Type)
	}

	return nil
}
