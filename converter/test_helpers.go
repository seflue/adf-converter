package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/seflue/adf-converter/adf_types"
)

// Helper function to parse ADF payload from JSON string
// Used across all test files for consistency
func parseTestADFPayload(t *testing.T, payload string) adf_types.ADFDocument {
	var parsed struct {
		Fields struct {
			Description adf_types.ADFDocument `json:"description"`
		} `json:"fields"`
	}

	err := json.Unmarshal([]byte(payload), &parsed)
	require.NoError(t, err, "Failed to parse ADF payload")
	return parsed.Fields.Description
}

// Real ADF content samples extracted from Jira API debug logs
// These can be used as reference for creating focused tests

// Basic text with simple formatting
var BasicTextADF = `{
	"fields": {
		"description": {
			"content": [
				{
					"attrs": {
						"level": 1
					},
					"content": [
						{
							"text": "Test Heading",
							"type": "text"
						}
					],
					"type": "heading"
				},
				{
					"content": [
						{
							"text": "This is a simple paragraph with some ",
							"type": "text"
						},
						{
							"marks": [
								{
									"type": "strong"
								}
							],
							"text": "bold text",
							"type": "text"
						},
						{
							"text": " and some ",
							"type": "text"
						},
						{
							"marks": [
								{
									"type": "em"
								}
							],
							"text": "italic text",
							"type": "text"
						},
						{
							"text": ".",
							"type": "text"
						}
					],
					"type": "paragraph"
				}
			],
			"type": "doc",
			"version": 1
		}
	}
}`

// Simple list structure
var BasicListADF = `{
	"fields": {
		"description": {
			"content": [
				{
					"content": [
						{
							"content": [
								{
									"content": [
										{
											"text": "First item",
											"type": "text"
										}
									],
									"type": "paragraph"
								}
							],
							"type": "listItem"
						},
						{
							"content": [
								{
									"content": [
										{
											"text": "Second item",
											"type": "text"
										}
									],
									"type": "paragraph"
								}
							],
							"type": "listItem"
						}
					],
					"type": "bulletList"
				}
			],
			"type": "doc",
			"version": 1
		}
	}
}`
