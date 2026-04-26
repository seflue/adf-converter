package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeBlock_ToMarkdown_WithLanguage(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type:  adf.NodeTypeCodeBlock,
		Attrs: map[string]any{"language": "go"},
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "fmt.Println()"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "```go\nfmt.Println()\n```\n\n", result.Content)
	assert.Equal(t, adf.MarkdownCodeBlock, result.Strategy)
}

func TestCodeBlock_ToMarkdown_NoLanguage(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type: adf.NodeTypeCodeBlock,
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "code"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "```\ncode\n```\n\n", result.Content)
}

func TestCodeBlock_ToMarkdown_EmptyLanguage(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type:  adf.NodeTypeCodeBlock,
		Attrs: map[string]any{"language": ""},
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "code"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "```\ncode\n```\n\n", result.Content)
}

func TestCodeBlock_ToMarkdown_EmptyContent(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type: adf.NodeTypeCodeBlock,
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "```\n\n```\n\n", result.Content)
}

func TestCodeBlock_ToMarkdown_MultilineContent(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type:  adf.NodeTypeCodeBlock,
		Attrs: map[string]any{"language": "go"},
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "func main() {\n\tfmt.Println(\"hello\")\n}"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\n", result.Content)
}

func TestCodeBlock_ToMarkdown_ContentWithBackticks(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type: adf.NodeTypeCodeBlock,
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "use ``` for code"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "````\nuse ``` for code\n````\n\n", result.Content)
}

func TestCodeBlock_ToMarkdown_ContentWithLongBacktickRun(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type: adf.NodeTypeCodeBlock,
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "````` is a long fence"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "``````\n````` is a long fence\n``````\n\n", result.Content)
}

func TestCodeBlock_ToMarkdown_WarningOnExtraAttrs(t *testing.T) {
	cb := NewCodeBlockRenderer()
	node := adf.Node{
		Type:  adf.NodeTypeCodeBlock,
		Attrs: map[string]any{"language": "go", "uniqueId": "abc-123"},
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "code"},
		},
	}

	result, err := cb.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	// Should still produce normal markdown
	assert.Equal(t, "```go\ncode\n```\n\n", result.Content)
	// Should have a warning about extra attrs
	assert.NotEmpty(t, result.Warnings)
}

// --- FromMarkdown Tests ---

func TestCodeBlock_FromMarkdown_WithLanguage(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```go", "code", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	assert.Equal(t, adf.NodeTypeCodeBlock, node.Type)
	assert.Equal(t, "go", node.Attrs["language"])
	require.Len(t, node.Content, 1)
	assert.Equal(t, "code", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_NoLanguage(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```", "code", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	assert.Equal(t, adf.NodeTypeCodeBlock, node.Type)
	assert.Nil(t, node.Attrs)
	require.Len(t, node.Content, 1)
	assert.Equal(t, "code", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_EmptyBlock(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 2, consumed)
	assert.Equal(t, adf.NodeTypeCodeBlock, node.Type)
	require.Len(t, node.Content, 1)
	assert.Equal(t, "", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_Multiline(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```go", "line1", "line2", "line3", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 5, consumed)
	require.Len(t, node.Content, 1)
	assert.Equal(t, "line1\nline2\nline3", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_DynamicFence(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"````", "use ``` for code", "````"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	require.Len(t, node.Content, 1)
	assert.Equal(t, "use ``` for code", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_LinesConsumed(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```go", "code", "```", "next paragraph"}

	_, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
}

func TestCodeBlock_FromMarkdown_SpecialLanguages(t *testing.T) {
	cb := NewCodeBlockRenderer()
	languages := []string{"c++", "c#", "objective-c", "f#"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			lines := []string{"```" + lang, "code", "```"}
			node, _, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
			require.NoError(t, err)
			assert.Equal(t, lang, node.Attrs["language"])
		})
	}
}

func TestCodeBlock_FromMarkdown_UnclosedFence(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```go", "code"}

	_, _, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	assert.Error(t, err)
}

func TestCodeBlock_FromMarkdown_StartIndex(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"some text", "```go", "code", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 1, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	assert.Equal(t, "go", node.Attrs["language"])
	assert.Equal(t, "code", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_BlankLinesInContent(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```go", "line1", "", "line3", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 5, consumed)
	require.Len(t, node.Content, 1)
	assert.Equal(t, "line1\n\nline3", node.Content[0].Text)
}

func TestCodeBlock_FromMarkdown_IndentedContent(t *testing.T) {
	cb := NewCodeBlockRenderer()
	lines := []string{"```go", "func main() {", "\tfmt.Println()", "}", "```"}

	node, consumed, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 5, consumed)
	require.Len(t, node.Content, 1)
	assert.Equal(t, "func main() {\n\tfmt.Println()\n}", node.Content[0].Text)
}

// --- Roundtrip Tests ---

func TestCodeBlock_RoundTrip_Simple(t *testing.T) {
	cb := NewCodeBlockRenderer()
	original := adf.Node{
		Type:  adf.NodeTypeCodeBlock,
		Attrs: map[string]any{"language": "go"},
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "fmt.Println(\"hello\")"},
		},
	}

	// ADF -> Markdown
	result, err := cb.ToMarkdown(original, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	// Markdown -> ADF
	lines := strings.Split(strings.TrimSuffix(result.Content, "\n\n"), "\n")
	restored, _, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Attrs["language"], restored.Attrs["language"])
	assert.Equal(t, original.Content[0].Text, restored.Content[0].Text)
}

func TestCodeBlock_RoundTrip_NoLanguage(t *testing.T) {
	cb := NewCodeBlockRenderer()
	original := adf.Node{
		Type: adf.NodeTypeCodeBlock,
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "plain code"},
		},
	}

	result, err := cb.ToMarkdown(original, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSuffix(result.Content, "\n\n"), "\n")
	restored, _, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	assert.Equal(t, original.Type, restored.Type)
	assert.Nil(t, restored.Attrs)
	assert.Equal(t, original.Content[0].Text, restored.Content[0].Text)
}

func TestCodeBlock_RoundTrip_BackticksInContent(t *testing.T) {
	cb := NewCodeBlockRenderer()
	original := adf.Node{
		Type: adf.NodeTypeCodeBlock,
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "use ``` for code blocks"},
		},
	}

	result, err := cb.ToMarkdown(original, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSuffix(result.Content, "\n\n"), "\n")
	restored, _, err := cb.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	assert.Equal(t, original.Content[0].Text, restored.Content[0].Text)
}

// --- Helper Tests ---

func TestComputeFenceLength(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"no backticks", "hello world", 3},
		{"single backtick", "use `code` here", 3},
		{"double backtick", "use ``code`` here", 3},
		{"triple backticks", "use ``` for code", 4},
		{"five backticks", "````` is a long fence", 6},
		{"backticks at end", "code```", 4},
		{"only backticks", "```", 4},
		{"empty string", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, computeFenceLength(tt.content))
		})
	}
}
