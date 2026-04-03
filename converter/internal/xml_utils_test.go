package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseXMLAttributes_StringValues(t *testing.T) {
	tests := []struct {
		name     string
		xmlTag   string
		expected map[string]interface{}
	}{
		{
			name:   "single string attribute with double quotes",
			xmlTag: `<table layout="default">`,
			expected: map[string]interface{}{
				"layout": "default",
			},
		},
		{
			name:   "single string attribute with single quotes",
			xmlTag: `<table layout='default'>`,
			expected: map[string]interface{}{
				"layout": "default",
			},
		},
		{
			name:   "multiple string attributes",
			xmlTag: `<element id="test" class="example" data-value="something">`,
			expected: map[string]interface{}{
				"id":         "test",
				"class":      "example",
				"data-value": "something",
			},
		},
		{
			name:   "empty string value",
			xmlTag: `<element value="">`,
			expected: map[string]interface{}{
				"value": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseXMLAttributes(tt.xmlTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseXMLAttributes_BooleanValues(t *testing.T) {
	tests := []struct {
		name     string
		xmlTag   string
		expected map[string]interface{}
	}{
		{
			name:   "true boolean",
			xmlTag: `<element checked="true">`,
			expected: map[string]interface{}{
				"checked": true,
			},
		},
		{
			name:   "false boolean",
			xmlTag: `<element checked="false">`,
			expected: map[string]interface{}{
				"checked": false,
			},
		},
		{
			name:   "multiple boolean attributes",
			xmlTag: `<element checked="true" disabled="false" readonly="true">`,
			expected: map[string]interface{}{
				"checked":  true,
				"disabled": false,
				"readonly": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseXMLAttributes(tt.xmlTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseXMLAttributes_NumericValues(t *testing.T) {
	tests := []struct {
		name     string
		xmlTag   string
		expected map[string]interface{}
	}{
		{
			name:   "positive integer",
			xmlTag: `<element width="100">`,
			expected: map[string]interface{}{
				"width": 100,
			},
		},
		{
			name:   "negative integer",
			xmlTag: `<element offset="-5">`,
			expected: map[string]interface{}{
				"offset": -5,
			},
		},
		{
			name:   "zero",
			xmlTag: `<element count="0">`,
			expected: map[string]interface{}{
				"count": 0,
			},
		},
		{
			name:   "multiple numeric attributes",
			xmlTag: `<element width="100" height="200" depth="50">`,
			expected: map[string]interface{}{
				"width":  100,
				"height": 200,
				"depth":  50,
			},
		},
		{
			name:   "decimal stays as string",
			xmlTag: `<element value="3.14">`,
			expected: map[string]interface{}{
				"value": "3.14",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseXMLAttributes(tt.xmlTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseXMLAttributes_MixedQuotes(t *testing.T) {
	tests := []struct {
		name     string
		xmlTag   string
		expected map[string]interface{}
	}{
		{
			name:   "mixed double and single quotes",
			xmlTag: `<element id="test" class='example' count="42">`,
			expected: map[string]interface{}{
				"id":    "test",
				"class": "example",
				"count": 42,
			},
		},
		{
			name:   "mixed types with different quotes",
			xmlTag: `<element checked='true' width="100" label="test">`,
			expected: map[string]interface{}{
				"checked": true,
				"width":   100,
				"label":   "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseXMLAttributes(tt.xmlTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseXMLAttributes_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		xmlTag   string
		expected map[string]interface{}
	}{
		{
			name:     "no attributes",
			xmlTag:   `<element>`,
			expected: map[string]interface{}{},
		},
		{
			name:     "empty tag",
			xmlTag:   ``,
			expected: map[string]interface{}{},
		},
		{
			name:     "self-closing tag",
			xmlTag:   `<element id="test" />`,
			expected: map[string]interface{}{"id": "test"},
		},
		{
			name:   "attributes with special characters in values",
			xmlTag: `<element data-info="value-with-dashes" class="test_underscore">`,
			expected: map[string]interface{}{
				"data-info": "value-with-dashes",
				"class":     "test_underscore",
			},
		},
		{
			name:   "mixed types all together",
			xmlTag: `<element id="test" width="100" checked="true" height="0" disabled="false" label="">`,
			expected: map[string]interface{}{
				"id":       "test",
				"width":    100,
				"checked":  true,
				"height":   0,
				"disabled": false,
				"label":    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseXMLAttributes(tt.xmlTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}
