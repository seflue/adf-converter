package converter

import (
	"fmt"
	"strings"
)

// EnhancedConversionResultBuilder helps build conversion results incrementally
type EnhancedConversionResultBuilder struct {
	content           strings.Builder
	preservedAttrs    map[string]interface{}
	strategy          ConversionStrategy
	warnings          []string
	elementsConverted int
	elementsPreserved int
}

// NewEnhancedConversionResultBuilder creates a new result builder
func NewEnhancedConversionResultBuilder(strategy ConversionStrategy) *EnhancedConversionResultBuilder {
	return &EnhancedConversionResultBuilder{
		preservedAttrs: make(map[string]interface{}),
		strategy:       strategy,
		warnings:       make([]string, 0),
	}
}

// AppendContent adds content to the result
func (crb *EnhancedConversionResultBuilder) AppendContent(content string) {
	crb.content.WriteString(content)
}

// AppendLine adds a line of content with newline
func (crb *EnhancedConversionResultBuilder) AppendLine(content string) {
	crb.content.WriteString(content)
	crb.content.WriteString("\n")
}

// PreserveAttribute adds an attribute to preserve
func (crb *EnhancedConversionResultBuilder) PreserveAttribute(key string, value interface{}) {
	crb.preservedAttrs[key] = value
}

// PreserveAttributes adds multiple attributes to preserve
func (crb *EnhancedConversionResultBuilder) PreserveAttributes(attrs map[string]interface{}) {
	for key, value := range attrs {
		crb.preservedAttrs[key] = value
	}
}

// AddWarning adds a warning message
func (crb *EnhancedConversionResultBuilder) AddWarning(message string) {
	crb.warnings = append(crb.warnings, message)
}

// AddWarningf adds a formatted warning message
func (crb *EnhancedConversionResultBuilder) AddWarningf(format string, args ...interface{}) {
	crb.warnings = append(crb.warnings, fmt.Sprintf(format, args...))
}

// IncrementConverted increments the converted elements counter
func (crb *EnhancedConversionResultBuilder) IncrementConverted() {
	crb.elementsConverted++
}

// IncrementPreserved increments the preserved elements counter
func (crb *EnhancedConversionResultBuilder) IncrementPreserved() {
	crb.elementsPreserved++
}

// AddConverted adds to the converted elements counter
func (crb *EnhancedConversionResultBuilder) AddConverted(count int) {
	crb.elementsConverted += count
}

// AddPreserved adds to the preserved elements counter
func (crb *EnhancedConversionResultBuilder) AddPreserved(count int) {
	crb.elementsPreserved += count
}

// SetStrategy updates the conversion strategy
func (crb *EnhancedConversionResultBuilder) SetStrategy(strategy ConversionStrategy) {
	crb.strategy = strategy
}

// Build creates the final conversion result
func (crb *EnhancedConversionResultBuilder) Build() EnhancedConversionResult {
	return EnhancedConversionResult{
		Content:           crb.content.String(),
		PreservedAttrs:    crb.preservedAttrs,
		Strategy:          crb.strategy,
		Warnings:          crb.warnings,
		ElementsConverted: crb.elementsConverted,
		ElementsPreserved: crb.elementsPreserved,
	}
}

// IsEmpty returns true if no content has been added
func (crb *EnhancedConversionResultBuilder) IsEmpty() bool {
	return crb.content.Len() == 0
}

// GetContentLength returns the current content length
func (crb *EnhancedConversionResultBuilder) GetContentLength() int {
	return crb.content.Len()
}

// HasWarnings returns true if there are any warnings
func (crb *EnhancedConversionResultBuilder) HasWarnings() bool {
	return len(crb.warnings) > 0
}

// GetWarningsCount returns the number of warnings
func (crb *EnhancedConversionResultBuilder) GetWarningsCount() int {
	return len(crb.warnings)
}

// CreateSuccessResult creates a successful conversion result
func CreateSuccessResult(content string, strategy ConversionStrategy) EnhancedConversionResult {
	return EnhancedConversionResult{
		Content:           content,
		PreservedAttrs:    make(map[string]interface{}),
		Strategy:          strategy,
		Warnings:          make([]string, 0),
		ElementsConverted: 1,
		ElementsPreserved: 0,
	}
}

