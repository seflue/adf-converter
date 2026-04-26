// Package defaultclass provides the default ContentClassifier implementation.
package defaultclass

// Classifier provides predefined rules for common ADF node types.
type Classifier struct {
	editableTypes     map[string]bool
	preservedTypes    map[string]bool
	inlineFormatTypes map[string]bool
}

// New creates a classifier with standard content type rules.
func New() *Classifier {
	return &Classifier{
		editableTypes: map[string]bool{
			"paragraph":   true,
			"heading":     true,
			"text":        true,
			"hardBreak":   true,
			"orderedList": true,
			"bulletList":  true,
			"listItem":    true,
			"inlineCard":  true,
			"emoji":       true,
			"mention":     true,
			"date":        true,
			"status":      true,
			"codeBlock":   true,
			"rule":        true,
			"table":       true,
			"panel":       true,
			"blockCard":   true,
			"blockquote":  true,
			"mediaSingle": true,
		},
		preservedTypes: map[string]bool{
			"mediaInline": true,
		},
		inlineFormatTypes: map[string]bool{
			"strong":    true,
			"em":        true,
			"code":      true,
			"link":      true,
			"underline": true,
			"strike":    true,
			"subsup":    true,
		},
	}
}

// IsEditable returns true if the node type can be safely converted to Markdown
// and edited by users without losing functionality.
func (c *Classifier) IsEditable(nodeType string) bool {
	return c.editableTypes[nodeType]
}

// IsPreserved returns true if the node type should be preserved as a placeholder
// to maintain round-trip fidelity and avoid data loss.
func (c *Classifier) IsPreserved(nodeType string) bool {
	return c.preservedTypes[nodeType]
}

// IsInlineFormattable returns true if the mark type can be converted to
// standard Markdown inline formatting (bold, italic, code, etc.).
func (c *Classifier) IsInlineFormattable(nodeType string) bool {
	return c.inlineFormatTypes[nodeType]
}
