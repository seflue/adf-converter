package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter/internal"
)

// TableStructure represents the parsed structure of an ADF table
type TableStructure struct {
	RowCount    int
	ColumnCount int
	HasHeaders  bool
	Headers     []string
	Rows        []TableRowStructure
}

// TableRowStructure represents a row within a table
type TableRowStructure struct {
	IsHeader bool
	Cells    []TableCellStructure
}

// TableCellStructure represents a cell within a table row
type TableCellStructure struct {
	Content  string
	IsHeader bool
}

// TableConverter implements markdown table conversion for ADF table nodes
type TableConverter struct{}

func NewTableConverter() *TableConverter {
	return &TableConverter{}
}

func (tc *TableConverter) ToMarkdown(node adf_types.ADFNode, context ConversionContext) (EnhancedConversionResult, error) {
	if node.Type != "table" {
		return EnhancedConversionResult{}, fmt.Errorf("table converter can only handle table nodes, got: %s", node.Type)
	}

	var markdown strings.Builder
	var headers []string
	var rows [][]string

	for _, row := range node.Content {
		if row.Type != "tableRow" {
			continue
		}

		var cells []string
		isHeaderRow := len(headers) == 0

		for _, cell := range row.Content {
			cellText := tc.extractCellText(cell)
			cells = append(cells, cellText)
		}

		if isHeaderRow {
			headers = cells
		} else {
			rows = append(rows, cells)
		}
	}

	if len(headers) > 0 {
		markdown.WriteString("| ")
		markdown.WriteString(strings.Join(headers, " | "))
		markdown.WriteString(" |\n")

		markdown.WriteString("|")
		for i, header := range headers {
			// Create separator based on header length: 6 dashes for 4+ chars, 5 for 3 chars
			headerLen := len(header)
			var dashCount int
			if headerLen >= 4 {
				dashCount = 6
			} else {
				dashCount = 5
			}
			separator := strings.Repeat("-", dashCount)

			if i == len(headers)-1 {
				markdown.WriteString(separator + "|\n")
			} else {
				markdown.WriteString(separator + "|")
			}
		}

		for _, row := range rows {
			markdown.WriteString("| ")
			markdown.WriteString(strings.Join(row, " | "))
			markdown.WriteString(" |\n")
		}
	}

	// Only use XML wrapper when there are ADF attributes to preserve
	var finalMarkdown string
	if context.PreserveAttrs && node.Attrs != nil && len(node.Attrs) > 0 {
		wrappedMarkdown, err := tc.wrapTableWithXML(markdown.String(), node.Attrs)
		if err != nil {
			return CreateErrorResult(err.Error(), MarkdownTable), err
		}
		finalMarkdown = wrappedMarkdown
	} else {
		finalMarkdown = markdown.String()
	}

	result := CreateSuccessResult(finalMarkdown, MarkdownTable)
	result.ElementsConverted = 1

	// Preserve ADF attributes for round-trip fidelity
	if context.PreserveAttrs && node.Attrs != nil {
		result.PreservedAttrs = node.Attrs
	}

	return result, nil
}

// extractCellText extracts text content from a table cell, preserving markdown formatting
func (tc *TableConverter) extractCellText(cell adf_types.ADFNode) string {
	var text strings.Builder

	for _, content := range cell.Content {
		if content.Type == "paragraph" {
			text.WriteString(tc.convertParagraphToMarkdown(content))
		}
	}

	return text.String()
}

