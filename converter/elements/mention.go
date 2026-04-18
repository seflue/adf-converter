package elements

import (
	"fmt"
	"net/url"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

// MentionConverter handles conversion of ADF mention nodes to/from markdown
// Format: [@DisplayName](accountid:id?accessLevel=X&userType=Y)
type MentionConverter struct{}

func NewMentionConverter() converter.ElementConverter {
	return &MentionConverter{}
}

func (mc *MentionConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Attrs == nil {
		return converter.EnhancedConversionResult{}, fmt.Errorf("mention node missing attrs")
	}

	id, _ := node.Attrs["id"].(string)
	if id == "" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("mention node missing id attribute")
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

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent(fmt.Sprintf("[%s](%s)", text, destination))
	builder.IncrementConverted()
	return builder.Build(), nil
}

// buildMentionQuery builds query parameters from optional mention attributes
func buildMentionQuery(attrs map[string]interface{}) string {
	params := url.Values{}

	if accessLevel, ok := attrs["accessLevel"].(string); ok && accessLevel != "" {
		params.Set("accessLevel", accessLevel)
	}
	if userType, ok := attrs["userType"].(string); ok && userType != "" {
		params.Set("userType", userType)
	}

	return params.Encode()
}

func (mc *MentionConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, fmt.Errorf("mention is an inline element and should be parsed within parent blocks")
}

func (mc *MentionConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeMention)
}

func (mc *MentionConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (mc *MentionConverter) ValidateInput(input interface{}) error {
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
