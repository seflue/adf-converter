package converter

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"

	"adf-converter/converter/internal"
)

// ========== XML ELEMENT EDGE CASES ==========

// TestExpandToXML_AttributeHandling tests XML element builder with various attributes
func TestExpandToXML_AttributeHandling(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		localID  interface{}
		expected string
	}{
		{
			name:     "with localId",
			title:    "Test Title",
			localID:  "local123",
			expected: `<expand title="Test Title" localId="local123">`,
		},
		{
			name:     "without localId",
			title:    "Test Title",
			localID:  nil,
			expected: `<expand title="Test Title"`,
		},
		{
			name:     "empty title",
			title:    "",
			localID:  "local123",
			expected: `<expand title="" localId="local123"`,
		},
		{
			name:     "special chars in title",
			title:    "Title with \"quotes\"",
			localID:  nil,
			expected: `<expand title="Title with &#34;quotes&#34;"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := internal.NewXMLElementBuilder("expand")
			builder.SetAttribute("title", tt.title)
			if tt.localID != nil {
				if localIDStr, ok := tt.localID.(string); ok {
					builder.SetAttribute("localId", localIDStr)
				}
			}

			element := builder.BuildExpand()

			// Test XML marshaling to check attributes
			xmlBytes, err := xml.Marshal(element)
			assert.NoError(t, err)
			xmlString := string(xmlBytes)
			assert.Contains(t, xmlString, tt.expected)
		})
	}
}

// TestTaskElement_StateValidation tests task element state validation
func TestTaskElement_StateValidation(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		content  string
		expected string
	}{
		{
			name:     "TODO state",
			state:    "TODO",
			content:  "Task content",
			expected: `state="TODO"`,
		},
		{
			name:     "DONE state",
			state:    "DONE",
			content:  "Task content",
			expected: `state="DONE"`,
		},
		{
			name:     "invalid state",
			state:    "INVALID",
			content:  "Task content",
			expected: `state="INVALID"`, // Should still work
		},
		{
			name:     "empty state",
			state:    "",
			content:  "Task content",
			expected: `state=""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := TaskElement{
				State:   tt.state,
				Content: tt.content,
			}

			// Test XML marshaling
			xmlBytes, err := xml.Marshal(element)
			assert.NoError(t, err)
			assert.Contains(t, string(xmlBytes), tt.expected)
		})
	}
}

// TestXMLElementBuilder_EdgeCases tests the XMLElementBuilder with edge cases
func TestXMLElementBuilder_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "empty builder",
			testFunc: func(t *testing.T) {
				builder := internal.NewXMLElementBuilder("test")
				result := builder.BuildGeneric()
				assert.NotEqual(t, internal.GenericADFElement{}, result)
			},
		},
		{
			name: "builder with content only",
			testFunc: func(t *testing.T) {
				builder := internal.NewXMLElementBuilder("test")
				builder.SetContent("Test content")
				result := builder.BuildGeneric()
				assert.Equal(t, "Test content", result.Content)
			},
		},
		{
			name: "builder with attributes only",
			testFunc: func(t *testing.T) {
				builder := internal.NewXMLElementBuilder("test")
				builder.SetAttribute("type", "test-type")
				result := builder.BuildGeneric()
				assert.Equal(t, "test-type", result.Type)
			},
		},
		{
			name: "builder with special character content",
			testFunc: func(t *testing.T) {
				builder := internal.NewXMLElementBuilder("test")
				builder.SetContent("value & < > \"")
				result := builder.BuildGeneric()
				// Should preserve content as-is (XML escaping happens during marshaling)
				assert.Equal(t, "value & < > \"", result.Content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// TestMentionElement_Attributes tests mention element attribute handling
func TestMentionElement_Attributes(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		text        string
		accessLevel string
		expectValid bool
	}{
		{
			name:        "valid mention",
			id:          "user123",
			text:        "@john.doe",
			accessLevel: "CONTAINER",
			expectValid: true,
		},
		{
			name:        "empty ID",
			id:          "",
			text:        "@unknown",
			accessLevel: "CONTAINER",
			expectValid: true, // Should still work
		},
		{
			name:        "special characters in text",
			id:          "user123",
			text:        "@user.with-special_chars",
			accessLevel: "CONTAINER",
			expectValid: true,
		},
		{
			name:        "empty access level",
			id:          "user123",
			text:        "@john.doe",
			accessLevel: "",
			expectValid: true, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := MentionElement{
				ID:          tt.id,
				Text:        tt.text,
				AccessLevel: tt.accessLevel,
			}

			// Test XML marshaling
			xmlBytes, err := xml.Marshal(element)
			if tt.expectValid {
				assert.NoError(t, err)
				assert.NotEmpty(t, xmlBytes)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestHardBreakElement_XMLGeneration tests hard break element XML generation
func TestHardBreakElement_XMLGeneration(t *testing.T) {
	element := HardBreakElement{}

	// Test XML marshaling
	xmlBytes, err := xml.Marshal(element)
	assert.NoError(t, err)

	// Should generate hardBreak tag
	xmlString := string(xmlBytes)
	assert.Contains(t, xmlString, "hardBreak")
}

// TestXMLElementHelper_EdgeCases tests XMLElementHelper with various edge cases
func TestXMLElementHelper_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "get type from empty XML data",
			testFunc: func(t *testing.T) {
				helper := internal.NewXMLElementHelper()
				elementType, err := helper.GetElementType([]byte{})
				// Should handle empty data gracefully
				assert.Error(t, err) // Expected to error on empty data
				assert.Empty(t, elementType)
			},
		},
		{
			name: "get type from invalid XML",
			testFunc: func(t *testing.T) {
				helper := internal.NewXMLElementHelper()
				elementType, err := helper.GetElementType([]byte("invalid xml"))
				// Should handle invalid XML gracefully
				assert.Error(t, err)
				assert.Empty(t, elementType)
			},
		},
		{
			name: "validate structure with invalid XML",
			testFunc: func(t *testing.T) {
				helper := internal.NewXMLElementHelper()
				err := helper.ValidateElementStructure("invalid", []byte("invalid xml"))
				assert.Error(t, err) // Should return error, not bool
			},
		},
		{
			name: "extract attributes from nil data",
			testFunc: func(t *testing.T) {
				helper := internal.NewXMLElementHelper()
				attrs, err := helper.ExtractAttributes(nil)
				assert.Error(t, err) // Should error on nil data
				assert.Nil(t, attrs) // Should return nil on error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
