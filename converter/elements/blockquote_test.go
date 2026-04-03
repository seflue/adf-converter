package elements

import (
	"testing"

	"adf-converter/adf_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockquoteConverter_FromMarkdown_Simple(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	markdown := "> This is a simple blockquote"

	node, err := converter.FromMarkdown(markdown, ctx)
	require.NoError(t, err)

	assert.Equal(t, "blockquote", node.Type)
	require.Len(t, node.Content, 1)

	paragraph := node.Content[0]
	assert.Equal(t, "paragraph", paragraph.Type)
	require.Len(t, paragraph.Content, 1)
	assert.Equal(t, "text", paragraph.Content[0].Type)
	assert.Equal(t, "This is a simple blockquote", paragraph.Content[0].Text)
}

func TestBlockquoteConverter_FromMarkdown_MultiLine(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	markdown := `> This is line one
> This is line two`

	node, err := converter.FromMarkdown(markdown, ctx)
	require.NoError(t, err)

	assert.Equal(t, "blockquote", node.Type)
	require.Len(t, node.Content, 1)

	paragraph := node.Content[0]
	assert.Equal(t, "paragraph", paragraph.Type)
	require.Len(t, paragraph.Content, 1)
	assert.Equal(t, "text", paragraph.Content[0].Type)
	assert.Equal(t, "This is line one This is line two", paragraph.Content[0].Text)
}

func TestBlockquoteConverter_FromMarkdown_MultiParagraph(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	markdown := `> First paragraph
>
> Second paragraph`

	node, err := converter.FromMarkdown(markdown, ctx)
	require.NoError(t, err)

	assert.Equal(t, "blockquote", node.Type)
	require.Len(t, node.Content, 2)

	// First paragraph
	paragraph1 := node.Content[0]
	assert.Equal(t, "paragraph", paragraph1.Type)
	require.Len(t, paragraph1.Content, 1)
	assert.Equal(t, "First paragraph", paragraph1.Content[0].Text)

	// Second paragraph
	paragraph2 := node.Content[1]
	assert.Equal(t, "paragraph", paragraph2.Type)
	require.Len(t, paragraph2.Content, 1)
	assert.Equal(t, "Second paragraph", paragraph2.Content[0].Text)
}

func TestBlockquoteConverter_FromMarkdown_XMLWrapped(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: true}

	markdown := `<blockquote localId="abc123">
> This is a blockquote with attributes
</blockquote>`

	node, err := converter.FromMarkdown(markdown, ctx)
	require.NoError(t, err)

	assert.Equal(t, "blockquote", node.Type)
	require.NotNil(t, node.Attrs)
	assert.Equal(t, "abc123", node.Attrs["localId"])

	require.Len(t, node.Content, 1)
	paragraph := node.Content[0]
	assert.Equal(t, "paragraph", paragraph.Type)
	require.Len(t, paragraph.Content, 1)
	assert.Equal(t, "This is a blockquote with attributes", paragraph.Content[0].Text)
}

func TestBlockquoteConverter_FromMarkdown_Empty(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	markdown := ""

	node, err := converter.FromMarkdown(markdown, ctx)
	require.NoError(t, err)

	assert.Equal(t, "blockquote", node.Type)
	assert.Len(t, node.Content, 0)
}

func TestBlockquoteConverter_FromMarkdown_EmptyBlockquote(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	markdown := "> "

	node, err := converter.FromMarkdown(markdown, ctx)
	require.NoError(t, err)

	assert.Equal(t, "blockquote", node.Type)
	assert.Len(t, node.Content, 0)
}

func TestBlockquoteConverter_RoundTrip_Simple(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	originalMarkdown := "> This is a simple blockquote"

	// Markdown -> ADF
	node, err := converter.FromMarkdown(originalMarkdown, ctx)
	require.NoError(t, err)

	// ADF -> Markdown
	result, err := converter.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.Equal(t, "> This is a simple blockquote", result.Content)
}

func TestBlockquoteConverter_RoundTrip_WithAttributes(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: true}

	originalMarkdown := `<blockquote localId="test123">
> This is a blockquote with attributes
</blockquote>`

	// Markdown -> ADF
	node, err := converter.FromMarkdown(originalMarkdown, ctx)
	require.NoError(t, err)
	require.NotNil(t, node.Attrs)
	assert.Equal(t, "test123", node.Attrs["localId"])

	// ADF -> Markdown
	result, err := converter.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// Should preserve attributes
	assert.Contains(t, result.Content, "<blockquote localId=\"test123\">")
	assert.Contains(t, result.Content, "> This is a blockquote with attributes")
	assert.Contains(t, result.Content, "</blockquote>")
}

