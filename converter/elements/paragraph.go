package elements

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/converter/elements/inline"
)

// ParagraphConverter handles conversion of ADF paragraph nodes to/from markdown
type ParagraphConverter struct{}

func NewParagraphConverter() *ParagraphConverter {
	return &ParagraphConverter{}
}

func (pc *ParagraphConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if len(node.Content) == 0 {
		builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
		builder.AppendContent("\n")
		return builder.Build(), nil
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	for _, child := range node.Content {
		// Check if this is a preserved inline node (e.g., emoji, mention, date)
		if context.Classifier != nil && context.Classifier.IsPreserved(child.Type) {
			// Handle preserved inline nodes
			placeholderID, preview, err := context.PlaceholderManager.Store(child)
			if err != nil {
				return converter.EnhancedConversionResult{}, fmt.Errorf("failed to store placeholder for %s: %w", child.Type, err)
			}

			comment := fmt.Sprintf("<!-- %s: %s -->", placeholderID, preview)

			// Inline nodes: just the comment, no spacing (surrounding text provides spacing)
			// Block nodes: add double newline for block separation
			if adf_types.IsInlineNode(child.Type) {
				builder.AppendContent(comment)
			} else {
				builder.AppendContent(comment + "\n\n")
			}
			continue
		}

		childConverter := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(child.Type))
		if childConverter == nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("no converter found for child type: %s", child.Type)
		}

		childResult, err := childConverter.ToMarkdown(child, context)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("failed to convert child node: %w", err)
		}

		builder.AppendContent(childResult.Content)
	}

	builder.AppendContent("\n\n")

	return builder.Build(), nil
}

func (pc *ParagraphConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf_types.ADFNode{}, 1, nil
	}

	var paragraphLines []string
	consumed := 0

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			consumed = i - startIndex + 1
			break
		}

		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "- ") ||
			strings.HasPrefix(trimmed, "<!--") ||
			strings.HasPrefix(trimmed, "```") {
			consumed = i - startIndex
			break
		}

		if matched, _ := regexp.MatchString(`^\s*\d+\.\s`, line); matched {
			consumed = i - startIndex
			break
		}

		paragraphLines = append(paragraphLines, line)

		if i == len(lines)-1 {
			consumed = i - startIndex + 1
		}
	}

	if len(paragraphLines) == 0 {
		return adf_types.ADFNode{}, consumed, nil
	}

	text := strings.Join(paragraphLines, " ")
	text = strings.TrimSpace(text)

	if text == "" {
		return adf_types.ADFNode{}, consumed, nil
	}

	textNodes, err := inline.ParseContent(text)
	if err != nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("failed to parse inline content: %w", err)
	}

	node := adf_types.ADFNode{
		Type:    adf_types.NodeTypeParagraph,
		Content: textNodes,
	}

	return node, consumed, nil
}

func (pc *ParagraphConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeParagraph)
}

func (pc *ParagraphConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (pc *ParagraphConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeParagraph {
		return fmt.Errorf("node type must be paragraph, got: %s", node.Type)
	}

	return nil
}
