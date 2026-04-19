package adf_types

import (
	"testing"
)

func TestIsInlineNode(t *testing.T) {
	tests := []struct {
		nodeType string
		expected bool
	}{
		{NodeTypeText, true},
		{NodeTypeHardBreak, true},
		{NodeTypeMention, true},
		{NodeTypeStatus, true},
		{NodeTypeMediaInline, true},
		{NodeTypeParagraph, false},
		{NodeTypeHeading, false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			result := IsInlineNode(tt.nodeType)
			if result != tt.expected {
				t.Errorf("IsInlineNode(%s) = %v, want %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}

func TestNewDocument(t *testing.T) {
	doc := NewDocument()

	if doc.Version != 1 {
		t.Errorf("NewDocument() version = %d, want 1", doc.Version)
	}

	if doc.Type != NodeTypeDoc {
		t.Errorf("NewDocument() type = %s, want %s", doc.Type, NodeTypeDoc)
	}

	if doc.Content == nil {
		t.Error("NewDocument() content should not be nil")
	}

	if len(doc.Content) != 0 {
		t.Errorf("NewDocument() content length = %d, want 0", len(doc.Content))
	}
}

func TestNewTextNode(t *testing.T) {
	text := "Hello, world!"
	mark := NewMark(MarkTypeStrong, nil)

	node := NewTextNode(text, mark)

	if node.Type != NodeTypeText {
		t.Errorf("NewTextNode() type = %s, want %s", node.Type, NodeTypeText)
	}

	if node.Text != text {
		t.Errorf("NewTextNode() text = %s, want %s", node.Text, text)
	}

	if len(node.Marks) != 1 {
		t.Errorf("NewTextNode() marks length = %d, want 1", len(node.Marks))
	}

	if node.Marks[0].Type != MarkTypeStrong {
		t.Errorf("NewTextNode() mark type = %s, want %s", node.Marks[0].Type, MarkTypeStrong)
	}
}

func TestNewParagraphNode(t *testing.T) {
	textNode := NewTextNode("Hello")

	para := NewParagraphNode(textNode)

	if para.Type != NodeTypeParagraph {
		t.Errorf("NewParagraphNode() type = %s, want %s", para.Type, NodeTypeParagraph)
	}

	if len(para.Content) != 1 {
		t.Errorf("NewParagraphNode() content length = %d, want 1", len(para.Content))
	}

	if para.Content[0].Text != "Hello" {
		t.Errorf("NewParagraphNode() content text = %s, want Hello", para.Content[0].Text)
	}
}

func TestGetHeadingLevel(t *testing.T) {
	tests := []struct {
		name     string
		node     ADFNode
		expected int
	}{
		{
			name: "valid heading with level 3",
			node: ADFNode{
				Type:    NodeTypeHeading,
				Attrs:   map[string]any{"level": 3},
				Content: []ADFNode{NewTextNode("Test")},
			},
			expected: 3,
		},
		{
			name: "heading with no attrs",
			node: ADFNode{
				Type: NodeTypeHeading,
			},
			expected: 1, // Default level
		},
		{
			name: "non-heading node",
			node: ADFNode{
				Type: NodeTypeParagraph,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := tt.node.GetHeadingLevel()
			if level != tt.expected {
				t.Errorf("GetHeadingLevel() = %d, want %d", level, tt.expected)
			}
		})
	}
}

func TestNewMark(t *testing.T) {
	markType := MarkTypeLink
	attrs := map[string]any{
		"href": "https://example.com",
	}

	mark := NewMark(markType, attrs)

	if mark.Type != markType {
		t.Errorf("NewMark() type = %s, want %s", mark.Type, markType)
	}

	if mark.Attrs["href"] != attrs["href"] {
		t.Errorf("NewMark() href = %v, want %v", mark.Attrs["href"], attrs["href"])
	}
}

func TestADFDocumentString(t *testing.T) {
	doc := NewDocument()
	doc.Content = append(doc.Content, NewParagraphNode(NewTextNode("test")))

	str := doc.String()
	expected := "ADFDocument{Version: 1, Type: doc, Content: 1 nodes}"

	if str != expected {
		t.Errorf("ADFDocument.String() = %s, want %s", str, expected)
	}
}

func TestADFNodeString(t *testing.T) {
	tests := []struct {
		name     string
		node     ADFNode
		contains string
	}{
		{
			name:     "text node",
			node:     NewTextNode("Hello, world!"),
			contains: "Type: text, Text: Hello, world!",
		},
		{
			name:     "paragraph node",
			node:     NewParagraphNode(NewTextNode("test")),
			contains: "Type: paragraph, Content: 1 nodes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.node.String()
			if str != tt.contains && len(str) < len(tt.contains) {
				t.Errorf("ADFNode.String() = %s, should contain %s", str, tt.contains)
			}
		})
	}
}
