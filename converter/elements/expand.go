package elements

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
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
			htmlBuilder.WriteString(fmt.Sprintf(` id="%s"`, localIDStr))
		}
	}

	switch node.Type {
	case adf_types.NodeTypeNestedExpand:
		htmlBuilder.WriteString(` data-adf-type="nestedExpand"`)
	case adf_types.NodeTypeExpand:
		htmlBuilder.WriteString(` data-adf-type="expand"`)
	}

	htmlBuilder.WriteString(">\n")

	htmlBuilder.WriteString(fmt.Sprintf("  <summary>%s</summary>\n", title))

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

func (ec *ExpandConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if len(lines) == 0 || startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, nil
	}

	firstLine := strings.TrimSpace(lines[startIndex])

	detailsRegex := regexp.MustCompile(`^<details(\s+[^>]*)?>`)
	if !detailsRegex.MatchString(firstLine) {
		return adf_types.ADFNode{}, 0, nil
	}

	attributes := make(map[string]interface{})

	if strings.Contains(firstLine, " open") || strings.Contains(firstLine, " open>") {
		attributes["expanded"] = true
	}

	idRegex := regexp.MustCompile(`\sid\s*=\s*["|']([^"']*)["|']`)
	if idMatch := idRegex.FindStringSubmatch(firstLine); len(idMatch) > 1 {
		attributes["localId"] = idMatch[1]
	}

	nodeType := adf_types.NodeTypeExpand
	adfTypeRegex := regexp.MustCompile(`data-adf-type="([^"]+)"`)
	if matches := adfTypeRegex.FindStringSubmatch(firstLine); len(matches) > 1 {
		if matches[1] == "nestedExpand" {
			nodeType = adf_types.NodeTypeNestedExpand
		}
	}

	summaryIdx := -1
	summaryEndIdx := -1
	detailsEndIdx := -1

	for i := startIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if summaryIdx == -1 && strings.Contains(line, "<summary>") {
			summaryIdx = i
		}

		if summaryIdx != -1 && summaryEndIdx == -1 && strings.Contains(line, "</summary>") {
			summaryEndIdx = i
		}

		if strings.Contains(line, "</details>") {
			detailsEndIdx = i
			break
		}
	}

	if summaryIdx == -1 || summaryEndIdx == -1 || detailsEndIdx == -1 {
		return adf_types.ADFNode{}, 0, fmt.Errorf("malformed details element: missing required tags")
	}

	title := ""
	summaryLine := lines[summaryIdx]
	summaryStart := strings.Index(summaryLine, "<summary>")
	summaryEnd := strings.Index(summaryLine, "</summary>")
	if summaryStart != -1 && summaryEnd != -1 {
		title = strings.TrimSpace(summaryLine[summaryStart+9 : summaryEnd])
	}

	if title == "" {
		return adf_types.ADFNode{}, 0, fmt.Errorf("expand element missing required title")
	}

	attributes["title"] = title

	var contentNodes []adf_types.ADFNode
	if detailsEndIdx > summaryEndIdx+1 {
		contentLines := lines[summaryEndIdx+1 : detailsEndIdx]

		// Strip minimum common indentation to preserve relative indentation (for nested lists, etc.)
		cleanedLines := DedentLines(contentLines)

		if len(cleanedLines) > 0 {
			contentText := strings.TrimSpace(strings.Join(cleanedLines, "\n"))
			if contentText != "" {
				textNode := adf_types.ADFNode{
					Type: adf_types.NodeTypeText,
					Text: contentText,
				}
				paragraphNode := adf_types.ADFNode{
					Type:    adf_types.NodeTypeParagraph,
					Content: []adf_types.ADFNode{textNode},
				}
				contentNodes = append(contentNodes, paragraphNode)
			}
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
