package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface check
var _ adf.Renderer = (*tableConverter)(nil)

// ============================================================================
// FromMarkdown Tests (new Renderer signature)
// ============================================================================

func TestTableConverter_FromMarkdown(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name              string
		lines             []string
		startIndex        int
		wantType          string
		wantRows          int
		wantConsumed      int
		wantHeaderCells   int
		wantFirstCellType string
		wantErr           bool
	}{
		{
			name: "plain table",
			lines: []string{
				"| Header 1 | Header 2 | Header 3 |",
				"|----------|----------|----------|",
				"| Cell 1   | Cell 2   | Cell 3   |",
				"| Cell 4   | Cell 5   | Cell 6   |",
			},
			wantType:          "table",
			wantRows:          3,
			wantConsumed:      4,
			wantHeaderCells:   3,
			wantFirstCellType: "tableHeader",
		},
		{
			name: "minimal table with header only",
			lines: []string{
				"| Header |",
				"|--------|",
			},
			wantType:          "table",
			wantRows:          1,
			wantConsumed:      2,
			wantHeaderCells:   1,
			wantFirstCellType: "tableHeader",
		},
		{
			name: "XML-wrapped table",
			lines: []string{
				`<table localId="abc123" layout="wide">`,
				"| Header 1 | Header 2 |",
				"|----------|----------|",
				"| Cell 1   | Cell 2   |",
				"</table>",
			},
			wantType:          "table",
			wantRows:          2,
			wantConsumed:      5,
			wantHeaderCells:   2,
			wantFirstCellType: "tableHeader",
		},
		{
			name: "XML-wrapped table with boolean attr",
			lines: []string{
				`<table isNumberColumnEnabled="true">`,
				"| Header 1 | Header 2 |",
				"|----------|----------|",
				"| Cell 1   | Cell 2   |",
				"</table>",
			},
			wantType:     "table",
			wantRows:     2,
			wantConsumed: 5,
		},
		{
			name: "startIndex skips preceding lines",
			lines: []string{
				"Some paragraph text",
				"| Header 1 | Header 2 |",
				"|----------|----------|",
				"| Cell 1   | Cell 2   |",
			},
			startIndex:        1,
			wantType:          "table",
			wantRows:          2,
			wantConsumed:      3,
			wantHeaderCells:   2,
			wantFirstCellType: "tableHeader",
		},
		{
			name: "table followed by other content",
			lines: []string{
				"| Header 1 | Header 2 |",
				"|----------|----------|",
				"| Cell 1   | Cell 2   |",
				"",
				"Another paragraph",
			},
			wantType:     "table",
			wantRows:     2,
			wantConsumed: 3,
		},
		{
			name:         "empty lines produces empty table",
			lines:        []string{},
			wantType:     "table",
			wantRows:     0,
			wantConsumed: 0,
		},
		{
			name: "malformed XML table missing closing tag",
			lines: []string{
				`<table localId="abc123">`,
				"| Header 1 | Header 2 |",
				"|----------|----------|",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, consumed, err := tc.FromMarkdown(tt.lines, tt.startIndex, ctx)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, result.Type)
			assert.Equal(t, tt.wantConsumed, consumed)
			assert.Len(t, result.Content, tt.wantRows)

			if tt.wantHeaderCells > 0 && tt.wantRows > 0 {
				headerRow := result.Content[0]
				assert.Equal(t, "tableRow", headerRow.Type)
				assert.Len(t, headerRow.Content, tt.wantHeaderCells)
				assert.Equal(t, tt.wantFirstCellType, headerRow.Content[0].Type)
			}

			// Data rows should use tableCell
			if tt.wantRows > 1 && tt.wantFirstCellType == "tableHeader" {
				dataRow := result.Content[1]
				assert.Equal(t, "tableCell", dataRow.Content[0].Type)
			}
		})
	}
}

