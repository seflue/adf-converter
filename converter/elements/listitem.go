package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// ListItemConverter handles conversion of ADF list item nodes to/from markdown
type ListItemConverter struct{}

func NewListItemConverter() *ListItemConverter {
	return &ListItemConverter{}
}

func (lic *ListItemConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	depth := context.ListDepth
	if depth < 1 {
		depth = 1
	}
	indent := strings.Repeat("  ", depth-1)
	continuationIndent := indent + "  " // Additional 2 spaces for continuation lines

	builder.AppendContent(indent + "- ")

	for i, child := range node.Content {
		childConverter := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(child.Type))
		if childConverter == nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("no converter found for child type: %s", child.Type)
		}

		childResult, err := childConverter.ToMarkdown(child, context)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("failed to convert child node: %w", err)
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

func (lic *ListItemConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no lines to parse")
	}

	line := lines[startIndex]
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "-") {
		return adf_types.ADFNode{}, 0, fmt.Errorf("invalid list item format: %s", line)
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
		node := adf_types.ADFNode{
			Type:    adf_types.NodeTypeListItem,
			Content: []adf_types.ADFNode{},
		}
		return node, 1, nil
	}

	textNode := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: text,
	}

	paragraphNode := adf_types.ADFNode{
		Type:    adf_types.NodeTypeParagraph,
		Content: []adf_types.ADFNode{textNode},
	}

	node := adf_types.ADFNode{
		Type:    adf_types.NodeTypeListItem,
		Content: []adf_types.ADFNode{paragraphNode},
	}

	return node, 1, nil
}

func (lic *ListItemConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeListItem)
}

func (lic *ListItemConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (lic *ListItemConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeListItem {
		return fmt.Errorf("node type must be listItem, got: %s", node.Type)
	}

	return nil
}
