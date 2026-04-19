// Package tables provides goldmark-based table parsing for ADF conversion.
package tables

import (
	"strings"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/elements/internal/inline"
)

// ParseTable parses a markdown table string into an ADF table node.
// Uses goldmark's Table extension for CommonMark-compliant parsing.
//
// Handles:
//   - Tables with header rows (tableHeader cells in ADF)
//   - Tables with synthetic empty header rows (all-tableCell ADF tables serialised
//     as "| | |" to satisfy CommonMark separator requirement — ac-0037)
//   - Inline formatting in cells via inline.ParseContent
//
// Input without a separator row is not a valid CommonMark table; goldmark
// returns no Table node and ParseTable returns an empty table node.
func ParseTable(markdown string) (adf_types.ADFNode, error) {
	empty := adf_types.ADFNode{Type: "table", Content: []adf_types.ADFNode{}}
	if strings.TrimSpace(markdown) == "" {
		return empty, nil
	}

	source := []byte(markdown)
	gm := goldmark.New(goldmark.WithExtensions(extension.Table))
	doc := gm.Parser().Parse(text.NewReader(source))

	tableAST := findTable(doc)
	if tableAST == nil {
		return empty, nil
	}

	return convertTable(tableAST, source), nil
}

// findTable returns the first Table node in the document, or nil.
func findTable(doc gast.Node) *east.Table {
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*east.Table); ok {
			return t
		}
	}
	return nil
}

// convertTable walks the goldmark Table AST and builds an ADF table node.
func convertTable(table *east.Table, source []byte) adf_types.ADFNode {
	var rows []adf_types.ADFNode

	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *east.TableHeader:
			// Synthetic empty header: ADF tables with no tableHeader cells are
			// serialised as an all-empty first row. Drop it and emit all
			// subsequent rows as tableCell (ac-0037).
			if allCellsEmpty(n, source) {
				break
			}
			rows = append(rows, convertRowNode(n, source, "tableHeader"))
		case *east.TableRow:
			rows = append(rows, convertRowNode(n, source, "tableCell"))
		}
	}

	return adf_types.ADFNode{Type: "table", Content: rows}
}

// allCellsEmpty reports whether every TableCell child has empty content.
func allCellsEmpty(row gast.Node, source []byte) bool {
	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if tc, ok := cell.(*east.TableCell); ok {
			if extractCellText(tc, source) != "" {
				return false
			}
		}
	}
	return true
}

// convertRowNode converts a TableHeader or TableRow node into an ADF tableRow.
// cellType is "tableHeader" or "tableCell".
func convertRowNode(row gast.Node, source []byte, cellType string) adf_types.ADFNode {
	var cells []adf_types.ADFNode
	for child := row.FirstChild(); child != nil; child = child.NextSibling() {
		tc, ok := child.(*east.TableCell)
		if !ok {
			continue
		}
		cellText := extractCellText(tc, source)
		paragraphContent := parseCellContent(cellText)
		cells = append(cells, adf_types.ADFNode{
			Type: cellType,
			Content: []adf_types.ADFNode{
				{Type: "paragraph", Content: paragraphContent},
			},
		})
	}
	return adf_types.ADFNode{Type: "tableRow", Content: cells}
}

// extractCellText returns the raw markdown content of a table cell.
// Goldmark stores each cell's trimmed source bytes in cell.Lines(), so
// source[seg.Start:seg.Stop] gives us the original markdown (e.g. "**bold**")
// which we pass directly to inline.ParseContent.
func extractCellText(cell *east.TableCell, source []byte) string {
	lines := cell.Lines()
	if lines.Len() == 0 {
		return ""
	}
	seg := lines.At(0)
	if seg.Start == seg.Stop {
		return ""
	}
	return string(source[seg.Start:seg.Stop])
}

// parseCellContent parses raw cell markdown into ADF inline nodes.
// Falls back to a plain text node on error.
func parseCellContent(cellText string) []adf_types.ADFNode {
	if cellText == "" {
		return []adf_types.ADFNode{{Type: "text", Text: ""}}
	}
	nodes, err := inline.ParseContent(cellText)
	if err != nil || len(nodes) == 0 {
		return []adf_types.ADFNode{{Type: "text", Text: cellText}}
	}
	return nodes
}
