package converter

import (
	"fmt"

	"adf-converter/adf_types"
)

// ADFStructureValidator provides utility functions to validate ADF structure compliance

// ValidateListItemStructure validates that listItem nodes contain only paragraphs as direct children
func ValidateListItemStructure(node adf_types.ADFNode) error {
	violations := []string{}
	validateListItemStructureRecursive([]adf_types.ADFNode{node}, &violations, "root")

	if len(violations) > 0 {
		return fmt.Errorf("ADF structure violations found: %v", violations)
	}
	return nil
}

// ValidateADFCompliance validates overall ADF document structure compliance
func ValidateADFCompliance(doc adf_types.ADFDocument) []string {
	violations := []string{}

	// Validate document-level structure
	if doc.Type != "doc" {
		violations = append(violations, "document type must be 'doc', got: "+doc.Type)
	}

	if doc.Version != 1 {
		violations = append(violations, fmt.Sprintf("document version must be 1, got: %d", doc.Version))
	}

	// Validate content structure
	validateADFComplianceRecursive(doc.Content, &violations, "document")

	return violations
}

// validateListItemStructureRecursive recursively validates listItem structure
func validateListItemStructureRecursive(nodes []adf_types.ADFNode, violations *[]string, path string) {
	for i, node := range nodes {
		currentPath := fmt.Sprintf("%s.%s[%d]", path, node.Type, i)

		if node.Type == adf_types.NodeTypeListItem {
			// CRITICAL RULE: listItem nodes MUST contain only paragraph nodes as direct children
			for j, child := range node.Content {
				if child.Type != adf_types.NodeTypeParagraph {
					*violations = append(*violations,
						fmt.Sprintf("%s.content[%d] contains %s instead of paragraph", currentPath, j, child.Type))
				}
			}
		}

		// Recursively check child nodes
		if len(node.Content) > 0 {
			validateListItemStructureRecursive(node.Content, violations, currentPath)
		}
	}
}

// validateADFComplianceRecursive validates general ADF structure rules
func validateADFComplianceRecursive(nodes []adf_types.ADFNode, violations *[]string, path string) {
	for i, node := range nodes {
		currentPath := fmt.Sprintf("%s.%s[%d]", path, node.Type, i)

		// Validate node-specific rules
		switch node.Type {
		case adf_types.NodeTypeListItem:
			// listItem nodes must contain paragraphs
			if len(node.Content) == 0 {
				*violations = append(*violations, currentPath+" is empty (listItem must contain paragraphs)")
			}
			for j, child := range node.Content {
				if child.Type != adf_types.NodeTypeParagraph {
					*violations = append(*violations,
						fmt.Sprintf("%s.content[%d] contains %s instead of paragraph", currentPath, j, child.Type))
				}
			}

		case adf_types.NodeTypeBulletList, adf_types.NodeTypeOrderedList:
			// List nodes must contain listItem nodes
			if len(node.Content) == 0 {
				*violations = append(*violations, currentPath+" is empty (lists must contain listItems)")
			}
			for j, child := range node.Content {
				if child.Type != adf_types.NodeTypeListItem {
					*violations = append(*violations,
						fmt.Sprintf("%s.content[%d] contains %s instead of listItem", currentPath, j, child.Type))
				}
			}

		case adf_types.NodeTypeParagraph:
			// Paragraph nodes should contain text, marks, or inline elements
			if len(node.Content) == 0 {
				*violations = append(*violations, currentPath+" is empty (paragraphs should contain content)")
			}

		case adf_types.NodeTypeText:
			// Text nodes should have text content
			if node.Text == "" {
				*violations = append(*violations, currentPath+" has empty text")
			}

		case adf_types.NodeTypeInlineCard:
			// InlineCard nodes should have url attribute
			if node.Attrs == nil || node.Attrs["url"] == nil {
				*violations = append(*violations, currentPath+" missing required url attribute")
			}
		}

		// Recursively check child nodes
		if len(node.Content) > 0 {
			validateADFComplianceRecursive(node.Content, violations, currentPath)
		}
	}
}

// ValidateDocumentStructure validates the structure of a complete ADF document
func ValidateDocumentStructure(doc adf_types.ADFDocument) error {
	violations := ValidateADFCompliance(doc)
	if len(violations) > 0 {
		return fmt.Errorf("ADF document has structure violations: %v", violations)
	}
	return nil
}

// ValidateNodeHierarchy validates that a node follows proper ADF hierarchy rules
func ValidateNodeHierarchy(node adf_types.ADFNode, parentType string) []string {
	violations := []string{}

	// Define valid parent-child relationships
	validChildren := map[string][]string{
		"doc":                         {adf_types.NodeTypeParagraph, adf_types.NodeTypeBulletList, adf_types.NodeTypeOrderedList, adf_types.NodeTypeHeading},
		adf_types.NodeTypeParagraph:   {adf_types.NodeTypeText, adf_types.NodeTypeInlineCard},
		adf_types.NodeTypeBulletList:  {adf_types.NodeTypeListItem},
		adf_types.NodeTypeOrderedList: {adf_types.NodeTypeListItem},
		adf_types.NodeTypeListItem:    {adf_types.NodeTypeParagraph},
		adf_types.NodeTypeHeading:     {adf_types.NodeTypeText},
	}

	if validTypes, exists := validChildren[parentType]; exists {
		validType := false
		for _, validNodeType := range validTypes {
			if node.Type == validNodeType {
				validType = true
				break
			}
		}
		if !validType {
			violations = append(violations,
				fmt.Sprintf("invalid child type %s under parent %s (valid: %v)", node.Type, parentType, validTypes))
		}
	}

	// Recursively validate children
	for _, child := range node.Content {
		childViolations := ValidateNodeHierarchy(child, node.Type)
		violations = append(violations, childViolations...)
	}

	return violations
}

// IsValidADFStructure performs a quick validation check
func IsValidADFStructure(doc adf_types.ADFDocument) bool {
	violations := ValidateADFCompliance(doc)
	return len(violations) == 0
}

// GetStructureViolationSummary returns a human-readable summary of structure violations
func GetStructureViolationSummary(doc adf_types.ADFDocument) string {
	violations := ValidateADFCompliance(doc)
	if len(violations) == 0 {
		return "✅ ADF structure is valid"
	}

	summary := fmt.Sprintf("❌ Found %d ADF structure violations:\n", len(violations))
	for i, violation := range violations {
		summary += fmt.Sprintf("  %d. %s\n", i+1, violation)
	}
	return summary
}
