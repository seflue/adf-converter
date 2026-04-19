package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

// TestMain sets up the converter registry before running tests.
// Registers ALL converters once — individual tests must NOT call Clear() or RegisterDefaultConverters().
func TestMain(m *testing.M) {
	converter.RegisterDefaultConverters(
		NewTextConverter(),
		NewHardBreakConverter(),
		NewParagraphConverter(),
		NewHeadingConverter(),
		NewListItemConverter(),
		NewBulletListConverter(),
		NewOrderedListConverter(),
		NewExpandConverter(),
		NewInlineCardConverter(),
		NewBlockCardConverter(),
		NewEmojiConverter(),
		NewCodeBlockConverter(),
		NewRuleConverter(),
		NewMentionConverter(),
		NewTableConverter(),
		NewPanelConverter(),
		NewDateConverter(),
		NewStatusConverter(),
		NewBlockquoteConverter(),
		NewTaskListConverter(),
		NewMediaSingleConverter(),
	)

	m.Run()
}

func TestParagraphConverter_ToMarkdown(t *testing.T) {
	pc := NewParagraphConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
		wantErr  bool
	}{
		{
			name: "empty paragraph",
			node: adf_types.ADFNode{
				Type:    adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{},
			},
			expected: "\n",
			wantErr:  false,
		},
		{
			name: "simple paragraph with plain text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Hello world",
					},
				},
			},
			expected: "Hello world\n\n",
			wantErr:  false,
		},
		{
			name: "paragraph with bold text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "bold text",
						Marks: []adf_types.ADFMark{
							{Type: "strong"},
						},
					},
				},
			},
			expected: "**bold text**\n\n",
			wantErr:  false,
		},
		{
			name: "paragraph with italic text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "italic text",
						Marks: []adf_types.ADFMark{
							{Type: "em"},
						},
					},
				},
			},
			expected: "*italic text*\n\n",
			wantErr:  false,
		},
		{
			name: "paragraph with code text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "code text",
						Marks: []adf_types.ADFMark{
							{Type: "code"},
						},
					},
				},
			},
			expected: "`code text`\n\n",
			wantErr:  false,
		},
		{
			name: "paragraph with multiple text nodes",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Hello ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "bold",
						Marks: []adf_types.ADFMark{
							{Type: "strong"},
						},
					},
					{
						Type: adf_types.NodeTypeText,
						Text: " world",
					},
				},
			},
			expected: "Hello **bold** world\n\n",
			wantErr:  false,
		},
		{
			name: "paragraph with hardbreak",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Line 1",
					},
					{
						Type: adf_types.NodeTypeHardBreak,
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "Line 2",
					},
				},
			},
			expected: "Line 1\nLine 2\n\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pc.ToMarkdown(tt.node, converter.ConversionContext{})
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result.Content != tt.expected {
				t.Errorf("ToMarkdown() got = %q, want %q", result.Content, tt.expected)
			}
			if result.Strategy != converter.StandardMarkdown {
				t.Errorf("ToMarkdown() strategy = %v, want %v", result.Strategy, converter.StandardMarkdown)
			}
		})
	}
}

