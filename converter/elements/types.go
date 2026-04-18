package elements

import (
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/internal/convresult"
)

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

// Helpers from internal convresult package
var (
	NewEnhancedConversionResultBuilder = convresult.NewEnhancedConversionResultBuilder
	CreateErrorResult                  = convresult.CreateErrorResult
	CreateSuccessResult                = convresult.CreateSuccessResult
)