func TestTableConverter_FromMarkdown_XMLWrappedAttributes(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	t.Run("localId and layout preserved", func(t *testing.T) {
		lines := []string{
			`<table localId="abc123" layout="wide">`,
			"| Header 1 | Header 2 |",
			"|----------|----------|",
			"| Cell 1   | Cell 2   |",
			"</table>",
		}

		result, _, err := tc.FromMarkdown(lines, 0, ctx)
		require.NoError(t, err)
		require.NotNil(t, result.Attrs)

		assert.Equal(t, "abc123", result.Attrs["localId"])
		assert.Equal(t, "wide", result.Attrs["layout"])
	})

	t.Run("boolean attribute preserved", func(t *testing.T) {
		lines := []string{
			`<table isNumberColumnEnabled="true">`,
			"| Header 1 | Header 2 |",
			"|----------|----------|",
			"| Cell 1   | Cell 2   |",
			"</table>",
		}

		result, _, err := tc.FromMarkdown(lines, 0, ctx)
		require.NoError(t, err)
		require.NotNil(t, result.Attrs)

		enabled, ok := result.Attrs["isNumberColumnEnabled"].(bool)
		assert.True(t, ok, "isNumberColumnEnabled should be bool")
		assert.True(t, enabled)
	})
}

// ============================================================================
// ToMarkdown Tests
// ============================================================================

func TestTableConverter_ToMarkdown_PlainTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false}

	node := adf.Node{
		Type: "table",
		Content: []adf.Node{
			{
				Type: "tableRow",
				Content: []adf.Node{
					{
						Type: "tableHeader",
						Content: []adf.Node{
							{
								Type: "paragraph",
								Content: []adf.Node{
									{Type: "text", Text: "Header 1"},
								},
							},
						},
					},
					{
						Type: "tableHeader",
						Content: []adf.Node{
							{
								Type: "paragraph",
								Content: []adf.Node{
									{Type: "text", Text: "Header 2"},
								},
							},
						},
					},
				},
			},
			{
				Type: "tableRow",
				Content: []adf.Node{
					{
						Type: "tableCell",
						Content: []adf.Node{
							{
								Type: "paragraph",
								Content: []adf.Node{
									{Type: "text", Text: "Cell 1"},
								},
							},
						},
					},
					{
						Type: "tableCell",
						Content: []adf.Node{
							{
								Type: "paragraph",
								Content: []adf.Node{
									{Type: "text", Text: "Cell 2"},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	require.NoError(t, err)
	assert.Equal(t, adf.MarkdownTable, result.Strategy)
	assert.Equal(t, 1, result.ElementsConverted)
	assert.Contains(t, result.Content, "Header 1")
	assert.Contains(t, result.Content, "Cell 1")
	assert.True(t, strings.HasSuffix(result.Content, "\n\n"),
		"table output must end with double newline for block spacing")
}

func TestTableConverter_ToMarkdown_NoHeader(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false}

	// ADF table where ALL cells are tableCell (no tableHeader)
	node := adf.Node{
		Type: "table",
		Content: []adf.Node{
			{
				Type: "tableRow",
				Content: []adf.Node{
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "A1"}}},
					}},
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "B1"}}},
					}},
				},
			},
			{
				Type: "tableRow",
				Content: []adf.Node{
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "A2"}}},
					}},
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "B2"}}},
					}},
				},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	require.Len(t, lines, 4, "should have empty header + separator + 2 data rows")

	// First line: empty synthetic header
	assert.Equal(t, "|  |  |", lines[0])
	// Second line: separator
	assert.Contains(t, lines[1], "--")
	// Data rows preserve content
	assert.Contains(t, lines[2], "A1")
	assert.Contains(t, lines[2], "B1")
	assert.Contains(t, lines[3], "A2")
	assert.Contains(t, lines[3], "B2")
}

