package converter

import (
	"fmt"
	"regexp"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// MarkdownParser processes markdown using stack-based iterative parsing
type MarkdownParser struct {
	stack       []*parseFrame
	session     *placeholder.EditSession
	manager     placeholder.Manager
	nestedLevel int // Nesting depth inherited from parent parser (for recursive content)
}

// parseFrame represents a parsing context on the stack.
// Currently only used for stack-depth tracking in ParseMarkdownToADFNodes.
type parseFrame struct{}

// NewMarkdownParser creates a new parser instance
func NewMarkdownParser(session *placeholder.EditSession, manager placeholder.Manager) *MarkdownParser {
	return &MarkdownParser{
		stack:   make([]*parseFrame, 0, 100), // Pre-allocate for performance
		session: session,
		manager: manager,
	}
}

// NewMarkdownParserWithNesting creates a parser that inherits a nesting level from a parent context.
func NewMarkdownParserWithNesting(session *placeholder.EditSession, manager placeholder.Manager, nestedLevel int) *MarkdownParser {
	return &MarkdownParser{
		stack:       make([]*parseFrame, 0, 100),
		session:     session,
		manager:     manager,
		nestedLevel: nestedLevel,
	}
}

// conversionContext returns a ConversionContext pre-filled with the parser's placeholder session, manager, and nesting level.
func (p *MarkdownParser) conversionContext() ConversionContext {
	return ConversionContext{
		PlaceholderSession: p.session,
		PlaceholderManager: p.manager,
		NestedLevel:        p.nestedLevel,
	}
}

// ParseMarkdownToADFNodes converts markdown lines to ADF nodes using iterative stack processing
func (p *MarkdownParser) ParseMarkdownToADFNodes(lines []string) ([]adf_types.ADFNode, error) {
	var result []adf_types.ADFNode

	// Reset parser state
	p.stack = p.stack[:0]

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Skip empty lines at document level
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Try to parse the current line
		node, consumed, err := p.parseNext(lines[i:])
		if err != nil {
			return nil, fmt.Errorf("parsing failed at line %d: %w", i+1, err)
		}

		if node != nil {
			result = append(result, *node)
		}

		i += consumed
	}

	// Verify stack is empty (all elements properly closed)
	if len(p.stack) > 0 {
		return nil, fmt.Errorf("unclosed elements detected: %d frames remain on stack", len(p.stack))
	}

	return result, nil
}

// parseNext processes the next markdown content using strategy-based parsing
func (p *MarkdownParser) parseNext(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	line := strings.TrimSpace(lines[0])

	// Strategy-based parsing for special elements
	switch {
	case strings.HasPrefix(line, "<details"):
		return p.parseExpand(lines)
	case strings.HasPrefix(line, `<div data-adf-type="blockCard"`):
		return p.parseBlockCard(lines)
	case strings.HasPrefix(line, "<table"):
		return p.parseMarkdownTable(lines)
	case strings.HasPrefix(line, "<taskList"):
		return p.parseTaskList(lines)
	case strings.HasPrefix(line, "<blockquote"):
		return p.parseBlockquote(lines)
	case strings.HasPrefix(line, "<!--"):
		return p.parsePlaceholder(lines)
	case strings.HasPrefix(line, ":::"):
		return p.parsePanel(lines)
	case isGitHubAdmonitionLine(line):
		return p.parsePanel(lines)
	case strings.HasPrefix(line, "```"):
		return p.parseCodeBlock(lines)
	case strings.HasPrefix(line, "#"):
		return p.parseHeading(lines)
	case isThematicBreak(line):
		return p.parseRule(lines)
	case strings.HasPrefix(line, "|"):
		return p.parseMarkdownTable(lines)
	case strings.HasPrefix(strings.TrimSpace(line), "- "):
		return p.parseBulletList(lines)
	default:
		if matched, _ := regexp.MatchString(`^\s*\d+\.\s`, line); matched {
			return p.parseOrderedList(lines)
		}
		return p.parseParagraph(lines)
	}
}

func (p *MarkdownParser) parseExpand(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("expand"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("expand converter not registered")
}

// Placeholder parsing methods that delegate to existing implementations
func (p *MarkdownParser) parsePlaceholder(lines []string) (*adf_types.ADFNode, int, error) {
	return parsePlaceholderNode(lines, p.session, p.manager)
}

func (p *MarkdownParser) parseTaskList(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("taskList"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("taskList converter not registered")
}

func (p *MarkdownParser) parseBlockquote(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("blockquote"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("blockquote converter not registered")
}

func (p *MarkdownParser) parseHeading(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("heading"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("heading converter not registered")
}

func (p *MarkdownParser) parseBulletList(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("bulletList"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 0, fmt.Errorf("bulletList converter not registered")
}

func (p *MarkdownParser) parseOrderedList(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("orderedList"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 0, fmt.Errorf("orderedList converter not registered")
}

// admonitionDispatchRegex matches GitHub-style admonition headers for panel dispatch
var admonitionDispatchRegex = regexp.MustCompile(`(?i)^>\s*\[!(INFO|WARNING|ERROR|SUCCESS|NOTE|TIP)\]\s*$`)

// isGitHubAdmonitionLine checks if a line matches > [!TYPE] pattern for supported panel types
func isGitHubAdmonitionLine(line string) bool {
	return admonitionDispatchRegex.MatchString(line)
}

func (p *MarkdownParser) parsePanel(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("panel"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("panel converter not registered")
}

func (p *MarkdownParser) parseCodeBlock(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("codeBlock"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("codeBlock converter not registered")
}

func (p *MarkdownParser) parseParagraph(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("paragraph"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("paragraph converter not registered")
}

func (p *MarkdownParser) parseMarkdownTable(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("table"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("table converter not registered")
}

func (p *MarkdownParser) parseRule(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("rule"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("rule converter not registered")
}

func (p *MarkdownParser) parseBlockCard(lines []string) (*adf_types.ADFNode, int, error) {
	if converter := globalRegistry.GetConverter("blockCard"); converter != nil {
		node, consumed, err := converter.FromMarkdown(lines, 0, p.conversionContext())
		return &node, consumed, err
	}
	return nil, 1, fmt.Errorf("blockCard converter not registered")
}

// isThematicBreak returns true if the line is a Markdown thematic break (---, ***, ___).
func isThematicBreak(line string) bool {
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

// GetStackDepth returns current stack depth for debugging/monitoring
func (p *MarkdownParser) GetStackDepth() int {
	return len(p.stack)
}

// IsStackEmpty checks if parser stack is clean
func (p *MarkdownParser) IsStackEmpty() bool {
	return len(p.stack) == 0
}