// convertParagraphToMarkdown converts paragraph content to markdown, handling marks
func (tc *TableConverter) convertParagraphToMarkdown(paragraph adf_types.ADFNode) string {
	var result strings.Builder

	for _, textNode := range paragraph.Content {
		if textNode.Type == "text" {
			text := textNode.Text

			for _, mark := range textNode.Marks {
				switch mark.Type {
				case "strong":
					text = "**" + text + "**"
				case "em":
					text = "*" + text + "*"
				case "code":
					text = "`" + text + "`"
				}
			}

			result.WriteString(text)
		}
	}

	return result.String()
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
func (tc *TableConverter) FromMarkdown(markdown string, context ConversionContext) (adf_types.ADFNode, error) {
	lines := strings.Split(strings.TrimSpace(markdown), "\n")
	if len(lines) == 0 {
		return adf_types.ADFNode{Type: "table", Content: []adf_types.ADFNode{}}, nil
	}

	// Check if this is an XML-wrapped table
	firstLine := strings.TrimSpace(lines[0])
	if strings.HasPrefix(firstLine, "<table") {
		return tc.parseXMLWrappedTable(lines)
	}

	// Parse as plain markdown table
	return tc.parseMarkdownTableLines(lines)
}

// parseXMLWrappedTable parses XML-wrapped markdown table with ADF attributes
func (tc *TableConverter) parseXMLWrappedTable(lines []string) (adf_types.ADFNode, error) {
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

	// Parse the markdown table content into ADF
	tableNode, err := tc.parseMarkdownTableLines(markdownLines)
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

// parseMarkdownTableLines parses plain markdown table lines into ADF table node
func (tc *TableConverter) parseMarkdownTableLines(lines []string) (adf_types.ADFNode, error) {
	if len(lines) < 2 {
		return adf_types.ADFNode{Type: "table", Content: []adf_types.ADFNode{}}, nil
	}

	var tableRows []adf_types.ADFNode

	// Process each line as a potential table row
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip separator row (contains only |, -, and spaces)
		if i == 1 && strings.Contains(trimmed, "---") {
			continue
		}

		// Parse table row
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			cells := strings.Split(trimmed[1:len(trimmed)-1], "|")
			var cellNodes []adf_types.ADFNode

			isHeader := (i == 0) // First row is header

			for _, cell := range cells {
				cellText := strings.TrimSpace(cell)
				cellType := "tableCell"
				if isHeader {
					cellType = "tableHeader"
				}

				cellNodes = append(cellNodes, adf_types.ADFNode{
					Type: cellType,
					Content: []adf_types.ADFNode{
						{
							Type: "paragraph",
							Content: []adf_types.ADFNode{
								{Type: "text", Text: cellText},
							},
						},
					},
				})
			}

			tableRows = append(tableRows, adf_types.ADFNode{
				Type:    "tableRow",
				Content: cellNodes,
			})
		}
	}

	return adf_types.ADFNode{
		Type:    "table",
		Content: tableRows,
	}, nil
}

// CanHandle returns true if this converter can handle the given node type
func (tc *TableConverter) CanHandle(nodeType ADFNodeType) bool {
	return nodeType == NodeTable
}

// GetStrategy returns the conversion strategy this converter implements
func (tc *TableConverter) GetStrategy() ConversionStrategy {
	return MarkdownTable
}

