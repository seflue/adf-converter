package elements

import (
	"testing"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

func TestHeadingConverter_ToMarkdown(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
		wantErr  bool
	}{
		{
			name: "h1 heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 1,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Heading 1",
					},
				},
			},
			expected: "# Heading 1\n\n",
			wantErr:  false,
		},
		{
			name: "h2 heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 2,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Heading 2",
					},
				},
			},
			expected: "## Heading 2\n\n",
			wantErr:  false,
		},
		{
			name: "h3 heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 3,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Heading 3",
					},
				},
			},
			expected: "### Heading 3\n\n",
			wantErr:  false,
		},
		{
			name: "h4 heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 4,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Heading 4",
					},
				},
			},
			expected: "#### Heading 4\n\n",
			wantErr:  false,
		},
		{
			name: "h5 heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 5,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Heading 5",
					},
				},
			},
			expected: "##### Heading 5\n\n",
			wantErr:  false,
		},
		{
			name: "h6 heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 6,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Heading 6",
					},
				},
			},
			expected: "###### Heading 6\n\n",
			wantErr:  false,
		},
		{
			name: "heading with bold text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 2,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Bold ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "heading",
						Marks: []adf_types.ADFMark{
							{Type: "strong"},
						},
					},
				},
			},
			expected: "## Bold **heading**\n\n",
			wantErr:  false,
		},
		{
			name: "heading with italic text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 3,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Italic ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "heading",
						Marks: []adf_types.ADFMark{
							{Type: "em"},
						},
					},
				},
			},
			expected: "### Italic *heading*\n\n",
			wantErr:  false,
		},
		{
			name: "heading with code",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 4,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Code: ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "example",
						Marks: []adf_types.ADFMark{
							{Type: "code"},
						},
					},
				},
			},
			expected: "#### Code: `example`\n\n",
			wantErr:  false,
		},
		{
			name: "heading with mixed formatting",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 2,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Mixed ",
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
						Text: " and ",
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "italic",
						Marks: []adf_types.ADFMark{
							{Type: "em"},
						},
					},
				},
			},
			expected: "## Mixed **bold** and *italic*\n\n",
			wantErr:  false,
		},
		{
			name: "empty heading",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 1,
				},
				Content: []adf_types.ADFNode{},
			},
			expected: "# \n\n",
			wantErr:  false,
		},
		{
			name: "heading removes newlines from content",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
				Attrs: map[string]interface{}{
					"level": 1,
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeText,
						Text: "Text",
					},
					{
						Type: adf_types.NodeTypeHardBreak,
					},
					{
						Type: adf_types.NodeTypeText,
						Text: "More",
					},
				},
			},
			expected: "# Text More\n\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hc.ToMarkdown(tt.node, converter.ConversionContext{})
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

