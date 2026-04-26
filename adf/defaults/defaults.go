// Package defaults wires the eager registration of all standard ADF element
// converters. Callers that want a ready-to-use converter without assembling
// the registry themselves should use NewDefaultConverter or NewDisplayConverter.
//
// This package is intentionally separate from converter/ to avoid an import
// cycle: the elements/ package depends on converter/, so converter/ cannot
// import elements/ to eager-register. defaults/ closes that gap by being the
// single place that knows about both.
package defaults

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/elements"
	"github.com/seflue/adf-converter/placeholder"
)

// NewRegistry builds a fresh registry with all standard element converters
// registered and the block-parser dispatch order wired. The node list and the
// dispatch order are sourced from elements.StandardNodes and
// elements.StandardBlockParserOrder so defaults and the elements-package test
// helpers share a single source of truth (ac-0114).
func NewRegistry() *adf.ConverterRegistry {
	r := adf.NewConverterRegistry()

	for _, reg := range elements.StandardNodes() {
		r.MustRegister(reg.NodeType, reg.Converter)
	}
	for _, nodeType := range elements.StandardBlockParserOrder {
		r.MustRegisterBlockParser(nodeType)
	}

	return r
}

// NewDefaultConverter returns a converter wired with all standard element
// converters and the default classifier and placeholder manager.
func NewDefaultConverter() *adf.DefaultConverter {
	c, err := adf.NewConverter(
		adf.WithRegistry(NewRegistry()),
	)
	if err != nil {
		panic(fmt.Sprintf("defaults: unreachable: %v", err))
	}
	return c
}

// NewDisplayConverter returns a converter for read-only display mode.
// It uses a NullManager that produces preview text instead of placeholder comments.
func NewDisplayConverter() *adf.DefaultConverter {
	c, err := adf.NewConverter(
		adf.WithRegistry(NewRegistry()),
		adf.WithPlaceholderManager(placeholder.NewNullManager()),
	)
	if err != nil {
		panic(fmt.Sprintf("defaults: unreachable: %v", err))
	}
	return c
}
