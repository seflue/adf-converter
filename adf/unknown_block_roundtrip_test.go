package adf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

// TestUnknownBlockNodeRoundtrip pins the existing block-level fallback: any
// unknown block node type must survive ADF→Markdown→ADF via the placeholder
// mechanism. Complements TestUnknownInlineNodeRoundtrip; together they form
// the explicit roundtrip guarantee for unknown ADF node types (ac-0073).
func TestUnknownBlockNodeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		node adf.Node
	}{
		{
			name: "unknown block node with scalar attrs",
			node: adf.Node{
				Type: "futureBlockThing",
				Attrs: map[string]any{
					"key":   "value",
					"count": float64(3),
				},
			},
		},
		{
			name: "unknown block node without attrs",
			node: adf.Node{
				Type: "mysteryBlock",
			},
		},
		{
			name: "unknown block node with nested content",
			node: adf.Node{
				Type: "bodiedExtension2",
				Attrs: map[string]any{
					"extensionKey": "com.example.widget",
				},
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "nested body"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := adf.Document{
				Version: 1,
				Type:    "doc",
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Vor dem Block."},
						},
					},
					tt.node,
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Nach dem Block."},
						},
					},
				},
			}

			classifier := adf.NewDefaultClassifier()
			manager := placeholder.NewManager()

			markdown, session, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
			require.NoError(t, err)
			t.Logf("Markdown: %q", markdown)

			resultDoc, err := testFromMarkdown(markdown, session, manager, defaults.NewRegistry())
			require.NoError(t, err)

			require.Len(t, resultDoc.Content, 3, "expected 3 top-level nodes (para, unknown, para)")

			restored := resultDoc.Content[1]
			assert.Equal(t, tt.node.Type, restored.Type, "block type must survive roundtrip")
			assert.Equal(t, tt.node.Attrs, restored.Attrs, "attrs must be preserved")
			assert.Equal(t, tt.node.Content, restored.Content, "nested content must be preserved")
		})
	}
}
