package linkclass

import (
	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/nodeclass"
)

// ConversionStrategyMapping defines the complete mapping from link characteristics to strategy.
type ConversionStrategyMapping struct {
	IsInternal   bool
	HasMetadata  bool
	MetadataKeys []string
}

// DetermineStrategy analyzes link characteristics and returns the appropriate strategy.
func (csm *ConversionStrategyMapping) DetermineStrategy() adf.ConversionStrategy {
	if !csm.IsInternal {
		return adf.StandardMarkdown
	}
	if !csm.HasMetadata || len(csm.MetadataKeys) == 0 {
		return adf.StandardMarkdown
	}
	return adf.HTMLWrapped
}

func CreateMappingFromClassification(classification LinkClassification) ConversionStrategyMapping {
	return ConversionStrategyMapping{
		IsInternal:   classification.IsInternal,
		HasMetadata:  classification.HasMetadata,
		MetadataKeys: classification.MetadataKeys,
	}
}

// GetStrategyForLinkType returns the appropriate conversion strategy for a given link type.
func GetStrategyForLinkType(linkType LinkType) adf.ConversionStrategy {
	switch linkType {
	case WebLink:
		return adf.StandardMarkdown
	case SimpleInternalLink:
		return adf.StandardMarkdown
	case ComplexInternalLink:
		return adf.HTMLWrapped
	default:
		return adf.Placeholder
	}
}

// StrategyClassifier determines the appropriate conversion strategy for ADF nodes.
type StrategyClassifier interface {
	ClassifyNode(node adf.Node) adf.ConversionStrategy
	RegisterNodeType(nodeType adf.NodeType, strategy adf.ConversionStrategy, priority int)
	GetSupportedTypes() []adf.NodeType
	GetStrategyForType(nodeType adf.NodeType) adf.ConversionStrategy
}

// DefaultStrategyClassifier implements the StrategyClassifier interface.
type DefaultStrategyClassifier struct {
	nodeClassifier *nodeclass.NodeClassifier
	linkClassifier *DefaultLinkClassifier
}

func NewDefaultStrategyClassifier() *DefaultStrategyClassifier {
	return &DefaultStrategyClassifier{
		nodeClassifier: nodeclass.NewNodeClassifier(),
		linkClassifier: &DefaultLinkClassifier{},
	}
}

func (dsc *DefaultStrategyClassifier) ClassifyNode(node adf.Node) adf.ConversionStrategy {
	if node.Type == adf.NodeTypeText && len(node.Marks) > 0 {
		for _, mark := range node.Marks {
			if mark.Type == adf.MarkTypeLink {
				classification := dsc.linkClassifier.ClassifyLink(mark)
				return GetStrategyForLinkType(classification.Type)
			}
		}
	}
	if node.Type == adf.NodeTypeInlineCard {
		return adf.StandardMarkdown
	}
	return dsc.nodeClassifier.ClassifyNode(node)
}

func (dsc *DefaultStrategyClassifier) RegisterNodeType(nodeType adf.NodeType, strategy adf.ConversionStrategy, priority int) {
	dsc.nodeClassifier.RegisterNodeType(nodeType, strategy, priority)
}

func (dsc *DefaultStrategyClassifier) GetSupportedTypes() []adf.NodeType {
	return dsc.nodeClassifier.GetSupportedTypes()
}

func (dsc *DefaultStrategyClassifier) GetStrategyForType(nodeType adf.NodeType) adf.ConversionStrategy {
	return dsc.nodeClassifier.GetStrategyForType(nodeType)
}

func (dsc *DefaultStrategyClassifier) ClassifyWithContext(node adf.Node, context adf.ConversionContext) adf.ConversionStrategy {
	strategy := dsc.ClassifyNode(node)
	switch context.Strategy {
	case adf.XMLPreserved:
		if strategy == adf.StandardMarkdown {
			return adf.XMLPreserved
		}
	case adf.MarkdownTable:
		if node.Type == adf.NodeTypeTableCell || node.Type == adf.NodeTypeTableHeader {
			return adf.MarkdownTable
		}
	case adf.MarkdownTaskList:
		if node.Type == adf.NodeTypeTaskItem {
			return adf.MarkdownTaskList
		}
	}
	if context.NestedLevel > 3 {
		if strategy == adf.XMLPreserved {
			return adf.Placeholder
		}
	}
	return strategy
}

func (dsc *DefaultStrategyClassifier) GetStrategyMetadata(strategy adf.ConversionStrategy) StrategyMetadata {
	switch strategy {
	case adf.StandardMarkdown:
		return StrategyMetadata{ReadabilityScore: 100, PreservesAttributes: false, RequiresPostProcess: false, SupportsNesting: true}
	case adf.MarkdownTable:
		return StrategyMetadata{ReadabilityScore: 90, PreservesAttributes: false, RequiresPostProcess: false, SupportsNesting: false}
	case adf.MarkdownTaskList:
		return StrategyMetadata{ReadabilityScore: 95, PreservesAttributes: true, RequiresPostProcess: false, SupportsNesting: true}
	case adf.MarkdownBlockquote:
		return StrategyMetadata{ReadabilityScore: 85, PreservesAttributes: false, RequiresPostProcess: false, SupportsNesting: true}
	case adf.XMLPreserved:
		return StrategyMetadata{ReadabilityScore: 60, PreservesAttributes: true, RequiresPostProcess: true, SupportsNesting: true}
	case adf.HTMLWrapped:
		return StrategyMetadata{ReadabilityScore: 70, PreservesAttributes: true, RequiresPostProcess: true, SupportsNesting: true}
	case adf.Placeholder:
		return StrategyMetadata{ReadabilityScore: 0, PreservesAttributes: true, RequiresPostProcess: true, SupportsNesting: true}
	default:
		return StrategyMetadata{}
	}
}

// StrategyMetadata provides information about conversion strategy characteristics.
type StrategyMetadata struct {
	ReadabilityScore    int
	PreservesAttributes bool
	RequiresPostProcess bool
	SupportsNesting     bool
}
