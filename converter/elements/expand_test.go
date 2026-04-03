package elements

import (
	"strings"
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandConverter_CanHandle(t *testing.T) {
	ec := NewExpandConverter()

	tests := []struct {
		name     string
		nodeType converter.ADFNodeType
		expected bool
	}{
		{
			name:     "handles expand",
			nodeType: converter.ADFNodeType(adf_types.NodeTypeExpand),
			expected: true,
		},
		{
			name:     "handles nestedExpand",
			nodeType: converter.ADFNodeType(adf_types.NodeTypeNestedExpand),
			expected: true,
		},
		{
			name:     "does not handle paragraph",
			nodeType: converter.ADFNodeType(adf_types.NodeTypeParagraph),
			expected: false,
		},
		{
			name:     "does not handle heading",
			nodeType: converter.ADFNodeType(adf_types.NodeTypeHeading),
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
	ec := NewExpandConverter()
	strategy := ec.GetStrategy()
	assert.Equal(t, converter.StandardMarkdown, strategy)
}

func TestExpandConverter_ValidateInput(t *testing.T) {
	ec := NewExpandConverter()

	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{
			name: "valid expand node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeExpand,
				Attrs: map[string]interface{}{
					"title": "Test Title",
				},
			},
			expectErr: false,
		},
		{
			name: "valid nestedExpand node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeNestedExpand,
				Attrs: map[string]interface{}{
					"title": "Test Title",
				},
			},
			expectErr: false,
		},
		{
			name: "missing title attribute",
			input: adf_types.ADFNode{
				Type:  adf_types.NodeTypeExpand,
				Attrs: map[string]interface{}{},
			},
			expectErr: true,
		},
		{
			name: "nil attributes",
			input: adf_types.ADFNode{
				Type:  adf_types.NodeTypeExpand,
				Attrs: nil,
			},
			expectErr: true,
		},
		{
			name: "wrong node type",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Attrs: map[string]interface{}{
					"title": "Test",
				},
			},
			expectErr: true,
		},
		{
			name:      "not an ADFNode",
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
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeExpand,
		Attrs: map[string]interface{}{
			"title": "Click to expand",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Hidden content here",
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// Should contain details tag with expand data-adf-type
	assert.Contains(t, result.Content, "<details")
	assert.Contains(t, result.Content, `data-adf-type="expand"`)
	assert.Contains(t, result.Content, "<summary>Click to expand</summary>")
	assert.Contains(t, result.Content, "Hidden content here")
	assert.Contains(t, result.Content, "</details>")
}

func TestExpandConverter_ToMarkdown_NestedExpand(t *testing.T) {
	// Register necessary converters
	// Converters already registered in TestMain

	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeNestedExpand,
		Attrs: map[string]interface{}{
			"title": "Nested section",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Nested content",
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// Should contain nestedExpand data-adf-type
	assert.Contains(t, result.Content, `data-adf-type="nestedExpand"`)
	assert.Contains(t, result.Content, "<summary>Nested section</summary>")
	assert.Contains(t, result.Content, "Nested content")
}

func TestExpandConverter_ToMarkdown_ExpandedState(t *testing.T) {
	// Register necessary converters
	// Converters already registered in TestMain

	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeExpand,
		Attrs: map[string]interface{}{
			"title":    "Open section",
			"expanded": true,
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Visible by default",
					},
				},
			},
		},
	}

	result, err := ec.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// Should have 'open' attribute
	assert.Contains(t, result.Content, "<details open")
}

