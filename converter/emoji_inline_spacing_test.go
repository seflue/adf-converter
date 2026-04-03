package converter_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/converter/elements"
	"adf-converter/placeholder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmojiInlineSpacing tests that emoji nodes in inline contexts (within paragraphs/list items)
// are converted to unicode characters with inline spacing (no extra newlines).
//
// Emojis are now editable by users as unicode characters instead of preserved as placeholders:
//   - ADF emoji node → Unicode character in markdown (e.g., 👍, ✅)
//   - Users can edit emojis directly in markdown
//   - Unicode emojis detected and converted back to ADF emoji nodes on round-trip
func TestEmojiInlineSpacing(t *testing.T) {
	// Clear registry and register element converters
	converter.GetGlobalRegistry().Clear()
	converter.RegisterDefaultConverters(
		elements.NewTextConverter(),
		elements.NewHardBreakConverter(),
		elements.NewParagraphConverter(),
		elements.NewHeadingConverter(),
		elements.NewListItemConverter(),
		elements.NewBulletListConverter(),
		elements.NewOrderedListConverter(),
		elements.NewExpandConverter(),
		elements.NewInlineCardConverter(),
		elements.NewEmojiConverter(),
	)

	// Load test ADF document with emojis in various inline contexts
	adfBytes, err := os.ReadFile("../testdata/adf_samples/emoji_inline_spacing_test.json")
	require.NoError(t, err, "Failed to load test ADF file")

	var doc adf_types.ADFDocument
	err = json.Unmarshal(adfBytes, &doc)
	require.NoError(t, err, "Failed to parse ADF document")

	// Convert to markdown
	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := converter.ToMarkdown(doc, classifier, manager)
	require.NoError(t, err, "Failed to convert ADF to Markdown")
	require.NotNil(t, session, "Session should not be nil")

	// Load expected markdown output
	expectedBytes, err := os.ReadFile("../testdata/markdown_samples/emoji_inline_spacing_expected.md")
	require.NoError(t, err, "Failed to load expected markdown file")
	expected := string(expectedBytes)

	// Normalize both for comparison (handle different line endings)
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	expected = strings.ReplaceAll(expected, "\r\n", "\n")

	// Assert exact match
	assert.Equal(t, expected, markdown, "Markdown output should match expected output")

	// Additional structural assertions to ensure emojis are rendered inline

	// 1. Count unicode emojis - should be exactly 2
	thumbsUpCount := strings.Count(markdown, "👍")
	checkMarkCount := strings.Count(markdown, "✅")
	assert.Equal(t, 1, thumbsUpCount, "Should have exactly 1 thumbs up emoji")
	assert.Equal(t, 1, checkMarkCount, "Should have exactly 1 check mark emoji")

	// 2. Verify list structure is preserved (check for proper nesting)
	assert.Contains(t, markdown, "- Ut enim ad minim", "First list item should be present")
	assert.Contains(t, markdown, "  - `exercitation`", "Nested list item should have 2-space indent")
	assert.Contains(t, markdown, "- 👍 `mollit anim`",
		"Thumbs up emoji list item should be on single line with content")
	assert.Contains(t, markdown, "- ✅ `sed ut perspiciatis`",
		"Check mark emoji list item should be on single line with content")
}