func TestTableConverter_ToMarkdown_WithAttributes(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true}

	node := adf.Node{
		Type: "table",
		Attrs: map[string]any{
			"layout": "wide",
		},
		Content: []adf.Node{
			{
				Type: "tableRow",
				Content: []adf.Node{
					{
						Type: "tableHeader",
						Content: []adf.Node{
							{
								Type: "paragraph",
								Content: []adf.Node{
									{Type: "text", Text: "Header 1"},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "<table")
	assert.Contains(t, result.Content, `layout="wide"`)
	assert.NotContains(t, result.Content, "localId", "localId is filtered as default")
	assert.True(t, strings.HasSuffix(result.Content, "\n\n"),
		"table output must end with double newline for block spacing")
}

func TestTableConverter_ToMarkdown_DefaultAttrsOmitWrapper(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true}

	tableContent := []adf.Node{
		{
			Type: "tableRow",
			Content: []adf.Node{
				{Type: "tableHeader", Content: []adf.Node{
					{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "H1"}}},
				}},
			},
		},
	}

	tests := []struct {
		name        string
		attrs       map[string]any
		wantWrapper bool
	}{
		{
			name:        "only defaults produces no wrapper",
			attrs:       map[string]any{"isNumberColumnEnabled": false, "layout": "center"},
			wantWrapper: false,
		},
		{
			name:        "layout default produces no wrapper",
			attrs:       map[string]any{"layout": "default"},
			wantWrapper: false,
		},
		{
			name:        "only default displayMode produces no wrapper",
			attrs:       map[string]any{"displayMode": "default"},
			wantWrapper: false,
		},
		{
			name:        "localId alone produces no wrapper",
			attrs:       map[string]any{"localId": "abc123", "isNumberColumnEnabled": false},
			wantWrapper: false,
		},
		{
			name:        "non-default layout triggers wrapper",
			attrs:       map[string]any{"layout": "align-start"},
			wantWrapper: true,
		},
		{
			name:        "isNumberColumnEnabled true triggers wrapper",
			attrs:       map[string]any{"isNumberColumnEnabled": true},
			wantWrapper: true,
		},
		{
			name:        "width always triggers wrapper",
			attrs:       map[string]any{"width": 960},
			wantWrapper: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := adf.Node{
				Type:    "table",
				Attrs:   tt.attrs,
				Content: tableContent,
			}
			result, err := tc.ToMarkdown(node, ctx)
			require.NoError(t, err)

			hasWrapper := strings.Contains(result.Content, "<table")
			assert.Equal(t, tt.wantWrapper, hasWrapper,
				"wrapper presence mismatch for attrs %v", tt.attrs)
		})
	}
}

// ============================================================================
// Roundtrip Tests (converter-level, not pipeline)
// ============================================================================

func TestTableConverter_RoundTrip_PlainTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false}

	lines := []string{
		"| Header 1 | Header 2 |",
		"|----------|----------|",
		"| Cell 1   | Cell 2   |",
		"| Cell 3   | Cell 4   |",
	}

	// MD → ADF
	adfNode, consumed, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)

	// ADF → MD
	result, err := tc.ToMarkdown(adfNode, ctx)
	require.NoError(t, err)

	// MD → ADF again
	lines2 := strings.Split(strings.TrimSpace(result.Content), "\n")
	adfNode2, _, err := tc.FromMarkdown(lines2, 0, ctx)
	require.NoError(t, err)

	assert.Equal(t, len(adfNode.Content), len(adfNode2.Content),
		"row count should be preserved after roundtrip")
}

func TestTableConverter_RoundTrip_XMLWrappedTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true}

	lines := []string{
		`<table layout="wide" isNumberColumnEnabled="true">`,
		"| Header 1 | Header 2 |",
		"|----------|----------|",
		"| Cell 1   | Cell 2   |",
		"</table>",
	}

	// MD → ADF
	adfNode, _, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	require.NotNil(t, adfNode.Attrs)
	assert.Equal(t, "wide", adfNode.Attrs["layout"])

	// ADF → MD
	result, err := tc.ToMarkdown(adfNode, ctx)
	require.NoError(t, err)

	// MD → ADF again
	lines2 := strings.Split(strings.TrimSpace(result.Content), "\n")
	adfNode2, _, err := tc.FromMarkdown(lines2, 0, ctx)
	require.NoError(t, err)

	require.NotNil(t, adfNode2.Attrs)
	assert.Equal(t, "wide", adfNode2.Attrs["layout"])
	enabled, ok := adfNode2.Attrs["isNumberColumnEnabled"].(bool)
	assert.True(t, ok)
	assert.True(t, enabled)
}

func TestTableConverter_FromMarkdown_EmptyHeaderMeansNoHeader(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	lines := []string{
		"|  |  |",
		"|--|--|",
		"| A1 | B1 |",
		"| A2 | B2 |",
	}

	result, consumed, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)

	// All rows should be tableCell, no tableHeader
	require.Len(t, result.Content, 2, "empty header row should be dropped")
	for i, row := range result.Content {
		for j, cell := range row.Content {
			assert.Equal(t, "tableCell", cell.Type,
				"row %d cell %d should be tableCell, not tableHeader", i, j)
		}
	}
}

