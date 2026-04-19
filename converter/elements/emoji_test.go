package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmojiConverter_ValidateInput(t *testing.T) {
	ec := NewEmojiConverter()

	tests := []struct {
		name    string
		node    adf_types.ADFNode
		wantErr bool
		errMsg  string
	}{
		{
			name: "shortName only — valid per spec",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeEmoji,
				Attrs: map[string]any{"shortName": ":white_check_mark:"},
			},
			wantErr: false,
		},
		{
			name: "shortName and text — valid",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeEmoji,
				Attrs: map[string]any{
					"shortName": ":white_check_mark:",
					"text":      "✅",
				},
			},
			wantErr: false,
		},
		{
			name: "text only without shortName — invalid",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeEmoji,
				Attrs: map[string]any{"text": "✅"},
			},
			wantErr: true,
			errMsg:  "shortName",
		},
		{
			name: "neither text nor shortName — invalid",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeEmoji,
				Attrs: map[string]any{},
			},
			wantErr: true,
			errMsg:  "shortName",
		},
		{
			name: "nil attrs — invalid",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeEmoji,
			},
			wantErr: true,
			errMsg:  "attrs",
		},
		{
			name: "wrong node type — invalid",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeText,
				Attrs: map[string]any{"shortName": ":white_check_mark:"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ec.ValidateInput(tt.node)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmojiConverter_ToMarkdown(t *testing.T) {
	ec := NewEmojiConverter()
	ctx := converter.ConversionContext{Registry: converter.GetGlobalRegistry()}

	tests := []struct {
		name    string
		node    adf_types.ADFNode
		want    string
		wantErr bool
	}{
		{
			name: "text takes priority over shortName",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeEmoji,
				Attrs: map[string]any{
					"shortName": ":white_check_mark:",
					"text":      "✅",
				},
			},
			want: "✅",
		},
		{
			name: "shortName fallback when text missing",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeEmoji,
				Attrs: map[string]any{"shortName": ":white_check_mark:"},
			},
			want: ":white_check_mark:",
		},
		{
			name: "shortName fallback with id also present",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeEmoji,
				Attrs: map[string]any{
					"id":        "2705",
					"shortName": ":white_check_mark:",
				},
			},
			want: ":white_check_mark:",
		},
		{
			name: "missing shortName and text — error",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeEmoji,
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
