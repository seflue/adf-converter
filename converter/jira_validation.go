package converter

import (
	"fmt"
	"log/slog"

	"adf-converter/adf_types"
)

// ValidateJiraADFCompliance validates ADF structures for Jira API compliance
// This is the main validation function that ensures generated ADF will be accepted by Jira
func ValidateJiraADFCompliance(doc adf_types.ADFDocument) error {
	validator := &JiraADFValidator{}
	return validator.ValidateDocument(doc)
}

// JiraADFValidator validates ADF structures against Jira API requirements
type JiraADFValidator struct{}

// ValidateDocument validates an entire ADF document for Jira compliance
func (v *JiraADFValidator) ValidateDocument(doc adf_types.ADFDocument) error {
	var issues []string

	// Basic document structure validation
	if doc.Type != "doc" {
		issues = append(issues, fmt.Sprintf("invalid document type: %s (expected 'doc')", doc.Type))
	}
	if doc.Version != 1 {
		issues = append(issues, fmt.Sprintf("invalid document version: %d (expected 1)", doc.Version))
	}
	if len(doc.Content) == 0 {
		issues = append(issues, "document has no content nodes")
	}

	// Validate each content node
	for i, node := range doc.Content {
		nodePath := fmt.Sprintf("content[%d]", i)
		nodeIssues := v.validateNode(node, nodePath)
		issues = append(issues, nodeIssues...)
	}

	if len(issues) > 0 {
		return fmt.Errorf("jira ADF compliance issues: %v", issues)
	}
	return nil
}

// validateNode validates individual ADF nodes for Jira compliance
func (v *JiraADFValidator) validateNode(node adf_types.ADFNode, path string) []string {
	var issues []string

	// Log node validation for debugging
	slog.Debug("Jira ADF compliance validation",
		"path", path,
		"type", node.Type,
		"hasText", node.Text != "",
		"hasAttrs", node.Attrs != nil,
		"childCount", len(node.Content),
		"markCount", len(node.Marks))

	// Validate node type
	if node.Type == "" {
		issues = append(issues, fmt.Sprintf("%s: node missing type", path))
	}

	// Type-specific validation
	switch node.Type {
	case "paragraph":
		issues = append(issues, v.validateParagraphNode(node, path)...)
	case "text":
		issues = append(issues, v.validateTextNode(node, path)...)
	case "bulletList":
		issues = append(issues, v.validateBulletListNode(node, path)...)
	case "listItem":
		issues = append(issues, v.validateListItemNode(node, path)...)
	case "inlineCard":
		// This is the critical validation for our HTTP 400 fix
		issues = append(issues, v.validateInlineCardNode(node, path)...)
	default:
		slog.Debug("Unknown node type in Jira validation", "path", path, "type", node.Type)
	}

	// Recursively validate child nodes
	for i, child := range node.Content {
		childPath := fmt.Sprintf("%s.content[%d]", path, i)
		childIssues := v.validateNode(child, childPath)
		issues = append(issues, childIssues...)
	}

	return issues
}

// validateParagraphNode validates paragraph nodes for Jira compliance
func (v *JiraADFValidator) validateParagraphNode(node adf_types.ADFNode, path string) []string {
	var issues []string

	if node.Text != "" {
		issues = append(issues, fmt.Sprintf("%s: paragraph node should not have direct text", path))
	}

	return issues
}

// validateTextNode validates text nodes and their marks for Jira compliance
func (v *JiraADFValidator) validateTextNode(node adf_types.ADFNode, path string) []string {
	var issues []string

	if node.Text == "" {
		issues = append(issues, fmt.Sprintf("%s: text node missing text content", path))
	}
	if len(node.Content) > 0 {
		issues = append(issues, fmt.Sprintf("%s: text node should not have child content", path))
	}

	// Validate marks
	for i, mark := range node.Marks {
		markPath := fmt.Sprintf("%s.marks[%d]", path, i)
		markIssues := v.validateMark(mark, markPath)
		issues = append(issues, markIssues...)
	}

	return issues
}

// validateBulletListNode validates bullet list nodes for Jira compliance
func (v *JiraADFValidator) validateBulletListNode(node adf_types.ADFNode, path string) []string {
	var issues []string

	if len(node.Content) == 0 {
		issues = append(issues, fmt.Sprintf("%s: bulletList must have listItem children", path))
	}

	// All children must be listItems
	for i, child := range node.Content {
		if child.Type != "listItem" {
			issues = append(issues, fmt.Sprintf("%s.content[%d]: bulletList child must be listItem, got %s", path, i, child.Type))
		}
	}

	return issues
}

