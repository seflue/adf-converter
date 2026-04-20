package converter_test

import (
	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/placeholder"
)

// testToMarkdown wraps DefaultConverter.ToMarkdown with the classifier/manager/registry
// injection that older tests used. Kept package-local so tests exercise the public
// method API without a wide mechanical rewrite.
func testToMarkdown(doc adf_types.ADFDocument, classifier converter.ContentClassifier, manager placeholder.Manager, registry *converter.ConverterRegistry) (string, *placeholder.EditSession, error) {
	c, err := converter.NewConverter(
		converter.WithClassifier(classifier),
		converter.WithPlaceholderManager(manager),
		converter.WithRegistry(registry),
	)
	if err != nil {
		return "", nil, err
	}
	return c.ToMarkdown(doc)
}

// testFromMarkdown wraps DefaultConverter.FromMarkdown and unwraps ConversionResult.Document
// to match the legacy top-level FromMarkdown return shape used in tests.
func testFromMarkdown(markdown string, session *placeholder.EditSession, manager placeholder.Manager, registry *converter.ConverterRegistry) (adf_types.ADFDocument, error) {
	c, err := converter.NewConverter(
		converter.WithPlaceholderManager(manager),
		converter.WithRegistry(registry),
	)
	if err != nil {
		return adf_types.ADFDocument{}, err
	}
	result, err := c.FromMarkdown(markdown, session)
	return result.Document, err
}
