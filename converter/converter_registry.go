package converter

import (
	"fmt"
	"sync"

	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/placeholder"
)

// BlockParserEntry is an alias for element.BlockParserEntry.
type BlockParserEntry = element.BlockParserEntry

// ConverterRegistry manages registration and lookup of element converters.
//
// Two dispatch paths:
// - ADF→MD: converters map, keyed by node type (GetConverter)
// - MD→ADF: blockParsers slice, ordered by registration (BlockParsers)
//
// Thread-safe for concurrent access via RWMutex.
type ConverterRegistry struct {
	converters   map[ADFNodeType]ElementConverter
	blockParsers []BlockParserEntry
	mu           sync.RWMutex
}

// NewConverterRegistry creates a new empty converter registry
//
// The registry starts empty to ensure zero behavior change during infrastructure setup.
// Converters are registered incrementally as they are extracted from legacy code.
func NewConverterRegistry() *ConverterRegistry {
	return &ConverterRegistry{
		converters: make(map[ADFNodeType]ElementConverter),
	}
}

// Register registers a converter for the given node type.
//
// If a converter is already registered for this node type, it will be replaced.
// This allows for testing different converter implementations.
//
// Returns an error if converter is nil. Use MustRegister for init-time
// registration where a nil converter is a programmer error.
func (r *ConverterRegistry) Register(nodeType ADFNodeType, converter ElementConverter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if converter == nil {
		return fmt.Errorf("cannot register nil converter for node type %s", nodeType)
	}

	r.converters[nodeType] = converter
	return nil
}

// MustRegister calls Register and panics if it returns an error.
// Intended for init-time registration where a failure is a programmer error.
func (r *ConverterRegistry) MustRegister(nodeType ADFNodeType, converter ElementConverter) {
	if err := r.Register(nodeType, converter); err != nil {
		panic(err)
	}
}

// GetConverter retrieves the converter for the given node type
//
// Returns nil if no converter is registered for this node type.
//
// Parameters:
//   - nodeType: The ADF node type to look up
//
// Returns:
//   - ElementConverter if registered, nil otherwise
func (r *ConverterRegistry) GetConverter(nodeType ADFNodeType) ElementConverter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.converters[nodeType]
}

// Lookup implements element.Registry by returning the converter for nodeType
// together with a presence flag.
func (r *ConverterRegistry) Lookup(nodeType ADFNodeType) (ElementConverter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.converters[nodeType]
	return c, ok
}

// IsRegistered checks if a converter is registered for the given node type
func (r *ConverterRegistry) IsRegistered(nodeType ADFNodeType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.converters[nodeType]
	return exists
}

// GetRegisteredTypes returns all node types with registered converters
func (r *ConverterRegistry) GetRegisteredTypes() []ADFNodeType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]ADFNodeType, 0, len(r.converters))
	for nodeType := range r.converters {
		types = append(types, nodeType)
	}
	return types
}

// Count returns the number of registered converters
func (r *ConverterRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.converters)
}

// Clear removes all registered converters and block parsers.
//
// Primarily used for testing to reset registry state between tests.
func (r *ConverterRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.converters = make(map[ADFNodeType]ElementConverter)
	r.blockParsers = r.blockParsers[:0]
}

// RegisterBlockParser adds the converter for nodeType to the ordered block parser list.
// The converter must already be registered via Register() and implement BlockParser.
// Registration order determines dispatch priority (first match wins).
//
// Returns an error if the converter is not registered or does not implement
// BlockParser. Use MustRegisterBlockParser for init-time registration.
func (r *ConverterRegistry) RegisterBlockParser(nodeType ADFNodeType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, exists := r.converters[nodeType]
	if !exists {
		return fmt.Errorf("RegisterBlockParser: converter %q not registered", nodeType)
	}

	bp, ok := conv.(BlockParser)
	if !ok {
		return fmt.Errorf("RegisterBlockParser: converter %q does not implement BlockParser", nodeType)
	}

	r.blockParsers = append(r.blockParsers, BlockParserEntry{NodeType: nodeType, Parser: bp})
	return nil
}

// MustRegisterBlockParser calls RegisterBlockParser and panics if it returns an error.
// Intended for init-time registration where a failure is a programmer error.
func (r *ConverterRegistry) MustRegisterBlockParser(nodeType ADFNodeType) {
	if err := r.RegisterBlockParser(nodeType); err != nil {
		panic(err)
	}
}

// BlockParsers returns the ordered block parser list for MD→ADF dispatch.
func (r *ConverterRegistry) BlockParsers() []BlockParserEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.blockParsers
}

// RegisterGlobal registers a converter with the global registry.
//
// This function is intended to be called from init() functions in the elements/
// package to register converters without creating circular import dependencies.
//
// Returns an error if converter is nil.
func RegisterGlobal(nodeType ADFNodeType, converter ElementConverter) error {
	return globalRegistry.Register(nodeType, converter)
}

// adaptContext adapts markdownConversionContext to ConversionContext for registry converters
//
// This adapter bridges the legacy context structure used in the switch-based dispatch
// with the new ConversionContext structure expected by ElementConverter implementations.
//
// During incremental migration, both context types coexist:
// - markdownConversionContext: Used by legacy switch-based conversion functions
// - ConversionContext: Used by new ElementConverter implementations
//
// Parameters:
//   - ctx: Legacy markdown conversion context with list depth tracking
//   - classifier: Content classifier for determining conversion strategy
//   - nodeType: The type of node being converted (for strategy determination)
//
// Returns:
//   - ConversionContext with appropriate strategy and settings
func adaptContext(ctx *markdownConversionContext, classifier ContentClassifier, manager placeholder.Manager, nodeType ADFNodeType) ConversionContext {
	// Determine strategy based on classifier
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
		PreserveAttrs:      true, // Default to preserving attributes for round-trip fidelity
		NestedLevel:        ctx.ListDepth,
		ParentNodeType:     "", // Will be set by container converters if needed
		RoundTripMode:      true,
		ErrorRecovery:      true,
		Classifier:         classifier,
		PlaceholderManager: manager,
		Registry:           globalRegistry,
	}
}
