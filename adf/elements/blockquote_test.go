package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockquoteConverter_FromMarkdown(t *testing.T) {
	conv := NewBlockquoteRenderer()

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		ctx              adf.ConversionContext
		expectedType     adf.NodeType
		expectedContent  int // number of content nodes
		expectedConsumed int
		expectedText     []string // expected text in each paragraph (optional)
		expectedAttrs    map[string]any
	}{
		{
			name:             "simple blockquote",
			lines:            []string{"> This is a simple blockquote"},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 1,
			expectedText:     []string{"This is a simple blockquote"},
		},
		{
			name:             "multi-line same paragraph",
			lines:            []string{"> This is line one", "> This is line two"},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 2,
			expectedText:     []string{"This is line one This is line two"},
		},
		{
			name:             "multi-paragraph",
			lines:            []string{"> First paragraph", ">", "> Second paragraph"},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
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
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 3,
			expectedText:     []string{"This is a blockquote with attributes"},
			expectedAttrs:    map[string]any{"localId": "abc123"},
		},
		{
			name:             "empty lines slice",
			lines:            []string{},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  0,
			expectedConsumed: 0,
		},
		{
			name:             "startIndex out of bounds",
			lines:            []string{"> something"},
			startIndex:       5,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  0,
			expectedConsumed: 0,
		},
		{
			name:             "empty blockquote line",
			lines:            []string{"> "},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  0,
			expectedConsumed: 1,
		},
		{
			name:             "startIndex skips prefix lines",
			lines:            []string{"ignored line", "> actual blockquote"},
			startIndex:       1,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 1,
			expectedText:     []string{"actual blockquote"},
		},
		{
			name:             "boundary: stops at non-blockquote line",
			lines:            []string{"> line one", "> line two", "not a blockquote"},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 2,
			expectedText:     []string{"line one line two"},
		},
		{
			name:             "boundary: empty line between blockquote paragraphs",
			lines:            []string{"> first", "", "> second"},
			startIndex:       0,
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
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
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 3,
			expectedText:     []string{"content"},
			expectedAttrs:    map[string]any{"localId": "test"},
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
			ctx:              adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true},
			expectedType:     "blockquote",
			expectedContent:  1,
			expectedConsumed: 3,
			expectedText:     []string{"inner"},
			expectedAttrs:    map[string]any{"localId": "skip"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := conv.FromMarkdown(tt.lines, tt.startIndex, tt.ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedType, node.Type)
			assert.Len(t, node.Content, tt.expectedContent)
			assert.Equal(t, tt.expectedConsumed, consumed)

			for i, expected := range tt.expectedText {
				paragraph := node.Content[i]
				assert.Equal(t, adf.NodeTypeParagraph, paragraph.Type)
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
	conv := NewBlockquoteRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false}

	lines := []string{"> This is a simple blockquote"}

	// Markdown -> ADF
	node, consumed, err := conv.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, consumed)

	// ADF -> Markdown
	result, err := conv.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.Equal(t, "> This is a simple blockquote\n\n", result.Content)
}

func TestBlockquoteConverter_RoundTrip_WithAttributes(t *testing.T) {
	conv := NewBlockquoteRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true}

	lines := []string{
		`<blockquote localId="test123">`,
		"> This is a blockquote with attributes",
		"</blockquote>",
	}

	// Markdown -> ADF
	node, consumed, err := conv.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	require.NotNil(t, node.Attrs)
	assert.Equal(t, "test123", node.Attrs["localId"])

	// ADF -> Markdown
	result, err := conv.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.Contains(t, result.Content, `<blockquote localId="test123">`)
	assert.Contains(t, result.Content, "> This is a blockquote with attributes")
	assert.Contains(t, result.Content, "</blockquote>")
}

