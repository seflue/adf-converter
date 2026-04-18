package converter

// ADFNodeType represents the type of ADF node for classification
type ADFNodeType string

const (
	// Document structure
	NodeDoc       ADFNodeType = "doc"
	NodeParagraph ADFNodeType = "paragraph"
	NodeHeading   ADFNodeType = "heading"

	// Markdown-native elements
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

	// XML-preserved elements
	NodeExpand    ADFNodeType = "expand"
	NodeMention   ADFNodeType = "mention"
	NodeHardBreak ADFNodeType = "hardBreak"

	// Text formatting (existing)
	NodeText ADFNodeType = "text"
	MarkLink ADFNodeType = "link"
)
