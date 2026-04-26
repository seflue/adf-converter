package elements

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
	"github.com/seflue/adf-converter/placeholder"
)

// inlineCardConverter handles conversion of ADF inlineCard nodes to/from markdown
type inlineCardConverter struct{}

func NewInlineCardConverter() adf.Renderer {
	return &inlineCardConverter{}
}

var complexMetadataAttrs = []string{"id", "space", "type", "version", "status", "localId", "key"}

func hasComplexMetadata(attrs map[string]any) bool {
	for _, attr := range complexMetadataAttrs {
		if _, exists := attrs[attr]; exists {
			return true
		}
	}
	return false
}

func buildComplexMetadataHTML(attrs map[string]any, linkURL string) string {
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

func (ic *inlineCardConverter) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)
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

	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)

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

func (ic *inlineCardConverter) dataOnlyToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	if context.PlaceholderManager != nil {
		placeholderID, preview, err := context.PlaceholderManager.Store(node)
		if err != nil {
			return adf.EnhancedConversionResult{}, fmt.Errorf("storing inlineCard placeholder: %w", err)
		}
		builder := convresult.NewEnhancedConversionResultBuilder(adf.Placeholder)
		if placeholderID == "" {
			builder.AppendContent(preview)
		} else {
			builder.AppendContent(placeholder.GeneratePlaceholderComment(placeholderID, preview))
		}
		return builder.Build(), nil
	}

	// No PlaceholderManager available — lossy fallback
	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)
	builder.AppendContent("[InlineCard]")
	return builder.Build(), nil
}

func (ic *inlineCardConverter) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("inlineCard is an inline element and should be parsed within parent blocks")
}

func (ic *inlineCardConverter) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypeInlineCard)
}

func (ic *inlineCardConverter) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (ic *inlineCardConverter) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be an Node")
	}

	if node.Type != adf.NodeTypeInlineCard {
		return fmt.Errorf("node type must be inlineCard, got: %s", node.Type)
	}

	return nil
}
