package converter_test

import (
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/converter/elements"
	"adf-converter/placeholder"
)

// TestMain sets up the converter registry before running tests
func TestMain(m *testing.M) {
	// Register all element converters for testing
	converter.RegisterDefaultConverters(
		elements.NewTextConverter(),
		elements.NewHardBreakConverter(),
		elements.NewParagraphConverter(),
		elements.NewHeadingConverter(),
		elements.NewListItemConverter(),
		elements.NewBulletListConverter(),
		elements.NewOrderedListConverter(),
		elements.NewExpandConverter(),
		elements.NewInlineCardConverter(),
		elements.NewCodeBlockConverter(),
		elements.NewRuleConverter(),
	)

	// Run tests
	m.Run()
}

// TestTextConverter_Integration verifies TextConverter works end-to-end through ToMarkdown
func TestTextConverter_Integration(t *testing.T) {
	classifier := converter.NewDefaultClassifier()
	manager := placeholder.NewManager()

	tests := []struct {
		name     string
		adf      adf_types.ADFDocument
		expected string
	}{
		{
			name: "plain text",
			adf: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Hello, World!"},
						},
					},
				},
			},
			expected: "Hello, World!\n\n",
		},
		{
			name: "bold text",
			adf: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Bold text",
								Marks: []adf_types.ADFMark{
									{Type: adf_types.MarkTypeStrong},
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
			adf: adf_types.ADFDocument{
				Version: 1,
				Type:    "doc",
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Italic text",
								Marks: []adf_types.ADFMark{
									{Type: adf_types.MarkTypeEm},
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
			markdown, _, err := converter.ToMarkdown(tt.adf, classifier, manager)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if markdown != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, markdown)
			}
		})
	}
}
