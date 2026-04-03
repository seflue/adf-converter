package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"adf-converter/adf_types"
)

// TestDetailsContentParsingRegression tests the specific production issue
// reported where details content has malformed headings and unprocessed markdown
func TestDetailsContentParsingRegression(t *testing.T) {
	// This is the exact ADF from production logs that was causing issues
	originalADF := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: "expand",
				Attrs: map[string]interface{}{
					"title": "This is for testing",
				},
				Content: []adf_types.ADFNode{
					{
						Type: "heading",
						Attrs: map[string]interface{}{
							"level": 1,
						},
						Content: []adf_types.ADFNode{
							{
								Type: "text",
								Text: "Heading",
							},
						},
					},
					{
						Type: "bulletList",
						Content: []adf_types.ADFNode{
							{
								Type: "listItem",
								Content: []adf_types.ADFNode{
									{
										Type: "paragraph",
										Content: []adf_types.ADFNode{
											{
												Type: "text",
												Text: "bold item",
												Marks: []adf_types.ADFMark{
													{Type: "strong"},
												},
											},
										},
									},
								},
							},
							{
								Type: "listItem",
								Content: []adf_types.ADFNode{
									{
										Type: "paragraph",
										Content: []adf_types.ADFNode{
											{
												Type: "text",
												Text: "italic item",
												Marks: []adf_types.ADFMark{
													{Type: "em"},
												},
											},
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

	conv := NewDefaultConverter()

	// Convert to markdown and back
	markdown, restoredADF, err := conv.ConvertRoundTrip(originalADF)
	require.NoError(t, err, "Round trip conversion should succeed")

	t.Logf("Generated Markdown:\n%s", markdown)

	// Verify the restored ADF structure
	require.Equal(t, 1, len(restoredADF.Content), "Should have one top-level element")

	expand := restoredADF.Content[0]
	assert.Equal(t, "expand", string(expand.Type), "Top-level element should be expand")
	assert.Equal(t, "This is for testing", expand.Attrs["title"], "Title should be preserved")

	// CRITICAL: Check that heading text is clean (not malformed)
	require.Greater(t, len(expand.Content), 0, "Expand should have content")
	heading := expand.Content[0]
	assert.Equal(t, "heading", string(heading.Type), "First content should be heading")

	require.Greater(t, len(heading.Content), 0, "Heading should have text content")
	headingText := heading.Content[0]

	// This was the production bug: heading text was " #  # Heading" instead of "Heading"
	assert.Equal(t, "Heading", headingText.Text, "CRITICAL: Heading text should be clean, not malformed with '# #'")

	// CRITICAL: Check that markdown formatting is processed into ADF marks
	require.Greater(t, len(expand.Content), 1, "Expand should have bullet list")
	bulletList := expand.Content[1]
	assert.Equal(t, "bulletList", string(bulletList.Type), "Second content should be bullet list")

	require.Greater(t, len(bulletList.Content), 0, "Bullet list should have items")

	// Check first list item (bold)
	item1 := bulletList.Content[0]
	require.Greater(t, len(item1.Content), 0, "List item should have content")
	require.Greater(t, len(item1.Content[0].Content), 0, "List item paragraph should have text")

	boldText := item1.Content[0].Content[0]
	assert.Equal(t, "bold item", boldText.Text, "Bold text content should be correct")
	// This was the production bug: markdown "**bold** item" was not converted to ADF marks
	assert.Equal(t, 1, len(boldText.Marks), "Bold text should have exactly one mark")
	assert.Equal(t, "strong", string(boldText.Marks[0].Type), "First item should have strong mark")

	// Check second list item (italic)
	require.Greater(t, len(bulletList.Content), 1, "Should have second list item")
	item2 := bulletList.Content[1]
	require.Greater(t, len(item2.Content), 0, "Second list item should have content")
	require.Greater(t, len(item2.Content[0].Content), 0, "Second list item paragraph should have text")

	italicText := item2.Content[0].Content[0]
	assert.Equal(t, "italic item", italicText.Text, "Italic text content should be correct")
	// This was the production bug: markdown "_italic_ item" was not converted to ADF marks
	assert.Equal(t, 1, len(italicText.Marks), "Italic text should have exactly one mark")
	assert.Equal(t, "em", string(italicText.Marks[0].Type), "Second item should have em mark")
}

// TestDetailsContentParsingEdgeCases tests additional edge cases around details content
func TestDetailsContentParsingEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		inputADF    adf_types.ADFDocument
		expectError bool
		validate    func(t *testing.T, restored adf_types.ADFDocument)
	}{
		{
			name: "nested_details_with_mixed_content",
			inputADF: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: "expand",
						Attrs: map[string]interface{}{
							"title": "Outer Section",
						},
						Content: []adf_types.ADFNode{
							{
								Type: "paragraph",
								Content: []adf_types.ADFNode{
									{Type: "text", Text: "Some text"},
								},
							},
							{
								Type: "expand",
								Attrs: map[string]interface{}{
									"title": "Inner Section",
								},
								Content: []adf_types.ADFNode{
									{
										Type:  "heading",
										Attrs: map[string]interface{}{"level": 2},
										Content: []adf_types.ADFNode{
											{Type: "text", Text: "Nested Heading"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, restored adf_types.ADFDocument) {
				// Should not hang or cause infinite recursion
				assert.Equal(t, 1, len(restored.Content), "Should have single expand element")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conv := NewDefaultConverter()

			_, restored, err := conv.ConvertRoundTrip(tc.inputADF)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, restored)
				}
			}
		})
	}
}
