package elements

import (
	"strings"
	"testing"

	"adf-converter/adf_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockquoteConverter_FromMarkdown(t *testing.T) {
	converter := NewBlockquoteConverter()

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		ctx              ConversionContext
		expectedType     string
		expectedContent  int // number of content nodes
		expectedConsumed int
		expectedText     []string // expected text in each paragraph (optional)
		expectedAttrs    map[string]interface{}
	}{
		{
			name:             "simple blockquote",
			lines:            []string{"> This is a simple blockquote"},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 1,
			expectedText:     []string{"This is a simple blockquote"},
		},
		{
			name:             "multi-line same paragraph",
			lines:            []string{"> This is line one", "> This is line two"},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 2,
			expectedText:     []string{"This is line one This is line two"},
		},
		{
			name:             "multi-paragraph",
			lines:            []string{"> First paragraph", ">", "> Second paragraph"},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  2,
			expectedConsumed: 3,
			expectedText:     []string{"First paragraph", "Second paragraph"},
		},
		{
			name: "XML-wrapped with localId",
			lines: []string{
				`<blockquote localId="abc123">`,
				"> This is a blockquote with attributes",
				"</blockquote>",
			},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: true},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 3,
			expectedText:     []string{"This is a blockquote with attributes"},
			expectedAttrs:    map[string]interface{}{"localId": "abc123"},
		},
		{
			name:             "empty lines slice",
			lines:            []string{},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  0,
			expectedConsumed: 0,
		},
		{
			name:             "startIndex out of bounds",
			lines:            []string{"> something"},
			startIndex:       5,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  0,
			expectedConsumed: 0,
		},
		{
			name:             "empty blockquote line",
			lines:            []string{"> "},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  0,
			expectedConsumed: 1,
		},
		{
			name:             "startIndex skips prefix lines",
			lines:            []string{"ignored line", "> actual blockquote"},
			startIndex:       1,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 1,
			expectedText:     []string{"actual blockquote"},
		},
		{
			name:             "boundary: stops at non-blockquote line",
			lines:            []string{"> line one", "> line two", "not a blockquote"},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 2,
			expectedText:     []string{"line one line two"},
		},
		{
			name:             "boundary: empty line between blockquote paragraphs",
			lines:            []string{"> first", "", "> second"},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  2,
			expectedConsumed: 3,
			expectedText:     []string{"first", "second"},
		},
		{
			name: "XML-wrapped with trailing lines",
			lines: []string{
				`<blockquote localId="test">`,
				"> content",
				"</blockquote>",
				"trailing line",
				"another trailing",
			},
			startIndex:       0,
			ctx:              ConversionContext{PreserveAttrs: true},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 3,
			expectedText:     []string{"content"},
			expectedAttrs:    map[string]interface{}{"localId": "test"},
		},
		{
			name: "XML-wrapped with startIndex",
			lines: []string{
				"prefix",
				`<blockquote localId="skip">`,
				"> inner",
				"</blockquote>",
			},
			startIndex:       1,
			ctx:              ConversionContext{PreserveAttrs: true},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 3,
			expectedText:     []string{"inner"},
			expectedAttrs:    map[string]interface{}{"localId": "skip"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := converter.FromMarkdown(tt.lines, tt.startIndex, tt.ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedType, node.Type)
			assert.Len(t, node.Content, tt.expectedContent)
			assert.Equal(t, tt.expectedConsumed, consumed)

			for i, expected := range tt.expectedText {
				paragraph := node.Content[i]
				assert.Equal(t, "paragraph", paragraph.Type)
				require.NotEmpty(t, paragraph.Content)
				assert.Equal(t, expected, paragraph.Content[0].Text)
			}

			if tt.expectedAttrs != nil {
				require.NotNil(t, node.Attrs)
				for key, val := range tt.expectedAttrs {
					assert.Equal(t, val, node.Attrs[key])
				}
			}
		})
	}
}

func TestBlockquoteConverter_RoundTrip_Simple(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: false}

	lines := []string{"> This is a simple blockquote"}

	// Markdown -> ADF
	node, consumed, err := converter.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, consumed)

	// ADF -> Markdown
	result, err := converter.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.Equal(t, "> This is a simple blockquote", result.Content)
}

func TestBlockquoteConverter_RoundTrip_WithAttributes(t *testing.T) {
	converter := NewBlockquoteConverter()
	ctx := ConversionContext{PreserveAttrs: true}

	lines := []string{
		`<blockquote localId="test123">`,
		"> This is a blockquote with attributes",
		"</blockquote>",
	}

	// Markdown -> ADF
	node, consumed, err := converter.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	require.NotNil(t, node.Attrs)
	assert.Equal(t, "test123", node.Attrs["localId"])

	// ADF -> Markdown
	result, err := converter.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.Contains(t, result.Content, `<blockquote localId="test123">`)
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
				"level":   1,
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

// Suppress unused import warning
var _ = strings.Split
