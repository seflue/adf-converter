package adf_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/adf/elements"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// asBlockParser asserts that a converter implements BlockParser and returns it.
func asBlockParser(t *testing.T, name string, conv adf.Renderer) adf.BlockParser {
	t.Helper()
	bp, ok := conv.(adf.BlockParser)
	require.True(t, ok, "%s must implement BlockParser", name)
	return bp
}

func TestCanParseLine_Expand(t *testing.T) {
	bp := asBlockParser(t, "ExpandConverter", elements.NewExpandRenderer())
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
	bp := asBlockParser(t, "BlockCardConverter", elements.NewBlockCardRenderer())
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
	bp := asBlockParser(t, "PanelConverter", elements.NewPanelRenderer())
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
	bp := asBlockParser(t, "TableConverter", elements.NewTableRenderer())
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
	bp := asBlockParser(t, "TaskListConverter", elements.NewTaskListRenderer())
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
	bp := asBlockParser(t, "BlockquoteConverter", elements.NewBlockquoteRenderer())
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
	bp := asBlockParser(t, "CodeBlockConverter", elements.NewCodeBlockRenderer())
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
	bp := asBlockParser(t, "HeadingConverter", elements.NewHeadingRenderer())
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
	bp := asBlockParser(t, "RuleConverter", elements.NewRuleRenderer())
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
		// CommonMark: spaces between chars are valid thematic breaks
		{"- - -", true},
		{"* * *", true},
		{"_ _ _", true},
		// Negatives
		{"- list item", false},
		{"some text", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, bp.CanParseLine(tt.line), "line: %q", tt.line)
	}
}

func TestCanParseLine_BulletList(t *testing.T) {
	bp := asBlockParser(t, "BulletListConverter", elements.NewBulletListRenderer())
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
	bp := asBlockParser(t, "OrderedListConverter", elements.NewOrderedListRenderer())
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
	registry := defaults.NewRegistry()

	tests := []struct {
		name     string
		line     string
		wantType adf.NodeType
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
					assert.Equal(t, tt.wantType, entry.NodeType,
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
	conv := elements.NewParagraphRenderer()
	_, ok := any(conv).(adf.BlockParser)
	assert.False(t, ok, "ParagraphConverter must NOT implement BlockParser")
}

// TestInlineConvertersAreNotBlockParsers verifies that inline converters do not
// implement BlockParser.
func TestInlineConvertersAreNotBlockParsers(t *testing.T) {
	inlineConverters := map[string]adf.Renderer{
		"text":       elements.NewTextRenderer(),
		"hardBreak":  elements.NewHardBreakRenderer(),
		"emoji":      elements.NewEmojiRenderer(),
		"mention":    elements.NewMentionRenderer(),
		"status":     elements.NewStatusRenderer(),
		"date":       elements.NewDateRenderer(),
		"inlineCard": elements.NewInlineCardRenderer(),
	}
	for name, conv := range inlineConverters {
		_, ok := conv.(adf.BlockParser)
		assert.False(t, ok, "%s must NOT implement BlockParser", name)
	}
}
