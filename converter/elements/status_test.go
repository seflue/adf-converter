package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConverter_ToMarkdown(t *testing.T) {
	sc := NewStatusConverter()
	ctx := converter.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name    string
		node    adf_types.ADFNode
		want    string
		wantErr bool
	}{
		{
			name: "basic status blue",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "In Progress",
					"color": "blue",
				},
			},
			want: "[status:In Progress|blue]",
		},
		{
			name: "status green",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "Done",
					"color": "green",
				},
			},
			want: "[status:Done|green]",
		},
		{
			name: "status neutral",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "TODO",
					"color": "neutral",
				},
			},
			want: "[status:TODO|neutral]",
		},
		{
			name: "localId and style ignored",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
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
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
			},
			wantErr: true,
		},
		{
			name: "missing text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]any{
					"color": "blue",
				},
			},
			wantErr: true,
		},
		{
			name: "missing color",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
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
	sc := NewStatusConverter()
	assert.True(t, sc.CanHandle(converter.ADFNodeType(adf_types.NodeTypeStatus)))
	assert.False(t, sc.CanHandle(converter.ADFNodeType(adf_types.NodeTypeDate)))
}

func TestStatusConverter_GetStrategy(t *testing.T) {
	sc := NewStatusConverter()
	assert.Equal(t, converter.StandardMarkdown, sc.GetStrategy())
}

func TestStatusConverter_ValidateInput(t *testing.T) {
	sc := NewStatusConverter()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid status node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]any{
					"text":  "In Progress",
					"color": "blue",
				},
			},
		},
		{
			name:    "non-ADFNode input",
			input:   "not a node",
			wantErr: true,
		},
		{
			name: "wrong node type",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeDate,
				Attrs: map[string]any{
					"text":  "X",
					"color": "blue",
				},
			},
			wantErr: true,
		},
		{
			name: "missing attrs",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
			},
			wantErr: true,
		},
		{
			name: "missing text attr",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]any{
					"color": "blue",
				},
			},
			wantErr: true,
		},
		{
			name: "missing color attr",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
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
