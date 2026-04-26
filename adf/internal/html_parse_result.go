package internal

// HTMLLinkParseResult contains the result of parsing an HTML-wrapped link
type HTMLLinkParseResult struct {
	Success     bool
	LinkText    string
	Href        string
	Attributes  map[string]string
	TotalLength int
}

// NewHTMLLinkParseResult creates a new HTMLLinkParseResult
func NewHTMLLinkParseResult() HTMLLinkParseResult {
	return HTMLLinkParseResult{
		Success:     false,
		LinkText:    "",
		Href:        "",
		Attributes:  make(map[string]string),
		TotalLength: 0,
	}
}

// NewSuccessfulHTMLLinkParseResult creates a successful parse result
func NewSuccessfulHTMLLinkParseResult(linkText, href string, attributes map[string]string, totalLength int) HTMLLinkParseResult {
	if attributes == nil {
		attributes = make(map[string]string)
	}

	return HTMLLinkParseResult{
		Success:     true,
		LinkText:    linkText,
		Href:        href,
		Attributes:  attributes,
		TotalLength: totalLength,
	}
}

// IsValid performs basic validation on the parse result
func (hpr *HTMLLinkParseResult) IsValid() bool {
	if !hpr.Success {
		return false
	}

	// Successful results should have non-empty link text and href
	if hpr.LinkText == "" || hpr.Href == "" {
		return false
	}

	// Total length should be positive
	if hpr.TotalLength <= 0 {
		return false
	}

	return true
}

// GetAttribute retrieves an attribute value
func (hpr *HTMLLinkParseResult) GetAttribute(key string) (string, bool) {
	if hpr.Attributes == nil {
		return "", false
	}
	value, exists := hpr.Attributes[key]
	return value, exists
}

// HasAttribute checks if an attribute exists
func (hpr *HTMLLinkParseResult) HasAttribute(key string) bool {
	_, exists := hpr.GetAttribute(key)
	return exists
}

// AttributeCount returns the number of attributes
func (hpr *HTMLLinkParseResult) AttributeCount() int {
	if hpr.Attributes == nil {
		return 0
	}
	return len(hpr.Attributes)
}

// AttributeKeys returns all attribute keys
func (hpr *HTMLLinkParseResult) AttributeKeys() []string {
	if hpr.Attributes == nil {
		return []string{}
	}

	keys := make([]string, 0, len(hpr.Attributes))
	for key := range hpr.Attributes {
		keys = append(keys, key)
	}
	return keys
}
