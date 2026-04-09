package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/placeholder"
)

// InlineCardConverter handles conversion of ADF inlineCard nodes to/from markdown
type InlineCardConverter struct{}

func NewInlineCardConverter() *InlineCardConverter {
	return &InlineCardConverter{}
}

var complexMetadataAttrs = []string{"id", "space", "type", "version", "status", "localId", "key"}

func hasComplexMetadata(attrs map[string]interface{}) bool {
	for _, attr := range complexMetadataAttrs {
		if _, exists := attrs[attr]; exists {
			return true
		}
	}
	return false
}

func buildComplexMetadataHTML(attrs map[string]interface{}, linkURL string) string {
	var b strings.Builder
	b.WriteString("<a")
	for _, attr := range complexMetadataAttrs {
		value, exists := attrs[attr]
		if !exists {
			continue
		}
		if strValue, ok := value.(string); ok {
			fmt.Fprintf(&b, ` %s="%s"`, attr, strValue)
		} else if intValue, ok := value.(int); ok {
			fmt.Fprintf(&b, ` %s="%d"`, attr, intValue)
		}
	}
	b.WriteString(">")
	if linkURL != "" {
		fmt.Fprintf(&b, "[%s](%s)", linkURL, linkURL)
	} else {
		b.WriteString("[InlineCard]")
	}
	b.WriteString("</a>")
	return b.String()
}

func (ic *InlineCardConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
		builder.AppendContent("[InlineCard]")
		return builder.Build(), nil
	}

	linkURL, _ := node.Attrs["url"].(string)

	// data-only inlineCards can't be edited as markdown — preserve as placeholder
	if linkURL == "" {
		if _, hasData := node.Attrs["data"]; hasData {
			return ic.dataOnlyToMarkdown(node, context)
		}
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	if hasComplexMetadata(node.Attrs) {
		builder.AppendContent(buildComplexMetadataHTML(node.Attrs, linkURL))
		return builder.Build(), nil
	}

	if linkURL != "" {
		builder.AppendContent(fmt.Sprintf("[%s](%s)", linkURL, linkURL))
	} else {
		builder.AppendContent("[InlineCard]")
	}
	return builder.Build(), nil
}

func (ic *InlineCardConverter) dataOnlyToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if context.PlaceholderManager != nil {
		placeholderID, preview, err := context.PlaceholderManager.Store(node)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("storing inlineCard placeholder: %w", err)
		}
		builder := converter.NewEnhancedConversionResultBuilder(converter.Placeholder)
		if placeholderID == "" {
			builder.AppendContent(preview)
		} else {
			builder.AppendContent(placeholder.GeneratePlaceholderComment(placeholderID, preview))
		}
		return builder.Build(), nil
	}

	// No PlaceholderManager available — lossy fallback
	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent("[InlineCard]")
	return builder.Build(), nil
}

func (ic *InlineCardConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("inlineCard is an inline element and should be parsed within parent blocks")
}

func (ic *InlineCardConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeInlineCard)
}

func (ic *InlineCardConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (ic *InlineCardConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeInlineCard {
		return fmt.Errorf("node type must be inlineCard, got: %s", node.Type)
	}

	return nil
}
