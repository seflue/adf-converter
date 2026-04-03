package elements

import (
	"testing"

	"adf-converter/adf_types"
)

func TestTableConverter_FromMarkdown_PlainTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	markdown := `| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |`

	result, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	if result.Type != "table" {
		t.Errorf("Expected type 'table', got '%s'", result.Type)
	}

	if len(result.Content) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(result.Content))
	}

	// Check header row
	headerRow := result.Content[0]
	if headerRow.Type != "tableRow" {
		t.Errorf("Expected first row type 'tableRow', got '%s'", headerRow.Type)
	}

	if len(headerRow.Content) != 3 {
		t.Errorf("Expected 3 header cells, got %d", len(headerRow.Content))
	}

	if headerRow.Content[0].Type != "tableHeader" {
		t.Errorf("Expected first cell type 'tableHeader', got '%s'", headerRow.Content[0].Type)
	}

	// Check data rows
	dataRow := result.Content[1]
	if dataRow.Content[0].Type != "tableCell" {
		t.Errorf("Expected data cell type 'tableCell', got '%s'", dataRow.Content[0].Type)
	}
}

func TestTableConverter_FromMarkdown_XMLWrappedTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	markdown := `<table localId="abc123" layout="wide">
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
</table>`

	result, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	if result.Type != "table" {
		t.Errorf("Expected type 'table', got '%s'", result.Type)
	}

	// Check attributes
	if result.Attrs == nil {
		t.Fatal("Expected attributes to be present")
	}

	if localId, ok := result.Attrs["localId"].(string); !ok || localId != "abc123" {
		t.Errorf("Expected localId='abc123', got %v", result.Attrs["localId"])
	}

	if layout, ok := result.Attrs["layout"].(string); !ok || layout != "wide" {
		t.Errorf("Expected layout='wide', got %v", result.Attrs["layout"])
	}

	// Check content
	if len(result.Content) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(result.Content))
	}
}

func TestTableConverter_FromMarkdown_XMLWrappedTableWithBooleanAttr(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	markdown := `<table isNumberColumnEnabled="true">
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
</table>`

	result, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	// Check boolean attribute
	if result.Attrs == nil {
		t.Fatal("Expected attributes to be present")
	}

	if enabled, ok := result.Attrs["isNumberColumnEnabled"].(bool); !ok || !enabled {
		t.Errorf("Expected isNumberColumnEnabled=true, got %v", result.Attrs["isNumberColumnEnabled"])
	}
}

func TestTableConverter_FromMarkdown_EmptyTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	markdown := ""

	result, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	if result.Type != "table" {
		t.Errorf("Expected type 'table', got '%s'", result.Type)
	}

	if len(result.Content) != 0 {
		t.Errorf("Expected 0 rows for empty table, got %d", len(result.Content))
	}
}

func TestTableConverter_FromMarkdown_MinimalTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	markdown := `| Header |
|--------|`

	result, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	if len(result.Content) != 1 {
		t.Errorf("Expected 1 row (header only), got %d", len(result.Content))
	}
}

func TestTableConverter_FromMarkdown_MalformedXMLTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{}

	// Missing closing tag
	markdown := `<table localId="abc123">
| Header 1 | Header 2 |
|----------|----------|`

	_, err := tc.FromMarkdown(markdown, ctx)
	if err == nil {
		t.Error("Expected error for malformed XML table (missing closing tag)")
	}
}

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
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	if result.Strategy != MarkdownTable {
		t.Errorf("Expected strategy MarkdownTable, got %v", result.Strategy)
	}

	if result.ElementsConverted != 1 {
		t.Errorf("Expected 1 element converted, got %d", result.ElementsConverted)
	}
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
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	// Should contain XML wrapper
	if !stringContains(result.Content, "<table") {
		t.Error("Expected XML wrapper in output")
	}

	if !stringContains(result.Content, "localId=") {
		t.Error("Expected localId attribute in output")
	}

	if result.PreservedAttrs == nil {
		t.Error("Expected preserved attributes")
	}
}

func TestTableConverter_RoundTrip_PlainTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	markdown := `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |`

	// Convert to ADF
	adfNode, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	// Convert back to markdown
	result, err := tc.ToMarkdown(adfNode, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	// Convert to ADF again
	adfNode2, err := tc.FromMarkdown(result.Content, ctx)
	if err != nil {
		t.Fatalf("Second FromMarkdown failed: %v", err)
	}

	// Check that the structure is preserved
	if len(adfNode.Content) != len(adfNode2.Content) {
		t.Errorf("Row count mismatch after round-trip: %d vs %d", len(adfNode.Content), len(adfNode2.Content))
	}
}

func TestTableConverter_RoundTrip_XMLWrappedTable(t *testing.T) {
	tc := NewTableConverter()
	ctx := ConversionContext{PreserveAttrs: true}

	markdown := `<table localId="abc123" layout="wide" isNumberColumnEnabled="true">
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
</table>`

	// Convert to ADF
	adfNode, err := tc.FromMarkdown(markdown, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}

	// Verify attributes
	if adfNode.Attrs == nil {
		t.Fatal("Expected attributes after FromMarkdown")
	}

	if localId, ok := adfNode.Attrs["localId"].(string); !ok || localId != "abc123" {
		t.Errorf("Expected localId='abc123' after FromMarkdown, got %v", adfNode.Attrs["localId"])
	}

	// Convert back to markdown
	result, err := tc.ToMarkdown(adfNode, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	// Convert to ADF again
	adfNode2, err := tc.FromMarkdown(result.Content, ctx)
	if err != nil {
		t.Fatalf("Second FromMarkdown failed: %v", err)
	}

	// Check that attributes are preserved
	if adfNode2.Attrs == nil {
		t.Fatal("Expected attributes after round-trip")
	}

	if localId, ok := adfNode2.Attrs["localId"].(string); !ok || localId != "abc123" {
		t.Errorf("Expected localId='abc123' after round-trip, got %v", adfNode2.Attrs["localId"])
	}

	if layout, ok := adfNode2.Attrs["layout"].(string); !ok || layout != "wide" {
		t.Errorf("Expected layout='wide' after round-trip, got %v", adfNode2.Attrs["layout"])
	}
}

// Helper function to check if string contains substring
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
