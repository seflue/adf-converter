package adf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

func TestMentionRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		node adf.Node
	}{
		{
			name: "basic mention",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":   "abc123",
					"text": "@john.doe",
				},
			},
		},
		{
			name: "mention with all attrs",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":          "user456",
					"text":        "@jane.smith",
					"accessLevel": "CONTAINER",
					"userType":    "DEFAULT",
				},
			},
		},
		{
			name: "mention with only accessLevel",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":          "user789",
					"text":        "@bob",
					"accessLevel": "APPLICATION",
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
						Type:    adf.NodeTypeParagraph,
						Content: []adf.Node{tt.node},
					},
				},
			}

			classifier := adf.NewDefaultClassifier()
			manager := placeholder.NewManager()

			// ADF → Markdown
			markdown, session, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
			require.NoError(t, err)
			t.Logf("Markdown: %q", markdown)

			// Markdown → ADF
			resultDoc, err := testFromMarkdown(markdown, session, manager, defaults.NewRegistry())
			require.NoError(t, err)

			// Verify roundtrip
			require.Len(t, resultDoc.Content, 1, "expected 1 paragraph")
			para := resultDoc.Content[0]
			require.Equal(t, adf.NodeTypeParagraph, para.Type)

			// Find mention node in paragraph content
			var mentionNode *adf.Node
			for i, child := range para.Content {
				if child.Type == adf.NodeTypeMention {
					mentionNode = &para.Content[i]
					break
				}
			}

			require.NotNil(t, mentionNode, "expected mention node in paragraph")
			assert.Equal(t, tt.node.Attrs["id"], mentionNode.Attrs["id"])
			assert.Equal(t, tt.node.Attrs["text"], mentionNode.Attrs["text"])

			if al, ok := tt.node.Attrs["accessLevel"]; ok {
				assert.Equal(t, al, mentionNode.Attrs["accessLevel"])
			}
			if ut, ok := tt.node.Attrs["userType"]; ok {
				assert.Equal(t, ut, mentionNode.Attrs["userType"])
			}
		})
	}
}

func TestUnresolvedMentionRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantID   string
		wantText string
	}{
		{
			name:     "unresolved mention with single word",
			markdown: "[@John]()",
			wantID:   "John",
			wantText: "@John",
		},
		{
			name:     "unresolved mention with spaces",
			markdown: "[@Some Name]()",
			wantID:   "Some Name",
			wantText: "@Some Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullMarkdown := tt.markdown + "\n"
			session := &placeholder.EditSession{}
			manager := placeholder.NewManager()

			resultDoc, err := testFromMarkdown(fullMarkdown, session, manager, defaults.NewRegistry())
			require.NoError(t, err)

			require.Len(t, resultDoc.Content, 1)
			para := resultDoc.Content[0]
			require.Equal(t, adf.NodeTypeParagraph, para.Type)

			var mentionNode *adf.Node
			for i, child := range para.Content {
				if child.Type == adf.NodeTypeMention {
					mentionNode = &para.Content[i]
					break
				}
			}
			require.NotNil(t, mentionNode, "expected mention node")
			assert.Equal(t, tt.wantID, mentionNode.Attrs["id"])
			assert.Equal(t, tt.wantText, mentionNode.Attrs["text"])
		})
	}
}

func TestMentionIDURLEncoding(t *testing.T) {
	// id with spaces roundtrips correctly via URL encoding
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeMention,
						Attrs: map[string]any{
							"id":   "Some Name",
							"text": "@Some Name",
						},
					},
				},
			},
		},
	}

	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err)
	assert.Contains(t, markdown, "[@Some Name](accountid:Some%20Name)")

	resultDoc, err := testFromMarkdown(markdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err)

	require.Len(t, resultDoc.Content, 1)
	para := resultDoc.Content[0]
	var mentionNode *adf.Node
	for i, child := range para.Content {
		if child.Type == adf.NodeTypeMention {
			mentionNode = &para.Content[i]
			break
		}
	}
	require.NotNil(t, mentionNode)
	assert.Equal(t, "Some Name", mentionNode.Attrs["id"])
	assert.Equal(t, "@Some Name", mentionNode.Attrs["text"])
}

func TestMentionInMixedParagraph(t *testing.T) {
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					adf.NewTextNode("Hello "),
					{
						Type: adf.NodeTypeMention,
						Attrs: map[string]any{
							"id":   "abc123",
							"text": "@john",
						},
					},
					adf.NewTextNode(" how are you?"),
				},
			},
		},
	}

	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// ADF → Markdown
	markdown, session, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err)

	assert.Contains(t, markdown, "Hello ")
	assert.Contains(t, markdown, "[@john](accountid:abc123)")
	assert.Contains(t, markdown, " how are you?")

	// Markdown → ADF
	resultDoc, err := testFromMarkdown(markdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err)

	require.Len(t, resultDoc.Content, 1)
	para := resultDoc.Content[0]

	mentionFound := false
	for _, child := range para.Content {
		if child.Type == adf.NodeTypeMention {
			mentionFound = true
			assert.Equal(t, "abc123", child.Attrs["id"])
			assert.Equal(t, "@john", child.Attrs["text"])
		}
	}
	assert.True(t, mentionFound, "mention node should survive roundtrip")
}
