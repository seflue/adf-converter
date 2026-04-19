package element

// Registry is the lookup abstraction element converters use to dispatch into
// sibling converters during nested rendering or block-boundary checks.
//
// The parent converter package provides a concrete implementation and wires it
// into ConversionContext.Registry; element converters must not depend on the
// concrete registry type.
type Registry interface {
	// Lookup returns the converter registered for the given node type, if any.
	Lookup(t ADFNodeType) (Converter, bool)

	// BlockParsers returns the ordered block parser list for MD→ADF dispatch.
	BlockParsers() []BlockParserEntry
}
