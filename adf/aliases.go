package adf

// EnhancedConversionResult contains the result of a single element conversion.
type EnhancedConversionResult struct {
	Content           string
	PreservedAttrs    map[string]any
	Strategy          ConversionStrategy
	Warnings          []string
	ElementsConverted int
	ElementsPreserved int
}

// BlockParserEntry pairs a node type with its BlockParser for ordered dispatch.
type BlockParserEntry struct {
	NodeType NodeType
	Parser   BlockParser
}
