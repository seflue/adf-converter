// Package nodeclass provides ADF node type → conversion strategy classification.
package nodeclass

import (
	"github.com/seflue/adf-converter/adf"
)

// NodeClassification maps ADF node types to conversion strategies
type NodeClassification struct {
	NodeType adf.NodeType
	Strategy adf.ConversionStrategy
	Priority int // For conflict resolution
}

// NodeClassifier determines conversion strategies for ADF nodes
type NodeClassifier struct {
	classifications map[adf.NodeType]NodeClassification
}

// NewNodeClassifier creates a new node classifier with default mappings
func NewNodeClassifier() *NodeClassifier {
	nc := &NodeClassifier{
		classifications: make(map[adf.NodeType]NodeClassification),
	}
	nc.registerDefaultClassifications()
	return nc
}

// registerDefaultClassifications sets up the default node type to strategy mappings
func (nc *NodeClassifier) registerDefaultClassifications() {
	// Markdown-native elements
	nc.RegisterNodeType(adf.NodeTable, adf.MarkdownTable, 100)
	nc.RegisterNodeType(adf.NodeTableRow, adf.MarkdownTable, 100)
	nc.RegisterNodeType(adf.NodeTableCell, adf.MarkdownTable, 100)
	nc.RegisterNodeType(adf.NodeTableHeader, adf.MarkdownTable, 100)
	nc.RegisterNodeType(adf.NodeTaskList, adf.MarkdownTaskList, 100)
	nc.RegisterNodeType(adf.NodeTaskItem, adf.MarkdownTaskList, 100)
	nc.RegisterNodeType(adf.NodeBlockquote, adf.MarkdownBlockquote, 100)
	nc.RegisterNodeType(adf.NodeCodeBlock, adf.MarkdownCodeBlock, 100)

	// HTML details elements
	nc.RegisterNodeType(adf.NodeExpand, adf.HTMLDetails, 100)
	nc.RegisterNodeType(adf.NodeMention, adf.StandardMarkdown, 100)
	nc.RegisterNodeType(adf.NodeHardBreak, adf.XMLPreserved, 50)

	// Standard markdown for basic elements
	nc.RegisterNodeType(adf.NodeParagraph, adf.StandardMarkdown, 100)
	nc.RegisterNodeType(adf.NodeHeading, adf.StandardMarkdown, 100)
	nc.RegisterNodeType(adf.NodeBulletList, adf.StandardMarkdown, 80)
	nc.RegisterNodeType(adf.NodeOrderedList, adf.StandardMarkdown, 100)
	nc.RegisterNodeType(adf.NodeListItem, adf.StandardMarkdown, 100)
	nc.RegisterNodeType(adf.NodeText, adf.StandardMarkdown, 100)

	// Placeholder fallback for unhandled types
	nc.RegisterNodeType("unknown", adf.Placeholder, 1)
}

// RegisterNodeType registers a node type with its conversion strategy and priority
func (nc *NodeClassifier) RegisterNodeType(nodeType adf.NodeType, strategy adf.ConversionStrategy, priority int) {
	nc.classifications[nodeType] = NodeClassification{
		NodeType: nodeType,
		Strategy: strategy,
		Priority: priority,
	}
}

// ClassifyNode determines the conversion strategy for an ADF node
func (nc *NodeClassifier) ClassifyNode(node adf.Node) adf.ConversionStrategy {
	nodeType := adf.NodeType(node.Type)

	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Strategy
	}

	if nodeType == adf.NodeBulletList {
		if nc.isTaskList(node) {
			return adf.MarkdownTaskList
		}
		return adf.StandardMarkdown
	}

	return adf.Placeholder
}

// isTaskList determines if a bullet list is actually a task list
func (nc *NodeClassifier) isTaskList(node adf.Node) bool {
	if node.Type != string(adf.NodeBulletList) {
		return false
	}

	for _, item := range node.Content {
		if item.Type == string(adf.NodeListItem) {
			if state, exists := item.Attrs["state"]; exists {
				if stateStr, ok := state.(string); ok {
					if stateStr == "TODO" || stateStr == "DONE" {
						return true
					}
				}
			}

			if _, exists := item.Attrs["localId"]; exists {
				return true
			}
		}
	}

	return false
}

// GetSupportedTypes returns all supported ADF node types
func (nc *NodeClassifier) GetSupportedTypes() []adf.NodeType {
	var types []adf.NodeType
	for nodeType := range nc.classifications {
		types = append(types, nodeType)
	}
	return types
}

// GetStrategyForType returns the strategy for a specific node type
func (nc *NodeClassifier) GetStrategyForType(nodeType adf.NodeType) adf.ConversionStrategy {
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Strategy
	}
	return adf.Placeholder
}

// GetPriorityForType returns the priority for a specific node type
func (nc *NodeClassifier) GetPriorityForType(nodeType adf.NodeType) int {
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Priority
	}
	return 0
}
