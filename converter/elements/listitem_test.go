package elements

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

func TestListItemConverter_ToMarkdown(t *testing.T) {
	lic := NewListItemConverter()

	tests := []struct {
		name     string
		node     adf_types.ADFNode
		context  converter.ConversionContext
		expected string
		wantErr  bool
	}{
		{
			name: "simple list item at depth 1",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "First item",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 1,
			},
			expected: "- First item\n",
			wantErr:  false,
		},
		{
			name: "list item at depth 2 (nested)",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Nested item",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 2,
			},
			expected: "  - Nested item\n",
			wantErr:  false,
		},
		{
			name: "list item at depth 3 (deeply nested)",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Deep nested item",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 3,
			},
			expected: "    - Deep nested item\n",
			wantErr:  false,
		},
		{
			name: "list item with bold text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Item with ",
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
								Text: " text",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 1,
			},
			expected: "- Item with **bold** text\n",
			wantErr:  false,
		},
		{
			name: "list item with italic text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Item with ",
							},
							{
								Type: adf_types.NodeTypeText,
								Text: "italic",
								Marks: []adf_types.ADFMark{
									{Type: "em"},
								},
							},
							{
								Type: adf_types.NodeTypeText,
								Text: " text",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 1,
			},
			expected: "- Item with *italic* text\n",
			wantErr:  false,
		},
		{
			name: "list item with code text",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Item with ",
							},
							{
								Type: adf_types.NodeTypeText,
								Text: "code",
								Marks: []adf_types.ADFMark{
									{Type: "code"},
								},
							},
							{
								Type: adf_types.NodeTypeText,
								Text: " text",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 1,
			},
			expected: "- Item with `code` text\n",
			wantErr:  false,
		},
		{
			name: "empty list item",
			node: adf_types.ADFNode{
				Type:    adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{},
			},
			context: converter.ConversionContext{
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
	lic := NewListItemConverter()

	tests := []struct {
		name         string
		lines        []string
		startIndex   int
		expectedNode adf_types.ADFNode
		consumed     int
		wantErr      bool
	}{
		{
			name:       "simple list item",
			lines:      []string{"- First item"},
			startIndex: 0,
			expectedNode: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
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
			expectedNode: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
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
			expectedNode: adf_types.ADFNode{
				Type:    adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{},
			},
			consumed: 1,
			wantErr:  false,
		},
		{
			name:       "list item with extra content after",
			lines:      []string{"- First item", "- Second item"},
			startIndex: 0,
			expectedNode: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
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
			node, consumed, err := lic.FromMarkdown(tt.lines, tt.startIndex, converter.ConversionContext{})
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
				if node.Content[0].Type != adf_types.NodeTypeParagraph {
					t.Errorf("FromMarkdown() first child type = %s, want paragraph", node.Content[0].Type)
				}
			}
		})
	}
}

func TestListItemConverter_RoundTrip(t *testing.T) {
	lic := NewListItemConverter()

	tests := []struct {
		name    string
		node    adf_types.ADFNode
		context converter.ConversionContext
	}{
		{
			name: "simple list item",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Test item",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
				ListDepth: 1,
			},
		},
		{
			name: "nested list item",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeText,
								Text: "Nested test item",
							},
						},
					},
				},
			},
			context: converter.ConversionContext{
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
	lic := NewListItemConverter()

	tests := []struct {
		nodeType converter.ADFNodeType
		want     bool
	}{
		{converter.ADFNodeType(adf_types.NodeTypeListItem), true},
		{converter.ADFNodeType(adf_types.NodeTypeParagraph), false},
		{converter.ADFNodeType(adf_types.NodeTypeHeading), false},
		{converter.ADFNodeType(adf_types.NodeTypeText), false},
		{converter.ADFNodeType(adf_types.NodeTypeBulletList), false},
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
	lic := NewListItemConverter()
	strategy := lic.GetStrategy()

	if strategy != converter.StandardMarkdown {
		t.Errorf("GetStrategy() = %v, want %v", strategy, converter.StandardMarkdown)
	}
}

func TestListItemConverter_ValidateInput(t *testing.T) {
	lic := NewListItemConverter()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name: "valid list item node",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeListItem,
			},
			wantErr: false,
		},
		{
			name:    "invalid type - not ADFNode",
			input:   "not a node",
			wantErr: true,
		},
		{
			name: "invalid type - wrong node type",
			input: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
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
