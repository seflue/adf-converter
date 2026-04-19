package converter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

// TestUnknownInlineNodeRoundtrip verifies that inline nodes with unknown types
// survive ADF→Markdown→ADF via the placeholder mechanism, the same way unknown
// block nodes already do. Regression guard for ac-0073.
func TestUnknownInlineNodeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		node adf_types.ADFNode
	}{
		{
			name: "unknown inline node with scalar attrs",
			node: adf_types.ADFNode{
				Type: "mention2",
				Attrs: map[string]any{
					"id":          "u-42",
					"accessLevel": "ADMIN",
				},
			},
		},
		{
			name: "unknown inline node without attrs",
			node: adf_types.ADFNode{
				Type: "futureInlineThing",
			},
		},
		{
			name: "unknown inline node with nested attrs",
			node: adf_types.ADFNode{
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
			doc := adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Hallo "},
							tt.node,
							{Type: adf_types.NodeTypeText, Text: ", hier ist der Link."},
						},
					},
				},
			}

			classifier := converter.NewDefaultClassifier()
			manager := placeholder.NewManager()

			markdown, session, err := converter.ToMarkdown(doc, classifier, manager, defaults.NewRegistry())
			require.NoError(t, err)
			t.Logf("Markdown: %q", markdown)

			resultDoc, err := converter.FromMarkdown(markdown, session, manager, defaults.NewRegistry())
			require.NoError(t, err)

			require.Len(t, resultDoc.Content, 1, "expected 1 paragraph")
			para := resultDoc.Content[0]
			require.Equal(t, adf_types.NodeTypeParagraph, para.Type)

			var restored *adf_types.ADFNode
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
