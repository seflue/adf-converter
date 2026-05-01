package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestHardBreakConverter_ToMarkdown_BasicHardBreak(t *testing.T) {
	hc := NewHardBreakRenderer()
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
	hc := NewHardBreakRenderer()
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
	hc := NewHardBreakRenderer()
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
	hc := NewHardBreakRenderer()
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

