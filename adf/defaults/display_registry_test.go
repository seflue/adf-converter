package defaults_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/adf/elements"
)

// TestNewDisplayRegistry_ParityWithEditMode guards the skeleton phase
// of ac-0126: NewDisplayRegistry must register the same nodes and block
// parsers as NewRegistry. Display-specific overrides are layered on top
// in later phases.
func TestNewDisplayRegistry_ParityWithEditMode(t *testing.T) {
	r := defaults.NewDisplayRegistry()

	standard := elements.StandardNodes()
	if got, want := r.Count(), len(standard); got != want {
		t.Fatalf("display registry converter count = %d, want %d", got, want)
	}
	for _, reg := range standard {
		if _, ok := r.Lookup(reg.NodeType); !ok {
			t.Errorf("node %q: converter not registered in display registry", reg.NodeType)
		}
	}

	parsers := r.BlockParsers()
	if got, want := len(parsers), len(elements.StandardBlockParserOrder); got != want {
		t.Fatalf("display registry block parser count = %d, want %d", got, want)
	}
	for i, entry := range parsers {
		if entry.NodeType != elements.StandardBlockParserOrder[i] {
			t.Errorf("display registry block parser[%d] = %q, want %q",
				i, entry.NodeType, elements.StandardBlockParserOrder[i])
		}
	}
}

