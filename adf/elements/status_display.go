package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// statusDisplayRenderer renders status nodes as bracketed labels wrapped
// in a ColorSpan ("<span style=\"color: #HEX\">[Done]</span>"). Edit-mode's
// "[status:Text|color]" form is rejected for display because Glamour would
// print the literal pipe-suffix. The display/ Glamour bridge parses the
// span and applies the colour via lipgloss.
type statusDisplayRenderer struct{}

// statusPillColors maps ADF status colors to Atlassian-pill hex codes.
// lipgloss only accepts hex (not CSS names), so we resolve here.
var statusPillColors = map[string]string{
	"neutral": "#42526E",
	"purple":  "#5243AA",
	"blue":    "#0052CC",
	"red":     "#DE350B",
	"yellow":  "#FF991F",
	"green":   "#00875A",
}

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

	color, _ := node.Attrs["color"].(string)
	hex, ok := statusPillColors[color]
	if !ok {
		hex = statusPillColors["neutral"]
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf(`<span style="color: %s">[%s]</span>`, hex, text))
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (r *statusDisplayRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("status display renderer is read-only")
}
