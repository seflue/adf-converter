package adf


// ConvertRoundTrip performs a full round-trip conversion for testing
// ADF → Markdown → ADF and returns both the intermediate markdown and final ADF.
// This is a free function over the Converter interface.
func ConvertRoundTrip(c Converter, original Document) (markdown string, restored Document, err error) {
	markdown, session, err := c.ToMarkdown(original)
	if err != nil {
		return "", Document{}, err
	}

	restored, _, err = c.FromMarkdown(markdown, session)
	if err != nil {
		return markdown, Document{}, err
	}

	return markdown, restored, nil
}
