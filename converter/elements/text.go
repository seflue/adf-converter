package elements

import (
	"errors"
	"fmt"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// TextConverter handles bidirectional conversion of text nodes with marks (formatting)
//
// Text nodes are atomic inline elements that can have formatting marks applied:
// - strong: **bold**
// - em: *italic*
// - code: `monospace`
// - link: [text](url)
// - strike: ~~strikethrough~~
// - underline: <u>underline</u>
//
// Note: Text nodes are inline and typically processed within container elements
// (paragraphs, headings, list items). The FromMarkdown direction is primarily
// handled by container converters that call parseInlineContent().
type TextConverter struct{}

func NewTextConverter() *TextConverter {
	return &TextConverter{}
}

func (tc *TextConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if err := tc.ValidateInput(node); err != nil {
		return converter.EnhancedConversionResult{}, err
	}

	text := node.Text

	for _, mark := range node.Marks {
		text = tc.applyMarkToText(text, mark)
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent(text)
	builder.IncrementConverted()

	return builder.Build(), nil
}

func (tc *TextConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, errors.New("text nodes are inline elements - use paragraph/heading converters for parsing")
}

func (tc *TextConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == adf_types.NodeTypeText
}

func (tc *TextConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (tc *TextConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("invalid input type: expected ADFNode, got %T", input)
	}

	if node.Type != adf_types.NodeTypeText {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf_types.NodeTypeText, node.Type)
	}

	if node.Text == "" {
		return fmt.Errorf("text node has empty text content")
	}

	return nil
}

func (tc *TextConverter) applyMarkToText(text string, mark adf_types.ADFMark) string {
	switch mark.Type {
	case adf_types.MarkTypeStrong:
		return fmt.Sprintf("**%s**", text)
	case adf_types.MarkTypeEm:
		return fmt.Sprintf("*%s*", text)
	case adf_types.MarkTypeCode:
		return fmt.Sprintf("`%s`", text)
	case adf_types.MarkTypeLink:
		if href, ok := mark.Attrs["href"].(string); ok {
			return fmt.Sprintf("[%s](%s)", text, href)
		}
		return text
	case adf_types.MarkTypeStrike:
		return fmt.Sprintf("~~%s~~", text)
	case adf_types.MarkTypeUnderline:
		return fmt.Sprintf("<u>%s</u>", text)
	case adf_types.MarkTypeTextColor:
		if color, ok := mark.Attrs["color"].(string); ok {
			return fmt.Sprintf(`<span style="color: %s">%s</span>`, color, text)
		}
		return text
	case adf_types.MarkTypeSubsup:
		if subType, ok := mark.Attrs["type"].(string); ok && subType == "sup" {
			return fmt.Sprintf("<sup>%s</sup>", text)
		}
		return fmt.Sprintf("<sub>%s</sub>", text)
	default:
		return text
	}
}