// validateListItemNode validates list item nodes for Jira compliance
func (v *JiraADFValidator) validateListItemNode(node adf_types.ADFNode, path string) []string {
	var issues []string

	if len(node.Content) == 0 {
		issues = append(issues, fmt.Sprintf("%s: listItem must have content", path))
	}

	// Check for the specific pattern that causes HTTP 400 errors
	if len(node.Content) > 1 {
		hasLinkMark := false
		hasAdditionalText := false

		for _, child := range node.Content {
			if child.Type == "text" {
				if len(child.Marks) > 0 {
					for _, mark := range child.Marks {
						if mark.Type == "link" {
							hasLinkMark = true
						}
					}
				} else if child.Text != "" && child.Text != " " {
					hasAdditionalText = true
				}
			}
		}

		if hasLinkMark && hasAdditionalText {
			issues = append(issues, fmt.Sprintf("%s: listItem has link followed by additional text (known to cause HTTP 400)", path))
		}
	}

	return issues
}

// validateInlineCardNode validates InlineCard nodes for Jira API compliance
// This is the key function that validates our Phase 5 fixes
func (v *JiraADFValidator) validateInlineCardNode(node adf_types.ADFNode, path string) []string {
	var issues []string

	slog.Debug("InlineCard Jira compliance validation",
		"path", path,
		"hasAttrs", node.Attrs != nil,
		"attrCount", func() int {
			if node.Attrs != nil {
				return len(node.Attrs)
			}
			return 0
		}())

	if node.Attrs == nil {
		issues = append(issues, fmt.Sprintf("%s: InlineCard node missing attrs", path))
		return issues
	}

	// Log all attributes for analysis
	for key, value := range node.Attrs {
		slog.Debug("InlineCard attribute validation",
			"path", path,
			"attrKey", key,
			"attrValue", value,
			"attrType", fmt.Sprintf("%T", value))
	}

	// PHASE 5 CRITICAL: Validate that InlineCard uses 'url' not 'href'
	if _, hasHref := node.Attrs["href"]; hasHref {
		issues = append(issues, fmt.Sprintf("%s: InlineCard uses forbidden 'href' attribute (must use 'url')", path))
	}

	// PHASE 5 CRITICAL: Validate required 'url' attribute
	urlVal, hasURL := node.Attrs["url"]
	if !hasURL {
		issues = append(issues, fmt.Sprintf("%s: InlineCard missing required 'url' attribute", path))
	} else {
		urlStr, ok := urlVal.(string)
		if !ok || urlStr == "" {
			issues = append(issues, fmt.Sprintf("%s: InlineCard 'url' must be non-empty string", path))
		}
	}

	// PHASE 5: Validate safe attributes only
	safeAttrs := map[string]bool{
		"url":   true,
		"title": true,
		"id":    true,
		"space": true,
	}

	for key := range node.Attrs {
		if !safeAttrs[key] {
			issues = append(issues, fmt.Sprintf("%s: InlineCard has potentially unsafe attribute '%s'", path, key))
		}
	}

	// PHASE 5: Check for conflicting attributes
	if hasHref, hasURL := node.Attrs["href"], node.Attrs["url"]; hasHref != nil && hasURL != nil {
		issues = append(issues, fmt.Sprintf("%s: InlineCard has both 'href' and 'url' attributes (conflict)", path))
	}

	return issues
}

// validateMark validates ADF marks for Jira compliance
func (v *JiraADFValidator) validateMark(mark adf_types.ADFMark, path string) []string {
	var issues []string

	if mark.Type == "" {
		issues = append(issues, fmt.Sprintf("%s: mark missing type", path))
		return issues
	}

	switch mark.Type {
	case "link":
		issues = append(issues, v.validateLinkMark(mark, path)...)
	default:
		slog.Debug("Unknown mark type in Jira validation", "path", path, "type", mark.Type)
	}

	return issues
}

// validateLinkMark validates link marks for Jira compliance
func (v *JiraADFValidator) validateLinkMark(mark adf_types.ADFMark, path string) []string {
	var issues []string

	if mark.Attrs == nil {
		issues = append(issues, fmt.Sprintf("%s: link mark missing attrs", path))
		return issues
	}

	// Validate href attribute
	href, hasHref := mark.Attrs["href"]
	if !hasHref {
		issues = append(issues, fmt.Sprintf("%s: link mark missing href attribute", path))
	} else if hrefStr, ok := href.(string); !ok || hrefStr == "" {
		issues = append(issues, fmt.Sprintf("%s: link mark href must be non-empty string", path))
	}

	// Check for excessive attributes that might cause API rejection
	if len(mark.Attrs) > 5 {
		issues = append(issues, fmt.Sprintf("%s: link mark has excessive attributes (%d), may exceed API limits", path, len(mark.Attrs)))
	}

	// Log all link attributes for debugging
	slog.Debug("Link mark attribute validation",
		"path", path,
		"href", href,
		"attrCount", len(mark.Attrs),
		"allAttrs", mark.Attrs)

	return issues
}
