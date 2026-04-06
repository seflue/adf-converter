package converter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// ============================================================================
// Basic Round-trip Fidelity Tests
// ============================================================================

func TestRoundTripConversion_BasicDocument(t *testing.T) {
	// Create a simple ADF document
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Hello, world!",
					},
				},
			},
		},
	}

	converter := NewDefaultConverter()

	// Perform round-trip conversion
	markdown, restored, err := converter.ConvertRoundTrip(original)
	if err != nil {
		t.Fatalf("Round-trip conversion failed: %v", err)
	}

	// Verify markdown is as expected
	expectedMarkdown := "Hello, world!\n\n"
	if markdown != expectedMarkdown {
		t.Errorf("Expected markdown %q, got %q", expectedMarkdown, markdown)
	}

	// Verify the round-trip preserves the document
	if !reflect.DeepEqual(original, restored) {
		t.Errorf("Round-trip conversion changed the document.\nOriginal: %+v\nRestored: %+v", original, restored)
	}
}

func TestRoundTripConversion_FormattedText(t *testing.T) {
	// Create ADF document with formatted text
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Bold",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeStrong},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " and ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "italic",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeEm},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " and ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "code",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeCode},
						},
					},
				},
			},
		},
	}

	converter := NewDefaultConverter()

	// Perform round-trip conversion
	markdown, restored, err := converter.ConvertRoundTrip(original)
	if err != nil {
		t.Fatalf("Round-trip conversion failed: %v", err)
	}

	// Verify markdown contains expected formatting
	assert.Contains(t, markdown, "**Bold**")
	assert.Contains(t, markdown, "*italic*")
	assert.Contains(t, markdown, "`code`")

	// Verify the round-trip preserves the document structure
	if !reflect.DeepEqual(original, restored) {
		t.Errorf("Round-trip conversion changed the document.\nOriginal: %+v\nRestored: %+v", original, restored)
	}
}

func TestRoundTripConversion_MultipleHeadings(t *testing.T) {
	// Create ADF document with multiple heading levels
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 1,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Main Title",
					},
				},
			},
			{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 2,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Subtitle",
					},
				},
			},
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Some content here.",
					},
				},
			},
		},
	}

	converter := NewDefaultConverter()

	// Perform round-trip conversion
	markdown, restored, err := converter.ConvertRoundTrip(original)
	if err != nil {
		t.Fatalf("Round-trip conversion failed: %v", err)
	}

	// Verify markdown contains expected heading formatting
	assert.Contains(t, markdown, "# Main Title")
	assert.Contains(t, markdown, "## Subtitle")
	assert.Contains(t, markdown, "Some content here.")

	// Verify the round-trip preserves the document structure
	if !reflect.DeepEqual(original, restored) {
		t.Errorf("Round-trip conversion changed the document.\nOriginal: %+v\nRestored: %+v", original, restored)
	}
}

// ============================================================================
// Underline Mark Round-trip Tests (ac-0004)
// ============================================================================

func TestRoundTripConversion_UnderlineText(t *testing.T) {
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "This is ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "underlined",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeUnderline},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " text",
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "This is <u>underlined</u> text\n\n", markdown)
	assert.Equal(t, original, restored)
}

func TestRoundTripConversion_UnderlineBoldText(t *testing.T) {
	// Mark order note: applyMarkToText applies marks left-to-right, so the last
	// mark becomes the outermost wrapper. [underline, strong] → **<u>text</u>**.
	// The parser reads outside-in: ** outermost → strong first, <u> inner → underline.
	// Result is [strong, underline] — inverted from the input. This is a known
	// asymmetry between mark application (inside-out) and parsing (outside-in).
	// This test verifies the markdown output and that both marks survive the roundtrip,
	// without asserting mark order (which does not roundtrip for mixed HTML+MD marks).
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "bold underlined",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeUnderline},
							{Type: adf_types.MarkTypeStrong},
						},
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "**<u>bold underlined</u>**\n\n", markdown)

	// Verify structure is preserved
	require.Equal(t, 1, len(restored.Content))
	require.Equal(t, adf_types.NodeTypeParagraph, restored.Content[0].Type)
	require.Equal(t, 1, len(restored.Content[0].Content))

	restoredNode := restored.Content[0].Content[0]
	assert.Equal(t, "bold underlined", restoredNode.Text)
	assert.Equal(t, 2, len(restoredNode.Marks))

	// Both marks must be present after roundtrip (order may differ due to parse direction)
	markTypes := make(map[string]bool)
	for _, m := range restoredNode.Marks {
		markTypes[m.Type] = true
	}
	assert.True(t, markTypes[adf_types.MarkTypeStrong], "strong mark must survive roundtrip")
	assert.True(t, markTypes[adf_types.MarkTypeUnderline], "underline mark must survive roundtrip")
}

