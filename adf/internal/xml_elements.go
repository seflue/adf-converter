package internal

import (
	"encoding/xml"

	"github.com/seflue/adf-converter/adf"
)

// XML element structures for ADF-specific nodes
// These structures define the XML representation of ADF elements that have no direct markdown equivalent

// ExpandElement represents an ADF expand section preserved as XML
type ExpandElement struct {
	XMLName xml.Name `xml:"expand"`
	Title   string   `xml:"title,attr"`
	LocalID string   `xml:"localId,attr,omitempty"`
	Content string   `xml:",chardata"`

	// For now, handle a few common custom attributes explicitly
	// This will be enhanced in the refactor phase
	CustomAttr string `xml:"customAttr,attr,omitempty"`
	Numeric    string `xml:"numeric,attr,omitempty"`
	Boolean    string `xml:"boolean,attr,omitempty"`
}

// MentionElement represents an ADF mention preserved as XML
type MentionElement struct {
	XMLName     xml.Name `xml:"mention"`
	ID          string   `xml:"id,attr"`
	Text        string   `xml:"text,attr"`
	AccessLevel string   `xml:"accessLevel,attr,omitempty"`
	UserType    string   `xml:"userType,attr,omitempty"`
	Content     string   `xml:",chardata"` // @username format
}

// TaskListElement represents an ADF task list (when preservation is needed)
type TaskListElement struct {
	XMLName  xml.Name      `xml:"taskList"`
	LocalID  string        `xml:"localId,attr"`
	Children []TaskElement `xml:"taskItem"`
}

// TaskElement represents an individual task item within a task list
type TaskElement struct {
	XMLName xml.Name `xml:"taskItem"`
	LocalID string   `xml:"localId,attr"`
	State   string   `xml:"state,attr"` // Valid values: TODO, DONE
	Content string   `xml:",chardata"`  // Markdown checkbox syntax
}

// HardBreakElement represents an ADF hard break preserved as XML
type HardBreakElement struct {
	XMLName xml.Name `xml:"hardBreak"`
	LocalID string   `xml:"localId,attr,omitempty"`
}

// GenericADFElement represents a generic wrapper for any ADF node
type GenericADFElement struct {
	XMLName xml.Name `xml:"adf-node"`
	Type    string   `xml:"type,attr"`
	LocalID string   `xml:"localId,attr,omitempty"`
	Content string   `xml:",chardata"`
}

// TableElement represents an ADF table (when XML preservation is needed)
// Note: Tables normally use markdown-native conversion, but this provides fallback
type TableElement struct {
	XMLName               xml.Name   `xml:"table"`
	IsNumberColumnEnabled bool       `xml:"isNumberColumnEnabled,attr,omitempty"`
	Layout                string     `xml:"layout,attr,omitempty"`
	LocalID               string     `xml:"localId,attr,omitempty"`
	Rows                  []TableRow `xml:"tableRow"`
}

// TableRow represents a row within a table
type TableRow struct {
	XMLName xml.Name    `xml:"tableRow"`
	Cells   []TableCell `xml:",any"`
}

// TableCell represents a cell within a table row
type TableCell struct {
	XMLName    xml.Name `xml:"tableCell"`
	Colspan    int      `xml:"colspan,attr,omitempty"`
	Rowspan    int      `xml:"rowspan,attr,omitempty"`
	Background string   `xml:"background,attr,omitempty"`
	Content    string   `xml:",chardata"`
}

// TableHeader represents a header cell within a table row
type TableHeader struct {
	XMLName    xml.Name `xml:"tableHeader"`
	Colspan    int      `xml:"colspan,attr,omitempty"`
	Rowspan    int      `xml:"rowspan,attr,omitempty"`
	Background string   `xml:"background,attr,omitempty"`
	Content    string   `xml:",chardata"`
}

// MediaGroupElement represents an ADF media group for future extension
type MediaGroupElement struct {
	XMLName xml.Name       `xml:"mediaGroup"`
	Media   []MediaElement `xml:"media"`
}

// MediaElement represents a media item within a media group
type MediaElement struct {
	XMLName    xml.Name `xml:"media"`
	ID         string   `xml:"id,attr"`
	Type       string   `xml:"type,attr"`
	Collection string   `xml:"collection,attr,omitempty"`
	Width      int      `xml:"width,attr,omitempty"`
	Height     int      `xml:"height,attr,omitempty"`
}

// XMLElementBuilder helps construct XML elements programmatically
type XMLElementBuilder struct {
	elementType string
	attributes  map[string]string
	content     string
	children    []any
}

// NewXMLElementBuilder creates a new XML element builder
func NewXMLElementBuilder(elementType string) *XMLElementBuilder {
	return &XMLElementBuilder{
		elementType: elementType,
		attributes:  make(map[string]string),
		children:    make([]any, 0),
	}
}

// SetAttribute adds an attribute to the element
func (builder *XMLElementBuilder) SetAttribute(name, value string) *XMLElementBuilder {
	builder.attributes[name] = value
	return builder
}

// SetContent sets the text content of the element
func (builder *XMLElementBuilder) SetContent(content string) *XMLElementBuilder {
	builder.content = content
	return builder
}

// AddChild adds a child element
func (builder *XMLElementBuilder) AddChild(child any) *XMLElementBuilder {
	builder.children = append(builder.children, child)
	return builder
}

// BuildExpand creates an ExpandElement
func (builder *XMLElementBuilder) BuildExpand() ExpandElement {
	expand := ExpandElement{
		XMLName: xml.Name{Local: "expand"},
		Content: builder.content,
	}

	if title, exists := builder.attributes["title"]; exists {
		expand.Title = title
	}
	if localID, exists := builder.attributes["localId"]; exists {
		expand.LocalID = localID
	}

	return expand
}

