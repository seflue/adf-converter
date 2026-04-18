package elements

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/internal/convresult"
	"github.com/seflue/adf-converter/converter/elements/internal/lists"
)

// bulletListConverter handles conversion of ADF bullet list nodes to/from markdown
type bulletListConverter struct{}

func NewBulletListConverter() converter.ElementConverter {
	return &bulletListConverter{}
}

func (blc *bulletListConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := convresult.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	childContext := converter.ConversionContext{
		Strategy:           context.Strategy,
		PreserveAttrs:      context.PreserveAttrs,
		NestedLevel:        context.NestedLevel,
		ParentNodeType:     context.ParentNodeType,
		RoundTripMode:      context.RoundTripMode,
		ErrorRecovery:      context.ErrorRecovery,
		ListDepth:          context.ListDepth + 1,
		Classifier:         context.Classifier,
		PlaceholderManager: context.PlaceholderManager,
	}

	for _, item := range node.Content {
		itemConverter := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(item.Type))
		if itemConverter == nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("no converter found for list item type: %s", item.Type)
		}

		itemResult, err := itemConverter.ToMarkdown(item, childContext)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("failed to convert list item: %w", err)
		}

		builder.AppendContent(itemResult.Content)
	}

	builder.AppendContent("\n")

	return builder.Build(), nil
}

func (blc *bulletListConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no lines to parse")
	}

	// Count consecutive list lines starting from startIndex, including:
	// - Lines that start with bullet markers (-, *, +)
	// - Indented continuation lines (for multiline list items and nested lists)
	listLineCount := 0
	inList := false
	for i := startIndex; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Empty line ends the list
		if trimmed == "" {
			break
		}

		// Check if line starts with bullet marker (marker + space)
		isBulletLine := strings.HasPrefix(trimmed, "- ") ||
			strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "+ ")

		if isBulletLine {
			// This is a bullet line - always include
			inList = true
			listLineCount++
		} else if inList && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// This is an indented continuation line - include if we're in a list
			listLineCount++
		} else {
			// Non-indented, non-bullet line - end of list
			break
		}
	}

	if listLineCount == 0 {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no bullet list lines found")
	}

	// Parse only the list lines, but strip common indentation to prevent goldmark from treating it as code block
	// Preserve relative indentation for nesting and multi-line items
	listLines := lines[startIndex : startIndex+listLineCount]
	dedentedLines := DedentLines(listLines)
	markdown := strings.Join(dedentedLines, "\n")
	node, err := lists.ParseBulletList(markdown, context.PlaceholderManager)
	if err != nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("goldmark list parser failed: %w", err)
	}

	// Consume trailing empty line if present
	consumed := listLineCount
	if startIndex+listLineCount < len(lines) && strings.TrimSpace(lines[startIndex+listLineCount]) == "" {
		consumed++
	}

	return node, consumed, nil
}

func (blc *bulletListConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "- ")
}

func (blc *bulletListConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeBulletList)
}

func (blc *bulletListConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (blc *bulletListConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeBulletList {
		return fmt.Errorf("node type must be bulletList, got: %s", node.Type)
	}

	return nil
}
