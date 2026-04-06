package elements

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

var (
	detailsOpenRegex = regexp.MustCompile(`^<details(\s+[^>]*)?>`)
	idAttrRegex      = regexp.MustCompile(`\sid\s*=\s*["|']([^"']*)["|']`)
	adfTypeAttrRegex = regexp.MustCompile(`data-adf-type="([^"]+)"`)
)

// ExpandConverter handles conversion of ADF expand and nestedExpand nodes to/from markdown
//
// This converter handles BOTH "expand" AND "nestedExpand" node types with a single implementation.
// The two types are functionally identical in terms of conversion - they both use HTML <details>
// elements with data-adf-type attributes to preserve the original type for round-trip fidelity.
type ExpandConverter struct{}

func NewExpandConverter() *ExpandConverter {
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

	if title == "" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("expand node missing required title attribute")
	}

	isExpanded := false
	if expandedAttr, exists := node.Attrs["expanded"]; exists {
		if expandedBool, ok := expandedAttr.(bool); ok {
			isExpanded = expandedBool
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

	if isExpanded {
		htmlBuilder.WriteString("<details open")
	} else {
		htmlBuilder.WriteString("<details")
	}

	if localID, exists := node.Attrs["localId"]; exists {
		if localIDStr, ok := localID.(string); ok {
			fmt.Fprintf(&htmlBuilder, ` id="%s"`, localIDStr)
		}
	}

	switch node.Type {
	case adf_types.NodeTypeNestedExpand:
		htmlBuilder.WriteString(` data-adf-type="nestedExpand"`)
	case adf_types.NodeTypeExpand:
		htmlBuilder.WriteString(` data-adf-type="expand"`)
	}

	htmlBuilder.WriteString(">\n")

	fmt.Fprintf(&htmlBuilder, "  <summary>%s</summary>\n", title)

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

	if strings.Contains(firstLine, " open") || strings.Contains(firstLine, " open>") {
		attributes["expanded"] = true
	}

	if idMatch := idAttrRegex.FindStringSubmatch(firstLine); len(idMatch) > 1 {
		attributes["localId"] = idMatch[1]
	}

	nodeType := adf_types.NodeTypeExpand
	if matches := adfTypeAttrRegex.FindStringSubmatch(firstLine); len(matches) > 1 {
		if matches[1] == "nestedExpand" {
			nodeType = adf_types.NodeTypeNestedExpand
		}
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
			if title == "" {
				return 0, "", fmt.Errorf("expand element missing required title")
			}
			return i, title, nil
		}
	}
	return 0, "", fmt.Errorf("details element missing required summary tag")
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

	if _, exists := node.Attrs["title"]; !exists {
		return fmt.Errorf("expand node missing required title attribute")
	}

	return nil
}
