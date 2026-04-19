package element

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// ConversionContext carries per-conversion state for element converters.
type ConversionContext struct {
	Strategy       ConversionStrategy
	PreserveAttrs  bool
	NestedLevel    int
	ParentNodeType ADFNodeType
	RoundTripMode  bool
	ErrorRecovery  bool
	ListDepth      int

	Classifier         ContentClassifier
	PlaceholderManager placeholder.Manager
	PlaceholderSession *placeholder.EditSession

	// Registry exposes the converter registry to element converters that need
	// to dispatch into other converters (lists, panels, expand, paragraph block
	// boundary checks, inline renderer). The parent converter package wires
	// this field; element converters must not assume a global registry.
	Registry Registry

	// ParseNested parses markdown lines into ADF nodes for element converters
	// that recursively process inner markdown bodies (expand, ...). The parent
	// converter package wires this with a fully configured MarkdownParser; it
	// may be nil when a converter is used outside the parser pipeline.
	ParseNested func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error)
}
