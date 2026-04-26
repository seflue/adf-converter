package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestTextConverter_ToMarkdown_PlainText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy:      adf.StandardMarkdown,
		RoundTripMode: true,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "Hello, World!",
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", result.Content)
	}

	if result.Strategy != adf.StandardMarkdown {
		t.Errorf("expected adf.StandardMarkdown strategy, got %v", result.Strategy)
	}

	if result.ElementsConverted != 1 {
		t.Errorf("expected 1 element converted, got %d", result.ElementsConverted)
	}
}

func TestTextConverter_ToMarkdown_BoldText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "Bold text",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeStrong},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "**Bold text**"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_ItalicText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "Italic text",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeEm},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "*Italic text*"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_CodeText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "code snippet",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeCode},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "`code snippet`"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_StrikethroughText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "deleted text",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeStrike},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "~~deleted text~~"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_UnderlineText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "underlined text",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeUnderline},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<u>underlined text</u>"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_UnderlineBoldText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "bold underlined",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeUnderline},
			{Type: adf.MarkTypeStrong},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Marks applied in order: underline wraps first, then strong wraps
	expected := "**<u>bold underlined</u>**"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_TextColor(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		marks    []adf.Mark
		expected string
	}{
		{
			name: "red text",
			text: "colored text",
			marks: []adf.Mark{
				adf.NewMark(adf.MarkTypeTextColor, map[string]any{
					"color": "#ff0000",
				}),
			},
			expected: `<span style="color: #ff0000">colored text</span>`,
		},
		{
			name: "blue text",
			text: "blue words",
			marks: []adf.Mark{
				adf.NewMark(adf.MarkTypeTextColor, map[string]any{
					"color": "#0000ff",
				}),
			},
			expected: `<span style="color: #0000ff">blue words</span>`,
		},
		{
			name:     "missing color attr falls back to plain text",
			text:     "no color",
			marks:    []adf.Mark{{Type: adf.MarkTypeTextColor}},
			expected: "no color",
		},
	}

	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := adf.Node{
				Type:  adf.NodeTypeText,
				Text:  tt.text,
				Marks: tt.marks,
			}
			result, err := tc.ToMarkdown(node, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Content)
			}
		})
	}
}

func TestTextConverter_ToMarkdown_TextColorBold(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), Strategy: adf.StandardMarkdown}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "bold red",
		Marks: []adf.Mark{
			adf.NewMark(adf.MarkTypeTextColor, map[string]any{
				"color": "#ff0000",
			}),
			{Type: adf.MarkTypeStrong},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Marks in Reihenfolge: textColor zuerst, dann strong umschließt
	expected := `**<span style="color: #ff0000">bold red</span>**`
	if result.Content != expected {
		t.Errorf("expected %q, got %q", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_LinkText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "click here",
		Marks: []adf.Mark{
			{
				Type: adf.MarkTypeLink,
				Attrs: map[string]any{
					"href": "https://example.com",
				},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "[click here](https://example.com)"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_LinkWithoutHref(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "text without link",
		Marks: []adf.Mark{
			{
				Type:  adf.MarkTypeLink,
				Attrs: map[string]any{},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fallback to plain text when href is missing
	expected := "text without link"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_MultipleMarks(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "formatted",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeStrong},
			{Type: adf.MarkTypeEm},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Marks are applied in order, each wrapping the previous result:
	// Start: "formatted"
	// After strong: "**formatted**"
	// After em: "*formatted*"
	expected := "***formatted***"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_UnsupportedMark(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "text with unknown mark",
		Marks: []adf.Mark{
			{Type: "unknownMarkType"},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unsupported marks should return text as-is
	expected := "text with unknown mark"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_SubscriptText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "2",
		Marks: []adf.Mark{
			{
				Type: adf.MarkTypeSubsup,
				Attrs: map[string]any{
					"type": "sub",
				},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<sub>2</sub>"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_SubsupDefaultsToSub(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "text",
		Marks: []adf.Mark{
			{Type: adf.MarkTypeSubsup},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<sub>text</sub>"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_ToMarkdown_SuperscriptText(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "2",
		Marks: []adf.Mark{
			{
				Type: adf.MarkTypeSubsup,
				Attrs: map[string]any{
					"type": "sup",
				},
			},
		},
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<sup>2</sup>"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestTextConverter_FromMarkdown_ReturnsError(t *testing.T) {
	tc := NewTextRenderer()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	lines := []string{"Some **bold** text"}
	_, _, err := tc.FromMarkdown(lines, 0, ctx)

	if err == nil {
		t.Error("expected error for FromMarkdown, got nil")
	}

	expectedMsg := "text nodes are inline elements"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestTextConverter_CanHandle(t *testing.T) {
	tc := NewTextRenderer()

	if !tc.CanHandle(adf.NodeTypeText) {
		t.Error("textRenderer should handle NodeTypeText")
	}

	if tc.CanHandle(adf.NodeTypeParagraph) {
		t.Error("textRenderer should not handle NodeTypeParagraph")
	}
}

func TestTextConverter_GetStrategy(t *testing.T) {
	tc := NewTextRenderer()
	strategy := tc.GetStrategy()

	if strategy != adf.StandardMarkdown {
		t.Errorf("expected adf.StandardMarkdown strategy, got %v", strategy)
	}
}

func TestTextConverter_ValidateInput_Valid(t *testing.T) {
	tc := NewTextRenderer()

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "Valid text",
	}

	err := tc.ValidateInput(node)
	if err != nil {
		t.Errorf("unexpected error for valid node: %v", err)
	}
}

func TestTextConverter_ValidateInput_InvalidType(t *testing.T) {
	tc := NewTextRenderer()

	err := tc.ValidateInput("not a node")
	if err == nil {
		t.Error("expected error for invalid input type, got nil")
	}
}

func TestTextConverter_ValidateInput_WrongNodeType(t *testing.T) {
	tc := NewTextRenderer()

	node := adf.Node{
		Type: adf.NodeTypeParagraph,
		Text: "text",
	}

	err := tc.ValidateInput(node)
	if err == nil {
		t.Error("expected error for wrong node type, got nil")
	}
}

func TestTextConverter_ValidateInput_EmptyText(t *testing.T) {
	tc := NewTextRenderer()

	node := adf.Node{
		Type: adf.NodeTypeText,
		Text: "",
	}

	err := tc.ValidateInput(node)
	if err == nil {
		t.Error("expected error for empty text, got nil")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
