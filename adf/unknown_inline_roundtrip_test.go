package adf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

// TestUnknownInlineNodeRoundtrip verifies that inline nodes with unknown types
// survive ADF→Markdown→ADF via the placeholder mechanism, the same way unknown
// block nodes already do. Regression guard for ac-0073.
func TestUnknownInlineNodeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		node adf.Node
	}{
		{
			name: "unknown inline node with scalar attrs",
			node: adf.Node{
				Type: "mention2",
				Attrs: map[string]any{
					"id":          "u-42",
					"accessLevel": "ADMIN",
				},
			},
		},
		{
			name: "unknown inline node without attrs",
			node: adf.Node{
				Type: "futureInlineThing",
			},
		},
		{
			name: "unknown inline node with nested attrs",
			node: adf.Node{
				Type: "inlineExtension2",
				Attrs: map[string]any{
					"extensionKey":  "com.example.widget",
					"extensionType": "com.example",
					"parameters": map[string]any{
						"foo": "bar",
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
							{Type: adf.NodeTypeText, Text: "Hallo "},
							tt.node,
							{Type: adf.NodeTypeText, Text: ", hier ist der Link."},
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

			require.Len(t, resultDoc.Content, 1, "expected 1 paragraph")
			para := resultDoc.Content[0]
			require.Equal(t, adf.NodeTypeParagraph, para.Type)

			var restored *adf.Node
			for i, child := range para.Content {
				if child.Type == tt.node.Type {
					restored = &para.Content[i]
					break
				}
			}

			require.NotNil(t, restored, "expected %s node to survive roundtrip", tt.node.Type)
			assert.Equal(t, tt.node.Type, restored.Type)
			assert.Equal(t, tt.node.Attrs, restored.Attrs, "attrs must be preserved")
		})
	}
}
