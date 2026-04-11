package elements

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/elements/inline"
	"github.com/seflue/adf-converter/placeholder"
)

// isBlockBoundary checks if a trimmed line starts a new block element
// by asking all registered BlockParsers. This avoids hardcoding tag names.
func isBlockBoundary(trimmed string) bool {
	for _, entry := range converter.GetGlobalRegistry().BlockParsers() {
		if entry.Parser.CanParseLine(trimmed) {
			return true
		}
	}
	return false
}

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

	// Separate preserved nodes into placeholders, pass rest to inline renderer.
	// Unknown inline types (anything IsInlineNode does not recognize) also take
	// the placeholder path so their attrs/nesting survive roundtrip — otherwise
	// the inline renderer would either drop them or fail hard.
	var renderableContent []adf_types.ADFNode
	for _, child := range node.Content {
		isUnknownInline := !adf_types.IsInlineNode(child.Type)
		shouldPreserve := context.Classifier != nil && context.Classifier.IsPreserved(child.Type)
		if shouldPreserve || isUnknownInline {
			var err error
			renderableContent, err = appendPreservedChild(child, renderableContent, context, builder)
			if err != nil {
				return converter.EnhancedConversionResult{}, err
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

		if isBlockBoundary(trimmed) {
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

// appendPreservedChild flushes any pending inline nodes, stores child as placeholder,
// and appends the placeholder comment to builder. Returns empty pending slice.
func appendPreservedChild(
	child adf_types.ADFNode,
	pending []adf_types.ADFNode,
	context converter.ConversionContext,
	builder *converter.EnhancedConversionResultBuilder,
) ([]adf_types.ADFNode, error) {
	placeholderID, preview, err := context.PlaceholderManager.Store(child)
	if err != nil {
		return nil, fmt.Errorf("failed to store placeholder for %s: %w", child.Type, err)
	}

	if len(pending) > 0 {
		rendered, err := inline.RenderInlineNodes(pending, context)
		if err != nil {
			return nil, err
		}
		builder.AppendContent(rendered)
	}

	// appendPreservedChild is only called from ParagraphConverter, so every
	// child is inline by construction — even if IsInlineNode does not yet
	// recognize its type (unknown inline nodes, ac-0073).
	if placeholderID == "" {
		builder.AppendContent(preview)
	} else {
		comment := placeholder.GeneratePlaceholderComment(placeholderID, preview)
		builder.AppendContent(comment)
	}

	return nil, nil
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
