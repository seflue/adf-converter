package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ToMarkdown Tests ---

func TestPanelConverter_ToMarkdown(t *testing.T) {

	pc := NewPanelConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
	}{
		{
			name: "info panel with text",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "info"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Hello"},
					}},
				},
			},
			expected: ":::info\nHello\n:::\n\n",
		},
		{
			name: "warning panel",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "warning"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Be careful"},
					}},
				},
			},
			expected: ":::warning\nBe careful\n:::\n\n",
		},
		{
			name: "error panel",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "error"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Something broke"},
					}},
				},
			},
			expected: ":::error\nSomething broke\n:::\n\n",
		},
		{
			name: "success panel",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "success"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "All good"},
					}},
				},
			},
			expected: ":::success\nAll good\n:::\n\n",
		},
		{
			name: "note panel",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "note"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Remember this"},
					}},
				},
			},
			expected: ":::note\nRemember this\n:::\n\n",
		},
		{
			name: "panel without panelType defaults to info",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypePanel,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Default"},
					}},
				},
			},
			expected: ":::info\nDefault\n:::\n\n",
		},
		{
			name: "panel with empty attrs defaults to info",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Default"},
					}},
				},
			},
			expected: ":::info\nDefault\n:::\n\n",
		},
		{
			name: "multi-paragraph panel",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "info"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "First paragraph"},
					}},
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Second paragraph"},
					}},
				},
			},
			expected: ":::info\nFirst paragraph\n\nSecond paragraph\n:::\n\n",
		},
		{
			name: "empty panel",
			node: adf_types.ADFNode{
				Type:    adf_types.NodeTypePanel,
				Attrs:   map[string]any{"panelType": "info"},
				Content: []adf_types.ADFNode{},
			},
			expected: ":::info\n:::\n\n",
		},
		{
			name: "panel with inline formatting",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "warning"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "This is "},
						{Type: adf_types.NodeTypeText, Text: "important", Marks: []adf_types.ADFMark{
							{Type: adf_types.MarkTypeStrong},
						}},
					}},
				},
			},
			expected: ":::warning\nThis is **important**\n:::\n\n",
		},
		{
			name: "unknown panel type passes through with warning",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "custom"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Custom content"},
					}},
				},
			},
			expected: ":::custom\nCustom content\n:::\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pc.ToMarkdown(tt.node, converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Content)
			assert.Equal(t, converter.MarkdownPanel, result.Strategy)
		})
	}
}

func TestPanelConverter_ToMarkdown_UnknownTypeWarning(t *testing.T) {

	pc := NewPanelConverter()

	node := adf_types.ADFNode{
		Type:  adf_types.NodeTypePanel,
		Attrs: map[string]any{"panelType": "custom"},
		Content: []adf_types.ADFNode{
			{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "Content"},
			}},
		},
	}

	result, err := pc.ToMarkdown(node, converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()})
	require.NoError(t, err)
	assert.Equal(t, ":::custom\nContent\n:::\n\n", result.Content)
	// Unknown type should produce a warning
	assert.NotEmpty(t, result.Warnings, "expected warning for unknown panel type")
	assert.Contains(t, result.Warnings[0], "unknown panel type")
}

// --- FromMarkdown Fenced-Div Tests ---

func TestPanelConverter_FromMarkdown_FencedDiv(t *testing.T) {

	pc := NewPanelConverter()

	tests := []struct {
		name         string
		lines        []string
		wantType     string
		wantText     string
		wantConsumed int
		wantErr      bool
	}{
		{
			name:         "simple info panel",
			lines:        []string{":::info", "Hello", ":::"},
			wantType:     "info",
			wantText:     "Hello",
			wantConsumed: 3,
		},
		{
			name:         "warning panel",
			lines:        []string{":::warning", "Be careful", ":::"},
			wantType:     "warning",
			wantText:     "Be careful",
			wantConsumed: 3,
		},
		{
			name:         "case insensitive type",
			lines:        []string{":::INFO", "Content", ":::"},
			wantType:     "info",
			wantText:     "Content",
			wantConsumed: 3,
		},
		{
			name:         "mixed case type",
			lines:        []string{":::Warning", "Content", ":::"},
			wantType:     "warning",
			wantText:     "Content",
			wantConsumed: 3,
		},
		{
			name:         "empty panel",
			lines:        []string{":::info", ":::"},
			wantType:     "info",
			wantText:     "",
			wantConsumed: 2,
		},
		{
			name:         "multi-paragraph panel",
			lines:        []string{":::info", "First", "", "Second", ":::"},
			wantType:     "info",
			wantText:     "First",
			wantConsumed: 5,
		},
		{
			name:    "unclosed fence",
			lines:   []string{":::info", "Content"},
			wantErr: true,
		},
		{
			name:         "panel followed by more content",
			lines:        []string{":::info", "Panel text", ":::", "", "After panel"},
			wantType:     "info",
			wantText:     "Panel text",
			wantConsumed: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := pc.FromMarkdown(tt.lines, 0, converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, adf_types.NodeTypePanel, node.Type)
			assert.Equal(t, tt.wantType, node.Attrs["panelType"])
			assert.Equal(t, tt.wantConsumed, consumed)

			if tt.wantText != "" {
				require.NotEmpty(t, node.Content, "expected content nodes")
				// First content node should be a paragraph with text
				firstPara := node.Content[0]
				assert.Equal(t, adf_types.NodeTypeParagraph, firstPara.Type)
				require.NotEmpty(t, firstPara.Content)
				assert.Equal(t, tt.wantText, firstPara.Content[0].Text)
			}

			if tt.wantText == "" && !tt.wantErr {
				assert.Empty(t, node.Content)
			}
		})
	}
}

