package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestInlineCardDisplayRenderer_ToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		node     adf.Node
		expected string
	}{
		{
			name: "url only renders single autolink",
			node: adf.Node{
				Type: adf.NodeTypeInlineCard,
				Attrs: map[string]any{
					"url": "https://example.com/foo",
				},
			},
			expected: "<https://example.com/foo>",
		},
		{
			name: "url with complex metadata still single autolink",
			node: adf.Node{
				Type: adf.NodeTypeInlineCard,
				Attrs: map[string]any{
					"url":  "https://example.com/foo",
					"id":   "ABC-1",
					"type": "issue",
				},
			},
			expected: "<https://example.com/foo>",
		},
		{
			name:     "missing attrs falls back to placeholder text",
			node:     adf.Node{Type: adf.NodeTypeInlineCard},
			expected: "[InlineCard]",
		},
		{
			name: "data-only inlineCard without url falls back to placeholder text",
			node: adf.Node{
				Type: adf.NodeTypeInlineCard,
				Attrs: map[string]any{
					"data": map[string]any{"x": 1},
				},
			},
			expected: "[InlineCard]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewInlineCardDisplayRenderer()
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
