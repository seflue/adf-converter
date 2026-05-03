// Package display renders an ADF document to a styled, terminal-ready
// string by piping adf-converter's display-mode Markdown through Glamour.
//
// The split exists for layering reasons. adf-converter stays a pure
// ADF↔Markdown library with no terminal concerns. This module composes
// adf-converter and Glamour without leaking ANSI escapes or styling
// decisions back into the lower layer.
package display

import (
	"fmt"

	"github.com/seflue/glamour/v2"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
)

// Render converts the ADF document to display-mode Markdown via
// adf-converter and pipes that Markdown through Glamour. Options control
// the Glamour stage only — the Markdown stage is fully determined by
// adf-converter's display-mode renderers.
func Render(doc *adf.Document, opts ...Option) (string, error) {
	if doc == nil {
		return "", fmt.Errorf("display: nil document")
	}

	cfg := defaultConfig()
	for _, apply := range opts {
		apply(&cfg)
	}

	conv := defaults.NewDisplayConverter()
	md, _, err := conv.ToMarkdown(*doc)
	if err != nil {
		return "", fmt.Errorf("display: render markdown: %w", err)
	}

	gOpts, err := cfg.glamourOptions()
	if err != nil {
		return "", err
	}
	r, err := glamour.NewTermRenderer(gOpts...)
	if err != nil {
		return "", fmt.Errorf("display: init glamour: %w", err)
	}
	rendered, err := r.Render(md)
	if err != nil {
		return "", fmt.Errorf("display: glamour render: %w", err)
	}
	return rendered, nil
}

// Option configures the Render pipeline via the functional-options pattern.
type Option func(*config)

// config holds resolved Render options. Zero value picks Glamour's "dark"
// style and an 80-column word-wrap. Glamour v2 dropped the v1 "auto"
// style; callers that want background-aware selection should pass an
// explicit style via WithStyle.
type config struct {
	style     string
	styleJSON string
	wordWrap  int
}

func defaultConfig() config {
	return config{
		style:    "dark",
		wordWrap: 80,
	}
}

// glamourOptions translates config into the option slice Glamour expects.
// styleJSON wins over style if both are set — caller asked for an explicit
// theme override and we honour it.
func (c config) glamourOptions() ([]glamour.TermRendererOption, error) {
	opts := []glamour.TermRendererOption{
		glamour.WithWordWrap(c.wordWrap),
		colorSpanCustomRenderer(),
		underlineCustomRenderer(),
	}
	if c.styleJSON != "" {
		opts = append(opts, glamour.WithStylesFromJSONBytes([]byte(c.styleJSON)))
		return opts, nil
	}
	opts = append(opts, glamour.WithStandardStyle(c.style))
	return opts, nil
}

// WithStyle selects a built-in Glamour style by name (e.g. "dark",
// "light", "dracula", "auto", "notty"). Unknown names fall back to
// Glamour's auto-detection at render time.
func WithStyle(name string) Option {
	return func(c *config) { c.style = name }
}

// WithStyleJSON loads a custom Glamour theme from a JSON document. When
// set, it takes precedence over WithStyle.
func WithStyleJSON(json string) Option {
	return func(c *config) { c.styleJSON = json }
}

// WithWordWrap sets Glamour's word-wrap column. A value <= 0 disables
// wrapping (Glamour treats it as "no wrap").
func WithWordWrap(n int) Option {
	return func(c *config) { c.wordWrap = n }
}