// --- FromMarkdown GitHub Admonition Tests ---

func TestPanelConverter_FromMarkdown_Admonition(t *testing.T) {

	pc := NewPanelConverter()

	tests := []struct {
		name         string
		lines        []string
		wantType     string
		wantText     string
		wantConsumed int
	}{
		{
			name:         "simple info admonition",
			lines:        []string{"> [!INFO]", "> Hello"},
			wantType:     "info",
			wantText:     "Hello",
			wantConsumed: 2,
		},
		{
			name:         "warning admonition",
			lines:        []string{"> [!WARNING]", "> Be careful"},
			wantType:     "warning",
			wantText:     "Be careful",
			wantConsumed: 2,
		},
		{
			name:         "error admonition",
			lines:        []string{"> [!ERROR]", "> Something broke"},
			wantType:     "error",
			wantText:     "Something broke",
			wantConsumed: 2,
		},
		{
			name:         "success admonition",
			lines:        []string{"> [!SUCCESS]", "> All good"},
			wantType:     "success",
			wantText:     "All good",
			wantConsumed: 2,
		},
		{
			name:         "note admonition",
			lines:        []string{"> [!NOTE]", "> Remember this"},
			wantType:     "note",
			wantText:     "Remember this",
			wantConsumed: 2,
		},
		{
			name:         "tip maps to note",
			lines:        []string{"> [!TIP]", "> A helpful tip"},
			wantType:     "note",
			wantText:     "A helpful tip",
			wantConsumed: 2,
		},
		{
			name:         "multi-line admonition",
			lines:        []string{"> [!INFO]", "> First line", "> Second line"},
			wantType:     "info",
			wantText:     "First line Second line",
			wantConsumed: 3,
		},
		{
			name:         "admonition ends at non-quote line",
			lines:        []string{"> [!INFO]", "> Content", "", "After"},
			wantType:     "info",
			wantText:     "Content",
			wantConsumed: 2,
		},
		{
			name:         "lowercase type in admonition",
			lines:        []string{"> [!info]", "> Content"},
			wantType:     "info",
			wantText:     "Content",
			wantConsumed: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := pc.FromMarkdown(tt.lines, 0, converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()})
			require.NoError(t, err)
			assert.Equal(t, adf_types.NodeTypePanel, node.Type)
			assert.Equal(t, tt.wantType, node.Attrs["panelType"])
			assert.Equal(t, tt.wantConsumed, consumed)

			require.NotEmpty(t, node.Content)
			firstPara := node.Content[0]
			assert.Equal(t, adf_types.NodeTypeParagraph, firstPara.Type)
			require.NotEmpty(t, firstPara.Content)
			assert.Equal(t, tt.wantText, firstPara.Content[0].Text)
		})
	}
}

// --- Roundtrip Tests ---