// TestEmojiInlineSpacing_VerifyInlineVsBlockSpacing tests that emojis are rendered inline
// while block-level preserved nodes (like code blocks) still use placeholders.
func TestEmojiInlineSpacing_VerifyInlineVsBlockSpacing(t *testing.T) {
	// Clear registry and register element converters
	converter.GetGlobalRegistry().Clear()
	converter.RegisterDefaultConverters(
		elements.NewTextConverter(),
		elements.NewHardBreakConverter(),
		elements.NewParagraphConverter(),
		elements.NewHeadingConverter(),
		elements.NewListItemConverter(),
		elements.NewBulletListConverter(),
		elements.NewOrderedListConverter(),
		elements.NewExpandConverter(),
		elements.NewInlineCardConverter(),
		elements.NewEmojiConverter(),
	)

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Create a document with both inline emoji and block-level preserved content
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Text before emoji ",
					},
					{
						Type: adf_types.NodeTypeEmoji,
						Attrs: map[string]interface{}{
							"id":        "1f44d",
							"shortName": ":thumbsup:",
							"text":      "👍",
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " text after emoji",
					},
				},
			},
			{
				Type: adf_types.NodeTypeCodeBlock,
				Attrs: map[string]interface{}{
					"language": "go",
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "package main",
					},
				},
			},
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Text after code block",
					},
				},
			},
		},
	}

	markdown, _, err := converter.ToMarkdown(doc, classifier, manager)
	require.NoError(t, err, "Conversion should succeed")

	// Emoji should be rendered as unicode (inline spacing)
	// The paragraph should look like: "Text before emoji 👍 text after emoji\n\n"
	assert.Contains(t, markdown, "Text before emoji 👍 text after emoji",
		"Emoji should be inline with surrounding text as unicode")

	// Code block placeholder should have block spacing (double newline before next content)
	// Should see: "...\n\n<!-- CODE_BLOCK_PLACEHOLDER -->\n\nText after..."
	lines := strings.Split(markdown, "\n")
	var foundCodeBlockPlaceholder bool
	for i, line := range lines {
		if strings.Contains(line, "ADF_PLACEHOLDER_001") && strings.Contains(line, "Code Block") {
			foundCodeBlockPlaceholder = true

			// Check there's a blank line before (from previous paragraph's \n\n)
			if i > 0 {
				assert.Equal(t, "", lines[i-1], "Should have blank line before code block placeholder")
			}

			// Check there's a blank line after (from code block's \n\n)
			if i+1 < len(lines) {
				assert.Equal(t, "", lines[i+1], "Should have blank line after code block placeholder")
			}
		}
	}

	assert.True(t, foundCodeBlockPlaceholder, "Should have found code block placeholder")
}

// TestEmojiInlineSpacing_RoundTrip tests that documents with emoji produce correct markdown
// with unicode inline spacing (no placeholders, emojis are editable by users).
func TestEmojiInlineSpacing_RoundTrip(t *testing.T) {
	// Clear registry and register element converters
	converter.GetGlobalRegistry().Clear()
	converter.RegisterDefaultConverters(
		elements.NewTextConverter(),
		elements.NewHardBreakConverter(),
		elements.NewParagraphConverter(),
		elements.NewHeadingConverter(),
		elements.NewListItemConverter(),
		elements.NewBulletListConverter(),
		elements.NewOrderedListConverter(),
		elements.NewExpandConverter(),
		elements.NewInlineCardConverter(),
		elements.NewEmojiConverter(),
	)

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Create a simple document with emoji in a list
	originalDoc := adf_types.ADFDocument{
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
										Attrs: map[string]interface{}{
											"id":        "2705",
											"shortName": ":white_check_mark:",
											"text":      "✅",
										},
									},
									{
										Type: adf_types.NodeTypeText,
										Text: " Task completed",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Convert to Markdown
	markdown, session, err := converter.ToMarkdown(originalDoc, classifier, manager)
	require.NoError(t, err, "ADF to Markdown conversion should succeed")
	require.NotNil(t, session, "Session should not be nil")

	// Verify markdown structure - emoji should be unicode inline (no blank line)
	assert.Contains(t, markdown, "- ✅ Task completed",
		"Emoji should be inline with list item text as unicode")

	// Verify no blank lines are inserted after emoji
	lines := strings.Split(markdown, "\n")
	for i, line := range lines {
		if strings.Contains(line, "✅") && strings.Contains(line, "Task completed") {
			// Check no blank line follows the emoji line
			if i+1 < len(lines) && lines[i+1] == "" {
				// A blank line after would indicate the old bug
				if i+2 < len(lines) && lines[i+2] != "" {
					t.Error("Found blank line after emoji, this indicates the spacing bug")
				}
			}
			break
		}
	}
}
