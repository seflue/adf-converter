package elements

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/placeholder"
)

func TestInlineCardConverter_ToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
	}{
		{
			name: "simple_inline_card_with_url",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "https://example.com/page",
				},
			},
			expected: "[https://example.com/page](https://example.com/page)",
		},
		{
			name: "inline_card_without_attributes",
			node: adf_types.ADFNode{
				Type:  adf_types.NodeTypeInlineCard,
				Attrs: nil,
			},
			expected: "[InlineCard]",
		},
		{
			name: "inline_card_with_empty_url",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "",
				},
			},
			expected: "[InlineCard]",
		},
		{
			name: "inline_card_with_jira_browse_url",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "/browse/PROJ-123",
				},
			},
			expected: "[/browse/PROJ-123](/browse/PROJ-123)",
		},
		{
			name: "inline_card_with_confluence_page_url",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "/pages/viewpage.action?pageId=12345",
				},
			},
			expected: "[/pages/viewpage.action?pageId=12345](/pages/viewpage.action?pageId=12345)",
		},
		{
			name: "inline_card_with_complex_metadata",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url":   "https://example.com/page",
					"id":    "page-id-123",
					"space": "SPACE",
					"type":  "page",
				},
			},
			expected: `<a id="page-id-123" space="SPACE" type="page">[https://example.com/page](https://example.com/page)</a>`,
		},
		{
			name: "inline_card_with_version_metadata",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url":     "https://example.com/page",
					"version": 42,
					"status":  "published",
				},
			},
			expected: `<a version="42" status="published">[https://example.com/page](https://example.com/page)</a>`,
		},
		{
			name: "inline_card_with_key_metadata",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "/browse/PROJ-456",
					"key": "PROJ-456",
				},
			},
			expected: `<a key="PROJ-456">[/browse/PROJ-456](/browse/PROJ-456)</a>`,
		},
		{
			name: "inline_card_with_localId_metadata",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url":     "https://example.com/page",
					"localId": "local-123",
				},
			},
			expected: `<a localId="local-123">[https://example.com/page](https://example.com/page)</a>`,
		},
		{
			name: "inline_card_with_all_complex_metadata",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url":     "https://example.com/page",
					"id":      "page-id",
					"space":   "SPACE",
					"type":    "page",
					"version": 1,
					"status":  "published",
					"localId": "local-id",
					"key":     "KEY",
				},
			},
			// Note: Order of attributes may vary, so we'll check for presence in test
			expected: "should contain all attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewInlineCardConverter()
			context := converter.ConversionContext{}

			result, err := conv.ToMarkdown(tt.node, context)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}

			if tt.name == "inline_card_with_all_complex_metadata" {
				// Check that all expected attributes are present
				expectedAttrs := []string{`id="page-id"`, `space="SPACE"`, `type="page"`,
					`version="1"`, `status="published"`, `localId="local-id"`, `key="KEY"`}
				for _, attr := range expectedAttrs {
					if !strings.Contains(result.Content, attr) {
						t.Errorf("ToMarkdown() result missing attribute %s\nGot: %s", attr, result.Content)
					}
				}
				// Check that it has the link structure
				if !strings.Contains(result.Content, "<a") || !strings.Contains(result.Content, "</a>") {
					t.Errorf("ToMarkdown() result missing HTML wrapper\nGot: %s", result.Content)
				}
				if !strings.Contains(result.Content, "[https://example.com/page](https://example.com/page)") {
					t.Errorf("ToMarkdown() result missing markdown link\nGot: %s", result.Content)
				}
			} else if result.Content != tt.expected {
				t.Errorf("ToMarkdown() result mismatch\nExpected: %s\nGot:      %s", tt.expected, result.Content)
			}
		})
	}
}

func TestInlineCardConverter_ToMarkdown_RoundTripFidelity(t *testing.T) {
	// Test that simple inline cards produce the [url](url) pattern for round-trip fidelity
	tests := []struct {
		name string
		node adf_types.ADFNode
		want string
	}{
		{
			name: "url_as_text_and_target",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "https://jira.example.com/browse/PROJ-123",
				},
			},
			want: "[https://jira.example.com/browse/PROJ-123](https://jira.example.com/browse/PROJ-123)",
		},
		{
			name: "confluence_url_pattern",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "/wiki/spaces/SPACE/pages/123456/Page+Title",
				},
			},
			want: "[/wiki/spaces/SPACE/pages/123456/Page+Title](/wiki/spaces/SPACE/pages/123456/Page+Title)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewInlineCardConverter()
			context := converter.ConversionContext{}

			result, err := conv.ToMarkdown(tt.node, context)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}

			if result.Content != tt.want {
				t.Errorf("ToMarkdown() round-trip pattern mismatch\nWant: %s\nGot:  %s", tt.want, result.Content)
			}
		})
	}
}

func TestInlineCardConverter_FromMarkdown(t *testing.T) {
	// FromMarkdown for inline cards should return an error since they're inline elements
	// and should be parsed within parent blocks (paragraphs, headings, etc.)
	conv := NewInlineCardConverter()
	context := converter.ConversionContext{}
	lines := []string{"[url](url)"}

	_, _, err := conv.FromMarkdown(lines, 0, context)
	if err == nil {
		t.Error("FromMarkdown() should return error for inline element")
	}

	expectedMsg := "inline element"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("FromMarkdown() error should mention inline element\nGot: %v", err)
	}
}

