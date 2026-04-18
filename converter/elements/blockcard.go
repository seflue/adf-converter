package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

// blockCardRegex matches <div data-adf-type="blockCard">[url](url)</div> or bare url
var blockCardRegex = regexp.MustCompile(`^<div data-adf-type="blockCard">(?:\[([^\]]+)\]\([^)]+\)|(.+))</div>$`)

// BlockCardConverter handles conversion of ADF blockCard nodes to/from markdown.
//
// blockCard is a block-level smart link that Jira renders as a full-width card.
// Uses an HTML wrapper to preserve the type through roundtrip:
//
//	<div data-adf-type="blockCard">https://example.com</div>
type BlockCardConverter struct{}

func NewBlockCardConverter() converter.ElementConverter {
	return &BlockCardConverter{}
}

func (bc *BlockCardConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	url, _ := node.Attrs["url"].(string)
	if url == "" {
		builder.AppendContent("<div data-adf-type=\"blockCard\"></div>\n\n")
		return builder.Build(), nil
	}

	builder.AppendContent(fmt.Sprintf("<div data-adf-type=\"blockCard\">[%s](%s)</div>\n\n", url, url))
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (bc *BlockCardConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no lines to parse")
	}

	line := strings.TrimSpace(lines[startIndex])
	matches := blockCardRegex.FindStringSubmatch(line)
	if matches == nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("not a blockCard: %s", line)
	}

	// Group 1 = link text from [url](url), Group 2 = bare url fallback
	url := matches[1]
	if url == "" {
		url = matches[2]
	}
	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeBlockCard,
		Attrs: map[string]interface{}{
			"url": url,
		},
	}

	return node, 1, nil
}

func (bc *BlockCardConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, `<div data-adf-type="blockCard"`)
}

func (bc *BlockCardConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeBlockCard)
}

func (bc *BlockCardConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (bc *BlockCardConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypeBlockCard {
		return fmt.Errorf("node type must be blockCard, got: %s", node.Type)
	}
	return nil
}
