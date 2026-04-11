package converter

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayMode_NoPlaceholderComments(t *testing.T) {
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Visible paragraph"},
				},
			},
			{
				Type: adf_types.NodeTypeCodeBlock,
				Attrs: map[string]any{"language": "go"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "fmt.Println(\"hello\")"},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewNullManager()

	md, session, err := ToMarkdown(doc, classifier, manager)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Contains(t, md, "Visible paragraph")
	assert.NotContains(t, md, "<!--", "display mode must not contain placeholder comments")
	assert.NotContains(t, md, "ADF_PLACEHOLDER", "display mode must not contain placeholder IDs")
}

func TestDisplayMode_UnknownNodeShowsPreviewText(t *testing.T) {
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Before"},
				},
			},
			{
				Type: "unknownCustomNode",
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewNullManager()

	md, _, err := ToMarkdown(doc, classifier, manager)
	require.NoError(t, err)

	assert.Contains(t, md, "Before")
	assert.Contains(t, md, "complex content", "unknown node should show preview text in display mode")
	assert.NotContains(t, md, "<!--", "display mode must not produce placeholder comments")
}

func TestDisplayMode_InlinePreservedNodes(t *testing.T) {
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Hello "},
					{
						Type:  adf_types.NodeTypeStatus,
						Attrs: map[string]any{"text": "IN PROGRESS", "color": "blue"},
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewNullManager()

	md, _, err := ToMarkdown(doc, classifier, manager)
	require.NoError(t, err)

	assert.NotContains(t, md, "<!--", "inline preserved nodes must not produce comments")
}

func TestDisplayMode_MixedEditableAndPreserved(t *testing.T) {
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]any{"level": float64(1)},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Title"},
				},
			},
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Some text"},
				},
			},
			{
				Type: adf_types.NodeTypeBulletList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Item 1"},
								},
							},
						},
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewNullManager()

	md, _, err := ToMarkdown(doc, classifier, manager)
	require.NoError(t, err)

	assert.Contains(t, md, "# Title")
	assert.Contains(t, md, "Some text")
	assert.Contains(t, md, "Item 1")
	assert.NotContains(t, md, "<!--")

	// Verify no placeholder was stored
	assert.Equal(t, 0, manager.Count())
}

func TestNewDisplayConverter_Integration(t *testing.T) {
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Hello world"},
				},
			},
		},
	}

	conv := NewDisplayConverter()
	md, session, err := conv.ToMarkdown(doc)
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Contains(t, md, "Hello world")
	assert.False(t, strings.Contains(md, "<!--"), "display converter must not produce placeholder comments")
}
