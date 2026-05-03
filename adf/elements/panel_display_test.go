package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestPanelDisplayRenderer_ToMarkdown(t *testing.T) {
	body := []adf.Node{
		{
			Type: adf.NodeTypeParagraph,
			Content: []adf.Node{
				{Type: adf.NodeTypeText, Text: "Hello"},
			},
		},
	}

	tests := []struct {
		name      string
		panelType string
		expected  string
	}{
		{
			name:      "info",
			panelType: "info",
			expected:  "> <span style=\"color: #0052CC\">ℹ️ **INFO**</span>\n>\n> Hello\n\n",
		},
		{
			name:      "warning",
			panelType: "warning",
			expected:  "> <span style=\"color: #FF991F\">⚠️ **WARNING**</span>\n>\n> Hello\n\n",
		},
		{
			name:      "error",
			panelType: "error",
			expected:  "> <span style=\"color: #DE350B\">❌ **ERROR**</span>\n>\n> Hello\n\n",
		},
		{
			name:      "success",
			panelType: "success",
			expected:  "> <span style=\"color: #00875A\">✅ **SUCCESS**</span>\n>\n> Hello\n\n",
		},
		{
			name:      "note",
			panelType: "note",
			expected:  "> <span style=\"color: #6554C0\">✍️ **NOTE**</span>\n>\n> Hello\n\n",
		},
		{
			name:      "tip",
			panelType: "tip",
			expected:  "> <span style=\"color: #FFAB00\">💡 **TIP**</span>\n>\n> Hello\n\n",
		},
		{
			name:      "unknown panel type falls back to info",
			panelType: "exotic",
			expected:  "> <span style=\"color: #0052CC\">ℹ️ **INFO**</span>\n>\n> Hello\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPanelDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			node := adf.Node{
				Type:    adf.NodeTypePanel,
				Attrs:   map[string]any{"panelType": tt.panelType},
				Content: body,
			}
			result, err := r.ToMarkdown(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestPanelDisplayRenderer_MultilineBody(t *testing.T) {
	r := NewPanelDisplayRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}

	node := adf.Node{
		Type:  adf.NodeTypePanel,
		Attrs: map[string]any{"panelType": "info"},
		Content: []adf.Node{
			{Type: adf.NodeTypeParagraph, Content: []adf.Node{{Type: adf.NodeTypeText, Text: "Line A"}}},
			{Type: adf.NodeTypeParagraph, Content: []adf.Node{{Type: adf.NodeTypeText, Text: "Line B"}}},
		},
	}
	result, err := r.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "> <span style=\"color: #0052CC\">ℹ️ **INFO**</span>\n>\n> Line A\n>\n> Line B\n\n"
	if result.Content != expected {
		t.Errorf("expected %q, got %q", expected, result.Content)
	}
}
