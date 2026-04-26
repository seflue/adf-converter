// Package defaultclass provides the default ContentClassifier implementation.
package defaultclass

import "github.com/seflue/adf-converter/adf/adftypes"

// Classifier provides predefined rules for common ADF node types.
type Classifier struct {
	editableTypes     map[adftypes.NodeType]bool
	preservedTypes    map[adftypes.NodeType]bool
	inlineFormatTypes map[adftypes.MarkType]bool
}

// New creates a classifier with standard content type rules.
func New() *Classifier {
	return &Classifier{
		editableTypes: map[adftypes.NodeType]bool{
			adftypes.NodeTypeParagraph:   true,
			adftypes.NodeTypeHeading:     true,
			adftypes.NodeTypeText:        true,
			adftypes.NodeTypeHardBreak:   true,
			adftypes.NodeTypeOrderedList: true,
			adftypes.NodeTypeBulletList:  true,
			adftypes.NodeTypeListItem:    true,
			adftypes.NodeTypeInlineCard:  true,
			adftypes.NodeTypeEmoji:       true,
			adftypes.NodeTypeMention:     true,
			adftypes.NodeTypeDate:        true,
			adftypes.NodeTypeStatus:      true,
			adftypes.NodeTypeCodeBlock:   true,
			adftypes.NodeTypeRule:        true,
			adftypes.NodeTypeTable:       true,
			adftypes.NodeTypePanel:       true,
			adftypes.NodeTypeBlockCard:   true,
			adftypes.NodeTypeBlockquote:  true,
			adftypes.NodeTypeMediaSingle: true,
		},
		preservedTypes: map[adftypes.NodeType]bool{
			adftypes.NodeTypeMediaInline: true,
		},
		inlineFormatTypes: map[adftypes.MarkType]bool{
			adftypes.MarkTypeStrong:    true,
			adftypes.MarkTypeEm:        true,
			adftypes.MarkTypeCode:      true,
			adftypes.MarkTypeLink:      true,
			adftypes.MarkTypeUnderline: true,
			adftypes.MarkTypeStrike:    true,
			adftypes.MarkTypeSubsup:    true,
		},
	}
}

// IsEditable returns true if the node type can be safely converted to Markdown
// and edited by users without losing functionality.
func (c *Classifier) IsEditable(nodeType adftypes.NodeType) bool {
	return c.editableTypes[nodeType]
}

// IsPreserved returns true if the node type should be preserved as a placeholder
// to maintain round-trip fidelity and avoid data loss.
func (c *Classifier) IsPreserved(nodeType adftypes.NodeType) bool {
	return c.preservedTypes[nodeType]
}

// IsInlineFormattable returns true if the mark type can be converted to
// standard Markdown inline formatting (bold, italic, code, etc.).
func (c *Classifier) IsInlineFormattable(markType adftypes.MarkType) bool {
	return c.inlineFormatTypes[markType]
}
