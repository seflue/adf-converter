package converter

// EnhancedConversionResult contains the result of an enhanced conversion operation
type EnhancedConversionResult struct {
	Content           string
	PreservedAttrs    map[string]any
	Strategy          ConversionStrategy
	Warnings          []string
	ElementsConverted int
	ElementsPreserved int
}