func TestBlockquoteConverter_ValidateInput(t *testing.T) {
	conv := NewBlockquoteRenderer()

	tests := []struct {
		name      string
		input     any
		expectErr bool
	}{
		{
			name: "valid ADF node",
			input: adf.Node{
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
			input: adf.Node{
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
			err := conv.ValidateInput(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBlockquoteConverter_CanHandle(t *testing.T) {
	conv := NewBlockquoteRenderer()

	assert.True(t, conv.CanHandle(adf.NodeTypeBlockquote))
	assert.False(t, conv.CanHandle(adf.NodeTypeParagraph))
	assert.False(t, conv.CanHandle(adf.NodeTypeHeading))
}

func TestBlockquoteConverter_GetStrategy(t *testing.T) {
	conv := NewBlockquoteRenderer()

	strategy := conv.GetStrategy()
	assert.Equal(t, adf.MarkdownBlockquote, strategy)
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

			assert.Equal(t, adf.NodeTypeBlockquote, node.Type)
			assert.Len(t, node.Content, tt.expectedParagraphs)

			for i, expectedText := range tt.expectedText {
				paragraph := node.Content[i]
				assert.Equal(t, adf.NodeTypeParagraph, paragraph.Type)
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
		expectAttrs map[string]any
		expectErr   bool
	}{
		{
			name: "with localId",
			lines: []string{
				`<blockquote localId="abc123">`,
				"> Content",
				"</blockquote>",
			},
			expectAttrs: map[string]any{
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
			expectAttrs: map[string]any{
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

func TestBlockquoteConverter_FromMarkdown_NestedPrefix(t *testing.T) {
	bc := NewBlockquoteRenderer()

	// > > text is NOT a nested blockquote in ADF — the remaining > stays as literal text
	// Separated by empty blockquote line to create two paragraphs
	lines := []string{
		"> Outer text",
		">",
		"> > Inner text",
	}

	node, consumed, err := bc.FromMarkdown(lines, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	assert.Equal(t, adf.NodeTypeBlockquote, node.Type)

	// Two paragraphs — no nested blockquote node
	require.Len(t, node.Content, 2)
	assert.Equal(t, adf.NodeTypeParagraph, node.Content[0].Type)
	assert.Equal(t, adf.NodeTypeParagraph, node.Content[1].Type)

	// Second paragraph preserves literal > as text
	require.Len(t, node.Content[1].Content, 1)
	assert.Equal(t, "> Inner text", node.Content[1].Content[0].Text)
}

func TestBlockquoteConverter_FromMarkdown_InlineFormatting(t *testing.T) {
	bc := NewBlockquoteRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name     string
		lines    []string
		wantMark adf.MarkType // mark type: "strong", "em", "code"
		wantText string // expected text value in the marked node
	}{
		{
			name:     "bold text in blockquote",
			lines:    []string{"> **bold text**"},
			wantMark: "strong",
			wantText: "bold text",
		},
		{
			name:     "italic text in blockquote",
			lines:    []string{"> *italic text*"},
			wantMark: "em",
			wantText: "italic text",
		},
		{
			name:     "inline code in blockquote",
			lines:    []string{"> `code here`"},
			wantMark: "code",
			wantText: "code here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, _, err := bc.FromMarkdown(tt.lines, 0, ctx)
			require.NoError(t, err)
			require.Len(t, node.Content, 1)

			para := node.Content[0]
			assert.Equal(t, adf.NodeTypeParagraph, para.Type)

			// Find a text node with the expected mark
			found := false
			for _, textNode := range para.Content {
				if textNode.Type != "text" {
					continue
				}
				for _, mark := range textNode.Marks {
					if mark.Type == tt.wantMark && textNode.Text == tt.wantText {
						found = true
					}
				}
			}
			assert.True(t, found, "expected text node with mark %q and text %q", tt.wantMark, tt.wantText)
		})
	}
}

func TestParseMarkdownBlockquote_Lists(t *testing.T) {
	tests := []struct {
		name             string
		lines            []string
		expectedType     adf.NodeType // type of first content node
		expectedItems    int    // expected list item count
		expectedItemText []string
	}{
		{
			name:             "bullet list in blockquote",
			lines:            []string{"> - item1", "> - item2"},
			expectedType:     "bulletList",
			expectedItems:    2,
			expectedItemText: []string{"item1", "item2"},
		},
		{
			name:             "ordered list in blockquote",
			lines:            []string{"> 1. first", "> 2. second"},
			expectedType:     "orderedList",
			expectedItems:    2,
			expectedItemText: []string{"first", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parseMarkdownBlockquote(tt.lines)
			require.NoError(t, err)
			assert.Equal(t, adf.NodeTypeBlockquote, node.Type)
			require.Len(t, node.Content, 1)

			listNode := node.Content[0]
			assert.Equal(t, tt.expectedType, listNode.Type)
			require.Len(t, listNode.Content, tt.expectedItems)

			for i, expectedText := range tt.expectedItemText {
				item := listNode.Content[i]
				assert.Equal(t, adf.NodeTypeListItem, item.Type)
				require.NotEmpty(t, item.Content)
				para := item.Content[0]
				assert.Equal(t, adf.NodeTypeParagraph, para.Type)
				require.NotEmpty(t, para.Content)
				assert.Equal(t, expectedText, para.Content[0].Text)
			}
		})
	}
}

func TestParseMarkdownBlockquote_CodeBlock(t *testing.T) {
	lines := []string{"> ```go", "> x := 1", "> ```"}

	node, err := parseMarkdownBlockquote(lines)
	require.NoError(t, err)
	assert.Equal(t, adf.NodeTypeBlockquote, node.Type)
	require.Len(t, node.Content, 1)

	codeNode := node.Content[0]
	assert.Equal(t, adf.NodeTypeCodeBlock, codeNode.Type)
	assert.Equal(t, "go", codeNode.Attrs["language"])
	require.Len(t, codeNode.Content, 1)
	assert.Equal(t, "x := 1", codeNode.Content[0].Text)
}

func TestBlockquoteConverter_ToMarkdown_Lists(t *testing.T) {
	bc := NewBlockquoteRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name     string
		node     adf.Node
		wantLine string
	}{
		{
			name: "bullet list child",
			node: adf.Node{
				Type: "blockquote",
				Content: []adf.Node{{
					Type: "bulletList",
					Content: []adf.Node{
						{Type: "listItem", Content: []adf.Node{
							{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "item1"}}},
						}},
						{Type: "listItem", Content: []adf.Node{
							{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "item2"}}},
						}},
					},
				}},
			},
			wantLine: "> - item1\n> - item2",
		},
		{
			name: "ordered list child",
			node: adf.Node{
				Type: "blockquote",
				Content: []adf.Node{{
					Type: "orderedList",
					Content: []adf.Node{
						{Type: "listItem", Content: []adf.Node{
							{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "first"}}},
						}},
						{Type: "listItem", Content: []adf.Node{
							{Type: "paragraph", Content: []adf.Node{{Type: "text", Text: "second"}}},
						}},
					},
				}},
			},
			wantLine: "> 1. first\n> 2. second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bc.ToMarkdown(tt.node, ctx)
			require.NoError(t, err)
			assert.Contains(t, result.Content, tt.wantLine)
		})
	}
}

func TestBlockquoteConverter_ToMarkdown_CodeBlock(t *testing.T) {
	bc := NewBlockquoteRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	node := adf.Node{
		Type: "blockquote",
		Content: []adf.Node{{
			Type:  "codeBlock",
			Attrs: map[string]any{"language": "go"},
			Content: []adf.Node{
				{Type: "text", Text: "x := 1"},
			},
		}},
	}

	result, err := bc.ToMarkdown(node, ctx)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "> ```go\n> x := 1\n> ```")
}

func TestBlockquoteConverter_Roundtrip_BulletList(t *testing.T) {
	bc := NewBlockquoteRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	lines := []string{"> - item1", "> - item2"}

	node, consumed, err := bc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, consumed)

	result, err := bc.ToMarkdown(node, ctx)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "> - item1")
	assert.Contains(t, result.Content, "> - item2")
}

func TestBlockquoteConverter_shouldPreserveAttrs(t *testing.T) {
	bc := &blockquoteRenderer{}

	tests := []struct {
		name string
		ctx  adf.ConversionContext
		node adf.Node
		want bool
	}{
		{
			name: "preserve true, attrs with content",
			ctx:  adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true},
			node: adf.Node{Attrs: map[string]any{"localId": "x"}},
			want: true,
		},
		{
			name: "preserve true, attrs nil",
			ctx:  adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true},
			node: adf.Node{},
			want: false,
		},
		{
			name: "preserve true, attrs empty map",
			ctx:  adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: true},
			node: adf.Node{Attrs: map[string]any{}},
			want: false,
		},
		{
			name: "preserve false, attrs with content",
			ctx:  adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false},
			node: adf.Node{Attrs: map[string]any{"localId": "x"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bc.shouldPreserveAttrs(tt.ctx, tt.node)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Suppress unused import warning
var _ = strings.Split
