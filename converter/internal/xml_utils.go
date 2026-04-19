package internal

import (
	"regexp"
	"strconv"
)

// ParseXMLAttributes extracts attributes from an XML opening tag.
// Matches: attribute="value" or attribute='value'
// Converts boolean strings ("true"/"false") and numeric strings to appropriate types.
func ParseXMLAttributes(xmlTag string) map[string]any {
	attrs := make(map[string]any)

	// Simple regex-based attribute parsing
	// Matches: attribute="value" or attribute='value'
	// Support attribute names with hyphens and underscores
	attrRegex := regexp.MustCompile(`([\w-]+)=["']([^"']*)["']`)
	matches := attrRegex.FindAllStringSubmatch(xmlTag, -1)

	for _, match := range matches {
		if len(match) == 3 {
			key := match[1]
			value := match[2]

			// Convert boolean and numeric values
			switch value {
			case "true":
				attrs[key] = true
			case "false":
				attrs[key] = false
			default:
				// Try to parse as integer
				if intVal, err := strconv.Atoi(value); err == nil {
					attrs[key] = intVal
				} else {
					attrs[key] = value
				}
			}
		}
	}

	return attrs
}
