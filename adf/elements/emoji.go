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