func TestPanelConverter_Roundtrip_FencedDiv(t *testing.T) {

	pc := NewPanelConverter()
	ctx := converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()}

	// ADF -> Markdown -> ADF
	originalNode := adf_types.ADFNode{
		Type:  adf_types.NodeTypePanel,
		Attrs: map[string]any{"panelType": "info"},
		Content: []adf_types.ADFNode{
			{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "Round trip content"},
			}},
		},
	}

	// ToMarkdown
	mdResult, err := pc.ToMarkdown(originalNode, ctx)
	require.NoError(t, err)
	assert.Equal(t, ":::info\nRound trip content\n:::\n\n", mdResult.Content)

	// FromMarkdown
	lines := strings.Split(strings.TrimRight(mdResult.Content, "\n"), "\n")
	restoredNode, _, err := pc.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)

	// Verify restored matches original
	assert.Equal(t, originalNode.Type, restoredNode.Type)
	assert.Equal(t, originalNode.Attrs["panelType"], restoredNode.Attrs["panelType"])
	require.Len(t, restoredNode.Content, 1)
	assert.Equal(t, adf_types.NodeTypeParagraph, restoredNode.Content[0].Type)
	require.Len(t, restoredNode.Content[0].Content, 1)
	assert.Equal(t, "Round trip content", restoredNode.Content[0].Content[0].Text)
}

func TestPanelConverter_Roundtrip_AdmonitionNormalization(t *testing.T) {

	pc := NewPanelConverter()
	ctx := converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()}

	// GitHub Admonition -> ADF -> Fenced-Div (canonical normalization)
	admonitionLines := []string{"> [!INFO]", "> Admonition content"}
	node, _, err := pc.FromMarkdown(admonitionLines, 0, ctx)
	require.NoError(t, err)

	// ADF -> Markdown (should always be fenced-div)
	mdResult, err := pc.ToMarkdown(node, ctx)
	require.NoError(t, err)
	assert.Equal(t, ":::info\nAdmonition content\n:::\n\n", mdResult.Content)
	assert.NotContains(t, mdResult.Content, "> [!")
}

func TestPanelConverter_Integration_MixedDocument(t *testing.T) {

	conv := converter.NewConverter(converter.WithRegistry(newTestRegistry()))

	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: []adf_types.ADFNode{
			{Type: adf_types.NodeTypeHeading, Attrs: map[string]any{"level": 1},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Title"},
				}},
			{Type: adf_types.NodeTypePanel, Attrs: map[string]any{"panelType": "warning"},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
						{Type: adf_types.NodeTypeText, Text: "Important warning"},
					}},
				}},
			{Type: adf_types.NodeTypeParagraph, Content: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "After panel"},
			}},
		},
	}

	md, _, err := conv.ToMarkdown(doc)
	require.NoError(t, err)

	// Should contain heading, panel, and paragraph
	assert.Contains(t, md, "# Title")
	assert.Contains(t, md, ":::warning")
	assert.Contains(t, md, "Important warning")
	assert.Contains(t, md, ":::")
	assert.Contains(t, md, "After panel")
	// Should NOT contain placeholder comments
	assert.NotContains(t, md, "<!--")
}

// --- Edge Cases ---

func TestPanelConverter_FromMarkdown_FencedDiv_WithBulletList(t *testing.T) {

	pc := NewPanelConverter()

	lines := []string{":::info", "- Item 1", "- Item 2", ":::"}
	node, consumed, err := pc.FromMarkdown(lines, 0, converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()})
	require.NoError(t, err)
	assert.Equal(t, 4, consumed)
	assert.Equal(t, adf_types.NodeTypePanel, node.Type)
	assert.Equal(t, "info", node.Attrs["panelType"])

	// Should contain a bulletList node
	require.NotEmpty(t, node.Content)
	assert.Equal(t, adf_types.NodeTypeBulletList, node.Content[0].Type)
}

func TestPanelConverter_FromMarkdown_FencedDiv_WithCodeBlock(t *testing.T) {

	pc := NewPanelConverter()

	lines := []string{":::info", "```go", "fmt.Println(\"hello\")", "```", ":::"}
	node, consumed, err := pc.FromMarkdown(lines, 0, converter.ConversionContext{Registry: newTestRegistry(), ParseNested: testParseNested()})
	require.NoError(t, err)
	assert.Equal(t, 5, consumed)
	assert.Equal(t, adf_types.NodeTypePanel, node.Type)

	// Should contain a codeBlock node
	require.NotEmpty(t, node.Content)
	assert.Equal(t, adf_types.NodeTypeCodeBlock, node.Content[0].Type)
}

func TestPanelConverter_ValidateInput(t *testing.T) {
	pc := NewPanelConverter()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid panel node",
			input: adf_types.ADFNode{
				Type:  adf_types.NodeTypePanel,
				Attrs: map[string]any{"panelType": "info"},
			},
			wantErr: false,
		},
		{
			name:    "wrong node type",
			input:   adf_types.ADFNode{Type: adf_types.NodeTypeParagraph},
			wantErr: true,
		},
		{
			name:    "not an ADFNode",
			input:   "string input",
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.ValidateInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
