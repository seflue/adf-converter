package adf_test

import (
	"encoding/json"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealJiraADF tests conversion using actual ADF structure retrieved from Jira
// This includes the nestedExpand node type that Jira actually uses
func TestRealJiraADF(t *testing.T) {
	// This is actual ADF content retrieved from a Jira ticket after fixing nested details
	realJiraADF := `{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "heading",
      "attrs": {
        "level": 1
      },
      "content": [
        {
          "type": "text",
          "text": "Test 1"
        }
      ]
    },
    {
      "type": "expand",
      "attrs": {
        "title": "This is for testing"
      },
      "content": [
        {
          "type": "heading",
          "attrs": {
            "level": 1
          },
          "content": [
            {
              "type": "text",
              "text": "Heading"
            }
          ]
        },
        {
          "type": "bulletList",
          "content": [
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    {
                      "type": "text",
                      "text": "item 1"
                    }
                  ]
                }
              ]
            },
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    {
                      "type": "text",
                      "text": "item 2"
                    }
                  ]
                }
              ]
            },
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    {
                      "type": "text",
                      "marks": [
                        {
                          "type": "strong"
                        }
                      ],
                      "text": "bold"
                    },
                    {
                      "type": "text",
                      "text": " item"
                    }
                  ]
                }
              ]
            },
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    {
                      "type": "text",
                      "marks": [
                        {
                          "type": "em"
                        }
                      ],
                      "text": "italic"
                    },
                    {
                      "type": "text",
                      "text": " item"
                    }
                  ]
                }
              ]
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
              "text": "Subheading"
            }
          ]
        },
        {
          "type": "paragraph",
          "content": [
            {
              "type": "text",
              "text": "More content"
            }
          ]
        }
      ]
    },
    {
      "type": "heading",
      "attrs": {
        "level": 1
      },
      "content": [
        {
          "type": "text",
          "text": "Test 2"
        }
      ]
    },
    {
      "type": "expand",
      "attrs": {
        "title": "Further testing"
      },
      "content": [
        {
          "type": "orderedList",
          "content": [
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    {
                      "type": "text",
                      "text": "Some content"
                    }
                  ]
                }
              ]
            },
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    {
                      "type": "text",
                      "text": "More content"
                    }
                  ]
                }
              ]
            }
          ]
        },
        {
          "type": "nestedExpand",
          "attrs": {
            "title": "Nested"
          },
          "content": [
            {
              "type": "bulletList",
              "content": [
                {
                  "type": "listItem",
                  "content": [
                    {
                      "type": "paragraph",
                      "content": [
                        {
                          "type": "text",
                          "text": "more"
                        }
                      ]
                    }
                  ]
                },
                {
                  "type": "listItem",
                  "content": [
                    {
                      "type": "paragraph",
                      "content": [
                        {
                          "type": "text",
                          "text": "content"
                        }
                      ]
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    },
    {
      "type": "paragraph",
      "content": [
        {
          "type": "text",
          "text": "Test Instructions"
        }
      ]
    },
    {
      "type": "bulletList",
      "content": [
        {
          "type": "listItem",
          "content": [
            {
              "type": "paragraph",
              "content": [
                {
                  "type": "text",
                  "text": "Simple text item"
                }
              ]
            }
          ]
        },
        {
          "type": "listItem",
          "content": [
            {
              "type": "paragraph",
              "content": [
                {
                  "type": "text",
                  "marks": [
                    {
                      "type": "strong"
                    }
                  ],
                  "text": "Bold text item"
                }
              ]
            }
          ]
        },
        {
          "type": "listItem",
          "content": [
            {
              "type": "paragraph",
              "content": [
                {
                  "type": "text",
                  "text": "Item with "
                },
                {
                  "type": "text",
                  "marks": [
                    {
                      "type": "link",
                      "attrs": {
                        "href": "https://example.com"
                      }
                    }
                  ],
                  "text": "link"
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}`

	conv := defaults.NewDefaultConverter()
	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Parse the real Jira ADF
	var adfDoc adf.Document
	err := json.Unmarshal([]byte(realJiraADF), &adfDoc)
	require.NoError(t, err, "Should parse real Jira ADF successfully")

	t.Run("Real Jira ADF Structure Validation", func(t *testing.T) {
		// Validate that we can handle the real structure
		assert.Equal(t, 1, adfDoc.Version, "Should have correct ADF version")
		assert.Equal(t, "doc", adfDoc.Type, "Should have correct document type")
		assert.NotEmpty(t, adfDoc.Content, "Should have content")

		// Find the expand element with nested content
		var foundExpandWithNested bool
		for _, node := range adfDoc.Content {
			if node.Type == adf.NodeTypeExpand {
				if title, ok := node.Attrs["title"].(string); ok && title == "Further testing" {
					// Check for nestedExpand within this expand
					for _, child := range node.Content {
						if child.Type == "nestedExpand" {
							foundExpandWithNested = true
							assert.Equal(t, "Nested", child.Attrs["title"], "nestedExpand should have correct title")
							break
						}
					}
				}
			}
		}
		assert.True(t, foundExpandWithNested, "Should find expand element with nestedExpand child")
	})

	t.Run("Convert Real Jira ADF to Markdown", func(t *testing.T) {
		markdown, _, err := conv.ToMarkdown(adfDoc)
		require.NoError(t, err)

		assert.NotEmpty(t, markdown)
		assert.Contains(t, markdown, "Test 1", "Should contain heading content")
		assert.Contains(t, markdown, "This is for testing", "Should contain expand title")
		assert.Contains(t, markdown, "Further testing", "Should contain second expand title")
		assert.Contains(t, markdown, "**bold** item", "Should preserve bold formatting")
		assert.Contains(t, markdown, "*italic* item", "Should preserve italic formatting")
	})

	t.Run("Expected Markdown Structure for Real Jira ADF", func(t *testing.T) {
		markdown, _, err := conv.ToMarkdown(adfDoc)
		require.NoError(t, err)

		expectedElements := []string{
			"# Test 1",
			"<details>",
			"<summary>This is for testing</summary>",
			"# Heading",
			"- item 1",
			"- item 2",
			"- **bold** item",
			"- *italic* item",
			"## Subheading",
			"More content",
			"</details>",
			"# Test 2",
			"<details>",
			"<summary>Further testing</summary>",
			"1. Some content",
			"2. More content",
			"<details>",
			"<summary>Nested</summary>",
			"- more",
			"- content",
			"</details>",
			"</details>",
			"Test Instructions",
			"- Simple text item",
			"- **Bold text item**",
			"- Item with [link](https://example.com)",
		}

		for _, element := range expectedElements {
			assert.Contains(t, markdown, element)
		}
	})

	t.Run("Round-trip Test with Real Jira ADF", func(t *testing.T) {
		markdown, _, err := conv.ToMarkdown(adfDoc)
		require.NoError(t, err)

		newADF, _, err := conv.FromMarkdown(markdown, session)
		require.NoError(t, err)

		assert.Equal(t, adfDoc.Version, newADF.Version, "Version should be preserved")
		assert.Equal(t, adfDoc.Type, newADF.Type, "Document type should be preserved")

		// Count expand elements in both documents
		originalExpandCount := countAllExpandElements(adfDoc)
		newExpandCount := countAllExpandElements(newADF)

		t.Logf("Original expand count: %d, New expand count: %d", originalExpandCount, newExpandCount)

		// The exact structure might differ (nestedExpand vs expand), but content should be preserved
		assert.NotZero(t, newExpandCount, "Should have expand elements after round-trip")
	})
}

// Helper function to count expand-like elements (both expand and nestedExpand)
func countAllExpandElements(doc adf.Document) int {
	count := 0
	for _, node := range doc.Content {
		count += countExpandInNodeIncludingNested(node)
	}
	return count
}

func countExpandInNodeIncludingNested(node adf.Node) int {
	count := 0

	// Check if this node is an expand or nestedExpand type
	if node.Type == adf.NodeTypeExpand || node.Type == adf.NodeTypeNestedExpand {
		count++
	}

	// Recursively check content array
	for _, child := range node.Content {
		count += countExpandInNodeIncludingNested(child)
	}

	return count
}
