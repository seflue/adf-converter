package inline_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/elements"
	"github.com/seflue/adf-converter/converter/elements/internal/inline"
)

func newTestRegistry() *converter.ConverterRegistry {
	r := converter.NewConverterRegistry()
	r.MustRegister("text", elements.NewTextConverter())
	return r
}

func TestRenderInlineNodes_MarkSpanning(t *testing.T) {
	ctx := converter.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name  string
		nodes []adf_types.ADFNode
		want  string
	}{
		{
			name: "bold wrapping italic nested",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "Bold mit ", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: "italic nested", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}, {Type: adf_types.MarkTypeEm}}},
				{Type: adf_types.NodeTypeText, Text: " drin", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
			},
			want: "**Bold mit *italic nested* drin**",
		},
		{
			name: "simple bold",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "just bold", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
			},
			want: "**just bold**",
		},
		{
			name: "plain text",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "plain"},
			},
			want: "plain",
		},
		{
			name: "mixed text and plain",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "before "},
				{Type: adf_types.NodeTypeText, Text: "bold", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: " after"},
			},
			want: "before **bold** after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := inline.RenderInlineNodes(tt.nodes, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
