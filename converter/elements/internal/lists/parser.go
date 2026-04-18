// Package lists provides goldmark-based list parsing for ADF conversion.
package lists

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/elements/inline"
	"github.com/seflue/adf-converter/placeholder"
)

// ParseBulletList parses a markdown bullet list string into an ADF bulletList node.
// This is the top-level entry point for converting bullet lists.
// The manager parameter is used to restore placeholder nodes (e.g., emoji, mentions) during conversion.
func ParseBulletList(markdown string, manager placeholder.Manager) (adf_types.ADFNode, error) {
	// Parse markdown with goldmark
	source := []byte(markdown)
	parser := goldmark.New()
	doc := parser.Parser().Parse(text.NewReader(source))

	// Extract root List node from AST
	list := doc.FirstChild()
	if list == nil {
		return adf_types.ADFNode{}, fmt.Errorf("no list found in markdown")
	}

	// Verify it's a List node
	listNode, ok := list.(*ast.List)
	if !ok {
		return adf_types.ADFNode{}, fmt.Errorf("expected List node, got %T", list)
	}

	// Verify it's an unordered list
	if listNode.IsOrdered() {
		return adf_types.ADFNode{}, fmt.Errorf("expected bullet list, got ordered list")
	}

	// Convert to ADF
	return convertList(listNode, source, manager)
}

// ParseOrderedList parses a markdown ordered list string into an ADF orderedList node.
// This is the top-level entry point for converting ordered lists.
// The manager parameter is used to restore placeholder nodes (e.g., emoji, mentions) during conversion.
func ParseOrderedList(markdown string, manager placeholder.Manager) (adf_types.ADFNode, error) {
	// Parse markdown with goldmark
	source := []byte(markdown)
	parser := goldmark.New()
	doc := parser.Parser().Parse(text.NewReader(source))

	// Extract root List node from AST
	list := doc.FirstChild()
	if list == nil {
		return adf_types.ADFNode{}, fmt.Errorf("no list found in markdown")
	}

	// Verify it's a List node
	listNode, ok := list.(*ast.List)
	if !ok {
		return adf_types.ADFNode{}, fmt.Errorf("expected List node, got %T", list)
	}

	// Verify it's an ordered list
	if !listNode.IsOrdered() {
		return adf_types.ADFNode{}, fmt.Errorf("expected ordered list, got bullet list")
	}

	// Convert to ADF
	return convertList(listNode, source, manager)
}

// ConvertListNode converts a goldmark List AST node into an ADF bulletList or orderedList node.
// It is exported for use by other converters (e.g. blockquote) that already have a parsed AST.
func ConvertListNode(list *ast.List, source []byte, manager placeholder.Manager) (adf_types.ADFNode, error) {
	return convertList(list, source, manager)
}

// convertList converts a goldmark List AST node into an ADF bulletList or orderedList node.
// This handles the recursive conversion of nested lists and list items.
func convertList(list *ast.List, source []byte, manager placeholder.Manager) (adf_types.ADFNode, error) {
	// Determine list type
	listType := adf_types.NodeTypeBulletList
	if list.IsOrdered() {
		listType = adf_types.NodeTypeOrderedList
	}

	// Convert each list item
	var items []adf_types.ADFNode
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if listItem, ok := child.(*ast.ListItem); ok {
			item, err := convertListItem(listItem, source, manager)
			if err != nil {
				return adf_types.ADFNode{}, err
			}
			items = append(items, item)
		}
	}

	// Capture start number for ordered lists (suppress default of 1)
	var attrs map[string]interface{}
	if list.IsOrdered() && list.Start != 1 {
		attrs = map[string]interface{}{"order": float64(list.Start)}
	}

	return adf_types.ADFNode{
		Type:    listType,
		Attrs:   attrs,
		Content: items,
	}, nil
}

// convertListItem converts a goldmark ListItem AST node into an ADF listItem node.
// This handles the conversion of list item content, including text blocks and nested lists.
func convertListItem(listItem *ast.ListItem, source []byte, manager placeholder.Manager) (adf_types.ADFNode, error) {
	var content []adf_types.ADFNode

	// Process each child of the list item
	for child := listItem.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.TextBlock:
			// Convert text content to paragraph using inline parser
			para, err := convertTextBlock(n, source, manager)
			if err != nil {
				return adf_types.ADFNode{}, err
			}
			content = append(content, para)

		case *ast.HTMLBlock:
			// Handle HTML blocks (e.g., placeholder comments) as text content
			// Extract the raw HTML and parse it as inline content
			para, err := convertHTMLBlock(n, source, manager)
			if err != nil {
				return adf_types.ADFNode{}, err
			}
			content = append(content, para)

		case *ast.List:
			// Recursively convert nested list
			nestedList, err := convertList(n, source, manager)
			if err != nil {
				return adf_types.ADFNode{}, err
			}
			content = append(content, nestedList)
		}
	}

	return adf_types.ADFNode{
		Type:    adf_types.NodeTypeListItem,
		Content: content,
	}, nil
}

// convertHTMLBlock converts HTML blocks (like placeholder comments) to paragraphs
func convertHTMLBlock(htmlBlock *ast.HTMLBlock, source []byte, manager placeholder.Manager) (adf_types.ADFNode, error) {
	// Extract all lines from the HTML block
	var rawText strings.Builder
	for i := 0; i < htmlBlock.Lines().Len(); i++ {
		line := htmlBlock.Lines().At(i)
		rawText.Write(source[line.Start:line.Stop])
	}

	text := strings.TrimSpace(rawText.String())
	if text == "" {
		return adf_types.ADFNode{
			Type:    adf_types.NodeTypeParagraph,
			Content: []adf_types.ADFNode{},
		}, nil
	}

	// Parse as inline content with placeholder restoration
	inlineContent, err := inline.ParseContentWithPlaceholders(text, manager)
	if err != nil {
		return adf_types.ADFNode{}, err
	}

	return adf_types.ADFNode{
		Type:    adf_types.NodeTypeParagraph,
		Content: inlineContent,
	}, nil
}

// convertTextBlock converts a goldmark TextBlock AST node into an ADF paragraph node.
// For multiline list items, creates text nodes with hardBreak nodes between lines.
func convertTextBlock(textBlock *ast.TextBlock, source []byte, manager placeholder.Manager) (adf_types.ADFNode, error) {
	if textBlock.Lines().Len() == 0 {
		return adf_types.ADFNode{
			Type:    adf_types.NodeTypeParagraph,
			Content: []adf_types.ADFNode{},
		}, nil
	}

	var allContent []adf_types.ADFNode

	// Process each line separately to create text + hardBreak structure for multiline items
	for i := 0; i < textBlock.Lines().Len(); i++ {
		line := textBlock.Lines().At(i)
		rawText := strings.TrimSpace(string(source[line.Start:line.Stop]))

		if rawText == "" {
			// Skip empty lines (but preserve hardBreak from previous line if any)
			continue
		}

		// Parse this line for inline formatting with placeholder restoration
		inlineContent, err := inline.ParseContentWithPlaceholders(rawText, manager)
		if err != nil {
			return adf_types.ADFNode{}, err
		}

		// Add the inline content from this line
		allContent = append(allContent, inlineContent...)

		// Add hardBreak after each line except the last
		if i < textBlock.Lines().Len()-1 {
			allContent = append(allContent, adf_types.ADFNode{
				Type: adf_types.NodeTypeHardBreak,
			})
		}
	}

	return adf_types.ADFNode{
		Type:    adf_types.NodeTypeParagraph,
		Content: allContent,
	}, nil
}
