package converter

import "github.com/seflue/adf-converter/converter/element"

// ADFNodeType is an alias for element.ADFNodeType.
type ADFNodeType = element.ADFNodeType

const (
	NodeDoc       = element.NodeDoc
	NodeParagraph = element.NodeParagraph
	NodeHeading   = element.NodeHeading

	NodeTable       = element.NodeTable
	NodeTableRow    = element.NodeTableRow
	NodeTableCell   = element.NodeTableCell
	NodeTableHeader = element.NodeTableHeader
	NodeTaskList    = element.NodeTaskList
	NodeTaskItem    = element.NodeTaskItem
	NodeBlockquote  = element.NodeBlockquote
	NodeCodeBlock   = element.NodeCodeBlock
	NodeBulletList  = element.NodeBulletList
	NodeOrderedList = element.NodeOrderedList
	NodeListItem    = element.NodeListItem

	NodeExpand    = element.NodeExpand
	NodeMention   = element.NodeMention
	NodeHardBreak = element.NodeHardBreak

	NodeText = element.NodeText
	MarkLink = element.MarkLink
)
