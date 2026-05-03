package display_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/display"
)

// ansiUnderlineSGR matches any SGR escape that carries the underline
// parameter (4) as one of its semicolon-separated values. Glamour merges
// underline with adjacent style parameters (e.g. foreground colour), so
// the underline byte rarely appears as the standalone "\x1b[4m" sequence.
var ansiUnderlineSGR = regexp.MustCompile(`\x1b\[(?:[0-9]+;)*4(?:;[0-9]+)*m`)

func underlineDoc(content []adf.Node) *adf.Document {
	return &adf.Document{
		Type:    "doc",
		Version: 1,
		Content: []adf.Node{
			{
				Type:    "paragraph",
				Content: content,
			},
		},
	}
}

// TestRender_UnderlineMarkProducesAnsi checks that an underline mark
// reaches the terminal as an ANSI underline SGR, not as a literal
// <u>…</u> tag and not stripped to bare text.
//
// Composition with other inline marks (bold, textColor) is intentionally
// not exercised here: the parser eats the inner content as a raw text
// segment, so nested marks do not re-parse — the same limitation
// ColorSpan has. See backlog ac-0132 for the composition follow-up.
func TestRender_UnderlineMarkProducesAnsi(t *testing.T) {
	cases := []struct {
		name        string
		content     []adf.Node
		expects     []string // substrings that MUST appear
		excludes    []string // substrings that must NOT appear
		wantUnderln bool     // must contain ANSI underline SGR somewhere
	}{
		{
			name: "plain underline",
			content: []adf.Node{
				{Type: "text", Text: "under", Marks: []adf.Mark{{Type: adf.MarkTypeUnderline}}},
			},
			expects:     []string{"under"},
			excludes:    []string{"<u>", "</u>"},
			wantUnderln: true,
		},
		{
			name: "empty underline",
			content: []adf.Node{
				{Type: "text", Text: "before "},
				{Type: "text", Text: "", Marks: []adf.Mark{{Type: adf.MarkTypeUnderline}}},
				{Type: "text", Text: " after"},
			},
			expects:  []string{"before", "after"},
			excludes: []string{"<u>", "</u>"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := underlineDoc(tc.content)
			out, err := display.Render(doc)
			if err != nil {
				t.Fatalf("display.Render: %v", err)
			}
			for _, want := range tc.expects {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q\nfull output: %q", want, out)
				}
			}
			for _, bad := range tc.excludes {
				if strings.Contains(out, bad) {
					t.Errorf("output contains forbidden %q\nfull output: %q", bad, out)
				}
			}
			if tc.wantUnderln && !ansiUnderlineSGR.MatchString(out) {
				t.Errorf("output missing ANSI underline SGR\nfull output: %q", out)
			}
		})
	}
}
