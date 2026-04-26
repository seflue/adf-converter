package inline_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/elements"
	"github.com/seflue/adf-converter/adf/elements/internal/inline"
)

func newTestRegistry() *adf.ConverterRegistry {
	r := adf.NewConverterRegistry()
	r.MustRegister("text", elements.NewTextRenderer())
	return r
}

func TestRenderInlineNodes_MarkSpanning(t *testing.T) {
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name  string
		nodes []adf.Node
		want  string
	}{
		{
			name: "bold wrapping italic nested",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "Bold mit ", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: "italic nested", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}, {Type: adf.MarkTypeEm}}},
				{Type: adf.NodeTypeText, Text: " drin", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
			},
			want: "**Bold mit *italic nested* drin**",
		},
		{
			name: "simple bold",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "just bold", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
			},
			want: "**just bold**",
		},
		{
			name: "plain text",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "plain"},
			},
			want: "plain",
		},
		{
			name: "mixed text and plain",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "before "},
				{Type: adf.NodeTypeText, Text: "bold", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: " after"},
			},
			want: "before **bold** after",
		},
		{
			name: "adjacent different marks with trailing space",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "bold ", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: "italic", Marks: []adf.Mark{{Type: adf.MarkTypeEm}}},
			},
			want: "**bold** *italic*",
		},
		{
			name: "adjacent different marks with leading space",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "bold", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: " italic", Marks: []adf.Mark{{Type: adf.MarkTypeEm}}},
			},
			want: "**bold** *italic*",
		},
		{
			name: "strike wrapping em",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "deleted ", Marks: []adf.Mark{{Type: adf.MarkTypeStrike}}},
				{Type: adf.NodeTypeText, Text: "and italic", Marks: []adf.Mark{{Type: adf.MarkTypeStrike}, {Type: adf.MarkTypeEm}}},
			},
			want: "~~deleted *and italic*~~",
		},
		{
			name: "real jira nested bold italic",
			nodes: []adf.Node{
				{Type: adf.NodeTypeText, Text: "Bold mit ", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: "italic nested", Marks: []adf.Mark{{Type: adf.MarkTypeEm}, {Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: " drin", Marks: []adf.Mark{{Type: adf.MarkTypeStrong}}},
				{Type: adf.NodeTypeText, Text: " — verschachtelte Marks."},
			},
			want: "**Bold mit *italic nested* drin** — verschachtelte Marks.",
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
