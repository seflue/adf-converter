package elements

import (
	"fmt"
	"strconv"
	"time"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// dateConverter handles conversion of ADF date nodes to/from markdown
// Format: [date:2025-04-04]
type dateConverter struct{}

func NewDateConverter() adf.Renderer {
	return &dateConverter{}
}

func (dc *dateConverter) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		return adf.EnhancedConversionResult{}, fmt.Errorf("date node missing attrs")
	}

	timestamp, _ := node.Attrs["timestamp"].(string)
	if timestamp == "" {
		return adf.EnhancedConversionResult{}, fmt.Errorf("date node missing timestamp attribute")
	}

	dateStr, err := millisToDate(timestamp)
	if err != nil {
		return adf.EnhancedConversionResult{}, fmt.Errorf("parsing date timestamp %q: %w", timestamp, err)
	}

	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)
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

func (dc *dateConverter) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("date is an inline element and should be parsed within parent blocks")
}

func (dc *dateConverter) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypeDate)
}

func (dc *dateConverter) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (dc *dateConverter) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be an Node")
	}
	if node.Type != adf.NodeTypeDate {
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
