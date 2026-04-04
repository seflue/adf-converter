package elements

import (
	"fmt"
	"strconv"
	"time"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// DateConverter handles conversion of ADF date nodes to/from markdown
// Format: [date:2025-04-04]
type DateConverter struct{}

func NewDateConverter() *DateConverter {
	return &DateConverter{}
}

func (dc *DateConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		return converter.EnhancedConversionResult{}, fmt.Errorf("date node missing attrs")
	}

	timestamp, _ := node.Attrs["timestamp"].(string)
	if timestamp == "" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("date node missing timestamp attribute")
	}

	dateStr, err := millisToDate(timestamp)
	if err != nil {
		return converter.EnhancedConversionResult{}, fmt.Errorf("parsing date timestamp %q: %w", timestamp, err)
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf("[date:%s]", dateStr))
	builder.IncrementConverted()
	return builder.Build(), nil
}

// millisToDate converts a Unix-millis string to ISO-8601 date
func millisToDate(millis string) (string, error) {
	ms, err := strconv.ParseInt(millis, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid millis %q: %w", millis, err)
	}
	t := time.Unix(ms/1000, 0).UTC()
	return t.Format("2006-01-02"), nil
}

func (dc *DateConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("date is an inline element and should be parsed within parent blocks")
}

func (dc *DateConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeDate)
}

func (dc *DateConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (dc *DateConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypeDate {
		return fmt.Errorf("node type must be date, got: %s", node.Type)
	}
	if node.Attrs == nil {
		return fmt.Errorf("date node missing attrs")
	}
	if _, ok := node.Attrs["timestamp"].(string); !ok {
		return fmt.Errorf("date node missing timestamp attribute")
	}
	return nil
}