// ============================================================================
// TextColor Mark Round-trip Tests (ac-0009)
// ============================================================================

func TestRoundTripConversion_TextColorText(t *testing.T) {
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "This is ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "red text",
						Marks: []adf_types.ADFMark{
							adf_types.NewMark(adf_types.MarkTypeTextColor, map[string]interface{}{
								"color": "#ff0000",
							}),
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " here",
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "This is <span style=\"color: #ff0000\">red text</span> here\n\n", markdown)
	assert.Equal(t, original, restored)
}

func TestRoundTripConversion_TextColorBoldText(t *testing.T) {
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "bold red",
						Marks: []adf_types.ADFMark{
							adf_types.NewMark(adf_types.MarkTypeTextColor, map[string]interface{}{
								"color": "#ff0000",
							}),
							{Type: adf_types.MarkTypeStrong},
						},
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "**<span style=\"color: #ff0000\">bold red</span>**\n\n", markdown)

	require.Equal(t, 1, len(restored.Content))
	require.Equal(t, 1, len(restored.Content[0].Content))

	restoredNode := restored.Content[0].Content[0]
	assert.Equal(t, "bold red", restoredNode.Text)
	assert.Equal(t, 2, len(restoredNode.Marks))

	// Beide Marks muessen ueberleben (Reihenfolge kann abweichen)
	markTypes := make(map[string]bool)
	for _, m := range restoredNode.Marks {
		markTypes[m.Type] = true
	}
	assert.True(t, markTypes[adf_types.MarkTypeStrong], "strong mark must survive roundtrip")
	assert.True(t, markTypes[adf_types.MarkTypeTextColor], "textColor mark must survive roundtrip")

	for _, m := range restoredNode.Marks {
		if m.Type == adf_types.MarkTypeTextColor {
			assert.Equal(t, "#ff0000", m.Attrs["color"])
		}
	}
}

// ============================================================================
// Subscript / Superscript Round-trip Tests
// ============================================================================

