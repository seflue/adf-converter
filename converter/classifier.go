package converter

import (
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/converter/internal/defaultclass"
)

// ContentClassifier is an alias for element.ContentClassifier.
type ContentClassifier = element.ContentClassifier

// NewDefaultClassifier creates a classifier with standard content type rules.
func NewDefaultClassifier() ContentClassifier {
	return defaultclass.New()
}
