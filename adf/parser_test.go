package adf

import (
	"strings"
	"testing"
)

func TestDefaultParser_Parse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		validate    func(*testing.T, *Document)
	}{
		{
			name: "valid simple document",
			jsonData: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "Hello, world!"
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, doc *Document) {
				if doc.Version != 1 {
					t.Errorf("Expected version 1, got %d", doc.Version)
				}
				if doc.Type != "doc" {
					t.Errorf("Expected type 'doc', got %s", doc.Type)
				}
				if len(doc.Content) != 1 {
					t.Errorf("Expected 1 content node, got %d", len(doc.Content))
				}
				if doc.Content[0].Type != "paragraph" {
					t.Errorf("Expected paragraph node, got %s", doc.Content[0].Type)
				}
			},
		},
		{
			name: "document with no version (should default to 1)",
			jsonData: `{
				"type": "doc",
				"content": []
			}`,
			expectError: false,
			validate: func(t *testing.T, doc *Document) {
				if doc.Version != 1 {
					t.Errorf("Expected default version 1, got %d", doc.Version)
				}
			},
		},
		{
			name: "document with heading",
			jsonData: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "heading",
						"attrs": {
							"level": 2
						},
						"content": [
							{
								"type": "text",
								"text": "Heading Text"
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Content) != 1 {
					t.Errorf("Expected 1 content node, got %d", len(doc.Content))
				}
				heading := doc.Content[0]
				if heading.Type != "heading" {
					t.Errorf("Expected heading node, got %s", heading.Type)
				}
				level := heading.GetHeadingLevel()
				if level != 2 {
					t.Errorf("Expected heading level 2, got %d", level)
				}
			},
		},
		{
			name: "document with text marks",
			jsonData: `{
				"version": 1,
				"type": "doc",
				"content": [
					{
						"type": "paragraph",
						"content": [
							{
								"type": "text",
								"text": "Bold text",
								"marks": [
									{
										"type": "strong"
									}
								]
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, doc *Document) {
				textNode := doc.Content[0].Content[0]
				if len(textNode.Marks) != 1 {
					t.Errorf("Expected 1 mark, got %d", len(textNode.Marks))
				}
				if textNode.Marks[0].Type != "strong" {
					t.Errorf("Expected strong mark, got %s", textNode.Marks[0].Type)
				}
			},
		},
		{
			name:        "empty JSON",
			jsonData:    "",
			expectError: true,
		},
		{
			name:        "invalid JSON",
			jsonData:    `{"invalid": json}`,
			expectError: true,
		},
		{
			name: "invalid document type",
			jsonData: `{
				"version": 1,
				"type": "invalid",
				"content": []
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parser.Parse([]byte(tt.jsonData))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if doc == nil {
				t.Error("Expected document but got nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, doc)
			}
		})
	}
}

func TestDefaultParser_Serialize(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		doc         *Document
		expectError bool
		validate    func(*testing.T, []byte)
	}{
		{
			name: "simple document",
			doc: &Document{
				Version: 1,
				Type:    "doc",
				Content: []Node{
					{
						Type: "paragraph",
						Content: []Node{
							{
								Type: "text",
								Text: "Hello",
							},
						},
					},
				},
			},
			expectError: false,
			validate: func(t *testing.T, jsonData []byte) {
				jsonStr := string(jsonData)
				if !strings.Contains(jsonStr, `"type": "doc"`) {
					t.Error("Serialized JSON should contain doc type")
				}
				if !strings.Contains(jsonStr, `"text": "Hello"`) {
					t.Error("Serialized JSON should contain text content")
				}
			},
		},
		{
			name: "document with missing type (should default)",
			doc: &Document{
				Version: 1,
				Content: []Node{},
			},
			expectError: false,
			validate: func(t *testing.T, jsonData []byte) {
				jsonStr := string(jsonData)
				if !strings.Contains(jsonStr, `"type": "doc"`) {
					t.Error("Serialized JSON should contain default doc type")
				}
			},
		},
		{
			name: "document with missing version (should default)",
			doc: &Document{
				Type:    "doc",
				Content: []Node{},
			},
			expectError: false,
			validate: func(t *testing.T, jsonData []byte) {
				jsonStr := string(jsonData)
				if !strings.Contains(jsonStr, `"version": 1`) {
					t.Error("Serialized JSON should contain default version")
				}
			},
		},
		{
			name:        "nil document",
			doc:         nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := parser.Serialize(tt.doc)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if jsonData == nil {
				t.Error("Expected JSON data but got nil")
				return
			}

			if tt.validate != nil {
				tt.validate(t, jsonData)
			}
		})
	}
}

func TestParseFromString(t *testing.T) {
	jsonStr := `{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{
						"type": "text",
						"text": "Test"
					}
				]
			}
		]
	}`

	doc, err := ParseFromString(jsonStr)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document but got nil")
	}

	if doc.Version != 1 {
		t.Errorf("Expected version 1, got %d", doc.Version)
	}
}

func TestSerializeToString(t *testing.T) {
	doc := &Document{
		Version: 1,
		Type:    "doc",
		Content: []Node{
			{
				Type: "paragraph",
				Content: []Node{
					{
						Type: "text",
						Text: "Test",
					},
				},
			},
		},
	}

	jsonStr, err := SerializeToString(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected JSON string but got empty")
	}

	if !strings.Contains(jsonStr, `"text": "Test"`) {
		t.Error("Serialized JSON should contain text content")
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that parsing and serializing preserves the document structure
	originalJSON := `{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{
						"type": "text",
						"text": "Hello, ",
						"marks": [
							{
								"type": "strong"
							}
						]
					},
					{
						"type": "text",
						"text": "world!"
					}
				]
			},
			{
				"type": "heading",
				"attrs": {
					"level": 2
				},
				"content": [
					{
						"type": "text",
						"text": "Heading"
					}
				]
			}
		]
	}`

	parser := NewParser()

	// Parse the original JSON
	doc, err := parser.Parse([]byte(originalJSON))
	if err != nil {
		t.Fatalf("Failed to parse original JSON: %v", err)
	}

	// Serialize it back to JSON
	serializedJSON, err := parser.Serialize(doc)
	if err != nil {
		t.Fatalf("Failed to serialize document: %v", err)
	}

	// Parse the serialized JSON again
	doc2, err := parser.Parse(serializedJSON)
	if err != nil {
		t.Fatalf("Failed to parse serialized JSON: %v", err)
	}

	// Verify the documents are equivalent
	if doc2.Version != doc.Version {
		t.Errorf("Version mismatch: original %d, round-trip %d", doc.Version, doc2.Version)
	}

	if doc2.Type != doc.Type {
		t.Errorf("Type mismatch: original %s, round-trip %s", doc.Type, doc2.Type)
	}

	if len(doc2.Content) != len(doc.Content) {
		t.Errorf("Content length mismatch: original %d, round-trip %d", len(doc.Content), len(doc2.Content))
	}

	// Verify specific content
	if len(doc2.Content) >= 2 {
		// Check paragraph
		para := doc2.Content[0]
		if para.Type != "paragraph" || len(para.Content) != 2 {
			t.Error("Paragraph structure not preserved")
		}

		// Check heading
		heading := doc2.Content[1]
		if heading.Type != "heading" || heading.GetHeadingLevel() != 2 {
			t.Error("Heading structure not preserved")
		}
	}
}