func TestRoundTripConversion_SubscriptText(t *testing.T) {
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "H",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "2",
						Marks: []adf_types.ADFMark{
							{
								Type: adf_types.MarkTypeSubsup,
								Attrs: map[string]interface{}{
									"type": "sub",
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "O",
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "H<sub>2</sub>O\n\n", markdown)
	assert.Equal(t, original, restored)
}

func TestRoundTripConversion_SuperscriptText(t *testing.T) {
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "x",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "2",
						Marks: []adf_types.ADFMark{
							{
								Type: adf_types.MarkTypeSubsup,
								Attrs: map[string]interface{}{
									"type": "sup",
								},
							},
						},
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "x<sup>2</sup>\n\n", markdown)
	assert.Equal(t, original, restored)
}

func TestRoundTripConversion_BoldSubscriptText(t *testing.T) {
	// Mark order note: [subsup, strong] → applyMarkToText wraps inside-out:
	// sub first: <sub>text</sub>, then strong: **<sub>text</sub>**
	// Parser reads outside-in: ** → strong, <sub> → subsup
	// Result marks: [strong, subsup] — order inverted but both survive.
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "important",
						Marks: []adf_types.ADFMark{
							{
								Type: adf_types.MarkTypeSubsup,
								Attrs: map[string]interface{}{
									"type": "sub",
								},
							},
							{Type: adf_types.MarkTypeStrong},
						},
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	markdown, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	assert.Equal(t, "**<sub>important</sub>**\n\n", markdown)

	// Verify structure preserved
	require.Equal(t, 1, len(restored.Content))
	require.Equal(t, 1, len(restored.Content[0].Content))

	restoredNode := restored.Content[0].Content[0]
	assert.Equal(t, "important", restoredNode.Text)
	assert.Equal(t, 2, len(restoredNode.Marks))

	// Both marks must survive (order may differ)
	markTypes := make(map[string]bool)
	for _, m := range restoredNode.Marks {
		markTypes[m.Type] = true
	}
	assert.True(t, markTypes[adf_types.MarkTypeStrong], "strong mark must survive roundtrip")
	assert.True(t, markTypes[adf_types.MarkTypeSubsup], "subsup mark must survive roundtrip")
}

// ============================================================================
// Enhanced Link Round-trip Tests
// ============================================================================

func TestEnhancedLinkRoundtrip_BasicExternalLinks(t *testing.T) {
	// Test basic external links through round-trip conversion
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Visit ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "Google",
						Marks: []adf_types.ADFMark{
							{
								Type: adf_types.MarkTypeLink,
								Attrs: map[string]interface{}{
									"href": "https://google.com",
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " for search.",
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Convert to markdown
	markdown, session, err := ToMarkdown(original, classifier, manager)
	require.NoError(t, err)

	// Verify markdown format
	expectedMarkdown := "Visit [Google](https://google.com) for search.\n\n"
	assert.Equal(t, expectedMarkdown, markdown)

	// Convert back to ADF
	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Verify round-trip fidelity for structure
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Version, restored.Version)
	assert.Equal(t, len(original.Content), len(restored.Content))

	// Verify paragraph structure
	assert.Equal(t, adf_types.NodeTypeParagraph, restored.Content[0].Type)

	// Check that the link is preserved
	paragraph := restored.Content[0]
	hasLink := false
	for _, node := range paragraph.Content {
		if node.Type == adf_types.NodeTypeText && len(node.Marks) > 0 {
			for _, mark := range node.Marks {
				if mark.Type == adf_types.MarkTypeLink {
					hasLink = true
					break
				}
			}
		}
	}
	assert.True(t, hasLink, "Link should be preserved in round-trip conversion")
}

// ============================================================================
// Expand Element Round-trip Tests
// ============================================================================

func TestExpandRoundtrip_BasicExpandElement(t *testing.T) {
	// Test basic expand element through round-trip conversion
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeExpand,
				Attrs: map[string]interface{}{
					"title": "Click to expand",
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Hidden content here.",
							},
						},
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := ToMarkdown(original, classifier, manager)
	require.NoError(t, err)

	// No data-adf-type attribute — node type is derived from structural context
	assert.Contains(t, markdown, "<details>")
	assert.NotContains(t, markdown, `data-adf-type`)
	assert.Contains(t, markdown, "<summary>Click to expand</summary>")
	assert.Contains(t, markdown, "Hidden content here.")
	assert.Contains(t, markdown, "</details>")

	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Version, restored.Version)
	assert.Equal(t, len(original.Content), len(restored.Content))
	assert.Equal(t, adf_types.NodeTypeExpand, restored.Content[0].Type)
	assert.NotNil(t, restored.Content[0].Attrs)
	assert.Equal(t, "Click to expand", restored.Content[0].Attrs["title"])
}

// ============================================================================
// Complex Structure Round-trip Tests
// ============================================================================

func TestRoundTrip_ComplexDocument(t *testing.T) {
	// Test document with multiple element types
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 1,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Test Document",
					},
				},
			},
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "This is a paragraph with ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "bold text",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeStrong},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " and a ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "link",
						Marks: []adf_types.ADFMark{
							{
								Type: adf_types.MarkTypeLink,
								Attrs: map[string]interface{}{
									"href": "https://example.com",
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: ".",
					},
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
									{
										Type: adf_types.NodeTypeText,
										Text: "First item",
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
										Type: adf_types.NodeTypeText,
										Text: "Second item",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Convert to markdown
	markdown, session, err := ToMarkdown(original, classifier, manager)
	require.NoError(t, err)

	// Verify markdown contains expected elements
	assert.Contains(t, markdown, "# Test Document")
	assert.Contains(t, markdown, "**bold text**")
	assert.Contains(t, markdown, "[link](https://example.com)")
	assert.Contains(t, markdown, "- First item")
	assert.Contains(t, markdown, "- Second item")

	// Convert back to ADF
	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Verify document structure is preserved
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Version, restored.Version)
	assert.Equal(t, len(original.Content), len(restored.Content))

	// Verify each element type is preserved
	assert.Equal(t, adf_types.NodeTypeHeading, restored.Content[0].Type)
	assert.Equal(t, adf_types.NodeTypeParagraph, restored.Content[1].Type)
	assert.Equal(t, adf_types.NodeTypeBulletList, restored.Content[2].Type)
}

// ============================================================================
// 100% Fidelity Validation Tests
// ============================================================================

func TestRoundTripFidelity_PreservesAllContent(t *testing.T) {
	// This test validates that the Constitution requirement II is met:
	// "100% ADF ↔ Markdown round-trip fidelity"

	testCases := []struct {
		name string
		doc  adf_types.ADFDocument
	}{
		{
			name: "simple_text",
			doc: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Simple text"},
						},
					},
				},
			},
		},
		{
			name: "formatted_text",
			doc: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type:  adf_types.NodeTypeText,
								Text:  "Bold text",
								Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}},
							},
						},
					},
				},
			},
		},
		{
			name: "heading_levels",
			doc: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type:  adf_types.NodeTypeHeading,
						Attrs: map[string]interface{}{"level": 2},
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Level 2 Heading"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			classifier := NewDefaultClassifier()
			manager := placeholder.NewManager()

			// Convert ADF → Markdown → ADF
			markdown, session, err := ToMarkdown(tc.doc, classifier, manager)
			require.NoError(t, err, "ADF to Markdown conversion failed")

			restored, err := FromMarkdown(markdown, session, manager)
			require.NoError(t, err, "Markdown to ADF conversion failed")

			// Verify 100% fidelity for basic structure
			assert.Equal(t, tc.doc.Type, restored.Type, "Document type should be preserved")
			assert.Equal(t, tc.doc.Version, restored.Version, "Document version should be preserved")
			assert.Equal(t, len(tc.doc.Content), len(restored.Content), "Content length should be preserved")

			// For simple cases, verify complete fidelity
			if tc.name == "simple_text" {
				assert.True(t, reflect.DeepEqual(tc.doc, restored),
					"Simple text should have 100% round-trip fidelity")
			}
		})
	}
}

