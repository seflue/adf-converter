package elements

import (
	"errors"
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// textConverter handles bidirectional conversion of text nodes with marks (formatting)
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
type textConverter struct{}

func NewTextConverter() adf.Renderer {
	return &textConverter{}
}

func (tc *textConverter) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	if err := tc.ValidateInput(node); err != nil {
		return adf.EnhancedConversionResult{}, err
	}

	text := node.Text

	for _, mark := range node.Marks {
		text = tc.applyMarkToText(text, mark)
	}

	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(text)
	builder.IncrementConverted()

	return builder.Build(), nil
}

func (tc *textConverter) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, errors.New("text nodes are inline elements - use paragraph/heading converters for parsing")
}

func (tc *textConverter) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeTypeText
}

func (tc *textConverter) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (tc *textConverter) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("invalid input type: expected Node, got %T", input)
	}

	if node.Type != adf.NodeTypeText {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf.NodeTypeText, node.Type)
	}

	if node.Text == "" {
		return fmt.Errorf("text node has empty text content")
	}

	return nil
}

func (tc *textConverter) applyMarkToText(text string, mark adf.Mark) string {
	switch mark.Type {
	case adf.MarkTypeStrong:
		return fmt.Sprintf("**%s**", text)
	case adf.MarkTypeEm:
		return fmt.Sprintf("*%s*", text)
	case adf.MarkTypeCode:
		return fmt.Sprintf("`%s`", text)
	case adf.MarkTypeLink:
		if href, ok := mark.Attrs["href"].(string); ok {
			if title, ok := mark.Attrs["title"].(string); ok && title != "" {
				return fmt.Sprintf(`[%s](%s "%s")`, text, href, title)
			}
			return fmt.Sprintf("[%s](%s)", text, href)
		}
		return text
	case adf.MarkTypeStrike:
		return fmt.Sprintf("~~%s~~", text)
	case adf.MarkTypeUnderline:
		return fmt.Sprintf("<u>%s</u>", text)
	case adf.MarkTypeTextColor:
		if color, ok := mark.Attrs["color"].(string); ok {
			return fmt.Sprintf(`<span style="color: %s">%s</span>`, color, text)
		}
		return text
	case adf.MarkTypeSubsup:
		if subType, ok := mark.Attrs["type"].(string); ok && subType == "sup" {
			return fmt.Sprintf("<sup>%s</sup>", text)
		}
		return fmt.Sprintf("<sub>%s</sub>", text)
	default:
		return text
	}
}
