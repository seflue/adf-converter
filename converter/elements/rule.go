package elements

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

// RuleConverter handles bidirectional conversion of horizontal rule nodes
//
// In ADF: { type: "rule" } — no attrs, no children
// In Markdown: "---" (thematic break)
type RuleConverter struct{}

func NewRuleConverter() converter.ElementConverter {
	return &RuleConverter{}
}

func (rc *RuleConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if err := rc.ValidateInput(node); err != nil {
		return converter.EnhancedConversionResult{}, err
	}

	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)
	builder.AppendContent("---\n\n")
	builder.IncrementConverted()

	return builder.Build(), nil
}

func (rc *RuleConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("startIndex %d out of range", startIndex)
	}

	// Parse only the single line so Goldmark cannot reinterpret context
	// (e.g. "text\n---" would be a Setext heading, not a thematic break).
	source := []byte(lines[startIndex])
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() == ast.KindThematicBreak {
			return adf_types.ADFNode{Type: adf_types.NodeTypeRule}, 1, nil
		}
	}

	return adf_types.ADFNode{}, 0, fmt.Errorf("not a thematic break: %q", lines[startIndex])
}

// CanParseLine returns true if the line is a CommonMark thematic break.
// Accepts sequences of -, *, or _ (all the same char) with optional spaces
// between them, as long as at least 3 of the char appear.
func (rc *RuleConverter) CanParseLine(line string) bool {
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

func (rc *RuleConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == adf_types.NodeTypeRule
}

func (rc *RuleConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (rc *RuleConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("invalid input type: expected ADFNode, got %T", input)
	}

	if node.Type != adf_types.NodeTypeRule {
		return fmt.Errorf("invalid node type: expected %s, got %s", adf_types.NodeTypeRule, node.Type)
	}

	if len(node.Content) > 0 {
		return fmt.Errorf("rule node should not have content")
	}

	return nil
}
