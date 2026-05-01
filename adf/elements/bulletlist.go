package elements

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/elements/internal/dedent"
	"github.com/seflue/adf-converter/adf/elements/internal/lists"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// bulletListRenderer handles conversion of ADF bullet list nodes to/from markdown
type bulletListRenderer struct{}

func NewBulletListRenderer() adf.Renderer {
	return &bulletListRenderer{}
}

func (blc *bulletListRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)

	childContext := context
	childContext.ListDepth = context.ListDepth + 1

	for _, item := range node.Content {
		itemRenderer , _ := context.Registry.Lookup(adf.NodeType(item.Type))
		if itemRenderer == nil {
			return adf.RenderResult{}, fmt.Errorf("no converter found for list item type: %s", item.Type)
		}

		itemResult, err := itemRenderer.ToMarkdown(item, childContext)
		if err != nil {
			return adf.RenderResult{}, fmt.Errorf("failed to convert list item: %w", err)
		}

		builder.AppendContent(itemResult.Content)
	}

	builder.AppendContent("\n")

	return builder.Build(), nil
}

func (blc *bulletListRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("no lines to parse")
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
		return adf.Node{}, 0, fmt.Errorf("no bullet list lines found")
	}

	// Parse only the list lines, but strip common indentation to prevent goldmark from treating it as code block
	// Preserve relative indentation for nesting and multi-line items
	listLines := lines[startIndex : startIndex+listLineCount]
	dedentedLines := dedent.DedentLines(listLines)
	markdown := strings.Join(dedentedLines, "\n")
	node, err := lists.ParseBulletList(markdown, context.PlaceholderManager)
	if err != nil {
		return adf.Node{}, 0, fmt.Errorf("goldmark list parser failed: %w", err)
	}

	// Consume trailing empty line if present
	consumed := listLineCount
	if startIndex+listLineCount < len(lines) && strings.TrimSpace(lines[startIndex+listLineCount]) == "" {
		consumed++
	}

	return node, consumed, nil
}

func (blc *bulletListRenderer) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "- ")
}

