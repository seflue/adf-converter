package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
	"github.com/seflue/adf-converter/placeholder"
)

// mediaSingleRegex parses Markdown image syntax with optional layout title:
//
//	![alt](url)                  → group1=alt, group2=url, group3=""
//	![alt](url "layout:wide")    → group1=alt, group2=url, group3="wide"
var mediaSingleRegex = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)\s"]+)(?:\s+"layout:([^"]+)")?\)\s*$`)

// mediaSingleRenderer handles conversion of ADF mediaSingle nodes to/from markdown.
//
// External images (media[type=external]) are converted to standard Markdown image syntax:
//
//	![alt](url)               // center layout (default, title suppressed)
//	![alt](url "layout:wide") // non-default layout encoded in title field
//
// Internal media (media[type=file/link] with id+collection) are preserved as placeholders.
type mediaSingleRenderer struct{}

func NewMediaSingleRenderer() adf.Renderer {
	return &mediaSingleRenderer{}
}

func (mc *mediaSingleRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if isExternalMedia(node) {
		return mc.externalToMarkdown(node)
	}
	return mc.internalToMarkdown(node, context)
}

func (mc *mediaSingleRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	if startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("no lines to parse at index %d", startIndex)
	}

	line := strings.TrimSpace(lines[startIndex])
	matches := mediaSingleRegex.FindStringSubmatch(line)
	if matches == nil {
		return adf.Node{}, 0, fmt.Errorf("not an external image: %s", line)
	}

	alt := matches[1]
	url := matches[2]
	layout := matches[3]
	if layout == "" {
		layout = "center"
	}

	node := adf.Node{
		Type: adf.NodeTypeMediaSingle,
		Attrs: map[string]any{
			"layout": layout,
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeMedia,
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

func (mc *mediaSingleRenderer) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "![")
}

// isExternalMedia returns true if node has a media child with type="external".
func isExternalMedia(node adf.Node) bool {
	if len(node.Content) == 0 {
		return false
	}
	child := node.Content[0]
	if child.Type != adf.NodeTypeMedia || child.Attrs == nil {
		return false
	}
	mediaType, _ := child.Attrs["type"].(string)
	return mediaType == "external"
}

func (mc *mediaSingleRenderer) externalToMarkdown(node adf.Node) (adf.RenderResult, error) {
	media := node.Content[0]
	url, _ := media.Attrs["url"].(string)
	alt, _ := media.Attrs["alt"].(string)
	layout, _ := node.Attrs["layout"].(string)

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)

	if layout == "" || layout == "center" {
		builder.AppendContent(fmt.Sprintf("![%s](%s)\n\n", alt, url))
	} else {
		builder.AppendContent(fmt.Sprintf("![%s](%s \"layout:%s\")\n\n", alt, url, layout))
	}

	builder.IncrementConverted()
	return builder.Build(), nil
}

func (mc *mediaSingleRenderer) internalToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if context.PlaceholderManager == nil {
		builder := convresult.NewRenderResultBuilder(adf.Placeholder)
		builder.AppendContent("<!-- mediaSingle: preserved -->\n\n")
		return builder.Build(), nil
	}

	placeholderID, preview, err := context.PlaceholderManager.Store(node)
	if err != nil {
		return adf.RenderResult{}, fmt.Errorf("storing mediaSingle placeholder: %w", err)
	}

	builder := convresult.NewRenderResultBuilder(adf.Placeholder)
	if placeholderID == "" {
		builder.AppendContent(preview + "\n\n")
	} else {
		builder.AppendContent(placeholder.GeneratePlaceholderComment(placeholderID, preview) + "\n\n")
	}
	return builder.Build(), nil
}
