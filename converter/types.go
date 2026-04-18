package converter

import "github.com/seflue/adf-converter/converter/internal"

// Re-export types from internal package for backward compatibility
type (
	XMLMarshaler        = internal.XMLMarshaler
	DefaultXMLMarshaler = internal.DefaultXMLMarshaler
	XMLValidator        = internal.XMLValidator
	ExpandElement       = internal.ExpandElement
	MentionElement      = internal.MentionElement
	TaskElement         = internal.TaskElement
	HardBreakElement    = internal.HardBreakElement
)

// Re-export functions from internal package
var (
	NewDefaultXMLMarshaler = internal.NewDefaultXMLMarshaler
	NewXMLValidator        = internal.NewXMLValidator
)

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

// EnhancedConverterError represents an error from the enhanced converter
type EnhancedConverterError struct {
	Message string
	Cause   error
}

// NewEnhancedConverterError creates a new enhanced converter error
func NewEnhancedConverterError(message string) *EnhancedConverterError {
	return &EnhancedConverterError{
		Message: message,
		Cause:   nil,
	}
}

// Error implements the error interface
func (e *EnhancedConverterError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}
