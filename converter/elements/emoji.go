package elements

import (
	"fmt"
	"strings"

	"github.com/forPelevin/gomoji"
	"adf-converter/adf_types"
	"adf-converter/converter"
)

// EmojiConverter handles conversion of ADF emoji nodes to/from markdown
type EmojiConverter struct{}

// NewEmojiConverter creates a new emoji converter
func NewEmojiConverter() *EmojiConverter {
	return &EmojiConverter{}
}

// ToMarkdown converts an ADF emoji node to unicode character
// Example: {"type": "emoji", "attrs": {"text": "✅"}} -> "✅"
func (c *EmojiConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Type != adf_types.NodeTypeEmoji {
		return converter.EnhancedConversionResult{}, fmt.Errorf("expected emoji node, got %s", node.Type)
	}

	// Extract unicode character from attrs
	var emojiChar string
	if node.Attrs != nil {
		// Try "text" attribute first (contains the actual unicode character)
		if text, ok := node.Attrs["text"].(string); ok && text != "" {
			emojiChar = text
		}
	}

	if emojiChar == "" {
		// Fallback: try to reconstruct from ID if text is missing
		if node.Attrs != nil {
			if id, ok := node.Attrs["id"].(string); ok && id != "" {
				// Try to find emoji by searching gomoji's database
				// This is a fallback case for incomplete ADF nodes
				allEmojis := gomoji.AllEmojis()
				for _, emoji := range allEmojis {
					// Compare code point (remove "U+" prefix and compare)
					codePoint := strings.TrimPrefix(emoji.CodePoint, "U+")
					if strings.EqualFold(codePoint, id) {
						emojiChar = emoji.Character
						break
					}
				}
			}
		}
	}

	if emojiChar == "" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("emoji node missing text attribute")
	}

	// Return the unicode character directly (no placeholder needed)
	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent(emojiChar)
	builder.IncrementConverted()
	return builder.Build(), nil
}

// FromMarkdown is not called for emoji nodes - they're detected by the inline parser
// This method exists to satisfy the ElementConverter interface
func (c *EmojiConverter) FromMarkdown(lines []string, startLine int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("emoji nodes are handled by inline parser, not block converter")
}

// CanHandle checks if this converter can handle the given node type
func (c *EmojiConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == adf_types.NodeTypeEmoji
}

// GetStrategy returns the conversion strategy for emoji nodes
func (c *EmojiConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

// ValidateInput validates that the input is a valid emoji node
func (c *EmojiConverter) ValidateInput(input interface{}) error {
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

	// Must have at least text attribute
	if text, ok := node.Attrs["text"].(string); !ok || text == "" {
		return fmt.Errorf("emoji node missing or empty text attribute")
	}

	return nil
}
