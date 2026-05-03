package elements

import (
	"errors"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// textDisplayRenderer renders text nodes for read-only display mode. It
// reuses the edit-mode mark pipeline for marks that survive Markdown
// rendering verbatim and overrides only the marks that look ugly in a
// terminal — currently subsup (<sub>/<sup> → Unicode super-/subscripts
// with an ASCII fallback). The textColor mark inherits the edit-mode
// rendering (HTML span); the display/ Glamour pipeline picks the span
// up via a custom renderer extension.
type textDisplayRenderer struct {
	pipeline markPipeline
}

// NewTextDisplayRenderer returns a display-mode text renderer. Subsup
// uses a Unicode-first rendering with ASCII fallback; textColor falls
// through to the edit-mode span renderer so the display/ Glamour bridge
// can colorize it.
func NewTextDisplayRenderer() adf.Renderer {
	overrides := markPipeline{
		adf.MarkTypeSubsup: renderSubsupDisplay,
		adf.MarkTypeLink:   renderLinkMarkDisplay,
	}
	return &textDisplayRenderer{
		pipeline: editMarkPipeline.withOverrides(overrides),
	}
}

func (r *textDisplayRenderer) ToMarkdown(node adf.Node, _ adf.ConversionContext) (adf.RenderResult, error) {
	text := node.Text
	for _, mark := range node.Marks {
		text = r.pipeline.apply(text, mark)
	}

	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)
	builder.AppendContent(text)
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (r *textDisplayRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, errors.New("text display renderer is read-only")
}

// supMap and subMap encode the Unicode coverage for super-/subscript
// characters. Missing keys force the ASCII fallback for the whole mark
// content (no partial mixing — see Plan, Phase 5).
var supMap = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
	'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ',
	'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ', 'o': 'ᵒ',
	'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ',
	'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
}

var subMap = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
	'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
	'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
	'v': 'ᵥ', 'x': 'ₓ',
}

// renderLinkMarkDisplay collapses links whose visible text equals the
// href into Markdown autolink form (<URL>). When a title is present the
// inline form wins because autolinks cannot carry titles. All other
// cases delegate to the edit-mode renderLinkMark.
func renderLinkMarkDisplay(text string, mark adf.Mark) string {
	href, ok := mark.Attrs["href"].(string)
	if !ok || href != text {
		return renderLinkMark(text, mark)
	}
	if title, ok := mark.Attrs["title"].(string); ok && title != "" {
		return renderLinkMark(text, mark)
	}
	return "<" + href + ">"
}

func renderSubsupDisplay(text string, mark adf.Mark) string {
	isSup := false
	if subType, ok := mark.Attrs["type"].(string); ok && subType == "sup" {
		isSup = true
	}

	table := subMap
	asciiPrefix := "_"
	if isSup {
		table = supMap
		asciiPrefix = "^"
	}

	if mapped, ok := tryMapAll(text, table); ok {
		return mapped
	}

	// ASCII fallback. Single-char form (_x / ^x) only when text holds
	// exactly one rune; multi-rune content uses the brace form to keep
	// the boundary unambiguous against following text.
	if runeCount(text) == 1 {
		return asciiPrefix + text
	}
	return asciiPrefix + "{" + text + "}"
}

func tryMapAll(text string, table map[rune]rune) (string, bool) {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		mapped, ok := table[r]
		if !ok {
			return "", false
		}
		b.WriteRune(mapped)
	}
	return b.String(), true
}

func runeCount(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
