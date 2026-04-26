package adf

import (
	"github.com/seflue/adf-converter/adf/adftypes"
)

// Re-exports of core ADF data types from internal/adftypes. The split exists
// because placeholder/ depends on these types and adf/ depends on placeholder/;
// keeping the types in a leaf package avoids the import cycle. Public-API
// surface of adf.* is unchanged.

// Document represents the root ADF document structure.
type Document = adftypes.Document

// Node represents any node in the ADF tree structure.
type Node = adftypes.Node

// Mark represents text formatting marks (bold, italic, etc.).
type Mark = adftypes.Mark

const (
	NodeTypeDoc          = adftypes.NodeTypeDoc
	NodeTypeParagraph    = adftypes.NodeTypeParagraph
	NodeTypeHeading      = adftypes.NodeTypeHeading
	NodeTypeCodeBlock    = adftypes.NodeTypeCodeBlock
	NodeTypeTable        = adftypes.NodeTypeTable
	NodeTypeTableRow     = adftypes.NodeTypeTableRow
	NodeTypeTableCell    = adftypes.NodeTypeTableCell
	NodeTypePanel        = adftypes.NodeTypePanel
	NodeTypeBlockquote   = adftypes.NodeTypeBlockquote
	NodeTypeRule         = adftypes.NodeTypeRule
	NodeTypeMedia        = adftypes.NodeTypeMedia
	NodeTypeMediaSingle  = adftypes.NodeTypeMediaSingle
	NodeTypeMediaInline  = adftypes.NodeTypeMediaInline
	NodeTypeOrderedList  = adftypes.NodeTypeOrderedList
	NodeTypeBulletList   = adftypes.NodeTypeBulletList
	NodeTypeListItem     = adftypes.NodeTypeListItem
	NodeTypeExpand       = adftypes.NodeTypeExpand
	NodeTypeNestedExpand = adftypes.NodeTypeNestedExpand
	NodeTypeBlockCard    = adftypes.NodeTypeBlockCard

	NodeTypeText       = adftypes.NodeTypeText
	NodeTypeHardBreak  = adftypes.NodeTypeHardBreak
	NodeTypeInlineCard = adftypes.NodeTypeInlineCard
	NodeTypeMention    = adftypes.NodeTypeMention
	NodeTypeDate       = adftypes.NodeTypeDate
	NodeTypeEmoji      = adftypes.NodeTypeEmoji
	NodeTypeStatus     = adftypes.NodeTypeStatus

	MarkTypeStrong    = adftypes.MarkTypeStrong
	MarkTypeEm        = adftypes.MarkTypeEm
	MarkTypeCode      = adftypes.MarkTypeCode
	MarkTypeLink      = adftypes.MarkTypeLink
	MarkTypeTextColor = adftypes.MarkTypeTextColor
	MarkTypeUnderline = adftypes.MarkTypeUnderline
	MarkTypeStrike    = adftypes.MarkTypeStrike
	MarkTypeSubsup    = adftypes.MarkTypeSubsup
)

// IsInlineNode returns true if the node type is an inline node.
func IsInlineNode(nodeType string) bool { return adftypes.IsInlineNode(nodeType) }

// NewDocument creates a new ADF document with the standard version.
func NewDocument() *Document { return adftypes.NewDocument() }

// NewTextNode creates a text node with the given content.
func NewTextNode(text string, marks ...Mark) Node {
	return adftypes.NewTextNode(text, marks...)
}

// NewParagraphNode creates a paragraph node with the given content.
func NewParagraphNode(content ...Node) Node { return adftypes.NewParagraphNode(content...) }

// NewMark creates a new text mark.
func NewMark(markType string, attrs map[string]any) Mark { return adftypes.NewMark(markType, attrs) }
