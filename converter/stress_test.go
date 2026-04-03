package converter

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// TestStreamingParser_StressTest validates new parser architecture with deeply nested content
// This addresses critical QA gate requirement for comprehensive stress testing
func TestStreamingParser_StressTest(t *testing.T) {
	tests := []struct {
		name        string
		depth       int
		expectError bool
		timeout     time.Duration
	}{
		{
			name:        "reasonable nesting (10 levels)",
			depth:       10,
			expectError: false,
			timeout:     5 * time.Second,
		},
		{
			name:        "deep nesting (50 levels)",
			depth:       50,
			expectError: false,
			timeout:     10 * time.Second,
		},
		{
			name:        "extreme nesting (100+ levels)",
			depth:       120,
			expectError: true, // Should fail gracefully, not hang
			timeout:     2 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create deeply nested details content
			markdown := generateNestedDetailsMarkdown(test.depth)

			manager := placeholder.NewManager()
			session := manager.GetSession()

			// Test with timeout to prevent infinite recursion hangs
			done := make(chan bool, 1)
			var err error

			go func() {
				parser := NewMarkdownParser(session, manager)
				_, err = parser.ParseMarkdownToADFNodes(strings.Split(markdown, "\n"))
				done <- true
			}()

			select {
			case <-done:
				// Parsing completed
				if test.expectError {
					assert.Error(t, err, "Expected error for extreme nesting depth")
				} else {
					assert.NoError(t, err, "Should parse reasonable nesting without error")
				}
			case <-time.After(test.timeout):
				t.Fatalf("CRITICAL: Parser hung for %d levels - infinite recursion detected!", test.depth)
			}
		})
	}
}

// TestStreamingParser_InfiniteRecursionPrevention tests the specific vulnerability that caused production issues
func TestStreamingParser_InfiniteRecursionPrevention(t *testing.T) {
	// This is the exact pattern that caused infinite recursion in production
	problematicMarkdown := `<details>
<summary>Outer expand</summary>
Some content
<details>
<summary>Inner expand</summary>
<details>
<summary>Even deeper</summary>
Deeply nested content
</details>
</details>
</details>`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// This must complete within reasonable time (no hanging)
	done := make(chan bool, 1)
	var result []adf_types.ADFNode
	var err error

	go func() {
		parser := NewMarkdownParser(session, manager)
		result, err = parser.ParseMarkdownToADFNodes(strings.Split(problematicMarkdown, "\n"))
		done <- true
	}()

	select {
	case <-done:
		// Success - no infinite recursion
		require.NoError(t, err)
		assert.NotEmpty(t, result, "Should parse nested details successfully")

		// Verify structure was preserved
		assert.Equal(t, adf_types.NodeTypeExpand, result[0].Type)
		assert.Equal(t, "Outer expand", result[0].Attrs["title"])

	case <-time.After(3 * time.Second):
		t.Fatal("CRITICAL: Parser hung on nested details - streaming parser failed to prevent infinite recursion!")
	}
}

// TestStreamingParser_StackManagement validates stack-based architecture works correctly
func TestStreamingParser_StackManagement(t *testing.T) {
	markdown := `<details>
<summary>Test</summary>
Content here
</details>`

	manager := placeholder.NewManager()
	session := manager.GetSession()
	parser := NewMarkdownParser(session, manager)

	// Stack should be empty initially
	assert.True(t, parser.IsStackEmpty(), "Stack should be empty initially")
	assert.Equal(t, 0, parser.GetStackDepth(), "Stack depth should be 0 initially")

	result, err := parser.ParseMarkdownToADFNodes(strings.Split(markdown, "\n"))
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Stack should be clean after parsing
	assert.True(t, parser.IsStackEmpty(), "Stack should be empty after parsing")
	assert.Equal(t, 0, parser.GetStackDepth(), "Stack depth should be 0 after parsing")
}

