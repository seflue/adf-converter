package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/placeholder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandConverter_CanHandle(t *testing.T) {
	ec := NewExpandRenderer()

	tests := []struct {
		name     string
		nodeType adf.NodeType
		expected bool
	}{
		{
			name:     "handles expand",
			nodeType: adf.NodeType(adf.NodeTypeExpand),
			expected: true,
		},
		{
			name:     "handles nestedExpand",
			nodeType: adf.NodeType(adf.NodeTypeNestedExpand),
			expected: true,
		},
		{
			name:     "does not handle paragraph",
			nodeType: adf.NodeType(adf.NodeTypeParagraph),
			expected: false,
		},
		{
			name:     "does not handle heading",
			nodeType: adf.NodeType(adf.NodeTypeHeading),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ec.CanHandle(tt.nodeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandConverter_GetStrategy(t *testing.T) {
	ec := NewExpandRenderer()
	strategy := ec.GetStrategy()
	assert.Equal(t, adf.StandardMarkdown, strategy)
}

func TestExpandConverter_ValidateInput(t *testing.T) {
	ec := NewExpandRenderer()

	tests := []struct {
		name      string
		input     any
		expectErr bool
	}{
		{
			name: "valid expand node",
			input: adf.Node{
				Type: adf.NodeTypeExpand,
				Attrs: map[string]any{
					"title": "Test Title",
				},
			},
			expectErr: false,
		},
		{
			name: "valid nestedExpand node",
			input: adf.Node{
				Type: adf.NodeTypeNestedExpand,
				Attrs: map[string]any{
					"title": "Test Title",
				},
			},
			expectErr: false,
		},
		{
			name: "missing title attribute",
			input: adf.Node{
				Type:  adf.NodeTypeExpand,
				Attrs: map[string]any{},
			},
			expectErr: false,
		},
		{
			name: "empty title string",
			input: adf.Node{
				Type: adf.NodeTypeExpand,
				Attrs: map[string]any{
					"title": "",
				},
			},
			expectErr: false,
		},
		{
			name: "nil attributes",
			input: adf.Node{
				Type:  adf.NodeTypeExpand,
				Attrs: nil,
			},
			expectErr: true,
		},
		{
			name: "wrong node type",
			input: adf.Node{
				Type: adf.NodeTypeParagraph,
				Attrs: map[string]any{
					"title": "Test",
				},
			},
			expectErr: true,
		},
		{
			name:      "not a Node",
			input:     "not a node",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ec.ValidateInput(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpandConverter_ToMarkdown_BasicExpand(t *testing.T) {
	// Converters already registered in TestMain
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	node := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title": "Click to expand",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeText,
						Text: "Hidden content here",
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// No data-adf-type attribute — node type is derived from structural context
	assert.Contains(t, result.Content, "<details>")
	assert.NotContains(t, result.Content, `data-adf-type`)
	assert.Contains(t, result.Content, "<summary>Click to expand</summary>")
	assert.Contains(t, result.Content, "Hidden content here")
	assert.Contains(t, result.Content, "</details>")
}

func TestExpandConverter_ToMarkdown_NestedExpand(t *testing.T) {
	// Register necessary converters
	// Converters already registered in TestMain

	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	node := adf.Node{
		Type: adf.NodeTypeNestedExpand,
		Attrs: map[string]any{
			"title": "Nested section",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeText,
						Text: "Nested content",
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// No data-adf-type attribute — node type is derived from structural context
	assert.NotContains(t, result.Content, `data-adf-type`)
	assert.Contains(t, result.Content, "<details>")
	assert.Contains(t, result.Content, "<summary>Nested section</summary>")
	assert.Contains(t, result.Content, "Nested content")
}

func TestExpandConverter_ToMarkdown_WithLocalId(t *testing.T) {
	// Register necessary converters
	// Converters already registered in TestMain

	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	node := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title":   "Section with ID",
			"localId": "my-section-123",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeText,
						Text: "Content with ID",
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// Should have id attribute
	assert.Contains(t, result.Content, `id="my-section-123"`)
}

func TestExpandConverter_ToMarkdown_EmptyTitle(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	tests := []struct {
		name  string
		attrs map[string]any
	}{
		{
			name:  "empty title string",
			attrs: map[string]any{"title": ""},
		},
		{
			name:  "title key missing",
			attrs: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := adf.Node{
				Type:  adf.NodeTypeExpand,
				Attrs: tt.attrs,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Content without title"},
						},
					},
				},
			}

			result, err := ec.ToMarkdown(node, ctx)
			require.NoError(t, err)

			assert.Contains(t, result.Content, "<details>")
			assert.NotContains(t, result.Content, "<summary>")
			assert.Contains(t, result.Content, "Content without title")
			assert.Contains(t, result.Content, "</details>")
		})
	}
}

func TestExpandConverter_FromMarkdown_BasicExpand(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details>`,
		`  <summary>Test Title</summary>`,
		`  Some content here`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf.NodeTypeExpand, node.Type)
	assert.Equal(t, "Test Title", node.Attrs["title"])
}

func TestExpandConverter_FromMarkdown_NestedExpand(t *testing.T) {
	ec := NewExpandRenderer()
	// NestedLevel > 0 signals this <details> is inside another expand → nestedExpand
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, NestedLevel: 1, ParseNested: testParseNested()}

	markdown := []string{
		`<details>`,
		`  <summary>Nested Title</summary>`,
		`  Nested content`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf.NodeTypeNestedExpand, node.Type)
	assert.Equal(t, "Nested Title", node.Attrs["title"])
}

func TestExpandConverter_FromMarkdown_WithOpenAttribute(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details open>`,
		`  <summary>Open Title</summary>`,
		`  Visible content`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	// Jira's ADF API rejects 'expanded' as an attribute — must not appear in ADF output
	_, hasExpanded := node.Attrs["expanded"]
	assert.False(t, hasExpanded, "expanded must not be in ADF attrs (Jira API rejects it)")
}

func TestExpandConverter_FromMarkdown_WithLocalId(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details id="section-123">`,
		`  <summary>Section Title</summary>`,
		`  Content with ID`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, "section-123", node.Attrs["localId"])
}

func TestExpandConverter_FromMarkdown_DefaultsToExpand(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	// Without data-adf-type attribute, should default to expand
	markdown := []string{
		`<details>`,
		`  <summary>Default Title</summary>`,
		`  Default content`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf.NodeTypeExpand, node.Type)
}

func TestExpandConverter_FromMarkdown_NotDetailsElement(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<div>Not a details element</div>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, consumed)
	assert.Equal(t, "", node.Type)
}

func TestExpandConverter_FromMarkdown_NoSummary(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details>`,
		`  Some content here`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, consumed)
	assert.Equal(t, adf.NodeTypeExpand, node.Type)
	assert.Equal(t, "", node.Attrs["title"])
	require.NotEmpty(t, node.Content)
}

func TestExpandConverter_FromMarkdown_EmptySummary(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details>`,
		`  <summary></summary>`,
		`  Some content here`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf.NodeTypeExpand, node.Type)
	assert.Equal(t, "", node.Attrs["title"])
	require.NotEmpty(t, node.Content)
}

func TestExpandConverter_FromMarkdown_MissingClosingTag(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details>`,
		`  <summary>Title</summary>`,
		`  Content without closing`,
	}

	_, _, err := ec.FromMarkdown(markdown, 0, ctx)
	assert.Error(t, err)
}

func TestExpandConverter_RoundTrip_BasicExpand(t *testing.T) {
	// Register converters
	// Converters already registered in TestMain

	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	// Create original node
	original := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title": "Round Trip Test",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeText,
						Text: "Test content",
					},
				},
			},
		},
	}

	// Convert to markdown
	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)

	// Convert back to ADF
	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	// Verify node type and title preserved
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Attrs["title"], restored.Attrs["title"])
}

