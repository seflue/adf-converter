package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// statusConverter handles conversion of ADF status nodes to/from markdown
// Format: [status:Text|color]
type statusConverter struct{}

func NewStatusConverter() adf.Renderer {
	return &statusConverter{}
}

func (sc *statusConverter) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		return adf.EnhancedConversionResult{}, fmt.Errorf("status node missing attrs")
	}

	text, _ := node.Attrs["text"].(string)
	if text == "" {
		return adf.EnhancedConversionResult{}, fmt.Errorf("status node missing text attribute")
	}

	color, _ := node.Attrs["color"].(string)
	if color == "" {
		return adf.EnhancedConversionResult{}, fmt.Errorf("status node missing color attribute")
	}

	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf("[status:%s|%s]", text, color))
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (sc *statusConverter) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("status is an inline element and should be parsed within parent blocks")
}

func (sc *statusConverter) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypeStatus)
}

func (sc *statusConverter) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (sc *statusConverter) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be an Node")
	}
	if node.Type != adf.NodeTypeStatus {
		return fmt.Errorf("node type must be status, got: %s", node.Type)
	}
	if node.Attrs == nil {
		return fmt.Errorf("status node missing attrs")
	}
	if _, ok := node.Attrs["text"].(string); !ok {
		return fmt.Errorf("status node missing text attribute")
	}
	if _, ok := node.Attrs["color"].(string); !ok {
		return fmt.Errorf("status node missing color attribute")
	}
	return nil
}
