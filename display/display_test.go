package display_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/display"
)

// fixturePath points at the hand-composed display sample owned by
// adf-converter. The display module reuses it instead of duplicating a
// fixture: the fixture is the contract between the two modules and lives
// where the renderers themselves live.
const fixturePath = "../adf/defaults/testdata/display-sample.json"

func loadFixtureDoc(t *testing.T) *adf.Document {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(fixturePath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	doc, err := adf.ParseFromString(string(raw))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return doc
}

// TestRender_PipelineRunsCrashFree exercises the full ADF → display-MD →
// Glamour pipeline on the hand-composed fixture. We assert no error and
// non-empty output; we do not pin Glamour-specific styling because that
// drifts across Glamour versions and themes.
func TestRender_PipelineRunsCrashFree(t *testing.T) {
	doc := loadFixtureDoc(t)

	out, err := display.Render(doc)
	if err != nil {
		t.Fatalf("display.Render: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("display.Render returned empty output")
	}
}

// TestRender_TextColorSpanConsumed verifies the color-span extension
// captures <span style="color: …">…</span> tokens emitted by
// adf-converter so they neither leak as raw HTML nor get stripped to
// bare text. The visible word survives, the markup tag does not.
func TestRender_TextColorSpanConsumed(t *testing.T) {
	doc := loadFixtureDoc(t)

	out, err := display.Render(doc)
	if err != nil {
		t.Fatalf("display.Render: %v", err)
	}
	if strings.Contains(out, `<span style="color:`) {
		t.Errorf("color span leaked as raw HTML in rendered output")
	}
	if !strings.Contains(out, "red") {
		t.Errorf("expected wrapped text 'red' to survive in rendered output")
	}
	if !strings.Contains(out, "\x1b[") {
		t.Errorf("expected ANSI escape sequences in rendered output (color applied?)")
	}
}

// TestRender_OptionsAccepted asserts the functional-options API compiles
// and the renderer accepts the public Option set. Behavioural pinning
// stays out — Glamour theme + wrap behaviour is not the contract we own.
func TestRender_OptionsAccepted(t *testing.T) {
	doc := loadFixtureDoc(t)

	out, err := display.Render(doc,
		display.WithStyle("dark"),
		display.WithWordWrap(80),
	)
	if err != nil {
		t.Fatalf("display.Render with options: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("display.Render with options returned empty output")
	}
}
