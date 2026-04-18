package convresult

import (
	"github.com/seflue/adf-converter/converter"
)

// EnhancedConversionResultAnalyzer analyzes conversion results.
type EnhancedConversionResultAnalyzer struct{}

// NewEnhancedConversionResultAnalyzer creates a new result analyzer.
func NewEnhancedConversionResultAnalyzer() *EnhancedConversionResultAnalyzer {
	return &EnhancedConversionResultAnalyzer{}
}

// QualityMetrics provides detailed quality analysis.
type QualityMetrics struct {
	ContentLength     int
	AttributeCount    int
	WarningCount      int
	ConversionRatio   float64
	PreservationRatio float64
	QualityScore      float64
}

// ResultSummary provides aggregate statistics for multiple results.
type ResultSummary struct {
	TotalResults   int
	TotalWarnings  int
	TotalConverted int
	TotalPreserved int
	AverageQuality float64
	StrategyUsage  map[converter.ConversionStrategy]int
}

// AnalyzeQuality provides quality metrics for a conversion result.
func (cra *EnhancedConversionResultAnalyzer) AnalyzeQuality(result converter.EnhancedConversionResult) QualityMetrics {
	metrics := QualityMetrics{
		ContentLength:  len(result.Content),
		AttributeCount: len(result.PreservedAttrs),
		WarningCount:   len(result.Warnings),
	}

	totalElements := result.ElementsConverted + result.ElementsPreserved
	if totalElements > 0 {
		metrics.ConversionRatio = float64(result.ElementsConverted) / float64(totalElements)
		metrics.PreservationRatio = float64(result.ElementsPreserved) / float64(totalElements)
	}

	baseScore := 1.0

	warningPenalty := float64(len(result.Warnings)) * 0.1
	if warningPenalty > 0.5 {
		warningPenalty = 0.5
	}
	baseScore -= warningPenalty

	if len(result.PreservedAttrs) > 0 {
		baseScore += 0.1
	}

	if baseScore < 0.0 {
		baseScore = 0.0
	}
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	metrics.QualityScore = baseScore

	return metrics
}

// SummarizeResults provides a summary of multiple results.
func (cra *EnhancedConversionResultAnalyzer) SummarizeResults(results []converter.EnhancedConversionResult) ResultSummary {
	summary := ResultSummary{
		TotalResults:  len(results),
		StrategyUsage: make(map[converter.ConversionStrategy]int),
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
