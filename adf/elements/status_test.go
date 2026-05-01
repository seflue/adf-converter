package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConverter_ToMarkdown(t *testing.T) {
	sc := NewStatusRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name    string
		node    adf.Node
		want    string
		wantErr bool
	}{
		{
			name: "basic status blue",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "In Progress",
					"color": "blue",
				},
			},
			want: "[status:In Progress|blue]",
		},
		{
			name: "status green",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "Done",
					"color": "green",
				},
			},
			want: "[status:Done|green]",
		},
		{
			name: "status neutral",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "TODO",
					"color": "neutral",
				},
			},
			want: "[status:TODO|neutral]",
		},
		{
			name: "localId and style ignored",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"text":    "Blocked",
					"color":   "red",
					"localId": "abc-123",
					"style":   "",
				},
			},
			want: "[status:Blocked|red]",
		},
		{
			name: "nil attrs",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
			},
			wantErr: true,
		},
		{
			name: "missing text",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"color": "blue",
				},
			},
			wantErr: true,
		},
		{
			name: "missing color",
			node: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"text": "Done",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sc.ToMarkdown(tt.node, ctx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Content)
		})
	}
}

