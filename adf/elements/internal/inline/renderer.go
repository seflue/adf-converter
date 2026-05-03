package inline

import (
	"fmt"
	"sort"
	"strings"

	"github.com/seflue/adf-converter/adf"
)

// Marks with a fixed open/close pair. When adjacent text nodes share one
// of these, the wrapper is opened once and closed at the end of the run
// (so three text nodes with **strong** become one **abc**, not three).
var markDelimiter = map[adf.MarkType]string{
	adf.MarkTypeStrong: "**",
	adf.MarkTypeEm:     "*", // italic
	adf.MarkTypeStrike: "~~",
}

// markPriority determines nesting order: lower = outermost.
// Strong wraps everything, em is innermost.
var markPriority = map[adf.MarkType]int{
	adf.MarkTypeStrong: 0,
	adf.MarkTypeStrike: 1,
	adf.MarkTypeEm:     2,
}

func hasSharedDelimiter(markType adf.MarkType) bool {
	_, ok := markDelimiter[markType]
	return ok
}

// Converts ADF inline nodes to markdown.
//
// Marks listed in markDelimiter (strong/em/strike) get one wrapper
// across adjacent nodes that share them; everything else wraps each
// node on its own.
//
// Whitespace extrusion: leading/trailing spaces are moved outside the
// delimiters, so a closing **bold** followed by an opening *italic*
// doesn't fuse into ***.
func RenderInlineNodes(nodes []adf.Node, context adf.ConversionContext) (string, error) {
	var result strings.Builder
	var openMarks []adf.MarkType // currently open shared-wrapper marks (sorted by priority)
	var deferredSpace string

	for _, node := range nodes {
		if node.Type != adf.NodeTypeText {
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

		sharedTarget, perNode := splitMarks(node.Marks)
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

		transition(&result, &openMarks, sharedTarget, separator)

		// Extrude trailing whitespace (defer for next transition)
		if trimmed := strings.TrimRight(text, " "); len(trimmed) < len(text) {
			deferredSpace = text[len(trimmed):]
			text = trimmed
		}

		rendered, err := applyPerNodeMarks(text, perNode, node, context)
		if err != nil {
			return "", err
		}
		result.WriteString(rendered)
	}

	closeAll(&result, &openMarks)
	result.WriteString(deferredSpace)
	return result.String(), nil
}

// Splits marks into the two render strategies: shared (one wrapper for
// a run of nodes) vs. per-node.
func splitMarks(marks []adf.Mark) (shared []adf.MarkType, perNode []adf.Mark) {
	for _, m := range marks {
		if hasSharedDelimiter(m.Type) {
			shared = append(shared, m.Type)
		} else {
			perNode = append(perNode, m)
		}
	}
	sort.Slice(shared, func(i, j int) bool {
		return markPriority[shared[i]] < markPriority[shared[j]]
	})
	return
}

// transition closes marks that are no longer needed and opens new ones.
// The separator (extruded whitespace) is placed between closing and opening
// delimiters to prevent ambiguous sequences like ***.
func transition(w *strings.Builder, openMarks *[]adf.MarkType, target []adf.MarkType, separator string) {
	commonLen := 0
	for commonLen < len(*openMarks) && commonLen < len(target) {
		if (*openMarks)[commonLen] != target[commonLen] {
			break
		}
		commonLen++
	}

	// Close marks after common prefix (innermost first = reverse order)
	for i := len(*openMarks) - 1; i >= commonLen; i-- {
		w.WriteString(markDelimiter[(*openMarks)[i]])
	}

	// Place separator between closing and opening delimiters.
	w.WriteString(separator)

	// Open marks after common prefix
	for i := commonLen; i < len(target); i++ {
		w.WriteString(markDelimiter[target[i]])
	}

	*openMarks = make([]adf.MarkType, len(target))
	copy(*openMarks, target)
}

// closeAll closes all currently open marks (innermost first).
func closeAll(w *strings.Builder, openMarks *[]adf.MarkType) {
	for i := len(*openMarks) - 1; i >= 0; i-- {
		w.WriteString(markDelimiter[(*openMarks)[i]])
	}
	*openMarks = nil
}

// Wraps one text node with its per-node marks (code, link, underline,
// textColor, subsup) by delegating to the registry's text renderer. That
// detour matters: RenderInlineNodes writes the shared marks directly
// without going through the registry, so display-mode overrides for
// textColor/subsup would otherwise never run.
//
// Workaround: routing through the text renderer is the cheap seam to
// reach the mark pipeline. The cleaner home is a MarkPipeline on
// ConversionContext that both inline and text can use directly.
func applyPerNodeMarks(text string, marks []adf.Mark, node adf.Node, context adf.ConversionContext) (string, error) {
	if len(marks) == 0 {
		return text, nil
	}

	renderer, ok := context.Registry.Lookup(adf.NodeTypeText)
	if !ok || renderer == nil {
		return "", fmt.Errorf("inline: text renderer not registered")
	}
	textNode := adf.Node{Type: node.Type, Text: text, Marks: marks}
	res, err := renderer.ToMarkdown(textNode, context)
	if err != nil {
		return "", fmt.Errorf("inline text renderer: %w", err)
	}
	return res.Content, nil
}

// renderNonTextNode delegates rendering of non-text inline nodes to their converters.
func renderNonTextNode(node adf.Node, context adf.ConversionContext) (string, error) {
	c, _ := context.Registry.Lookup(adf.NodeType(node.Type))
	if c == nil {
		return "", fmt.Errorf("no converter found for inline node type: %s", node.Type)
	}
	res, err := c.ToMarkdown(node, context)
	if err != nil {
		return "", err
	}
	return res.Content, nil
}
