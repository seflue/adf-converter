package inline

import (
	"testing"

	"adf-converter/adf_types"
)

func TestParseContent_PlainText(t *testing.T) {
	nodes, err := ParseContent("plain text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Type != adf_types.NodeTypeText {
		t.Errorf("expected text node, got %s", nodes[0].Type)
	}

	if nodes[0].Text != "plain text" {
		t.Errorf("expected 'plain text', got '%s'", nodes[0].Text)
	}
}

func TestParseContent_BoldText(t *testing.T) {
	nodes, err := ParseContent("**bold**")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "bold" {
		t.Errorf("expected 'bold', got '%s'", nodes[0].Text)
	}

	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != "strong" {
		t.Errorf("expected strong mark")
	}
}

func TestParseContent_ItalicText(t *testing.T) {
	nodes, err := ParseContent("*italic*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "italic" {
		t.Errorf("expected 'italic', got '%s'", nodes[0].Text)
	}

	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != "em" {
		t.Errorf("expected em mark")
	}
}

func TestParseContent_CodeText(t *testing.T) {
	nodes, err := ParseContent("`code`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "code" {
		t.Errorf("expected 'code', got '%s'", nodes[0].Text)
	}

	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != "code" {
		t.Errorf("expected code mark")
	}
}

func TestParseContent_Links(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantText string
		wantHref string
	}{
		{
			name:     "simple link",
			input:    "[text](url)",
			wantText: "text",
			wantHref: "url",
		},
		{
			name:     "InlineCard (same text and url)",
			input:    "[https://example.com](https://example.com)",
			wantText: "",
			wantHref: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d", len(nodes))
			}

			if tt.input == "[https://example.com](https://example.com)" {
				// Should be InlineCard
				if nodes[0].Type != adf_types.NodeTypeInlineCard {
					t.Errorf("expected InlineCard node, got %s", nodes[0].Type)
				}
			} else {
				// Should be text with link mark
				if nodes[0].Text != tt.wantText {
					t.Errorf("expected text '%s', got '%s'", tt.wantText, nodes[0].Text)
				}
			}
		})
	}
}

func TestParseContent_MixedFormatting(t *testing.T) {
	nodes, err := ParseContent("plain **bold** *italic* `code` plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 7 {
		t.Fatalf("expected 7 nodes, got %d", len(nodes))
	}

	// Verify sequence
	if nodes[0].Text != "plain " {
		t.Errorf("expected 'plain ', got '%s'", nodes[0].Text)
	}
	if nodes[1].Text != "bold" || len(nodes[1].Marks) == 0 || nodes[1].Marks[0].Type != "strong" {
		t.Errorf("expected bold node")
	}
	if nodes[2].Text != " " {
		t.Errorf("expected ' ', got '%s'", nodes[2].Text)
	}
	if nodes[3].Text != "italic" || len(nodes[3].Marks) == 0 || nodes[3].Marks[0].Type != "em" {
		t.Errorf("expected italic node")
	}
	if nodes[4].Text != " " {
		t.Errorf("expected ' ', got '%s'", nodes[4].Text)
	}
	if nodes[5].Text != "code" || len(nodes[5].Marks) == 0 || nodes[5].Marks[0].Type != "code" {
		t.Errorf("expected code node")
	}
	if nodes[6].Text != " plain" {
		t.Errorf("expected ' plain', got '%s'", nodes[6].Text)
	}
}

func TestParseContent_NestedFormatting(t *testing.T) {
	// Note: Current implementation doesn't support true nesting,
	// but we should handle adjacent formatting correctly
	nodes, err := ParseContent("**bold*italic***")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should parse as bold "bold*italic*" (asterisks inside are treated as content)
	if len(nodes) == 0 {
		t.Fatal("expected at least one node")
	}
}

// Comprehensive goldmark-based tests
func TestParseContent_Goldmark(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, nodes []adf_types.ADFNode)
	}{
		{
			name:  "nested_bold_italic",
			input: "This is **bold with *italic* inside** text",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				// Goldmark produces 5 text nodes:
				// 1. "This is " (plain)
				// 2. "bold with " (strong)
				// 3. "italic" (strong+em)
				// 4. " inside" (strong)
				// 5. " text" (plain)
				if len(nodes) != 5 {
					t.Fatalf("expected 5 nodes, got %d", len(nodes))
				}
				// Check for proper nesting
				foundNested := false
				for _, node := range nodes {
					if len(node.Marks) == 2 {
						// Should have both strong and em marks
						foundNested = true
						hasStrong := false
						hasEm := false
						for _, mark := range node.Marks {
							if mark.Type == "strong" {
								hasStrong = true
							}
							if mark.Type == "em" {
								hasEm = true
							}
						}
						if !hasStrong || !hasEm {
							t.Errorf("nested node should have both strong and em marks")
						}
					}
				}
				if !foundNested {
					t.Errorf("should have a node with nested marks")
				}
			},
		},
		{
			name:  "inline_card_detection",
			input: "Check [https://example.com](https://example.com) out",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				// Should have InlineCard node
				found := false
				for _, node := range nodes {
					if node.Type == adf_types.NodeTypeInlineCard {
						found = true
						if node.Attrs["url"] != "https://example.com" {
							t.Errorf("InlineCard should have url attribute")
						}
					}
				}
				if !found {
					t.Errorf("Should have InlineCard node")
				}
			},
		},
		{
			name:  "link_with_formatted_text",
			input: "[**bold link**](https://example.com)",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				if len(nodes) != 1 {
					t.Fatalf("expected 1 node, got %d", len(nodes))
				}
				// Should have both link and strong marks
				if len(nodes[0].Marks) != 2 {
					t.Errorf("should have 2 marks (link + strong), got %d", len(nodes[0].Marks))
				}
			},
		},
		{
			name:  "text_node_merging",
			input: "Hello, world!",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				// Should be single merged text node (goldmark splits at !)
				if len(nodes) != 1 {
					t.Errorf("expected 1 merged node, got %d", len(nodes))
				}
				if nodes[0].Text != "Hello, world!" {
					t.Errorf("expected 'Hello, world!', got '%s'", nodes[0].Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, nodes)
		})
	}
}
