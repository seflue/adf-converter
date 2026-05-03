package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// inlineCardDisplayRenderer emits inlineCard as a single Markdown autolink
// (e.g. <https://example.com>). Edit-mode wraps URLs as [url](url) which
// causes Glamour to render the URL twice — once as label, once as link
// target. The autolink form prints exactly once and stays stable across
// Glamour styles.
type inlineCardDisplayRenderer struct{}

// NewInlineCardDisplayRenderer returns a display-mode renderer for
// inlineCard nodes. Output is a Goldmark autolink for nodes with a URL,
// or [InlineCard] as fallback for data-only / missing-attrs nodes.
func NewInlineCardDisplayRenderer() adf.Renderer {
	return &inlineCardDisplayRenderer{}
}

func (r *inlineCardDisplayRenderer) ToMarkdown(node adf.Node, _ adf.ConversionContext) (adf.RenderResult, error) {
	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)

	url := ""
	if node.Attrs != nil {
		url, _ = node.Attrs["url"].(string)
	}

	if url == "" {
		builder.AppendContent("[InlineCard]")
		return builder.Build(), nil
	}

	builder.AppendContent("<" + url + ">")
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (r *inlineCardDisplayRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("inlineCard display renderer is read-only")
}
