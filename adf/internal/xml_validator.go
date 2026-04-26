package internal

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// XMLValidator provides validation utilities for ADF-preserved XML elements
type XMLValidator struct {
	schema *XMLSchema
}

// NewXMLValidator creates a new XML validator with the default schema
func NewXMLValidator() *XMLValidator {
	return &XMLValidator{
		schema: NewDefaultXMLSchema(),
	}
}

// XMLSchema defines the validation rules for ADF XML elements
type XMLSchema struct {
	elements map[string]ElementSchema
}

// ElementSchema defines validation rules for a specific XML element
type ElementSchema struct {
	Name              string
	RequiredAttrs     []string
	OptionalAttrs     []string
	AllowedChildren   []string
	ContentType       ContentType
	ValidationPattern *regexp.Regexp
}

// ContentType defines what type of content an element can contain
type ContentType int

const (
	// EmptyContent means the element should have no content
	EmptyContent ContentType = iota
	// TextContent means the element can contain text
	TextContent
	// ElementContent means the element can contain child elements
	ElementContent
	// MixedContent means the element can contain both text and child elements
	MixedContent
)

// NewDefaultXMLSchema creates the default schema for ADF XML elements
func NewDefaultXMLSchema() *XMLSchema {
	schema := &XMLSchema{
		elements: make(map[string]ElementSchema),
	}

	// Define schema for expand elements
	schema.elements["expand"] = ElementSchema{
		Name:            "expand",
		RequiredAttrs:   []string{"title"},
		OptionalAttrs:   []string{"localId"},
		AllowedChildren: []string{},
		ContentType:     TextContent,
	}

	// Define schema for mention elements
	schema.elements["mention"] = ElementSchema{
		Name:              "mention",
		RequiredAttrs:     []string{"id", "text"},
		OptionalAttrs:     []string{"accessLevel", "userType"},
		ContentType:       TextContent,
		ValidationPattern: regexp.MustCompile(`^@.+$`), // Must start with @
	}

	// Define schema for taskList elements
	schema.elements["taskList"] = ElementSchema{
		Name:            "taskList",
		RequiredAttrs:   []string{"localId"},
		OptionalAttrs:   []string{},
		AllowedChildren: []string{"taskItem"},
		ContentType:     ElementContent,
	}

	// Define schema for taskItem elements
	schema.elements["taskItem"] = ElementSchema{
		Name:              "taskItem",
		RequiredAttrs:     []string{"localId", "state"},
		OptionalAttrs:     []string{},
		ContentType:       TextContent,
		ValidationPattern: regexp.MustCompile(`^- \[[ x]\] .+`), // Checkbox format
	}

	// Define schema for hardBreak elements
	schema.elements["hardBreak"] = ElementSchema{
		Name:          "hardBreak",
		RequiredAttrs: []string{},
		OptionalAttrs: []string{"localId"},
		ContentType:   EmptyContent,
	}

	// Define schema for generic adf-node wrapper
	schema.elements["adf-node"] = ElementSchema{
		Name:          "adf-node",
		RequiredAttrs: []string{"type"},
		OptionalAttrs: []string{"localId"},
		ContentType:   MixedContent,
	}

	return schema
}

// ValidateElement validates a single XML element against the schema
func (v *XMLValidator) ValidateElement(elementName string, attrs map[string]string, content string, children []string) error {
	schema, exists := v.schema.elements[elementName]
	if !exists {
		return fmt.Errorf("unknown element type: %s", elementName)
	}

	// Validate required attributes
	for _, requiredAttr := range schema.RequiredAttrs {
		if _, exists := attrs[requiredAttr]; !exists {
			return fmt.Errorf("element %s missing required attribute: %s", elementName, requiredAttr)
		}
	}

	// Validate attribute names (optional check)
	allowedAttrs := make(map[string]bool)
	for _, attr := range schema.RequiredAttrs {
		allowedAttrs[attr] = true
	}
	for _, attr := range schema.OptionalAttrs {
		allowedAttrs[attr] = true
	}

	// Unknown attributes are silently accepted for forward-compatibility.
	// Validation only checks required attrs and content type below.

	// Validate content type
	switch schema.ContentType {
	case EmptyContent:
		if content != "" || len(children) > 0 {
			return fmt.Errorf("element %s should be empty but has content", elementName)
		}
	case TextContent:
		if len(children) > 0 {
			return fmt.Errorf("element %s should contain only text but has child elements", elementName)
		}
	case ElementContent:
		if content != "" && strings.TrimSpace(content) != "" {
			return fmt.Errorf("element %s should contain only child elements but has text content", elementName)
		}
	case MixedContent:
		// Mixed content is always allowed
	}

	// Validate children
	if len(schema.AllowedChildren) > 0 {
		allowedChildren := make(map[string]bool)
		for _, child := range schema.AllowedChildren {
			allowedChildren[child] = true
		}

		for _, child := range children {
			if !allowedChildren[child] {
				return fmt.Errorf("element %s cannot contain child element: %s", elementName, child)
			}
		}
	}

	// Validate content pattern
	if schema.ValidationPattern != nil && content != "" {
		if !schema.ValidationPattern.MatchString(content) {
			return fmt.Errorf("element %s content does not match required pattern", elementName)
		}
	}

	return nil
}