// ValidateInput validates that the input can be processed
func (tc *TableConverter) ValidateInput(input interface{}) error {
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

// ParseTableStructure parses an ADF table node and returns its structure
func (tc *TableConverter) ParseTableStructure(node adf_types.ADFNode) (TableStructure, error) {
	if node.Type != "table" {
		return TableStructure{}, fmt.Errorf("node must be of type 'table', got: %s", node.Type)
	}

	structure := TableStructure{
		RowCount:    len(node.Content),
		ColumnCount: 0,
		HasHeaders:  false,
		Headers:     []string{},
		Rows:        []TableRowStructure{},
	}

	for rowIdx, row := range node.Content {
		if row.Type != "tableRow" {
			continue
		}

		rowStructure := tc.parseTableRow(row, rowIdx == 0)

		if rowStructure.IsHeader {
			structure.HasHeaders = true
			for _, cell := range rowStructure.Cells {
				structure.Headers = append(structure.Headers, cell.Content)
			}
		}

		// Update column count if this row has more columns
		if len(rowStructure.Cells) > structure.ColumnCount {
			structure.ColumnCount = len(rowStructure.Cells)
		}

		structure.Rows = append(structure.Rows, rowStructure)
	}

	return structure, nil
}

// parseTableRow parses a single table row and returns its structure
func (tc *TableConverter) parseTableRow(row adf_types.ADFNode, isFirstRow bool) TableRowStructure {
	rowStructure := TableRowStructure{
		IsHeader: false,
		Cells:    []TableCellStructure{},
	}

	// Determine if this is a header row (first row with tableHeader cells)
	isHeaderRow := isFirstRow && len(row.Content) > 0 && row.Content[0].Type == "tableHeader"

	if isHeaderRow {
		rowStructure.IsHeader = true
	}

	for _, cell := range row.Content {
		cellContent := tc.extractCellText(cell)

		cellStructure := TableCellStructure{
			Content:  cellContent,
			IsHeader: cell.Type == "tableHeader",
		}

		rowStructure.Cells = append(rowStructure.Cells, cellStructure)
	}

	return rowStructure
}

// GenerateMarkdownTable generates markdown table syntax from a table structure
func (tc *TableConverter) GenerateMarkdownTable(structure TableStructure) (string, error) {
	if structure.ColumnCount == 0 {
		return "", fmt.Errorf("table structure has no columns")
	}

	var markdown strings.Builder

	if structure.HasHeaders && len(structure.Headers) > 0 {
		markdown.WriteString("| ")
		markdown.WriteString(strings.Join(structure.Headers, " | "))
		markdown.WriteString(" |\n")

		markdown.WriteString("|")
		for i := 0; i < structure.ColumnCount; i++ {
			if i == structure.ColumnCount-1 {
				markdown.WriteString("-----|\n")
			} else {
				markdown.WriteString("------|")
			}
		}
	}

	for _, row := range structure.Rows {
		if row.IsHeader {
			continue
		}

		markdown.WriteString("| ")
		var cellContents []string
		for _, cell := range row.Cells {
			cellContents = append(cellContents, cell.Content)
		}
		markdown.WriteString(strings.Join(cellContents, " | "))
		markdown.WriteString(" |\n")
	}

	return markdown.String(), nil
}

// ProcessTableHeadersAndCells processes an ADF table and handles headers and cells with formatting
func (tc *TableConverter) ProcessTableHeadersAndCells(node adf_types.ADFNode) (TableStructure, error) {
	if node.Type != "table" {
		return TableStructure{}, fmt.Errorf("node must be of type 'table', got: %s", node.Type)
	}

	structure := TableStructure{
		RowCount:    len(node.Content),
		ColumnCount: 0,
		HasHeaders:  false,
		Headers:     []string{},
		Rows:        []TableRowStructure{},
	}

	for rowIdx, row := range node.Content {
		if row.Type != "tableRow" {
			continue
		}

		rowStructure := tc.processTableRowWithFormatting(row, rowIdx == 0)

		if rowStructure.IsHeader {
			structure.HasHeaders = true
			for _, cell := range rowStructure.Cells {
				structure.Headers = append(structure.Headers, cell.Content)
			}
		}

		// Update column count if this row has more columns
		if len(rowStructure.Cells) > structure.ColumnCount {
			structure.ColumnCount = len(rowStructure.Cells)
		}

		structure.Rows = append(structure.Rows, rowStructure)
	}

	return structure, nil
}

// processTableRowWithFormatting processes a table row and handles cell content with markdown formatting
func (tc *TableConverter) processTableRowWithFormatting(row adf_types.ADFNode, isFirstRow bool) TableRowStructure {
	rowStructure := TableRowStructure{
		IsHeader: false,
		Cells:    []TableCellStructure{},
	}

	// Determine if this is a header row (first row with tableHeader cells)
	isHeaderRow := isFirstRow && len(row.Content) > 0 && row.Content[0].Type == "tableHeader"

	if isHeaderRow {
		rowStructure.IsHeader = true
	}

	for _, cell := range row.Content {
		cellContent := tc.extractCellTextWithFormatting(cell)

		cellStructure := TableCellStructure{
			Content:  cellContent,
			IsHeader: cell.Type == "tableHeader",
		}

		rowStructure.Cells = append(rowStructure.Cells, cellStructure)
	}

	return rowStructure
}

// extractCellTextWithFormatting extracts text content from a table cell with markdown formatting
func (tc *TableConverter) extractCellTextWithFormatting(cell adf_types.ADFNode) string {
	var text strings.Builder

	for _, content := range cell.Content {
		if content.Type == "paragraph" {
			text.WriteString(tc.processParagraphWithFormatting(content))
		}
	}

	return text.String()
}

// processParagraphWithFormatting processes a paragraph and applies markdown formatting to text nodes
func (tc *TableConverter) processParagraphWithFormatting(paragraph adf_types.ADFNode) string {
	var result strings.Builder

	for _, textNode := range paragraph.Content {
		if textNode.Type == "text" {
			text := textNode.Text

			for _, mark := range textNode.Marks {
				switch mark.Type {
				case "strong":
					text = "**" + text + "**"
				case "em":
					text = "*" + text + "*"
				case "link":
					if href, exists := mark.Attrs["href"]; exists {
						if hrefStr, ok := href.(string); ok {
							text = "[" + text + "](" + hrefStr + ")"
						}
					}
				}
			}

			result.WriteString(text)
		}
	}

	return result.String()
}

// wrapTableWithXML wraps markdown table content with XML tags and ADF attributes
//
//nolint:unparam // error return reserved for future use
func (tc *TableConverter) wrapTableWithXML(markdownTable string, attrs map[string]interface{}) (string, error) {
	var xmlBuilder strings.Builder

	xmlBuilder.WriteString("<table")

	if attrs != nil {
		if localId, exists := attrs["localId"]; exists {
			if localIdStr, ok := localId.(string); ok {
				xmlBuilder.WriteString(fmt.Sprintf(` localId="%s"`, localIdStr))
			}
		}

		if layout, exists := attrs["layout"]; exists {
			if layoutStr, ok := layout.(string); ok {
				xmlBuilder.WriteString(fmt.Sprintf(` layout="%s"`, layoutStr))
			}
		}

		if isNumberColumnEnabled, exists := attrs["isNumberColumnEnabled"]; exists {
			if isNumberColumnEnabledBool, ok := isNumberColumnEnabled.(bool); ok {
				xmlBuilder.WriteString(fmt.Sprintf(` isNumberColumnEnabled="%t"`, isNumberColumnEnabledBool))
			}
		}
	}

	xmlBuilder.WriteString(">\n")

	xmlBuilder.WriteString(markdownTable)

	if !strings.HasSuffix(markdownTable, "\n") {
		xmlBuilder.WriteString("\n")
	}
	xmlBuilder.WriteString("</table>")

	return xmlBuilder.String(), nil
}