// TestStreamingParser_ComplexMixedContent tests realistic content scenarios
func TestStreamingParser_ComplexMixedContent(t *testing.T) {
	complexMarkdown := `# Document Title

Some introductory text.

<details>
<summary>Expand Section 1</summary>

- List item 1
- List item 2
  - Nested item

<details>
<summary>Nested expand</summary>
Nested content with **formatting**.
</details>

More content after nested section.
</details>

## Another section

Final paragraph content.`

	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Must complete without hanging
	done := make(chan bool, 1)
	var result []adf_types.ADFNode
	var err error

	go func() {
		parser := NewMarkdownParser(session, manager)
		result, err = parser.ParseMarkdownToADFNodes(strings.Split(complexMarkdown, "\n"))
		done <- true
	}()

	select {
	case <-done:
		require.NoError(t, err)
		assert.NotEmpty(t, result, "Should parse complex mixed content")

		// Verify we got expected node types
		nodeTypes := make([]string, len(result))
		for i, node := range result {
			nodeTypes[i] = string(node.Type)
		}

		assert.Contains(t, nodeTypes, string(adf_types.NodeTypeHeading), "Should contain heading")
		assert.Contains(t, nodeTypes, string(adf_types.NodeTypeExpand), "Should contain expand element")
		assert.Contains(t, nodeTypes, string(adf_types.NodeTypeParagraph), "Should contain paragraphs")

	case <-time.After(5 * time.Second):
		t.Fatal("CRITICAL: Parser hung on complex mixed content!")
	}
}

// TestStreamingParser_MalformedContent tests error recovery
func TestStreamingParser_MalformedContent(t *testing.T) {
	malformedCases := []struct {
		name     string
		markdown string
	}{
		{
			name: "unclosed details tag",
			markdown: `<details>
<summary>Never closed</summary>
Content without closing tag`,
		},
		{
			name: "missing summary tag",
			markdown: `<details>
Content without summary
</details>`,
		},
		{
			name:     "deeply nested malformed",
			markdown: generateMalformedNestedDetails(20),
		},
	}

	for _, testCase := range malformedCases {
		t.Run(testCase.name, func(t *testing.T) {
			manager := placeholder.NewManager()
			session := manager.GetSession()

			// Must handle malformed content gracefully (no hangs)
			done := make(chan bool, 1)
			var err error

			go func() {
				parser := NewMarkdownParser(session, manager)
				_, err = parser.ParseMarkdownToADFNodes(strings.Split(testCase.markdown, "\n"))
				done <- true
			}()

			select {
			case <-done:
				// Should either succeed (by falling back to paragraph parsing) or fail gracefully
				// The important thing is NO HANGING
				t.Logf("Malformed content handled: err=%v", err)

			case <-time.After(3 * time.Second):
				t.Fatalf("CRITICAL: Parser hung on malformed content: %s", testCase.name)
			}
		})
	}
}

// generateNestedDetailsMarkdown creates test content with specified nesting depth
func generateNestedDetailsMarkdown(depth int) string {
	if depth <= 0 {
		return "Base content"
	}

	var builder strings.Builder

	// Opening tags
	for i := 0; i < depth; i++ {
		builder.WriteString(fmt.Sprintf("<details>\n<summary>Level %d</summary>\n", i+1))
	}

	builder.WriteString("Deepest content\n")

	// Closing tags
	for i := 0; i < depth; i++ {
		builder.WriteString("</details>\n")
	}

	return builder.String()
}

// generateMalformedNestedDetails creates intentionally malformed nested content
func generateMalformedNestedDetails(depth int) string {
	var builder strings.Builder

	// Create nested structure but deliberately omit some closing tags
	for i := 0; i < depth; i++ {
		builder.WriteString(fmt.Sprintf("<details>\n<summary>Level %d</summary>\n", i+1))
	}

	builder.WriteString("Content without proper closing\n")

	// Only close some of the tags (creating malformed structure)
	for i := 0; i < depth/2; i++ {
		builder.WriteString("</details>\n")
	}

	return builder.String()
}

// TestStreamingParser_PerformanceBenchmark benchmarks the new parser vs memory usage
func TestStreamingParser_PerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmark in short mode")
	}

	depths := []int{10, 25, 50, 75}

	for _, depth := range depths {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			markdown := generateNestedDetailsMarkdown(depth)
			manager := placeholder.NewManager()
			session := manager.GetSession()

			start := time.Now()
			parser := NewMarkdownParser(session, manager)
			result, err := parser.ParseMarkdownToADFNodes(strings.Split(markdown, "\n"))
			duration := time.Since(start)

			require.NoError(t, err)
			assert.NotEmpty(t, result)

			t.Logf("Depth %d: parsed in %v", depth, duration)

			// Performance should be reasonable (not exponential)
			if duration > 1*time.Second {
				t.Errorf("Performance concern: depth %d took %v", depth, duration)
			}
		})
	}
}
