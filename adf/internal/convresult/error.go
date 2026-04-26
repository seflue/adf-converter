package convresult

// EnhancedConverterError represents an error from the enhanced converter.
type EnhancedConverterError struct {
	Message string
	Cause   error
}

// NewEnhancedConverterError creates a new enhanced converter error.
func NewEnhancedConverterError(message string) *EnhancedConverterError {
	return &EnhancedConverterError{Message: message}
}

// Error implements the error interface.
func (e *EnhancedConverterError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}
