package converter

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf_types"
)

func TestInlineCard_DataAttributeIsValid(t *testing.T) {
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
	violations := ValidateADFCompliance(doc)
	for _, v := range violations {
		if strings.Contains(v, "missing required url") {
			t.Errorf("data-only inlineCard should not trigger url violation, got: %s", v)
		}
	}
}
