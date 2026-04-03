package converter

import (
	"fmt"
	"sort"
	"strings"
)

// IsWebLink returns true if the given href is a web link (https:// or http://)
func IsWebLink(href string) bool {
	return strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "http://")
}

// IsInternalLink returns true if the given href is an internal link (relative path)
func IsInternalLink(href string) bool {
	return strings.HasPrefix(href, "/")
}

// HasMetadata returns true if the link attributes contain metadata beyond just href
func HasMetadata(attrs map[string]interface{}) bool {
	// If there's only href or no attributes, no metadata
	if len(attrs) <= 1 {
		return false
	}

	// If there are attributes other than href, there's metadata
	for key := range attrs {
		if key != "href" {
			return true
		}
	}

	return false
}

// ClassifyLink classifies a link based on its attributes
func ClassifyLink(attrs map[string]interface{}) string {
	href, ok := attrs["href"].(string)
	if !ok {
		return "Unknown"
	}

	if IsWebLink(href) {
		return "WebLink"
	}

	if IsInternalLink(href) {
		if HasMetadata(attrs) {
			return "ComplexInternalLink"
		}
		return "SimpleInternalLink"
	}

	return "Unknown"
}

// DetermineConversionStrategy determines the conversion strategy based on link attributes
func DetermineConversionStrategy(attrs map[string]interface{}) string {
	linkType := ClassifyLink(attrs)

	switch linkType {
	case "WebLink", "SimpleInternalLink":
		return "StandardMarkdown"
	case "ComplexInternalLink":
		return "HTMLWrapped"
	default:
		return "Placeholder"
	}
}

// GenerateHTMLWrapper generates an HTML wrapper for link attributes and text
func GenerateHTMLWrapper(attrs map[string]interface{}, text string) string {
	var parts []string

	// Always include href first if present
	if href, ok := attrs["href"].(string); ok {
		parts = append(parts, fmt.Sprintf(`href="%s"`, href))
	}

	// Add other attributes in sorted order for consistency
	var otherKeys []string
	for key := range attrs {
		if key != "href" {
			otherKeys = append(otherKeys, key)
		}
	}
	sort.Strings(otherKeys)

	for _, key := range otherKeys {
		if value, ok := attrs[key].(string); ok {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, key, value))
		}
	}

	attributesStr := strings.Join(parts, " ")
	return fmt.Sprintf(`<a %s>%s</a>`, attributesStr, text)
}
