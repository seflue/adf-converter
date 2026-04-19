// Package convresult provides internal helpers for building, analyzing, and
// producing EnhancedConversionResult values. The result type itself lives in
// the public converter package; this package holds the non-public machinery.
package convresult

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/converter"
)

// EnhancedConversionResultBuilder builds conversion results incrementally.
type EnhancedConversionResultBuilder struct {
	content           strings.Builder
	preservedAttrs    map[string]any
	strategy          converter.ConversionStrategy
	warnings          []string
	elementsConverted int
	elementsPreserved int
}

// NewEnhancedConversionResultBuilder creates a new result builder.
func NewEnhancedConversionResultBuilder(strategy converter.ConversionStrategy) *EnhancedConversionResultBuilder {
	return &EnhancedConversionResultBuilder{
		preservedAttrs: make(map[string]any),
		strategy:       strategy,
		warnings:       make([]string, 0),
	}
}

func (crb *EnhancedConversionResultBuilder) AppendContent(content string) {
	crb.content.WriteString(content)
}

func (crb *EnhancedConversionResultBuilder) AppendLine(content string) {
	crb.content.WriteString(content)
	crb.content.WriteString("\n")
}

func (crb *EnhancedConversionResultBuilder) PreserveAttribute(key string, value any) {
	crb.preservedAttrs[key] = value
}

func (crb *EnhancedConversionResultBuilder) PreserveAttributes(attrs map[string]any) {
	for key, value := range attrs {
		crb.preservedAttrs[key] = value
	}
}

func (crb *EnhancedConversionResultBuilder) AddWarning(message string) {
	crb.warnings = append(crb.warnings, message)
}

func (crb *EnhancedConversionResultBuilder) AddWarningf(format string, args ...any) {
	crb.warnings = append(crb.warnings, fmt.Sprintf(format, args...))
}

func (crb *EnhancedConversionResultBuilder) IncrementConverted() {
	crb.elementsConverted++
}

func (crb *EnhancedConversionResultBuilder) IncrementPreserved() {
	crb.elementsPreserved++
}

func (crb *EnhancedConversionResultBuilder) AddConverted(count int) {
	crb.elementsConverted += count
}

func (crb *EnhancedConversionResultBuilder) AddPreserved(count int) {
	crb.elementsPreserved += count
}

func (crb *EnhancedConversionResultBuilder) SetStrategy(strategy converter.ConversionStrategy) {
	crb.strategy = strategy
}

func (crb *EnhancedConversionResultBuilder) Build() converter.EnhancedConversionResult {
	return converter.EnhancedConversionResult{
		Content:           crb.content.String(),
		PreservedAttrs:    crb.preservedAttrs,
		Strategy:          crb.strategy,
		Warnings:          crb.warnings,
		ElementsConverted: crb.elementsConverted,
		ElementsPreserved: crb.elementsPreserved,
	}
}

func (crb *EnhancedConversionResultBuilder) IsEmpty() bool {
	return crb.content.Len() == 0
}

func (crb *EnhancedConversionResultBuilder) GetContentLength() int {
	return crb.content.Len()
}

func (crb *EnhancedConversionResultBuilder) HasWarnings() bool {
	return len(crb.warnings) > 0
}

func (crb *EnhancedConversionResultBuilder) GetWarningsCount() int {
	return len(crb.warnings)
}

// CreateSuccessResult creates a successful conversion result.
func CreateSuccessResult(content string, strategy converter.ConversionStrategy) converter.EnhancedConversionResult {
	return converter.EnhancedConversionResult{
		Content:           content,
		PreservedAttrs:    make(map[string]any),
		Strategy:          strategy,
		Warnings:          make([]string, 0),
		ElementsConverted: 1,
		ElementsPreserved: 0,
	}
}

// CreatePreservedResult creates a result for preserved content.
func CreatePreservedResult(content string, attrs map[string]any, strategy converter.ConversionStrategy) converter.EnhancedConversionResult {
	if attrs == nil {
		attrs = make(map[string]any)
	}

	return converter.EnhancedConversionResult{
		Content:           content,
		PreservedAttrs:    attrs,
		Strategy:          strategy,
		Warnings:          make([]string, 0),
		ElementsConverted: 0,
		ElementsPreserved: 1,
	}
}

// CreateErrorResult creates a result for error scenarios.
func CreateErrorResult(errorMsg string, strategy converter.ConversionStrategy) converter.EnhancedConversionResult {
	return converter.EnhancedConversionResult{
		Content:           "",
		PreservedAttrs:    make(map[string]any),
		Strategy:          strategy,
		Warnings:          []string{errorMsg},
		ElementsConverted: 0,
		ElementsPreserved: 0,
	}
}

// MergeResults combines multiple conversion results.
func MergeResults(results ...converter.EnhancedConversionResult) converter.EnhancedConversionResult {
	if len(results) == 0 {
		return converter.EnhancedConversionResult{
			PreservedAttrs: make(map[string]any),
			Warnings:       make([]string, 0),
		}
	}

	if len(results) == 1 {
		return results[0]
	}

	builder := NewEnhancedConversionResultBuilder(results[0].Strategy)

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
