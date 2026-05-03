package defaults_test

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
)

// TestDisplaySample_AcceptanceCriteria asserts the display-mode markdown
// output meets the ac-0126 acceptance bar. We pin substrings rather than
// full output so downstream renderer drift cannot flake the suite. ANSI
// escapes and Glamour rendering live in the separate display/ module —
// adf-converter only emits plain Markdown.
func TestDisplaySample_AcceptanceCriteria(t *testing.T) {
	md := renderDisplaySampleMarkdown(t)

	t.Run("no fenced-div panel leakage", func(t *testing.T) {
		if strings.Contains(md, ":::info") {
			t.Errorf(":::info residue found in display output")
		}
		if strings.Contains(md, ":::warning") {
			t.Errorf(":::warning residue found in display output")
		}
	})

	t.Run("mention is not double-rendered as accountid link", func(t *testing.T) {
		if strings.Contains(md, "accountid:") {
			t.Errorf("accountid: residue from edit-mode mention rendering")
		}
		if strings.Contains(md, "[@john.doe](") {
			t.Errorf("mention rendered as Markdown link, expected plain @name")
		}
	})

	t.Run("textColor emits HTML span for downstream color pickup", func(t *testing.T) {
		if !strings.Contains(md, `<span style="color:`) {
			t.Errorf("textColor missing span; display/ Glamour bridge cannot colorize")
		}
	})

	t.Run("subsup mark not leaked as HTML tag", func(t *testing.T) {
		if strings.Contains(md, "<sub>") || strings.Contains(md, "<sup>") {
			t.Errorf("subsup leaked as <sub>/<sup> HTML")
		}
	})

	t.Run("inlineCard URL appears exactly once per occurrence", func(t *testing.T) {
		const url = "https://example.com/page"
		got := strings.Count(md, url)
		if got != 1 {
			t.Errorf("URL %q occurrences = %d, want 1", url, got)
		}
	})

	t.Run("subsup unicode mapping applied", func(t *testing.T) {
		if !strings.Contains(md, "H₂O") {
			t.Errorf("expected sub digit Unicode H₂O in display output")
		}
		if !strings.Contains(md, "xⁿ") {
			t.Errorf("expected sup letter Unicode xⁿ in display output")
		}
	})

	t.Run("status renders as bracketed label", func(t *testing.T) {
		if !strings.Contains(md, "[In Progress]") {
			t.Errorf("expected bracketed status label [In Progress] in display output")
		}
	})
}

func renderDisplaySampleMarkdown(t *testing.T) string {
	t.Helper()
	raw := readFixture(t)
	doc, err := adf.ParseFromString(string(raw))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	conv := defaults.NewDisplayConverter()
	md, _, err := conv.ToMarkdown(*doc)
	if err != nil {
		t.Fatalf("display ToMarkdown: %v", err)
	}
	return md
}