// ValidateXMLStructure validates the overall structure of XML data
func (v *XMLValidator) ValidateXMLStructure(data []byte) error {
	// Parse the XML to get structure information
	var doc XMLDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse XML structure: %w", err)
	}

	// Validate the root element
	return v.validateElementRecursive(doc.Root)
}

// XMLDocument represents a parsed XML document for validation
type XMLDocument struct {
	Root XMLElement `xml:",any"`
}

// XMLElement represents an XML element with attributes and children
type XMLElement struct {
	XMLName  xml.Name
	Attrs    []xml.Attr   `xml:",any,attr"`
	Content  string       `xml:",chardata"`
	Children []XMLElement `xml:",any"`
}

// validateElementRecursive validates an element and all its children
func (v *XMLValidator) validateElementRecursive(element XMLElement) error {
	// Convert attributes to map
	attrs := make(map[string]string)
	for _, attr := range element.Attrs {
		attrs[attr.Name.Local] = attr.Value
	}

	// Get child element names
	var childNames []string
	for _, child := range element.Children {
		childNames = append(childNames, child.XMLName.Local)
	}

	// Validate this element
	if err := v.ValidateElement(element.XMLName.Local, attrs, element.Content, childNames); err != nil {
		return fmt.Errorf("validating element %s: %w", element.XMLName.Local, err)
	}

	// Recursively validate children
	for _, child := range element.Children {
		if err := v.validateElementRecursive(child); err != nil {
			return fmt.Errorf("validating child of %s: %w", element.XMLName.Local, err)
		}
	}

	return nil
}

// ValidateSpecificElement validates a specific element type with custom validation
func (v *XMLValidator) ValidateSpecificElement(elementType string, data []byte) error {
	switch elementType {
	case "expand":
		return v.validateExpandElement(data)
	case "mention":
		return v.validateMentionElement(data)
	case "taskList":
		return v.validateTaskListElement(data)
	case "taskItem":
		return v.validateTaskItemElement(data)
	case "hardBreak":
		return v.validateHardBreakElement(data)
	default:
		return v.ValidateXMLStructure(data)
	}
}

// validateExpandElement performs specific validation for expand elements
func (v *XMLValidator) validateExpandElement(data []byte) error {
	var expand ExpandElement
	if err := xml.Unmarshal(data, &expand); err != nil {
		return fmt.Errorf("invalid expand element: %w", err)
	}

	// Validate title length (reasonable limit)
	if len(expand.Title) > 200 {
		return fmt.Errorf("expand title too long (max 200 characters)")
	}

	// Validate content if present
	if expand.Content != "" {
		if len(expand.Content) > 10000 {
			return fmt.Errorf("expand content too long (max 10000 characters)")
		}
	}

	return nil
}

// validateMentionElement performs specific validation for mention elements
func (v *XMLValidator) validateMentionElement(data []byte) error {
	var mention MentionElement
	if err := xml.Unmarshal(data, &mention); err != nil {
		return fmt.Errorf("invalid mention element: %w", err)
	}

	// Validate required fields
	if mention.ID == "" {
		return fmt.Errorf("mention element missing required id attribute")
	}
	if mention.Text == "" {
		return fmt.Errorf("mention element missing required text attribute")
	}

	// Validate ID format (should be a valid user identifier)
	if len(mention.ID) < 1 || len(mention.ID) > 100 {
		return fmt.Errorf("mention id invalid length")
	}

	// Validate content format
	if mention.Content != "" && !strings.HasPrefix(mention.Content, "@") {
		return fmt.Errorf("mention content should start with @")
	}

	// Validate access level if present
	if mention.AccessLevel != "" {
		validLevels := map[string]bool{
			"reader": true,
			"writer": true,
			"admin":  true,
		}
		if !validLevels[mention.AccessLevel] {
			return fmt.Errorf("invalid access level: %s", mention.AccessLevel)
		}
	}

	return nil
}

