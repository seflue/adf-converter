package elements

import "adf-converter/converter"

// Type aliases for converter package types
type (
	ConversionContext        = converter.ConversionContext
	EnhancedConversionResult = converter.EnhancedConversionResult
	ConversionStrategy       = converter.ConversionStrategy
	ADFNodeType              = converter.ADFNodeType
	XMLMarshaler             = converter.XMLMarshaler
)

// Constants from converter package
const (
	MarkdownTable      = converter.MarkdownTable
	MarkdownTaskList   = converter.MarkdownTaskList
	MarkdownBlockquote = converter.MarkdownBlockquote
	XMLPreserved       = converter.XMLPreserved

	NodeTable      = converter.NodeTable
	NodeTaskList   = converter.NodeTaskList
	NodeBlockquote = converter.NodeBlockquote
	NodeParagraph  = converter.NodeParagraph
	NodeHeading    = converter.NodeHeading
)

// Functions from converter package
var (
	NewEnhancedConversionResultBuilder = converter.NewEnhancedConversionResultBuilder
	CreateErrorResult                  = converter.CreateErrorResult
	CreateSuccessResult                = converter.CreateSuccessResult
	NewDefaultXMLMarshaler             = converter.NewDefaultXMLMarshaler
)
