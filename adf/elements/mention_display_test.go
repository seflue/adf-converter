package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestMentionDisplayRenderer_ToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		node     adf.Node
		expected string
	}{
		{
			name: "plain mention",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":   "abc123",
					"text": "@john.doe",
				},
			},
			expected: "@john.doe",
		},
		{
			name: "mention without text falls back to id",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id": "abc123",
				},
			},
			expected: "@abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMentionDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			result, err := r.ToMarkdown(tt.node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestMentionDisplayRenderer_ToMarkdown_Errors(t *testing.T) {
	r := NewMentionDisplayRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}

	t.Run("nil attrs", func(t *testing.T) {
		_, err := r.ToMarkdown(adf.Node{Type: adf.NodeTypeMention}, ctx)
		if err == nil {
			t.Error("expected error for nil attrs")
		}
	})

	t.Run("missing id", func(t *testing.T) {
		_, err := r.ToMarkdown(adf.Node{
			Type:  adf.NodeTypeMention,
			Attrs: map[string]any{"text": "@x"},
		}, ctx)
		if err == nil {
			t.Error("expected error for missing id")
		}
	})
}
