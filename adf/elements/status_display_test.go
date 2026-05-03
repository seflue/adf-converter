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
		{"blue", "blue", "In Review", `<span style="color: #0052CC">[In Review]</span>`},
		{"green", "green", "Done", `<span style="color: #00875A">[Done]</span>`},
		{"red", "red", "Blocked", `<span style="color: #DE350B">[Blocked]</span>`},
		{"yellow", "yellow", "Pending", `<span style="color: #FF991F">[Pending]</span>`},
		{"purple", "purple", "Custom", `<span style="color: #5243AA">[Custom]</span>`},
		{"neutral", "neutral", "Open", `<span style="color: #42526E">[Open]</span>`},
		{"unknown color falls back to neutral", "exotic", "Mystery", `<span style="color: #42526E">[Mystery]</span>`},
		{"missing color falls back to neutral", "", "NoColor", `<span style="color: #42526E">[NoColor]</span>`},
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