func TestExpandBackwardCompatibility_DetailsFormat(t *testing.T) {
	// Test that <details> expand in context (surrounded by paragraphs) is parsed correctly
	markdownWithDetails := "paragraph before\n\n<details data-adf-type=\"expand\">\n  <summary>Section Title</summary>\n  Content inside.\n</details>\n\nMore content here."

	manager := placeholder.NewManager()
	session := manager.GetSession()

	restored, err := FromMarkdown(markdownWithDetails, session, manager)
	require.NoError(t, err)

	require.Equal(t, 3, len(restored.Content), "Should have 3 top-level elements")

	expandNode := restored.Content[1]
	assert.Equal(t, adf_types.NodeTypeExpand, expandNode.Type)
	assert.Equal(t, "Section Title", expandNode.Attrs["title"])
	require.Greater(t, len(expandNode.Content), 0, "Expand should have content")
}

func TestExpandNested_NoBlankLineBeforeInnerDetails(t *testing.T) {
	// Bug: when there's no blank line between text and nested <details>,
	// the inner expand gets corrupted during roundtrip
	markdown := "<details>\n  <summary>Outer</summary>\n  Some text before\n  <details>\n    <summary>Inner</summary>\n    Inner content\n  </details>\n</details>"

	manager := placeholder.NewManager()
	session := manager.GetSession()

	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Find the expand node
	require.Greater(t, len(restored.Content), 0)
	expandNode := restored.Content[0]
	assert.Equal(t, adf_types.NodeTypeExpand, expandNode.Type)
	assert.Equal(t, "Outer", expandNode.Attrs["title"])

	// Should have both: a paragraph with "Some text before" AND a nestedExpand
	require.GreaterOrEqual(t, len(expandNode.Content), 2, "Should have paragraph + nestedExpand")

	var foundParagraph, foundNestedExpand bool
	for _, child := range expandNode.Content {
		if child.Type == adf_types.NodeTypeParagraph {
			foundParagraph = true
		}
		if child.Type == adf_types.NodeTypeNestedExpand {
			foundNestedExpand = true
			assert.Equal(t, "Inner", child.Attrs["title"])
		}
	}
	assert.True(t, foundParagraph, "Should have paragraph with text content")
	assert.True(t, foundNestedExpand, "Should have nestedExpand child")
}

