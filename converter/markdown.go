package converter

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// Global converter registry for incremental migration
// Starts empty; converters are registered as they are extracted from the switch statement
//
// IMPORTANT: Call RegisterDefaultConverters() before using ToMarkdown/FromMarkdown
// to register migrated element converters. This must be done explicitly to avoid
// circular import dependencies (converter ← elements → converter).
var globalRegistry = NewConverterRegistry()

// RegisterDefaultConverters registers all migrated element converters
//
// This function must be called once before using ToMarkdown or FromMarkdown.
// Typically called from application init() or test setup:
//
//	func init() {
//	    converter.RegisterDefaultConverters()
//	}
//
// The function is idempotent - calling it multiple times is safe.
//
// Note: We cannot import elements here due to circular dependency.
// Callers must pass converters via the elementConverters parameter.
func RegisterDefaultConverters(elementConverters ...ElementConverter) {
	if globalRegistry.Count() > 0 {
		return // Already registered
	}

	nodeTypes := []ADFNodeType{
		"text", "hardBreak", "paragraph", "heading",
		"listItem", "bulletList", "orderedList",
		"expand", "nestedExpand", "inlineCard", "emoji",
		"codeBlock", "rule", "mention", "table", "panel", "date", "status",
	}

	for _, converter := range elementConverters {
		if converter != nil {
			for _, nodeType := range nodeTypes {
				if converter.CanHandle(nodeType) {
					globalRegistry.Register(nodeType, converter)
				}
			}
		}
	}
}

// GetGlobalRegistry provides access to the global registry for manual registration
//
// Use this to register converters from test setup or application code:
//
//	import "adf-converter/converter/elements"
//	converter.GetGlobalRegistry().Register(adf_types.NodeTypeText, elements.NewTextConverter())
func GetGlobalRegistry() *ConverterRegistry {
	return globalRegistry
}

// MarkdownConversionContext tracks state during ADF to Markdown conversion
type MarkdownConversionContext struct {
	ListDepth int // Current nesting depth for lists (0 = top level)
}

// ToMarkdown converts an ADF document to Markdown, preserving complex content as placeholders
func ToMarkdown(doc adf_types.ADFDocument, classifier ContentClassifier, manager placeholder.Manager) (string, *placeholder.EditSession, error) {
	if doc.Type != "doc" {
		return "", nil, fmt.Errorf("expected document type 'doc', got '%s'", doc.Type)
	}

	var result strings.Builder
	ctx := &MarkdownConversionContext{ListDepth: 0}

	for _, node := range doc.Content {
		markdown, err := convertNodeToMarkdownWithContext(node, ctx, classifier, manager)
		if err != nil {
			return "", nil, fmt.Errorf("failed to convert node: %w", err)
		}
		result.WriteString(markdown)
	}

	session := manager.GetSession()
	return result.String(), session, nil
}

// convertNodeToMarkdownWithContext recursively converts an ADF node to Markdown with context
func convertNodeToMarkdownWithContext(node adf_types.ADFNode, ctx *MarkdownConversionContext, classifier ContentClassifier, manager placeholder.Manager) (string, error) {
	// Check if this node should be preserved as a placeholder
	if classifier.IsPreserved(node.Type) {
		placeholderID, preview, err := manager.Store(node)
		if err != nil {
			return "", fmt.Errorf("failed to store placeholder for %s: %w", node.Type, err)
		}

		// Generate a markdown comment with the placeholder
		comment := placeholder.GeneratePlaceholderComment(placeholderID, preview)

		// Use inline spacing for inline nodes, block spacing for block nodes
		if adf_types.IsInlineNode(node.Type) {
			return comment + " ", nil
		}
		return comment + "\n\n", nil
	}

	// Try registry first (NEW - incremental migration pattern)
	// Registry starts empty, so this has zero behavior change initially.
	// As converters are registered, they take precedence over the switch statement.
	nodeType := ADFNodeType(node.Type)
	if converter := globalRegistry.GetConverter(nodeType); converter != nil {
		// Adapt context from legacy MarkdownConversionContext to ConversionContext
		conversionCtx := adaptContext(ctx, classifier, manager, nodeType)

		// Use registered converter
		result, err := converter.ToMarkdown(node, conversionCtx)
		if err != nil {
			return "", fmt.Errorf("converter failed for %s: %w", node.Type, err)
		}

		// Extract markdown content from enhanced result
		return result.Content, nil
	}

	// Unknown or unsupported node type - preserve as placeholder
	placeholderID, preview, err := manager.Store(node)
	if err != nil {
		return "", fmt.Errorf("failed to store placeholder for unknown type %s: %w", node.Type, err)
	}

	comment := placeholder.GeneratePlaceholderComment(placeholderID, preview)
	return comment + "\n\n", nil
}
