// Package adftypes holds the core ADF data structures (Node, Document,
// Mark) and node-type constants. Lives as an internal leaf package so that
// peer libraries (placeholder/) can depend on it without forming a cycle with
// the parent adf/ package, which would arise if these types lived directly in
// adf/.
//
// The public adf package re-exports every type and constant from here via type
// aliases and const declarations.
package adftypes

import (
	"encoding/json"
	"fmt"
)

// Document represents the root ADF document structure
type Document struct {
	Version int       `json:"version"`
	Type    string    `json:"type"` // Always "doc"
	Content []Node `json:"content"`
}

// Node represents any node in the ADF tree structure
type Node struct {
	Type    string         `json:"type"`
	Attrs   map[string]any `json:"attrs,omitempty"`
	Content []Node      `json:"content,omitempty"`
	Marks   []Mark      `json:"marks,omitempty"`
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
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// Common ADF node types based on Atlassian specification
const (
	// Block nodes
	NodeTypeDoc          = "doc"
	NodeTypeParagraph    = "paragraph"
	NodeTypeHeading      = "heading"
	NodeTypeCodeBlock    = "codeBlock"
	NodeTypeTable        = "table"
	NodeTypeTableRow     = "tableRow"
	NodeTypeTableCell    = "tableCell"
	NodeTypePanel        = "panel"
	NodeTypeBlockquote   = "blockquote"
	NodeTypeRule         = "rule"
	NodeTypeMedia        = "media"
	NodeTypeMediaSingle  = "mediaSingle"
	NodeTypeMediaInline  = "mediaInline"
	NodeTypeOrderedList  = "orderedList"
	NodeTypeBulletList   = "bulletList"
	NodeTypeListItem     = "listItem"
	NodeTypeExpand       = "expand"
	NodeTypeNestedExpand = "nestedExpand"
	NodeTypeBlockCard    = "blockCard"

	// Inline nodes
	NodeTypeText       = "text"
	NodeTypeHardBreak  = "hardBreak"
	NodeTypeInlineCard = "inlineCard"
	NodeTypeMention    = "mention"
	NodeTypeDate       = "date"
	NodeTypeEmoji      = "emoji"
	NodeTypeStatus     = "status"

	// Mark types for text formatting
	MarkTypeStrong    = "strong"
	MarkTypeEm        = "em"
	MarkTypeCode      = "code"
	MarkTypeLink      = "link"
	MarkTypeTextColor = "textColor"
	MarkTypeUnderline = "underline"
	MarkTypeStrike    = "strike"
	MarkTypeSubsup    = "subsup"
)

// IsInlineNode returns true if the node type is an inline node
func IsInlineNode(nodeType string) bool {
	inlineTypes := map[string]bool{
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
func NewMark(markType string, attrs map[string]any) Mark {
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
