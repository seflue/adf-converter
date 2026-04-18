package converter

// EnhancedConversionResult contains the result of an enhanced conversion operation
type EnhancedConversionResult struct {
	Content           string
	PreservedAttrs    map[string]interface{}
	Strategy          ConversionStrategy
	Warnings          []string
	ElementsConverted int
	ElementsPreserved int
}

// ValidationMetrics provides conversion quality statistics
type ValidationMetrics struct {
	TotalConversions int
	SuccessfulRounds int
	AttributesLost   int
	ContentModified  int
	FidelityScore    float64 // 0.0 to 1.0
}

