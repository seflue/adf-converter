package elements

import (
	"strings"
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface check
var _ converter.ElementConverter = (*TableConverter)(nil)

// ============================================================================
// FromMarkdown Tests (new ElementConverter signature)
// ============================================================================

func TestTableConverter_FromMarkdown(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name            string
		lines           []string
		startIndex      int
		wantType        string
		wantRows        int
		wantConsumed    int
		wantHeaderCells int
		wantFirstCellType string
		wantErr         bool
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
	ctx := ConversionContext{}

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
	ctx := ConversionContext{PreserveAttrs: false}

	node := adf_types.ADFNode{
		Type: "table",
		Content: []adf_types.ADFNode{
			{
				Type: "tableRow",
				Content: []adf_types.ADFNode{
					{
						Type: "tableHeader",
						Content: []adf_types.ADFNode{
							{
								Type: "paragraph",
								Content: []adf_types.ADFNode{
									{Type: "text", Text: "Header 1"},
								},
							},
						},
					},
					{
						Type: "tableHeader",
						Content: []adf_types.ADFNode{
							{
								Type: "paragraph",
								Content: []adf_types.ADFNode{
									{Type: "text", Text: "Header 2"},
								},
							},
						},
					},
				},
			},
			{
				Type: "tableRow",
				Content: []adf_types.ADFNode{
					{
						Type: "tableCell",
						Content: []adf_types.ADFNode{
							{
								Type: "paragraph",
								Content: []adf_types.ADFNode{
									{Type: "text", Text: "Cell 1"},
								},
							},
						},
					},
					{
						Type: "tableCell",
						Content: []adf_types.ADFNode{
							{
								Type: "paragraph",
								Content: []adf_types.ADFNode{
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
	assert.Equal(t, MarkdownTable, result.Strategy)
	assert.Equal(t, 1, result.ElementsConverted)
	assert.Contains(t, result.Content, "Header 1")
	assert.Contains(t, result.Content, "Cell 1")
}

func TestTableConverter_ToMarkdown_WithAttributes(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{PreserveAttrs: true}

	node := adf_types.ADFNode{
		Type: "table",
		Attrs: map[string]interface{}{
			"localId": "abc123",
			"layout":  "wide",
		},
		Content: []adf_types.ADFNode{
			{
				Type: "tableRow",
				Content: []adf_types.ADFNode{
					{
						Type: "tableHeader",
						Content: []adf_types.ADFNode{
							{
								Type: "paragraph",
								Content: []adf_types.ADFNode{
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
	assert.Contains(t, result.Content, "localId=")
	assert.NotNil(t, result.PreservedAttrs)
}

// ============================================================================
// Roundtrip Tests (converter-level, not pipeline)
// ============================================================================

func TestTableConverter_RoundTrip_PlainTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{PreserveAttrs: false}

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
	ctx := ConversionContext{PreserveAttrs: true}

	lines := []string{
		`<table localId="abc123" layout="wide" isNumberColumnEnabled="true">`,
		"| Header 1 | Header 2 |",
		"|----------|----------|",
		"| Cell 1   | Cell 2   |",
		"</table>",
	}

	// MD → ADF
	adfNode, _, err := tc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	require.NotNil(t, adfNode.Attrs)
	assert.Equal(t, "abc123", adfNode.Attrs["localId"])

	// ADF → MD
	result, err := tc.ToMarkdown(adfNode, ctx)
	require.NoError(t, err)

	// MD → ADF again
	lines2 := strings.Split(strings.TrimSpace(result.Content), "\n")
	adfNode2, _, err := tc.FromMarkdown(lines2, 0, ctx)
	require.NoError(t, err)

	require.NotNil(t, adfNode2.Attrs)
	assert.Equal(t, "abc123", adfNode2.Attrs["localId"])
	assert.Equal(t, "wide", adfNode2.Attrs["layout"])
}

// ============================================================================
// Inline Content Tests
// ============================================================================

func TestTableConverter_FromMarkdown_InlineFormatting(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

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
	ctx := ConversionContext{}

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
			name: "table without separator row",
			lines: []string{
				"| Col 1 | Col 2 |",
				"| Val 1 | Val 2 |",
			},
			wantRows:     2,
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
	ctx := ConversionContext{}

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
	assert.True(t, tc.CanHandle(NodeTable))
	assert.False(t, tc.CanHandle("paragraph"))
}

func TestTableConverter_GetStrategy(t *testing.T) {
	tc := NewTableConverter()
	assert.Equal(t, MarkdownTable, tc.GetStrategy())
}

func TestTableConverter_ValidateInput(t *testing.T) {
	tc := NewTableConverter()

	t.Run("valid ADF node", func(t *testing.T) {
		assert.NoError(t, tc.ValidateInput(adf_types.ADFNode{Type: "table"}))
	})

	t.Run("wrong node type", func(t *testing.T) {
		assert.Error(t, tc.ValidateInput(adf_types.ADFNode{Type: "paragraph"}))
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
