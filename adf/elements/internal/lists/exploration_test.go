package lists_test

import (
	"fmt"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// This test explores goldmark's list AST structure to understand what we need to convert
func TestExplore_GoldmarkListAST(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
	}{
		{
			name: "simple bullet list",
			markdown: `- Item 1
- Item 2
- Item 3`,
		},
		{
			name: "multi-line list item",
			markdown: `- First line
  continued line
  still same item
- Second item`,
		},
		{
			name: "nested bullet list",
			markdown: `- Level 1
  - Level 2
    - Level 3
  - Back to Level 2
- Level 1 again`,
		},
		{
			name: "ordered list",
			markdown: `1. First
2. Second
3. Third`,
		},
		{
			name: "mixed nesting",
			markdown: `- Bullet item
  1. Ordered inside
  2. Another ordered
- Another bullet`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.markdown)
			parser := goldmark.New()
			doc := parser.Parser().Parse(text.NewReader(source))

			fmt.Printf("\n=== %s ===\n", tt.name)
			fmt.Printf("Markdown:\n%s\n\n", tt.markdown)
			fmt.Println("Goldmark AST:")
			dumpAST(doc, source, 0)
		})
	}
}

// dumpAST recursively prints the AST structure
func dumpAST(node ast.Node, source []byte, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch n := node.(type) {
	case *ast.Document:
		fmt.Printf("%sDocument\n", indent)

	case *ast.List:
		listType := "bullet"
		if n.IsOrdered() {
			listType = "ordered"
		}
		fmt.Printf("%sList (type=%s, marker='%c', tight=%v, start=%d)\n",
			indent, listType, n.Marker, n.IsTight, n.Start)

	case *ast.ListItem:
		fmt.Printf("%sListItem\n", indent)

	case *ast.Paragraph:
		fmt.Printf("%sParagraph\n", indent)

	case *ast.Text:
		segment := n.Segment
		txt := string(source[segment.Start:segment.Stop])
		fmt.Printf("%sText: %q\n", indent, txt)

	case *ast.TextBlock:
		fmt.Printf("%sTextBlock\n", indent)

	default:
		fmt.Printf("%s%T\n", indent, node)
	}

	// Recurse to children
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		dumpAST(child, source, depth+1)
	}
}

// This test shows what the target ADF structure should look like
func TestExplore_TargetADFStructure(t *testing.T) {
	t.Skip("Reference only - shows expected ADF output structure")

	// For this markdown:
	// - First line
	//   continued line
	// - Second item

	// Expected ADF structure:
	/*
		{
			Type: "bulletList",
			Content: [
				{
					Type: "listItem",
					Content: [
						{
							Type: "paragraph",
							Content: [
								{Type: "text", Text: "First line continued line"}
							]
						}
					]
				},
				{
					Type: "listItem",
					Content: [
						{
							Type: "paragraph",
							Content: [
								{Type: "text", Text: "Second item"}
							]
						}
					]
				}
			]
		}
	*/

	// For nested lists:
	// - Level 1
	//   - Level 2

	// Expected ADF structure:
	/*
		{
			Type: "bulletList",
			Content: [
				{
					Type: "listItem",
					Content: [
						{
							Type: "paragraph",
							Content: [{Type: "text", Text: "Level 1"}]
						},
						{
							Type: "bulletList",  // Nested list is a child of listItem
							Content: [
								{
									Type: "listItem",
									Content: [
										{
											Type: "paragraph",
											Content: [{Type: "text", Text: "Level 2"}]
										}
									]
								}
							]
						}
					]
				}
			]
		}
	*/
}
