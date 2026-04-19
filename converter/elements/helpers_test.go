package elements

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/placeholder"
)

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
		p := converter.NewMarkdownParserWithNesting(mgr.GetSession(), mgr, nestedLevel)
		return p.ParseMarkdownToADFNodes(lines)
	}
}
