// Package element defines the core interfaces and value types used by individual
// ADF element converters.
//
// This is a leaf package: it does not import the parent converter/ package, so
// element converters in converter/elements/ can depend on it without creating a
// circular import.
package element

// ADFNodeType represents the type of ADF node for classification.
type ADFNodeType string

const (
	NodeDoc       ADFNodeType = "doc"
	NodeParagraph ADFNodeType = "paragraph"
	NodeHeading   ADFNodeType = "heading"

	NodeTable       ADFNodeType = "table"
	NodeTableRow    ADFNodeType = "tableRow"
	NodeTableCell   ADFNodeType = "tableCell"
	NodeTableHeader ADFNodeType = "tableHeader"
	NodeTaskList    ADFNodeType = "taskList"
	NodeTaskItem    ADFNodeType = "taskItem"
	NodeBlockquote  ADFNodeType = "blockquote"
	NodeCodeBlock   ADFNodeType = "codeBlock"
	NodeBulletList  ADFNodeType = "bulletList"
	NodeOrderedList ADFNodeType = "orderedList"
	NodeListItem    ADFNodeType = "listItem"

	NodeExpand    ADFNodeType = "expand"
	NodeMention   ADFNodeType = "mention"
	NodeHardBreak ADFNodeType = "hardBreak"

	NodeText ADFNodeType = "text"
	MarkLink ADFNodeType = "link"
)

// EnhancedConversionResult contains the result of a single element conversion.
type EnhancedConversionResult struct {
	Content           string
	PreservedAttrs    map[string]any
	Strategy          ConversionStrategy
	Warnings          []string
	ElementsConverted int
	ElementsPreserved int
}

// ContentClassifier determines how different ADF node types should be handled.
type ContentClassifier interface {
	IsEditable(nodeType string) bool
	IsPreserved(nodeType string) bool
	IsInlineFormattable(nodeType string) bool
}

// BlockParserEntry pairs a node type with its BlockParser for ordered dispatch.
type BlockParserEntry struct {
	NodeType ADFNodeType
	Parser   BlockParser
}
