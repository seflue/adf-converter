package converter

import (
	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// StrategyClassifier determines the appropriate conversion strategy for ADF nodes
type StrategyClassifier interface {
	// ClassifyNode determines the conversion strategy for the given ADF node
	ClassifyNode(node adf_types.ADFNode) ConversionStrategy

	// RegisterNodeType registers a new node type with its conversion strategy
	RegisterNodeType(nodeType ADFNodeType, strategy ConversionStrategy, priority int)

	// GetSupportedTypes returns all supported ADF node types
	GetSupportedTypes() []ADFNodeType

	// GetStrategyForType returns the strategy for a specific node type
	GetStrategyForType(nodeType ADFNodeType) ConversionStrategy
}

// DefaultStrategyClassifier implements the StrategyClassifier interface
type DefaultStrategyClassifier struct {
	nodeClassifier *NodeClassifier
	linkClassifier *DefaultLinkClassifier // For link-specific classification
}

// NewDefaultStrategyClassifier creates a new default strategy classifier
func NewDefaultStrategyClassifier() *DefaultStrategyClassifier {
	return &DefaultStrategyClassifier{
		nodeClassifier: NewNodeClassifier(),
		linkClassifier: &DefaultLinkClassifier{},
	}
}

// ClassifyNode determines the conversion strategy for the given ADF node
func (dsc *DefaultStrategyClassifier) ClassifyNode(node adf_types.ADFNode) ConversionStrategy {
	// Handle link marks specially
	if node.Type == string(NodeText) && len(node.Marks) > 0 {
		for _, mark := range node.Marks {
			if mark.Type == string(MarkLink) {
				// Use link-specific classification
				classification := dsc.linkClassifier.ClassifyLink(mark)
				return GetStrategyForLinkType(classification.Type)
			}
		}
	}

	// Handle inline card nodes
	if node.Type == adf_types.NodeTypeInlineCard {
		// InlineCard nodes should be converted to markdown links
		return StandardMarkdown
	}

	// Use node classifier for all other types
	return dsc.nodeClassifier.ClassifyNode(node)
}

// RegisterNodeType registers a new node type with its conversion strategy
func (dsc *DefaultStrategyClassifier) RegisterNodeType(nodeType ADFNodeType, strategy ConversionStrategy, priority int) {
	dsc.nodeClassifier.RegisterNodeType(nodeType, strategy, priority)
}

// GetSupportedTypes returns all supported ADF node types
func (dsc *DefaultStrategyClassifier) GetSupportedTypes() []ADFNodeType {
	return dsc.nodeClassifier.GetSupportedTypes()
}

// GetStrategyForType returns the strategy for a specific node type
func (dsc *DefaultStrategyClassifier) GetStrategyForType(nodeType ADFNodeType) ConversionStrategy {
	return dsc.nodeClassifier.GetStrategyForType(nodeType)
}

// ClassifyWithContext provides context-aware classification for complex scenarios
func (dsc *DefaultStrategyClassifier) ClassifyWithContext(node adf_types.ADFNode, context ConversionContext) ConversionStrategy {
	// Base classification
	strategy := dsc.ClassifyNode(node)

	// Context-based adjustments
	switch context.Strategy {
	case XMLPreserved:
		// If parent is XML-preserved, children should also be XML-preserved for consistency
		if strategy == StandardMarkdown {
			return XMLPreserved
		}
	case MarkdownTable:
		// If we're inside a table, preserve table-related strategies
		if node.Type == string(NodeTableCell) || node.Type == string(NodeTableHeader) {
			return MarkdownTable
		}
	case MarkdownTaskList:
		// If we're inside a task list, preserve task-related strategies
		if node.Type == string(NodeTaskItem) {
			return MarkdownTaskList
		}
	}

	// Apply nesting level constraints
	if context.NestedLevel > 3 {
		// For deeply nested content, prefer placeholder to avoid complexity
		if strategy == XMLPreserved {
			return Placeholder
		}
	}

	return strategy
}

// GetStrategyMetadata returns metadata about a conversion strategy
func (dsc *DefaultStrategyClassifier) GetStrategyMetadata(strategy ConversionStrategy) StrategyMetadata {
	switch strategy {
	case StandardMarkdown:
		return StrategyMetadata{
			ReadabilityScore:    100,
			PreservesAttributes: false,
			RequiresPostProcess: false,
			SupportsNesting:     true,
		}
	case MarkdownTable:
		return StrategyMetadata{
			ReadabilityScore:    90,
			PreservesAttributes: false,
			RequiresPostProcess: false,
			SupportsNesting:     false, // Tables have limited nesting support
		}
	case MarkdownTaskList:
		return StrategyMetadata{
			ReadabilityScore:    95,
			PreservesAttributes: true, // Preserves state in checkbox syntax
			RequiresPostProcess: false,
			SupportsNesting:     true,
		}
	case MarkdownBlockquote:
		return StrategyMetadata{
			ReadabilityScore:    85,
			PreservesAttributes: false,
			RequiresPostProcess: false,
			SupportsNesting:     true,
		}
	case XMLPreserved:
		return StrategyMetadata{
			ReadabilityScore:    60, // Less readable due to XML tags
			PreservesAttributes: true,
			RequiresPostProcess: true, // Requires XML parsing on round-trip
			SupportsNesting:     true,
		}
	case HTMLWrapped:
		return StrategyMetadata{
			ReadabilityScore:    70,
			PreservesAttributes: true,
			RequiresPostProcess: true,
			SupportsNesting:     true,
		}
	case Placeholder:
		return StrategyMetadata{
			ReadabilityScore:    0, // Not readable
			PreservesAttributes: true,
			RequiresPostProcess: true,
			SupportsNesting:     true,
		}
	default:
		return StrategyMetadata{
			ReadabilityScore:    0,
			PreservesAttributes: false,
			RequiresPostProcess: false,
			SupportsNesting:     false,
		}
	}
}

// StrategyMetadata provides information about conversion strategy characteristics
type StrategyMetadata struct {
	ReadabilityScore    int  // 0-100, higher is more readable
	PreservesAttributes bool // Whether attributes are preserved
	RequiresPostProcess bool // Whether post-processing is needed for round-trip
	SupportsNesting     bool // Whether nested content is supported
}

// ConversionContext provides context information during conversion
type ConversionContext struct {
	Strategy       ConversionStrategy
	PreserveAttrs  bool
	NestedLevel    int
	ParentNodeType ADFNodeType
	RoundTripMode  bool
	ErrorRecovery  bool
	ListDepth      int // Current nesting depth for lists (1 = top level)

	// Classifier and Manager for handling preserved nodes in element converters
	Classifier           ContentClassifier
	PlaceholderManager   placeholder.Manager
	PlaceholderSession   *placeholder.EditSession
}
