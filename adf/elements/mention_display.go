package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// mentionDisplayRenderer emits mentions as plain @Name text for read-only
// display mode. Edit-mode link syntax is dropped because Glamour would
// render the link target verbatim, producing visible accountid noise.
type mentionDisplayRenderer struct{}

// NewMentionDisplayRenderer returns a display-mode renderer for mention
// nodes. Output is plain @Name text; terminal styling is the consumer's
// responsibility (e.g. via the separate display/ module's Glamour bridge).
func NewMentionDisplayRenderer() adf.Renderer {
	return &mentionDisplayRenderer{}
}

func (r *mentionDisplayRenderer) ToMarkdown(node adf.Node, _ adf.ConversionContext) (adf.RenderResult, error) {
	if node.Attrs == nil {
		return adf.RenderResult{}, fmt.Errorf("mention node missing attrs")
	}

	id, _ := node.Attrs["id"].(string)
	if id == "" {
		return adf.RenderResult{}, fmt.Errorf("mention node missing id attribute")
	}

	text, _ := node.Attrs["text"].(string)
	if text == "" {
		text = "@" + id
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(text)
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (r *mentionDisplayRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("mention display renderer is read-only")
}
