package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/internal/convresult"
	"github.com/seflue/adf-converter/converter/elements/internal/lists"
)

var orderedListLinePattern = regexp.MustCompile(`^\s*\d+\.\s`)

// OrderedListConverter handles conversion of ADF ordered list nodes to/from markdown
type OrderedListConverter struct{}

func NewOrderedListConverter() converter.ElementConverter {
	return &OrderedListConverter{}
}

func (olc *OrderedListConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := convresult.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	childContext := converter.ConversionContext{
		Strategy:       context.Strategy,
		PreserveAttrs:  context.PreserveAttrs,
		NestedLevel:    context.NestedLevel,
		ParentNodeType: context.ParentNodeType,
		RoundTripMode:  context.RoundTripMode,
		ErrorRecovery:  context.ErrorRecovery,
		ListDepth:      context.ListDepth + 1,
	}

	start := 1
	if order, ok := node.Attrs["order"]; ok {
		if v, ok := order.(float64); ok {
			start = int(v)
		}
	}

	for i, item := range node.Content {
		itemConverter := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(item.Type))
		if itemConverter == nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("no converter found for list item type: %s", item.Type)
		}

		itemResult, err := itemConverter.ToMarkdown(item, childContext)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("failed to convert list item: %w", err)
		}

		itemContent := replaceListMarker(itemResult.Content, start+i)

		builder.AppendContent(itemContent)
	}

	builder.AppendContent("\n")

	return builder.Build(), nil
}

func (olc *OrderedListConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	// Count consecutive list lines starting from startIndex, including:
	// - Lines that start with ordered list markers (1., 2., etc.)
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

		if isOrderedListLine(trimmed) {
			// This is an ordered list line - always include
			inList = true
			listLineCount++
		} else if inList && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// This is an indented continuation line - include if we're in a list
			listLineCount++
		} else {
			// Non-indented, non-ordered-list line - end of list
			break
		}
	}

	if listLineCount == 0 {
		return adf_types.ADFNode{}, 0, fmt.Errorf("no ordered list lines found")
	}

	// Parse only the list lines, but strip common indentation to prevent goldmark from treating it as code block
	// Preserve relative indentation for nesting and multi-line items
	listLines := lines[startIndex : startIndex+listLineCount]
	dedentedLines := DedentLines(listLines)
	markdown := strings.Join(dedentedLines, "\n")
	node, err := lists.ParseOrderedList(markdown, context.PlaceholderManager)
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

// replaceListMarker replaces the leading "- " marker in a list item with "N. ".
// Preserves any leading indentation before the dash.
func replaceListMarker(content string, num int) string {
	dashIndex := strings.Index(content, "- ")
	if dashIndex >= 0 {
		return fmt.Sprintf("%s%d. %s", content[:dashIndex], num, content[dashIndex+2:])
	}
	return fmt.Sprintf("%d. %s", num, content)
}

// isOrderedListLine reports whether trimmed starts with an ordered list marker (e.g. "1." or "1)").
func isOrderedListLine(trimmed string) bool {
	if len(trimmed) < 2 || trimmed[0] < '0' || trimmed[0] > '9' {
		return false
	}
	for j := 0; j < len(trimmed); j++ {
		if trimmed[j] == '.' || trimmed[j] == ')' {
			return true
		}
		if trimmed[j] < '0' || trimmed[j] > '9' {
			return false
		}
	}
	return false
}

func (olc *OrderedListConverter) CanParseLine(line string) bool {
	return orderedListLinePattern.MatchString(line)
}

func (olc *OrderedListConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeOrderedList)
}

func (olc *OrderedListConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (olc *OrderedListConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeOrderedList {
		return fmt.Errorf("node type must be orderedList, got: %s", node.Type)
	}

	return nil
}
