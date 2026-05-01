package elements

import (
	"fmt"
	"strconv"
	"time"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// dateRenderer handles conversion of ADF date nodes to/from markdown
// Format: [date:2025-04-04]
type dateRenderer struct{}

func NewDateRenderer() adf.Renderer {
	return &dateRenderer{}
}

func (dc *dateRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if node.Attrs == nil {
		return adf.RenderResult{}, fmt.Errorf("date node missing attrs")
	}

	timestamp, _ := node.Attrs["timestamp"].(string)
	if timestamp == "" {
		return adf.RenderResult{}, fmt.Errorf("date node missing timestamp attribute")
	}

	dateStr, err := millisToDate(timestamp)
	if err != nil {
		return adf.RenderResult{}, fmt.Errorf("parsing date timestamp %q: %w", timestamp, err)
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
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

func (dc *dateRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("date is an inline element and should be parsed within parent blocks")
}

