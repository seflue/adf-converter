package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// ============================================================================
// ADF-to-Markdown Conversion Tests (from adf_test.go)
// ============================================================================

func TestFromMarkdown_DeletedPlaceholder(t *testing.T) {
	// Test the FIXED behavior: when user deletes a placeholder comment,
	// the conversion should succeed and skip the deleted element.

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Store multiple placeholders to test selective deletion
	codeBlock := adf_types.ADFNode{
		Type: adf_types.NodeTypeCodeBlock,
		Attrs: map[string]interface{}{
			"language": "go",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeText,
				Text: `fmt.Println("Hello, World!")`,
			},
		},
	}

	panel := adf_types.ADFNode{
		Type: adf_types.NodeTypePanel,
		Attrs: map[string]interface{}{
			"panelType": "info",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Important information",
					},
				},
			},
		},
	}

	// Store placeholders - ADF_PLACEHOLDER_001 and ADF_PLACEHOLDER_002
	_, _, err := manager.Store(codeBlock)
	if err != nil {
		t.Fatalf("Failed to store code block: %v", err)
	}

	_, _, err = manager.Store(panel)
	if err != nil {
		t.Fatalf("Failed to store panel: %v", err)
	}

	// Create markdown that references both placeholders
	markdownWithBothPlaceholders := `This is before the first element.

<!-- ADF_PLACEHOLDER_001: Code Block (go, 1 lines): fmt.Println("Hello, World!") -->

This is between elements.

<!-- ADF_PLACEHOLDER_002: Info Panel: Important information -->

This is after the second element.
`

	// First, test that both placeholders work
	doc, err := FromMarkdown(markdownWithBothPlaceholders, session, manager)
	if err != nil {
		t.Fatalf("FromMarkdown failed with both placeholders: %v", err)
	}

	// Should have 5 nodes: paragraph, codeBlock, paragraph, panel, paragraph
	if len(doc.Content) != 5 {
		t.Errorf("Expected 5 nodes with both placeholders, got %d", len(doc.Content))
	}

	// Now test deletion: create markdown with middle placeholder deleted
	markdownWithDeletedPlaceholder := `This is before the first element.

<!-- ADF_PLACEHOLDER_001: Code Block (go, 1 lines): fmt.Println("Hello, World!") -->

This is between elements.

This is after the second element.
`

	// This should succeed after our fix - deleted placeholder is gracefully skipped
	doc, err = FromMarkdown(markdownWithDeletedPlaceholder, session, manager)
	if err != nil {
		t.Fatalf("FromMarkdown failed with deleted placeholder (should succeed after fix): %v", err)
	}

	// Should have 4 nodes: paragraph, codeBlock, paragraph, paragraph (deleted element missing)
	if len(doc.Content) != 4 {
		t.Errorf("Expected 4 nodes with deleted placeholder, got %d", len(doc.Content))
	}

	// Verify correct node types
	expectedTypes := []string{adf_types.NodeTypeParagraph, adf_types.NodeTypeCodeBlock, adf_types.NodeTypeParagraph, adf_types.NodeTypeParagraph}
	for i, expectedType := range expectedTypes {
		if doc.Content[i].Type != expectedType {
			t.Errorf("Node %d: expected %s, got %s", i, expectedType, doc.Content[i].Type)
		}
	}
}

// ============================================================================
// Markdown-to-ADF Conversion Tests (from markdown_test.go)
// ============================================================================

func TestToMarkdown_BasicDocument(t *testing.T) {
	// Create a simple ADF document with a paragraph
	doc := adf_types.ADFDocument{
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

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := ToMarkdown(doc, classifier, manager)
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	expected := "Hello, world!\n\n"
	if markdown != expected {
		t.Errorf("Expected %q, got %q", expected, markdown)
	}

	if session == nil {
		t.Fatal("Expected session to be returned")
	}

	if session.ID == "" {
		t.Error("Expected session to have an ID")
	}
}

func TestToMarkdown_HeadingWithText(t *testing.T) {
	// Create ADF document with heading
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 2,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "My Heading",
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, _, err := ToMarkdown(doc, classifier, manager)
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	expected := "## My Heading\n\n"
	if markdown != expected {
		t.Errorf("Expected %q, got %q", expected, markdown)
	}
}

func TestToMarkdown_TextWithMarks(t *testing.T) {
	// Create ADF document with formatted text
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Bold text",
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
						Text: "italic text",
						Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeEm},
						},
					},
				},
			},
		},
	}

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, _, err := ToMarkdown(doc, classifier, manager)
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	expected := "**Bold text** and *italic text*\n\n"
	if markdown != expected {
		t.Errorf("Expected %q, got %q", expected, markdown)
	}
}