// CreatePreservedResult creates a result for preserved content
func CreatePreservedResult(content string, attrs map[string]interface{}, strategy ConversionStrategy) EnhancedConversionResult {
	if attrs == nil {
		attrs = make(map[string]interface{})
	}

	return EnhancedConversionResult{
		Content:           content,
		PreservedAttrs:    attrs,
		Strategy:          strategy,
		Warnings:          make([]string, 0),
		ElementsConverted: 0,
		ElementsPreserved: 1,
	}
}

// CreateErrorResult creates a result for error scenarios
func CreateErrorResult(errorMsg string, strategy ConversionStrategy) EnhancedConversionResult {
	return EnhancedConversionResult{
		Content:           "",
		PreservedAttrs:    make(map[string]interface{}),
		Strategy:          strategy,
		Warnings:          []string{errorMsg},
		ElementsConverted: 0,
		ElementsPreserved: 0,
	}
}

// MergeResults combines multiple conversion results
func MergeResults(results ...EnhancedConversionResult) EnhancedConversionResult {
	if len(results) == 0 {
		return EnhancedConversionResult{
			PreservedAttrs: make(map[string]interface{}),
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

// EnhancedConversionResultAnalyzer provides analysis methods for conversion results
type EnhancedConversionResultAnalyzer struct{}

// NewEnhancedConversionResultAnalyzer creates a new result analyzer
func NewEnhancedConversionResultAnalyzer() *EnhancedConversionResultAnalyzer {
	return &EnhancedConversionResultAnalyzer{}
}

// AnalyzeQuality provides quality metrics for a conversion result
func (cra *EnhancedConversionResultAnalyzer) AnalyzeQuality(result EnhancedConversionResult) QualityMetrics {
	metrics := QualityMetrics{
		ContentLength:     len(result.Content),
		AttributeCount:    len(result.PreservedAttrs),
		WarningCount:      len(result.Warnings),
		ConversionRatio:   0.0,
		PreservationRatio: 0.0,
		QualityScore:      0.0,
	}

	totalElements := result.ElementsConverted + result.ElementsPreserved
	if totalElements > 0 {
		metrics.ConversionRatio = float64(result.ElementsConverted) / float64(totalElements)
		metrics.PreservationRatio = float64(result.ElementsPreserved) / float64(totalElements)
	}

	// Calculate quality score (0.0 to 1.0)
	baseScore := 1.0

	// Reduce score for warnings
	warningPenalty := float64(len(result.Warnings)) * 0.1
	if warningPenalty > 0.5 {
		warningPenalty = 0.5 // Cap at 50% penalty
	}
	baseScore -= warningPenalty

	// Bonus for preserved attributes (indicates completeness)
	if len(result.PreservedAttrs) > 0 {
		baseScore += 0.1
	}

	// Ensure score is within bounds
	if baseScore < 0.0 {
		baseScore = 0.0
	}
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	metrics.QualityScore = baseScore

	return metrics
}

// QualityMetrics provides detailed quality analysis
type QualityMetrics struct {
	ContentLength     int     // Length of generated content
	AttributeCount    int     // Number of preserved attributes
	WarningCount      int     // Number of warnings generated
	ConversionRatio   float64 // Ratio of converted to total elements
	PreservationRatio float64 // Ratio of preserved to total elements
	QualityScore      float64 // Overall quality score (0.0 to 1.0)
}

// SummarizeResults provides a summary of multiple results
func (cra *EnhancedConversionResultAnalyzer) SummarizeResults(results []EnhancedConversionResult) ResultSummary {
	summary := ResultSummary{
		TotalResults:   len(results),
		TotalWarnings:  0,
		TotalConverted: 0,
		TotalPreserved: 0,
		AverageQuality: 0.0,
		StrategyUsage:  make(map[ConversionStrategy]int),
	}

	if len(results) == 0 {
		return summary
	}

	var qualitySum float64

	for _, result := range results {
		summary.TotalWarnings += len(result.Warnings)
		summary.TotalConverted += result.ElementsConverted
		summary.TotalPreserved += result.ElementsPreserved

		metrics := cra.AnalyzeQuality(result)
		qualitySum += metrics.QualityScore

		summary.StrategyUsage[result.Strategy]++
	}

	summary.AverageQuality = qualitySum / float64(len(results))

	return summary
}

// ResultSummary provides aggregate statistics for multiple results
type ResultSummary struct {
	TotalResults   int                        // Number of results analyzed
	TotalWarnings  int                        // Total warnings across all results
	TotalConverted int                        // Total elements converted
	TotalPreserved int                        // Total elements preserved
	AverageQuality float64                    // Average quality score
	StrategyUsage  map[ConversionStrategy]int // Usage count per strategy
}