func TestHeadingConverter_FromMarkdown(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		name          string
		lines         []string
		startIndex    int
		expectedLevel int
		expectedText  string
		expectedLines int
		wantErr       bool
	}{
		{
			name:          "h1 heading",
			lines:         []string{"# Heading 1", ""},
			startIndex:    0,
			expectedLevel: 1,
			expectedText:  "Heading 1",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "h2 heading",
			lines:         []string{"## Heading 2", ""},
			startIndex:    0,
			expectedLevel: 2,
			expectedText:  "Heading 2",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "h3 heading",
			lines:         []string{"### Heading 3", ""},
			startIndex:    0,
			expectedLevel: 3,
			expectedText:  "Heading 3",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "h4 heading",
			lines:         []string{"#### Heading 4", ""},
			startIndex:    0,
			expectedLevel: 4,
			expectedText:  "Heading 4",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "h5 heading",
			lines:         []string{"##### Heading 5", ""},
			startIndex:    0,
			expectedLevel: 5,
			expectedText:  "Heading 5",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "h6 heading",
			lines:         []string{"###### Heading 6", ""},
			startIndex:    0,
			expectedLevel: 6,
			expectedText:  "Heading 6",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "heading with bold",
			lines:         []string{"## Heading with **bold**", ""},
			startIndex:    0,
			expectedLevel: 2,
			expectedText:  "Heading with bold",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "heading with italic",
			lines:         []string{"### Heading with *italic*", ""},
			startIndex:    0,
			expectedLevel: 3,
			expectedText:  "Heading with italic",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "heading with code",
			lines:         []string{"#### Heading with `code`", ""},
			startIndex:    0,
			expectedLevel: 4,
			expectedText:  "Heading with code",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "heading with mixed formatting",
			lines:         []string{"## **Bold** and *italic* heading", ""},
			startIndex:    0,
			expectedLevel: 2,
			expectedText:  "Bold and italic heading",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "empty heading",
			lines:         []string{"# ", ""},
			startIndex:    0,
			expectedLevel: 1,
			expectedText:  "",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "invalid level too low",
			lines:         []string{"Heading without #", ""},
			startIndex:    0,
			expectedLines: 0,
			wantErr:       true,
		},
		{
			name:          "invalid level too high",
			lines:         []string{"####### Too many #", ""},
			startIndex:    0,
			expectedLines: 0,
			wantErr:       true,
		},
		{
			name:          "indented heading",
			lines:         []string{"   ## Indented Heading", "next line"},
			startIndex:    0,
			expectedLevel: 2,
			expectedText:  "Indented Heading",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "tab-indented heading",
			lines:         []string{"\t### Tab Heading", ""},
			startIndex:    0,
			expectedLevel: 3,
			expectedText:  "Tab Heading",
			expectedLines: 1,
			wantErr:       false,
		},
		{
			name:          "empty lines",
			lines:         []string{},
			startIndex:    0,
			expectedLines: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := hc.FromMarkdown(tt.lines, tt.startIndex, converter.ConversionContext{})
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if consumed != tt.expectedLines {
				t.Errorf("FromMarkdown() consumed = %v, want %v", consumed, tt.expectedLines)
			}

			if tt.wantErr {
				return
			}

			if node.Type != adf_types.NodeTypeHeading {
				t.Errorf("FromMarkdown() node type = %v, want %v", node.Type, adf_types.NodeTypeHeading)
			}

			level, ok := node.Attrs["level"].(int)
			if !ok {
				t.Errorf("FromMarkdown() level attribute missing or wrong type")
				return
			}

			if level != tt.expectedLevel {
				t.Errorf("FromMarkdown() level = %v, want %v", level, tt.expectedLevel)
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

func TestHeadingConverter_RoundTrip(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		name     string
		markdown string
		level    int
	}{
		{
			name:     "h1 simple",
			markdown: "# Simple Heading\n\n",
			level:    1,
		},
		{
			name:     "h2 simple",
			markdown: "## Simple Heading\n\n",
			level:    2,
		},
		{
			name:     "h3 with bold",
			markdown: "### Heading with **bold**\n\n",
			level:    3,
		},
		{
			name:     "h4 with italic",
			markdown: "#### Heading with *italic*\n\n",
			level:    4,
		},
		{
			name:     "h5 with code",
			markdown: "##### Heading with `code`\n\n",
			level:    5,
		},
		{
			name:     "h6 with mixed",
			markdown: "###### **Bold** and *italic*\n\n",
			level:    6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse markdown to ADF
			lines := splitLines(tt.markdown)
			node, _, err := hc.FromMarkdown(lines, 0, converter.ConversionContext{})
			if err != nil {
				t.Fatalf("FromMarkdown() failed: %v", err)
			}

			// Verify level was preserved
			level, ok := node.Attrs["level"].(int)
			if !ok || level != tt.level {
				t.Errorf("Level not preserved: got %v, want %v", level, tt.level)
			}

			// Convert back to markdown
			result, err := hc.ToMarkdown(node, converter.ConversionContext{})
			if err != nil {
				t.Fatalf("ToMarkdown() failed: %v", err)
			}

			// Compare
			if result.Content != tt.markdown {
				t.Errorf("Round-trip failed:\nOriginal:  %q\nConverted: %q", tt.markdown, result.Content)
			}
		})
	}
}

func TestHeadingConverter_CanHandle(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		nodeType converter.ADFNodeType
		expected bool
	}{
		{converter.ADFNodeType(adf_types.NodeTypeHeading), true},
		{converter.ADFNodeType(adf_types.NodeTypeText), false},
		{converter.ADFNodeType(adf_types.NodeTypeParagraph), false},
		{converter.ADFNodeType(adf_types.NodeTypeHardBreak), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			result := hc.CanHandle(tt.nodeType)
			if result != tt.expected {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

func TestHeadingConverter_GetStrategy(t *testing.T) {
	hc := NewHeadingConverter()
	strategy := hc.GetStrategy()
	if strategy != converter.StandardMarkdown {
		t.Errorf("GetStrategy() = %v, want %v", strategy, converter.StandardMarkdown)
	}
}

func TestHeadingConverter_ValidateInput(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid heading node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeHeading,
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
			err := hc.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
