package linkclass

import (
	"fmt"
	"sort"
	"strings"
)

func IsWebLink(href string) bool {
	return strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "http://")
}

func IsInternalLink(href string) bool {
	return strings.HasPrefix(href, "/")
}

func HasMetadata(attrs map[string]any) bool {
	if len(attrs) <= 1 {
		return false
	}
	for key := range attrs {
		if key != "href" {
			return true
		}
	}
	return false
}

func ClassifyLink(attrs map[string]any) string {
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

func DetermineConversionStrategy(attrs map[string]any) string {
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

func GenerateHTMLWrapper(attrs map[string]any, text string) string {
	var parts []string
	if href, ok := attrs["href"].(string); ok {
		parts = append(parts, fmt.Sprintf(`href="%s"`, href))
	}
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
	return fmt.Sprintf(`<a %s>%s</a>`, strings.Join(parts, " "), text)
}