func TestParagraphConverter_FromMarkdown(t *testing.T) {
	pc := NewParagraphConverter()

	tests := []struct {
		name          string
		lines         []string
		startIndex    int
		expectedType  string
		expectedText  string
		expectedLines int
		wantErr       bool
	}{
		{
			name:          "empty lines",
			lines:         []string{},
			startIndex:    0,
			expectedType:  "",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "simple paragraph",
			lines:         []string{"Hello world", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "Hello world",
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name:          "paragraph with bold",
			lines:         []string{"This is **bold** text", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "This is bold text",
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name:          "paragraph with italic",
			lines:         []string{"This is *italic* text", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "This is italic text",
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name:          "paragraph with code",
			lines:         []string{"This is `code` text", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "This is code text",
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name:          "multi-line paragraph",
			lines:         []string{"Line one", "Line two", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "Line one Line two",
			expectedLines: 3,
			wantErr:       false,
		},
		{
			name:          "paragraph stops at heading",
			lines:         []string{"Paragraph text", "# Heading", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "Paragraph text",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "paragraph stops at list",
			lines:         []string{"Paragraph text", "- List item", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "Paragraph text",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "paragraph stops at ordered list",
			lines:         []string{"Paragraph text", "1. List item", ""},
			startIndex:    0,
			expectedType:  adf_types.NodeTypeParagraph,
			expectedText:  "Paragraph text",
			expectedLines: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := pc.FromMarkdown(tt.lines, tt.startIndex, converter.ConversionContext{})
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if consumed != tt.expectedLines {
				t.Errorf("FromMarkdown() consumed = %v, want %v", consumed, tt.expectedLines)
			}

			if tt.expectedType == "" {
				// Empty result expected
				return
			}

			if node.Type != tt.expectedType {
				t.Errorf("FromMarkdown() node type = %v, want %v", node.Type, tt.expectedType)
			}

			// Collect text content from all text nodes
			var actualText string
			for _, child := range node.Content {
				if child.Type == adf_types.NodeTypeText {
					actualText += child.Text
				}
			}

			if actualText != tt.expectedText {
				t.Errorf("FromMarkdown() text content = %q, want %q", actualText, tt.expectedText)
			}
		})
	}
}

func TestParagraphConverter_RoundTrip(t *testing.T) {
	pc := NewParagraphConverter()

	tests := []struct {
		name     string
		markdown string
	}{
		{
			name:     "simple text",
			markdown: "Hello world\n\n",
		},
		{
			name:     "bold text",
			markdown: "This is **bold** text\n\n",
		},
		{
			name:     "italic text",
			markdown: "This is *italic* text\n\n",
		},
		{
			name:     "code text",
			markdown: "This is `code` text\n\n",
		},
		{
			name:     "mixed formatting",
			markdown: "Normal **bold** *italic* `code` text\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse markdown to ADF
			lines := splitLines(tt.markdown)
			node, _, err := pc.FromMarkdown(lines, 0, converter.ConversionContext{})
			if err != nil {
				t.Fatalf("FromMarkdown() failed: %v", err)
			}

			// Convert back to markdown
			result, err := pc.ToMarkdown(node, converter.ConversionContext{})
			if err != nil {
				t.Fatalf("ToMarkdown() failed: %v", err)
			}

			// Compare (allowing for minor whitespace differences)
			if result.Content != tt.markdown {
				t.Errorf("Round-trip failed:\nOriginal:  %q\nConverted: %q", tt.markdown, result.Content)
			}
		})
	}
}

func TestParagraphConverter_CanHandle(t *testing.T) {
	pc := NewParagraphConverter()

	tests := []struct {
		nodeType converter.ADFNodeType
		expected bool
	}{
		{converter.ADFNodeType(adf_types.NodeTypeParagraph), true},
		{converter.ADFNodeType(adf_types.NodeTypeText), false},
		{converter.ADFNodeType(adf_types.NodeTypeHeading), false},
		{converter.ADFNodeType(adf_types.NodeTypeHardBreak), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			result := pc.CanHandle(tt.nodeType)
			if result != tt.expected {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

func TestParagraphConverter_GetStrategy(t *testing.T) {
	pc := NewParagraphConverter()
	strategy := pc.GetStrategy()
	if strategy != converter.StandardMarkdown {
		t.Errorf("GetStrategy() = %v, want %v", strategy, converter.StandardMarkdown)
	}
}

func TestParagraphConverter_ValidateInput(t *testing.T) {
	pc := NewParagraphConverter()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid paragraph node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
			},
			wantErr: false,
		},
		{
			name: "invalid node type",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeText,
			},
			wantErr: true,
		},
		{
			name:    "invalid input type",
			input:   "not a node",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to split markdown into lines for testing
func splitLines(markdown string) []string {
	if markdown == "" {
		return []string{}
	}
	lines := make([]string, 0)
	current := ""
	for _, r := range markdown {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
