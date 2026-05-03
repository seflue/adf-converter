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
type textRenderer struct {
	pipeline markPipeline
}

func NewTextRenderer() adf.Renderer {
	return &textRenderer{pipeline: editMarkPipeline}
}

func (tc *textRenderer) ToMarkdown(node adf.Node, _ adf.ConversionContext) (adf.RenderResult, error) {
	text := node.Text

	for _, mark := range node.Marks {
		text = tc.pipeline.apply(text, mark)
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(text)
	builder.IncrementConverted()

	return builder.Build(), nil
}

func (tc *textRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, errors.New("text nodes are inline elements - use paragraph/heading converters for parsing")
}

// markRenderFunc renders a single mark by wrapping its already-rendered
// inner text. Returning the input verbatim signals "drop this mark".
type markRenderFunc func(text string, mark adf.Mark) string

// markPipeline maps ADF mark types to render functions. Marks not in the
// map fall through to the default unknown-mark behaviour (text returned
// unchanged).
type markPipeline map[adf.MarkType]markRenderFunc

func (p markPipeline) apply(text string, mark adf.Mark) string {
	if fn, ok := p[mark.Type]; ok {
		return fn(text, mark)
	}
	return text
}

// editMarkPipeline is the canonical edit-mode rendering of all supported
// marks. Display-mode pipelines are constructed by overlaying overrides
// on top of this base.
var editMarkPipeline = markPipeline{
	adf.MarkTypeStrong:    func(t string, _ adf.Mark) string { return "**" + t + "**" },
	adf.MarkTypeEm:        func(t string, _ adf.Mark) string { return "*" + t + "*" },
	adf.MarkTypeCode:      func(t string, _ adf.Mark) string { return "`" + t + "`" },
	adf.MarkTypeStrike:    func(t string, _ adf.Mark) string { return "~~" + t + "~~" },
	adf.MarkTypeUnderline: func(t string, _ adf.Mark) string { return "<u>" + t + "</u>" },
	adf.MarkTypeLink:      renderLinkMark,
	adf.MarkTypeTextColor: renderTextColorMarkEdit,
	adf.MarkTypeSubsup:    renderSubsupMarkEdit,
}

func renderLinkMark(text string, mark adf.Mark) string {
	href, ok := mark.Attrs["href"].(string)
	if !ok {
		return text
	}
	if title, ok := mark.Attrs["title"].(string); ok && title != "" {
		return fmt.Sprintf(`[%s](%s "%s")`, text, href, title)
	}
	return "[" + text + "](" + href + ")"
}

func renderTextColorMarkEdit(text string, mark adf.Mark) string {
	color, ok := mark.Attrs["color"].(string)
	if !ok {
		return text
	}
	return fmt.Sprintf(`<span style="color: %s">%s</span>`, color, text)
}

func renderSubsupMarkEdit(text string, mark adf.Mark) string {
	if subType, ok := mark.Attrs["type"].(string); ok && subType == "sup" {
		return "<sup>" + text + "</sup>"
	}
	return "<sub>" + text + "</sub>"
}

// withOverrides returns a new pipeline that copies p and overlays the
// given overrides. The base pipeline is not mutated so callers can build
// independent variants from a shared edit-mode baseline.
func (p markPipeline) withOverrides(overrides markPipeline) markPipeline {
	out := make(markPipeline, len(p)+len(overrides))
	for k, v := range p {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}
