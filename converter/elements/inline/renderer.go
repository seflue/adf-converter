package inline

import (
	"fmt"
	"sort"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter"
)

// markDelimiter maps spannable mark types to their markdown delimiters.
// These marks can be opened once and span across multiple text nodes.
var markDelimiter = map[string]string{
	adf_types.MarkTypeStrong: "**",
	adf_types.MarkTypeEm:    "*",
	adf_types.MarkTypeStrike: "~~",
}

// markPriority determines nesting order: lower = outermost.
// Strong wraps everything, em is innermost.
var markPriority = map[string]int{
	adf_types.MarkTypeStrong: 0,
	adf_types.MarkTypeStrike: 1,
	adf_types.MarkTypeEm:    2,
}

func isSpannable(markType string) bool {
	_, ok := markDelimiter[markType]
	return ok
}

// RenderInlineNodes converts a slice of ADF inline nodes to markdown with
// proper mark spanning. Shared marks across adjacent text nodes are opened
// once instead of being duplicated at each node boundary.
//
// Whitespace extrusion: trailing/leading spaces are moved outside mark
// delimiters to prevent ambiguous sequences like *** (bold close + italic open).
// The extruded space is placed between closing and opening delimiters in the
// transition function, preserving mark spanning for shared marks.
func RenderInlineNodes(nodes []adf_types.ADFNode, context converter.ConversionContext) (string, error) {
	var result strings.Builder
	var openMarks []string // stack of currently open spannable marks (sorted by priority)
	var deferredSpace string

	for _, node := range nodes {
		if node.Type != adf_types.NodeTypeText {
			closeAll(&result, &openMarks)
			result.WriteString(deferredSpace)
			deferredSpace = ""
			rendered, err := renderNonTextNode(node, context)
			if err != nil {
				return "", err
			}
			result.WriteString(rendered)
			continue
		}

		targetSpannable, nonSpannable := splitMarks(node.Marks)
		text := node.Text

		// Extrude leading whitespace from text
		var leadingSpace string
		if trimmed := strings.TrimLeft(text, " "); len(trimmed) < len(text) {
			leadingSpace = text[:len(text)-len(trimmed)]
			text = trimmed
		}

		// Combine deferred trailing space + leading space as separator
		separator := deferredSpace + leadingSpace
		deferredSpace = ""

		transition(&result, &openMarks, targetSpannable, separator)

		// Extrude trailing whitespace (defer for next transition)
		if trimmed := strings.TrimRight(text, " "); len(trimmed) < len(text) {
			deferredSpace = text[len(trimmed):]
			text = trimmed
		}

		for _, m := range nonSpannable {
			text = applyWrappingMark(text, m)
		}
		result.WriteString(text)
	}

	closeAll(&result, &openMarks)
	result.WriteString(deferredSpace)
	return result.String(), nil
}

// splitMarks separates marks into spannable (strong/em/strike) and
// non-spannable (link/code/underline/textColor/subsup).
func splitMarks(marks []adf_types.ADFMark) (spannable []string, nonSpannable []adf_types.ADFMark) {
	for _, m := range marks {
		if isSpannable(m.Type) {
			spannable = append(spannable, m.Type)
		} else {
			nonSpannable = append(nonSpannable, m)
		}
	}
	sort.Slice(spannable, func(i, j int) bool {
		return markPriority[spannable[i]] < markPriority[spannable[j]]
	})
	return
}

// transition closes marks that are no longer needed and opens new ones.
// The separator (extruded whitespace) is placed between closing and opening
// delimiters to prevent ambiguous sequences like ***.
func transition(w *strings.Builder, openMarks *[]string, target []string, separator string) {
	commonLen := 0
	for commonLen < len(*openMarks) && commonLen < len(target) {
		if (*openMarks)[commonLen] != target[commonLen] {
			break
		}
		commonLen++
	}

	closing := len(*openMarks) > commonLen
	opening := commonLen < len(target)

	// Close marks after common prefix (innermost first = reverse order)
	for i := len(*openMarks) - 1; i >= commonLen; i-- {
		w.WriteString(markDelimiter[(*openMarks)[i]])
	}

	// Place separator between closing and opening delimiters.
	// When nothing is closing or opening, emit inside current marks.
	if closing || opening {
		w.WriteString(separator)
	} else {
		w.WriteString(separator)
	}

	// Open marks after common prefix
	for i := commonLen; i < len(target); i++ {
		w.WriteString(markDelimiter[target[i]])
	}

	*openMarks = make([]string, len(target))
	copy(*openMarks, target)
}

// closeAll closes all currently open marks (innermost first).
func closeAll(w *strings.Builder, openMarks *[]string) {
	for i := len(*openMarks) - 1; i >= 0; i-- {
		w.WriteString(markDelimiter[(*openMarks)[i]])
	}
	*openMarks = nil
}

// applyWrappingMark applies a non-spannable mark to text (per-node wrapping).
func applyWrappingMark(text string, mark adf_types.ADFMark) string {
	switch mark.Type {
	case adf_types.MarkTypeCode:
		return fmt.Sprintf("`%s`", text)
	case adf_types.MarkTypeLink:
		if href, ok := mark.Attrs["href"].(string); ok {
			if title, ok := mark.Attrs["title"].(string); ok && title != "" {
				return fmt.Sprintf(`[%s](%s "%s")`, text, href, title)
			}
			return fmt.Sprintf("[%s](%s)", text, href)
		}
		return text
	case adf_types.MarkTypeUnderline:
		return fmt.Sprintf("<u>%s</u>", text)
	case adf_types.MarkTypeTextColor:
		if color, ok := mark.Attrs["color"].(string); ok {
			return fmt.Sprintf(`<span style="color: %s">%s</span>`, color, text)
		}
		return text
	case adf_types.MarkTypeSubsup:
		if subType, ok := mark.Attrs["type"].(string); ok && subType == "sup" {
			return fmt.Sprintf("<sup>%s</sup>", text)
		}
		return fmt.Sprintf("<sub>%s</sub>", text)
	default:
		return text
	}
}

// renderNonTextNode delegates rendering of non-text inline nodes to their converters.
func renderNonTextNode(node adf_types.ADFNode, context converter.ConversionContext) (string, error) {
	c := converter.GetGlobalRegistry().GetConverter(converter.ADFNodeType(node.Type))
	if c == nil {
		return "", fmt.Errorf("no converter found for inline node type: %s", node.Type)
	}
	res, err := c.ToMarkdown(node, context)
	if err != nil {
		return "", err
	}
	return res.Content, nil
}
