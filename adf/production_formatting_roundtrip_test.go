package adf_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProductionFormattingRoundtrip tests the exact scenario that failed in production
// Based on logs from NT-146 ticket processing
func TestProductionFormattingRoundtrip(t *testing.T) {
	// This is the exact markdown content that caused the production failure
	productionMarkdown := `# Test 1

<details>
  <summary>This is for testing</summary>
  # Heading

  - item 1
  - item 2
  - **bold** item
  - _italic_ item

  ## Subheading

  More content
</details>

# Test 2

<details>
  <summary>Further testing</summary>
  1. Some content
  2. More content
  <details>
    <summary>Further testing</summary>
    1. Some content
    2. More content
  </details>
</details>

Test Instructions

1. Test Markdown Data (paste this into a Jira ticket description):

Simple bullet list:

- First item
- Second item

Simple ordered list:

1. First item
2. Second item

Mixed content list:

- Simple text item
- **Bold text item**
- Item with [link](https://example.com)

Nested list:

- Parent item- Child item
- Another child

Complex list:

1. First numbered item
2. Second with **formatting**
3. Third item`

	conv := defaults.NewDefaultConverter()
	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Step 1: Convert Markdown to ADF
	adf, _, err := conv.FromMarkdown(productionMarkdown, session)
	require.NoError(t, err, "Markdown to ADF conversion should not fail")

	// DEBUG: Print ADF structure to see what we're getting
	if adfBytes, err := json.MarshalIndent(adf, "", "  "); err == nil {
		t.Logf("Generated ADF structure:\n%s", string(adfBytes))
	}

	// CRITICAL VALIDATION: Check for ADF structure integrity
	t.Run("ADF Structure Validation", func(t *testing.T) {
		// Validate no duplicate expand elements
		expandCount := countExpandElements(adf)
		assert.LessOrEqual(t, expandCount, 3, "Should not have duplicate expand elements")

		// Validate markdown formatting is converted to proper ADF marks
		validateFormattingMarks(t, adf)
	})

	// Step 2: Convert ADF back to Markdown
	roundtripMarkdown, _, err := conv.ToMarkdown(adf)
	require.NoError(t, err, "ADF to Markdown conversion should not fail")

	// CRITICAL VALIDATION: Check formatting preservation
	t.Run("Formatting Preservation", func(t *testing.T) {
		// Bold formatting should be preserved
		assert.Contains(t, roundtripMarkdown, "**bold** item", "Bold formatting in list items should be preserved")
		assert.Contains(t, roundtripMarkdown, "**Bold text item**", "Bold formatting should be preserved")
		assert.Contains(t, roundtripMarkdown, "**formatting**", "Bold formatting in numbered lists should be preserved")

		// Italic formatting should be preserved (note: ADF converts back to * syntax)
		assert.Contains(t, roundtripMarkdown, "*italic* item", "Italic formatting in list items should be preserved")

		// Heading formatting should be clean (no extra # characters)
		assert.Contains(t, roundtripMarkdown, "# Heading", "Heading should be clean without extra # characters")
		assert.NotContains(t, roundtripMarkdown, " #  # Heading", "Should not have malformed heading text")
	})

	// Step 3: Test full roundtrip (this is what failed in production)
	secondADF, _, err := conv.FromMarkdown(roundtripMarkdown, session)
	require.NoError(t, err, "Second markdown to ADF conversion should not fail")

	// CRITICAL: This is where the production failure occurred
	t.Run("Production API Compatibility", func(t *testing.T) {
		// Validate ADF structure is valid for Jira API
		err := validateJiraAPICompatibility(secondADF)
		assert.NoError(t, err, "ADF should be compatible with Jira API (production failed with HTTP 400)")
	})

	// Final validation: Multiple roundtrips should be stable
	thirdMarkdown, _, err := conv.ToMarkdown(secondADF)
	require.NoError(t, err, "Third conversion should not fail")

	// Content should stabilize after first roundtrip
	assert.Equal(t, roundtripMarkdown, thirdMarkdown, "Content should be stable across multiple roundtrips")
}

// Helper function to count expand elements in ADF structure
func countExpandElements(adfDoc adf.Document) int {
	count := 0
	for _, node := range adfDoc.Content {
		count += countExpandInNode(node)
	}
	return count
}

func countExpandInNode(node adf.Node) int {
	count := 0

	// Check if this node is an expand type
	if node.Type == adf.NodeTypeExpand {
		count++
	}

	// Recursively check content array
	for _, child := range node.Content {
		count += countExpandInNode(child)
	}

	return count
}

// Helper function to validate formatting marks in ADF
func validateFormattingMarks(t *testing.T, adfDoc adf.Document) {
	t.Helper()
	var issues []string

	for _, node := range adfDoc.Content {
		traverseAndValidateMarksInNode(node, &issues)
	}

	// Debug: Log all found issues
	t.Logf("Found %d formatting issues:", len(issues))
	for i, issue := range issues {
		t.Logf("Issue %d: %s", i+1, issue)
	}

	// Assert no formatting issues found
	for _, issue := range issues {
		t.Errorf("Formatting validation failed: %s", issue)
	}
}

