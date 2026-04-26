package adf

import (
	"github.com/seflue/adf-converter/adf/internal/defaultclass"
)

// ContentClassifier determines how different ADF node types should be handled.
type ContentClassifier interface {
	IsEditable(nodeType NodeType) bool
	IsPreserved(nodeType NodeType) bool
	IsInlineFormattable(nodeType MarkType) bool
}

// NewDefaultClassifier creates a classifier with standard content type rules.
func NewDefaultClassifier() ContentClassifier {
	return defaultclass.New()
}
