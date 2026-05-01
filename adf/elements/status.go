package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// statusRenderer handles conversion of ADF status nodes to/from markdown
// Format: [status:Text|color]
type statusRenderer struct{}

func NewStatusRenderer() adf.Renderer {
	return &statusRenderer{}
}

func (sc *statusRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if node.Attrs == nil {
		return adf.RenderResult{}, fmt.Errorf("status node missing attrs")
	}

	text, _ := node.Attrs["text"].(string)
	if text == "" {
		return adf.RenderResult{}, fmt.Errorf("status node missing text attribute")
	}

	color, _ := node.Attrs["color"].(string)
	if color == "" {
		return adf.RenderResult{}, fmt.Errorf("status node missing color attribute")
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf("[status:%s|%s]", text, color))
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (sc *statusRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("status is an inline element and should be parsed within parent blocks")
}

