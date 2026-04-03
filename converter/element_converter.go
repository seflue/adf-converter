package converter

import "adf-converter/adf_types"

// ElementConverter defines the interface for bidirectional conversion of ADF elements
//
// Each converter handles both directions for round-trip fidelity:
// - ToMarkdown: Convert ADF node to Markdown representation
// - FromMarkdown: Parse Markdown lines into ADF node
//
// This interface enables the Open/Closed Principle:
// - Open for extension: Add new converters without modifying core dispatch logic
// - Closed for modification: Core conversion flow remains stable
type ElementConverter interface {
	// ToMarkdown converts an ADF node to Markdown
	//
	// Parameters:
	//   - node: The ADF node to convert
	//   - context: Conversion context with strategy and nesting info
	//
	// Returns:
	//   - EnhancedConversionResult with markdown content and metadata
	//   - error if conversion fails
	ToMarkdown(node adf_types.ADFNode, context ConversionContext) (EnhancedConversionResult, error)

	// FromMarkdown parses Markdown lines into an ADF node
	//
	// This method handles multi-line parsing with line consumption tracking.
	// For inline elements (text, hardBreak), this may return an error indicating
	// that parsing is handled by container converters (paragraph, heading).
	//
	// Parameters:
	//   - lines: All markdown lines being parsed
	//   - startIndex: Index of first line to parse
	//   - context: Conversion context
	//
	// Returns:
	//   - ADFNode: Parsed ADF node
	//   - int: Number of lines consumed (0 for inline elements)
	//   - error: If parsing fails or element type is inline-only
	FromMarkdown(lines []string, startIndex int, context ConversionContext) (adf_types.ADFNode, int, error)

	// CanHandle returns true if this converter can handle the given node type
	CanHandle(nodeType ADFNodeType) bool

	// GetStrategy returns the conversion strategy for this converter
	GetStrategy() ConversionStrategy

	// ValidateInput validates that the input can be processed by this converter
	//
	// For ToMarkdown: validates node structure
	// For FromMarkdown: validates markdown format
	ValidateInput(input interface{}) error
}
