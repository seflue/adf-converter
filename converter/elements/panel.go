package elements

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// admonitionRegex matches GitHub-style admonition headers: > [!TYPE]
var admonitionRegex = regexp.MustCompile(`(?i)^>\s*\[!(INFO|WARNING|ERROR|SUCCESS|NOTE|TIP)\]\s*$`)

// PanelConverter handles conversion of ADF panel nodes to/from markdown
//
// Output (ADF -> MD): Fenced-div syntax (:::type ... :::)
// Input (MD -> ADF): Fenced-div AND GitHub Admonition (> [!TYPE])
type PanelConverter struct{}

func NewPanelConverter() *PanelConverter {
	return &PanelConverter{}
}

func (pc *PanelConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	builder := converter.NewEnhancedConversionResultBuilder(converter.MarkdownPanel)

	panelType := pc.extractPanelType(node)

	if !isKnownPanelType(panelType) {
		builder.AddWarningf("unknown panel type: %s", panelType)
	}

	var contentBuilder strings.Builder
	for _, child := range node.Content {
		childConverter := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(child.Type))
		if childConverter == nil {
			builder.AddWarningf("no converter for child type: %s", child.Type)
			continue
		}

		childResult, err := childConverter.ToMarkdown(child, context)
		if err != nil {
			return converter.EnhancedConversionResult{}, fmt.Errorf("converting panel child %s: %w", child.Type, err)
		}

		contentBuilder.WriteString(childResult.Content)
	}

	content := strings.TrimRight(contentBuilder.String(), "\n")

	var result strings.Builder
	result.WriteString(":::" + panelType + "\n")
	if content != "" {
		result.WriteString(content + "\n")
	}
	result.WriteString(":::\n\n")

	builder.AppendContent(result.String())
	return builder.Build(), nil
}

func (pc *PanelConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	if startIndex >= len(lines) {
		return adf_types.ADFNode{}, 0, fmt.Errorf("startIndex out of range")
	}

	firstLine := strings.TrimSpace(lines[startIndex])

	if strings.HasPrefix(firstLine, ":::") {
		return pc.parseFencedDiv(lines, startIndex, context)
	}

	if isGitHubAdmonition(firstLine) {
		return pc.parseGitHubAdmonition(lines, startIndex, context)
	}

	return adf_types.ADFNode{}, 0, fmt.Errorf("unrecognized panel syntax: %s", firstLine)
}

// parseFencedDiv parses :::type ... ::: syntax
func (pc *PanelConverter) parseFencedDiv(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	firstLine := strings.TrimSpace(lines[startIndex])
	panelType := strings.ToLower(strings.TrimSpace(firstLine[3:]))
	if panelType == "" {
		panelType = "info"
	}

	// Find closing fence
	closingIdx := -1
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == ":::" {
			closingIdx = i
			break
		}
	}

	if closingIdx == -1 {
		return adf_types.ADFNode{}, 0, fmt.Errorf("unclosed panel fence starting at line %d", startIndex)
	}

	// Parse inner content
	contentLines := lines[startIndex+1 : closingIdx]
	contentNodes, err := parseInnerContentWithContext(contentLines, context)
	if err != nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("parsing panel content: %w", err)
	}

	node := adf_types.ADFNode{
		Type:    adf_types.NodeTypePanel,
		Attrs:   map[string]interface{}{"panelType": panelType},
		Content: contentNodes,
	}

	consumed := closingIdx - startIndex + 1
	return node, consumed, nil
}

func (pc *PanelConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, ":::") || admonitionRegex.MatchString(line)
}

func (pc *PanelConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.ADFNodeType(adf_types.NodeTypePanel)
}

func (pc *PanelConverter) GetStrategy() converter.ConversionStrategy {
	return converter.MarkdownPanel
}

func (pc *PanelConverter) ValidateInput(input interface{}) error {
	node, ok := input.(adf_types.ADFNode)
	if !ok {
		return fmt.Errorf("input must be an ADFNode")
	}
	if node.Type != adf_types.NodeTypePanel {
		return fmt.Errorf("node type must be panel, got: %s", node.Type)
	}
	return nil
}

// admonitionTypeMapping maps GitHub admonition types to ADF panel types
var admonitionTypeMapping = map[string]string{
	"info":    "info",
	"warning": "warning",
	"error":   "error",
	"success": "success",
	"note":    "note",
	"tip":     "note",
}

// isGitHubAdmonition checks if a line matches > [!TYPE] pattern
func isGitHubAdmonition(line string) bool {
	return admonitionRegex.MatchString(line)
}

// parseGitHubAdmonition parses > [!TYPE] ... syntax
func (pc *PanelConverter) parseGitHubAdmonition(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	firstLine := strings.TrimSpace(lines[startIndex])

	// Extract type from [!TYPE]
	matches := admonitionRegex.FindStringSubmatch(firstLine)
	if len(matches) < 2 {
		return adf_types.ADFNode{}, 0, fmt.Errorf("invalid admonition syntax: %s", firstLine)
	}

	rawType := strings.ToLower(matches[1])
	panelType := admonitionTypeMapping[rawType]
	if panelType == "" {
		panelType = "info"
	}

	// Collect content lines: all subsequent lines starting with >
	var contentLines []string
	i := startIndex + 1
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if !strings.HasPrefix(trimmed, ">") {
			break
		}

		// Strip > prefix
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "> ") {
			stripped = stripped[2:]
		} else if strings.HasPrefix(stripped, ">") {
			stripped = stripped[1:]
		}

		contentLines = append(contentLines, stripped)
		i++
	}

	consumed := i - startIndex

	// Parse inner content
	contentNodes, err := parseInnerContentWithContext(contentLines, context)
	if err != nil {
		return adf_types.ADFNode{}, 0, fmt.Errorf("parsing admonition content: %w", err)
	}

	node := adf_types.ADFNode{
		Type:    adf_types.NodeTypePanel,
		Attrs:   map[string]interface{}{"panelType": panelType},
		Content: contentNodes,
	}

	return node, consumed, nil
}

// knownPanelTypes contains the standard ADF panel types
var knownPanelTypes = map[string]bool{
	"info": true, "warning": true, "error": true, "success": true, "note": true,
}

func isKnownPanelType(panelType string) bool {
	return knownPanelTypes[panelType]
}

// extractPanelType returns the panel type from node attrs, defaulting to "info"
func (pc *PanelConverter) extractPanelType(node adf_types.ADFNode) string {
	if node.Attrs != nil {
		if pt, ok := node.Attrs["panelType"].(string); ok && pt != "" {
			return pt
		}
	}
	return "info"
}
