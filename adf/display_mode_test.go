package adf_test

import (
	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"strings"
	"testing"

	"github.com/seflue/adf-converter/placeholder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayMode_NoPlaceholderComments(t *testing.T) {
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Visible paragraph"},
				},
			},
			{
				Type:  adf.NodeTypeCodeBlock,
				Attrs: map[string]any{"language": "go"},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "fmt.Println(\"hello\")"},
				},
			},
		},
	}

	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewNoop()

	md, session, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Contains(t, md, "Visible paragraph")
	assert.NotContains(t, md, "<!--", "display mode must not contain placeholder comments")
	assert.NotContains(t, md, "ADF_PLACEHOLDER", "display mode must not contain placeholder IDs")
}

func TestDisplayMode_UnknownNodeShowsPreviewText(t *testing.T) {
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Before"},
				},
			},
			{
				Type: "unknownCustomNode",
			},
		},
	}

	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewNoop()

	md, _, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err)

	assert.Contains(t, md, "Before")
	assert.Contains(t, md, "complex content", "unknown node should show preview text in display mode")
	assert.NotContains(t, md, "<!--", "display mode must not produce placeholder comments")
}

func TestDisplayMode_InlinePreservedNodes(t *testing.T) {
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Hello "},
					{
						Type:  adf.NodeTypeStatus,
						Attrs: map[string]any{"text": "IN PROGRESS", "color": "blue"},
					},
				},
			},
		},
	}

	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewNoop()

	md, _, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err)

	assert.NotContains(t, md, "<!--", "inline preserved nodes must not produce comments")
}

func TestDisplayMode_MixedEditableAndPreserved(t *testing.T) {
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type:  adf.NodeTypeHeading,
				Attrs: map[string]any{"level": float64(1)},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Title"},
				},
			},
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Some text"},
				},
			},
			{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Item 1"},
								},
							},
						},
					},
				},
			},
		},
	}

	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewNoop()

	md, _, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err)

	assert.Contains(t, md, "# Title")
	assert.Contains(t, md, "Some text")
	assert.Contains(t, md, "Item 1")
	assert.NotContains(t, md, "<!--")

	// Verify no placeholder was stored
	assert.Equal(t, 0, manager.Count())
}

func TestNewDisplayConverter_Integration(t *testing.T) {
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Hello world"},
				},
			},
		},
	}

	conv := defaults.NewDisplayConverter()
	md, session, err := conv.ToMarkdown(doc)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Contains(t, md, "Hello world")
	assert.False(t, strings.Contains(md, "<!--"), "display converter must not produce placeholder comments")
}
