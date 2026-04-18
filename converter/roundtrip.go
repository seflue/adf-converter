package converter

import "github.com/seflue/adf-converter/adf_types"

// ConvertRoundTrip performs a full round-trip conversion for testing
// ADF → Markdown → ADF and returns both the intermediate markdown and final ADF.
// This is a free function over the Converter interface.
func ConvertRoundTrip(c Converter, original adf_types.ADFDocument) (markdown string, restored adf_types.ADFDocument, err error) {
	markdown, session, err := c.ToMarkdown(original)
	if err != nil {
		return "", adf_types.ADFDocument{}, err
	}

	result, err := c.FromMarkdown(markdown, session)
	if err != nil {
		return markdown, adf_types.ADFDocument{}, err
	}
	restored = result.Document

	return markdown, restored, nil
}
