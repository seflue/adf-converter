package converter

import (
	"testing"

	"adf-converter/adf_types"
)

func TestJiraValidation_InlineCardWithData(t *testing.T) {
	doc := adf_types.ADFDocument{
		Type:    "doc",
		Version: 1,
		Content: []adf_types.ADFNode{
			{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeInlineCard,
						Attrs: map[string]interface{}{
							"data": map[string]interface{}{
								"@type": "Document",
							},
						},
					},
				},
			},
		},
	}
	err := ValidateJiraADFCompliance(doc)
	if err != nil {
		t.Errorf("data-only inlineCard should pass Jira validation, got: %v", err)
	}
}