// BuildMention creates a MentionElement
func (builder *XMLElementBuilder) BuildMention() MentionElement {
	mention := MentionElement{
		XMLName: xml.Name{Local: "mention"},
		Content: builder.content,
	}

	if id, exists := builder.attributes["id"]; exists {
		mention.ID = id
	}
	if text, exists := builder.attributes["text"]; exists {
		mention.Text = text
	}
	if accessLevel, exists := builder.attributes["accessLevel"]; exists {
		mention.AccessLevel = accessLevel
	}
	if userType, exists := builder.attributes["userType"]; exists {
		mention.UserType = userType
	}

	return mention
}

// BuildTaskList creates a TaskListElement
func (builder *XMLElementBuilder) BuildTaskList() TaskListElement {
	taskList := TaskListElement{
		XMLName: xml.Name{Local: "taskList"},
	}

	if localID, exists := builder.attributes["localId"]; exists {
		taskList.LocalID = localID
	}

	// Convert children to TaskElement instances
	for _, child := range builder.children {
		if taskElement, ok := child.(TaskElement); ok {
			taskList.Children = append(taskList.Children, taskElement)
		}
	}

	return taskList
}

// BuildTaskItem creates a TaskElement
func (builder *XMLElementBuilder) BuildTaskItem() TaskElement {
	taskItem := TaskElement{
		XMLName: xml.Name{Local: "taskItem"},
		Content: builder.content,
	}

	if localID, exists := builder.attributes["localId"]; exists {
		taskItem.LocalID = localID
	}
	if state, exists := builder.attributes["state"]; exists {
		taskItem.State = state
	}

	return taskItem
}

// BuildHardBreak creates a HardBreakElement
func (builder *XMLElementBuilder) BuildHardBreak() HardBreakElement {
	hardBreak := HardBreakElement{
		XMLName: xml.Name{Local: "hardBreak"},
	}

	if localID, exists := builder.attributes["localId"]; exists {
		hardBreak.LocalID = localID
	}

	return hardBreak
}

// BuildGeneric creates a GenericADFElement
func (builder *XMLElementBuilder) BuildGeneric() GenericADFElement {
	generic := GenericADFElement{
		XMLName: xml.Name{Local: "adf-node"},
		Content: builder.content,
	}

	if elementType, exists := builder.attributes["type"]; exists {
		generic.Type = elementType
	} else {
		generic.Type = builder.elementType
	}

	if localID, exists := builder.attributes["localId"]; exists {
		generic.LocalID = localID
	}

	// Note: Additional attributes beyond type and localId are not preserved
	// in generic wrapper for simplicity

	return generic
}

// XMLElementHelper provides utility functions for working with XML elements
type XMLElementHelper struct{}

// NewXMLElementHelper creates a new XML element helper
func NewXMLElementHelper() *XMLElementHelper {
	return &XMLElementHelper{}
}

// GetElementType extracts the element type from XML data
func (helper *XMLElementHelper) GetElementType(data []byte) (string, error) {
	var temp struct {
		XMLName xml.Name
	}

	if err := xml.Unmarshal(data, &temp); err != nil {
		return "", err
	}

	return temp.XMLName.Local, nil
}

// HasAttribute checks if XML element has a specific attribute
func (helper *XMLElementHelper) HasAttribute(data []byte, attributeName string) (bool, error) {
	var temp map[string]any
	if err := xml.Unmarshal(data, &temp); err != nil {
		return false, err
	}

	_, exists := temp[attributeName]
	return exists, nil
}

// ExtractAttributes extracts all attributes from XML element
func (helper *XMLElementHelper) ExtractAttributes(data []byte) (map[string]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would need more sophisticated XML parsing

	var temp struct {
		Attributes []xml.Attr `xml:",any,attr"`
	}

	if err := xml.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	attributes := make(map[string]string)
	for _, attr := range temp.Attributes {
		attributes[attr.Name.Local] = attr.Value
	}

	return attributes, nil
}

// ValidateElementStructure performs basic structure validation
func (helper *XMLElementHelper) ValidateElementStructure(elementType string, data []byte) error {
	switch elementType {
	case "expand":
		var expand ExpandElement
		return xml.Unmarshal(data, &expand)
	case "mention":
		var mention MentionElement
		return xml.Unmarshal(data, &mention)
	case "taskList":
		var taskList TaskListElement
		return xml.Unmarshal(data, &taskList)
	case "taskItem":
		var taskItem TaskElement
		return xml.Unmarshal(data, &taskItem)
	case "hardBreak":
		var hardBreak HardBreakElement
		return xml.Unmarshal(data, &hardBreak)
	default:
		var generic GenericADFElement
		return xml.Unmarshal(data, &generic)
	}
}

// CreateXMLFromADFNode creates appropriate XML element from ADF node
func (helper *XMLElementHelper) CreateXMLFromADFNode(node adf.Node) ([]byte, error) {
	builder := NewXMLElementBuilder(node.Type)

	// Copy attributes
	for key, value := range node.Attrs {
		if valueStr, ok := value.(string); ok {
			builder.SetAttribute(key, valueStr)
		}
	}

	// Handle content based on node type
	switch node.Type {
	case "expand":
		element := builder.BuildExpand()
		return xml.Marshal(element)
	case "mention":
		element := builder.BuildMention()
		return xml.Marshal(element)
	case "taskList":
		element := builder.BuildTaskList()
		return xml.Marshal(element)
	case "taskItem":
		element := builder.BuildTaskItem()
		return xml.Marshal(element)
	case "hardBreak":
		element := builder.BuildHardBreak()
		return xml.Marshal(element)
	default:
		element := builder.BuildGeneric()
		return xml.Marshal(element)
	}
}
