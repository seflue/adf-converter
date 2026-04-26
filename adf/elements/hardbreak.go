package elements

import (
	"errors"
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// hardBreakConverter handles bidirectional conversion of hard break nodes
//
// Hard breaks are atomic inline elements that represent line breaks within
// block-level elements (paragraphs, headings, list items).
//
// In Markdown: represented as "\n" (newline)
// In ADF: represented as { type: "hardBreak" }
//
// Note: Hard breaks are inline and typically processed within container elements.
// The FromMarkdown direction is primarily handled by container converters that
// recognize newlines within their content.
type hardBreakConverter struct{}

// NewHardBreakConverter creates a new hard break converter instance
func NewHardBreakConverter() adf.Renderer {
	return &hardBreakConverter{}
}

// ToMarkdown converts an ADF hard break node to a Markdown newline
func (hc *hardBreakConverter) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	// Validate input
	if err := hc.ValidateInput(node); err != nil {
		return adf.EnhancedConversionResult{}, err
	}

	// Hard break is simply a newline character
	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)
	builder.AppendContent("\n")
	builder.IncrementConverted()

	return builder.Build(), nil
}

// FromMarkdown parses Markdown into an ADF hard break node
//
// Hard break nodes are inline elements, so this method returns an error indicating
// that parsing is handled by container converters (paragraph, heading, etc.).
// Container converters detect newlines within their content and create hard break nodes.
func (hc *hardBreakConverter) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, errors.New("hardBreak nodes are inline elements - use paragraph/heading converters for parsing")
}

// CanHandle returns true if this converter can handle the given node type
func (hc *hardBreakConverter) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeTypeHardBreak
}

// GetStrategy returns the conversion strategy for hard break nodes
func (hc *hardBreakConverter) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

// ValidateInput validates that the input node is a valid hard break node
func (hc *hardBreakConverter) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("invalid input type: expected Node, got %T", input)
	}

	if node.Type != adf.NodeTypeHardBreak {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf.NodeTypeHardBreak, node.Type)
	}

	// Hard break nodes should not have content or text
	if len(node.Content) > 0 {
		return fmt.Errorf("hardBreak node should not have content")
	}

	if node.Text != "" {
		return fmt.Errorf("hardBreak node should not have text")
	}

	return nil
}