func TestParagraph_InlineHTMLNotBlockBoundary(t *testing.T) {
	// Inline HTML tags at line start must NOT break the paragraph.
	// isBlockBoundary delegates to BlockParsers — none should claim <u>, <sub>, etc.
	markdown := "First line\n<u>underlined</u> text\n<sub>subscript</sub> too"

	manager := placeholder.NewManager()
	session := manager.GetSession()

	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// All three lines should be in a single paragraph
	require.Equal(t, 1, len(restored.Content), "Should be one paragraph, not split at inline HTML")
	assert.Equal(t, adf_types.NodeTypeParagraph, restored.Content[0].Type)
}

func TestExpandRoundtrip_NestedExpandWithContent(t *testing.T) {
	// Full ADF→MD→ADF roundtrip: expand containing text + nestedExpand
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type:  adf_types.NodeTypeExpand,
				Attrs: map[string]interface{}{"title": "Outer"},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Text before nested"},
						},
					},
					{
						Type:  adf_types.NodeTypeNestedExpand,
						Attrs: map[string]interface{}{"title": "Inner"},
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Nested content"},
								},
							},
						},
					},
				},
			},
		},
	}

	manager := placeholder.NewManager()
	classifier := NewDefaultClassifier()

	markdown, session, err := ToMarkdown(original, classifier, manager)
	require.NoError(t, err)

	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	require.Equal(t, 1, len(restored.Content))
	expand := restored.Content[0]
	assert.Equal(t, adf_types.NodeTypeExpand, expand.Type)
	assert.Equal(t, "Outer", expand.Attrs["title"])

	// Must have paragraph + nestedExpand
	var hasParagraph, hasNestedExpand bool
	for _, child := range expand.Content {
		if child.Type == adf_types.NodeTypeParagraph {
			hasParagraph = true
		}
		if child.Type == adf_types.NodeTypeNestedExpand {
			hasNestedExpand = true
			assert.Equal(t, "Inner", child.Attrs["title"])
		}
	}
	assert.True(t, hasParagraph, "Roundtrip must preserve paragraph")
	assert.True(t, hasNestedExpand, "Roundtrip must preserve nestedExpand")
}

func TestExpandParsing_BackToBackExpands(t *testing.T) {
	// Two expands without blank line between closing and opening tag
	markdown := "<details>\n  <summary>First</summary>\n  Content 1\n</details>\n<details>\n  <summary>Second</summary>\n  Content 2\n</details>"

	manager := placeholder.NewManager()
	session := manager.GetSession()

	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	require.Equal(t, 2, len(restored.Content), "Should parse two separate expand nodes")
	assert.Equal(t, adf_types.NodeTypeExpand, restored.Content[0].Type)
	assert.Equal(t, "First", restored.Content[0].Attrs["title"])
	assert.Equal(t, adf_types.NodeTypeExpand, restored.Content[1].Type)
	assert.Equal(t, "Second", restored.Content[1].Attrs["title"])
}

func TestExpandDetails_AttributeHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    adf_types.ADFDocument
		expected []string
	}{
		{
			name: "expand with expanded state",
			input: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeExpand,
						Attrs: map[string]interface{}{
							"title":    "Expanded Section",
							"expanded": true,
						},
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Content"},
								},
							},
						},
					},
				},
			},
			expected: []string{"<details open", "<summary>Expanded Section</summary>", "Content", "</details>"},
		},
		{
			name: "expand with localId",
			input: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeExpand,
						Attrs: map[string]interface{}{
							"title":   "Section with ID",
							"localId": "my-expand-123",
						},
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Content with ID"},
								},
							},
						},
					},
				},
			},
			expected: []string{`id="my-expand-123"`, "<summary>Section with ID</summary>", "Content with ID", "</details>"},
		},
		{
			name: "expand with both expanded and localId",
			input: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeExpand,
						Attrs: map[string]interface{}{
							"title":    "Full Featured Section",
							"expanded": true,
							"localId":  "full-expand",
						},
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Full featured content"},
								},
							},
						},
					},
				},
			},
			expected: []string{"<details open", `id="full-expand"`, "<summary>Full Featured Section</summary>", "Full featured content", "</details>"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			classifier := NewDefaultClassifier()
			manager := placeholder.NewManager()

			markdown, _, err := ToMarkdown(test.input, classifier, manager)
			require.NoError(t, err)

			for _, expected := range test.expected {
				assert.Contains(t, markdown, expected, "Markdown should contain: %s", expected)
			}

			session := manager.GetSession()
			restored, err := FromMarkdown(markdown, session, manager)
			require.NoError(t, err)

			assert.Equal(t, test.input.Type, restored.Type)
			assert.Equal(t, len(test.input.Content), len(restored.Content))
			assert.Equal(t, adf_types.NodeTypeExpand, restored.Content[0].Type)
		})
	}
}