func TestExpandConverter_RoundTrip_NestedExpand(t *testing.T) {
	// Register converters
	// Converters already registered in TestMain

	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	// Create original nestedExpand node
	original := adf.Node{
		Type: adf.NodeTypeNestedExpand,
		Attrs: map[string]any{
			"title": "Nested Round Trip",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeText,
						Text: "Nested test content",
					},
				},
			},
		},
	}

	// Convert to markdown
	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)

	// No data-adf-type in output — type derived from context
	assert.NotContains(t, result.Content, `data-adf-type`)

	// Convert back with NestedLevel > 0 to simulate nested context
	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	nestedCtx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, NestedLevel: 1, ParseNested: testParseNested()}
	restored, consumed, err := ec.FromMarkdown(lines, 0, nestedCtx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	// nestedExpand type derived from structural context
	assert.Equal(t, adf.NodeTypeNestedExpand, restored.Type)
	assert.Equal(t, original.Attrs["title"], restored.Attrs["title"])
}

func TestExpandConverter_RoundTrip_WithAllAttributes(t *testing.T) {
	// Register converters
	// Converters already registered in TestMain

	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	// Create node with all optional attributes
	original := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title":   "Complete Attributes",
			"localId": "test-id-456",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeText,
						Text: "Full attribute test",
					},
				},
			},
		},
	}

	// Convert to markdown
	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)

	// Verify attributes in markdown (no data-adf-type, only localId)
	assert.Contains(t, result.Content, `id="test-id-456"`)
	assert.NotContains(t, result.Content, `data-adf-type`)

	// Convert back to ADF
	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	// Verify attributes preserved
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Attrs["title"], restored.Attrs["title"])
	assert.Equal(t, original.Attrs["localId"], restored.Attrs["localId"])
}

