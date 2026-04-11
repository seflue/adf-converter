package converter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/placeholder"
)

// TestUnknownBlockNodeRoundtrip pins the existing block-level fallback: any
// unknown block node type must survive ADF→Markdown→ADF via the placeholder
// mechanism. Complements TestUnknownInlineNodeRoundtrip; together they form
// the explicit roundtrip guarantee for unknown ADF node types (ac-0073).
func TestUnknownBlockNodeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		node adf_types.ADFNode
	}{
		{
			name: "unknown block node with scalar attrs",
			node: adf_types.ADFNode{
				Type: "futureBlockThing",
				Attrs: map[string]interface{}{
					"key":   "value",
					"count": float64(3),
				},
			},
		},
		{
			name: "unknown block node without attrs",
			node: adf_types.ADFNode{
				Type: "mysteryBlock",
			},
		},
		{
			name: "unknown block node with nested content",
			node: adf_types.ADFNode{
				Type: "bodiedExtension2",
				Attrs: map[string]interface{}{
					"extensionKey": "com.example.widget",
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "nested body"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Vor dem Block."},
						},
					},
					tt.node,
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Nach dem Block."},
						},
					},
				},
			}

			classifier := converter.NewDefaultClassifier()
			manager := placeholder.NewManager()

			markdown, session, err := converter.ToMarkdown(doc, classifier, manager)
			require.NoError(t, err)
			t.Logf("Markdown: %q", markdown)

			resultDoc, err := converter.FromMarkdown(markdown, session, manager)
			require.NoError(t, err)

			require.Len(t, resultDoc.Content, 3, "expected 3 top-level nodes (para, unknown, para)")

			restored := resultDoc.Content[1]
			assert.Equal(t, tt.node.Type, restored.Type, "block type must survive roundtrip")
			assert.Equal(t, tt.node.Attrs, restored.Attrs, "attrs must be preserved")
			assert.Equal(t, tt.node.Content, restored.Content, "nested content must be preserved")
		})
	}
}