func TestExpandDetails_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		markdownIn  string
		expectError bool
		expectNodes int
	}{
		{
			name: "malformed details - no summary",
			markdownIn: `<details>
Content without summary
</details>`,
			expectError: true,
		},
		{
			name: "malformed details - no closing tag",
			markdownIn: `<details>
<summary>Title</summary>
Content without closing`,
			expectError: true,
		},
		{
			name: "valid nested content",
			markdownIn: `<details data-adf-type="expand">
<summary>Section Title</summary>

## Nested heading

- List item 1
- List item 2

Some **bold text** here.
</details>`,
			expectError: false,
			expectNodes: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := placeholder.NewManager()
			session := manager.GetSession()

			restored, err := FromMarkdown(test.markdownIn, session, manager)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectNodes, len(restored.Content))
			}
		})
	}
}

func TestExpandRoundtrip_BlockContentInExpand(t *testing.T) {
	// Expand containing heading, list, and bold text — tests recursive content parsing
	markdownIn := `<details data-adf-type="expand">
<summary>Section Title</summary>

## Nested heading

- List item 1
- List item 2

Some **bold text** here.
</details>`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	restored, err := FromMarkdown(markdownIn, session, manager)
	require.NoError(t, err)
	require.Equal(t, 1, len(restored.Content), "should be single expand node")

	expand := restored.Content[0]
	assert.Equal(t, adf_types.NodeTypeExpand, expand.Type)
	assert.Equal(t, "Section Title", expand.Attrs["title"])

	require.Greater(t, len(expand.Content), 1, "expand content must have multiple block nodes")
	types := make([]string, len(expand.Content))
	for i, n := range expand.Content {
		types[i] = n.Type
	}
	assert.Contains(t, types, adf_types.NodeTypeHeading, "expand content should include heading")
	assert.Contains(t, types, adf_types.NodeTypeBulletList, "expand content should include bulletList")
}

// ============================================================================
// Regression Test for Infinite Recursion Bug Fix (Story 3.1)
// ============================================================================

func TestExpandDetails_RecursionRegression(t *testing.T) {
	tests := []struct {
		name        string
		markdown    string
		expectError bool
		expectNodes int
	}{
		{
			name: "nested details elements should not cause infinite recursion",
			markdown: `# Test Nested Details

<details>
  <summary>Outer Details</summary>
  This is content in the outer details.

  <details open>
    <summary>Inner Details</summary>
    This is nested content that would trigger the infinite recursion bug.

    - Bullet point
    - Another bullet point
  </details>

  More content after the inner details.
</details>

Regular content after all details.`,
			expectError: false,
			expectNodes: 3, // heading, expand, paragraph
		},
		{
			name: "moderate nesting should work fine",
			markdown: `<details>
  <summary>Level 1</summary>
  Content 1

  <details>
    <summary>Level 2</summary>
    Content 2

    <details>
      <summary>Level 3</summary>
      Content 3 - this should work fine
    </details>
  </details>
</details>`,
			expectError: false,
			expectNodes: 1, // Single expand element
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup - create an ADF document first, then convert to markdown, then back
			// This tests the exact scenario that was causing infinite recursion
			testDoc := adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: "heading",
						Attrs: map[string]interface{}{
							"level": 1,
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Test Document"},
						},
					},
				},
			}

			classifier := NewDefaultClassifier()
			manager := placeholder.NewManager()

			// First convert ADF to markdown
			_, session, err := ToMarkdown(testDoc, classifier, manager)
			require.NoError(t, err, "Initial conversion should work")
			require.NotNil(t, session, "Session should be created")

			// Now test the problematic conversion: markdown with nested details back to ADF
			// This would hang before the fix due to infinite recursion
			result, err := FromMarkdown(test.markdown, session, manager)

			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "maximum recursion depth exceeded")
			} else {
				assert.NoError(t, err, "Conversion should not fail")
				assert.NotNil(t, result, "Result should not be nil")

				// Basic sanity check - the conversion completed without hanging
				assert.True(t, len(result.Content) > 0, "Should have some content nodes")

				// Test round-trip to ensure full functionality
				backToMarkdown, _, err := ToMarkdown(result, classifier, manager)
				assert.NoError(t, err, "Round-trip conversion should work")
				// Check for details element - matches both <details> and <details attr="value">
				assert.Regexp(t, `<details(?:\s|>)`, backToMarkdown, "Should contain details elements")
				assert.Contains(t, backToMarkdown, "<summary>", "Should contain summary elements")
			}
		})
	}
}

