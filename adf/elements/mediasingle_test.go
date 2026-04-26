package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/placeholder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// externalMediaNode builds a mediaSingle node with a media[type=external] child.
func externalMediaNode(layout, url, alt string) adf.Node {
	attrs := map[string]any{}
	if layout != "" {
		attrs["layout"] = layout
	}
	mediaAttrs := map[string]any{
		"type": "external",
		"url":  url,
	}
	if alt != "" {
		mediaAttrs["alt"] = alt
	}
	return adf.Node{
		Type:  adf.NodeTypeMediaSingle,
		Attrs: attrs,
		Content: []adf.Node{
			{Type: adf.NodeTypeMedia, Attrs: mediaAttrs},
		},
	}
}

// internalMediaNode builds a mediaSingle node with a media[type=file] child.
func internalMediaNode() adf.Node {
	return adf.Node{
		Type:  adf.NodeTypeMediaSingle,
		Attrs: map[string]any{"layout": "center"},
		Content: []adf.Node{
			{
				Type: adf.NodeTypeMedia,
				Attrs: map[string]any{
					"type":       "file",
					"id":         "abc-123",
					"collection": "contentId-456",
				},
			},
		},
	}
}

// ============================================================================
// ToMarkdown Tests
// ============================================================================

func TestMediaSingleConverter_ToMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		node    adf.Node
		want    string
		wantErr bool
	}{
		{
			name: "external, center layout (default) - no title",
			node: externalMediaNode("center", "https://example.com/img.png", "alt text"),
			want: "![alt text](https://example.com/img.png)\n\n",
		},
		{
			name: "external, no layout attr - no title",
			node: externalMediaNode("", "https://example.com/img.png", "alt text"),
			want: "![alt text](https://example.com/img.png)\n\n",
		},
		{
			name: "external, wide layout - title set",
			node: externalMediaNode("wide", "https://example.com/img.png", "alt text"),
			want: `![alt text](https://example.com/img.png "layout:wide")` + "\n\n",
		},
		{
			name: "external, full-width layout",
			node: externalMediaNode("full-width", "https://example.com/img.png", "alt text"),
			want: `![alt text](https://example.com/img.png "layout:full-width")` + "\n\n",
		},
		{
			name: "external, align-start layout",
			node: externalMediaNode("align-start", "https://example.com/img.png", "chart"),
			want: `![chart](https://example.com/img.png "layout:align-start")` + "\n\n",
		},
		{
			name: "external, empty alt",
			node: externalMediaNode("center", "https://example.com/img.png", ""),
			want: "![](https://example.com/img.png)\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMediaSingleConverter()
			result, err := mc.ToMarkdown(tt.node, adf.ConversionContext{Registry: newTestRegistry()})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Content)
		})
	}
}

func TestMediaSingleConverter_ToMarkdown_Internal_UsesPlaceholder(t *testing.T) {
	mgr := placeholder.NewManager()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PlaceholderManager: mgr}

	mc := NewMediaSingleConverter()
	node := internalMediaNode()

	result, err := mc.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(result.Content, "<!-- ADF_PLACEHOLDER_"), "expected placeholder comment, got: %q", result.Content)
	assert.Contains(t, result.Content, "ADF_PLACEHOLDER_")
	assert.Equal(t, 1, mgr.Count())
}

func TestMediaSingleConverter_ToMarkdown_NoContent_UsesPlaceholder(t *testing.T) {
	mgr := placeholder.NewManager()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PlaceholderManager: mgr}

	mc := NewMediaSingleConverter()
	node := adf.Node{
		Type:    adf.NodeTypeMediaSingle,
		Attrs:   map[string]any{"layout": "center"},
		Content: nil,
	}

	result, err := mc.ToMarkdown(node, ctx)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(result.Content, "<!-- ADF_PLACEHOLDER_"), "expected placeholder comment, got: %q", result.Content)
}

// ============================================================================
// FromMarkdown Tests
// ============================================================================

func TestMediaSingleConverter_FromMarkdown(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantLayout   string
		wantURL      string
		wantAlt      string
		wantConsumed int
		wantErr      bool
	}{
		{
			name:         "basic external image, center layout",
			line:         "![alt text](https://example.com/img.png)",
			wantLayout:   "center",
			wantURL:      "https://example.com/img.png",
			wantAlt:      "alt text",
			wantConsumed: 1,
		},
		{
			name:         "wide layout via title",
			line:         `![alt text](https://example.com/img.png "layout:wide")`,
			wantLayout:   "wide",
			wantURL:      "https://example.com/img.png",
			wantAlt:      "alt text",
			wantConsumed: 1,
		},
		{
			name:         "full-width layout",
			line:         `![chart](https://example.com/img.png "layout:full-width")`,
			wantLayout:   "full-width",
			wantURL:      "https://example.com/img.png",
			wantAlt:      "chart",
			wantConsumed: 1,
		},
		{
			name:         "empty alt text",
			line:         "![](https://example.com/img.png)",
			wantLayout:   "center",
			wantURL:      "https://example.com/img.png",
			wantAlt:      "",
			wantConsumed: 1,
		},
		{
			name:         "url with query params",
			line:         "![screenshot](https://cdn.example.com/img.png?v=2&size=full)",
			wantLayout:   "center",
			wantURL:      "https://cdn.example.com/img.png?v=2&size=full",
			wantAlt:      "screenshot",
			wantConsumed: 1,
		},
		{
			name:    "not an image line",
			line:    "# heading",
			wantErr: true,
		},
		{
			name:    "plain text",
			line:    "hello world",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMediaSingleConverter()
			node, consumed, err := mc.FromMarkdown([]string{tt.line}, 0, adf.ConversionContext{Registry: newTestRegistry()})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantConsumed, consumed)

			// mediaSingle wrapper
			assert.Equal(t, adf.NodeTypeMediaSingle, node.Type)
			assert.Equal(t, tt.wantLayout, node.Attrs["layout"])

			// media child
			require.Len(t, node.Content, 1)
			media := node.Content[0]
			assert.Equal(t, adf.NodeTypeMedia, media.Type)
			assert.Equal(t, "external", media.Attrs["type"])
			assert.Equal(t, tt.wantURL, media.Attrs["url"])
			assert.Equal(t, tt.wantAlt, media.Attrs["alt"])
		})
	}
}

