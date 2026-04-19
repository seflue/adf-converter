package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

func TestOrderedListConverter_ToMarkdown(t *testing.T) {
	olc := NewOrderedListConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		expected string
		wantErr  bool
	}{
		{
			name: "simple ordered list",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "First item"},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Second item"},
								},
							},
						},
					},
				},
			},
			expected: "1. First item\n2. Second item\n\n",
			wantErr:  false,
		},
		{
			name: "ordered list with formatted text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeText,
										Text: "Bold text",
										Marks: []adf_types.ADFMark{
											{Type: "strong"},
										},
									},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeText,
										Text: "Italic text",
										Marks: []adf_types.ADFMark{
											{Type: "em"},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "1. **Bold text**\n2. *Italic text*\n\n",
			wantErr:  false,
		},
		{
			name: "ordered list with three items",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "First"},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Second"},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Third"},
								},
							},
						},
					},
				},
			},
			expected: "1. First\n2. Second\n3. Third\n\n",
			wantErr:  false,
		},
		{
			name: "single item ordered list",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Only item"},
								},
							},
						},
					},
				},
			},
			expected: "1. Only item\n\n",
			wantErr:  false,
		},
		{
			name: "empty ordered list",
			node: adf_types.ADFNode{
				Type:    adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{},
			},
			expected: "\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := converter.ConversionContext{}
			result, err := olc.ToMarkdown(tt.node, ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.Content != tt.expected {
				t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.expected)
			}
		})
	}
}

func TestOrderedListConverter_ToMarkdown_NestedLists(t *testing.T) {
	olc := NewOrderedListConverter()

	// Test nested ordered list with proper depth tracking
	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeOrderedList,
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Parent"},
						},
					},
				},
			},
		},
	}

	// Test at depth 0 (top level)
	ctx := converter.ConversionContext{ListDepth: 0}
	result, err := olc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	expected := "1. Parent\n\n"
	if result.Content != expected {
		t.Errorf("ToMarkdown() at depth 0 = %q, want %q", result.Content, expected)
	}

	// Test at depth 1 (nested)
	ctx = converter.ConversionContext{ListDepth: 1}
	result, err = olc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	expected = "  1. Parent\n\n"
	if result.Content != expected {
		t.Errorf("ToMarkdown() at depth 1 = %q, want %q", result.Content, expected)
	}

	// Test at depth 2 (deeply nested)
	ctx = converter.ConversionContext{ListDepth: 2}
	result, err = olc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	expected = "    1. Parent\n\n"
	if result.Content != expected {
		t.Errorf("ToMarkdown() at depth 2 = %q, want %q", result.Content, expected)
	}
}

func TestOrderedListConverter_FromMarkdown(t *testing.T) {
	olc := NewOrderedListConverter()

	tests := []struct {
		name          string
		lines         []string
		startIndex    int
		expectedItems int
		expectedLines int
		wantErr       bool
	}{
		{
			name: "simple ordered list",
			lines: []string{
				"1. First item",
				"2. Second item",
				"",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 3,
			wantErr:       false,
		},
		{
			name: "single item",
			lines: []string{
				"1. Only item",
				"",
			},
			startIndex:    0,
			expectedItems: 1,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "list with indented items",
			lines: []string{
				"  1. Indented item",
				"  2. Another indented",
				"",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 3,
			wantErr:       false,
		},
		{
			name: "list ending at EOF",
			lines: []string{
				"1. Item one",
				"2. Item two",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "list stopped by non-list line",
			lines: []string{
				"1. Item one",
				"2. Item two",
				"# Heading",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "list with non-sequential numbers",
			lines: []string{
				"1. First",
				"3. Third (will parse as second)",
				"5. Fifth (will parse as third)",
				"",
			},
			startIndex:    0,
			expectedItems: 3,
			expectedLines: 4,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := converter.ConversionContext{}
			node, consumed, err := olc.FromMarkdown(tt.lines, tt.startIndex, ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if node.Type != adf_types.NodeTypeOrderedList {
				t.Errorf("FromMarkdown() node type = %v, want %v", node.Type, adf_types.NodeTypeOrderedList)
			}

			if len(node.Content) != tt.expectedItems {
				t.Errorf("FromMarkdown() items = %d, want %d", len(node.Content), tt.expectedItems)
			}

			if consumed != tt.expectedLines {
				t.Errorf("FromMarkdown() consumed = %d, want %d", consumed, tt.expectedLines)
			}
		})
	}
}

func TestOrderedListConverter_RoundTrip(t *testing.T) {
	olc := NewOrderedListConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		markdown string
	}{
		{
			name: "simple list round trip",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Item 1"},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Item 2"},
								},
							},
						},
					},
				},
			},
			markdown: "1. Item 1\n2. Item 2\n\n",
		},
		{
			name: "three items round trip",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "First"},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Second"},
								},
							},
						},
					},
					{
						Type: adf_types.NodeTypeListItem,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeParagraph,
								Content: []adf_types.ADFNode{
									{Type: adf_types.NodeTypeText, Text: "Third"},
								},
							},
						},
					},
				},
			},
			markdown: "1. First\n2. Second\n3. Third\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := converter.ConversionContext{}

			// Convert to markdown
			result, err := olc.ToMarkdown(tt.node, ctx)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}

			if result.Content != tt.markdown {
				t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.markdown)
			}

			// Convert back to ADF
			lines := strings.Split(strings.TrimSuffix(result.Content, "\n"), "\n")
			node, _, err := olc.FromMarkdown(lines, 0, ctx)
			if err != nil {
				t.Fatalf("FromMarkdown() error = %v", err)
			}

			// Verify structure
			if node.Type != tt.node.Type {
				t.Errorf("Round trip node type = %v, want %v", node.Type, tt.node.Type)
			}

			if len(node.Content) != len(tt.node.Content) {
				t.Errorf("Round trip items = %d, want %d", len(node.Content), len(tt.node.Content))
			}
		})
	}
}

