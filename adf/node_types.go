package adf

// NodeType represents the type of ADF node for classification.
type NodeType string

const (
	NodeDoc       NodeType = "doc"
	NodeParagraph NodeType = "paragraph"
	NodeHeading   NodeType = "heading"

	NodeTable       NodeType = "table"
	NodeTableRow    NodeType = "tableRow"
	NodeTableCell   NodeType = "tableCell"
	NodeTableHeader NodeType = "tableHeader"
	NodeTaskList    NodeType = "taskList"
	NodeTaskItem    NodeType = "taskItem"
	NodeBlockquote  NodeType = "blockquote"
	NodeCodeBlock   NodeType = "codeBlock"
	NodeBulletList  NodeType = "bulletList"
	NodeOrderedList NodeType = "orderedList"
	NodeListItem    NodeType = "listItem"

	NodeExpand    NodeType = "expand"
	NodeMention   NodeType = "mention"
	NodeHardBreak NodeType = "hardBreak"

	NodeText NodeType = "text"
	MarkLink NodeType = "link"
)
