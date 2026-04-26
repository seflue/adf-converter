package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// blockCardRegex matches <div data-adf-type="blockCard">[url](url)</div> or bare url
var blockCardRegex = regexp.MustCompile(`^<div data-adf-type="blockCard">(?:\[([^\]]+)\]\([^)]+\)|(.+))</div>$`)

// blockCardRenderer handles conversion of ADF blockCard nodes to/from markdown.
//
// blockCard is a block-level smart link that Jira renders as a full-width card.
// Uses an HTML wrapper to preserve the type through roundtrip:
//
//	<div data-adf-type="blockCard">https://example.com</div>
type blockCardRenderer struct{}

func NewBlockCardRenderer() adf.Renderer {
	return &blockCardRenderer{}
}

func (bc *blockCardRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)

	url, _ := node.Attrs["url"].(string)
	if url == "" {
		builder.AppendContent("<div data-adf-type=\"blockCard\"></div>\n\n")
		return builder.Build(), nil
	}

	builder.AppendContent(fmt.Sprintf("<div data-adf-type=\"blockCard\">[%s](%s)</div>\n\n", url, url))
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (bc *blockCardRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	if startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("no lines to parse")
	}

	line := strings.TrimSpace(lines[startIndex])
	matches := blockCardRegex.FindStringSubmatch(line)
	if matches == nil {
		return adf.Node{}, 0, fmt.Errorf("not a blockCard: %s", line)
	}

	// Group 1 = link text from [url](url), Group 2 = bare url fallback
	url := matches[1]
	if url == "" {
		url = matches[2]
	}
	node := adf.Node{
		Type: adf.NodeTypeBlockCard,
		Attrs: map[string]any{
			"url": url,
		},
	}

	return node, 1, nil
}

func (bc *blockCardRenderer) CanParseLine(line string) bool {
	return strings.HasPrefix(line, `<div data-adf-type="blockCard"`)
}

func (bc *blockCardRenderer) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypeBlockCard)
}

func (bc *blockCardRenderer) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (bc *blockCardRenderer) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be a Node")
	}
	if node.Type != adf.NodeTypeBlockCard {
		return fmt.Errorf("node type must be blockCard, got: %s", node.Type)
	}
	return nil
}
