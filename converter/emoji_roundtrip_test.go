package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/placeholder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmojiRoundTrip tests that emoji nodes survive a full ADF → Markdown → ADF round-trip
// Emojis are now converted to unicode characters in markdown (editable by users) and
// detected back during markdown → ADF conversion using gomoji library.
//
// Round-trip flow:
//
//	ADF with emoji → Markdown with unicode emoji → ADF conversion
//	Expected: Emoji node recreated from unicode with gomoji metadata
//	Result: Full round-trip fidelity with editable emojis
func TestEmojiRoundTrip(t *testing.T) {
	// Create original ADF document with emoji in list item (matches real Jira structure)
	originalADF := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeBulletList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeEmoji,
										Attrs: map[string]any{
											"id":        "2705",
											"shortName": ":white_check_mark:",
											"text":      "✅",
										},
									},
									{
										Type: adf_types.NodeTypeText,
										Text: " ",
									},
									{
										Type: adf_types.NodeTypeText,
										Text: "Task completed successfully",
										Marks: []adf_types.ADFMark{
											{Type: adf_types.MarkTypeCode},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Phase 1: ADF → Markdown
	markdown, session, err := converter.ToMarkdown(originalADF, classifier, manager)
	require.NoError(t, err, "ADF → Markdown conversion should succeed")
	require.NotNil(t, session, "Session should not be nil")

	t.Logf("Generated markdown:\n%s", markdown)

	// Verify markdown has unicode emoji (not placeholder)
	assert.Contains(t, markdown, "✅", "Markdown should contain unicode emoji")
	assert.Contains(t, markdown, "Task completed successfully", "Markdown should contain task text")

	// Phase 2: Markdown → ADF (the failing part)
	reconvertedADF, err := converter.FromMarkdown(markdown, session, manager)
	require.NoError(t, err, "Markdown → ADF conversion should succeed")

	t.Logf("Reconverted ADF: %+v", jsonPrettyPrint(reconvertedADF))

	// Validation: Structure should match
	require.Equal(t, "doc", reconvertedADF.Type, "Document type should be 'doc'")
	require.Len(t, reconvertedADF.Content, 1, "Should have 1 top-level node (bulletList)")

	bulletList := reconvertedADF.Content[0]
	require.Equal(t, adf_types.NodeTypeBulletList, bulletList.Type, "Should be bulletList")
	require.Len(t, bulletList.Content, 1, "BulletList should have 1 listItem")

	listItem := bulletList.Content[0]
	require.Equal(t, adf_types.NodeTypeListItem, listItem.Type, "Should be listItem")
	require.NotEmpty(t, listItem.Content, "ListItem must not be empty (CRITICAL: causes JIRA INVALID_INPUT)")
	require.Len(t, listItem.Content, 1, "ListItem should have 1 paragraph")

	para := listItem.Content[0]
	require.Equal(t, adf_types.NodeTypeParagraph, para.Type, "Should be paragraph")
	require.NotEmpty(t, para.Content, "Paragraph must have content")

	// Critical validation: Emoji should be restored
	require.GreaterOrEqual(t, len(para.Content), 2, "Paragraph should have emoji + text nodes")

	emojiNode := para.Content[0]
	assert.Equal(t, adf_types.NodeTypeEmoji, emojiNode.Type, "First node should be emoji (detected from unicode)")
	assert.NotNil(t, emojiNode.Attrs, "Emoji should have attributes")
	assert.Equal(t, "✅", emojiNode.Attrs["text"], "Emoji text should be preserved")
	assert.Equal(t, "2705", emojiNode.Attrs["id"], "Emoji ID should match unicode code point")
	// Note: shortName comes from gomoji, may differ from original Jira shortName
	assert.Contains(t, emojiNode.Attrs["shortName"], "check", "Emoji shortName should be related to check mark")

	// Verify text content is preserved
	hasText := false
	for _, node := range para.Content[1:] {
		if node.Type == adf_types.NodeTypeText && node.Text == "Task completed successfully" {
			hasText = true
			require.Len(t, node.Marks, 1, "Text should have code mark")
			assert.Equal(t, adf_types.MarkTypeCode, node.Marks[0].Type, "Should have code mark")
			break
		}
	}
	assert.True(t, hasText, "Text content should be preserved with formatting")
}

// TestEmojiRoundTrip_MultipleEmojis tests round-trip with multiple emoji nodes
func TestEmojiRoundTrip_MultipleEmojis(t *testing.T) {
	originalADF := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeBulletList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeEmoji,
										Attrs: map[string]any{
											"id":        "2705",
											"shortName": ":white_check_mark:",
											"text":      "✅",
										},
									},
									{
										Type: adf_types.NodeTypeText,
										Text: " First item",
									},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeEmoji,
										Attrs: map[string]any{
											"id":        "1f44d",
											"shortName": ":thumbsup:",
											"text":      "👍",
										},
									},
									{
										Type: adf_types.NodeTypeText,
										Text: " Second item",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Round-trip conversion
	markdown, session, err := converter.ToMarkdown(originalADF, classifier, manager)
	require.NoError(t, err)

	reconvertedADF, err := converter.FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Verify structure
	bulletList := reconvertedADF.Content[0]
	require.Len(t, bulletList.Content, 2, "Should have 2 list items")

	// Check first item
	listItem1 := bulletList.Content[0]
	require.NotEmpty(t, listItem1.Content, "First list item must not be empty")
	para1 := listItem1.Content[0]
	require.NotEmpty(t, para1.Content, "First paragraph must have content")
	assert.Equal(t, adf_types.NodeTypeEmoji, para1.Content[0].Type, "First item should have emoji")

	// Check second item
	listItem2 := bulletList.Content[1]
	require.NotEmpty(t, listItem2.Content, "Second list item must not be empty")
	para2 := listItem2.Content[0]
	require.NotEmpty(t, para2.Content, "Second paragraph must have content")
	assert.Equal(t, adf_types.NodeTypeEmoji, para2.Content[0].Type, "Second item should have emoji")
}

// Helper function for debugging
func jsonPrettyPrint(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}