// ============================================================================
// Link Processing Tests (from links_*.go files)
// ============================================================================

func TestBasicLinks(t *testing.T) {
	basicLinksADF := `{
		"fields": {
			"description": {
				"content": [
					{
						"content": [
							{
								"text": "Visit ",
								"type": "text"
							},
							{
								"marks": [
									{
										"attrs": {
											"href": "https://google.com"
										},
										"type": "link"
									}
								],
								"text": "Google",
								"type": "text"
							},
							{
								"text": " for search and ",
								"type": "text"
							},
							{
								"marks": [
									{
										"attrs": {
											"href": "https://github.com"
										},
										"type": "link"
									}
								],
								"text": "GitHub",
								"type": "text"
							},
							{
								"text": " for code.",
								"type": "text"
							}
						],
						"type": "paragraph"
					}
				],
				"type": "doc",
				"version": 1
			}
		}
	}`

	testDoc := parseTestADFPayload(t, basicLinksADF)

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := ToMarkdown(testDoc, classifier, manager)
	require.NoError(t, err)

	expectedMarkdown := "Visit [Google](https://google.com) for search and [GitHub](https://github.com) for code.\n\n"
	assert.Equal(t, expectedMarkdown, markdown)

	// Verify round-trip conversion
	convertedBack, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Should preserve link structure
	assert.Equal(t, adf_types.NodeTypeParagraph, convertedBack.Content[0].Type)

	// Check that links are preserved in the paragraph content
	paragraph := convertedBack.Content[0]
	hasLinks := false
	for _, node := range paragraph.Content {
		if node.Type == adf_types.NodeTypeText && len(node.Marks) > 0 {
			for _, mark := range node.Marks {
				if mark.Type == adf_types.MarkTypeLink {
					hasLinks = true
					break
				}
			}
		}
	}
	assert.True(t, hasLinks, "Links should be preserved in round-trip conversion")
}

func TestLinksWithFormatting(t *testing.T) {
	formattingLinksADF := `{
		"fields": {
			"description": {
				"content": [
					{
						"content": [
							{
								"text": "This is ",
								"type": "text"
							},
							{
								"marks": [
									{
										"attrs": {
											"href": "https://example.com"
										},
										"type": "link"
									},
									{
										"type": "strong"
									}
								],
								"text": "bold link",
								"type": "text"
							},
							{
								"text": " and ",
								"type": "text"
							},
							{
								"marks": [
									{
										"attrs": {
											"href": "https://github.com"
										},
										"type": "link"
									},
									{
										"type": "em"
									}
								],
								"text": "italic link",
								"type": "text"
							},
							{
								"text": ".",
								"type": "text"
							}
						],
						"type": "paragraph"
					}
				],
				"type": "doc",
				"version": 1
			}
		}
	}`

	testDoc := parseTestADFPayload(t, formattingLinksADF)

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := ToMarkdown(testDoc, classifier, manager)
	require.NoError(t, err)

	// Should contain both formatted links (formatting around whole link)
	assert.Contains(t, markdown, "**[bold link](https://example.com)**")
	assert.Contains(t, markdown, "*[italic link](https://github.com)*")

	// Verify round-trip conversion
	convertedBack, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Should preserve both links and formatting
	assert.Equal(t, adf_types.NodeTypeParagraph, convertedBack.Content[0].Type)
}

// ============================================================================
// Helper Functions
// ============================================================================