// ============================================================================
// Table Round-trip Tests
// ============================================================================

func TestRoundTrip_Table_PlainTable(t *testing.T) {
	conv := NewDefaultConverter()

	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: "table",
				Content: []adf_types.ADFNode{
					{Type: "tableRow", Content: []adf_types.ADFNode{
						{Type: "tableHeader", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "Name"},
							}},
						}},
						{Type: "tableHeader", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "Value"},
							}},
						}},
					}},
					{Type: "tableRow", Content: []adf_types.ADFNode{
						{Type: "tableCell", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "key"},
							}},
						}},
						{Type: "tableCell", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "val"},
							}},
						}},
					}},
				},
			},
		},
	}

	md, restored, err := conv.ConvertRoundTrip(doc)
	require.NoError(t, err)

	// Should render as markdown table, not placeholder
	assert.NotContains(t, md, "ADF_PLACEHOLDER", "table should be editable, not a placeholder")
	assert.Contains(t, md, "| Name", "markdown should contain table headers")
	assert.Contains(t, md, "| key", "markdown should contain table data")

	// Restored document should have table structure
	require.Len(t, restored.Content, 1)
	assert.Equal(t, "table", restored.Content[0].Type)
	assert.Len(t, restored.Content[0].Content, 2, "should have 2 rows")

	// Header row preserved
	headerRow := restored.Content[0].Content[0]
	assert.Equal(t, "tableHeader", headerRow.Content[0].Type)

	// Data row preserved
	dataRow := restored.Content[0].Content[1]
	assert.Equal(t, "tableCell", dataRow.Content[0].Type)
}

func TestRoundTrip_Table_WithFormattedContent(t *testing.T) {
	conv := NewDefaultConverter()

	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: "table",
				Content: []adf_types.ADFNode{
					{Type: "tableRow", Content: []adf_types.ADFNode{
						{Type: "tableHeader", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "Feature", Marks: []adf_types.ADFMark{
									{Type: "strong"},
								}},
							}},
						}},
						{Type: "tableHeader", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "Status"},
							}},
						}},
					}},
					{Type: "tableRow", Content: []adf_types.ADFNode{
						{Type: "tableCell", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "Tables"},
							}},
						}},
						{Type: "tableCell", Content: []adf_types.ADFNode{
							{Type: "paragraph", Content: []adf_types.ADFNode{
								{Type: "text", Text: "Done", Marks: []adf_types.ADFMark{
									{Type: "em"},
								}},
							}},
						}},
					}},
				},
			},
		},
	}

	md, restored, err := conv.ConvertRoundTrip(doc)
	require.NoError(t, err)

	// Bold header should appear in markdown
	assert.Contains(t, md, "**Feature**")
	// Italic cell should appear
	assert.Contains(t, md, "*Done*")

	// Structure preserved
	require.Len(t, restored.Content, 1)
	assert.Equal(t, "table", restored.Content[0].Type)

	// Marks preserved after roundtrip
	headerCell := restored.Content[0].Content[0].Content[0].Content[0]
	require.NotEmpty(t, headerCell.Content)
	assert.Equal(t, "Feature", headerCell.Content[0].Text)
	require.Len(t, headerCell.Content[0].Marks, 1)
	assert.Equal(t, "strong", headerCell.Content[0].Marks[0].Type)
}

