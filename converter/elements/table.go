package elements

import (
	"fmt"
	"sort"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/elements/internal/inline"
	"github.com/seflue/adf-converter/converter/elements/internal/tables"
	"github.com/seflue/adf-converter/converter/internal"
	"github.com/seflue/adf-converter/converter/internal/convresult"
)

// tableConverter implements markdown table conversion for ADF table nodes
type tableConverter struct{}

func NewTableConverter() converter.ElementConverter {
	return &tableConverter{}
}

func (tc *tableConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Type != "table" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("table converter can only handle table nodes, got: %s", node.Type)
	}

	var markdown strings.Builder
	var headers []string
	var dataRows [][]string
	hasRealHeader := tc.firstRowIsHeader(node)

	for i, row := range node.Content {
		if row.Type != "tableRow" {
			continue
		}

		var cells []string
		for _, cell := range row.Content {
			cellText := tc.extractCellText(cell, context)
			cells = append(cells, cellText)
		}

		if i == 0 && hasRealHeader {
			headers = cells
		} else {
			dataRows = append(dataRows, cells)
		}
	}

	// Determine column count for synthetic header
	colCount := len(headers)
	if colCount == 0 && len(dataRows) > 0 {
		colCount = len(dataRows[0])
		// Synthetic empty header for tables without tableHeader cells
		headers = make([]string, colCount)
	}

	if colCount > 0 {
		markdown.WriteString("| ")
		markdown.WriteString(strings.Join(headers, " | "))
		markdown.WriteString(" |\n")

		markdown.WriteString("|")
		for i, header := range headers {
			headerLen := len(header)
			var dashCount int
			if headerLen >= 4 {
				dashCount = 6
			} else {
				dashCount = 2
			}
			separator := strings.Repeat("-", dashCount)

			if i == len(headers)-1 {
				markdown.WriteString(separator + "|\n")
			} else {
				markdown.WriteString(separator + "|")
			}
		}

		for _, row := range dataRows {
			markdown.WriteString("| ")
			markdown.WriteString(strings.Join(row, " | "))
			markdown.WriteString(" |\n")
		}
	}

	// Only use XML wrapper when non-default attributes need preserving
	var finalMarkdown string
	nonDefaultAttrs := filterDefaultTableAttrs(node.Attrs)
	if context.PreserveAttrs && len(nonDefaultAttrs) > 0 {
		wrappedMarkdown, err := tc.wrapTableWithXML(markdown.String(), nonDefaultAttrs)
		if err != nil {
			return convresult.CreateErrorResult(err.Error(), converter.MarkdownTable), err
		}
		finalMarkdown = wrappedMarkdown
	} else {
		finalMarkdown = markdown.String()
	}

	// Block-level elements need trailing double newline for spacing
	if !strings.HasSuffix(finalMarkdown, "\n\n") {
		if strings.HasSuffix(finalMarkdown, "\n") {
			finalMarkdown += "\n"
		} else {
			finalMarkdown += "\n\n"
		}
	}

	result := convresult.CreateSuccessResult(finalMarkdown, converter.MarkdownTable)
	result.ElementsConverted = 1

	// Preserve ADF attributes for round-trip fidelity
	if context.PreserveAttrs && node.Attrs != nil {
		result.PreservedAttrs = node.Attrs
	}

	return result, nil
}

// filterDefaultTableAttrs returns a copy of attrs with ADF spec defaults removed.
// Defaults per spec: layout="center", isNumberColumnEnabled=false, displayMode="default".
func filterDefaultTableAttrs(attrs map[string]interface{}) map[string]interface{} {
	if len(attrs) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range attrs {
		switch k {
		case "localId":
			continue // Jira regenerates on save
		case "layout":
			if s, ok := v.(string); ok && (s == "center" || s == "default") {
				continue
			}
		case "isNumberColumnEnabled":
			if b, ok := v.(bool); ok && !b {
				continue
			}
		case "displayMode":
			if s, ok := v.(string); ok && s == "default" {
				continue
			}
		}
		result[k] = v
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// firstRowIsHeader checks whether the first row contains tableHeader cells.
func (tc *tableConverter) firstRowIsHeader(node adf_types.ADFNode) bool {
	for _, row := range node.Content {
		if row.Type != "tableRow" {
			continue
		}
		return len(row.Content) > 0 && row.Content[0].Type == "tableHeader"
	}
	return false
}

// extractCellText extracts text content from a table cell, preserving markdown formatting
func (tc *tableConverter) extractCellText(cell adf_types.ADFNode, context converter.ConversionContext) string {
	var text strings.Builder

	for _, content := range cell.Content {
		if content.Type == "paragraph" {
			rendered, err := inline.RenderInlineNodes(content.Content, context)
			if err != nil {
				continue
			}
			text.WriteString(rendered)
		}
	}

	return text.String()
}

// FromMarkdown converts markdown table syntax back to ADF table node.
// Supports both plain markdown tables and XML-wrapped tables with ADF attributes.
//
// Plain markdown table:
//
//	| Header 1 | Header 2 |
//	|----------|----------|
//	| Cell 1   | Cell 2   |
//
// XML-wrapped table with attributes:
//
//	<table localId="abc123" layout="wide">
//	| Header 1 | Header 2 |
//	|----------|----------|
//	| Cell 1   | Cell 2   |
//	</table>
func (tc *tableConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{Type: "table", Content: []adf_types.ADFNode{}}, 0, nil
	}

	firstLine := strings.TrimSpace(lines[startIndex])

	// XML-wrapped table: consume from <table> to </table>
	if strings.HasPrefix(firstLine, "<table") {
		consumed := tc.countXMLTableLines(lines, startIndex)
		node, err := tc.parseXMLWrappedTable(lines[startIndex : startIndex+consumed])
		return node, consumed, err
	}

	// Plain markdown table: consume consecutive table lines
	consumed := tc.countPlainTableLines(lines, startIndex)
	if consumed == 0 {
		return adf_types.ADFNode{Type: "table", Content: []adf_types.ADFNode{}}, 0, nil
	}

	tableLines := lines[startIndex : startIndex+consumed]
	markdown := strings.Join(tableLines, "\n") + "\n"
	node, err := tables.ParseTable(markdown)
	return node, consumed, err
}

// countPlainTableLines counts consecutive lines that belong to a markdown table.
// A table line starts with |.
func (tc *tableConverter) countPlainTableLines(lines []string, startIndex int) int {
	count := 0
	for i := startIndex; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			break
		}
		count++
	}
	return count
}

