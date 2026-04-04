package elements

import (
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

func TestTextConverter_ToMarkdown_PlainText(t *testing.T) {
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy:      converter.StandardMarkdown,
		RoundTripMode: true,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "Hello, World!",
	}

	result, err := tc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got '%s'", result.Content)
	}

	if result.Strategy != converter.StandardMarkdown {
		t.Errorf("expected StandardMarkdown strategy, got %v", result.Strategy)
	}

	if result.ElementsConverted != 1 {
		t.Errorf("expected 1 element converted, got %d", result.ElementsConverted)
	}
}

func TestTextConverter_ToMarkdown_BoldText(t *testing.T) {
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "Bold text",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeStrong},
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "Italic text",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeEm},
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "code snippet",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeCode},
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "deleted text",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeStrike},
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "underlined text",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeUnderline},
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "bold underlined",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeUnderline},
			{Type: adf_types.MarkTypeStrong},
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

func TestTextConverter_ToMarkdown_LinkText(t *testing.T) {
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "click here",
		Marks: []adf_types.ADFMark{
			{
				Type: adf_types.MarkTypeLink,
				Attrs: map[string]interface{}{
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "text without link",
		Marks: []adf_types.ADFMark{
			{
				Type:  adf_types.MarkTypeLink,
				Attrs: map[string]interface{}{}, // No href
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "formatted",
		Marks: []adf_types.ADFMark{
			{Type: adf_types.MarkTypeStrong},
			{Type: adf_types.MarkTypeEm},
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
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
	}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "text with unknown mark",
		Marks: []adf_types.ADFMark{
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

func TestTextConverter_FromMarkdown_ReturnsError(t *testing.T) {
	tc := NewTextConverter()
	ctx := converter.ConversionContext{
		Strategy: converter.StandardMarkdown,
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
	tc := NewTextConverter()

	if !tc.CanHandle(adf_types.NodeTypeText) {
		t.Error("TextConverter should handle NodeTypeText")
	}

	if tc.CanHandle(adf_types.NodeTypeParagraph) {
		t.Error("TextConverter should not handle NodeTypeParagraph")
	}
}

func TestTextConverter_GetStrategy(t *testing.T) {
	tc := NewTextConverter()
	strategy := tc.GetStrategy()

	if strategy != converter.StandardMarkdown {
		t.Errorf("expected StandardMarkdown strategy, got %v", strategy)
	}
}

func TestTextConverter_ValidateInput_Valid(t *testing.T) {
	tc := NewTextConverter()

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
		Text: "Valid text",
	}

	err := tc.ValidateInput(node)
	if err != nil {
		t.Errorf("unexpected error for valid node: %v", err)
	}
}

func TestTextConverter_ValidateInput_InvalidType(t *testing.T) {
	tc := NewTextConverter()

	err := tc.ValidateInput("not a node")
	if err == nil {
		t.Error("expected error for invalid input type, got nil")
	}
}

func TestTextConverter_ValidateInput_WrongNodeType(t *testing.T) {
	tc := NewTextConverter()

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeParagraph,
		Text: "text",
	}

	err := tc.ValidateInput(node)
	if err == nil {
		t.Error("expected error for wrong node type, got nil")
	}
}

func TestTextConverter_ValidateInput_EmptyText(t *testing.T) {
	tc := NewTextConverter()

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeText,
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
