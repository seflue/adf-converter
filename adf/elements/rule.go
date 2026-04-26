package elements

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// ruleRenderer handles bidirectional conversion of horizontal rule nodes
//
// In ADF: { type: "rule" } — no attrs, no children
// In Markdown: "---" (thematic break)
type ruleRenderer struct{}

func NewRuleRenderer() adf.Renderer {
	return &ruleRenderer{}
}

func (rc *ruleRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	if err := rc.ValidateInput(node); err != nil {
		return adf.RenderResult{}, err
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent("---\n\n")
	builder.IncrementConverted()

	return builder.Build(), nil
}

func (rc *ruleRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	if startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("startIndex %d out of range", startIndex)
	}

	// Parse only the single line so Goldmark cannot reinterpret context
	// (e.g. "text\n---" would be a Setext heading, not a thematic break).
	source := []byte(lines[startIndex])
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() == ast.KindThematicBreak {
			return adf.Node{Type: adf.NodeTypeRule}, 1, nil
		}
	}

	return adf.Node{}, 0, fmt.Errorf("not a thematic break: %q", lines[startIndex])
}

// CanParseLine returns true if the line is a CommonMark thematic break.
// Accepts sequences of -, *, or _ (all the same char) with optional spaces
// between them, as long as at least 3 of the char appear.
func (rc *ruleRenderer) CanParseLine(line string) bool {
	if len(line) < 3 {
		return false
	}

	ch := line[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}

	count := 0
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == ch {
			count++
		} else if c != ' ' && c != '\t' {
			return false
		}
	}

	return count >= 3
}

func (rc *ruleRenderer) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeTypeRule
}

func (rc *ruleRenderer) GetStrategy() adf.ConversionStrategy {
	return adf.StandardMarkdown
}

func (rc *ruleRenderer) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("invalid input type: expected Node, got %T", input)
	}

	if node.Type != adf.NodeTypeRule {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf.NodeTypeRule, node.Type)
	}

	if len(node.Content) > 0 {
		return fmt.Errorf("rule node should not have content")
	}

	return nil
}
