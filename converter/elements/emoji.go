package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/internal/convresult"
)

// emojiConverter handles conversion of ADF emoji nodes to/from markdown
type emojiConverter struct{}

// NewEmojiConverter creates a new emoji converter
func NewEmojiConverter() converter.ElementConverter {
	return &emojiConverter{}
}

// ToMarkdown converts an ADF emoji node to its text representation.
// Uses text (unicode char) if present, falls back to shortName (e.g. ":white_check_mark:").
func (c *emojiConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Type != adf_types.NodeTypeEmoji {
		return converter.EnhancedConversionResult{}, fmt.Errorf("expected emoji node, got %s", node.Type)
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
		return converter.EnhancedConversionResult{}, fmt.Errorf("emoji node missing shortName attribute")
	}

	// Return the unicode character directly (no placeholder needed)
	builder := convresult.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent(emojiChar)
	builder.IncrementConverted()
	return builder.Build(), nil
}

// FromMarkdown is not called for emoji nodes - they're detected by the inline parser
// This method exists to satisfy the ElementConverter interface
func (c *emojiConverter) FromMarkdown(lines []string, startLine int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("emoji nodes are handled by inline parser, not block converter")
}

// CanHandle checks if this converter can handle the given node type
func (c *emojiConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == adf_types.NodeTypeEmoji
}

// GetStrategy returns the conversion strategy for emoji nodes
func (c *emojiConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

// ValidateInput validates that the input is a valid emoji node
func (c *emojiConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("invalid input type: expected ADFNode, got %T", input)
	}

	if node.Type != adf_types.NodeTypeEmoji {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf_types.NodeTypeEmoji, node.Type)
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
