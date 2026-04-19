package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuleConverter_ToMarkdown(t *testing.T) {
	rc := NewRuleConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
	}{
		{
			name:     "rule node produces horizontal rule",
			node:     adf_types.ADFNode{Type: adf_types.NodeTypeRule},
			expected: "---\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rc.ToMarkdown(tt.node, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Content)
			assert.Equal(t, 1, result.ElementsConverted)
		})
	}
}

func TestRuleConverter_ValidateInput(t *testing.T) {
	rc := NewRuleConverter()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "valid rule node",
			input:   adf_types.ADFNode{Type: adf_types.NodeTypeRule},
			wantErr: false,
		},
		{
			name:    "wrong node type",
			input:   adf_types.ADFNode{Type: adf_types.NodeTypeParagraph},
			wantErr: true,
		},
		{
			name: "rule with content is invalid",
			input: adf_types.ADFNode{
				Type:    adf_types.NodeTypeRule,
				Content: []adf_types.ADFNode{{Type: adf_types.NodeTypeText}},
			},
			wantErr: true,
		},
		{
			name:    "wrong input type",
			input:   "not a node",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rc.ValidateInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRuleConverter_FromMarkdown(t *testing.T) {
	rc := NewRuleConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	tests := []struct {
		name         string
		lines        []string
		startIndex   int
		expectedType string
		consumed     int
	}{
		{
			name:         "three dashes",
			lines:        []string{"---"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		{
			name:         "three asterisks",
			lines:        []string{"***"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		{
			name:         "three underscores",
			lines:        []string{"___"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		{
			name:         "extended dashes",
			lines:        []string{"------"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		// CommonMark: spaces between chars are valid thematic breaks
		{
			name:         "dashes with spaces",
			lines:        []string{"- - -"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		{
			name:         "asterisks with spaces",
			lines:        []string{"* * *"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		{
			name:         "underscores with spaces",
			lines:        []string{"_ _ _"},
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
		// consumed=1: lines after the rule must not be consumed
		{
			name:         "rule with trailing lines",
			lines:        []string{"---", "paragraph text", "more text"},
			startIndex:   0,
			expectedType: adf_types.NodeTypeRule,
			consumed:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := rc.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, node.Type)
			assert.Equal(t, tt.consumed, consumed)
		})
	}
}

func TestRuleConverter_FromMarkdown_OutOfBounds(t *testing.T) {
	rc := NewRuleConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	tests := []struct {
		name       string
		lines      []string
		startIndex int
	}{
		{
			name:       "empty lines",
			lines:      []string{},
			startIndex: 0,
		},
		{
			name:       "startIndex past end",
			lines:      []string{"---"},
			startIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := rc.FromMarkdown(tt.lines, tt.startIndex, ctx)
			assert.Error(t, err)
		})
	}
}

func TestRuleConverter_FromMarkdown_Invalid(t *testing.T) {
	rc := NewRuleConverter()
	ctx := converter.ConversionContext{Strategy: converter.StandardMarkdown}

	tests := []struct {
		name  string
		lines []string
	}{
		{
			name:  "too short dashes",
			lines: []string{"--"},
		},
		{
			name:  "mixed characters",
			lines: []string{"-*-"},
		},
		{
			name:  "regular text",
			lines: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := rc.FromMarkdown(tt.lines, 0, ctx)
			assert.Error(t, err)
		})
	}
}

func TestRuleConverter_ADFToMarkdown_Integration(t *testing.T) {
	conv := converter.NewDefaultConverter()

	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "before"},
				},
			},
			{Type: adf_types.NodeTypeRule},
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "after"},
				},
			},
		},
	}

	md, _, err := conv.ToMarkdown(doc)
	require.NoError(t, err)

	// Should produce --- not a placeholder comment
	assert.Contains(t, md, "---")
	assert.NotContains(t, md, "<!--")
}

func TestRuleConverter_Roundtrip(t *testing.T) {
	conv := converter.NewDefaultConverter()

	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "before"},
				},
			},
			{Type: adf_types.NodeTypeRule},
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "after"},
				},
			},
		},
	}

	md, restored, err := converter.ConvertRoundTrip(conv, doc)
	require.NoError(t, err)

	assert.Contains(t, md, "---")
	assert.NotContains(t, md, "<!--")

	// MD→ADF: should have rule node
	require.Len(t, restored.Content, 3)
	assert.Equal(t, adf_types.NodeTypeRule, restored.Content[1].Type)
}

func TestRuleConverter_CanHandle(t *testing.T) {
	rc := NewRuleConverter()

	assert.True(t, rc.CanHandle(adf_types.NodeTypeRule))
	assert.False(t, rc.CanHandle(adf_types.NodeTypeParagraph))
}

func TestRuleConverter_GetStrategy(t *testing.T) {
	rc := NewRuleConverter()
	assert.Equal(t, converter.StandardMarkdown, rc.GetStrategy())
}

func TestRuleConverter_EdgeCases(t *testing.T) {
	conv := converter.NewDefaultConverter()

	tests := []struct {
		name     string
		content  []adf_types.ADFNode
		ruleIdx  int
		numNodes int
	}{
		{
			name: "rule at document start",
			content: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeRule},
				{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "after"},
				}},
			},
			ruleIdx:  0,
			numNodes: 2,
		},
		{
			name: "rule at document end",
			content: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "before"},
				}},
				{Type: adf_types.NodeTypeRule},
			},
			ruleIdx:  1,
			numNodes: 2,
		},
		{
			name: "multiple consecutive rules",
			content: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeRule},
				{Type: adf_types.NodeTypeRule},
				{Type: adf_types.NodeTypeRule},
			},
			ruleIdx:  0,
			numNodes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: tt.content,
			}

			md, restored, err := converter.ConvertRoundTrip(conv, doc)
			require.NoError(t, err)
			assert.Contains(t, md, "---")
			require.Len(t, restored.Content, tt.numNodes)
			assert.Equal(t, adf_types.NodeTypeRule, restored.Content[tt.ruleIdx].Type)
		})
	}
}
