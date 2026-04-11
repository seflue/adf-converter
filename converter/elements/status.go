package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

// StatusConverter handles conversion of ADF status nodes to/from markdown
// Format: [status:Text|color]
type StatusConverter struct{}

func NewStatusConverter() *StatusConverter {
	return &StatusConverter{}
}

func (sc *StatusConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		return converter.EnhancedConversionResult{}, fmt.Errorf("status node missing attrs")
	}

	text, _ := node.Attrs["text"].(string)
	if text == "" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("status node missing text attribute")
	}

	color, _ := node.Attrs["color"].(string)
	if color == "" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("status node missing color attribute")
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf("[status:%s|%s]", text, color))
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (sc *StatusConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("status is an inline element and should be parsed within parent blocks")
}

func (sc *StatusConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeStatus)
}

func (sc *StatusConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (sc *StatusConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypeStatus {
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
