package adf

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/placeholder"
)

// markdownConversionContext tracks state during ADF to Markdown conversion
type markdownConversionContext struct {
	ListDepth int // Current nesting depth for lists (0 = top level)
}

// toMarkdown converts an ADF document to Markdown, preserving complex content as placeholders.
// Internal helper shared by DefaultConverter.ToMarkdown.
func toMarkdown(doc Document, classifier ContentClassifier, manager placeholder.Manager, registry *ConverterRegistry) (string, *placeholder.EditSession, error) {
	if doc.Type != "doc" {
		return "", nil, fmt.Errorf("expected document type 'doc', got '%s'", doc.Type)
	}

	var result strings.Builder
	ctx := &markdownConversionContext{ListDepth: 0}

	for _, node := range doc.Content {
		markdown, err := convertNodeToMarkdownWithContext(node, ctx, classifier, manager, registry)
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert node: %w", err)
		}
		result.WriteString(markdown)
	}

	session := manager.GetSession()
	return result.String(), session, nil
}

// convertNodeToMarkdownWithContext recursively converts an ADF node to Markdown with context
func convertNodeToMarkdownWithContext(node Node, ctx *markdownConversionContext, classifier ContentClassifier, manager placeholder.Manager, registry *ConverterRegistry) (string, error) {
	// Check if this node should be preserved as a placeholder
	if classifier.IsPreserved(node.Type) {
		placeholderID, preview, err := manager.Store(node)
		if err != nil {
			return "", fmt.Errorf("failed to store placeholder for %s: %w", node.Type, err)
		}

		if placeholderID == "" {
			// Display mode: preview text only, no comment wrapper
			if IsInlineNode(node.Type) {
				return preview + " ", nil
			}
			return preview + "\n\n", nil
		}

		comment := placeholder.GeneratePlaceholderComment(placeholderID, preview)
		if IsInlineNode(node.Type) {
			return comment + " ", nil
		}
		return comment + "\n\n", nil
	}

	// Dispatch via registry.
	nodeType := NodeType(node.Type)
	if conv, ok := registry.Lookup(nodeType); ok {
		conversionCtx := adaptContext(ctx, classifier, manager, registry, nodeType)
		result, err := conv.ToMarkdown(node, conversionCtx)
		if err != nil {
			return "", fmt.Errorf("converter failed for %s: %w", node.Type, err)
		}
		return result.Content, nil
	}

	// No registered converter for this node type - preserve as placeholder
	// so the node round-trips losslessly (ac-0094).
	placeholderID, preview, err := manager.Store(node)
	if err != nil {
		return "", fmt.Errorf("failed to store placeholder for unknown type %s: %w", node.Type, err)
	}

	if placeholderID == "" {
		return preview + "\n\n", nil
	}
	comment := placeholder.GeneratePlaceholderComment(placeholderID, preview)
	return comment + "\n\n", nil
}
