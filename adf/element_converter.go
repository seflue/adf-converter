package adf

// Renderer is the bidirectional conversion interface for a single ADF element type.
type Renderer interface {
	ToMarkdown(node Node, context ConversionContext) (RenderResult, error)
	FromMarkdown(lines []string, startIndex int, context ConversionContext) (Node, int, error)
	CanHandle(nodeType NodeType) bool
	GetStrategy() ConversionStrategy
	ValidateInput(input any) error
}

// BlockParser extends Renderer with line-based dispatch for MD→ADF parsing.
// Block-level converters implement this to declare which markdown lines they handle.
type BlockParser interface {
	Renderer
	CanParseLine(line string) bool
}
