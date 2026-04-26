package elements

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// listItemConverter handles conversion of ADF list item nodes to/from markdown
type listItemConverter struct{}

func NewListItemConverter() adf.Renderer {
	return &listItemConverter{}
}

func (lic *listItemConverter) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.EnhancedConversionResult, error) {
	builder := convresult.NewEnhancedConversionResultBuilder(adf.StandardMarkdown)

	depth := context.ListDepth
	if depth < 1 {
		depth = 1
	}
	indent := strings.Repeat("  ", depth-1)
	continuationIndent := indent + "  " // Additional 2 spaces for continuation lines

	builder.AppendContent(indent + "- ")

	for i, child := range node.Content {
		childConverter , _ := context.Registry.Lookup(adf.NodeType(child.Type))
		if childConverter == nil {
			return adf.EnhancedConversionResult{}, fmt.Errorf("no converter found for child type: %s", child.Type)
		}

		childResult, err := childConverter.ToMarkdown(child, context)
		if err != nil {
			return adf.EnhancedConversionResult{}, fmt.Errorf("failed to convert child node: %w", err)
		}

		childContent := strings.TrimSuffix(childResult.Content, "\n\n")
		childContent = strings.TrimSuffix(childContent, "\n")

		// Handle multiline content: indent continuation lines
		lines := strings.Split(childContent, "\n")
		for j, line := range lines {
			if j == 0 {
				// First line goes right after the "- " marker
				builder.AppendContent(line)
			} else {
				// Continuation lines need proper indentation to stay part of list item
				builder.AppendContent("\n" + continuationIndent + line)
			}
		}

		// Add spacing between child nodes (e.g., multiple paragraphs in one list item)
		if i < len(node.Content)-1 {
			builder.AppendContent("\n" + continuationIndent)
		}
	}

	builder.AppendContent("\n")

	return builder.Build(), nil
}

func (lic *listItemConverter) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("no lines to parse")
	}

	line := lines[startIndex]
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "-") {
		return adf.Node{}, 0, fmt.Errorf("invalid list item format: %s", line)
	}

	var text string
	if len(trimmed) > 1 && trimmed[1] == ' ' {
		text = strings.TrimSpace(trimmed[2:])
	} else if len(trimmed) > 1 {
		text = strings.TrimSpace(trimmed[1:])
	} else {
		text = ""
	}

	if text == "" {
		node := adf.Node{
			Type:    adf.NodeTypeListItem,
			Content: []adf.Node{},
		}
		return node, 1, nil
	}

	textNode := adf.Node{
		Type: adf.NodeTypeText,
		Text: text,
	}

	paragraphNode := adf.Node{
		Type:    adf.NodeTypeParagraph,
		Content: []adf.Node{textNode},
	}

	node := adf.Node{
		Type:    adf.NodeTypeListItem,
		Content: []adf.Node{paragraphNode},
	}

	return node, 1, nil
}

func (lic *listItemConverter) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypeListItem)
}

func (lic *listItemConverter) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (lic *listItemConverter) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be an Node")
	}

	if node.Type != adf.NodeTypeListItem {
		return fmt.Errorf("node type must be listItem, got: %s", node.Type)
	}

	return nil
}
