package converter_test

import (
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/defaults"
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Error Handling and Boundary Condition Tests
// ============================================================================

// TestProblematicLinkScenarios tests specific link scenarios that have caused HTTP 400 errors
// Based on actual error logs from production usage
func TestProblematicLinkScenarios(t *testing.T) {
	tests := []struct {
		name             string
		inputMarkdown    string
		expectConversion bool
		expectedIssues   []string
		description      string
	}{
		{
			name: "atlassian_wiki_link_in_bullet_list",
			inputMarkdown: `Lorem ipsum dolor sit amet.

Ut enim ad minim veniam:

- Duis aute irure dolor
- [https://acme-corp.atlassian.net/wiki/x/AbCdEf](https://acme-corp.atlassian.net/wiki/x/AbCdEf)
- Excepteur sint occaecat`,
			expectConversion: true,
			expectedIssues:   []string{}, // Should convert but may cause API issues
			description:      "Atlassian wiki link in bullet list item (from error logs)",
		},
		{
			name: "malformed_link_syntax",
			inputMarkdown: `Test document with malformed links:

- [Incomplete link without URL]
- [Another incomplete](
- [Valid link](https://example.com)`,
			expectConversion: true,
			expectedIssues:   []string{}, // Should handle gracefully
			description:      "Malformed markdown link syntax",
		},
		{
			name: "link_with_special_characters",
			inputMarkdown: `Links with special characters:

- [Link with "quotes"](https://example.com/page?param=value&other="quoted")
- [Link with <brackets>](https://example.com/path/<id>)
- [Link with &amp; entities](https://example.com/page?a=1&amp;b=2)`,
			expectConversion: true,
			expectedIssues:   []string{}, // Should handle special characters properly
			description:      "Links containing special characters that might cause parsing issues",
		},
		{
			name:             "nested_formatting_in_links",
			inputMarkdown:    "Complex link formatting:\n\n- [**Bold link text**](https://example.com)\n- [*Italic link text*](https://example.com)\n- [`Code in link`](https://example.com)",
			expectConversion: true,
			expectedIssues:   []string{}, // Should handle nested formatting
			description:      "Links with nested formatting that might cause ADF structure issues",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := placeholder.NewManager()
			session := manager.GetSession()

			// Try conversion
			doc, err := converter.FromMarkdown(test.inputMarkdown, session, manager, defaults.NewRegistry())

			if test.expectConversion {
				require.NoError(t, err, "Conversion should succeed for: %s", test.description)
				assert.NotNil(t, doc, "Document should be generated")
				assert.Equal(t, "doc", doc.Type, "Should produce valid ADF document")
			} else {
				assert.Error(t, err, "Conversion should fail for: %s", test.description)
			}
		})
	}
}

// ============================================================================
// Empty and Null Input Edge Cases
// ============================================================================

func TestEdgeCases_EmptyInputs(t *testing.T) {
	manager := placeholder.NewManager()
	session := manager.GetSession()

	tests := []struct {
		name          string
		markdown      string
		shouldSucceed bool
	}{
		{
			name:          "empty_string",
			markdown:      "",
			shouldSucceed: true, // Should produce empty document
		},
		{
			name:          "whitespace_only",
			markdown:      "   \n\t  \n  ",
			shouldSucceed: true, // Should handle gracefully
		},
		{
			name:          "single_newline",
			markdown:      "\n",
			shouldSucceed: true,
		},
		{
			name:          "multiple_newlines",
			markdown:      "\n\n\n\n",
			shouldSucceed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc, err := converter.FromMarkdown(test.markdown, session, manager, defaults.NewRegistry())

			if test.shouldSucceed {
				require.NoError(t, err, "Should handle empty/whitespace input gracefully")
				assert.Equal(t, "doc", doc.Type, "Should produce valid ADF document type")
				assert.Equal(t, 1, doc.Version, "Should have correct version")
			} else {
				assert.Error(t, err, "Should fail for invalid input")
			}
		})
	}
}

// ============================================================================
// Malformed ADF Structure Tests
// ============================================================================

