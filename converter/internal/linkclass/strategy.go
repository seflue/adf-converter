package linkclass

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/internal/nodeclass"
)

// ConversionStrategyMapping defines the complete mapping from link characteristics to strategy.
type ConversionStrategyMapping struct {
	IsInternal   bool
	HasMetadata  bool
	MetadataKeys []string
}

// DetermineStrategy analyzes link characteristics and returns the appropriate strategy.
func (csm *ConversionStrategyMapping) DetermineStrategy() converter.ConversionStrategy {
	if !csm.IsInternal {
		return converter.StandardMarkdown
	}
	if !csm.HasMetadata || len(csm.MetadataKeys) == 0 {
		return converter.StandardMarkdown
	}
	return converter.HTMLWrapped
}

func CreateMappingFromClassification(classification LinkClassification) ConversionStrategyMapping {
	return ConversionStrategyMapping{
		IsInternal:   classification.IsInternal,
		HasMetadata:  classification.HasMetadata,
		MetadataKeys: classification.MetadataKeys,
	}
}

// GetStrategyForLinkType returns the appropriate conversion strategy for a given link type.
func GetStrategyForLinkType(linkType LinkType) converter.ConversionStrategy {
	switch linkType {
	case WebLink:
		return converter.StandardMarkdown
	case SimpleInternalLink:
		return converter.StandardMarkdown
	case ComplexInternalLink:
		return converter.HTMLWrapped
	default:
		return converter.Placeholder
	}
}

// StrategyClassifier determines the appropriate conversion strategy for ADF nodes.
type StrategyClassifier interface {
	ClassifyNode(node adf_types.ADFNode) converter.ConversionStrategy
	RegisterNodeType(nodeType converter.ADFNodeType, strategy converter.ConversionStrategy, priority int)
	GetSupportedTypes() []converter.ADFNodeType
	GetStrategyForType(nodeType converter.ADFNodeType) converter.ConversionStrategy
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

func (dsc *DefaultStrategyClassifier) ClassifyNode(node adf_types.ADFNode) converter.ConversionStrategy {
	if node.Type == string(converter.NodeText) && len(node.Marks) > 0 {
		for _, mark := range node.Marks {
			if mark.Type == string(converter.MarkLink) {
				classification := dsc.linkClassifier.ClassifyLink(mark)
				return GetStrategyForLinkType(classification.Type)
			}
		}
	}
	if node.Type == adf_types.NodeTypeInlineCard {
		return converter.StandardMarkdown
	}
	return dsc.nodeClassifier.ClassifyNode(node)
}

func (dsc *DefaultStrategyClassifier) RegisterNodeType(nodeType converter.ADFNodeType, strategy converter.ConversionStrategy, priority int) {
	dsc.nodeClassifier.RegisterNodeType(nodeType, strategy, priority)
}

func (dsc *DefaultStrategyClassifier) GetSupportedTypes() []converter.ADFNodeType {
	return dsc.nodeClassifier.GetSupportedTypes()
}

func (dsc *DefaultStrategyClassifier) GetStrategyForType(nodeType converter.ADFNodeType) converter.ConversionStrategy {
	return dsc.nodeClassifier.GetStrategyForType(nodeType)
}

func (dsc *DefaultStrategyClassifier) ClassifyWithContext(node adf_types.ADFNode, context converter.ConversionContext) converter.ConversionStrategy {
	strategy := dsc.ClassifyNode(node)
	switch context.Strategy {
	case converter.XMLPreserved:
		if strategy == converter.StandardMarkdown {
			return converter.XMLPreserved
		}
	case converter.MarkdownTable:
		if node.Type == string(converter.NodeTableCell) || node.Type == string(converter.NodeTableHeader) {
			return converter.MarkdownTable
		}
	case converter.MarkdownTaskList:
		if node.Type == string(converter.NodeTaskItem) {
			return converter.MarkdownTaskList
		}
	}
	if context.NestedLevel > 3 {
		if strategy == converter.XMLPreserved {
			return converter.Placeholder
		}
	}
	return strategy
}

func (dsc *DefaultStrategyClassifier) GetStrategyMetadata(strategy converter.ConversionStrategy) StrategyMetadata {
	switch strategy {
	case converter.StandardMarkdown:
		return StrategyMetadata{ReadabilityScore: 100, PreservesAttributes: false, RequiresPostProcess: false, SupportsNesting: true}
	case converter.MarkdownTable:
		return StrategyMetadata{ReadabilityScore: 90, PreservesAttributes: false, RequiresPostProcess: false, SupportsNesting: false}
	case converter.MarkdownTaskList:
		return StrategyMetadata{ReadabilityScore: 95, PreservesAttributes: true, RequiresPostProcess: false, SupportsNesting: true}
	case converter.MarkdownBlockquote:
		return StrategyMetadata{ReadabilityScore: 85, PreservesAttributes: false, RequiresPostProcess: false, SupportsNesting: true}
	case converter.XMLPreserved:
		return StrategyMetadata{ReadabilityScore: 60, PreservesAttributes: true, RequiresPostProcess: true, SupportsNesting: true}
	case converter.HTMLWrapped:
		return StrategyMetadata{ReadabilityScore: 70, PreservesAttributes: true, RequiresPostProcess: true, SupportsNesting: true}
	case converter.Placeholder:
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