// validateTaskListElement performs specific validation for task list elements
func (v *XMLValidator) validateTaskListElement(data []byte) error {
	var taskList TaskListElement
	if err := xml.Unmarshal(data, &taskList); err != nil {
		return fmt.Errorf("invalid task list element: %w", err)
	}

	// Validate required fields
	if taskList.LocalID == "" {
		return fmt.Errorf("task list element missing required localId attribute")
	}

	// Validate child task items
	for i, child := range taskList.Children {
		if child.LocalID == "" {
			return fmt.Errorf("task item %d missing required localId attribute", i)
		}
		if child.State == "" {
			return fmt.Errorf("task item %d missing required state attribute", i)
		}

		// Validate state values
		if child.State != "TODO" && child.State != "DONE" {
			return fmt.Errorf("task item %d has invalid state: %s", i, child.State)
		}

		// Validate content format
		if child.Content != "" {
			expectedPrefix := "- [ ] "
			if child.State == "DONE" {
				expectedPrefix = "- [x] "
			}
			if !strings.HasPrefix(child.Content, expectedPrefix) {
				return fmt.Errorf("task item %d content does not match expected checkbox format", i)
			}
		}
	}

	return nil
}

// validateTaskItemElement performs specific validation for individual task item elements
func (v *XMLValidator) validateTaskItemElement(data []byte) error {
	var taskItem TaskElement
	if err := xml.Unmarshal(data, &taskItem); err != nil {
		return fmt.Errorf("invalid task item element: %w", err)
	}

	// Validate required fields
	if taskItem.LocalID == "" {
		return fmt.Errorf("task item element missing required localId attribute")
	}
	if taskItem.State == "" {
		return fmt.Errorf("task item element missing required state attribute")
	}

	// Validate state
	if taskItem.State != "TODO" && taskItem.State != "DONE" {
		return fmt.Errorf("invalid task item state: %s", taskItem.State)
	}

	// Validate content format
	if taskItem.Content != "" {
		expectedPrefix := "- [ ] "
		if taskItem.State == "DONE" {
			expectedPrefix = "- [x] "
		}
		if !strings.HasPrefix(taskItem.Content, expectedPrefix) {
			return fmt.Errorf("task item content does not match expected checkbox format")
		}
	}

	return nil
}

// validateHardBreakElement performs specific validation for hard break elements
func (v *XMLValidator) validateHardBreakElement(data []byte) error {
	var hardBreak HardBreakElement
	if err := xml.Unmarshal(data, &hardBreak); err != nil {
		return fmt.Errorf("invalid hard break element: %w", err)
	}

	// Hard break elements should be empty
	// localId is optional, no other validation needed

	return nil
}

// GetValidationReport provides a detailed validation report for XML data
func (v *XMLValidator) GetValidationReport(data []byte) ValidationReport {
	report := ValidationReport{
		IsValid:      true,
		Errors:       make([]string, 0),
		Warnings:     make([]string, 0),
		ElementCount: 0,
	}

	// Basic well-formedness check
	if err := v.ValidateXMLStructure(data); err != nil {
		report.IsValid = false
		report.Errors = append(report.Errors, err.Error())
		return report
	}

	// Count elements and check for potential issues
	var doc XMLDocument
	if err := xml.Unmarshal(data, &doc); err == nil {
		report.ElementCount = v.countElements(doc.Root)

		// Check for overly complex structures
		if report.ElementCount > 100 {
			report.Warnings = append(report.Warnings, "XML structure is very complex (>100 elements)")
		}

		// Check for deeply nested structures
		maxDepth := v.calculateMaxDepth(doc.Root, 0)
		if maxDepth > 10 {
			report.Warnings = append(report.Warnings, fmt.Sprintf("XML structure is deeply nested (depth: %d)", maxDepth))
		}
	}

	return report
}

// ValidationReport provides detailed information about XML validation
type ValidationReport struct {
	IsValid      bool
	Errors       []string
	Warnings     []string
	ElementCount int
}

// countElements recursively counts the number of elements in the XML
func (v *XMLValidator) countElements(element XMLElement) int {
	count := 1 // Count this element
	for _, child := range element.Children {
		count += v.countElements(child)
	}
	return count
}

// calculateMaxDepth calculates the maximum nesting depth
func (v *XMLValidator) calculateMaxDepth(element XMLElement, currentDepth int) int {
	maxDepth := currentDepth
	for _, child := range element.Children {
		childDepth := v.calculateMaxDepth(child, currentDepth+1)
		if childDepth > maxDepth {
			maxDepth = childDepth
		}
	}
	return maxDepth
}
