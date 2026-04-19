package converter

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
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
		Registry:           globalRegistry,
		ParseNested: func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
			nested := NewMarkdownParserWithNesting(p.session, p.manager, nestedLevel)
			return nested.ParseMarkdownToADFNodes(lines)
		},
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

// parseNext dispatches to the first matching BlockParser or falls back to paragraph.
func (p *MarkdownParser) parseNext(lines []string) (*adf_types.ADFNode, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	line := strings.TrimSpace(lines[0])

	// Placeholder: infrastructure special case, no converter
	if strings.HasPrefix(line, "<!--") {
		return parsePlaceholderNode(lines, p.manager)
	}

	// Block parser dispatch (first match wins, order from RegisterDefaultConverters)
	for _, entry := range globalRegistry.BlockParsers() {
		if entry.Parser.CanParseLine(line) {
			node, consumed, err := entry.Parser.FromMarkdown(lines, 0, p.conversionContext())
			return &node, consumed, err
		}
	}

	// Fallback: paragraph
	conv := globalRegistry.GetConverter("paragraph")
	if conv == nil {
		return nil, 1, fmt.Errorf("paragraph converter not registered")
	}
	node, consumed, err := conv.FromMarkdown(lines, 0, p.conversionContext())
	return &node, consumed, err
}

// GetStackDepth returns current stack depth for debugging/monitoring
func (p *MarkdownParser) GetStackDepth() int {
	return len(p.stack)
}

// IsStackEmpty checks if parser stack is clean
func (p *MarkdownParser) IsStackEmpty() bool {
	return len(p.stack) == 0
}
