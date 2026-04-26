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
	Renderer adf.Renderer
}

// StandardNodes returns the canonical list of 22 element-converter
// registrations that ship with the library. Callers get fresh adf.Renderer
// instances on every call so independent registries never share state.
func StandardNodes() []Registration {
	return []Registration{
		{"text", NewTextRenderer()},
		{"hardBreak", NewHardBreakRenderer()},
		{"paragraph", NewParagraphRenderer()},
		{"heading", NewHeadingRenderer()},
		{"listItem", NewListItemRenderer()},
		{"bulletList", NewBulletListRenderer()},
		{"orderedList", NewOrderedListRenderer()},
		{"expand", NewExpandRenderer()},
		{"nestedExpand", NewExpandRenderer()},
		{"inlineCard", NewInlineCardRenderer()},
		{"blockCard", NewBlockCardRenderer()},
		{"emoji", NewEmojiRenderer()},
		{"codeBlock", NewCodeBlockRenderer()},
		{"rule", NewRuleRenderer()},
		{"mention", NewMentionRenderer()},
		{"table", NewTableRenderer()},
		{"panel", NewPanelRenderer()},
		{"date", NewDateRenderer()},
		{"status", NewStatusRenderer()},
		{"blockquote", NewBlockquoteRenderer()},
		{"taskList", NewTaskListRenderer()},
		{"mediaSingle", NewMediaSingleRenderer()},
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