func TestTableConverter_RoundTrip_NoHeader(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false}

	// Start with ADF: all tableCell, no tableHeader
	originalNode := adf.Node{
		Type: "table",
		Content: []adf.Node{
			{
				Type: "tableRow",
				Content: []adf.Node{
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "A1"}}},
					}},
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "B1"}}},
					}},
				},
			},
			{
				Type: "tableRow",
				Content: []adf.Node{
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "A2"}}},
					}},
					{Type: "tableCell", Content: []adf.Node{
						{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "B2"}}},
					}},
				},
			},
		},
	}

	// ADF → MD
	mdResult, err := tc.ToMarkdown(originalNode, ctx)
	require.NoError(t, err)

	// MD → ADF
	lines := strings.Split(strings.TrimSpace(mdResult.Content), "\n")
	roundtrippedNode, _, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)

	// Same number of rows
	assert.Equal(t, len(originalNode.Content), len(roundtrippedNode.Content),
		"row count must survive roundtrip")

	// All cells remain tableCell (not promoted to tableHeader)
	for i, row := range roundtrippedNode.Content {
		for j, cell := range row.Content {
			assert.Equal(t, "tableCell", cell.Type,
				"row %d cell %d should be tableCell after roundtrip", i, j)
		}
	}

	// Content preserved
	cell00 := roundtrippedNode.Content[0].Content[0].Content[0].Content[0]
	assert.Equal(t, "A1", cell00.Text)
}

func TestTableConverter_ToMarkdown_InlineMarks(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false}

	tests := []struct {
		name     string
		marks    []adf.Mark
		text     string
		wantCell string
	}{
		{
			name:     "textColor",
			marks:    []adf.Mark{{Type: "textColor", Attrs: map[string]any{"color": "#ff5630"}}},
			text:     "red",
			wantCell: `<span style="color: #ff5630">red</span>`,
		},
		{
			name:     "strikethrough",
			marks:    []adf.Mark{{Type: "strike"}},
			text:     "deleted",
			wantCell: "~~deleted~~",
		},
		{
			name:     "subscript",
			marks:    []adf.Mark{{Type: "subsup", Attrs: map[string]any{"type": "sub"}}},
			text:     "2",
			wantCell: "<sub>2</sub>",
		},
		{
			name: "bold with textColor",
			marks: []adf.Mark{
				{Type: "strong"},
				{Type: "textColor", Attrs: map[string]any{"color": "#36b37e"}},
			},
			text:     "green bold",
			wantCell: `**<span style="color: #36b37e">green bold</span>**`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := adf.Node{
				Type: "table",
				Content: []adf.Node{
					{Type: "tableRow", Content: []adf.Node{
						{Type: "tableHeader", Content: []adf.Node{
							{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "Header"}}},
						}},
					}},
					{Type: "tableRow", Content: []adf.Node{
						{Type: "tableCell", Content: []adf.Node{
							{Type: "paragraph", Content: []adf.Node{
								{Type: "text", Text: tt.text, Marks: tt.marks},
							}},
						}},
					}},
				},
			}

			result, err := tc.ToMarkdown(node, ctx)
			require.NoError(t, err)
			assert.Contains(t, result.Content, tt.wantCell,
				"cell should contain formatted text")
		})
	}
}

// ============================================================================
// Inline Content Tests
// ============================================================================

