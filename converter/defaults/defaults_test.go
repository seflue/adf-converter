package defaults_test

import (
	"testing"

	"github.com/seflue/adf-converter/converter/defaults"
	"github.com/seflue/adf-converter/converter/elements"
)

// TestRegistrationSourceOfTruth guards against drift between
// defaults.NewRegistry() and elements.StandardNodes/StandardBlockParserOrder.
// If someone adds a converter to NewRegistry without updating StandardNodes
// (or vice versa), this test fails. Covers ac-0114 Finding E22.
func TestRegistrationSourceOfTruth(t *testing.T) {
	r := defaults.NewRegistry()

	standard := elements.StandardNodes()
	if got, want := r.Count(), len(standard); got != want {
		t.Fatalf("registry converter count = %d, want %d", got, want)
	}
	for _, reg := range standard {
		if r.GetConverter(reg.NodeType) == nil {
			t.Errorf("node %q: converter not registered", reg.NodeType)
		}
	}

	parsers := r.BlockParsers()
	if got, want := len(parsers), len(elements.StandardBlockParserOrder); got != want {
		t.Fatalf("block parser count = %d, want %d", got, want)
	}
	for i, entry := range parsers {
		if entry.NodeType != elements.StandardBlockParserOrder[i] {
			t.Errorf("block parser[%d] = %q, want %q",
				i, entry.NodeType, elements.StandardBlockParserOrder[i])
		}
	}
}
