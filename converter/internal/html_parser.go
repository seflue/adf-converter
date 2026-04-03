package internal

import (
	"regexp"
	"strings"
)

// DefaultHTMLLinkParser is the default implementation of HTMLLinkParser
type DefaultHTMLLinkParser struct{}

// NewDefaultHTMLLinkParser creates a new DefaultHTMLLinkParser
func NewDefaultHTMLLinkParser() *DefaultHTMLLinkParser {
	return &DefaultHTMLLinkParser{}
}

// ParseHTMLLink parses HTML tags containing markdown links
// Input: "<a meta-attr="value">[text](url)</a>"
// Output: HTMLLinkParseResult with extracted data
func (dhlp *DefaultHTMLLinkParser) ParseHTMLLink(text string) HTMLLinkParseResult {
	if text == "" {
		return NewHTMLLinkParseResult()
	}

	// Check if it starts with an HTML link
	if !dhlp.IsHTMLLinkStart(text) {
		return NewHTMLLinkParseResult()
	}

	// Find the opening tag
	openTagEnd := strings.Index(text, ">")
	if openTagEnd == -1 {
		return NewHTMLLinkParseResult()
	}

	openTag := text[:openTagEnd+1]
	remainder := text[openTagEnd+1:]

	// Find the closing tag
	closeTagStart := strings.Index(remainder, "</a>")
	if closeTagStart == -1 {
		return NewHTMLLinkParseResult()
	}

	content := remainder[:closeTagStart]
	totalLength := openTagEnd + 1 + closeTagStart + 4 // Include </a>

	// Extract attributes from opening tag
	attributes, attrParseOk := dhlp.extractAttributesWithValidation(openTag)
	if !attrParseOk {
		return NewHTMLLinkParseResult()
	}

	// Parse markdown link from content
	linkText, href := dhlp.parseMarkdownLinkFromContent(content)
	if linkText == "" || href == "" {
		return NewHTMLLinkParseResult()
	}

	return NewSuccessfulHTMLLinkParseResult(linkText, href, attributes, totalLength)
}

// IsHTMLLinkStart checks if text begins with HTML link tag
// Returns true for strings starting with "<a "
func (dhlp *DefaultHTMLLinkParser) IsHTMLLinkStart(text string) bool {
	if len(text) < 3 {
		return false
	}

	return strings.HasPrefix(text, "<a ") || strings.HasPrefix(text, "<a>")
}

// extractAttributesWithValidation parses HTML attributes and returns success status
func (dhlp *DefaultHTMLLinkParser) extractAttributesWithValidation(htmlTag string) (map[string]string, bool) {
	attributes := make(map[string]string)

	// Remove < and > brackets
	content := strings.TrimSpace(htmlTag)
	if strings.HasPrefix(content, "<") {
		content = content[1:]
	}
	if strings.HasSuffix(content, ">") {
		content = content[:len(content)-1]
	}

	// Remove the 'a' tag name
	if strings.HasPrefix(content, "a ") {
		content = content[2:]
	} else if content == "a" {
		return attributes, true
	}

	// Parse attributes using regex - strict about quote matching
	// Check for unclosed quotes first
	if strings.Contains(content, `="`) && !regexp.MustCompile(`="[^"]*"`).MatchString(content) {
		// Has opening quote but no properly closed quote
		return attributes, false
	}
	if strings.Contains(content, `='`) && !regexp.MustCompile(`='[^']*'`).MatchString(content) {
		// Has opening quote but no properly closed quote
		return attributes, false
	}

	attrRegex := regexp.MustCompile(`(\w+(?:-\w+)*)(?:\s*=\s*"([^"]*)"|=\s*'([^']*)'|=\s*([^\s>]+))?`)
	matches := attrRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			key := match[1]
			value := ""

			// Check which capture group has the value
			if len(match) > 2 && match[2] != "" {
				value = match[2] // Double-quoted value
			} else if len(match) > 3 && match[3] != "" {
				value = match[3] // Single-quoted value
			} else if len(match) > 4 && match[4] != "" {
				value = match[4] // Unquoted value
			}

			// Skip href attribute as it will be in the markdown link
			if key != "href" {
				attributes[key] = value
			}
		}
	}

	return attributes, true
}

// ExtractAttributes parses HTML attributes from opening tag
// Input: "<a meta-attr=\"value\" title=\"test\">"
// Output: map["meta-attr"]="value", map["title"]="test"
func (dhlp *DefaultHTMLLinkParser) ExtractAttributes(htmlTag string) map[string]string {
	attributes, _ := dhlp.extractAttributesWithValidation(htmlTag)
	return attributes
}

// parseMarkdownLinkFromContent extracts markdown link text and URL from content
func (dhlp *DefaultHTMLLinkParser) parseMarkdownLinkFromContent(content string) (string, string) {
	// Look for [text](url) pattern
	linkRegex := regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)
	matches := linkRegex.FindStringSubmatch(content)

	if len(matches) >= 3 {
		linkText := matches[1]
		href := matches[2]
		return linkText, href
	}

	return "", ""
}

// ValidateHTMLLink performs validation on HTML link syntax
func (dhlp *DefaultHTMLLinkParser) ValidateHTMLLink(htmlLink string) bool {
	result := dhlp.ParseHTMLLink(htmlLink)
	return result.Success
}

// ExtractLinkTextOnly extracts just the link text from HTML link
func (dhlp *DefaultHTMLLinkParser) ExtractLinkTextOnly(htmlLink string) string {
	result := dhlp.ParseHTMLLink(htmlLink)
	if result.Success {
		return result.LinkText
	}
	return ""
}

// ExtractHrefOnly extracts just the href from HTML link
func (dhlp *DefaultHTMLLinkParser) ExtractHrefOnly(htmlLink string) string {
	result := dhlp.ParseHTMLLink(htmlLink)
	if result.Success {
		return result.Href
	}
	return ""
}

// ParseMultipleHTMLLinks parses multiple HTML links from text
func (dhlp *DefaultHTMLLinkParser) ParseMultipleHTMLLinks(text string) []HTMLLinkParseResult {
	var results []HTMLLinkParseResult
	remaining := text
	offset := 0

	for len(remaining) > 0 {
		// Find next HTML link start
		linkStart := strings.Index(remaining, "<a ")
		if linkStart == -1 {
			break
		}

		// Try to parse from this position
		linkText := remaining[linkStart:]
		result := dhlp.ParseHTMLLink(linkText)

		if result.Success {
			results = append(results, result)
			// Move past this link
			remaining = remaining[linkStart+result.TotalLength:]
			offset += linkStart + result.TotalLength
		} else {
			// Move past this failed attempt
			remaining = remaining[linkStart+1:]
			offset += linkStart + 1
		}
	}

	return results
}
