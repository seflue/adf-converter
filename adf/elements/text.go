package elements

import (
	"errors"
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// textRenderer handles bidirectional conversion of text nodes with marks (formatting)
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
type textRenderer struct{}

func NewTextRenderer() adf.Renderer {
	return &textRenderer{}
}

func (tc *textRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	text := node.Text

	for _, mark := range node.Marks {
		text = tc.applyMarkToText(text, mark)
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(text)
	builder.IncrementConverted()

	return builder.Build(), nil
}

func (tc *textRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, errors.New("text nodes are inline elements - use paragraph/heading converters for parsing")
}

func (tc *textRenderer) applyMarkToText(text string, mark adf.Mark) string {
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
