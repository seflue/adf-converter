package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// statusDisplayRenderer renders status nodes as plain bracketed labels
// ("[Done]"). Edit-mode's "[status:Text|color]" form is rejected for
// display because Glamour would print the literal pipe-suffix. Terminal
// colouring (status pills) is delegated to the separate display/ module.
type statusDisplayRenderer struct{}

// NewStatusDisplayRenderer returns a display-mode renderer for status
// nodes.
func NewStatusDisplayRenderer() adf.Renderer {
	return &statusDisplayRenderer{}
}

func (r *statusDisplayRenderer) ToMarkdown(node adf.Node, _ adf.ConversionContext) (adf.RenderResult, error) {
	if node.Attrs == nil {
		return adf.RenderResult{}, fmt.Errorf("status node missing attrs")
	}

	text, _ := node.Attrs["text"].(string)
	if text == "" {
		return adf.RenderResult{}, fmt.Errorf("status node missing text attribute")
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent("[" + text + "]")
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (r *statusDisplayRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("status display renderer is read-only")
}
