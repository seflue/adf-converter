package converter

import "github.com/seflue/adf-converter/converter/internal/defaultclass"

// ContentClassifier determines how different ADF node types should be handled
type ContentClassifier interface {
	IsEditable(nodeType string) bool
	IsPreserved(nodeType string) bool
	IsInlineFormattable(nodeType string) bool
}

// NewDefaultClassifier creates a classifier with standard content type rules
func NewDefaultClassifier() ContentClassifier {
	return defaultclass.New()
}
