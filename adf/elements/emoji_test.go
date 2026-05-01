package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmojiConverter_ToMarkdown(t *testing.T) {
	ec := NewEmojiRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name    string
		node    adf.Node
		want    string
		wantErr bool
	}{
		{
			name: "text takes priority over shortName",
			node: adf.Node{
				Type: adf.NodeTypeEmoji,
				Attrs: map[string]any{
					"shortName": ":white_check_mark:",
					"text":      "✅",
				},
			},
			want: "✅",
		},
		{
			name: "shortName fallback when text missing",
			node: adf.Node{
				Type:  adf.NodeTypeEmoji,
				Attrs: map[string]any{"shortName": ":white_check_mark:"},
			},
			want: ":white_check_mark:",
		},
		{
			name: "shortName fallback with id also present",
			node: adf.Node{
				Type: adf.NodeTypeEmoji,
				Attrs: map[string]any{
					"id":        "2705",
					"shortName": ":white_check_mark:",
				},
			},
			want: ":white_check_mark:",
		},
		{
			name: "missing shortName and text — error",
			node: adf.Node{
				Type:  adf.NodeTypeEmoji,
				Attrs: map[string]any{"id": "2705"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ec.ToMarkdown(tt.node, ctx)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result.Content)
			}
		})
	}
}
