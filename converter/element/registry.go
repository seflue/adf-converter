package element

import (
	"fmt"
	"sync"
)

// Registry is the lookup abstraction element converters use to dispatch into
// sibling converters during nested rendering or block-boundary checks.
//
// ConverterRegistry below is the concrete implementation.
type Registry interface {
	// Lookup returns the converter registered for the given node type, if any.
	Lookup(t ADFNodeType) (Converter, bool)

	// BlockParsers returns the ordered block parser list for MD→ADF dispatch.
	BlockParsers() []BlockParserEntry
}

// ConverterRegistry manages registration and lookup of element converters.
//
// Two dispatch paths:
// - ADF→MD: converters map, keyed by node type (GetConverter)
// - MD→ADF: blockParsers slice, ordered by registration (BlockParsers)
//
// Thread-safe for concurrent access via RWMutex.
type ConverterRegistry struct {
	converters   map[ADFNodeType]Converter
	blockParsers []BlockParserEntry
	mu           sync.RWMutex
}

// NewConverterRegistry creates a new empty converter registry.
func NewConverterRegistry() *ConverterRegistry {
	return &ConverterRegistry{
		converters: make(map[ADFNodeType]Converter),
	}
}

// Register registers a converter for the given node type.
//
// If a converter is already registered for this node type, it will be replaced.
// Returns an error if converter is nil.
func (r *ConverterRegistry) Register(nodeType ADFNodeType, converter Converter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if converter == nil {
		return fmt.Errorf("cannot register nil converter for node type %s", nodeType)
	}

	r.converters[nodeType] = converter
	return nil
}

// MustRegister calls Register and panics if it returns an error.
func (r *ConverterRegistry) MustRegister(nodeType ADFNodeType, converter Converter) {
	if err := r.Register(nodeType, converter); err != nil {
		panic(err)
	}
}

// GetConverter retrieves the converter for the given node type, or nil if absent.
func (r *ConverterRegistry) GetConverter(nodeType ADFNodeType) Converter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.converters[nodeType]
}

// Lookup implements Registry by returning the converter for nodeType
// together with a presence flag.
func (r *ConverterRegistry) Lookup(nodeType ADFNodeType) (Converter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.converters[nodeType]
	return c, ok
}

// IsRegistered checks if a converter is registered for the given node type.
func (r *ConverterRegistry) IsRegistered(nodeType ADFNodeType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.converters[nodeType]
	return exists
}

// GetRegisteredTypes returns all node types with registered converters.
func (r *ConverterRegistry) GetRegisteredTypes() []ADFNodeType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]ADFNodeType, 0, len(r.converters))
	for nodeType := range r.converters {
		types = append(types, nodeType)
	}
	return types
}

// Count returns the number of registered converters.
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

	r.converters = make(map[ADFNodeType]Converter)
	r.blockParsers = r.blockParsers[:0]
}

// RegisterBlockParser adds the converter for nodeType to the ordered block parser list.
// The converter must already be registered via Register() and implement BlockParser.
// Registration order determines dispatch priority (first match wins).
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