func TestMediaSingleConverter_FromMarkdown_EmptyLines(t *testing.T) {
	mc := NewMediaSingleConverter()
	_, _, err := mc.FromMarkdown([]string{}, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.Error(t, err)
}

func TestMediaSingleConverter_FromMarkdown_StartIndexOutOfBounds(t *testing.T) {
	mc := NewMediaSingleConverter()
	_, _, err := mc.FromMarkdown([]string{"![alt](url)"}, 5, adf.ConversionContext{Registry: newTestRegistry()})
	require.Error(t, err)
}

// ============================================================================
// CanParseLine Tests
// ============================================================================

func TestMediaSingleConverter_CanParseLine(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"![alt](https://example.com/img.png)", true},
		{"![](https://example.com/img.png)", true},
		{`![alt](https://example.com/img.png "layout:wide")`, true},
		{"# heading", false},
		{"> blockquote", false},
		{"- list item", false},
		{"plain text", false},
		{`<div data-adf-type="blockCard">...</div>`, false},
		{"```go", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			mc := &mediaSingleConverter{}
			assert.Equal(t, tt.want, mc.CanParseLine(tt.line))
		})
	}
}

// ============================================================================
// CanHandle / GetStrategy / ValidateInput Tests
// ============================================================================

func TestMediaSingleConverter_CanHandle(t *testing.T) {
	mc := NewMediaSingleConverter()
	assert.True(t, mc.CanHandle(adf.NodeType(adf.NodeTypeMediaSingle)))
	assert.False(t, mc.CanHandle(adf.NodeType(adf.NodeTypeMediaInline)))
	assert.False(t, mc.CanHandle(adf.NodeType("paragraph")))
}

func TestMediaSingleConverter_GetStrategy(t *testing.T) {
	mc := NewMediaSingleConverter()
	assert.Equal(t, adf.StandardMarkdown, mc.GetStrategy())
}

func TestMediaSingleConverter_ValidateInput(t *testing.T) {
	mc := NewMediaSingleConverter()

	assert.NoError(t, mc.ValidateInput(adf.Node{Type: adf.NodeTypeMediaSingle}))
	assert.Error(t, mc.ValidateInput(adf.Node{Type: "paragraph"}))
	assert.Error(t, mc.ValidateInput("not a node"))
}

// ============================================================================
// Roundtrip Tests
// ============================================================================

func TestMediaSingleConverter_Roundtrip_ExternalCenter(t *testing.T) {
	mc := NewMediaSingleConverter()
	original := externalMediaNode("center", "https://example.com/img.png", "alt text")

	// ADF → MD
	md, err := mc.ToMarkdown(original, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, "![alt text](https://example.com/img.png)\n\n", md.Content)

	// MD → ADF
	line := strings.TrimSpace(md.Content)
	restored, _, err := mc.FromMarkdown([]string{line}, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	assert.Equal(t, adf.NodeTypeMediaSingle, restored.Type)
	assert.Equal(t, "center", restored.Attrs["layout"])
	require.Len(t, restored.Content, 1)
	media := restored.Content[0]
	assert.Equal(t, "external", media.Attrs["type"])
	assert.Equal(t, "https://example.com/img.png", media.Attrs["url"])
	assert.Equal(t, "alt text", media.Attrs["alt"])
}

func TestMediaSingleConverter_Roundtrip_ExternalWide(t *testing.T) {
	mc := NewMediaSingleConverter()
	original := externalMediaNode("wide", "https://example.com/chart.png", "chart")

	// ADF → MD
	md, err := mc.ToMarkdown(original, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)
	assert.Equal(t, `![chart](https://example.com/chart.png "layout:wide")`+"\n\n", md.Content)

	// MD → ADF
	line := strings.TrimSpace(md.Content)
	restored, _, err := mc.FromMarkdown([]string{line}, 0, adf.ConversionContext{Registry: newTestRegistry()})
	require.NoError(t, err)

	assert.Equal(t, "wide", restored.Attrs["layout"])
	assert.Equal(t, "https://example.com/chart.png", restored.Content[0].Attrs["url"])
}

func TestMediaSingleConverter_Roundtrip_Internal_Placeholder(t *testing.T) {
	mgr := placeholder.NewManager()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), PlaceholderManager: mgr}

	mc := NewMediaSingleConverter()
	original := internalMediaNode()

	// ADF → MD (placeholder)
	md, err := mc.ToMarkdown(original, ctx)
	require.NoError(t, err)
	assert.Contains(t, md.Content, "ADF_PLACEHOLDER_")

	// adf.Placeholder restore
	assert.Equal(t, 1, mgr.Count())
	placeholderID, _ := placeholder.ParsePlaceholderComment(strings.TrimSpace(md.Content))
	require.NotEmpty(t, placeholderID)

	restored, err := mgr.Restore(placeholderID)
	require.NoError(t, err)
	assert.Equal(t, adf.NodeTypeMediaSingle, restored.Type)
	assert.Equal(t, "center", restored.Attrs["layout"])
	require.Len(t, restored.Content, 1)
	assert.Equal(t, "file", restored.Content[0].Attrs["type"])
	assert.Equal(t, "abc-123", restored.Content[0].Attrs["id"])
}
