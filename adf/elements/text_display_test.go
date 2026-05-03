package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestTextDisplayRenderer_TextColor(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    string
		expected string
	}{
		{
			name:     "drops mark, keeps text",
			text:     "x",
			color:    "#ff5630",
			expected: "x",
		},
		{
			name:     "missing color attr keeps plain text",
			text:     "no color",
			color:    "",
			expected: "no color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTextDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			marks := []adf.Mark{}
			if tt.color != "" {
				marks = append(marks, adf.NewMark(adf.MarkTypeTextColor, map[string]any{"color": tt.color}))
			} else {
				marks = append(marks, adf.Mark{Type: adf.MarkTypeTextColor})
			}
			node := adf.Node{
				Type:  adf.NodeTypeText,
				Text:  tt.text,
				Marks: marks,
			}
			result, err := r.ToMarkdown(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestTextDisplayRenderer_Subsup_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		subType  string
		expected string
	}{
		{name: "sub digits", text: "2", subType: "sub", expected: "₂"},
		{name: "sub digit 0", text: "0", subType: "sub", expected: "₀"},
		{name: "sub multi-digit", text: "123", subType: "sub", expected: "₁₂₃"},
		{name: "sup digits", text: "2", subType: "sup", expected: "²"},
		{name: "sup multi-digit", text: "10", subType: "sup", expected: "¹⁰"},
		{name: "sup letter n", text: "n", subType: "sup", expected: "ⁿ"},
		{name: "sub letter i", text: "i", subType: "sub", expected: "ᵢ"},
		{name: "sup operator plus", text: "+", subType: "sup", expected: "⁺"},
		{name: "sub operator equals", text: "=", subType: "sub", expected: "₌"},
		{name: "sup default when type missing falls to sub digits", text: "2", subType: "", expected: "₂"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTextDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			attrs := map[string]any{}
			if tt.subType != "" {
				attrs["type"] = tt.subType
			}
			node := adf.Node{
				Type:  adf.NodeTypeText,
				Text:  tt.text,
				Marks: []adf.Mark{adf.NewMark(adf.MarkTypeSubsup, attrs)},
			}
			result, err := r.ToMarkdown(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestTextDisplayRenderer_Subsup_ASCIIFallback(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		subType  string
		expected string
	}{
		{
			name:     "sub single unmappable char uses _x form",
			text:     "q",
			subType:  "sub",
			expected: "_q",
		},
		{
			name:     "sub multi-char unmappable uses brace form",
			text:     "foo",
			subType:  "sub",
			expected: "_{foo}",
		},
		{
			name:     "sup single unmappable uppercase letter uses ^X form",
			text:     "Z",
			subType:  "sup",
			expected: "^Z",
		},
		{
			name:     "sup multi-char unmappable uses brace form",
			text:     "ABC",
			subType:  "sup",
			expected: "^{ABC}",
		},
		{
			name:     "sub partial-mappable falls fully to ASCII (single char unmappable)",
			text:     "qx",
			subType:  "sub",
			expected: "_{qx}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTextDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			node := adf.Node{
				Type: adf.NodeTypeText,
				Text: tt.text,
				Marks: []adf.Mark{
					adf.NewMark(adf.MarkTypeSubsup, map[string]any{"type": tt.subType}),
				},
			}
			result, err := r.ToMarkdown(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestTextDisplayRenderer_OtherMarksDelegate(t *testing.T) {
	tests := []struct {
		name     string
		marks    []adf.Mark
		text     string
		expected string
	}{
		{
			name:     "strong",
			text:     "x",
			marks:    []adf.Mark{{Type: adf.MarkTypeStrong}},
			expected: "**x**",
		},
		{
			name:     "em",
			text:     "x",
			marks:    []adf.Mark{{Type: adf.MarkTypeEm}},
			expected: "*x*",
		},
		{
			name:     "link",
			text:     "click",
			marks:    []adf.Mark{{Type: adf.MarkTypeLink, Attrs: map[string]any{"href": "https://example.com"}}},
			expected: "[click](https://example.com)",
		},
		{
			name: "strong wraps textColor (mark dropped)",
			text: "bold red",
			marks: []adf.Mark{
				adf.NewMark(adf.MarkTypeTextColor, map[string]any{"color": "#ff0000"}),
				{Type: adf.MarkTypeStrong},
			},
			expected: "**bold red**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTextDisplayRenderer()
			ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
			node := adf.Node{Type: adf.NodeTypeText, Text: tt.text, Marks: tt.marks}
			result, err := r.ToMarkdown(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestTextDisplayRenderer_PlainTextNoMarks(t *testing.T) {
	r := NewTextDisplayRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}
	node := adf.Node{Type: adf.NodeTypeText, Text: "Hello"}
	result, err := r.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "Hello" {
		t.Errorf("expected %q, got %q", "Hello", result.Content)
	}
}
