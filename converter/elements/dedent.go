package elements

import "strings"

// DedentLines removes the minimum common indentation from a slice of lines
// while preserving relative indentation. This is useful for processing content
// that has base indentation (e.g., from being inside HTML tags) while preserving
// nested structure (e.g., nested list items).
//
// Example:
//
//	Input:  ["  - first", "    - sub", "  - second"]
//	Output: ["- first", "  - sub", "- second"]
//
// The function finds the minimum indentation across all non-empty lines,
// then strips only that amount from each line, preserving the relative
// indentation that indicates nesting or structure.
func DedentLines(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}

	// Find minimum indentation across all non-empty lines
	minIndent := -1
	for _, line := range lines {
		if trimmed := strings.TrimLeft(line, " \t"); len(trimmed) > 0 {
			indent := len(line) - len(trimmed)
			if minIndent == -1 || indent < minIndent {
				minIndent = indent
			}
		}
	}

	// If all lines are empty or whitespace-only, normalize whitespace to empty
	if minIndent == -1 {
		normalized := make([]string, len(lines))
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				normalized[i] = ""
			} else {
				normalized[i] = line
			}
		}
		return normalized
	}

	// Strip minimum indentation from all lines
	dedented := make([]string, 0, len(lines))
	for _, line := range lines {
		// Handle empty or whitespace-only lines first
		if strings.TrimSpace(line) == "" {
			dedented = append(dedented, "")
			continue
		}

		// Strip minimum indentation from content lines
		if len(line) >= minIndent {
			dedented = append(dedented, line[minIndent:])
		} else {
			// Line is shorter than minIndent but has content - keep as-is
			dedented = append(dedented, line)
		}
	}

	return dedented
}
