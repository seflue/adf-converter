package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/placeholder"
)

func TestDateConverter_ToMarkdown(t *testing.T) {
	dc := NewDateConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
		wantErr  bool
	}{
		{
			name: "basic date with unix millis",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeDate,
				Attrs: map[string]any{
					"timestamp": "1743724800000",
				},
			},
			expected: "[date:2025-04-04]",
		},
		{
			name: "date with nil attrs",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeDate,
			},
			wantErr: true,
		},
		{
			name: "date with empty timestamp",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeDate,
				Attrs: map[string]any{
					"timestamp": "",
				},
			},
			wantErr: true,
		},
		{
			name: "date with non-midnight timestamp",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeDate,
				Attrs: map[string]any{
					"timestamp": "1743771600000",
				},
			},
			expected: "[date:2025-04-04]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := converter.ConversionContext{Registry: newTestRegistry(), Strategy: converter.StandardMarkdown}
			result, err := dc.ToMarkdown(tt.node, ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestDateConverter_CanHandle(t *testing.T) {
	dc := NewDateConverter()

	if !dc.CanHandle("date") {
		t.Error("should handle date")
	}
	if dc.CanHandle("mention") {
		t.Error("should not handle mention")
	}
}

func TestDateConverter_GetStrategy(t *testing.T) {
	dc := NewDateConverter()

	if dc.GetStrategy() != converter.StandardMarkdown {
		t.Errorf("expected StandardMarkdown, got %v", dc.GetStrategy())
	}
}

func TestDateConverter_Roundtrip(t *testing.T) {

	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					adf_types.NewTextNode("Due: "),
					{
						Type: adf_types.NodeTypeDate,
						Attrs: map[string]any{
							"timestamp": "1743724800000",
						},
					},
					adf_types.NewTextNode(" end"),
				},
			},
		},
	}

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// ADF → Markdown
	registry := newTestRegistry()
	conv, err := converter.NewConverter(
		converter.WithClassifier(classifier),
		converter.WithPlaceholderManager(manager),
		converter.WithRegistry(registry),
	)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	markdown, session, err := conv.ToMarkdown(original)
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	expectedMD := "Due: [date:2025-04-04] end\n\n"
	if markdown != expectedMD {
		t.Errorf("expected markdown %q, got %q", expectedMD, markdown)
	}

	// Markdown → ADF
	result, err := conv.FromMarkdown(markdown, session)
	if err != nil {
		t.Fatalf("FromMarkdown failed: %v", err)
	}
	restored := result.Document

	// Verify date node roundtripped
	if len(restored.Content) != 1 {
		t.Fatalf("expected 1 block, got %d", len(restored.Content))
	}
	para := restored.Content[0]
	if para.Type != adf_types.NodeTypeParagraph {
		t.Fatalf("expected paragraph, got %s", para.Type)
	}

	var dateNode *adf_types.ADFNode
	for i, child := range para.Content {
		if child.Type == adf_types.NodeTypeDate {
			dateNode = &para.Content[i]
			break
		}
	}
	if dateNode == nil {
		t.Fatalf("no date node found in restored content: %+v", para.Content)
	}

	timestamp, ok := dateNode.Attrs["timestamp"].(string)
	if !ok {
		t.Fatal("date node missing timestamp")
	}
	if timestamp != "1743724800000" {
		t.Errorf("expected timestamp 1743724800000, got %s", timestamp)
	}
}
