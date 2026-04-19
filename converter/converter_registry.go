package converter

import (
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/placeholder"
)

// BlockParserEntry is an alias for element.BlockParserEntry.
type BlockParserEntry = element.BlockParserEntry

// ConverterRegistry is an alias for element.ConverterRegistry. The concrete
// type lives in the element package so element converters can build their own
// registries without creating a cycle with the parent package.
type ConverterRegistry = element.ConverterRegistry

// NewConverterRegistry creates a new empty converter registry.
func NewConverterRegistry() *ConverterRegistry {
	return element.NewConverterRegistry()
}

// adaptContext adapts markdownConversionContext to ConversionContext for registry converters.
//
// This adapter bridges the legacy context structure used in the switch-based dispatch
// with the new ConversionContext structure expected by ElementConverter implementations.
func adaptContext(ctx *markdownConversionContext, classifier ContentClassifier, manager placeholder.Manager, registry *ConverterRegistry, nodeType ADFNodeType) ConversionContext {
	var strategy ConversionStrategy
	nodeTypeStr := string(nodeType)

	switch {
	case classifier.IsPreserved(nodeTypeStr):
		strategy = Placeholder
	case classifier.IsEditable(nodeTypeStr):
		strategy = StandardMarkdown
	default:
		strategy = Placeholder
	}

	return ConversionContext{
		Strategy:           strategy,
		PreserveAttrs:      true,
		NestedLevel:        ctx.ListDepth,
		ParentNodeType:     "",
		RoundTripMode:      true,
		ErrorRecovery:      true,
		Classifier:         classifier,
		PlaceholderManager: manager,
		Registry:           registry,
	}
}
