package elements

import (
	"fmt"
	"net/url"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/converter/internal/convresult"
)

// mentionConverter handles conversion of ADF mention nodes to/from markdown
// Format: [@DisplayName](accountid:id?accessLevel=X&userType=Y)
type mentionConverter struct{}

func NewMentionConverter() element.Converter {
	return &mentionConverter{}
}

func (mc *mentionConverter) ToMarkdown(node adf_types.ADFNode, context element.ConversionContext) (element.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		return element.EnhancedConversionResult{}, fmt.Errorf("mention node missing attrs")
	}

	id, _ := node.Attrs["id"].(string)
	if id == "" {
		return element.EnhancedConversionResult{}, fmt.Errorf("mention node missing id attribute")
	}

	text, _ := node.Attrs["text"].(string)
	if text == "" {
		text = "@" + id
	}

	destination := "accountid:" + url.PathEscape(id)
	query := buildMentionQuery(node.Attrs)
	if query != "" {
		destination += "?" + query
	}

	builder := convresult.NewEnhancedConversionResultBuilder(element.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf("[%s](%s)", text, destination))
	builder.IncrementConverted()
	return builder.Build(), nil
}

// buildMentionQuery builds query parameters from optional mention attributes
func buildMentionQuery(attrs map[string]any) string {
	params := url.Values{}

	if accessLevel, ok := attrs["accessLevel"].(string); ok && accessLevel != "" {
		params.Set("accessLevel", accessLevel)
	}
	if userType, ok := attrs["userType"].(string); ok && userType != "" {
		params.Set("userType", userType)
	}

	return params.Encode()
}

func (mc *mentionConverter) FromMarkdown(lines []string, startIndex int, context element.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("mention is an inline element and should be parsed within parent blocks")
}

func (mc *mentionConverter) CanHandle(nodeType element.ADFNodeType) bool {
	return nodeType == element.ADFNodeType(adf_types.NodeTypeMention)
}

func (mc *mentionConverter) GetStrategy() element.ConversionStrategy {
	return element.StandardMarkdown
}

func (mc *mentionConverter) ValidateInput(input any) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypeMention {
		return fmt.Errorf("node type must be mention, got: %s", node.Type)
	}
	if node.Attrs == nil {
		return fmt.Errorf("mention node missing attrs")
	}
	if _, ok := node.Attrs["id"].(string); !ok {
		return fmt.Errorf("mention node missing id attribute")
	}
	return nil
}