func TestTableConverter_FromMarkdown_InlineFormatting(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	lines := []string{
		"| **Bold** | *Italic* | `code` | [link](http://example.com) |",
		"|----------|----------|--------|----------------------------|",
		"| normal   | text     | more   | data                       |",
	}

	result, _, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	require.Len(t, result.Content, 2, "should have header + data row")

	headerRow := result.Content[0]
	require.Len(t, headerRow.Content, 4, "header row should have 4 cells")

	// Cell 1: **Bold** → text "Bold" with strong mark
	cell1Para := headerRow.Content[0].Content[0]
	require.NotEmpty(t, cell1Para.Content)
	assert.Equal(t, "Bold", cell1Para.Content[0].Text)
	require.Len(t, cell1Para.Content[0].Marks, 1)
	assert.Equal(t, "strong", cell1Para.Content[0].Marks[0].Type)

	// Cell 2: *Italic* → text "Italic" with em mark
	cell2Para := headerRow.Content[1].Content[0]
	require.NotEmpty(t, cell2Para.Content)
	assert.Equal(t, "Italic", cell2Para.Content[0].Text)
	require.Len(t, cell2Para.Content[0].Marks, 1)
	assert.Equal(t, "em", cell2Para.Content[0].Marks[0].Type)

	// Cell 3: `code` → text "code" with code mark
	cell3Para := headerRow.Content[2].Content[0]
	require.NotEmpty(t, cell3Para.Content)
	assert.Equal(t, "code", cell3Para.Content[0].Text)
	require.Len(t, cell3Para.Content[0].Marks, 1)
	assert.Equal(t, "code", cell3Para.Content[0].Marks[0].Type)

	// Cell 4: [link](url) → text "link" with link mark
	cell4Para := headerRow.Content[3].Content[0]
	require.NotEmpty(t, cell4Para.Content)
	assert.Equal(t, "link", cell4Para.Content[0].Text)
	require.Len(t, cell4Para.Content[0].Marks, 1)
	assert.Equal(t, "link", cell4Para.Content[0].Marks[0].Type)
	assert.Equal(t, "http://example.com", cell4Para.Content[0].Marks[0].Attrs["href"])
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestTableConverter_FromMarkdown_EdgeCases(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name         string
		lines        []string
		wantRows     int
		wantConsumed int
		wantErr      bool
	}{
		{
			name: "empty cells",
			lines: []string{
				"| Header 1 | Header 2 |",
				"|----------|----------|",
				"|          | Cell 2   |",
				"| Cell 3   |          |",
			},
			wantRows:     3,
			wantConsumed: 4,
		},
		{
			name: "single column",
			lines: []string{
				"| Header |",
				"|--------|",
				"| Cell 1 |",
				"| Cell 2 |",
			},
			wantRows:     3,
			wantConsumed: 4,
		},
		{
			// Goldmark requires a separator row for CommonMark compliance.
			// Input without separator is not a valid table — 0 rows returned.
			// Jira never generates such output, so this is not a real regression.
			// consumed=2 because countPlainTableLines still counts both | lines.
			name: "table without separator row",
			lines: []string{
				"| Col 1 | Col 2 |",
				"| Val 1 | Val 2 |",
			},
			wantRows:     0,
			wantConsumed: 2,
		},
		{
			name: "table with trailing empty line stops consumption",
			lines: []string{
				"| H1 | H2 |",
				"|----|-----|",
				"| C1 | C2 |",
				"",
				"Next paragraph",
			},
			wantRows:     2,
			wantConsumed: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, consumed, err := tc.FromMarkdown(tt.lines, 0, ctx)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "table", result.Type)
			assert.Equal(t, tt.wantConsumed, consumed)
			assert.Len(t, result.Content, tt.wantRows)
		})
	}
}

func TestTableConverter_FromMarkdown_EmptyCellContent(t *testing.T) {
	tc := NewTableConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	lines := []string{
		"| Header 1 | Header 2 |",
		"|----------|----------|",
		"|          | Cell 2   |",
	}

	result, _, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)

	// Data row, first cell should have empty text
	dataRow := result.Content[1]
	firstCell := dataRow.Content[0]
	para := firstCell.Content[0]
	require.NotEmpty(t, para.Content)
	assert.Equal(t, "", para.Content[0].Text, "empty cell should have empty text node")
}

// ============================================================================
// Helper & Validation Tests
// ============================================================================

func TestTableConverter_CanHandle(t *testing.T) {
	tc := NewTableConverter()
	assert.True(t, tc.CanHandle(adf.NodeTable))
	assert.False(t, tc.CanHandle("paragraph"))
}

func TestTableConverter_GetStrategy(t *testing.T) {
	tc := NewTableConverter()
	assert.Equal(t, adf.MarkdownTable, tc.GetStrategy())
}

func TestTableConverter_ValidateInput(t *testing.T) {
	tc := NewTableConverter()

	t.Run("valid ADF node", func(t *testing.T) {
		assert.NoError(t, tc.ValidateInput(adf.Node{Type: "table"}))
	})

	t.Run("wrong node type", func(t *testing.T) {
		assert.Error(t, tc.ValidateInput(adf.Node{Type: "paragraph"}))
	})

	t.Run("valid string", func(t *testing.T) {
		assert.NoError(t, tc.ValidateInput("| H |"))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Error(t, tc.ValidateInput(""))
	})

	t.Run("nil input", func(t *testing.T) {
		assert.Error(t, tc.ValidateInput(nil))
	})
}