func TestOrderedListConverter_CanHandle(t *testing.T) {
	olc := NewOrderedListConverter()

	tests := []struct {
		nodeType converter.ADFNodeType
		expected bool
	}{
		{converter.ADFNodeType(adf_types.NodeTypeOrderedList), true},
		{converter.ADFNodeType(adf_types.NodeTypeBulletList), false},
		{converter.ADFNodeType(adf_types.NodeTypeListItem), false},
		{converter.ADFNodeType(adf_types.NodeTypeParagraph), false},
		{converter.ADFNodeType(adf_types.NodeTypeText), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			result := olc.CanHandle(tt.nodeType)
			if result != tt.expected {
				t.Errorf("CanHandle(%v) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

func TestOrderedListConverter_GetStrategy(t *testing.T) {
	olc := NewOrderedListConverter()

	strategy := olc.GetStrategy()
	if strategy != converter.StandardMarkdown {
		t.Errorf("GetStrategy() = %v, want %v", strategy, converter.StandardMarkdown)
	}
}

func TestOrderedListConverter_ValidateInput(t *testing.T) {
	olc := NewOrderedListConverter()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid ordered list node",
			input: adf_types.ADFNode{
				Type:    adf_types.NodeTypeOrderedList,
				Content: []adf_types.ADFNode{},
			},
			wantErr: false,
		},
		{
			name: "wrong node type",
			input: adf_types.ADFNode{
				Type:    adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{},
			},
			wantErr: true,
		},
		{
			name:    "not an ADFNode",
			input:   "not a node",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := olc.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderedListConverter_StartNumber(t *testing.T) {
	olc := NewOrderedListConverter()

	t.Run("ToMarkdown", func(t *testing.T) {
		tests := []struct {
			name     string
			attrs    map[string]any
			expected string
		}{
			{
				name:     "start at 5",
				attrs:    map[string]any{"order": float64(5)},
				expected: "5. First\n6. Second\n\n",
			},
			{
				name:     "start at 0",
				attrs:    map[string]any{"order": float64(0)},
				expected: "0. First\n1. Second\n\n",
			},
			{
				name:     "default start (no attrs)",
				attrs:    nil,
				expected: "1. First\n2. Second\n\n",
			},
			{
				name:     "explicit start at 1 (default)",
				attrs:    map[string]any{"order": float64(1)},
				expected: "1. First\n2. Second\n\n",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := adf_types.ADFNode{
					Type:  adf_types.NodeTypeOrderedList,
					Attrs: tt.attrs,
					Content: []adf_types.ADFNode{
						{
							Type: adf_types.NodeTypeListItem,
							Content: []adf_types.ADFNode{
								{
									Type:    adf_types.NodeTypeParagraph,
									Content: []adf_types.ADFNode{{Type: adf_types.NodeTypeText, Text: "First"}},
								},
							},
						},
						{
							Type: adf_types.NodeTypeListItem,
							Content: []adf_types.ADFNode{
								{
									Type:    adf_types.NodeTypeParagraph,
									Content: []adf_types.ADFNode{{Type: adf_types.NodeTypeText, Text: "Second"}},
								},
							},
						},
					},
				}

				ctx := converter.ConversionContext{}
				result, err := olc.ToMarkdown(node, ctx)
				if err != nil {
					t.Fatalf("ToMarkdown() error = %v", err)
				}
				if result.Content != tt.expected {
					t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.expected)
				}
			})
		}
	})

	t.Run("FromMarkdown", func(t *testing.T) {
		tests := []struct {
			name          string
			lines         []string
			expectedAttrs map[string]any
		}{
			{
				name:          "start at 5",
				lines:         []string{"5. First", "6. Second", ""},
				expectedAttrs: map[string]any{"order": float64(5)},
			},
			{
				name:          "start at 0",
				lines:         []string{"0. First", "1. Second", ""},
				expectedAttrs: map[string]any{"order": float64(0)},
			},
			{
				name:          "start at 1 (default, no attrs)",
				lines:         []string{"1. First", "2. Second", ""},
				expectedAttrs: nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := converter.ConversionContext{}
				node, _, err := olc.FromMarkdown(tt.lines, 0, ctx)
				if err != nil {
					t.Fatalf("FromMarkdown() error = %v", err)
				}

				if tt.expectedAttrs == nil {
					if node.Attrs != nil {
						t.Errorf("FromMarkdown() attrs = %v, want nil", node.Attrs)
					}
				} else {
					if node.Attrs == nil {
						t.Fatalf("FromMarkdown() attrs = nil, want %v", tt.expectedAttrs)
					}
					expectedOrder := tt.expectedAttrs["order"]
					gotOrder := node.Attrs["order"]
					if gotOrder != expectedOrder {
						t.Errorf("FromMarkdown() order = %v, want %v", gotOrder, expectedOrder)
					}
				}
			})
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		tests := []struct {
			name     string
			attrs    map[string]any
			markdown string
		}{
			{
				name:     "start at 5 survives roundtrip",
				attrs:    map[string]any{"order": float64(5)},
				markdown: "5. Alpha\n6. Beta\n\n",
			},
			{
				name:     "default start survives roundtrip",
				attrs:    nil,
				markdown: "1. Alpha\n2. Beta\n\n",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := adf_types.ADFNode{
					Type:  adf_types.NodeTypeOrderedList,
					Attrs: tt.attrs,
					Content: []adf_types.ADFNode{
						{
							Type: adf_types.NodeTypeListItem,
							Content: []adf_types.ADFNode{
								{
									Type:    adf_types.NodeTypeParagraph,
									Content: []adf_types.ADFNode{{Type: adf_types.NodeTypeText, Text: "Alpha"}},
								},
							},
						},
						{
							Type: adf_types.NodeTypeListItem,
							Content: []adf_types.ADFNode{
								{
									Type:    adf_types.NodeTypeParagraph,
									Content: []adf_types.ADFNode{{Type: adf_types.NodeTypeText, Text: "Beta"}},
								},
							},
						},
					},
				}

				ctx := converter.ConversionContext{}

				// ADF → Markdown
				result, err := olc.ToMarkdown(node, ctx)
				if err != nil {
					t.Fatalf("ToMarkdown() error = %v", err)
				}
				if result.Content != tt.markdown {
					t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.markdown)
				}

				// Markdown → ADF
				lines := strings.Split(strings.TrimSuffix(result.Content, "\n"), "\n")
				roundTripped, _, err := olc.FromMarkdown(lines, 0, ctx)
				if err != nil {
					t.Fatalf("FromMarkdown() error = %v", err)
				}

				// Verify attrs survived
				if tt.attrs == nil {
					if roundTripped.Attrs != nil {
						t.Errorf("RoundTrip attrs = %v, want nil", roundTripped.Attrs)
					}
				} else {
					if roundTripped.Attrs == nil {
						t.Fatalf("RoundTrip attrs = nil, want %v", tt.attrs)
					}
					if roundTripped.Attrs["order"] != tt.attrs["order"] {
						t.Errorf("RoundTrip order = %v, want %v", roundTripped.Attrs["order"], tt.attrs["order"])
					}
				}

				// Verify content survived
				if len(roundTripped.Content) != len(node.Content) {
					t.Errorf("RoundTrip items = %d, want %d", len(roundTripped.Content), len(node.Content))
				}
			})
		}
	})
}

// NOTE: TestMain is defined in paragraph_test.go for the entire elements package
// It registers all converters including orderedListConverter
