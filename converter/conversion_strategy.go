package converter

import "github.com/seflue/adf-converter/converter/element"

// ConversionStrategy is an alias for element.ConversionStrategy.
type ConversionStrategy = element.ConversionStrategy

const (
	StandardMarkdown = element.StandardMarkdown
	HTMLWrapped      = element.HTMLWrapped
	Placeholder      = element.Placeholder

	MarkdownTable      = element.MarkdownTable
	MarkdownTaskList   = element.MarkdownTaskList
	MarkdownBlockquote = element.MarkdownBlockquote
	MarkdownCodeBlock  = element.MarkdownCodeBlock

	XMLPreserved = element.XMLPreserved

	HTMLDetails = element.HTMLDetails

	MarkdownPanel = element.MarkdownPanel
)
