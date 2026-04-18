package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
)

var (
	detailsOpenRegex = regexp.MustCompile(`^<details(\s+[^>]*)?>`)
	idAttrRegex      = regexp.MustCompile(`\sid\s*=\s*["|']([^"']*)["|']`)
)

// ExpandConverter handles conversion of ADF expand and nestedExpand nodes to/from markdown
//
// This converter handles BOTH "expand" AND "nestedExpand" node types with a single implementation.
// Both use plain HTML <details> elements. The node type is derived from structural context:
// top-level <details> → expand, nested <details> (NestedLevel > 0) → nestedExpand.
type ExpandConverter struct{}

func NewExpandConverter() converter.ElementConverter {
	return &ExpandConverter{}
}

func (ec *ExpandConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := converter.NewEnhancedConversionResultBuilder(converter.StandardMarkdown)

	title := ""
	if titleAttr, exists := node.Attrs["title"]; exists {
		if titleStr, ok := titleAttr.(string); ok {
			title = titleStr
		}
	}

	var contentBuilder strings.Builder
	for i, child := range node.Content {
		childConverter := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(child.Type))
		if childConverter == nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("no converter found for child type: %s", child.Type)
		}

		childResult, err := childConverter.ToMarkdown(child, context)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("failed to convert expand content: %w", err)
		}

		contentBuilder.WriteString(strings.TrimSpace(childResult.Content))

		if i < len(node.Content)-1 {
			contentBuilder.WriteString("\n\n")
		}
	}

	var htmlBuilder strings.Builder

	htmlBuilder.WriteString("<details")

	if expanded, ok := node.Attrs["expanded"].(bool); ok && expanded {
		htmlBuilder.WriteString(" open")
	}

	if localID, exists := node.Attrs["localId"]; exists {
		if localIDStr, ok := localID.(string); ok {
			fmt.Fprintf(&htmlBuilder, ` id="%s"`, localIDStr)
		}
	}

	htmlBuilder.WriteString(">\n")

	if title != "" {
		fmt.Fprintf(&htmlBuilder, "  <summary>%s</summary>\n", title)
	}

	content := strings.TrimSpace(contentBuilder.String())
	if content != "" {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				htmlBuilder.WriteString("  " + line + "\n")
			} else {
				htmlBuilder.WriteString("\n")
			}
		}
	}

	htmlBuilder.WriteString("</details>")

	builder.AppendContent(htmlBuilder.String() + "\n\n")
	return builder.Build(), nil
}

const maxExpandNestingDepth = 100

func (ec *ExpandConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, nil
	}

	if context.NestedLevel > maxExpandNestingDepth {
		return adf_types.ADFNode{}, 0, fmt.Errorf("maximum nesting depth exceeded (%d levels)", maxExpandNestingDepth)
	}

	firstLine := strings.TrimSpace(lines[startIndex])

	if !detailsOpenRegex.MatchString(firstLine) {
		return adf_types.ADFNode{}, 0, nil
	}

	// Parse attributes from opening tag
	attributes := make(map[string]interface{})

	if idMatch := idAttrRegex.FindStringSubmatch(firstLine); len(idMatch) > 1 {
		attributes["localId"] = idMatch[1]
	}

	// Derive node type from structural context: nested <details> = nestedExpand
	nodeType := adf_types.NodeTypeExpand
	if context.NestedLevel > 0 {
		nodeType = adf_types.NodeTypeNestedExpand
	}

	// Find summary and closing tag with nesting-aware scan
	summaryEndIdx, title, err := ec.findSummary(lines, startIndex)
	if err != nil {
		return adf_types.ADFNode{}, 0, err
	}

	detailsEndIdx, err := ec.findClosingTag(lines, summaryEndIdx+1)
	if err != nil {
		return adf_types.ADFNode{}, 0, err
	}

	attributes["title"] = title

	// Parse inner content recursively with incremented nesting depth
	var contentNodes []adf_types.ADFNode
	if detailsEndIdx > summaryEndIdx+1 {
		contentLines := lines[summaryEndIdx+1 : detailsEndIdx]
		cleanedLines := DedentLines(contentLines)

		innerContext := context
		innerContext.NestedLevel++
		contentNodes, err = parseInnerContentWithContext(cleanedLines, innerContext)
		if err != nil {
			return adf_types.ADFNode{}, 0, fmt.Errorf("parsing expand content: %w", err)
		}
	}

	node := adf_types.ADFNode{
		Type:    nodeType,
		Attrs:   attributes,
		Content: contentNodes,
	}

	linesConsumed := detailsEndIdx - startIndex + 1
	return node, linesConsumed, nil
}

// findSummary scans for <summary>...</summary> near the opening tag.
// Returns the line index of the summary end and the extracted title.
func (ec *ExpandConverter) findSummary(lines []string, startIndex int) (summaryEndIdx int, title string, err error) {
	for i := startIndex; i < len(lines) && i <= startIndex+5; i++ {
		line := lines[i]
		summaryStart := strings.Index(line, "<summary>")
		summaryEnd := strings.Index(line, "</summary>")

		if summaryStart != -1 && summaryEnd != -1 {
			title = strings.TrimSpace(line[summaryStart+9 : summaryEnd])
			return i, title, nil
		}
	}
	// No <summary> found — empty title, content starts after <details> line
	return startIndex, "", nil
}

// findClosingTag finds the matching </details> considering nested <details> elements.
func (ec *ExpandConverter) findClosingTag(lines []string, searchStart int) (int, error) {
	nestingLevel := 0
	for i := searchStart; i < len(lines); i++ {
		line := lines[i]

		if strings.Contains(line, "<details") {
			nestingLevel++
		}

		if strings.Contains(line, "</details>") {
			if nestingLevel > 0 {
				nestingLevel--
			} else {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("unclosed details element")
}

// parseInnerContentWithContext parses markdown content using a MarkdownParser
// that inherits placeholder support and nesting level from the parent context.
func parseInnerContentWithContext(lines []string, context converter.ConversionContext) ([]adf_types.ADFNode, error) {
	hasContent := false
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return nil, nil
	}

	parser := converter.NewMarkdownParserWithNesting(context.PlaceholderSession, context.PlaceholderManager, context.NestedLevel)
	return parser.ParseMarkdownToADFNodes(lines)
}

func (ec *ExpandConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "<details")
}

func (ec *ExpandConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypeExpand) ||
		nodeType == converter.ADFNodeType(adf_types.NodeTypeNestedExpand)
}

func (ec *ExpandConverter) GetStrategy() converter.ConversionStrategy {
	return converter.StandardMarkdown
}

func (ec *ExpandConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}

	if node.Type != adf_types.NodeTypeExpand && node.Type != adf_types.NodeTypeNestedExpand {
		return fmt.Errorf("node type must be expand or nestedExpand, got: %s", node.Type)
	}

	if node.Attrs == nil {
		return fmt.Errorf("expand node missing attributes")
	}

	return nil
}
