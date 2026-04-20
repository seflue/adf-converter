package element

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// MarkdownParser processes markdown using iterative block-parser dispatch.
type MarkdownParser struct {
	session     *placeholder.EditSession
	manager     placeholder.Manager
	registry    Registry
	nestedLevel int // Nesting depth inherited from parent parser (for recursive content)
}

// NewMarkdownParser creates a new parser instance bound to the given registry.
func NewMarkdownParser(session *placeholder.EditSession, manager placeholder.Manager, registry Registry) *MarkdownParser {
	return &MarkdownParser{
		session:  session,
		manager:  manager,
		registry: registry,
	}
}

// NewMarkdownParserWithNesting creates a parser that inherits a nesting level from a parent context.
func NewMarkdownParserWithNesting(session *placeholder.EditSession, manager placeholder.Manager, registry Registry, nestedLevel int) *MarkdownParser {
	return &MarkdownParser{
		session:     session,
		manager:     manager,
		registry:    registry,
		nestedLevel: nestedLevel,
	}
}

// conversionContext returns a ConversionContext pre-filled with the parser's placeholder session, manager, registry, and nesting level.
func (p *MarkdownParser) conversionContext() ConversionContext {
	return ConversionContext{
		PlaceholderSession: p.session,
		PlaceholderManager: p.manager,
		NestedLevel:        p.nestedLevel,
		Registry:           p.registry,
		ParseNested: func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
			nested := NewMarkdownParserWithNesting(p.session, p.manager, p.registry, nestedLevel)
			return nested.ParseMarkdownToADFNodes(lines)
		},
	}
}

// ParseMarkdownToADFNodes converts markdown lines to ADF nodes.
func (p *MarkdownParser) ParseMarkdownToADFNodes(lines []string) ([]adf_types.ADFNode, error) {
	var result []adf_types.ADFNode

	i := 0
	for i < len(lines) {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		node, consumed, err := p.parseNext(lines[i:])
		if err != nil {
			return nil, fmt.Errorf("parsing failed at line %d: %w", i+1, err)
		}

		if node != nil {
			result = append(result, *node)
		}

		i += consumed
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

	// Block parser dispatch (first match wins, order from Registry registration)
	for _, entry := range p.registry.BlockParsers() {
		if entry.Parser.CanParseLine(line) {
			node, consumed, err := entry.Parser.FromMarkdown(lines, 0, p.conversionContext())
			return &node, consumed, err
		}
	}

	// Fallback: paragraph
	conv, ok := p.registry.Lookup("paragraph")
	if !ok {
		return nil, 1, fmt.Errorf("paragraph converter not registered")
	}
	node, consumed, err := conv.FromMarkdown(lines, 0, p.conversionContext())
	return &node, consumed, err
}