func traverseAndValidateMarksInNode(node adf.Node, issues *[]string) {
	// Check for text nodes with raw markdown formatting
	if node.Type == adf.NodeTypeText {
		// Check for raw italic formatting that should be converted
		if strings.Contains(node.Text, "_italic_") {
			*issues = append(*issues, fmt.Sprintf("Raw italic formatting found in text: '%s' - should be converted to em marks", node.Text))
		}

		// Check for raw bold formatting that should be converted
		if strings.Contains(node.Text, "**bold**") {
			*issues = append(*issues, fmt.Sprintf("Raw bold formatting found in text: '%s' - should be converted to strong marks", node.Text))
		}

		// Validate that italic text has proper marks
		if strings.Contains(node.Text, "italic") && strings.Contains(node.Text, "_") {
			hasEmMark := false
			for _, mark := range node.Marks {
				if mark.Type == adf.MarkTypeEm {
					hasEmMark = true
					break
				}
			}
			if !hasEmMark {
				*issues = append(*issues, fmt.Sprintf("Text '%s' contains italic content but lacks em mark", node.Text))
			}
		}

		// Validate that bold text has proper marks
		if strings.Contains(node.Text, "bold") && strings.Contains(node.Text, "**") {
			hasStrongMark := false
			for _, mark := range node.Marks {
				if mark.Type == adf.MarkTypeStrong {
					hasStrongMark = true
					break
				}
			}
			if !hasStrongMark {
				*issues = append(*issues, fmt.Sprintf("Text '%s' contains bold content but lacks strong mark", node.Text))
			}
		}
	}

	// Recursively check content array
	for _, child := range node.Content {
		traverseAndValidateMarksInNode(child, issues)
	}
}

// Helper function to validate Jira API compatibility
func validateJiraAPICompatibility(adfDoc adf.Document) error {
	var errors []string

	// Track expand elements to detect duplicates
	expandTitles := make(map[string]int)

	for _, node := range adfDoc.Content {
		validateJiraStructureInNode(node, &errors, expandTitles, 0)
	}

	// Check for duplicate expand elements with same titles
	for title, count := range expandTitles {
		if count > 1 {
			errors = append(errors, fmt.Sprintf("Duplicate expand elements found with title '%s' (count: %d)", title, count))
		}
	}

	// Debug: Always log what we found
	fmt.Printf("DEBUG: Found %d expand titles, %d total errors\n", len(expandTitles), len(errors))
	for title, count := range expandTitles {
		fmt.Printf("DEBUG: Expand title '%s': %d occurrences\n", title, count)
	}

	if len(errors) > 0 {
		return fmt.Errorf("Jira API compatibility issues: %s", strings.Join(errors, "; "))
	}

	return nil
}

func validateJiraStructureInNode(node adf.Node, errors *[]string, expandTitles map[string]int, depth int) {
	// Prevent infinite recursion
	if depth > 50 {
		*errors = append(*errors, "Structure too deeply nested (max depth exceeded)")
		return
	}

	// Validate expand elements specifically
	if node.Type == adf.NodeTypeExpand {
		// Check for proper expand structure
		if title, ok := node.Attrs["title"].(string); ok {
			// Only count top-level expand elements (depth 0) for duplicate checking
			// Nested expand elements are allowed to have duplicate titles
			if depth == 0 {
				expandTitles[title]++
			}
		} else {
			*errors = append(*errors, "Expand element missing title attribute")
		}
	}

	// Validate text nodes don't contain raw markdown
	if node.Type == adf.NodeTypeText {
		// Critical: Check for raw markdown that should be processed
		if strings.Contains(node.Text, "_italic_") {
			*errors = append(*errors, fmt.Sprintf("Text node contains unprocessed italic markdown: '%s'", node.Text))
		}
		if strings.Contains(node.Text, "**bold**") {
			*errors = append(*errors, fmt.Sprintf("Text node contains unprocessed bold markdown: '%s'", node.Text))
		}

		// Check for malformed heading text
		if strings.Contains(node.Text, " #  # ") {
			*errors = append(*errors, fmt.Sprintf("Malformed heading text detected: '%s'", node.Text))
		}
	}

	// Recursively validate content
	for _, child := range node.Content {
		validateJiraStructureInNode(child, errors, expandTitles, depth+1)
	}
}

// TestSpecificFormattingIssues tests the individual components that failed
func TestSpecificFormattingIssues(t *testing.T) {
	conv := defaults.NewDefaultConverter()

	testCases := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Italic in list items",
			markdown: "- _italic_ item",
			expected: "*italic* item", // Should preserve italic formatting (ADF converts back to * syntax)
		},
		{
			name:     "Bold in list items",
			markdown: "- **bold** item",
			expected: "**bold** item", // Should preserve bold formatting
		},
		{
			name:     "Mixed formatting in lists",
			markdown: "- **Bold text item**\n- _italic_ item",
			expected: "**Bold text item**", // Both should be preserved
		},
		{
			name:     "Clean headings in details",
			markdown: "<details>\n<summary>Test</summary>\n# Heading\n</details>",
			expected: "# Heading", // Should not become " #  # Heading"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := placeholder.NewManager()
			session := manager.GetSession()

			adf, _, err := conv.FromMarkdown(tc.markdown, session)
			require.NoError(t, err)

			result, _, err := conv.ToMarkdown(adf)
			require.NoError(t, err)

			assert.Contains(t, result, tc.expected, "Formatting should be preserved in roundtrip")
		})
	}
}