func TestBlockquoteConverter_ValidateInput(t *testing.T) {
	converter := NewBlockquoteConverter()

	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{
			name: "valid ADF node",
			input: adf_types.ADFNode{
				Type: "blockquote",
			},
			expectErr: false,
		},
		{
			name:      "valid markdown string",
			input:     "> blockquote",
			expectErr: false,
		},
		{
			name: "invalid ADF node type",
			input: adf_types.ADFNode{
				Type: "paragraph",
			},
			expectErr: true,
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: true,
		},
		{
			name:      "nil input",
			input:     nil,
			expectErr: true,
		},
		{
			name:      "invalid type",
			input:     123,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.ValidateInput(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBlockquoteConverter_CanHandle(t *testing.T) {
	converter := NewBlockquoteConverter()

	assert.True(t, converter.CanHandle(NodeBlockquote))
	assert.False(t, converter.CanHandle(NodeParagraph))
	assert.False(t, converter.CanHandle(NodeHeading))
}

func TestBlockquoteConverter_GetStrategy(t *testing.T) {
	converter := NewBlockquoteConverter()

	strategy := converter.GetStrategy()
	assert.Equal(t, MarkdownBlockquote, strategy)
}

// Test helper functions

func TestCreateParagraphFromLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{
			name:     "single line",
			lines:    []string{"Hello"},
			expected: "Hello",
		},
		{
			name:     "multiple lines",
			lines:    []string{"Hello", "World"},
			expected: "Hello World",
		},
		{
			name:     "three lines",
			lines:    []string{"One", "Two", "Three"},
			expected: "One Two Three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paragraph := createParagraphFromLines(tt.lines)
			assert.Equal(t, "paragraph", paragraph.Type)
			require.Len(t, paragraph.Content, 1)
			assert.Equal(t, "text", paragraph.Content[0].Type)
			assert.Equal(t, tt.expected, paragraph.Content[0].Text)
		})
	}
}

func TestParseMarkdownBlockquote(t *testing.T) {
	tests := []struct {
		name               string
		lines              []string
		expectedParagraphs int
		expectedText       []string
	}{
		{
			name:               "single line",
			lines:              []string{"> Hello"},
			expectedParagraphs: 1,
			expectedText:       []string{"Hello"},
		},
		{
			name:               "multiple lines same paragraph",
			lines:              []string{"> Hello", "> World"},
			expectedParagraphs: 1,
			expectedText:       []string{"Hello World"},
		},
		{
			name:               "multiple paragraphs",
			lines:              []string{"> First", "> ", "> Second"},
			expectedParagraphs: 2,
			expectedText:       []string{"First", "Second"},
		},
		{
			name:               "empty line separator",
			lines:              []string{"> First", "", "> Second"},
			expectedParagraphs: 2,
			expectedText:       []string{"First", "Second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parseMarkdownBlockquote(tt.lines)
			require.NoError(t, err)

			assert.Equal(t, "blockquote", node.Type)
			assert.Len(t, node.Content, tt.expectedParagraphs)

			for i, expectedText := range tt.expectedText {
				paragraph := node.Content[i]
				assert.Equal(t, "paragraph", paragraph.Type)
				require.Len(t, paragraph.Content, 1)
				assert.Equal(t, expectedText, paragraph.Content[0].Text)
			}
		})
	}
}

func TestParseXMLBlockquote(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		expectAttrs map[string]interface{}
		expectErr   bool
	}{
		{
			name: "with localId",
			lines: []string{
				`<blockquote localId="abc123">`,
				"> Content",
				"</blockquote>",
			},
			expectAttrs: map[string]interface{}{
				"localId": "abc123",
			},
			expectErr: false,
		},
		{
			name: "with multiple attributes",
			lines: []string{
				`<blockquote localId="abc123" level="1">`,
				"> Content",
				"</blockquote>",
			},
			expectAttrs: map[string]interface{}{
				"localId": "abc123",
				"level":   1, // ParseXMLAttributes converts numeric strings to integers
			},
			expectErr: false,
		},
		{
			name: "missing closing tag",
			lines: []string{
				`<blockquote localId="abc123">`,
				"> Content",
			},
			expectErr: true,
		},
		{
			name: "missing opening tag",
			lines: []string{
				"> Content",
				"</blockquote>",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, linesConsumed, err := parseXMLBlockquote(tt.lines)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, node)
			assert.Equal(t, len(tt.lines), linesConsumed)

			if tt.expectAttrs != nil {
				require.NotNil(t, node.Attrs)
				for key, expectedValue := range tt.expectAttrs {
					assert.Equal(t, expectedValue, node.Attrs[key])
				}
			}
		})
	}
}
