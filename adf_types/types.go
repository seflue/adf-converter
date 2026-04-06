package adf_types

import (
	"encoding/json"
	"fmt"
)

// ADFDocument represents the root ADF document structure
type ADFDocument struct {
	Version int       `json:"version"`
	Type    string    `json:"type"` // Always "doc"
	Content []ADFNode `json:"content"`
}

// ADFNode represents any node in the ADF tree structure
type ADFNode struct {
	Type    string                 `json:"type"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Content []ADFNode              `json:"content,omitempty"`
	Marks   []ADFMark              `json:"marks,omitempty"`
	Text    string                 `json:"text,omitempty"` // Only for text nodes
}

// MarshalJSON ensures text nodes always include the "text" field, even when empty.
// ADF spec requires "text" on text nodes but not on other node types.
func (n ADFNode) MarshalJSON() ([]byte, error) {
	type Alias ADFNode

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

// ADFMark represents text formatting marks (bold, italic, etc.)
type ADFMark struct {
	Type  string                 `json:"type"`
	Attrs map[string]interface{} `json:"attrs,omitempty"`
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

// IsBlockNode returns true if the node type is a block-level node
func IsBlockNode(nodeType string) bool {
	blockTypes := map[string]bool{
		NodeTypeDoc:         true,
		NodeTypeParagraph:   true,
		NodeTypeHeading:     true,
		NodeTypeCodeBlock:   true,
		NodeTypeTable:       true,
		NodeTypeTableRow:    true,
		NodeTypeTableCell:   true,
		NodeTypePanel:       true,
		NodeTypeBlockquote:  true,
		NodeTypeRule:        true,
		NodeTypeMediaSingle: true,
		NodeTypeOrderedList: true,
		NodeTypeBulletList:  true,
		NodeTypeListItem:    true,
		NodeTypeExpand:      true,
	}
	return blockTypes[nodeType]
}

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
func (d ADFDocument) String() string {
	return fmt.Sprintf("ADFDocument{Version: %d, Type: %s, Content: %d nodes}",
		d.Version, d.Type, len(d.Content))
}

// String provides a human-readable representation of an ADF node
func (n ADFNode) String() string {
	if n.Text != "" {
		return fmt.Sprintf("ADFNode{Type: %s, Text: %.20s...}", n.Type, n.Text)
	}
	return fmt.Sprintf("ADFNode{Type: %s, Content: %d nodes, Marks: %d}",
		n.Type, len(n.Content), len(n.Marks))
}

// NewDocument creates a new ADF document with the standard version
func NewDocument() *ADFDocument {
	return &ADFDocument{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: make([]ADFNode, 0),
	}
}

// NewTextNode creates a text node with the given content
func NewTextNode(text string, marks ...ADFMark) ADFNode {
	return ADFNode{
		Type:  NodeTypeText,
		Text:  text,
		Marks: marks,
	}
}

// NewParagraphNode creates a paragraph node with the given content
func NewParagraphNode(content ...ADFNode) ADFNode {
	return ADFNode{
		Type:    NodeTypeParagraph,
		Content: content,
	}
}

// NewHeadingNode creates a heading node with the specified level (1-6)
func NewHeadingNode(level int, content ...ADFNode) ADFNode {
	attrs := map[string]interface{}{
		"level": level,
	}
	return ADFNode{
		Type:    NodeTypeHeading,
		Attrs:   attrs,
		Content: content,
	}
}

// NewMark creates a new text mark
func NewMark(markType string, attrs map[string]interface{}) ADFMark {
	return ADFMark{
		Type:  markType,
		Attrs: attrs,
	}
}

// GetHeadingLevel returns the heading level from a heading node's attributes
func (n ADFNode) GetHeadingLevel() int {
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
func (n ADFNode) GetAttribute(key string) (interface{}, bool) {
	if n.Attrs == nil {
		return nil, false
	}
	value, exists := n.Attrs[key]
	return value, exists
}

// SetAttribute sets an attribute value
func (n *ADFNode) SetAttribute(key string, value interface{}) {
	if n.Attrs == nil {
		n.Attrs = make(map[string]interface{})
	}
	n.Attrs[key] = value
}
