package converter

import (
	"github.com/seflue/adf-converter/placeholder"
)

// ConversionContext provides context information during conversion
type ConversionContext struct {
	Strategy       ConversionStrategy
	PreserveAttrs  bool
	NestedLevel    int
	ParentNodeType ADFNodeType
	RoundTripMode  bool
	ErrorRecovery  bool
	ListDepth      int // Current nesting depth for lists (1 = top level)

	// Classifier and Manager for handling preserved nodes in element converters
	Classifier         ContentClassifier
	PlaceholderManager placeholder.Manager
	PlaceholderSession *placeholder.EditSession
}
