package elements

import "github.com/seflue/adf-converter/converter"

// Type aliases for converter package types
type (
	EnhancedConversionResult = converter.EnhancedConversionResult
	ADFNodeType              = converter.ADFNodeType
)

// Constants from converter package
const (
	MarkdownTable      = converter.MarkdownTable
	MarkdownTaskList   = converter.MarkdownTaskList
	MarkdownBlockquote = converter.MarkdownBlockquote
	MarkdownPanel      = converter.MarkdownPanel
	XMLPreserved       = converter.XMLPreserved

	NodeTable      = converter.NodeTable
	NodeTaskList   = converter.NodeTaskList
	NodeBlockquote = converter.NodeBlockquote
	NodeParagraph  = converter.NodeParagraph
	NodeHeading    = converter.NodeHeading
	NodePanel      = converter.ADFNodeType("panel")
)

// Functions from converter package
var (
	NewEnhancedConversionResultBuilder = converter.NewEnhancedConversionResultBuilder
	CreateErrorResult                  = converter.CreateErrorResult
	CreateSuccessResult                = converter.CreateSuccessResult
)