func TestLinksInLists(t *testing.T) {
	listsWithLinksADF := `{
		"fields": {
			"description": {
				"content": [
					{
						"content": [
							{
								"content": [
									{
										"content": [
											{
												"text": "External: ",
												"type": "text"
											},
											{
												"marks": [
													{
														"attrs": {
															"href": "https://stackoverflow.com"
														},
														"type": "link"
													}
												],
												"text": "Stack Overflow",
												"type": "text"
											}
										],
										"type": "paragraph"
									}
								],
								"type": "listItem"
							}
						],
						"type": "bulletList"
					}
				],
				"type": "doc",
				"version": 1
			}
		}
	}`

	testDoc := parseTestADFPayload(t, listsWithLinksADF)

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := ToMarkdown(testDoc, classifier, manager)
	require.NoError(t, err)

	// Should contain bullet list with link
	assert.Contains(t, markdown, "- External: [Stack Overflow](https://stackoverflow.com)")

	// Verify round-trip conversion
	convertedBack, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Should preserve list structure
	assert.Equal(t, adf_types.NodeTypeBulletList, convertedBack.Content[0].Type)
}

func TestLinksWithSpecialCharacters(t *testing.T) {
	specialCharsLinksADF := `{
		"fields": {
			"description": {
				"content": [
					{
						"content": [
							{
								"marks": [
									{
										"attrs": {
											"href": "https://company.atlassian.com/browse/PAREN-123"
										},
										"type": "link"
									}
								],
								"text": "Link with (parentheses)",
								"type": "text"
							}
						],
						"type": "paragraph"
					}
				],
				"type": "doc",
				"version": 1
			}
		}
	}`

	testDoc := parseTestADFPayload(t, specialCharsLinksADF)

	classifier := NewDefaultClassifier()
	manager := placeholder.NewManager()

	markdown, session, err := ToMarkdown(testDoc, classifier, manager)
	require.NoError(t, err)

	// Should handle special characters in link text
	assert.Contains(t, markdown, "[Link with (parentheses)](https://company.atlassian.com/browse/PAREN-123)")

	// Verify round-trip conversion
	convertedBack, err := FromMarkdown(markdown, session, manager)
	require.NoError(t, err)

	// Should preserve link structure
	assert.Equal(t, adf_types.NodeTypeParagraph, convertedBack.Content[0].Type)
}

// ============================================================================
// Classification and Enhanced Converter Tests
// ============================================================================

func TestDefaultClassifier_IsEditable(t *testing.T) {
	classifier := NewDefaultClassifier()

	tests := []struct {
		nodeType string
		expected bool
	}{
		{adf_types.NodeTypeParagraph, true},
		{adf_types.NodeTypeHeading, true},
		{adf_types.NodeTypeText, true},
		{adf_types.NodeTypeHardBreak, true},
		{adf_types.NodeTypeOrderedList, true},
		{adf_types.NodeTypeBulletList, true},
		{adf_types.NodeTypeListItem, true},
		{adf_types.NodeTypeCodeBlock, true},
		{adf_types.NodeTypeTable, true},
		{adf_types.NodeTypePanel, true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			result := classifier.IsEditable(tt.nodeType)
			if result != tt.expected {
				t.Errorf("IsEditable(%s) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

func TestDefaultClassifier_IsPreserved(t *testing.T) {
	classifier := NewDefaultClassifier()

	tests := []struct {
		nodeType string
		expected bool
	}{
		{adf_types.NodeTypeCodeBlock, false},
		{adf_types.NodeTypeTable, false},
		{adf_types.NodeTypeTableRow, false},
		{adf_types.NodeTypeTableCell, false},
		{adf_types.NodeTypePanel, false},
		{adf_types.NodeTypeExpand, false},
		{adf_types.NodeTypeMediaInline, true},
		{adf_types.NodeTypeParagraph, false},
		{adf_types.NodeTypeText, false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			result := classifier.IsPreserved(tt.nodeType)
			if result != tt.expected {
				t.Errorf("IsPreserved(%s) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

// Note: parseTestADFPayload helper function is available from test_helpers.go
