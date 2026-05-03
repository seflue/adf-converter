package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestStatusDisplayRenderer_ToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		text     string
		expected string
	}{
		{"blue", "blue", "In Review", "[In Review]"},
		{"green", "green", "Done", "[Done]"},
		{"red", "red", "Blocked", "[Blocked]"},
		{"yellow", "yellow", "Pending", "[Pending]"},
		{"purple", "purple", "Custom", "[Custom]"},
		{"neutral", "neutral", "Open", "[Open]"},
		{"unknown color falls back to bracket form", "exotic", "Mystery", "[Mystery]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewStatusDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			node := adf.Node{
				Type:  adf.NodeTypeStatus,
				Attrs: map[string]any{"text": tt.text, "color": tt.color},
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

func TestStatusDisplayRenderer_Errors(t *testing.T) {
	r := NewStatusDisplayRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}

	t.Run("nil attrs", func(t *testing.T) {
		_, err := r.ToMarkdown(adf.Node{Type: adf.NodeTypeStatus}, ctx)
		if err == nil {
			t.Error("expected error for nil attrs")
		}
	})
	t.Run("missing text", func(t *testing.T) {
		_, err := r.ToMarkdown(adf.Node{
			Type:  adf.NodeTypeStatus,
			Attrs: map[string]any{"color": "blue"},
		}, ctx)
		if err == nil {
			t.Error("expected error for missing text")
		}
	})
}
