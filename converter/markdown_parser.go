package converter

import (
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/placeholder"
)

// MarkdownParser is an alias for element.MarkdownParser.
type MarkdownParser = element.MarkdownParser

// NewMarkdownParser creates a new parser bound to the supplied registry.
func NewMarkdownParser(session *placeholder.EditSession, manager placeholder.Manager, registry *ConverterRegistry) *MarkdownParser {
	return element.NewMarkdownParser(session, manager, registry)
}

// NewMarkdownParserWithNesting creates a parser that inherits a nesting level from a parent context.
func NewMarkdownParserWithNesting(session *placeholder.EditSession, manager placeholder.Manager, registry *ConverterRegistry, nestedLevel int) *MarkdownParser {
	return element.NewMarkdownParserWithNesting(session, manager, registry, nestedLevel)
}
