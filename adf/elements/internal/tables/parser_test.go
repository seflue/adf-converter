package tables_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf/elements/internal/tables"
)

// ============================================================================
// Exploration: Goldmark Table AST Structure
// ============================================================================

// TestExplore_GoldmarkTableAST shows how goldmark's table extension represents
// tables in the AST. Run with -v to see the output.
func TestExplore_GoldmarkTableAST(t *testing.T) {
	cases := []struct {
		name     string
		markdown string
	}{
		{
			name: "plain table with header",
			markdown: "| Header 1 | Header 2 |\n" +
				"|----------|----------|\n" +
				"| Cell 1   | Cell 2   |\n",
		},
		{
			name: "bold cell content",
			markdown: "| **Bold** | normal |\n" +
				"|----------|--------|\n" +
				"| data     | more   |\n",
		},
		{
			name: "synthetic empty header (no ADF header)",
			markdown: "|  |  |\n" +
				"|--|--|\n" +
				"| A1 | B1 |\n",
		},
		{
			name: "no separator (invalid CommonMark — not a table)",
			markdown: "| Col 1 | Col 2 |\n" +
				"| Val 1 | Val 2 |\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(tc.markdown)
			gm := goldmark.New(goldmark.WithExtensions(extension.Table))
			doc := gm.Parser().Parse(text.NewReader(source))

			fmt.Printf("\n=== %s ===\n", tc.name)
			fmt.Printf("Markdown:\n%s\n", tc.markdown)
			fmt.Println("Goldmark AST:")
			dumpTableAST(doc, source, 0)
		})
	}
}

func dumpTableAST(node gast.Node, source []byte, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch n := node.(type) {
	case *gast.Document:
		fmt.Printf("%sDocument\n", indent)
	case *east.Table:
		fmt.Printf("%sTable (alignments=%v)\n", indent, n.Alignments)
	case *east.TableHeader:
		fmt.Printf("%sTableHeader\n", indent)
	case *east.TableRow:
		fmt.Printf("%sTableRow\n", indent)
	case *east.TableCell:
		fmt.Printf("%sTableCell (alignment=%v)\n", indent, n.Alignment)
		for i := 0; i < n.Lines().Len(); i++ {
			seg := n.Lines().At(i)
			fmt.Printf("%s  Lines[%d]: %q\n", indent, i, string(source[seg.Start:seg.Stop]))
		}
	case *gast.Paragraph:
		fmt.Printf("%sParagraph\n", indent)
	case *gast.Text:
		seg := n.Segment
		fmt.Printf("%sText: %q\n", indent, string(source[seg.Start:seg.Stop]))
	default:
		fmt.Printf("%s%T\n", indent, node)
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		dumpTableAST(child, source, depth+1)
	}
}

// ============================================================================
// ParseTable Tests
// ============================================================================

func TestParseTable_PlainTable(t *testing.T) {
	markdown := "| Header 1 | Header 2 |\n" +
		"|----------|----------|\n" +
		"| Cell 1   | Cell 2   |\n" +
		"| Cell 3   | Cell 4   |\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)

	assert.Equal(t, "table", node.Type)
	require.Len(t, node.Content, 3, "header row + 2 data rows")

	// First row: tableRow with tableHeader cells
	headerRow := node.Content[0]
	assert.Equal(t, "tableRow", headerRow.Type)
	require.Len(t, headerRow.Content, 2)
	assert.Equal(t, "tableHeader", headerRow.Content[0].Type)
	assert.Equal(t, "tableHeader", headerRow.Content[1].Type)

	// Header cell content
	headerCell1Para := headerRow.Content[0].Content[0]
	assert.Equal(t, "paragraph", headerCell1Para.Type)
	require.NotEmpty(t, headerCell1Para.Content)
	assert.Equal(t, "Header 1", headerCell1Para.Content[0].Text)

	// Data rows: tableRow with tableCell
	dataRow1 := node.Content[1]
	assert.Equal(t, "tableRow", dataRow1.Type)
	assert.Equal(t, "tableCell", dataRow1.Content[0].Type)
	assert.Equal(t, "Cell 1", dataRow1.Content[0].Content[0].Content[0].Text)
}

