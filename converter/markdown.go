package converter

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// markdownConversionContext tracks state during ADF to Markdown conversion
type markdownConversionContext struct {
	ListDepth int // Current nesting depth for lists (0 = top level)
}

// ToMarkdown converts an ADF document to Markdown, preserving complex content as placeholders.
// The registry must be populated with the element converters to dispatch to;
// use converter/defaults.NewRegistry() for the standard set.
func ToMarkdown(doc adf_types.ADFDocument, classifier ContentClassifier, manager placeholder.Manager, registry *ConverterRegistry) (string, *placeholder.EditSession, error) {
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
func convertNodeToMarkdownWithContext(node adf_types.ADFNode, ctx *markdownConversionContext, classifier ContentClassifier, manager placeholder.Manager, registry *ConverterRegistry) (string, error) {
	// Check if this node should be preserved as a placeholder
	if classifier.IsPreserved(node.Type) {
		placeholderID, preview, err := manager.Store(node)
		if err != nil {
			return "", fmt.Errorf("failed to store placeholder for %s: %w", node.Type, err)
		}

		if placeholderID == "" {
			// Display mode: preview text only, no comment wrapper
			if adf_types.IsInlineNode(node.Type) {
				return preview + " ", nil
			}
			return preview + "\n\n", nil
		}

		comment := placeholder.GeneratePlaceholderComment(placeholderID, preview)
		if adf_types.IsInlineNode(node.Type) {
			return comment + " ", nil
		}
		return comment + "\n\n", nil
	}

	// Dispatch via registry.
	nodeType := ADFNodeType(node.Type)
	if conv := registry.GetConverter(nodeType); conv != nil {
		conversionCtx := adaptContext(ctx, classifier, manager, registry, nodeType)
		result, err := conv.ToMarkdown(node, conversionCtx)
		if err != nil {
			return "", fmt.Errorf("converter failed for %s: %w", node.Type, err)
		}
		return result.Content, nil
	}

	// Unknown or unsupported node type - preserve as placeholder
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
