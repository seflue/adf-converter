package converter

import "adf-converter/adf_types"

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

// NodeClassification maps ADF node types to conversion strategies
type NodeClassification struct {
	NodeType ADFNodeType
	Strategy ConversionStrategy
	Priority int // For conflict resolution
}

// NodeClassifier determines conversion strategies for ADF nodes
type NodeClassifier struct {
	classifications map[ADFNodeType]NodeClassification
}

// NewNodeClassifier creates a new node classifier with default mappings
func NewNodeClassifier() *NodeClassifier {
	nc := &NodeClassifier{
		classifications: make(map[ADFNodeType]NodeClassification),
	}

	// Register default node type classifications
	nc.registerDefaultClassifications()

	return nc
}

// registerDefaultClassifications sets up the default node type to strategy mappings
func (nc *NodeClassifier) registerDefaultClassifications() {
	// Markdown-native elements
	nc.RegisterNodeType(NodeTable, MarkdownTable, 100)
	nc.RegisterNodeType(NodeTableRow, MarkdownTable, 100)
	nc.RegisterNodeType(NodeTableCell, MarkdownTable, 100)
	nc.RegisterNodeType(NodeTableHeader, MarkdownTable, 100)
	nc.RegisterNodeType(NodeTaskList, MarkdownTaskList, 100)
	nc.RegisterNodeType(NodeTaskItem, MarkdownTaskList, 100)
	nc.RegisterNodeType(NodeBlockquote, MarkdownBlockquote, 100)
	nc.RegisterNodeType(NodeCodeBlock, MarkdownCodeBlock, 100)

	// HTML details elements
	nc.RegisterNodeType(NodeExpand, HTMLDetails, 100)
	nc.RegisterNodeType(NodeMention, StandardMarkdown, 100)
	nc.RegisterNodeType(NodeHardBreak, XMLPreserved, 50) // Lower priority for optional handling

	// Standard markdown for basic elements
	nc.RegisterNodeType(NodeParagraph, StandardMarkdown, 100)
	nc.RegisterNodeType(NodeHeading, StandardMarkdown, 100)
	nc.RegisterNodeType(NodeBulletList, StandardMarkdown, 80) // Lower than task list for precedence
	nc.RegisterNodeType(NodeOrderedList, StandardMarkdown, 100)
	nc.RegisterNodeType(NodeListItem, StandardMarkdown, 100)
	nc.RegisterNodeType(NodeText, StandardMarkdown, 100)

	// Placeholder fallback for unhandled types
	nc.RegisterNodeType("unknown", Placeholder, 1)
}

// RegisterNodeType registers a node type with its conversion strategy and priority
func (nc *NodeClassifier) RegisterNodeType(nodeType ADFNodeType, strategy ConversionStrategy, priority int) {
	nc.classifications[nodeType] = NodeClassification{
		NodeType: nodeType,
		Strategy: strategy,
		Priority: priority,
	}
}

// ClassifyNode determines the conversion strategy for an ADF node
func (nc *NodeClassifier) ClassifyNode(node adf_types.ADFNode) ConversionStrategy {
	nodeType := ADFNodeType(node.Type)

	// Check for direct classification
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Strategy
	}

	// Special cases for task lists vs bullet lists
	if nodeType == NodeBulletList {
		// Check if this is actually a task list
		if nc.isTaskList(node) {
			return MarkdownTaskList
		}
		return StandardMarkdown
	}

	// Default to placeholder for unknown types
	return Placeholder
}

// isTaskList determines if a bullet list is actually a task list
func (nc *NodeClassifier) isTaskList(node adf_types.ADFNode) bool {
	if node.Type != string(NodeBulletList) {
		return false
	}

	// Check if any list item has task-specific attributes
	for _, item := range node.Content {
		if item.Type == string(NodeListItem) {
			// Check for task-specific attributes
			if state, exists := item.Attrs["state"]; exists {
				if stateStr, ok := state.(string); ok {
					if stateStr == "TODO" || stateStr == "DONE" {
						return true
					}
				}
			}

			// Check for localId which is common in task items
			if _, exists := item.Attrs["localId"]; exists {
				return true
			}
		}
	}

	return false
}

// GetSupportedTypes returns all supported ADF node types
func (nc *NodeClassifier) GetSupportedTypes() []ADFNodeType {
	var types []ADFNodeType
	for nodeType := range nc.classifications {
		types = append(types, nodeType)
	}
	return types
}

// GetStrategyForType returns the strategy for a specific node type
func (nc *NodeClassifier) GetStrategyForType(nodeType ADFNodeType) ConversionStrategy {
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Strategy
	}
	return Placeholder
}

// GetPriorityForType returns the priority for a specific node type
func (nc *NodeClassifier) GetPriorityForType(nodeType ADFNodeType) int {
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Priority
	}
	return 0
}
