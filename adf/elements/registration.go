package elements

import (
	"github.com/seflue/adf-converter/adf"
)

// Registration pairs an ADF node type with its adf.Renderer. It is the
// metadata shape used by StandardNodes so that callers (converter/defaults and
// the elements-package test helpers) can build a adf.ConverterRegistry from a
// single source of truth.
type Registration struct {
	NodeType  adf.NodeType
	Converter adf.Renderer
}

// StandardNodes returns the canonical list of 22 element-converter
// registrations that ship with the library. Callers get fresh adf.Renderer
// instances on every call so independent registries never share state.
func StandardNodes() []Registration {
	return []Registration{
		{"text", NewTextConverter()},
		{"hardBreak", NewHardBreakConverter()},
		{"paragraph", NewParagraphConverter()},
		{"heading", NewHeadingConverter()},
		{"listItem", NewListItemConverter()},
		{"bulletList", NewBulletListConverter()},
		{"orderedList", NewOrderedListConverter()},
		{"expand", NewExpandConverter()},
		{"nestedExpand", NewExpandConverter()},
		{"inlineCard", NewInlineCardConverter()},
		{"blockCard", NewBlockCardConverter()},
		{"emoji", NewEmojiConverter()},
		{"codeBlock", NewCodeBlockConverter()},
		{"rule", NewRuleConverter()},
		{"mention", NewMentionConverter()},
		{"table", NewTableConverter()},
		{"panel", NewPanelConverter()},
		{"date", NewDateConverter()},
		{"status", NewStatusConverter()},
		{"blockquote", NewBlockquoteConverter()},
		{"taskList", NewTaskListConverter()},
		{"mediaSingle", NewMediaSingleConverter()},
	}
}

// StandardBlockParserOrder is the canonical MD→ADF block-parser dispatch
// order (first match wins). Specific before general:
//   - panel before blockquote (> [!TYPE] vs >)
//   - taskList before bulletList (- [ ] vs -)
//   - rule before bulletList (--- vs -)
var StandardBlockParserOrder = []adf.NodeType{
	"expand", "blockCard", "panel", "table", "taskList",
	"blockquote", "codeBlock", "heading", "mediaSingle", "rule",
	"bulletList", "orderedList",
}