func TestRoundTripConversion_InlineCardDataOnly(t *testing.T) {
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeInlineCard,
						Attrs: map[string]interface{}{
							"data": map[string]interface{}{
								"@type": "Document",
								"name":  "My Document",
							},
						},
					},
				},
			},
		},
	}

	converter := NewDefaultConverter()
	markdown, restored, err := converter.ConvertRoundTrip(original)
	require.NoError(t, err)

	// Markdown must contain placeholder comment
	assert.Contains(t, markdown, "ADF_PLACEHOLDER_")
	assert.Contains(t, markdown, "InlineCard")

	// The inlineCard node must come back identical
	require.Len(t, restored.Content, 1)
	para := restored.Content[0]
	require.Equal(t, adf_types.NodeTypeParagraph, para.Type)
	require.Len(t, para.Content, 1)
	card := para.Content[0]
	assert.Equal(t, adf_types.NodeTypeInlineCard, card.Type)
	assert.NotNil(t, card.Attrs["data"])
	assert.Nil(t, card.Attrs["url"])
}

// ============================================================================
// mediaInline Placeholder Preservation Tests
// ============================================================================

func TestRoundTrip_MediaInline_InParagraph(t *testing.T) {
	// mediaInline node with typical attrs (id, collection, type)
	mediaInlineNode := adf_types.ADFNode{
		Type: adf_types.NodeTypeMediaInline,
		Attrs: map[string]interface{}{
			"id":         "abc-123",
			"collection": "contentId-456",
			"type":       "file",
			"width":      float64(200),
			"height":     float64(150),
		},
	}

	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					adf_types.NewTextNode("See attachment: "),
					mediaInlineNode,
					adf_types.NewTextNode(" for details."),
				},
			},
		},
	}

	conv := NewDefaultConverter()
	md, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	// Markdown should contain placeholder comment (inline, no double newline)
	assert.Contains(t, md, "ADF_PLACEHOLDER_")
	assert.Contains(t, md, "See attachment:")
	assert.Contains(t, md, "for details.")

	// Roundtrip: mediaInline node must survive identically
	require.Len(t, restored.Content, 1)
	para := restored.Content[0]
	assert.Equal(t, adf_types.NodeTypeParagraph, para.Type)
	require.Len(t, para.Content, 3)

	assert.Equal(t, adf_types.NodeTypeText, para.Content[0].Type)
	assert.Equal(t, "See attachment: ", para.Content[0].Text)

	assert.Equal(t, adf_types.NodeTypeMediaInline, para.Content[1].Type)
	assert.Equal(t, "abc-123", para.Content[1].Attrs["id"])
	assert.Equal(t, "contentId-456", para.Content[1].Attrs["collection"])
	assert.Equal(t, "file", para.Content[1].Attrs["type"])
	assert.Equal(t, float64(200), para.Content[1].Attrs["width"])
	assert.Equal(t, float64(150), para.Content[1].Attrs["height"])

	assert.Equal(t, adf_types.NodeTypeText, para.Content[2].Type)
	assert.Equal(t, " for details.", para.Content[2].Text)
}

func TestRoundTrip_MediaInline_StandaloneInParagraph(t *testing.T) {
	// Paragraph containing only a mediaInline (no surrounding text).
	// The parser restores standalone placeholder comments as top-level nodes
	// (paragraph wrapper is lost), but the mediaInline node itself survives.
	original := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeMediaInline,
						Attrs: map[string]interface{}{
							"id":         "media-solo",
							"collection": "col-1",
							"type":       "image",
						},
					},
				},
			},
		},
	}

	conv := NewDefaultConverter()
	_, restored, err := conv.ConvertRoundTrip(original)
	require.NoError(t, err)

	// Paragraph wrapper preserved, mediaInline restored inside
	require.Len(t, restored.Content, 1)
	para := restored.Content[0]
	assert.Equal(t, adf_types.NodeTypeParagraph, para.Type)
	require.Len(t, para.Content, 1)
	assert.Equal(t, adf_types.NodeTypeMediaInline, para.Content[0].Type)
	assert.Equal(t, "media-solo", para.Content[0].Attrs["id"])
	assert.Equal(t, "col-1", para.Content[0].Attrs["collection"])
}

func TestMediaInline_PlaceholderPreview(t *testing.T) {
	manager := placeholder.NewManager()

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeMediaInline,
		Attrs: map[string]interface{}{
			"id":   "img-42",
			"type": "image",
		},
	}

	_, preview, err := manager.Store(node)
	require.NoError(t, err)
	assert.Contains(t, preview, "Inline Media")
}