func TestInlineCardConverter_CanHandle(t *testing.T) {
	conv := NewInlineCardConverter()

	tests := []struct {
		nodeType converter.ADFNodeType
		expected bool
	}{
		{converter.ADFNodeType(adf_types.NodeTypeInlineCard), true},
		{converter.ADFNodeType(adf_types.NodeTypeParagraph), false},
		{converter.ADFNodeType(adf_types.NodeTypeHeading), false},
		{converter.ADFNodeType(adf_types.NodeTypeText), false},
		{converter.ADFNodeType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			result := conv.CanHandle(tt.nodeType)
			if result != tt.expected {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

func TestInlineCardConverter_GetStrategy(t *testing.T) {
	conv := NewInlineCardConverter()
	strategy := conv.GetStrategy()

	if strategy != converter.StandardMarkdown {
		t.Errorf("GetStrategy() = %v, want %v", strategy, converter.StandardMarkdown)
	}
}

func TestInlineCardConverter_ValidateInput(t *testing.T) {
	conv := NewInlineCardConverter()

	tests := []struct {
		name      string
		input     interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid_inline_card_node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": "https://example.com",
				},
			},
			wantError: false,
		},
		{
			name:      "invalid_input_type",
			input:     "not a node",
			wantError: true,
			errorMsg:  "must be an ADFNode",
		},
		{
			name: "wrong_node_type",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
			},
			wantError: true,
			errorMsg:  "must be inlineCard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conv.ValidateInput(tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateInput() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateInput() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateInput() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestInlineCardConverter_ToMarkdown_Strategy(t *testing.T) {
	// Verify that the conversion result has the correct strategy
	conv := NewInlineCardConverter()
	context := converter.ConversionContext{}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeInlineCard,
		Attrs: map[string]interface{}{
			"url": "https://example.com",
		},
	}

	result, err := conv.ToMarkdown(node, context)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	if result.Strategy != converter.StandardMarkdown {
		t.Errorf("ToMarkdown() strategy = %v, want %v", result.Strategy, converter.StandardMarkdown)
	}
}

func TestInlineCardConverter_ToMarkdown_ComplexAttributePreservation(t *testing.T) {
	// Test that complex metadata is preserved in HTML wrapper
	conv := NewInlineCardConverter()
	context := converter.ConversionContext{}

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeInlineCard,
		Attrs: map[string]interface{}{
			"url":     "https://confluence.example.com/pages/123",
			"id":      "page-123",
			"space":   "DOCS",
			"type":    "page",
			"version": 5,
		},
	}

	result, err := conv.ToMarkdown(node, context)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	// Should have HTML wrapper
	if !strings.HasPrefix(result.Content, "<a") {
		t.Error("Complex metadata should start with <a tag")
	}

	if !strings.HasSuffix(result.Content, "</a>") {
		t.Error("Complex metadata should end with </a>")
	}

	// Should preserve all complex attributes
	requiredAttrs := map[string]string{
		"id":      "page-123",
		"space":   "DOCS",
		"type":    "page",
		"version": "5",
	}

	for attrName, attrValue := range requiredAttrs {
		expectedAttr := attrName + `="` + attrValue + `"`
		if !strings.Contains(result.Content, expectedAttr) {
			t.Errorf("Result missing attribute %s=%q\nGot: %s", attrName, attrValue, result.Content)
		}
	}

	// Should still contain the markdown link
	if !strings.Contains(result.Content, "[https://confluence.example.com/pages/123](https://confluence.example.com/pages/123)") {
		t.Error("Complex metadata should contain markdown link")
	}
}

func TestInlineCardConverter_ToMarkdown_DataOnlyPlaceholder(t *testing.T) {
	ic := NewInlineCardConverter()
	mgr := placeholder.NewManager()

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeInlineCard,
		Attrs: map[string]interface{}{
			"data": map[string]interface{}{
				"@type": "Document",
				"name":  "My Document",
			},
		},
	}

	ctx := converter.ConversionContext{
		PlaceholderManager: mgr,
	}

	result, err := ic.ToMarkdown(node, ctx)
	require.NoError(t, err)

	// Must contain placeholder comment, not lossy [InlineCard]
	assert.Contains(t, result.Content, "ADF_PLACEHOLDER_")
	assert.Contains(t, result.Content, "InlineCard")
	assert.NotContains(t, result.Content, "[InlineCard]")

	// Verify the node was stored and can be restored
	session := mgr.GetSession()
	assert.Len(t, session.Preserved, 1)
}

func TestInlineCardConverter_ToMarkdown_DataOnlyFallback(t *testing.T) {
	ic := NewInlineCardConverter()

	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeInlineCard,
		Attrs: map[string]interface{}{
			"data": map[string]interface{}{
				"@type": "Document",
			},
		},
	}

	// No PlaceholderManager in context
	result, err := ic.ToMarkdown(node, converter.ConversionContext{})
	require.NoError(t, err)
	assert.Equal(t, "[InlineCard]", result.Content)
}
