package element

import "github.com/seflue/adf-converter/adf_types"

// Converter is the bidirectional conversion interface for a single ADF element type.
type Converter interface {
	ToMarkdown(node adf_types.ADFNode, context ConversionContext) (EnhancedConversionResult, error)
	FromMarkdown(lines []string, startIndex int, context ConversionContext) (adf_types.ADFNode, int, error)
	CanHandle(nodeType ADFNodeType) bool
	GetStrategy() ConversionStrategy
	ValidateInput(input any) error
}

// BlockParser extends Converter with line-based dispatch for MD→ADF parsing.
// Block-level converters implement this to declare which markdown lines they handle.
type BlockParser interface {
	Converter
	CanParseLine(line string) bool
}
