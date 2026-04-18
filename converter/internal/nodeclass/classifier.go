// Package nodeclass provides ADF node type → conversion strategy classification.
package nodeclass

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

// NodeClassification maps ADF node types to conversion strategies
type NodeClassification struct {
	NodeType converter.ADFNodeType
	Strategy converter.ConversionStrategy
	Priority int // For conflict resolution
}

// NodeClassifier determines conversion strategies for ADF nodes
type NodeClassifier struct {
	classifications map[converter.ADFNodeType]NodeClassification
}

// NewNodeClassifier creates a new node classifier with default mappings
func NewNodeClassifier() *NodeClassifier {
	nc := &NodeClassifier{
		classifications: make(map[converter.ADFNodeType]NodeClassification),
	}
	nc.registerDefaultClassifications()
	return nc
}

// registerDefaultClassifications sets up the default node type to strategy mappings
func (nc *NodeClassifier) registerDefaultClassifications() {
	// Markdown-native elements
	nc.RegisterNodeType(converter.NodeTable, converter.MarkdownTable, 100)
	nc.RegisterNodeType(converter.NodeTableRow, converter.MarkdownTable, 100)
	nc.RegisterNodeType(converter.NodeTableCell, converter.MarkdownTable, 100)
	nc.RegisterNodeType(converter.NodeTableHeader, converter.MarkdownTable, 100)
	nc.RegisterNodeType(converter.NodeTaskList, converter.MarkdownTaskList, 100)
	nc.RegisterNodeType(converter.NodeTaskItem, converter.MarkdownTaskList, 100)
	nc.RegisterNodeType(converter.NodeBlockquote, converter.MarkdownBlockquote, 100)
	nc.RegisterNodeType(converter.NodeCodeBlock, converter.MarkdownCodeBlock, 100)

	// HTML details elements
	nc.RegisterNodeType(converter.NodeExpand, converter.HTMLDetails, 100)
	nc.RegisterNodeType(converter.NodeMention, converter.StandardMarkdown, 100)
	nc.RegisterNodeType(converter.NodeHardBreak, converter.XMLPreserved, 50)

	// Standard markdown for basic elements
	nc.RegisterNodeType(converter.NodeParagraph, converter.StandardMarkdown, 100)
	nc.RegisterNodeType(converter.NodeHeading, converter.StandardMarkdown, 100)
	nc.RegisterNodeType(converter.NodeBulletList, converter.StandardMarkdown, 80)
	nc.RegisterNodeType(converter.NodeOrderedList, converter.StandardMarkdown, 100)
	nc.RegisterNodeType(converter.NodeListItem, converter.StandardMarkdown, 100)
	nc.RegisterNodeType(converter.NodeText, converter.StandardMarkdown, 100)

	// Placeholder fallback for unhandled types
	nc.RegisterNodeType("unknown", converter.Placeholder, 1)
}

// RegisterNodeType registers a node type with its conversion strategy and priority
func (nc *NodeClassifier) RegisterNodeType(nodeType converter.ADFNodeType, strategy converter.ConversionStrategy, priority int) {
	nc.classifications[nodeType] = NodeClassification{
		NodeType: nodeType,
		Strategy: strategy,
		Priority: priority,
	}
}

// ClassifyNode determines the conversion strategy for an ADF node
func (nc *NodeClassifier) ClassifyNode(node adf_types.ADFNode) converter.ConversionStrategy {
	nodeType := converter.ADFNodeType(node.Type)

	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Strategy
	}

	if nodeType == converter.NodeBulletList {
		if nc.isTaskList(node) {
			return converter.MarkdownTaskList
		}
		return converter.StandardMarkdown
	}

	return converter.Placeholder
}

// isTaskList determines if a bullet list is actually a task list
func (nc *NodeClassifier) isTaskList(node adf_types.ADFNode) bool {
	if node.Type != string(converter.NodeBulletList) {
		return false
	}

	for _, item := range node.Content {
		if item.Type == string(converter.NodeListItem) {
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
func (nc *NodeClassifier) GetSupportedTypes() []converter.ADFNodeType {
	var types []converter.ADFNodeType
	for nodeType := range nc.classifications {
		types = append(types, nodeType)
	}
	return types
}

// GetStrategyForType returns the strategy for a specific node type
func (nc *NodeClassifier) GetStrategyForType(nodeType converter.ADFNodeType) converter.ConversionStrategy {
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Strategy
	}
	return converter.Placeholder
}

// GetPriorityForType returns the priority for a specific node type
func (nc *NodeClassifier) GetPriorityForType(nodeType converter.ADFNodeType) int {
	if classification, exists := nc.classifications[nodeType]; exists {
		return classification.Priority
	}
	return 0
}
