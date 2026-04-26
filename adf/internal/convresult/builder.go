// Package convresult provides internal helpers for building, analyzing, and
// producing RenderResult values. The result type itself lives in
// the public converter package; this package holds the non-public machinery.
package convresult

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf"
)

// RenderResultBuilder builds conversion results incrementally.
type RenderResultBuilder struct {
	content           strings.Builder
	preservedAttrs    map[string]any
	strategy          adf.ConversionStrategy
	warnings          []string
	elementsConverted int
	elementsPreserved int
}

// NewRenderResultBuilder creates a new result builder.
func NewRenderResultBuilder(strategy adf.ConversionStrategy) *RenderResultBuilder {
	return &RenderResultBuilder{
		preservedAttrs: make(map[string]any),
		strategy:       strategy,
		warnings:       make([]string, 0),
	}
}

func (crb *RenderResultBuilder) AppendContent(content string) {
	crb.content.WriteString(content)
}

func (crb *RenderResultBuilder) AppendLine(content string) {
	crb.content.WriteString(content)
	crb.content.WriteString("\n")
}

func (crb *RenderResultBuilder) PreserveAttribute(key string, value any) {
	crb.preservedAttrs[key] = value
}

func (crb *RenderResultBuilder) PreserveAttributes(attrs map[string]any) {
	for key, value := range attrs {
		crb.preservedAttrs[key] = value
	}
}

func (crb *RenderResultBuilder) AddWarning(message string) {
	crb.warnings = append(crb.warnings, message)
}

func (crb *RenderResultBuilder) AddWarningf(format string, args ...any) {
	crb.warnings = append(crb.warnings, fmt.Sprintf(format, args...))
}

func (crb *RenderResultBuilder) IncrementConverted() {
	crb.elementsConverted++
}

func (crb *RenderResultBuilder) IncrementPreserved() {
	crb.elementsPreserved++
}

func (crb *RenderResultBuilder) AddConverted(count int) {
	crb.elementsConverted += count
}

func (crb *RenderResultBuilder) AddPreserved(count int) {
	crb.elementsPreserved += count
}

func (crb *RenderResultBuilder) SetStrategy(strategy adf.ConversionStrategy) {
	crb.strategy = strategy
}

func (crb *RenderResultBuilder) Build() adf.RenderResult {
	return adf.RenderResult{
		Content:           crb.content.String(),
		PreservedAttrs:    crb.preservedAttrs,
		Strategy:          crb.strategy,
		Warnings:          crb.warnings,
		ElementsConverted: crb.elementsConverted,
		ElementsPreserved: crb.elementsPreserved,
	}
}

func (crb *RenderResultBuilder) IsEmpty() bool {
	return crb.content.Len() == 0
}

func (crb *RenderResultBuilder) GetContentLength() int {
	return crb.content.Len()
}

func (crb *RenderResultBuilder) HasWarnings() bool {
	return len(crb.warnings) > 0
}

func (crb *RenderResultBuilder) GetWarningsCount() int {
	return len(crb.warnings)
}

// CreateSuccessResult creates a successful conversion result.
func CreateSuccessResult(content string, strategy adf.ConversionStrategy) adf.RenderResult {
	return adf.RenderResult{
		Content:           content,
		PreservedAttrs:    make(map[string]any),
		Strategy:          strategy,
		Warnings:          make([]string, 0),
		ElementsConverted: 1,
		ElementsPreserved: 0,
	}
}

// CreatePreservedResult creates a result for preserved content.
func CreatePreservedResult(content string, attrs map[string]any, strategy adf.ConversionStrategy) adf.RenderResult {
	if attrs == nil {
		attrs = make(map[string]any)
	}

	return adf.RenderResult{
		Content:           content,
		PreservedAttrs:    attrs,
		Strategy:          strategy,
		Warnings:          make([]string, 0),
		ElementsConverted: 0,
		ElementsPreserved: 1,
	}
}

// CreateErrorResult creates a result for error scenarios.
func CreateErrorResult(errorMsg string, strategy adf.ConversionStrategy) adf.RenderResult {
	return adf.RenderResult{
		Content:           "",
		PreservedAttrs:    make(map[string]any),
		Strategy:          strategy,
		Warnings:          []string{errorMsg},
		ElementsConverted: 0,
		ElementsPreserved: 0,
	}
}

// MergeResults combines multiple conversion results.
func MergeResults(results ...adf.RenderResult) adf.RenderResult {
	if len(results) == 0 {
		return adf.RenderResult{
			PreservedAttrs: make(map[string]any),
			Warnings:       make([]string, 0),
		}
	}

	if len(results) == 1 {
		return results[0]
	}

	builder := NewRenderResultBuilder(results[0].Strategy)

	for _, result := range results {
		builder.AppendContent(result.Content)
		builder.PreserveAttributes(result.PreservedAttrs)
		for _, warning := range result.Warnings {
			builder.AddWarning(warning)
		}
		builder.AddConverted(result.ElementsConverted)
		builder.AddPreserved(result.ElementsPreserved)
	}

	return builder.Build()
}
