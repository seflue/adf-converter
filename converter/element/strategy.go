package element

// ConversionStrategy defines how an ADF element should be rendered to markdown.
type ConversionStrategy int

const (
	StandardMarkdown ConversionStrategy = iota
	HTMLWrapped
	Placeholder

	MarkdownTable
	MarkdownTaskList
	MarkdownBlockquote
	MarkdownCodeBlock

	XMLPreserved

	HTMLDetails

	MarkdownPanel
)

// String returns a string representation of the ConversionStrategy.
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

// IsMarkdownReadable returns true if the strategy produces readable markdown.
func (cs ConversionStrategy) IsMarkdownReadable() bool {
	switch cs {
	case StandardMarkdown:
		return true
	case HTMLWrapped:
		return true
	case Placeholder:
		return false
	default:
		return false
	}
}

// RequiresMetadataPreservation returns true if the strategy preserves metadata.
func (cs ConversionStrategy) RequiresMetadataPreservation() bool {
	switch cs {
	case StandardMarkdown:
		return false
	case HTMLWrapped:
		return true
	case Placeholder:
		return true
	default:
		return false
	}
}

// SupportsAttributes returns true if the strategy can handle additional attributes.
func (cs ConversionStrategy) SupportsAttributes() bool {
	switch cs {
	case StandardMarkdown:
		return false
	case HTMLWrapped:
		return true
	case Placeholder:
		return true
	default:
		return false
	}
}
