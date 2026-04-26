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

func TestStatusConverter_CanHandle(t *testing.T) {
	sc := NewStatusRenderer()
	assert.True(t, sc.CanHandle(adf.NodeType(adf.NodeTypeStatus)))
	assert.False(t, sc.CanHandle(adf.NodeType(adf.NodeTypeDate)))
}

func TestStatusConverter_GetStrategy(t *testing.T) {
	sc := NewStatusRenderer()
	assert.Equal(t, adf.StandardMarkdown, sc.GetStrategy())
}

func TestStatusConverter_ValidateInput(t *testing.T) {
	sc := NewStatusRenderer()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid status node",
			input: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "In Progress",
					"color": "blue",
				},
			},
		},
		{
			name:    "non-Node input",
			input:   "not a node",
			wantErr: true,
		},
		{
			name: "wrong node type",
			input: adf.Node{
				Type: adf.NodeTypeDate,
				Attrs: map[string]any{
					"text":  "X",
					"color": "blue",
				},
			},
			wantErr: true,
		},
		{
			name: "missing attrs",
			input: adf.Node{
				Type: adf.NodeTypeStatus,
			},
			wantErr: true,
		},
		{
			name: "missing text attr",
			input: adf.Node{
				Type: adf.NodeTypeStatus,
				Attrs: map[string]any{
					"color": "blue",
				},
			},
			wantErr: true,
		},
		{
			name: "missing color attr",
			input: adf.Node{
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
			err := sc.ValidateInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
