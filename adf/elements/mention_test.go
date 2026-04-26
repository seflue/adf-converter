package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestMentionConverter_ToMarkdown(t *testing.T) {
	mc := NewMentionRenderer()

	tests := []struct {
		name     string
		node     adf.Node
		expected string
		wantErr  bool
	}{
		{
			name: "basic mention",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":   "abc123",
					"text": "@john.doe",
				},
			},
			expected: "[@john.doe](accountid:abc123)",
		},
		{
			name: "mention with accessLevel and userType",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":          "abc123",
					"text":        "@john.doe",
					"accessLevel": "CONTAINER",
					"userType":    "DEFAULT",
				},
			},
			expected: "[@john.doe](accountid:abc123?accessLevel=CONTAINER&userType=DEFAULT)",
		},
		{
			name: "mention with only accessLevel",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":          "abc123",
					"text":        "@john.doe",
					"accessLevel": "CONTAINER",
				},
			},
			expected: "[@john.doe](accountid:abc123?accessLevel=CONTAINER)",
		},
		{
			name: "mention with only userType",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":       "abc123",
					"text":     "@john.doe",
					"userType": "DEFAULT",
				},
			},
			expected: "[@john.doe](accountid:abc123?userType=DEFAULT)",
		},
		{
			name: "mention without text falls back to id",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id": "abc123",
				},
			},
			expected: "[@abc123](accountid:abc123)",
		},
		{
			name: "mention with spaces in id gets URL-encoded",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"id":   "Some Name",
					"text": "@Some Name",
				},
			},
			expected: "[@Some Name](accountid:Some%20Name)",
		},
		{
			name: "mention with nil attrs",
			node: adf.Node{
				Type: adf.NodeTypeMention,
			},
			wantErr: true,
		},
		{
			name: "mention with empty id",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"text": "@john",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			result, err := mc.ToMarkdown(tt.node, ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestMentionConverter_CanHandle(t *testing.T) {
	mc := NewMentionRenderer()

	if !mc.CanHandle("mention") {
		t.Error("should handle mention")
	}
	if mc.CanHandle("emoji") {
		t.Error("should not handle emoji")
	}
}

func TestMentionConverter_GetStrategy(t *testing.T) {
	mc := NewMentionRenderer()

	if mc.GetStrategy() != adf.StandardMarkdown {
		t.Errorf("expected adf.StandardMarkdown, got %v", mc.GetStrategy())
	}
}