func TestParseTable_SyntheticEmptyHeader(t *testing.T) {
	// ADF tables with only tableCell nodes are serialized as | | | + separator
	markdown := "|  |  |\n" +
		"|--|--|\n" +
		"| A1 | B1 |\n" +
		"| A2 | B2 |\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)

	// Empty header row must be dropped — only 2 data rows remain
	require.Len(t, node.Content, 2, "synthetic header should be dropped")

	for i, row := range node.Content {
		for j, cell := range row.Content {
			assert.Equal(t, "tableCell", cell.Type,
				"row %d cell %d should be tableCell", i, j)
		}
	}

	// Content preserved
	assert.Equal(t, "A1", node.Content[0].Content[0].Content[0].Content[0].Text)
	assert.Equal(t, "B1", node.Content[0].Content[1].Content[0].Content[0].Text)
}

func TestParseTable_HeaderOnly(t *testing.T) {
	markdown := "| Header |\n" +
		"|--------|\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)

	require.Len(t, node.Content, 1)
	assert.Equal(t, "tableHeader", node.Content[0].Content[0].Type)
	assert.Equal(t, "Header", node.Content[0].Content[0].Content[0].Content[0].Text)
}

func TestParseTable_InlineFormattingInCells(t *testing.T) {
	markdown := "| **Bold** | *Italic* | `code` |\n" +
		"|----------|----------|--------|\n" +
		"| normal   | text     | data   |\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)
	require.Len(t, node.Content, 2)

	headerRow := node.Content[0]

	// **Bold** → strong mark
	cell0 := headerRow.Content[0].Content[0].Content[0] // paragraph → text node
	assert.Equal(t, "Bold", cell0.Text)
	require.Len(t, cell0.Marks, 1)
	assert.Equal(t, "strong", cell0.Marks[0].Type)

	// *Italic* → em mark
	cell1 := headerRow.Content[1].Content[0].Content[0]
	assert.Equal(t, "Italic", cell1.Text)
	require.Len(t, cell1.Marks, 1)
	assert.Equal(t, "em", cell1.Marks[0].Type)

	// `code` → code mark
	cell2 := headerRow.Content[2].Content[0].Content[0]
	assert.Equal(t, "code", cell2.Text)
	require.Len(t, cell2.Marks, 1)
	assert.Equal(t, "code", cell2.Marks[0].Type)
}

func TestParseTable_EmptyCells(t *testing.T) {
	markdown := "| Header 1 | Header 2 |\n" +
		"|----------|----------|\n" +
		"|          | Cell 2   |\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)
	require.Len(t, node.Content, 2)

	// Empty cell should have empty text node
	dataRow := node.Content[1]
	emptyCell := dataRow.Content[0]
	para := emptyCell.Content[0]
	require.NotEmpty(t, para.Content)
	assert.Equal(t, "", para.Content[0].Text)
}

func TestParseTable_EmptyInput(t *testing.T) {
	node, err := tables.ParseTable("")
	require.NoError(t, err)
	assert.Equal(t, "table", node.Type)
	assert.Empty(t, node.Content)
}

func TestParseTable_NoSeparatorNotATable(t *testing.T) {
	// Without separator row, Goldmark does NOT parse as a table.
	// This is correct CommonMark behavior — Jira never generates such output.
	markdown := "| Col 1 | Col 2 |\n" +
		"| Val 1 | Val 2 |\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)
	assert.Equal(t, "table", node.Type)
	assert.Empty(t, node.Content, "input without separator is not a valid table")
}

func TestParseTable_SingleColumn(t *testing.T) {
	markdown := "| Header |\n" +
		"|--------|\n" +
		"| Cell 1 |\n" +
		"| Cell 2 |\n"

	node, err := tables.ParseTable(markdown)
	require.NoError(t, err)
	require.Len(t, node.Content, 3)
	assert.Equal(t, "tableHeader", node.Content[0].Content[0].Type)
	assert.Equal(t, "tableCell", node.Content[1].Content[0].Type)
	assert.Equal(t, "tableCell", node.Content[2].Content[0].Type)
}
