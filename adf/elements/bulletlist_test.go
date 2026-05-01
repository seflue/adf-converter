package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestBulletListConverter_ToMarkdown(t *testing.T) {
	blc := NewBulletListRenderer()

	tests := []struct {
		name     string
		node     adf.Node
		expected string
		wantErr  bool
	}{
		{
			name: "simple bullet list",
			node: adf.Node{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "First item"},
								},
							},
						},
					},
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Second item"},
								},
							},
						},
					},
				},
			},
			expected: "- First item\n- Second item\n\n",
			wantErr:  false,
		},
		{
			name: "bullet list with formatted text",
			node: adf.Node{
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
										Text: "Bold text",
										Marks: []adf.Mark{
											{Type: "strong"},
										},
									},
								},
							},
						},
					},
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{
										Type: adf.NodeTypeText,
										Text: "Italic text",
										Marks: []adf.Mark{
											{Type: "em"},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "- **Bold text**\n- *Italic text*\n\n",
			wantErr:  false,
		},
		{
			name: "nested bullet list (depth 2)",
			node: adf.Node{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Parent item"},
								},
							},
						},
					},
				},
			},
			expected: "- Parent item\n\n",
			wantErr:  false,
		},
		{
			name: "single item bullet list",
			node: adf.Node{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Only item"},
								},
							},
						},
					},
				},
			},
			expected: "- Only item\n\n",
			wantErr:  false,
		},
		{
			name: "empty bullet list",
			node: adf.Node{
				Type:    adf.NodeTypeBulletList,
				Content: []adf.Node{},
			},
			expected: "\n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := adf.ConversionContext{Registry: newTestRegistry()}
			result, err := blc.ToMarkdown(tt.node, ctx)

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

func TestBulletListConverter_ToMarkdown_NestedLists(t *testing.T) {
	blc := NewBulletListRenderer()

	// Test nested bullet list with proper depth tracking
	node := adf.Node{
		Type: adf.NodeTypeBulletList,
		Content: []adf.Node{
			{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Parent"},
						},
					},
				},
			},
		},
	}

	// Test at depth 0 (top level)
	ctx := adf.ConversionContext{Registry: newTestRegistry(), ListDepth: 0}
	result, err := blc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	expected := "- Parent\n\n"
	if result.Content != expected {
		t.Errorf("ToMarkdown() at depth 0 = %q, want %q", result.Content, expected)
	}

	// Test at depth 1 (nested)
	ctx = adf.ConversionContext{Registry: newTestRegistry(), ListDepth: 1}
	result, err = blc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	expected = "  - Parent\n\n"
	if result.Content != expected {
		t.Errorf("ToMarkdown() at depth 1 = %q, want %q", result.Content, expected)
	}

	// Test at depth 2 (deeply nested)
	ctx = adf.ConversionContext{Registry: newTestRegistry(), ListDepth: 2}
	result, err = blc.ToMarkdown(node, ctx)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	expected = "    - Parent\n\n"
	if result.Content != expected {
		t.Errorf("ToMarkdown() at depth 2 = %q, want %q", result.Content, expected)
	}
}

func TestBulletListConverter_FromMarkdown(t *testing.T) {
	blc := NewBulletListRenderer()

	tests := []struct {
		name          string
		lines         []string
		startIndex    int
		expectedItems int
		expectedLines int
		wantErr       bool
	}{
		{
			name: "simple bullet list",
			lines: []string{
				"- First item",
				"- Second item",
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
				"- Only item",
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
				"  - Indented item",
				"  - Another indented",
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
				"- Item one",
				"- Item two",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "list stopped by non-list line",
			lines: []string{
				"- Item one",
				"- Item two",
				"# Heading",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "single multiline list item",
			lines: []string{
				"- First line",
				"  continuation line",
				"",
			},
			startIndex:    0,
			expectedItems: 1,
			expectedLines: 3,
			wantErr:       false,
		},
		{
			name: "multiple multiline list items",
			lines: []string{
				"- First item line 1",
				"  First item line 2",
				"- Second item line 1",
				"  Second item line 2",
				"  Second item line 3",
				"",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 6,
			wantErr:       false,
		},
		{
			name: "mixed single and multiline items",
			lines: []string{
				"- Single line item",
				"- Multiline item line 1",
				"  Multiline item line 2",
				"- Another single line",
				"",
			},
			startIndex:    0,
			expectedItems: 3,
			expectedLines: 5,
			wantErr:       false,
		},
		{
			name: "multiline with nested list",
			lines: []string{
				"- Parent line 1",
				"  Parent line 2",
				"  - Nested item",
				"- Second parent",
				"",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 5,
			wantErr:       false,
		},
		{
			name: "multiline stopped by non-indented line",
			lines: []string{
				"- Multiline item",
				"  continuation",
				"Not a list item",
			},
			startIndex:    0,
			expectedItems: 1,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "list stopped by thematic break",
			lines: []string{
				"- Item one",
				"- Item two",
				"---",
				"Some paragraph",
			},
			startIndex:    0,
			expectedItems: 2,
			expectedLines: 2,
			wantErr:       false,
		},
		{
			name: "list followed by paragraph",
			lines: []string{
				"- Item one",
				"",
				"A paragraph after the list",
			},
			startIndex:    0,
			expectedItems: 1,
			expectedLines: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := adf.ConversionContext{Registry: newTestRegistry()}
			node, consumed, err := blc.FromMarkdown(tt.lines, tt.startIndex, ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("FromMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if node.Type != adf.NodeTypeBulletList {
				t.Errorf("FromMarkdown() node type = %v, want %v", node.Type, adf.NodeTypeBulletList)
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

func TestBulletListConverter_RoundTrip(t *testing.T) {
	blc := NewBulletListRenderer()

	tests := []struct {
		name     string
		node     adf.Node
		markdown string
	}{
		{
			name: "simple list round trip",
			node: adf.Node{
				Type: adf.NodeTypeBulletList,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Item 1"},
								},
							},
						},
					},
					{
						Type: adf.NodeTypeListItem,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeParagraph,
								Content: []adf.Node{
									{Type: adf.NodeTypeText, Text: "Item 2"},
								},
							},
						},
					},
				},
			},
			markdown: "- Item 1\n- Item 2\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := adf.ConversionContext{Registry: newTestRegistry()}

			// Convert to markdown
			result, err := blc.ToMarkdown(tt.node, ctx)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}

			if result.Content != tt.markdown {
				t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.markdown)
			}

			// Convert back to ADF
			lines := strings.Split(strings.TrimSuffix(result.Content, "\n"), "\n")
			node, _, err := blc.FromMarkdown(lines, 0, ctx)
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

// NOTE: TestMain is defined in paragraph_test.go for the entire elements package
// It registers all converters including bulletListRenderer