// countXMLTableLines counts lines from <table...> to </table> inclusive.
func (tc *tableConverter) countXMLTableLines(lines []string, startIndex int) int {
	for i := startIndex; i < len(lines); i++ {
		if strings.Contains(strings.TrimSpace(lines[i]), "</table>") {
			return i - startIndex + 1
		}
	}
	// No closing tag found — return all remaining lines so parseXMLWrappedTable can report the error
	return len(lines) - startIndex
}

// parseXMLWrappedTable parses XML-wrapped markdown table with ADF attributes
func (tc *tableConverter) parseXMLWrappedTable(lines []string) (adf_types.ADFNode, error) {
	if len(lines) == 0 {
		return adf_types.ADFNode{}, fmt.Errorf("no lines to parse")
	}

	// Find the opening and closing tags
	startIdx := -1
	endIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<table") {
			startIdx = i
		} else if strings.HasPrefix(trimmed, "</table>") {
			endIdx = i
			break
		}
	}

	if startIdx == -1 || endIdx == -1 {
		return adf_types.ADFNode{}, fmt.Errorf("malformed XML table: missing opening or closing tag")
	}

	// Parse XML attributes from opening tag
	openTag := strings.TrimSpace(lines[startIdx])
	attrs := internal.ParseXMLAttributes(openTag)

	// Extract markdown table content between tags
	var markdownLines []string
	for i := startIdx + 1; i < endIdx; i++ {
		markdownLines = append(markdownLines, lines[i])
	}

	// Parse the markdown table content into ADF using Goldmark
	markdown := strings.Join(markdownLines, "\n") + "\n"
	tableNode, err := tables.ParseTable(markdown)
	if err != nil {
		return adf_types.ADFNode{}, fmt.Errorf("failed to parse markdown table content: %w", err)
	}

	// Add the preserved XML attributes to the table node
	if len(attrs) > 0 {
		if tableNode.Attrs == nil {
			tableNode.Attrs = make(map[string]interface{})
		}
		for key, value := range attrs {
			tableNode.Attrs[key] = value
		}
	}

	return tableNode, nil
}

// CanHandle returns true if this converter can handle the given node type
func (tc *tableConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "<table") || strings.HasPrefix(line, "|")
}

func (tc *tableConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.NodeTable
}

// GetStrategy returns the conversion strategy this converter implements
func (tc *tableConverter) GetStrategy() converter.ConversionStrategy {
	return converter.MarkdownTable
}

// ValidateInput validates that the input can be processed
func (tc *tableConverter) ValidateInput(input interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	switch v := input.(type) {
	case adf_types.ADFNode:
		if v.Type != "table" {
			return fmt.Errorf("ADF node must be of type 'table', got: %s", v.Type)
		}
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("markdown input cannot be empty")
		}
		return nil
	default:
		return fmt.Errorf("input must be adf_types.ADFNode or string, got: %T", input)
	}
}

// wrapTableWithXML wraps markdown table content with XML tags and ADF attributes.
// Writes all attrs from the map in a stable order (localId first, then sorted).
//
//nolint:unparam // error return reserved for future use
func (tc *tableConverter) wrapTableWithXML(markdownTable string, attrs map[string]interface{}) (string, error) {
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString("<table")

	keys := sortedAttrKeys(attrs)
	for _, k := range keys {
		fmt.Fprintf(&xmlBuilder, ` %s="%v"`, k, attrs[k])
	}

	xmlBuilder.WriteString(">\n")
	xmlBuilder.WriteString(markdownTable)
	if !strings.HasSuffix(markdownTable, "\n") {
		xmlBuilder.WriteString("\n")
	}
	xmlBuilder.WriteString("</table>")

	return xmlBuilder.String(), nil
}

// sortedAttrKeys returns attr keys with localId first, rest alphabetically.
func sortedAttrKeys(attrs map[string]interface{}) []string {
	var keys []string
	hasLocalId := false
	for k := range attrs {
		if k == "localId" {
			hasLocalId = true
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if hasLocalId {
		keys = append([]string{"localId"}, keys...)
	}
	return keys
}
