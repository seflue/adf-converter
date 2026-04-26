package adf_test

import (
	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/placeholder"
)

// testToMarkdown wraps DefaultConverter.ToMarkdown with the classifier/manager/registry
// injection that older tests used. Kept package-local so tests exercise the public
// method API without a wide mechanical rewrite.
func testToMarkdown(doc adf.Document, classifier adf.ContentClassifier, manager placeholder.Manager, registry *adf.ConverterRegistry) (string, *placeholder.EditSession, error) {
	c, err := adf.NewConverter(
		adf.WithClassifier(classifier),
		adf.WithPlaceholderManager(manager),
		adf.WithRegistry(registry),
	)
	if err != nil {
		return "", nil, err
	}
	return c.ToMarkdown(doc)
}

// testFromMarkdown wraps DefaultConverter.FromMarkdown and unwraps ConversionResult.Document
// to match the legacy top-level FromMarkdown return shape used in tests.
func testFromMarkdown(markdown string, session *placeholder.EditSession, manager placeholder.Manager, registry *adf.ConverterRegistry) (adf.Document, error) {
	c, err := adf.NewConverter(
		adf.WithPlaceholderManager(manager),
		adf.WithRegistry(registry),
	)
	if err != nil {
		return adf.Document{}, err
	}
	result, err := c.FromMarkdown(markdown, session)
	return result.Document, err
}
