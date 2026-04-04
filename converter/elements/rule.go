package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// RuleConverter handles bidirectional conversion of horizontal rule nodes
//
// In ADF: { type: "rule" } — no attrs, no children
// In Markdown: "---" (thematic break)
type RuleConverter struct{}

func NewRuleConverter() *RuleConverter {
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

	line := strings.TrimSpace(lines[startIndex])
	if !IsThematicBreak(line) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("not a thematic break: %q", line)
	}

	return adf_types.ADFNode{Type: adf_types.NodeTypeRule}, 1, nil
}

// IsThematicBreak returns true if the line is a Markdown thematic break (---, ***, ___).
// Requires at least 3 identical characters from the set {-, *, _}.
func IsThematicBreak(line string) bool {
	if len(line) < 3 {
		return false
	}

	ch := line[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}

	for i := 1; i < len(line); i++ {
		if line[i] != ch {
			return false
		}
	}

	return true
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
