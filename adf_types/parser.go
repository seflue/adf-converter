package adf_types

import (
	"encoding/json"
	"fmt"
)

// Parser handles conversion between ADF JSON and Go structures
type Parser interface {
	Parse(jsonData []byte) (*ADFDocument, error)
	Serialize(doc *ADFDocument) ([]byte, error)
}

// DefaultParser implements the Parser interface using standard JSON encoding
type DefaultParser struct{}

// NewParser creates a new ADF parser
func NewParser() Parser {
	return &DefaultParser{}
}

// Parse converts ADF JSON bytes into an ADFDocument structure
func (p *DefaultParser) Parse(jsonData []byte) (*ADFDocument, error) {
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	var doc ADFDocument
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse ADF JSON: %w", err)
	}

	// Basic validation
	if doc.Type != NodeTypeDoc {
		return nil, fmt.Errorf("invalid document type: expected '%s', got '%s'", NodeTypeDoc, doc.Type)
	}

	if doc.Version == 0 {
		doc.Version = 1 // Set default version if missing
	}

	return &doc, nil
}

// Serialize converts an ADFDocument into JSON bytes
func (p *DefaultParser) Serialize(doc *ADFDocument) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("document is nil")
	}

	// Ensure required fields are set
	if doc.Type == "" {
		doc.Type = NodeTypeDoc
	}
	if doc.Version == 0 {
		doc.Version = 1
	}

	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ADF document: %w", err)
	}

	return jsonData, nil
}

// ParseFromString is a convenience method for parsing ADF from string
func ParseFromString(jsonStr string) (*ADFDocument, error) {
	parser := NewParser()
	return parser.Parse([]byte(jsonStr))
}

// SerializeToString is a convenience method for serializing ADF to string
func SerializeToString(doc *ADFDocument) (string, error) {
	parser := NewParser()
	jsonData, err := parser.Serialize(doc)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
