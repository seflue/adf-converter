// Package defaults wires the eager registration of all standard ADF element
// converters. Callers that want a ready-to-use converter without assembling
// the registry themselves should use NewDefaultConverter or NewDisplayConverter.
//
// This package is intentionally separate from converter/ to avoid an import
// cycle: the elements/ package depends on converter/, so converter/ cannot
// import elements/ to eager-register. defaults/ closes that gap by being the
// single place that knows about both.
package defaults

import (
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/elements"
	"github.com/seflue/adf-converter/placeholder"
)

// NewRegistry builds a fresh registry with all 21 standard element converters
// registered and the block-parser dispatch order wired.
func NewRegistry() *converter.ConverterRegistry {
	r := converter.NewConverterRegistry()

	converters := []struct {
		nodeType  converter.ADFNodeType
		converter converter.ElementConverter
	}{
		{"text", elements.NewTextConverter()},
		{"hardBreak", elements.NewHardBreakConverter()},
		{"paragraph", elements.NewParagraphConverter()},
		{"heading", elements.NewHeadingConverter()},
		{"listItem", elements.NewListItemConverter()},
		{"bulletList", elements.NewBulletListConverter()},
		{"orderedList", elements.NewOrderedListConverter()},
		{"expand", elements.NewExpandConverter()},
		{"nestedExpand", elements.NewExpandConverter()},
		{"inlineCard", elements.NewInlineCardConverter()},
		{"blockCard", elements.NewBlockCardConverter()},
		{"emoji", elements.NewEmojiConverter()},
		{"codeBlock", elements.NewCodeBlockConverter()},
		{"rule", elements.NewRuleConverter()},
		{"mention", elements.NewMentionConverter()},
		{"table", elements.NewTableConverter()},
		{"panel", elements.NewPanelConverter()},
		{"date", elements.NewDateConverter()},
		{"status", elements.NewStatusConverter()},
		{"blockquote", elements.NewBlockquoteConverter()},
		{"taskList", elements.NewTaskListConverter()},
		{"mediaSingle", elements.NewMediaSingleConverter()},
	}
	for _, c := range converters {
		r.MustRegister(c.nodeType, c.converter)
	}

	// Block parser dispatch order (MD→ADF, first match wins). Specific before general:
	// - panel before blockquote (> [!TYPE] vs >)
	// - taskList before bulletList (- [ ] vs -)
	// - rule before bulletList (--- vs -)
	blockParserOrder := []converter.ADFNodeType{
		"expand", "blockCard", "panel", "table", "taskList",
		"blockquote", "codeBlock", "heading", "mediaSingle", "rule",
		"bulletList", "orderedList",
	}
	for _, nodeType := range blockParserOrder {
		r.MustRegisterBlockParser(nodeType)
	}

	return r
}

// NewDefaultConverter returns a converter wired with all standard element
// converters and the default classifier and placeholder manager.
func NewDefaultConverter() *converter.DefaultConverter {
	return converter.NewConverter(
		converter.WithRegistry(NewRegistry()),
	)
}

// NewDisplayConverter returns a converter for read-only display mode.
// It uses a NullManager that produces preview text instead of placeholder comments.
func NewDisplayConverter() *converter.DefaultConverter {
	return converter.NewConverter(
		converter.WithRegistry(NewRegistry()),
		converter.WithPlaceholderManager(placeholder.NewNullManager()),
	)
}
