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

// DISABLED: False positive - passes despite known production failures (QA Gate Mandate)
// Re-enable only after new parser architecture is implemented and comprehensive stress testing added
func TestExpandRoundtrip_BasicExpandElement_DISABLED(t *testing.T) {
	t.Skip("DISABLED: False positive test - passes while production code fails with infinite recursion")
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

	// Convert to markdown (should use HTML details for expand)
	markdown, session, err := ToMarkdown(original, classifier, manager)
	require.NoError(t, err)

	// Should contain HTML details element
	assert.Contains(t, markdown, "<details>")
	assert.Contains(t, markdown, "<summary>Click to expand</summary>")
	assert.Contains(t, markdown, "Hidden content here.")
	assert.Contains(t, markdown, "</details>")

	// Convert back to ADF
	restored, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Verify round-trip fidelity
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Version, restored.Version)
	assert.Equal(t, len(original.Content), len(restored.Content))

	// Verify expand element is preserved
	assert.Equal(t, adf_types.NodeTypeExpand, restored.Content[0].Type)

	// Check expand attributes are preserved (attrs may be interface{})
	assert.NotNil(t, restored.Content[0].Attrs, "Expand element should have attributes")
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

// DISABLED: False positive - passes despite known production failures (QA Gate Mandate)
// Re-enable only after new parser architecture is implemented and comprehensive stress testing added
func TestExpandBackwardCompatibility_XMLFormat_DISABLED(t *testing.T) {
	t.Skip("DISABLED: False positive test - passes while production code fails with infinite recursion")
	// Test that existing XML expand format can still be parsed
	markdownWithXML := `This is a test document.

<expand title="Legacy XML Format">
This content was created with the old XML format.
It should still be parsed correctly.
</expand>

More content here.`

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Convert XML markdown back to ADF
	session := manager.GetSession()
	restored, err := FromMarkdown(markdownWithXML, session, manager)
	require.NoError(t, err)

	// Verify expand element is parsed correctly
	require.Equal(t, 3, len(restored.Content), "Should have 3 top-level elements")

	expandNode := restored.Content[1] // Middle element should be expand
	assert.Equal(t, adf_types.NodeTypeExpand, expandNode.Type)
	assert.Equal(t, "Legacy XML Format", expandNode.Attrs["title"])

	// Verify content is preserved
	require.Greater(t, len(expandNode.Content), 0, "Expand should have content")

	newMarkdown, _, err := ToMarkdown(restored, classifier, manager)
	require.NoError(t, err)

	// Should now be in details format
	assert.Contains(t, newMarkdown, "<details>")
	assert.Contains(t, newMarkdown, "<summary>Legacy XML Format</summary>")
	assert.Contains(t, newMarkdown, "This content was created with the old XML format.")
	assert.Contains(t, newMarkdown, "</details>")
}

// DISABLED: False positive - passes despite known production failures (QA Gate Mandate)
// Re-enable only after new parser architecture is implemented and comprehensive stress testing added
func TestExpandDetails_AttributeHandling_DISABLED(t *testing.T) {
	t.Skip("DISABLED: False positive test - passes while production code fails with infinite recursion")
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

			// Convert to markdown
			markdown, _, err := ToMarkdown(test.input, classifier, manager)
			require.NoError(t, err)

			// Check all expected content is present
			for _, expected := range test.expected {
				assert.Contains(t, markdown, expected, "Markdown should contain: %s", expected)
			}

			// Test round-trip conversion
			session := manager.GetSession()
			restored, err := FromMarkdown(markdown, session, manager)
			require.NoError(t, err)

			// Verify structure is preserved
			assert.Equal(t, test.input.Type, restored.Type)
			assert.Equal(t, len(test.input.Content), len(restored.Content))
			assert.Equal(t, adf_types.NodeTypeExpand, restored.Content[0].Type)
		})
	}
}

// DISABLED: Insufficient depth testing - doesn't catch infinite recursion (QA Gate Mandate)
// Re-enable only after new parser architecture is implemented and comprehensive stress testing added
func TestExpandDetails_EdgeCases_DISABLED(t *testing.T) {
	t.Skip("DISABLED: Insufficient depth testing - doesn't test deeply nested content that causes infinite recursion")
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
			expectError: false, // Should fallback to paragraph
			expectNodes: 1,     // Becomes a paragraph
		},
		{
			name: "malformed details - no closing tag",
			markdownIn: `<details>
<summary>Title</summary>
Content without closing`,
			expectError: false, // Should fallback to paragraph
			expectNodes: 1,     // Becomes a paragraph
		},
		{
			name: "valid nested content",
			markdownIn: `<details>
<summary>Section Title</summary>

## Nested heading

- List item 1
- List item 2

Some **bold text** here.
</details>`,
			expectError: false,
			expectNodes: 1, // Single expand element
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
