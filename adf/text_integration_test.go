package adf_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

// TestTextConverter_Integration verifies TextConverter works end-to-end through ToMarkdown
func TestTextConverter_Integration(t *testing.T) {
	classifier := adf.NewDefaultClassifier()
	manager := placeholder.NewManager()
	registry := defaults.NewRegistry()

	tests := []struct {
		name     string
		adf      adf.Document
		expected string
	}{
		{
			name: "plain text",
			adf: adf.Document{
				Version: 1,
				Type:    "doc",
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Hello, World!"},
						},
					},
				},
			},
			expected: "Hello, World!\n\n",
		},
		{
			name: "bold text",
			adf: adf.Document{
				Version: 1,
				Type:    "doc",
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Bold text",
								Marks: []adf.Mark{
									{Type: adf.MarkTypeStrong},
								},
							},
						},
					},
				},
			},
			expected: "**Bold text**\n\n",
		},
		{
			name: "italic text",
			adf: adf.Document{
				Version: 1,
				Type:    "doc",
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Italic text",
								Marks: []adf.Mark{
									{Type: adf.MarkTypeEm},
								},
							},
						},
					},
				},
			},
			expected: "*Italic text*\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, _, err := testToMarkdown(tt.adf, classifier, manager, registry)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if markdown != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, markdown)
			}
		})
	}
}
