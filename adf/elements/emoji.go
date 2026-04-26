package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// emojiRenderer handles conversion of ADF emoji nodes to/from markdown
type emojiRenderer struct{}

// NewEmojiRenderer creates a new emoji converter
func NewEmojiRenderer() adf.Renderer {
	return &emojiRenderer{}
}

// ToMarkdown converts an ADF emoji node to its text representation.
// Uses text (unicode char) if present, falls back to shortName (e.g. ":white_check_mark:").
func (c *emojiRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if node.Type != adf.NodeTypeEmoji {
		return adf.RenderResult{}, fmt.Errorf("expected emoji node, got %s", node.Type)
	}

	// Extract emoji character from attrs: text (unicode) takes priority, shortName as fallback
	var emojiChar string
	if node.Attrs != nil {
		if text, ok := node.Attrs["text"].(string); ok && text != "" {
			emojiChar = text
		} else if shortName, ok := node.Attrs["shortName"].(string); ok && shortName != "" {
			emojiChar = shortName
		}
	}

	if emojiChar == "" {
		return adf.RenderResult{}, fmt.Errorf("emoji node missing shortName attribute")
	}

	// Return the unicode character directly (no placeholder needed)
	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(emojiChar)
	builder.IncrementConverted()
	return builder.Build(), nil
}

// FromMarkdown is not called for emoji nodes - they're detected by the inline parser
// This method exists to satisfy the Renderer interface
func (c *emojiRenderer) FromMarkdown(lines []string, startLine int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("emoji nodes are handled by inline parser, not block converter")
}

// CanHandle checks if this converter can handle the given node type
func (c *emojiRenderer) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeTypeEmoji
}

// GetStrategy returns the conversion strategy for emoji nodes
func (c *emojiRenderer) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

// ValidateInput validates that the input is a valid emoji node
func (c *emojiRenderer) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("invalid input type: expected Node, got %T", input)
	}

	if node.Type != adf.NodeTypeEmoji {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf.NodeTypeEmoji, node.Type)
	}

	// Validate that emoji has required attributes
	if node.Attrs == nil {
		return fmt.Errorf("emoji node missing attrs")
	}

	// shortName is required per ADF spec
	if shortName, ok := node.Attrs["shortName"].(string); !ok || shortName == "" {
		return fmt.Errorf("emoji node missing or empty shortName attribute")
	}

	return nil
}
