// Package adftypes holds the core ADF data structures (Node, Document,
// Mark) and node-type / mark-type constants. Lives as an internal leaf
// package so that peer libraries (placeholder/) can depend on it without
// forming a cycle with the parent adf/ package, which would arise if these
// types lived directly in adf/.
//
// The public adf package re-exports every type and constant from here via type
// aliases and const declarations.
package adftypes

import (
	"encoding/json"
	"fmt"
)

// NodeType is the typed enumeration of ADF node types. Underlying type is
// string so JSON marshalling is transparent.
type NodeType string

// MarkType is the typed enumeration of ADF mark (text formatting) types.
// Underlying type is string so JSON marshalling is transparent.
type MarkType string

// Document represents the root ADF document structure
type Document struct {
	Version int      `json:"version"`
	Type    NodeType `json:"type"` // Always "doc"
	Content []Node   `json:"content"`
}

// Node represents any node in the ADF tree structure
type Node struct {
	Type    NodeType       `json:"type"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []Node         `json:"content,omitempty"`
	Marks   []Mark         `json:"marks,omitempty"`
	Text    string         `json:"text,omitempty"` // Only for text nodes
}

// MarshalJSON ensures text nodes always include the "text" field, even when empty.
// ADF spec requires "text" on text nodes but not on other node types.
func (n Node) MarshalJSON() ([]byte, error) {
	type Alias Node

	if n.Type == NodeTypeText {
		return json.Marshal(struct {
			Alias
			Text string `json:"text"`
		}{
			Alias: Alias(n),
			Text:  n.Text,
		})
	}

	return json.Marshal(Alias(n))
}

// Mark represents text formatting marks (bold, italic, etc.)
type Mark struct {
	Type  MarkType       `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// Common ADF node types based on Atlassian specification
const (
	// Block nodes
	NodeTypeDoc          NodeType = "doc"
	NodeTypeParagraph    NodeType = "paragraph"
	NodeTypeHeading      NodeType = "heading"
	NodeTypeCodeBlock    NodeType = "codeBlock"
	NodeTypeTable        NodeType = "table"
	NodeTypeTableRow     NodeType = "tableRow"
	NodeTypeTableCell    NodeType = "tableCell"
	NodeTypeTableHeader  NodeType = "tableHeader"
	NodeTypeTaskList     NodeType = "taskList"
	NodeTypeTaskItem     NodeType = "taskItem"
	NodeTypePanel        NodeType = "panel"
	NodeTypeBlockquote   NodeType = "blockquote"
	NodeTypeRule         NodeType = "rule"
	NodeTypeMedia        NodeType = "media"
	NodeTypeMediaSingle  NodeType = "mediaSingle"
	NodeTypeMediaInline  NodeType = "mediaInline"
	NodeTypeOrderedList  NodeType = "orderedList"
	NodeTypeBulletList   NodeType = "bulletList"
	NodeTypeListItem     NodeType = "listItem"
	NodeTypeExpand       NodeType = "expand"
	NodeTypeNestedExpand NodeType = "nestedExpand"
	NodeTypeBlockCard    NodeType = "blockCard"

	// Inline nodes
	NodeTypeText       NodeType = "text"
	NodeTypeHardBreak  NodeType = "hardBreak"
	NodeTypeInlineCard NodeType = "inlineCard"
	NodeTypeMention    NodeType = "mention"
	NodeTypeDate       NodeType = "date"
	NodeTypeEmoji      NodeType = "emoji"
	NodeTypeStatus     NodeType = "status"
)

// Mark types for text formatting.
const (
	MarkTypeStrong    MarkType = "strong"
	MarkTypeEm        MarkType = "em"
	MarkTypeCode      MarkType = "code"
	MarkTypeLink      MarkType = "link"
	MarkTypeTextColor MarkType = "textColor"
	MarkTypeUnderline MarkType = "underline"
	MarkTypeStrike    MarkType = "strike"
	MarkTypeSubsup    MarkType = "subsup"
)

// IsInlineNode returns true if the node type is an inline node
func IsInlineNode(nodeType NodeType) bool {
	inlineTypes := map[NodeType]bool{
		NodeTypeText:        true,
		NodeTypeHardBreak:   true,
		NodeTypeMention:     true,
		NodeTypeDate:        true,
		NodeTypeEmoji:       true,
		NodeTypeStatus:      true,
		NodeTypeInlineCard:  true,
		NodeTypeMediaInline: true,
	}
	return inlineTypes[nodeType]
}

// String provides a human-readable representation of an ADF document
func (d Document) String() string {
	return fmt.Sprintf("Document{Version: %d, Type: %s, Content: %d nodes}",
		d.Version, d.Type, len(d.Content))
}

// String provides a human-readable representation of an ADF node
func (n Node) String() string {
	if n.Text != "" {
		return fmt.Sprintf("Node{Type: %s, Text: %.20s...}", n.Type, n.Text)
	}
	return fmt.Sprintf("Node{Type: %s, Content: %d nodes, Marks: %d}",
		n.Type, len(n.Content), len(n.Marks))
}

// NewDocument creates a new ADF document with the standard version
func NewDocument() *Document {
	return &Document{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: make([]Node, 0),
	}
}

// NewTextNode creates a text node with the given content
func NewTextNode(text string, marks ...Mark) Node {
	return Node{
		Type:  NodeTypeText,
		Text:  text,
		Marks: marks,
	}
}

// NewParagraphNode creates a paragraph node with the given content
func NewParagraphNode(content ...Node) Node {
	return Node{
		Type:    NodeTypeParagraph,
		Content: content,
	}
}

// NewMark creates a new text mark
func NewMark(markType MarkType, attrs map[string]any) Mark {
	return Mark{
		Type:  markType,
		Attrs: attrs,
	}
}

// GetHeadingLevel returns the heading level from a heading node's attributes
func (n Node) GetHeadingLevel() int {
	if n.Type != NodeTypeHeading {
		return 0
	}
	if n.Attrs == nil {
		return 1 // Default level
	}

	// Handle both int (from Go code) and float64 (from JSON unmarshaling)
	switch level := n.Attrs["level"].(type) {
	case int:
		return level
	case float64:
		return int(level)
	}

	return 1 // Default level
}

// GetAttribute returns an attribute value and whether it exists
func (n Node) GetAttribute(key string) (any, bool) {
	if n.Attrs == nil {
		return nil, false
	}
	value, exists := n.Attrs[key]
	return value, exists
}

// SetAttribute sets an attribute value
func (n *Node) SetAttribute(key string, value any) {
	if n.Attrs == nil {
		n.Attrs = make(map[string]any)
	}
	n.Attrs[key] = value
}
