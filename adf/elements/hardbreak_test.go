package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestHardBreakConverter_ToMarkdown_BasicHardBreak(t *testing.T) {
	hc := NewHardBreakConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy:      adf.StandardMarkdown,
		RoundTripMode: true,
	}

	node := adf.Node{
		Type: adf.NodeTypeHardBreak,
	}

	result, err := hc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "\n" {
		t.Errorf("expected newline '\\n', got '%s'", result.Content)
	}

	if result.Strategy != adf.StandardMarkdown {
		t.Errorf("expected adf.StandardMarkdown strategy, got %v", result.Strategy)
	}

	if result.ElementsConverted != 1 {
		t.Errorf("expected 1 element converted, got %d", result.ElementsConverted)
	}
}

func TestHardBreakConverter_ToMarkdown_WithRoundTripMode(t *testing.T) {
	hc := NewHardBreakConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy:      adf.StandardMarkdown,
		RoundTripMode: true,
	}

	node := adf.Node{
		Type: adf.NodeTypeHardBreak,
	}

	result, err := hc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "\n"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestHardBreakConverter_ToMarkdown_WithoutRoundTripMode(t *testing.T) {
	hc := NewHardBreakConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy:      adf.StandardMarkdown,
		RoundTripMode: false,
	}

	node := adf.Node{
		Type: adf.NodeTypeHardBreak,
	}

	result, err := hc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Hard break should always be "\n" regardless of round-trip mode
	expected := "\n"
	if result.Content != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Content)
	}
}

func TestHardBreakConverter_FromMarkdown_ReturnsError(t *testing.T) {
	hc := NewHardBreakConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry(), 
		Strategy: adf.StandardMarkdown,
	}

	lines := []string{"Some text\nwith a line break"}
	_, _, err := hc.FromMarkdown(lines, 0, ctx)

	if err == nil {
		t.Error("expected error for FromMarkdown, got nil")
	}

	expectedMsg := "hardBreak nodes are inline elements"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestHardBreakConverter_CanHandle(t *testing.T) {
	hc := NewHardBreakConverter()

	if !hc.CanHandle(adf.NodeTypeHardBreak) {
		t.Error("hardBreakConverter should handle NodeTypeHardBreak")
	}

	if hc.CanHandle(adf.NodeTypeText) {
		t.Error("hardBreakConverter should not handle NodeTypeText")
	}

	if hc.CanHandle(adf.NodeTypeParagraph) {
		t.Error("hardBreakConverter should not handle NodeTypeParagraph")
	}
}

func TestHardBreakConverter_GetStrategy(t *testing.T) {
	hc := NewHardBreakConverter()
	strategy := hc.GetStrategy()

	if strategy != adf.StandardMarkdown {
		t.Errorf("expected adf.StandardMarkdown strategy, got %v", strategy)
	}
}

func TestHardBreakConverter_ValidateInput_Valid(t *testing.T) {
	hc := NewHardBreakConverter()

	node := adf.Node{
		Type: adf.NodeTypeHardBreak,
	}

	err := hc.ValidateInput(node)
	if err != nil {
		t.Errorf("unexpected error for valid node: %v", err)
	}
}

func TestHardBreakConverter_ValidateInput_InvalidType(t *testing.T) {
	hc := NewHardBreakConverter()

	err := hc.ValidateInput("not a node")
	if err == nil {
		t.Error("expected error for invalid input type, got nil")
	}

	expectedMsg := "invalid input type"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestHardBreakConverter_ValidateInput_WrongNodeType(t *testing.T) {
	hc := NewHardBreakConverter()

	node := adf.Node{
		Type: adf.NodeTypeText,
	}

	err := hc.ValidateInput(node)
	if err == nil {
		t.Error("expected error for wrong node type, got nil")
	}

	expectedMsg := "invalid node type"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestHardBreakConverter_ValidateInput_WithContent(t *testing.T) {
	hc := NewHardBreakConverter()

	node := adf.Node{
		Type: adf.NodeTypeHardBreak,
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "should not have content"},
		},
	}

	err := hc.ValidateInput(node)
	if err == nil {
		t.Error("expected error for hardBreak with content, got nil")
	}

	expectedMsg := "should not have content"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestHardBreakConverter_ValidateInput_WithText(t *testing.T) {
	hc := NewHardBreakConverter()

	node := adf.Node{
		Type: adf.NodeTypeHardBreak,
		Text: "should not have text",
	}

	err := hc.ValidateInput(node)
	if err == nil {
		t.Error("expected error for hardBreak with text, got nil")
	}

	expectedMsg := "should not have text"
	if err != nil && !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message containing '%s', got '%s'", expectedMsg, err.Error())
	}
}
