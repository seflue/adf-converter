package elements

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/element"
	"github.com/seflue/adf-converter/placeholder"
)

// newTestRegistry builds a registry populated with the standard element
// converters plus the canonical block-parser order. It iterates over the
// exported StandardNodes / StandardBlockParserOrder so converter/defaults and
// the elements-package tests stay in lockstep (ac-0114).
func newTestRegistry() *element.ConverterRegistry {
	r := element.NewConverterRegistry()
	for _, reg := range StandardNodes() {
		r.MustRegister(reg.NodeType, reg.Converter)
	}
	for _, nodeType := range StandardBlockParserOrder {
		r.MustRegisterBlockParser(nodeType)
	}
	return r
}

// testParseNested wires a MarkdownParser-backed ParseNested closure for tests
// that exercise element converters expecting nested markdown parsing.
func testParseNested() func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
	mgr := placeholder.NewManager()
	return testParseNestedWith(mgr)
}

// testParseNestedWith wires ParseNested using the given manager so tests that
// pre-stored placeholders can recover them during nested parsing.
func testParseNestedWith(mgr placeholder.Manager) func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
	return func(lines []string, nestedLevel int) ([]adf_types.ADFNode, error) {
		p := element.NewMarkdownParserWithNesting(mgr.GetSession(), mgr, newTestRegistry(), nestedLevel)
		return p.ParseMarkdownToADFNodes(lines)
	}
}
