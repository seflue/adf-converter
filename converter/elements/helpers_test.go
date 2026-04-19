package elements

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/placeholder"
)

// newTestRegistry builds a registry populated with all 21 standard element
// converters plus the canonical block-parser order, mirroring
// converter/defaults.NewRegistry. Tests in this package cannot import defaults
// (cycle), so the wiring is duplicated here.
func newTestRegistry() *element.ConverterRegistry {
	r := element.NewConverterRegistry()

	entries := []struct {
		nodeType  element.ADFNodeType
		converter element.Converter
	}{
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
	for _, e := range entries {
		r.MustRegister(e.nodeType, e.converter)
	}

	for _, nodeType := range []element.ADFNodeType{
		"expand", "blockCard", "panel", "table", "taskList",
		"blockquote", "codeBlock", "heading", "mediaSingle", "rule",
		"bulletList", "orderedList",
	} {
		r.MustRegisterBlockParser(nodeType)
	}

	return r
}

// testParseNested wires a MarkdownParser-backed ParseNested closure for tests
// that exercise element converters expecting nested markdown parsing.
func testParseNested() func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
	mgr := placeholder.NewManager()
	return testParseNestedWith(mgr)
}

// testParseNestedWith wires ParseNested using the given manager so tests that
// pre-stored placeholders can recover them during nested parsing.
func testParseNestedWith(mgr placeholder.Manager) func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
	return func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
		p := element.NewMarkdownParserWithNesting(mgr.GetSession(), mgr, newTestRegistry(), nestedLevel)
		return p.ParseMarkdownToADFNodes(lines)
	}
}
