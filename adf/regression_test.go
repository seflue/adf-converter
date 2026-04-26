package adf_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

// TestTextNodeRegression tests that text nodes are properly converted, not placeholdered
// This test uses the exact ADF structure from a real Confluence page that was failing
func TestTextNodeRegression(t *testing.T) {
	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewManager()

	// Real ADF structure from failing Confluence page
	doc := adf.Document{
		Version: 1,
		Type:    "doc",
		Content: []adf.Node{
			{
				Type: adf.NodeTypeHeading,
				Attrs: map[string]any{
					"level": 1,
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Test"},
				},
			},
			{
				Type: adf.NodeTypeHeading,
				Attrs: map[string]any{
					"level": 3,
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Heading"},
				},
			},
			{
				Type: adf.NodeTypeExpand,
				Attrs: map[string]any{
					"title": "Also expandable",
				},
				Content: []adf.Node{
					{
						Type: adf.NodeTypeHeading,
						Attrs: map[string]any{
							"level": 2,
						},
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Content"},
						},
					},
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Lorem ipsum"},
						},
					},
				},
			},
			{
				Type: adf.NodeTypeHeading,
				Attrs: map[string]any{
					"level": 1,
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Sonstiges"},
				},
			},
			{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{
										Type: adf.NodeTypeText,
										Text: "Why you should use Python and Rust together",
										Marks: []adf.Mark{
											{
												Type: adf.MarkTypeLink,
												Attrs: map[string]any{
													"href": "https://opensource.com/article/23/3/python-loves-rust",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	markdown, _, err := testToMarkdown(doc, classifier, manager, defaults.NewRegistry())
	if err != nil {
		t.Fatalf("ToMarkdown failed: %v", err)
	}

	t.Logf("Generated markdown:\n%s", markdown)

	// Text should NOT be placeholdered
	if containsPlaceholder(markdown, "Text") {
		t.Errorf("REGRESSION: Text nodes are being placeholdered!\nMarkdown:\n%s", markdown)
	}

	// Should contain actual text from headings
	expectedTexts := []string{
		"Test",
		"Heading",
		"Content",
		"Lorem ipsum",
		"Sonstiges",
		"Why you should use Python and Rust together",
	}

	for _, expected := range expectedTexts {
		if !contains(markdown, expected) {
			t.Errorf("Expected '%s' in markdown, got:\n%s", expected, markdown)
		}
	}
}

func containsPlaceholder(s, nodeType string) bool {
	placeholderPrefix := "<!-- ADF_PLACEHOLDER_"
	return contains(s, placeholderPrefix)
}

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