func TestExpandConverter_RoundTrip_EmptyTitle(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	original := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title": "",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Content without title"},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, "", restored.Attrs["title"])
	require.NotEmpty(t, restored.Content)
}

func TestExpandConverter_FromMarkdown_NestedDetailsElements(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	markdown := []string{
		`<details>`,
		`  <summary>Outer</summary>`,
		`  <details>`,
		`    <summary>Inner</summary>`,
		`    Inner content`,
		`  </details>`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 7, consumed)
	assert.Equal(t, adf.NodeTypeExpand, node.Type)
	assert.Equal(t, "Outer", node.Attrs["title"])

	// Inner expand should be parsed as a child node
	require.Len(t, node.Content, 1)
	inner := node.Content[0]
	assert.Equal(t, adf.NodeTypeNestedExpand, inner.Type)
	assert.Equal(t, "Inner", inner.Attrs["title"])
}

func TestExpandConverter_FromMarkdown_WithPlaceholderContent(t *testing.T) {
	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Store a placeholder via the manager
	placeholderID, _, err := manager.Store(adf.Node{
		Type: "mediaInline",
		Attrs: map[string]any{
			"id":         "abc-123",
			"collection": "test-collection",
			"type":       "file",
		},
	})
	require.NoError(t, err)

	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(),
		Strategy:           adf.StandardMarkdown,
		PlaceholderSession: session,
		PlaceholderManager: manager,
		ParseNested:        testParseNestedWith(manager),
	}

	placeholderComment := placeholder.GeneratePlaceholderComment(placeholderID, "mediaInline")

	markdown := []string{
		`<details>`,
		`  <summary>With Media</summary>`,
		`  ` + placeholderComment,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, "With Media", node.Attrs["title"])

	// The placeholder should have been resolved into content
	require.NotEmpty(t, node.Content)
}

func TestExpandConverter_FromMarkdown_NestingDepthLimit(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy:    adf.StandardMarkdown,
		NestedLevel: 101, // Already past limit
	}

	markdown := []string{
		`<details>`,
		`  <summary>Too Deep</summary>`,
		`  content`,
		`</details>`,
	}

	_, _, err := ec.FromMarkdown(markdown, 0, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum nesting depth")
}

func TestExpandConverter_RoundTrip_WithBulletList(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	original := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title": "List content",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Item one"},
								},
							},
						},
					},
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Item two"},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "Item one")
	assert.Contains(t, result.Content, "Item two")

	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	assert.Equal(t, adf.NodeTypeExpand, restored.Type)
	require.Len(t, restored.Content, 1)
	assert.Equal(t, adf.NodeTypeBulletList, restored.Content[0].Type, "expand content must be bulletList, not downgraded to paragraph")
}

func TestExpandConverter_RoundTrip_WithHeading(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	original := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title": "Heading content",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeHeading,
				Attrs: map[string]any{
					"level": float64(2),
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Section heading"},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "Section heading")

	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	assert.Equal(t, adf.NodeTypeExpand, restored.Type)
	require.Len(t, restored.Content, 1)
	assert.Equal(t, adf.NodeTypeHeading, restored.Content[0].Type, "expand content must be heading, not downgraded to paragraph")
}

func TestExpandConverter_RoundTrip_WithCodeBlock(t *testing.T) {
	ec := NewExpandRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown, ParseNested: testParseNested()}

	original := adf.Node{
		Type: adf.NodeTypeExpand,
		Attrs: map[string]any{
			"title": "Code example",
		},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeCodeBlock,
				Attrs: map[string]any{
					"language": "go",
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "fmt.Println(\"hello\")"},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "fmt.Println")

	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	assert.Equal(t, adf.NodeTypeExpand, restored.Type)
	require.Len(t, restored.Content, 1)
	assert.Equal(t, adf.NodeTypeCodeBlock, restored.Content[0].Type, "expand content must be codeBlock, not downgraded to paragraph")
}