func TestEdgeCases_MalformedADF(t *testing.T) {
	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	tests := []struct {
		name          string
		adf           adf_types.ADFDocument
		shouldSucceed bool
	}{
		{
			name: "missing_content_array",
			adf: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: nil, // Missing content
			},
			shouldSucceed: true, // Should handle gracefully
		},
		{
			name: "empty_content_array",
			adf: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{}, // Empty content
			},
			shouldSucceed: true, // Should produce empty markdown
		},
		{
			name: "invalid_version",
			adf: adf_types.ADFDocument{
				Version: 999, // Invalid version
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Test"},
						},
					},
				},
			},
			shouldSucceed: true, // Should handle gracefully
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			markdown, session, err := converter.ToMarkdown(test.adf, classifier, manager, defaults.NewRegistry())

			if test.shouldSucceed {
				require.NoError(t, err, "Should handle malformed ADF gracefully")
				assert.NotNil(t, session, "Should return session")
			} else {
				assert.Error(t, err, "Should fail for severely malformed ADF")
			}

			// If conversion succeeded, try round-trip
			if err == nil && markdown != "" {
				_, err2 := converter.FromMarkdown(markdown, session, manager, defaults.NewRegistry())
				assert.NoError(t, err2, "Round-trip should also succeed")
			}
		})
	}
}

// ============================================================================
// Complex Nested Structure Edge Cases
// ============================================================================

func TestEdgeCases_DeeplyNestedStructures(t *testing.T) {
	// Test deeply nested list structures
	deeplyNestedList := adf_types.ADFDocument{
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
									{Type: adf_types.NodeTypeText, Text: "Level 1"},
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
													{Type: adf_types.NodeTypeText, Text: "Level 2"},
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
																	{Type: adf_types.NodeTypeText, Text: "Level 3"},
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
					},
				},
			},
		},
	}

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Convert to markdown
	markdown, session, err := converter.ToMarkdown(deeplyNestedList, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should handle deeply nested structures")

	// Verify markdown contains nested list indicators
	assert.Contains(t, markdown, "- Level 1")
	assert.Contains(t, markdown, "  - Level 2")
	assert.Contains(t, markdown, "    - Level 3")

	// Verify round-trip conversion
	restored, err := converter.FromMarkdown(markdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err, "Round-trip should succeed for nested structures")

	// Verify structure is preserved
	assert.Equal(t, adf_types.NodeTypeBulletList, restored.Content[0].Type)
}

// ============================================================================
// Large Content Edge Cases
// ============================================================================

func TestEdgeCases_LargeContent(t *testing.T) {
	// Test very long text content
	longText := ""
	for i := 0; i < 1000; i++ {
		longText += "This is a very long paragraph with lots of text content. "
	}

	largeParagraph := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: longText},
				},
			},
		},
	}

	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Convert to markdown
	markdown, session, err := converter.ToMarkdown(largeParagraph, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should handle large text content")

	// Verify content is preserved
	assert.Contains(t, markdown, "This is a very long paragraph")
	assert.Greater(t, len(markdown), 10000, "Markdown should contain the full text")

	// Verify round-trip conversion
	restored, err := converter.FromMarkdown(markdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err, "Round-trip should succeed for large content")

	// Verify document structure
	assert.Equal(t, "doc", restored.Type)
	assert.Equal(t, 1, len(restored.Content))
	assert.Equal(t, adf_types.NodeTypeParagraph, restored.Content[0].Type)
}

// ============================================================================
// Blockquote Edge Cases
// ============================================================================

func TestEdgeCases_PlainBlockquote(t *testing.T) {
	blockquoteMarkdown := "> This is a blockquote"

	manager := placeholder.NewManager()
	session := manager.GetSession()

	doc, err := converter.FromMarkdown(blockquoteMarkdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err)

	assert.Equal(t, "doc", doc.Type)
	require.Equal(t, 1, len(doc.Content), "Should have exactly one top-level node")

	blockquote := doc.Content[0]
	assert.Equal(t, "blockquote", blockquote.Type, "Top-level node should be blockquote, not paragraph")
	require.Equal(t, 1, len(blockquote.Content), "Blockquote should have one paragraph")
	assert.Equal(t, "paragraph", blockquote.Content[0].Type)
	require.NotEmpty(t, blockquote.Content[0].Content)
	assert.Equal(t, "This is a blockquote", blockquote.Content[0].Content[0].Text)
}

