package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
				Attrs: map[string]any{
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
			result, err := hc.ToMarkdown(tt.node, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
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
			expectedLines: 0,
			wantErr:       true, // tab = 4 spaces indentation → not an ATX heading per CommonMark
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
			node, consumed, err := hc.FromMarkdown(tt.lines, tt.startIndex, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
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
			node, _, err := hc.FromMarkdown(lines, 0, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
			if err != nil {
				t.Fatalf("FromMarkdown() failed: %v", err)
			}

			// Verify level was preserved
			level, ok := node.Attrs["level"].(int)
			if !ok || level != tt.level {
				t.Errorf("Level not preserved: got %v, want %v", level, tt.level)
			}

			// Convert back to markdown
			result, err := hc.ToMarkdown(node, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
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

func TestHeadingConverter_CanParseLine(t *testing.T) {
	hc := &headingConverter{}

	tests := []struct {
		line     string
		expected bool
	}{
		{"# Heading", true},
		{"## Heading 2", true},
		{"###### H6", true},
		{"#", true},              // bare hash is a heading
		{"#NoSpace", false},      // no space after hash — not a valid CommonMark heading
		{"####### seven", false}, // 7 hashes — not a valid CommonMark heading level
		{"not a heading", false},
		{"  ## indented", false}, // leading spaces — HasPrefix does not trim
		{"", false},
		{" # space first", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := hc.CanParseLine(tt.line)
			if result != tt.expected {
				t.Errorf("CanParseLine(%q) = %v, want %v", tt.line, result, tt.expected)
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
		input   any
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

// TestHeadingConverter_FromMarkdown_MarksPreserved verifies that inline marks
// (bold, italic, code) are present on the correct nodes in the ADF output.
// The existing FromMarkdown table tests only collect plain text, so a silent
// mark-loss during a Goldmark migration would not be caught there.
func TestHeadingConverter_FromMarkdown_MarksPreserved(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		name      string
		line      string
		wantNodes []adf_types.ADFNode
	}{
		{
			name: "bold mark",
			line: "## **bold text**",
			wantNodes: []adf_types.ADFNode{
				{
					Type:  adf_types.NodeTypeText,
					Text:  "bold text",
					Marks: []adf_types.ADFMark{{Type: "strong"}},
				},
			},
		},
		{
			name: "italic mark",
			line: "### *italic text*",
			wantNodes: []adf_types.ADFNode{
				{
					Type:  adf_types.NodeTypeText,
					Text:  "italic text",
					Marks: []adf_types.ADFMark{{Type: "em"}},
				},
			},
		},
		{
			name: "code mark",
			line: "#### `code snippet`",
			wantNodes: []adf_types.ADFNode{
				{
					Type:  adf_types.NodeTypeText,
					Text:  "code snippet",
					Marks: []adf_types.ADFMark{{Type: "code"}},
				},
			},
		},
		{
			name: "mixed: plain + bold",
			line: "## Intro **bold** end",
			wantNodes: []adf_types.ADFNode{
				{Type: adf_types.NodeTypeText, Text: "Intro "},
				{
					Type:  adf_types.NodeTypeText,
					Text:  "bold",
					Marks: []adf_types.ADFMark{{Type: "strong"}},
				},
				{Type: adf_types.NodeTypeText, Text: " end"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := hc.FromMarkdown([]string{tt.line}, 0, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if consumed != 1 {
				t.Errorf("consumed = %d, want 1", consumed)
			}
			if len(node.Content) != len(tt.wantNodes) {
				t.Fatalf("content len = %d, want %d (content: %v)", len(node.Content), len(tt.wantNodes), node.Content)
			}
			for i, want := range tt.wantNodes {
				got := node.Content[i]
				if got.Text != want.Text {
					t.Errorf("node[%d].Text = %q, want %q", i, got.Text, want.Text)
				}
				if len(got.Marks) != len(want.Marks) {
					t.Errorf("node[%d].Marks len = %d, want %d", i, len(got.Marks), len(want.Marks))
					continue
				}
				for j, wm := range want.Marks {
					if got.Marks[j].Type != wm.Type {
						t.Errorf("node[%d].Marks[%d].Type = %q, want %q", i, j, got.Marks[j].Type, wm.Type)
					}
				}
			}
		})
	}
}

// TestHeadingConverter_FromMarkdown_StartIndex verifies that the parser
// correctly starts at a non-zero index and reports consumed relative to the
// slice, not the full document.
func TestHeadingConverter_FromMarkdown_StartIndex(t *testing.T) {
	hc := NewHeadingConverter()
	lines := []string{"paragraph text", "## The Heading", "more text"}

	node, consumed, err := hc.FromMarkdown(lines, 1, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumed != 1 {
		t.Errorf("consumed = %d, want 1", consumed)
	}
	level, ok := node.Attrs["level"].(int)
	if !ok || level != 2 {
		t.Errorf("level = %v, want 2", node.Attrs["level"])
	}
}

// TestHeadingConverter_FromMarkdown_StartIndexOutOfRange ensures an error is
// returned when startIndex is past the end of the slice.
func TestHeadingConverter_FromMarkdown_StartIndexOutOfRange(t *testing.T) {
	hc := NewHeadingConverter()
	lines := []string{"# Only one line"}

	_, _, err := hc.FromMarkdown(lines, 1, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
	if err == nil {
		t.Error("expected error for out-of-range startIndex, got nil")
	}
}

// TestHeadingConverter_ToMarkdown_InvalidLevel documents the fallback behaviour
// for heading levels outside [1, 6].  Both cases must render as an h1 so that
// ADF round-trips remain stable even when the source JSON contains bad data.
func TestHeadingConverter_ToMarkdown_InvalidLevel(t *testing.T) {
	hc := NewHeadingConverter()

	tests := []struct {
		name     string
		level    int
		expected string
	}{
		{"level 0 defaults to h1", 0, "# Text\n\n"},
		{"level 7 defaults to h1", 7, "# Text\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := adf_types.ADFNode{
				Type:  adf_types.NodeTypeHeading,
				Attrs: map[string]any{"level": tt.level},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Text"},
				},
			}
			result, err := hc.ToMarkdown(node, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Content != tt.expected {
				t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.expected)
			}
		})
	}
}

// TestHeadingConverter_FromMarkdown_HashWithoutSpace verifies that "#text"
// (no space after the hash) is rejected.  CommonMark requires a space or
// end-of-line after the opening sequence, so "#NoSpace" is not a heading.
// The current handwritten parser accepts it (bug); this test drives the fix
// via the Goldmark migration.
func TestHeadingConverter_FromMarkdown_HashWithoutSpace(t *testing.T) {
	hc := NewHeadingConverter()

	_, _, err := hc.FromMarkdown([]string{"#NoSpace"}, 0, converter.ConversionContext{Registry: converter.GetGlobalRegistry()})
	if err == nil {
		t.Error("expected error for heading without space after #, got nil")
	}
}