func TestExpandConverter_ToMarkdown_WithLocalId(t *testing.T) {
	// Register necessary converters
	// Converters already registered in TestMain

	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeExpand,
		Attrs: map[string]interface{}{
			"title":   "Section with ID",
			"localId": "my-section-123",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
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

func TestExpandConverter_ToMarkdown_MissingTitle(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	node := adf_types.ADFNode{
		Type:  adf_types.NodeTypeExpand,
		Attrs: map[string]interface{}{},
	}

	_, err := ec.ToMarkdown(node, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required title attribute")
}

func TestExpandConverter_FromMarkdown_BasicExpand(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	markdown := []string{
		`<details data-adf-type="expand">`,
		`  <summary>Test Title</summary>`,
		`  Some content here`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf_types.NodeTypeExpand, node.Type)
	assert.Equal(t, "Test Title", node.Attrs["title"])
}

func TestExpandConverter_FromMarkdown_NestedExpand(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	markdown := []string{
		`<details data-adf-type="nestedExpand">`,
		`  <summary>Nested Title</summary>`,
		`  Nested content`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf_types.NodeTypeNestedExpand, node.Type)
	assert.Equal(t, "Nested Title", node.Attrs["title"])
}

func TestExpandConverter_FromMarkdown_WithOpenAttribute(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	markdown := []string{
		`<details open data-adf-type="expand">`,
		`  <summary>Open Title</summary>`,
		`  Visible content`,
		`</details>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.True(t, node.Attrs["expanded"].(bool))
}

func TestExpandConverter_FromMarkdown_WithLocalId(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	markdown := []string{
		`<details id="section-123" data-adf-type="expand">`,
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
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

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
	assert.Equal(t, adf_types.NodeTypeExpand, node.Type)
}

func TestExpandConverter_FromMarkdown_NotDetailsElement(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	markdown := []string{
		`<div>Not a details element</div>`,
	}

	node, consumed, err := ec.FromMarkdown(markdown, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, consumed)
	assert.Equal(t, "", node.Type)
}

func TestExpandConverter_FromMarkdown_MalformedElement(t *testing.T) {
	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	tests := []struct {
		name     string
		markdown []string
	}{
		{
			name: "missing summary",
			markdown: []string{
				`<details>`,
				`  Content without summary`,
				`</details>`,
			},
		},
		{
			name: "missing closing tag",
			markdown: []string{
				`<details>`,
				`  <summary>Title</summary>`,
				`  Content without closing`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ec.FromMarkdown(tt.markdown, 0, ctx)
			assert.Error(t, err)
		})
	}
}

func TestExpandConverter_RoundTrip_BasicExpand(t *testing.T) {
	// Register converters
	// Converters already registered in TestMain

	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	// Create original node
	original := adf_types.ADFNode{
		Type: adf_types.NodeTypeExpand,
		Attrs: map[string]interface{}{
			"title": "Round Trip Test",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
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

	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	// Create original nestedExpand node
	original := adf_types.ADFNode{
		Type: adf_types.NodeTypeNestedExpand,
		Attrs: map[string]interface{}{
			"title": "Nested Round Trip",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Nested test content",
					},
				},
			},
		},
	}

	// Convert to markdown
	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)

	// Verify nestedExpand type is preserved in markdown
	assert.Contains(t, result.Content, `data-adf-type="nestedExpand"`)

	// Convert back to ADF
	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	// CRITICAL: Verify nestedExpand type is preserved through round-trip
	assert.Equal(t, adf_types.NodeTypeNestedExpand, restored.Type)
	assert.Equal(t, original.Attrs["title"], restored.Attrs["title"])
}

func TestExpandConverter_RoundTrip_WithAllAttributes(t *testing.T) {
	// Register converters
	// Converters already registered in TestMain

	ec := NewExpandConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	// Create node with all optional attributes
	original := adf_types.ADFNode{
		Type: adf_types.NodeTypeExpand,
		Attrs: map[string]interface{}{
			"title":    "Complete Attributes",
			"expanded": true,
			"localId":  "test-id-456",
		},
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Full attribute test",
					},
				},
			},
		},
	}

	// Convert to markdown
	result, err := ec.ToMarkdown(original, ctx)
	require.NoError(t, err)

	// Verify all attributes in markdown
	assert.Contains(t, result.Content, "open")
	assert.Contains(t, result.Content, `id="test-id-456"`)
	assert.Contains(t, result.Content, `data-adf-type="expand"`)

	// Convert back to ADF
	lines := strings.Split(strings.TrimSpace(result.Content), "\n")
	restored, consumed, err := ec.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Greater(t, consumed, 0)

	// Verify all attributes preserved
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Attrs["title"], restored.Attrs["title"])
	assert.Equal(t, original.Attrs["expanded"], restored.Attrs["expanded"])
	assert.Equal(t, original.Attrs["localId"], restored.Attrs["localId"])
}
