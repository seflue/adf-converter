package elements

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/converter/elements/inline"
)

var orderedListPattern = regexp.MustCompile(`^\s*\d+\.\s`)

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

	// Separate preserved nodes into placeholders, pass rest to inline renderer
	var renderableContent []adf_types.ADFNode
	for _, child := range node.Content {
		if context.Classifier != nil && context.Classifier.IsPreserved(child.Type) {
			placeholderID, preview, err := context.PlaceholderManager.Store(child)
			if err != nil {
				return converter.EnhancedConversionResult{}, fmt.Errorf("failed to store placeholder for %s: %w", child.Type, err)
			}

			comment := fmt.Sprintf("<!-- %s: %s -->", placeholderID, preview)

			// Flush accumulated nodes before placeholder
			if len(renderableContent) > 0 {
				rendered, err := inline.RenderInlineNodes(renderableContent, context)
				if err != nil {
					return converter.EnhancedConversionResult{}, err
				}
				builder.AppendContent(rendered)
				renderableContent = nil
			}

			if adf_types.IsInlineNode(child.Type) {
				builder.AppendContent(comment)
			} else {
				builder.AppendContent(comment + "\n\n")
			}
			continue
		}

		renderableContent = append(renderableContent, child)
	}

	// Render remaining inline nodes with mark spanning
	if len(renderableContent) > 0 {
		rendered, err := inline.RenderInlineNodes(renderableContent, context)
		if err != nil {
			return converter.EnhancedConversionResult{}, err
		}
		builder.AppendContent(rendered)
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
			strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "<!--") ||
			strings.HasPrefix(trimmed, "```") {
			consumed = i - startIndex
			break
		}

		if orderedListPattern.MatchString(line) {
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

	textNodes, err := inline.ParseContentWithPlaceholders(text, context.PlaceholderManager)
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
