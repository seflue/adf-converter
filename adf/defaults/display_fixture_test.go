package defaults_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
)

const fixturePath = "testdata/display-sample.json"

// requiredNodeTypes lists the ADF node types the display-sample fixture
// must contain (Phase 6 acceptance — display mode is only meaningful if
// the fixture exercises every renderer that ships with NewDisplayRegistry).
var requiredNodeTypes = []adf.NodeType{
	"doc", "paragraph", "text", "panel", "mention", "inlineCard", "status",
}

// requiredMarkTypes lists the mark types the fixture must carry. Mirrors
// the marks the display renderers can affect (textColor, subsup) plus
// the standard inline marks the text-display pipeline still delegates
// to the edit-mode pipeline.
var requiredMarkTypes = []adf.MarkType{
	"em", "strong", "code", "link", "subsup", "textColor",
}

// TestDisplaySample_IsValidADF is gate 1 for Phase 6: it guarantees the
// hand-composed fixture parses as ADF, survives an edit-mode roundtrip,
// and still carries every node/mark the smoke test depends on.
func TestDisplaySample_IsValidADF(t *testing.T) {
	raw := readFixture(t)

	doc, err := adf.ParseFromString(string(raw))
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	assertNodesPresent(t, "input fixture", doc.Content, requiredNodeTypes)
	assertMarksPresent(t, "input fixture", doc.Content, requiredMarkTypes)

	conv := defaults.NewDefaultConverter()
	md, session, err := conv.ToMarkdown(*doc)
	if err != nil {
		t.Fatalf("edit-mode ToMarkdown: %v", err)
	}
	if md == "" {
		t.Fatal("edit-mode ToMarkdown returned empty markdown")
	}

	roundtripDoc, _, err := conv.FromMarkdown(md, session)
	if err != nil {
		t.Fatalf("edit-mode FromMarkdown: %v", err)
	}

	assertNodesPresent(t, "edit-mode roundtrip", roundtripDoc.Content, requiredNodeTypes)
	assertMarksPresent(t, "edit-mode roundtrip", roundtripDoc.Content, requiredMarkTypes)
}

func readFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Clean(fixturePath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return raw
}

func assertNodesPresent(t *testing.T, label string, nodes []adf.Node, required []adf.NodeType) {
	t.Helper()
	seen := collectNodeTypes(nodes)
	// "doc" lives one level above the slice we're handed; treat its
	// presence as implicit because we already parsed a Document.
	seen["doc"] = struct{}{}
	for _, want := range required {
		if _, ok := seen[want]; !ok {
			t.Errorf("%s: required node type %q missing", label, want)
		}
	}
}

func assertMarksPresent(t *testing.T, label string, nodes []adf.Node, required []adf.MarkType) {
	t.Helper()
	seen := collectMarkTypes(nodes)
	for _, want := range required {
		if _, ok := seen[want]; !ok {
			t.Errorf("%s: required mark type %q missing", label, want)
		}
	}
}

func collectNodeTypes(nodes []adf.Node) map[adf.NodeType]struct{} {
	out := map[adf.NodeType]struct{}{}
	var walk func([]adf.Node)
	walk = func(ns []adf.Node) {
		for _, n := range ns {
			out[adf.NodeType(n.Type)] = struct{}{}
			walk(n.Content)
		}
	}
	walk(nodes)
	return out
}

func collectMarkTypes(nodes []adf.Node) map[adf.MarkType]struct{} {
	out := map[adf.MarkType]struct{}{}
	var walk func([]adf.Node)
	walk = func(ns []adf.Node) {
		for _, n := range ns {
			for _, m := range n.Marks {
				out[m.Type] = struct{}{}
			}
			walk(n.Content)
		}
	}
	walk(nodes)
	return out
}
