package elements

import (
	"fmt"
	"net/url"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// mentionRenderer handles conversion of ADF mention nodes to/from markdown
// Format: [@DisplayName](accountid:id?accessLevel=X&userType=Y)
type mentionRenderer struct{}

func NewMentionRenderer() adf.Renderer {
	return &mentionRenderer{}
}

func (mc *mentionRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if node.Attrs == nil {
		return adf.RenderResult{}, fmt.Errorf("mention node missing attrs")
	}

	id, _ := node.Attrs["id"].(string)
	if id == "" {
		return adf.RenderResult{}, fmt.Errorf("mention node missing id attribute")
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

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
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

func (mc *mentionRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("mention is an inline element and should be parsed within parent blocks")
}

func (mc *mentionRenderer) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypeMention)
}

func (mc *mentionRenderer) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (mc *mentionRenderer) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be a Node")
	}
	if node.Type != adf.NodeTypeMention {
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
