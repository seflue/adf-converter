package inline

import (
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

func TestRenderInlineNodes_MarkSpanning(t *testing.T) {
	ctx := converter.ConversionContext{}

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
			name: "no marks",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "plain text"},
			},
			want: "plain text",
		},
		{
			name: "adjacent different marks with trailing space",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "bold ", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: "italic", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeEm}}},
			},
			want: "**bold** *italic*",
		},
		{
			name: "adjacent different marks with leading space",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "bold", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: " italic", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeEm}}},
			},
			want: "**bold** *italic*",
		},
		{
			name: "strike wrapping em",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "deleted ", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrike}}},
				{Type: adf_types.NodeTypeText, Text: "and italic", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrike}, {Type: adf_types.MarkTypeEm}}},
			},
			want: "~~deleted *and italic*~~",
		},
		{
			name: "bold with link inside",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "text ", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: "link", Marks: []adf_types.ADFMark{
					{Type: adf_types.MarkTypeStrong},
					{Type: adf_types.MarkTypeLink, Attrs: map[string]interface{}{"href": "https://example.com"}},
				}},
			},
			want: "**text [link](https://example.com)**",
		},
		{
			name: "real jira nested bold italic",
			nodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "Bold mit ", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: "italic nested", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeEm}, {Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: " drin", Marks: []adf_types.ADFMark{{Type: adf_types.MarkTypeStrong}}},
				{Type: adf_types.NodeTypeText, Text: " — verschachtelte Marks."},
			},
			want: "**Bold mit *italic nested* drin** — verschachtelte Marks.",
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
			got, err := RenderInlineNodes(tt.nodes, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
