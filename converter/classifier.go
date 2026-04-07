package converter

import "adf-converter/adf_types"

// ContentClassifier determines how different ADF node types should be handled
type ContentClassifier interface {
	IsEditable(nodeType string) bool
	IsPreserved(nodeType string) bool
	IsInlineFormattable(nodeType string) bool
}

// DefaultClassifier provides predefined rules for common ADF node types
type DefaultClassifier struct {
	editableTypes     map[string]bool
	preservedTypes    map[string]bool
	inlineFormatTypes map[string]bool
}

// NewDefaultClassifier creates a classifier with standard content type rules
func NewDefaultClassifier() ContentClassifier {
	return &DefaultClassifier{
		editableTypes: map[string]bool{
			adf_types.NodeTypeParagraph: true,
			adf_types.NodeTypeHeading:   true,
			adf_types.NodeTypeText:      true,
			adf_types.NodeTypeHardBreak: true,
			// Simple lists can be editable
			adf_types.NodeTypeOrderedList: true,
			adf_types.NodeTypeBulletList:  true,
			adf_types.NodeTypeListItem:    true,
			// InlineCard can be converted to markdown links
			adf_types.NodeTypeInlineCard: true,
			// Emoji can be edited as unicode characters
			adf_types.NodeTypeEmoji: true,
			// Mention can be edited as [@DisplayName](accountid:id)
			adf_types.NodeTypeMention: true,
			// Date can be edited as [date:2025-04-04]
			adf_types.NodeTypeDate: true,
			// Status can be edited as [status:Text|color]
			adf_types.NodeTypeStatus: true,
			// CodeBlock can be converted to fenced code blocks
			adf_types.NodeTypeCodeBlock: true,
			// Rule is a simple block node (no attrs, no children)
			adf_types.NodeTypeRule: true,
			// Table can be converted to markdown pipe tables
			adf_types.NodeTypeTable: true,
			// Panel can be converted to fenced-div syntax (:::type content :::)
			adf_types.NodeTypePanel: true,
			// BlockCard is a block-level smart link
			adf_types.NodeTypeBlockCard: true,
			// Blockquote can be converted to markdown > syntax
			adf_types.NodeTypeBlockquote: true,
			// mediaSingle: external images → ![alt](url), internal → placeholder (handled in converter)
			adf_types.NodeTypeMediaSingle: true,
		},
		preservedTypes: map[string]bool{
			// mediaInline is preserved as placeholder (no editable representation)
			adf_types.NodeTypeMediaInline: true,
		},
		inlineFormatTypes: map[string]bool{
			adf_types.MarkTypeStrong:    true,
			adf_types.MarkTypeEm:        true,
			adf_types.MarkTypeCode:      true,
			adf_types.MarkTypeLink:      true,
			adf_types.MarkTypeUnderline: true,
			adf_types.MarkTypeStrike:    true,
			adf_types.MarkTypeSubsup:   true,
		},
	}
}

// IsEditable returns true if the node type can be safely converted to Markdown
// and edited by users without losing functionality
func (c *DefaultClassifier) IsEditable(nodeType string) bool {
	return c.editableTypes[nodeType]
}

// IsPreserved returns true if the node type should be preserved as a placeholder
// to maintain round-trip fidelity and avoid data loss
func (c *DefaultClassifier) IsPreserved(nodeType string) bool {
	return c.preservedTypes[nodeType]
}

// IsInlineFormattable returns true if the mark type can be converted to
// standard Markdown inline formatting (bold, italic, code, etc.)
func (c *DefaultClassifier) IsInlineFormattable(nodeType string) bool {
	return c.inlineFormatTypes[nodeType]
}

// GetContentStrategy returns the recommended strategy for handling a node
type ContentStrategy int

const (
	StrategyEdit ContentStrategy = iota
	StrategyPreserve
	StrategyUnknown
)

// GetContentStrategy determines the best strategy for handling a specific node
func (c *DefaultClassifier) GetContentStrategy(nodeType string) ContentStrategy {
	if c.IsEditable(nodeType) {
		return StrategyEdit
	}
	if c.IsPreserved(nodeType) {
		return StrategyPreserve
	}
	return StrategyUnknown
}

// String returns a human-readable description of the content strategy
func (s ContentStrategy) String() string {
	switch s {
	case StrategyEdit:
		return "edit"
	case StrategyPreserve:
		return "preserve"
	case StrategyUnknown:
		return "unknown"
	default:
		return "invalid"
	}
}
