package elements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// admonitionRegex matches GitHub-style admonition headers: > [!TYPE]
var admonitionRegex = regexp.MustCompile(`(?i)^>\s*\[!(INFO|WARNING|ERROR|SUCCESS|NOTE|TIP)\]\s*$`)

// panelRenderer handles conversion of ADF panel nodes to/from markdown
//
// Output (ADF -> MD): Fenced-div syntax (:::type ... :::)
// Input (MD -> ADF): Fenced-div AND GitHub Admonition (> [!TYPE])
type panelRenderer struct{}

func NewPanelRenderer() adf.Renderer {
	return &panelRenderer{}
}

func (pc *panelRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	builder := convresult.NewRenderResultBuilder(adf.MarkdownPanel)

	panelType := pc.extractPanelType(node)

	if !isKnownPanelType(panelType) {
		builder.AddWarningf("unknown panel type: %s", panelType)
	}

	var contentBuilder strings.Builder
	for _, child := range node.Content {
		childRenderer , _ := context.Registry.Lookup(adf.NodeType(child.Type))
		if childRenderer == nil {
			builder.AddWarningf("no converter for child type: %s", child.Type)
			continue
		}

		childResult, err := childRenderer.ToMarkdown(child, context)
		if err != nil {
			return adf.RenderResult{}, fmt.Errorf("converting panel child %s: %w", child.Type, err)
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

func (pc *panelRenderer) FromMarkdown(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	if startIndex >= len(lines) {
		return adf.Node{}, 0, fmt.Errorf("startIndex out of range")
	}

	firstLine := strings.TrimSpace(lines[startIndex])

	if strings.HasPrefix(firstLine, ":::") {
		return pc.parseFencedDiv(lines, startIndex, context)
	}

	if isGitHubAdmonition(firstLine) {
		return pc.parseGitHubAdmonition(lines, startIndex, context)
	}

	return adf.Node{}, 0, fmt.Errorf("unrecognized panel syntax: %s", firstLine)
}

// parseFencedDiv parses :::type ... ::: syntax
func (pc *panelRenderer) parseFencedDiv(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
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
		return adf.Node{}, 0, fmt.Errorf("unclosed panel fence starting at line %d", startIndex)
	}

	// Parse inner content
	contentLines := lines[startIndex+1 : closingIdx]
	contentNodes, err := parseInnerContentWithContext(contentLines, context)
	if err != nil {
		return adf.Node{}, 0, fmt.Errorf("parsing panel content: %w", err)
	}

	node := adf.Node{
		Type:    adf.NodeTypePanel,
		Attrs:   map[string]any{"panelType": panelType},
		Content: contentNodes,
	}

	consumed := closingIdx - startIndex + 1
	return node, consumed, nil
}

func (pc *panelRenderer) CanParseLine(line string) bool {
	return strings.HasPrefix(line, ":::") || admonitionRegex.MatchString(line)
}

func (pc *panelRenderer) CanHandle(nodeType adf.NodeType) bool {
	return nodeType == adf.NodeType(adf.NodeTypePanel)
}

func (pc *panelRenderer) GetStrategy() adf.ConversionStrategy {
	return adf.MarkdownPanel
}

func (pc *panelRenderer) ValidateInput(input any) error {
	node, ok := input.(adf.Node)
	if !ok {
		return fmt.Errorf("input must be a Node")
	}
	if node.Type != adf.NodeTypePanel {
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
func (pc *panelRenderer) parseGitHubAdmonition(lines []string, startIndex int, context adf.ConversionContext) (adf.Node, int, error) {
	firstLine := strings.TrimSpace(lines[startIndex])

	// Extract type from [!TYPE]
	matches := admonitionRegex.FindStringSubmatch(firstLine)
	if len(matches) < 2 {
		return adf.Node{}, 0, fmt.Errorf("invalid admonition syntax: %s", firstLine)
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
		return adf.Node{}, 0, fmt.Errorf("parsing admonition content: %w", err)
	}

	node := adf.Node{
		Type:    adf.NodeTypePanel,
		Attrs:   map[string]any{"panelType": panelType},
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
func (pc *panelRenderer) extractPanelType(node adf.Node) string {
	if node.Attrs != nil {
		if pt, ok := node.Attrs["panelType"].(string); ok && pt != "" {
			return pt
		}
	}
	return "info"
}
