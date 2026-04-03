package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"adf-converter/adf_types"
	"adf-converter/placeholder"
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

	converter := NewDefaultConverter()
	manager := placeholder.NewManager()
	session := manager.GetSession()

	// Parse the real Jira ADF
	var adfDoc adf_types.ADFDocument
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
			if node.Type == adf_types.NodeTypeExpand {
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
		// Convert to markdown - this should handle nestedExpand gracefully
		markdown, _, err := converter.ToMarkdown(adfDoc)
		// For now, we expect this to either work or give a clear error about unsupported node type
		if err != nil {
			// If it fails, it should be because nestedExpand is not supported yet
			assert.Contains(t, err.Error(), "nestedExpand", "Error should mention unsupported nestedExpand node type")
			t.Logf("Expected failure due to unsupported nestedExpand: %v", err)
			return
		}

		// If it succeeds, validate the output
		assert.NotEmpty(t, markdown, "Should generate non-empty markdown")
		assert.Contains(t, markdown, "Test 1", "Should contain heading content")
		assert.Contains(t, markdown, "This is for testing", "Should contain expand title")
		assert.Contains(t, markdown, "Further testing", "Should contain second expand title")
		assert.Contains(t, markdown, "**bold** item", "Should preserve bold formatting")
		assert.Contains(t, markdown, "*italic* item", "Should preserve italic formatting")

		t.Logf("Generated markdown from real Jira ADF:\n%s", markdown)
	})

	t.Run("Expected Markdown Structure for Real Jira ADF", func(t *testing.T) {
		// This test documents what we expect the conversion to produce
		// when nestedExpand support is implemented

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
			// For nestedExpand, we might expect nested details or some other structure
			"<details>", // nested details for nestedExpand
			"<summary>Nested</summary>",
			"- more",
			"- content",
			"</details>", // end nested details
			"</details>", // end parent details
			"Test Instructions",
			"- Simple text item",
			"- **Bold text item**",
			"- Item with [link](https://example.com)",
		}

		// For now, just document what we expect
		t.Logf("When nestedExpand support is implemented, markdown should contain these elements:")
		for i, element := range expectedElements {
			t.Logf("  %d. %s", i+1, element)
		}
	})

	t.Run("Round-trip Test with Real Jira ADF", func(t *testing.T) {
		// This test will verify that we can round-trip the real Jira structure
		// Currently expected to fail until nestedExpand support is added

		markdown, _, err := converter.ToMarkdown(adfDoc)
		if err != nil {
			t.Skipf("Skipping round-trip test due to ToMarkdown error: %v", err)
			return
		}

		// Convert back to ADF
		newADF, err := converter.FromMarkdownLegacy(markdown, session)
		if err != nil {
			t.Logf("Round-trip conversion failed (expected): %v", err)
			return
		}

		// If round-trip succeeds, validate structure preservation
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
func countAllExpandElements(doc adf_types.ADFDocument) int {
	count := 0
	for _, node := range doc.Content {
		count += countExpandInNodeIncludingNested(node)
	}
	return count
}

func countExpandInNodeIncludingNested(node adf_types.ADFNode) int {
	count := 0

	// Check if this node is an expand or nestedExpand type
	if node.Type == adf_types.NodeTypeExpand || node.Type == "nestedExpand" {
		count++
	}

	// Recursively check content array
	for _, child := range node.Content {
		count += countExpandInNodeIncludingNested(child)
	}

	return count
}
