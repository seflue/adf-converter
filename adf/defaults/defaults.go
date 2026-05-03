// Package defaults wires the eager registration of all standard ADF element
// converters. Callers that want a ready-to-use converter without assembling
// the registry themselves should use NewDefaultConverter or NewDisplayConverter.
//
// This package is intentionally separate from adf/ to avoid an import
// cycle: the elements/ package depends on adf/, so adf/ cannot
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
		r.MustRegister(reg.NodeType, reg.Renderer)
	}
	for _, nodeType := range elements.StandardBlockParserOrder {
		r.MustRegisterBlockParser(nodeType)
	}

	return r
}

// NewDisplayRegistry builds a registry for read-only display-mode rendering.
// It copies the standard registry and overlays display-specific renderers
// for the node types that render badly through Glamour without adjustment
// (panel, mention, inlineCard, status, text). The output is plain Markdown;
// terminal styling (ANSI) lives in a separate display/ module that pipes
// this Markdown through Glamour.
func NewDisplayRegistry() *adf.ConverterRegistry {
	r := adf.NewConverterRegistry()

	for _, reg := range elements.StandardNodes() {
		r.MustRegister(reg.NodeType, reg.Renderer)
	}
	for _, nodeType := range elements.StandardBlockParserOrder {
		r.MustRegisterBlockParser(nodeType)
	}

	r.MustRegister("mention", elements.NewMentionDisplayRenderer())
	r.MustRegister("inlineCard", elements.NewInlineCardDisplayRenderer())
	r.MustRegister("panel", elements.NewPanelDisplayRenderer())
	r.MustRegister("status", elements.NewStatusDisplayRenderer())
	r.MustRegister("text", elements.NewTextDisplayRenderer())

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
// It wires NewDisplayRegistry together with a noop placeholder manager so
// preserved nodes render as preview text instead of placeholder comments.
func NewDisplayConverter() *adf.DefaultConverter {
	c, err := adf.NewConverter(
		adf.WithRegistry(NewDisplayRegistry()),
		adf.WithPlaceholderManager(placeholder.NewNoop()),
	)
	if err != nil {
		panic(fmt.Sprintf("defaults: unreachable: %v", err))
	}
	return c
}
