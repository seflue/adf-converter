package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

func TestBlockCardConverter_ToMarkdown(t *testing.T) {
	tests := []struct {
		name string
		node adf_types.ADFNode
		want string
	}{
		{
			name: "simple url",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeBlockCard,
				Attrs: map[string]any{"url": "https://example.com"},
			},
			want: "<div data-adf-type=\"blockCard\">[https://example.com](https://example.com)</div>\n\n",
		},
		{
			name: "jira ticket url",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeBlockCard,
				Attrs: map[string]any{"url": "https://jira.example.com/browse/PROJ-123"},
			},
			want: "<div data-adf-type=\"blockCard\">[https://jira.example.com/browse/PROJ-123](https://jira.example.com/browse/PROJ-123)</div>\n\n",
		},
		{
			name: "empty url",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeBlockCard,
				Attrs: map[string]any{},
			},
			want: "<div data-adf-type=\"blockCard\"></div>\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBlockCardConverter()
			result, err := bc.ToMarkdown(tt.node, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.want {
				t.Errorf("got %q, want %q", result.Content, tt.want)
			}
		})
	}
}

func TestBlockCardConverter_FromMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantURL string
	}{
		{
			name:    "link syntax",
			line:    `<div data-adf-type="blockCard">[https://example.com](https://example.com)</div>`,
			wantURL: "https://example.com",
		},
		{
			name:    "jira ticket url",
			line:    `<div data-adf-type="blockCard">[https://jira.example.com/browse/PROJ-123](https://jira.example.com/browse/PROJ-123)</div>`,
			wantURL: "https://jira.example.com/browse/PROJ-123",
		},
		{
			name:    "bare url fallback",
			line:    `<div data-adf-type="blockCard">https://example.com</div>`,
			wantURL: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBlockCardConverter()
			node, consumed, err := bc.FromMarkdown([]string{tt.line}, 0, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if consumed != 1 {
				t.Errorf("expected 1 line consumed, got %d", consumed)
			}
			if node.Type != adf_types.NodeTypeBlockCard {
				t.Errorf("expected type blockCard, got %s", node.Type)
			}
			if url, _ := node.Attrs["url"].(string); url != tt.wantURL {
				t.Errorf("expected url %q, got %q", tt.wantURL, url)
			}
		})
	}
}

func TestBlockCardConverter_RoundTrip(t *testing.T) {
	bc := NewBlockCardConverter()
	ctx := converter.ConversionContext{Registry: converter.GetGlobalRegistry()}

	original := adf_types.ADFNode{
		Type:  adf_types.NodeTypeBlockCard,
		Attrs: map[string]any{"url": "https://example.com/page"},
	}

	// ADF → MD
	md, err := bc.ToMarkdown(original, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown error: %v", err)
	}

	// MD → ADF
	lines := []string{`<div data-adf-type="blockCard">[https://example.com/page](https://example.com/page)</div>`}
	restored, _, err := bc.FromMarkdown(lines, 0, ctx)
	if err != nil {
		t.Fatalf("FromMarkdown error: %v", err)
	}

	// Verify type preserved
	if restored.Type != adf_types.NodeTypeBlockCard {
		t.Errorf("node type not preserved: got %s, want blockCard", restored.Type)
	}

	// Verify url preserved
	originalURL := original.Attrs["url"].(string)
	restoredURL, _ := restored.Attrs["url"].(string)
	if restoredURL != originalURL {
		t.Errorf("url not preserved: got %q, want %q", restoredURL, originalURL)
	}

	// Verify roundtrip markdown is stable
	md2, err := bc.ToMarkdown(restored, ctx)
	if err != nil {
		t.Fatalf("second ToMarkdown error: %v", err)
	}
	if md.Content != md2.Content {
		t.Errorf("markdown not stable:\n  first:  %q\n  second: %q", md.Content, md2.Content)
	}
}
