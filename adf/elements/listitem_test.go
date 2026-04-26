package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
)

func TestListItemConverter_ToMarkdown(t *testing.T) {
	lic := NewListItemRenderer()

	tests := []struct {
		name     string
		node     adf.Node
		context  adf.ConversionContext
		expected string
		wantErr  bool
	}{
		{
			name: "simple list item at depth 1",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "First item",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 1,
			},
			expected: "- First item\n",
			wantErr:  false,
		},
		{
			name: "list item at depth 2 (nested)",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Nested item",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 2,
			},
			expected: "  - Nested item\n",
			wantErr:  false,
		},
		{
			name: "list item at depth 3 (deeply nested)",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Deep nested item",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 3,
			},
			expected: "    - Deep nested item\n",
			wantErr:  false,
		},
		{
			name: "list item with bold text",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Item with ",
							},
							{
								Type: adf.NodeTypeText,
								Text: "bold",
								Marks: []adf.Mark{
									{Type: "strong"},
								},
							},
							{
								Type: adf.NodeTypeText,
								Text: " text",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 1,
			},
			expected: "- Item with **bold** text\n",
			wantErr:  false,
		},
		{
			name: "list item with italic text",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Item with ",
							},
							{
								Type: adf.NodeTypeText,
								Text: "italic",
								Marks: []adf.Mark{
									{Type: "em"},
								},
							},
							{
								Type: adf.NodeTypeText,
								Text: " text",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 1,
			},
			expected: "- Item with *italic* text\n",
			wantErr:  false,
		},
		{
			name: "list item with code text",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Item with ",
							},
							{
								Type: adf.NodeTypeText,
								Text: "code",
								Marks: []adf.Mark{
									{Type: "code"},
								},
							},
							{
								Type: adf.NodeTypeText,
								Text: " text",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 1,
			},
			expected: "- Item with `code` text\n",
			wantErr:  false,
		},
		{
			name: "empty list item",
			node: adf.Node{
				Type:    adf.NodeTypeListItem,
				Content: []adf.Node{},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 1,
			},
			expected: "- \n",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := lic.ToMarkdown(tt.node, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Content != tt.expected {
				t.Errorf("ToMarkdown() = %q, want %q", result.Content, tt.expected)
			}
		})
	}
}

func TestListItemConverter_FromMarkdown(t *testing.T) {
	lic := NewListItemRenderer()

	tests := []struct {
		name         string
		lines        []string
		startIndex   int
		expectedNode adf.Node
		consumed     int
		wantErr      bool
	}{
		{
			name:       "simple list item",
			lines:      []string{"- First item"},
			startIndex: 0,
			expectedNode: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "First item",
							},
						},
					},
				},
			},
			consumed: 1,
			wantErr:  false,
		},
		{
			name:       "list item with leading spaces",
			lines:      []string{"  - Nested item"},
			startIndex: 0,
			expectedNode: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Nested item",
							},
						},
					},
				},
			},
			consumed: 1,
			wantErr:  false,
		},
		{
			name:       "empty list item",
			lines:      []string{"- "},
			startIndex: 0,
			expectedNode: adf.Node{
				Type:    adf.NodeTypeListItem,
				Content: []adf.Node{},
			},
			consumed: 1,
			wantErr:  false,
		},
		{
			name:       "list item with extra content after",
			lines:      []string{"- First item", "- Second item"},
			startIndex: 0,
			expectedNode: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "First item",
							},
						},
					},
				},
			},
			consumed: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, consumed, err := lic.FromMarkdown(tt.lines, tt.startIndex, adf.ConversionContext{Registry: newTestRegistry()})
			if (err != nil) != tt.wantErr {
				t.Errorf("FromMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if consumed != tt.consumed {
				t.Errorf("FromMarkdown() consumed = %d, want %d", consumed, tt.consumed)
			}

			// Check node type
			if node.Type != tt.expectedNode.Type {
				t.Errorf("FromMarkdown() node type = %s, want %s", node.Type, tt.expectedNode.Type)
			}

			// Check content length
			if len(node.Content) != len(tt.expectedNode.Content) {
				t.Errorf("FromMarkdown() content length = %d, want %d", len(node.Content), len(tt.expectedNode.Content))
				return
			}

			// For non-empty list items, verify the paragraph structure
			if len(node.Content) > 0 {
				if node.Content[0].Type != adf.NodeTypeParagraph {
					t.Errorf("FromMarkdown() first child type = %s, want paragraph", node.Content[0].Type)
				}
			}
		})
	}
}

func TestListItemConverter_RoundTrip(t *testing.T) {
	lic := NewListItemRenderer()

	tests := []struct {
		name    string
		node    adf.Node
		context adf.ConversionContext
	}{
		{
			name: "simple list item",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Test item",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 1,
			},
		},
		{
			name: "nested list item",
			node: adf.Node{
				Type: adf.NodeTypeListItem,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeText,
								Text: "Nested test item",
							},
						},
					},
				},
			},
			context: adf.ConversionContext{Registry: newTestRegistry(), 
				ListDepth: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to markdown
			result, err := lic.ToMarkdown(tt.node, tt.context)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}

			// Convert back to ADF
			lines := []string{result.Content[:len(result.Content)-1]} // Remove trailing newline
			node, _, err := lic.FromMarkdown(lines, 0, tt.context)
			if err != nil {
				t.Fatalf("FromMarkdown() error = %v", err)
			}

			// Verify structure is preserved
			if node.Type != tt.node.Type {
				t.Errorf("Round-trip node type = %s, want %s", node.Type, tt.node.Type)
			}

			if len(node.Content) != len(tt.node.Content) {
				t.Errorf("Round-trip content length = %d, want %d", len(node.Content), len(tt.node.Content))
			}
		})
	}
}

func TestListItemConverter_CanHandle(t *testing.T) {
	lic := NewListItemRenderer()

	tests := []struct {
		nodeType adf.NodeType
		want     bool
	}{
		{adf.NodeTypeListItem, true},
		{adf.NodeTypeParagraph, false},
		{adf.NodeTypeHeading, false},
		{adf.NodeTypeText, false},
		{adf.NodeTypeBulletList, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.nodeType), func(t *testing.T) {
			if got := lic.CanHandle(tt.nodeType); got != tt.want {
				t.Errorf("CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListItemConverter_GetStrategy(t *testing.T) {
	lic := NewListItemRenderer()
	strategy := lic.GetStrategy()

	if strategy != adf.StandardMarkdown {
		t.Errorf("GetStrategy() = %v, want %v", strategy, adf.StandardMarkdown)
	}
}

func TestListItemConverter_ValidateInput(t *testing.T) {
	lic := NewListItemRenderer()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid list item node",
			input: adf.Node{
				Type: adf.NodeTypeListItem,
			},
			wantErr: false,
		},
		{
			name:    "invalid type - not Node",
			input:   "not a node",
			wantErr: true,
		},
		{
			name: "invalid type - wrong node type",
			input: adf.Node{
				Type: adf.NodeTypeParagraph,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lic.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
