package adf_test

import (
	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"testing"

	"github.com/seflue/adf-converter/placeholder"
)

func TestDefaultClassifier_MentionIsEditable(t *testing.T) {
	c := adf.NewDefaultClassifier()

	if !c.IsEditable(adf.NodeTypeMention) {
		t.Error("mention should be editable")
	}
	if c.IsPreserved(adf.NodeTypeMention) {
		t.Error("mention should not be preserved")
	}
}

func TestDefaultClassifier_TableIsEditable(t *testing.T) {
	c := adf.NewDefaultClassifier()

	if !c.IsEditable(adf.NodeTypeTable) {
		t.Error("table should be editable")
	}
	if c.IsPreserved(adf.NodeTypeTable) {
		t.Error("table should not be preserved")
	}
}

func TestTableConverter_RegisteredInRegistry(t *testing.T) {
	c := defaults.NewRegistry().GetConverter("table")
	if c == nil {
		t.Fatal("table converter should be registered in global registry")
	}
	if !c.CanHandle("table") {
		t.Error("registered converter should handle table type")
	}
}

func TestDefaultClassifier_TableSubNodesNotPreserved(t *testing.T) {
	c := adf.NewDefaultClassifier()

	if c.IsPreserved(adf.NodeTypeTableRow) {
		t.Error("tableRow should not be preserved (sub-node of table)")
	}
	if c.IsPreserved(adf.NodeTypeTableCell) {
		t.Error("tableCell should not be preserved (sub-node of table)")
	}
}

func TestParseNext_DetectsPlainMarkdownTable(t *testing.T) {
	manager := placeholder.NewManager()
	session := manager.GetSession()
	parser := adf.NewMarkdownParser(session, manager, defaults.NewRegistry())

	lines := []string{
		"| Header 1 | Header 2 |",
		"|----------|----------|",
		"| Cell 1   | Cell 2   |",
	}

	nodes, err := parser.ParseMarkdownToADFNodes(lines)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != "table" {
		t.Errorf("expected table node, got %s", nodes[0].Type)
	}
	if len(nodes[0].Content) != 2 {
		t.Errorf("expected 2 rows, got %d", len(nodes[0].Content))
	}
}
