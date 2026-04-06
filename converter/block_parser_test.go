package converter_test

import (
	"testing"

	"adf-converter/converter"
	"adf-converter/converter/elements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// asBlockParser asserts that a converter implements BlockParser and returns it.
func asBlockParser(t *testing.T, name string, conv converter.ElementConverter) converter.BlockParser {
	t.Helper()
	bp, ok := conv.(converter.BlockParser)
	require.True(t, ok, "%s must implement BlockParser", name)
	return bp
}

func TestCanParseLine_Expand(t *testing.T) {
	bp := asBlockParser(t, "ExpandConverter", elements.NewExpandConverter())
	tests := []struct {
		line string
		want bool
	}{
		{"<details>", true},
		{"<details open>", true},
		{"<details title=\"foo\">", true},
		{"some text", false},
		{"# heading", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_BlockCard(t *testing.T) {
	bp := asBlockParser(t, "BlockCardConverter", elements.NewBlockCardConverter())
	tests := []struct {
		line string
		want bool
	}{
		{`<div data-adf-type="blockCard" data-url="http://example.com">`, true},
		{"<div>", false},
		{"<div class=\"other\">", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_Panel(t *testing.T) {
	bp := asBlockParser(t, "PanelConverter", elements.NewPanelConverter())
	tests := []struct {
		line string
		want bool
	}{
		// Fenced div syntax
		{":::info", true},
		{":::warning", true},
		{":::", true},
		// GitHub admonition syntax
		{"> [!INFO]", true},
		{"> [!WARNING]", true},
		{"> [!NOTE]", true},
		{"> [!TIP]", true},
		{"> [!ERROR]", true},
		{"> [!SUCCESS]", true},
		// Negatives
		{"> normal blockquote", false},
		{">", false},
		{"some text", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_Table(t *testing.T) {
	bp := asBlockParser(t, "TableConverter", elements.NewTableConverter())
	tests := []struct {
		line string
		want bool
	}{
		// HTML table
		{"<table>", true},
		{"<table class=\"foo\">", true},
		// Pipe table
		{"| col1 | col2 |", true},
		{"|---|---|", true},
		// Negatives
		{"some text", false},
		{"# heading", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_TaskList(t *testing.T) {
	bp := asBlockParser(t, "TaskListConverter", elements.NewTaskListConverter())
	tests := []struct {
		line string
		want bool
	}{
		// XML syntax
		{"<taskList localId=\"abc\">", true},
		{"<taskList>", true},
		// Plain checkbox syntax
		{"- [ ] unchecked", true},
		{"- [x] checked", true},
		{"- [X] checked", true},
		// Negatives
		{"- normal item", false},
		{"some text", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_Blockquote(t *testing.T) {
	bp := asBlockParser(t, "BlockquoteConverter", elements.NewBlockquoteConverter())
	tests := []struct {
		line string
		want bool
	}{
		// XML syntax
		{"<blockquote>", true},
		// Plain blockquote
		{"> text", true},
		{">", true},
		// Also matches admonitions (ordering resolves conflict, not negative checks)
		{"> [!INFO]", true},
		// Negatives
		{"some text", false},
		{"# heading", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_CodeBlock(t *testing.T) {
	bp := asBlockParser(t, "CodeBlockConverter", elements.NewCodeBlockConverter())
	tests := []struct {
		line string
		want bool
	}{
		{"```", true},
		{"```go", true},
		{"```javascript", true},
		{"some text", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_Heading(t *testing.T) {
	bp := asBlockParser(t, "HeadingConverter", elements.NewHeadingConverter())
	tests := []struct {
		line string
		want bool
	}{
		{"# h1", true},
		{"## h2", true},
		{"### h3", true},
		{"some text", false},
		{"> blockquote", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_Rule(t *testing.T) {
	bp := asBlockParser(t, "RuleConverter", elements.NewRuleConverter())
	tests := []struct {
		line string
		want bool
	}{
		{"---", true},
		{"----", true},
		{"***", true},
		{"___", true},
		// Not a rule (too short)
		{"--", false},
		// Not a rule (mixed chars)
		{"-*-", false},
		// Negatives
		{"- list item", false},
		{"some text", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_BulletList(t *testing.T) {
	bp := asBlockParser(t, "BulletListConverter", elements.NewBulletListConverter())
	tests := []struct {
		line string
		want bool
	}{
		{"- item", true},
		{"-  item with extra space", true},
		// Also matches checkboxes (ordering resolves, not negative checks)
		{"- [ ] task", true},
		// Negatives
		{"some text", false},
		{"1. ordered", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_OrderedList(t *testing.T) {
	bp := asBlockParser(t, "OrderedListConverter", elements.NewOrderedListConverter())
	tests := []struct {
		line string
		want bool
	}{
		{"1. item", true},
		{"10. item", true},
		{"1) item", false},
		{"- bullet", false},
		{"some text", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

// TestBlockParserDispatchOrder verifies that ambiguous lines resolve to the correct
// converter based on registration order (first match wins).
// Requires Registry.BlockParsers() — added in step 3.
func TestBlockParserDispatchOrder(t *testing.T) {
	registry := converter.GetGlobalRegistry()

	tests := []struct {
		name     string
		line     string
		wantType string
	}{
		{"admonition goes to panel, not blockquote", "> [!INFO]", "panel"},
		{"plain blockquote goes to blockquote", "> some text", "blockquote"},
		{"checkbox goes to taskList, not bulletList", "- [ ] task", "taskList"},
		{"plain bullet goes to bulletList", "- item", "bulletList"},
		{"thematic break goes to rule, not bulletList", "---", "rule"},
		{"long thematic break goes to rule", "-----", "rule"},
		{"pipe table goes to table", "| col |", "table"},
		{"HTML table goes to table", "<table>", "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := false
			for _, entry := range registry.BlockParsers() {
				if entry.Parser.CanParseLine(tt.line) {
					assert.Equal(t, tt.wantType, string(entry.NodeType),
						"line %q should dispatch to %s", tt.line, tt.wantType)
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("line %q did not match any block parser", tt.line)
			}
		})
	}
}

// TestParagraphIsNotBlockParser verifies that ParagraphConverter does not implement
// BlockParser — it is always the fallback.
func TestParagraphIsNotBlockParser(t *testing.T) {
	var conv converter.ElementConverter = elements.NewParagraphConverter()
	_, ok := conv.(converter.BlockParser)
	assert.False(t, ok, "ParagraphConverter must NOT implement BlockParser")
}

// TestInlineConvertersAreNotBlockParsers verifies that inline converters do not
// implement BlockParser.
func TestInlineConvertersAreNotBlockParsers(t *testing.T) {
	inlineConverters := map[string]converter.ElementConverter{
		"text":       elements.NewTextConverter(),
		"hardBreak":  elements.NewHardBreakConverter(),
		"emoji":      elements.NewEmojiConverter(),
		"mention":    elements.NewMentionConverter(),
		"status":     elements.NewStatusConverter(),
		"date":       elements.NewDateConverter(),
		"inlineCard": elements.NewInlineCardConverter(),
	}
	for name, conv := range inlineConverters {
		_, ok := conv.(converter.BlockParser)
		assert.False(t, ok, "%s must NOT implement BlockParser", name)
	}
}