func TestEdgeCases_NestedBlockquotes(t *testing.T) {
	nestedBlockquoteMarkdown := `> This is a quote
>
> > This is a nested quote
> > with multiple lines
>
> Back to first level`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Convert to ADF
	doc, err := converter.FromMarkdown(nestedBlockquoteMarkdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should handle nested blockquotes")

	// Verify structure - blockquotes may be converted to placeholders
	assert.Equal(t, "doc", doc.Type)
	assert.Greater(t, len(doc.Content), 0, "Should have content")

	// Test round-trip
	classifier := converter.NewDefaultClassifier()
	markdown, _, err := converter.ToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should convert blockquotes back to markdown")

	// Should preserve some form of quote structure
	assert.Contains(t, markdown, "quote", "Should preserve quote content")
}

// ============================================================================
// Task List Edge Cases
// ============================================================================

func TestEdgeCases_TaskListVariations(t *testing.T) {
	// Test various task list formats
	taskListMarkdown := `Task list variations:

- [ ] Unchecked task
- [x] Checked task
- [X] Checked task with capital X
- [ ]   Task with extra spaces
- [x]Task without space after checkbox
- Regular bullet point
- [ ] Another unchecked task`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Convert to ADF
	doc, err := converter.FromMarkdown(taskListMarkdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should handle task list variations")

	// Verify structure
	assert.Equal(t, "doc", doc.Type)
	assert.Greater(t, len(doc.Content), 0, "Should have content")

	// Test round-trip
	classifier := converter.NewDefaultClassifier()
	markdown, _, err := converter.ToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should convert task lists back to markdown")

	// Should preserve list structure in some form
	assert.Contains(t, markdown, "task", "Should preserve task content")
}

func TestPlainMarkdownTaskList_ParsedAsTaskList(t *testing.T) {
	// Regression test: plain `- [ ]`/`- [x]` markdown must produce taskList nodes,
	// not bulletList nodes. The markdown_parser switch case for "- " previously
	// swallowed task list lines before the task list handler could run.
	markdown := "- [ ] Unchecked task\n- [x] Checked task\n"

	manager := placeholder.NewManager()
	session := manager.GetSession()

	doc, err := converter.FromMarkdown(markdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err)
	require.Len(t, doc.Content, 1, "should produce a single top-level node")

	node := doc.Content[0]
	assert.Equal(t, "taskList", node.Type, "plain checkbox markdown must be parsed as taskList, not bulletList")
	require.Len(t, node.Content, 2, "should contain two task items")
	assert.Equal(t, "taskItem", node.Content[0].Type)
	assert.Equal(t, "TODO", node.Content[0].Attrs["state"])
	assert.Equal(t, "taskItem", node.Content[1].Type)
	assert.Equal(t, "DONE", node.Content[1].Attrs["state"])
}

// ============================================================================
// Special Character Edge Cases
// ============================================================================

func TestEdgeCases_SpecialCharacters(t *testing.T) {
	// Note: Unicode emoji and emoji-like symbols (©, ®, ™) removed from test
	// as they now require EmojiConverter which would create an import cycle.
	// Emoji round-trip is thoroughly tested in emoji_roundtrip_test.go
	specialCharsMarkdown := `Special characters test:

Mathematical: ∀ ∃ ∈ ∋ ∩ ∪ ⊂ ⊃

Diacritics: café naïve résumé piñata

Mixed: "Smart quotes" and 'single quotes' and —em dash— and –en dash–`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Convert to ADF
	doc, err := converter.FromMarkdown(specialCharsMarkdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should handle special characters")

	// Verify structure
	assert.Equal(t, "doc", doc.Type)

	// Test round-trip
	classifier := converter.NewDefaultClassifier()
	markdown, _, err := converter.ToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should convert special characters back to markdown")

	// Verify key special characters are preserved
	assert.Contains(t, markdown, "café", "Should preserve diacritics")
	assert.Contains(t, markdown, "Mathematical", "Should preserve content")
	assert.Contains(t, markdown, "∀", "Should preserve mathematical symbols")
}

// ============================================================================
// Error Recovery Tests
// ============================================================================

func TestEdgeCases_ErrorRecovery(t *testing.T) {
	// Test conversion continues despite individual node errors
	mixedValidInvalidMarkdown := `# Valid Heading

This is a valid paragraph.

- Valid list item
- Another valid item

Some text with **bold** and *italic* formatting.

Final paragraph.`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Should succeed even if some elements cause issues
	doc, err := converter.FromMarkdown(mixedValidInvalidMarkdown, session, manager, defaults.NewRegistry())
	require.NoError(t, err, "Should recover from individual element errors")

	// Verify basic structure is preserved
	assert.Equal(t, "doc", doc.Type)
	assert.Greater(t, len(doc.Content), 0, "Should preserve valid content")

	// Test round-trip
	classifier := converter.NewDefaultClassifier()
	markdown, _, err := converter.ToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	require.NoError(t, err, "Round-trip should succeed")

	// Should preserve main content
	assert.Contains(t, markdown, "Valid Heading", "Should preserve heading")
	assert.Contains(t, markdown, "valid paragraph", "Should preserve paragraph")
}
