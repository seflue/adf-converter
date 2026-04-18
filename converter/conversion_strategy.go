package converter

// ConversionStrategy defines how a link should be converted to markdown format
// This pattern establishes the foundation for all future ADF element conversion strategies
type ConversionStrategy int

const (
	// Existing strategies (maintain compatibility)
	StandardMarkdown ConversionStrategy = iota // [text](url) - Standard markdown format
	HTMLWrapped                                // <a meta="value">[text](url)</a> - HTML wrapper with metadata
	Placeholder                                // Preserve as placeholder - For complex cases requiring special handling

	// New markdown-native strategies for enhanced ADF support
	MarkdownTable      // Standard markdown table syntax
	MarkdownTaskList   // GitHub-style checkboxes (- [ ] / - [x])
	MarkdownBlockquote // Standard markdown blockquote (> text)
	MarkdownCodeBlock  // Code fence syntax (``` code ```)

	// New XML-preserved strategy for ADF-specific elements
	XMLPreserved // XML tags with ADF attributes preserved

	// New HTML details strategy for expand elements
	HTMLDetails // HTML <details> and <summary> tags for native markdown preview

	MarkdownPanel // Fenced-div syntax (:::type content :::)
)

// String returns a string representation of the ConversionStrategy
func (cs ConversionStrategy) String() string {
	switch cs {
	case StandardMarkdown:
		return "StandardMarkdown"
	case HTMLWrapped:
		return "HTMLWrapped"
	case Placeholder:
		return "Placeholder"
	case MarkdownTable:
		return "MarkdownTable"
	case MarkdownTaskList:
		return "MarkdownTaskList"
	case MarkdownBlockquote:
		return "MarkdownBlockquote"
	case MarkdownCodeBlock:
		return "MarkdownCodeBlock"
	case XMLPreserved:
		return "XMLPreserved"
	case HTMLDetails:
		return "HTMLDetails"
	case MarkdownPanel:
		return "MarkdownPanel"
	default:
		return "Unknown"
	}
}

// IsMarkdownReadable returns true if the strategy produces readable markdown
func (cs ConversionStrategy) IsMarkdownReadable() bool {
	switch cs {
	case StandardMarkdown:
		return true
	case HTMLWrapped:
		return true // HTML with markdown inside is still readable
	case Placeholder:
		return false // Placeholders are not readable markdown
	default:
		return false
	}
}

// RequiresMetadataPreservation returns true if the strategy preserves metadata
func (cs ConversionStrategy) RequiresMetadataPreservation() bool {
	switch cs {
	case StandardMarkdown:
		return false // Only preserves href
	case HTMLWrapped:
		return true // Preserves all metadata in HTML attributes
	case Placeholder:
		return true // Preserves everything in placeholder
	default:
		return false
	}
}

// SupportsAttributes returns true if the strategy can handle additional attributes
func (cs ConversionStrategy) SupportsAttributes() bool {
	switch cs {
	case StandardMarkdown:
		return false // Only href attribute supported
	case HTMLWrapped:
		return true // All attributes as HTML attributes
	case Placeholder:
		return true // All attributes preserved
	default:
		return false
	}
}

