package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/converter/internal/convresult"
	"github.com/seflue/adf-converter/placeholder"
)

// mediaSingleRegex parses Markdown image syntax with optional layout title:
//
//	![alt](url)                  → group1=alt, group2=url, group3=""
//	![alt](url "layout:wide")    → group1=alt, group2=url, group3="wide"
var mediaSingleRegex = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)\s"]+)(?:\s+"layout:([^"]+)")?\)\s*$`)

// mediaSingleConverter handles conversion of ADF mediaSingle nodes to/from markdown.
//
// External images (media[type=external]) are converted to standard Markdown image syntax:
//
//	![alt](url)               // center layout (default, title suppressed)
//	![alt](url "layout:wide") // non-default layout encoded in title field
//
// Internal media (media[type=file/link] with id+collection) are preserved as placeholders.
type mediaSingleConverter struct{}

func NewMediaSingleConverter() element.Converter {
	return &mediaSingleConverter{}
}

func (mc *mediaSingleConverter) ToMarkdown(node adf_types.ADFNode, context element.ConversionContext) (element.EnhancedConversionResult, error) {
	if isExternalMedia(node) {
		return mc.externalToMarkdown(node)
	}
	return mc.internalToMarkdown(node, context)
}

func (mc *mediaSingleConverter) FromMarkdown(lines []string, startIndex int, context element.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no lines to parse at index %d", startIndex)
	}

	line := strings.TrimSpace(lines[startIndex])
	matches := mediaSingleRegex.FindStringSubmatch(line)
	if matches == nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("not an external image: %s", line)
	}

	alt := matches[1]
	url := matches[2]
	layout := matches[3]
	if layout == "" {
		layout = "center"
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeMediaSingle,
		Attrs: map[string]any{
			"layout": layout,
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeMedia,
				Attrs: map[string]any{
					"type": "external",
					"url":  url,
					"alt":  alt,
				},
			},
		},
	}

	return node, 1, nil
}

func (mc *mediaSingleConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "![")
}

func (mc *mediaSingleConverter) CanHandle(nodeType element.ADFNodeType) bool {
	return nodeType == element.ADFNodeType(adf_types.NodeTypeMediaSingle)
}

func (mc *mediaSingleConverter) GetStrategy() element.ConversionStrategy {
	return element.StandardMarkdown
}

func (mc *mediaSingleConverter) ValidateInput(input any) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypeMediaSingle {
		return fmt.Errorf("node type must be mediaSingle, got: %s", node.Type)
	}
	return nil
}

// isExternalMedia returns true if node has a media child with type="external".
func isExternalMedia(node adf_types.ADFNode) bool {
	if len(node.Content) == 0 {
		return false
	}
	child := node.Content[0]
	if child.Type != adf_types.NodeTypeMedia || child.Attrs == nil {
		return false
	}
	mediaType, _ := child.Attrs["type"].(string)
	return mediaType == "external"
}

func (mc *mediaSingleConverter) externalToMarkdown(node adf_types.ADFNode) (element.EnhancedConversionResult, error) {
	media := node.Content[0]
	url, _ := media.Attrs["url"].(string)
	alt, _ := media.Attrs["alt"].(string)
	layout, _ := node.Attrs["layout"].(string)

	builder := convresult.NewEnhancedConversionResultBuilder(element.StandardMarkdown)

	if layout == "" || layout == "center" {
		builder.AppendContent(fmt.Sprintf("![%s](%s)\n\n", alt, url))
	} else {
		builder.AppendContent(fmt.Sprintf("![%s](%s \"layout:%s\")\n\n", alt, url, layout))
	}

	builder.IncrementConverted()
	return builder.Build(), nil
}

func (mc *mediaSingleConverter) internalToMarkdown(node adf_types.ADFNode, context element.ConversionContext) (element.EnhancedConversionResult, error) {
	if context.PlaceholderManager == nil {
		builder := convresult.NewEnhancedConversionResultBuilder(element.Placeholder)
		builder.AppendContent("<!-- mediaSingle: preserved -->\n\n")
		return builder.Build(), nil
	}

	placeholderID, preview, err := context.PlaceholderManager.Store(node)
	if err != nil {
		return element.EnhancedConversionResult{}, fmt.Errorf("storing mediaSingle placeholder: %w", err)
	}

	builder := convresult.NewEnhancedConversionResultBuilder(element.Placeholder)
	if placeholderID == "" {
		builder.AppendContent(preview + "\n\n")
	} else {
		builder.AppendContent(placeholder.GeneratePlaceholderComment(placeholderID, preview) + "\n\n")
	}
	return builder.Build(), nil
}
