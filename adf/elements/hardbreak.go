package elements

import (
	"errors"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// hardBreakRenderer handles bidirectional conversion of hard break nodes
//
// Hard breaks are atomic inline elements that represent line breaks within
// block-level elements (paragraphs, headings, list items).
//
// In Markdown: represented as "\n" (newline)
// In ADF: represented as { type: "hardBreak" }
//
// Note: Hard breaks are inline and typically processed within container elements.
// The FromMarkdown direction is primarily handled by container converters that
// recognize newlines within their content.
type hardBreakRenderer struct{}

// NewHardBreakRenderer creates a new hard break converter instance
func NewHardBreakRenderer() adf.Renderer {
	return &hardBreakRenderer{}
}

// ToMarkdown converts an ADF hard break node to a Markdown newline
func (hc *hardBreakRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	// Hard break is simply a newline character
	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent("\n")
	builder.IncrementConverted()

	return builder.Build(), nil
}

// FromMarkdown parses Markdown into an ADF hard break node
//
// Hard break nodes are inline elements, so this method returns an error indicating
// that parsing is handled by container converters (paragraph, heading, etc.).
// Container converters detect newlines within their content and create hard break nodes.
func (hc *hardBreakRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, errors.New("hardBreak nodes are inline elements - use paragraph/heading converters for parsing")
}

