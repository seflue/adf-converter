package converter

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ValidationMetricsTracker provides comprehensive tracking of conversion validation metrics
type ValidationMetricsTracker interface {
	// RecordConversion records a conversion attempt
	RecordConversion(nodeType string, strategy ConversionStrategy, success bool)

	// RecordAttributeLoss records when attributes are lost during conversion
	RecordAttributeLoss(nodeType string, lostAttributes []string)

	// RecordContentModification records when content is modified during conversion
	RecordContentModification(nodeType string, originalLength, newLength int)

	// RecordValidationTime records how long validation took
	RecordValidationTime(nodeType string, duration time.Duration)

	// GetOverallMetrics returns overall validation metrics
	GetOverallMetrics() ValidationMetrics

	// GetMetricsByNodeType returns metrics broken down by node type
	GetMetricsByNodeType() map[string]NodeTypeMetrics

	// GetMetricsByStrategy returns metrics broken down by conversion strategy
	GetMetricsByStrategy() map[ConversionStrategy]StrategyMetrics

	// Reset clears all metrics
	Reset()

	// GenerateReport generates a human-readable metrics report
	GenerateReport() string
}

// NodeTypeMetrics tracks metrics for a specific ADF node type
type NodeTypeMetrics struct {
	NodeType         string
	TotalConversions int
	SuccessfulRounds int
	AttributesLost   int
	ContentModified  int
	AverageTime      time.Duration
	MaxTime          time.Duration
	MinTime          time.Duration
	FidelityScore    float64
}

// StrategyMetrics tracks metrics for a specific conversion strategy
type StrategyMetrics struct {
	Strategy           ConversionStrategy
	TotalConversions   int
	SuccessfulRounds   int
	AttributesLost     int
	ContentModified    int
	AverageTime        time.Duration
	SupportedNodeTypes []string
	FidelityScore      float64
}

// DetailedValidationMetrics extends ValidationMetrics with more granular tracking
type DetailedValidationMetrics struct {
	ValidationMetrics // Embed base metrics

	// Time-based metrics
	TotalValidationTime   time.Duration
	AverageValidationTime time.Duration
	MaxValidationTime     time.Duration
	MinValidationTime     time.Duration

	// Error tracking
	ErrorsByType     map[string]int
	MostCommonErrors []string

	// Node type breakdown
	NodeTypeStats map[string]NodeTypeMetrics

	// Strategy breakdown
	StrategyStats map[ConversionStrategy]StrategyMetrics

	// Performance metrics
	ConversionsPerSecond float64
	MemoryUsage          int64 // bytes

	// Quality metrics
	AttributePreservationRate float64
	ContentPreservationRate   float64
	StructuralFidelityRate    float64
}

// DefaultValidationMetricsTracker implements ValidationMetricsTracker
type DefaultValidationMetricsTracker struct {
	metrics         DetailedValidationMetrics
	nodeMetrics     map[string]*NodeTypeMetrics
	strategyMetrics map[ConversionStrategy]*StrategyMetrics
	timings         []time.Duration
	mutex           sync.RWMutex
	startTime       time.Time
}

// NewDefaultValidationMetricsTracker creates a new metrics tracker
func NewDefaultValidationMetricsTracker() *DefaultValidationMetricsTracker {
	return &DefaultValidationMetricsTracker{
		metrics: DetailedValidationMetrics{
			ValidationMetrics: ValidationMetrics{
				FidelityScore: 1.0,
			},
			ErrorsByType:  make(map[string]int),
			NodeTypeStats: make(map[string]NodeTypeMetrics),
			StrategyStats: make(map[ConversionStrategy]StrategyMetrics),
		},
		nodeMetrics:     make(map[string]*NodeTypeMetrics),
		strategyMetrics: make(map[ConversionStrategy]*StrategyMetrics),
		timings:         make([]time.Duration, 0),
		startTime:       time.Now(),
	}
}

// RecordConversion records a conversion attempt
func (dvmt *DefaultValidationMetricsTracker) RecordConversion(nodeType string, strategy ConversionStrategy, success bool) {
	dvmt.mutex.Lock()
	defer dvmt.mutex.Unlock()

	// Update overall metrics
	dvmt.metrics.TotalConversions++
	if success {
		dvmt.metrics.SuccessfulRounds++
	}

	// Update node type metrics
	if nodeMetrics, exists := dvmt.nodeMetrics[nodeType]; exists {
		nodeMetrics.TotalConversions++
		if success {
			nodeMetrics.SuccessfulRounds++
		}
	} else {
		dvmt.nodeMetrics[nodeType] = &NodeTypeMetrics{
			NodeType:         nodeType,
			TotalConversions: 1,
			SuccessfulRounds: 0,
			MinTime:          time.Hour, // Initialize to high value
		}
		if success {
			dvmt.nodeMetrics[nodeType].SuccessfulRounds = 1
		}
	}

	// Update strategy metrics
	if strategyMetrics, exists := dvmt.strategyMetrics[strategy]; exists {
		strategyMetrics.TotalConversions++
		if success {
			strategyMetrics.SuccessfulRounds++
		}
	} else {
		dvmt.strategyMetrics[strategy] = &StrategyMetrics{
			Strategy:           strategy,
			TotalConversions:   1,
			SuccessfulRounds:   0,
			SupportedNodeTypes: []string{nodeType},
		}
		if success {
			dvmt.strategyMetrics[strategy].SuccessfulRounds = 1
		}
	}

	// Update conversion rate
	dvmt.updateConversionRate()

	// Recalculate fidelity scores
	dvmt.recalculateFidelityScores()
}

// RecordAttributeLoss records when attributes are lost during conversion
func (dvmt *DefaultValidationMetricsTracker) RecordAttributeLoss(nodeType string, lostAttributes []string) {
	dvmt.mutex.Lock()
	defer dvmt.mutex.Unlock()

	attributeCount := len(lostAttributes)
	dvmt.metrics.AttributesLost += attributeCount

	// Update node-specific metrics
	if nodeMetrics, exists := dvmt.nodeMetrics[nodeType]; exists {
		nodeMetrics.AttributesLost += attributeCount
	}

	// Update error tracking
	errorType := fmt.Sprintf("attribute_loss_%s", nodeType)
	dvmt.metrics.ErrorsByType[errorType] += attributeCount

	dvmt.recalculateFidelityScores()
}

// RecordContentModification records when content is modified during conversion
func (dvmt *DefaultValidationMetricsTracker) RecordContentModification(nodeType string, originalLength, newLength int) {
	dvmt.mutex.Lock()
	defer dvmt.mutex.Unlock()

	dvmt.metrics.ContentModified++

	// Update node-specific metrics
	if nodeMetrics, exists := dvmt.nodeMetrics[nodeType]; exists {
		nodeMetrics.ContentModified++
	}

	// Track the degree of modification
	if originalLength > 0 {
		modificationRatio := float64(abs(originalLength-newLength)) / float64(originalLength)
		if modificationRatio > 0.1 { // More than 10% change
			errorType := fmt.Sprintf("significant_content_modification_%s", nodeType)
			dvmt.metrics.ErrorsByType[errorType]++
		}
	}

	dvmt.recalculateFidelityScores()
}

// RecordValidationTime records how long validation took
func (dvmt *DefaultValidationMetricsTracker) RecordValidationTime(nodeType string, duration time.Duration) {
	dvmt.mutex.Lock()
	defer dvmt.mutex.Unlock()

	// Update overall timing metrics
	dvmt.timings = append(dvmt.timings, duration)
	dvmt.metrics.TotalValidationTime += duration

	if duration > dvmt.metrics.MaxValidationTime {
		dvmt.metrics.MaxValidationTime = duration
	}

	if dvmt.metrics.MinValidationTime == 0 || duration < dvmt.metrics.MinValidationTime {
		dvmt.metrics.MinValidationTime = duration
	}

	// Update average
	if len(dvmt.timings) > 0 {
		total := time.Duration(0)
		for _, t := range dvmt.timings {
			total += t
		}
		dvmt.metrics.AverageValidationTime = total / time.Duration(len(dvmt.timings))
	}

	// Update node-specific timing
	if nodeMetrics, exists := dvmt.nodeMetrics[nodeType]; exists {
		// Simple average calculation for node metrics
		if nodeMetrics.MinTime == time.Hour || duration < nodeMetrics.MinTime {
			nodeMetrics.MinTime = duration
		}
		if duration > nodeMetrics.MaxTime {
			nodeMetrics.MaxTime = duration
		}
		// Update average (simplified)
		nodeMetrics.AverageTime = (nodeMetrics.AverageTime + duration) / 2
	}

	dvmt.updateConversionRate()
}

// GetOverallMetrics returns overall validation metrics
func (dvmt *DefaultValidationMetricsTracker) GetOverallMetrics() ValidationMetrics {
	dvmt.mutex.RLock()
	defer dvmt.mutex.RUnlock()

	return dvmt.metrics.ValidationMetrics
}

// GetMetricsByNodeType returns metrics broken down by node type
func (dvmt *DefaultValidationMetricsTracker) GetMetricsByNodeType() map[string]NodeTypeMetrics {
	dvmt.mutex.RLock()
	defer dvmt.mutex.RUnlock()

	result := make(map[string]NodeTypeMetrics)
	for nodeType, metrics := range dvmt.nodeMetrics {
		// Create a copy to prevent external modification
		result[nodeType] = *metrics
	}

	return result
}

// GetMetricsByStrategy returns metrics broken down by conversion strategy
func (dvmt *DefaultValidationMetricsTracker) GetMetricsByStrategy() map[ConversionStrategy]StrategyMetrics {
	dvmt.mutex.RLock()
	defer dvmt.mutex.RUnlock()

	result := make(map[ConversionStrategy]StrategyMetrics)
	for strategy, metrics := range dvmt.strategyMetrics {
		// Create a copy to prevent external modification
		result[strategy] = *metrics
	}

	return result
}

// Reset clears all metrics
func (dvmt *DefaultValidationMetricsTracker) Reset() {
	dvmt.mutex.Lock()
	defer dvmt.mutex.Unlock()

	dvmt.metrics = DetailedValidationMetrics{
		ValidationMetrics: ValidationMetrics{
			FidelityScore: 1.0,
		},
		ErrorsByType:  make(map[string]int),
		NodeTypeStats: make(map[string]NodeTypeMetrics),
		StrategyStats: make(map[ConversionStrategy]StrategyMetrics),
	}
	dvmt.nodeMetrics = make(map[string]*NodeTypeMetrics)
	dvmt.strategyMetrics = make(map[ConversionStrategy]*StrategyMetrics)
	dvmt.timings = make([]time.Duration, 0)
	dvmt.startTime = time.Now()
}

// GenerateReport generates a human-readable metrics report
func (dvmt *DefaultValidationMetricsTracker) GenerateReport() string {
	dvmt.mutex.RLock()
	defer dvmt.mutex.RUnlock()

	var report strings.Builder

	report.WriteString("=== Validation Metrics Report ===\n\n")

	// Overall metrics
	report.WriteString("Overall Metrics:\n")
	fmt.Fprintf(&report, "  Total Conversions: %d\n", dvmt.metrics.TotalConversions)
	fmt.Fprintf(&report, "  Successful Rounds: %d\n", dvmt.metrics.SuccessfulRounds)
	fmt.Fprintf(&report, "  Success Rate: %.2f%%\n", dvmt.getSuccessRate())
	fmt.Fprintf(&report, "  Fidelity Score: %.3f\n", dvmt.metrics.FidelityScore)
	fmt.Fprintf(&report, "  Attributes Lost: %d\n", dvmt.metrics.AttributesLost)
	fmt.Fprintf(&report, "  Content Modified: %d\n", dvmt.metrics.ContentModified)
	fmt.Fprintf(&report, "  Conversions/sec: %.2f\n", dvmt.metrics.ConversionsPerSecond)

	// Timing metrics
	report.WriteString("\nTiming Metrics:\n")
	fmt.Fprintf(&report, "  Average Time: %v\n", dvmt.metrics.AverageValidationTime)
	fmt.Fprintf(&report, "  Min Time: %v\n", dvmt.metrics.MinValidationTime)
	fmt.Fprintf(&report, "  Max Time: %v\n", dvmt.metrics.MaxValidationTime)
	fmt.Fprintf(&report, "  Total Time: %v\n", dvmt.metrics.TotalValidationTime)

	// Node type breakdown
	report.WriteString("\nNode Type Breakdown:\n")
	for nodeType, metrics := range dvmt.nodeMetrics {
		successRate := 0.0
		if metrics.TotalConversions > 0 {
			successRate = float64(metrics.SuccessfulRounds) / float64(metrics.TotalConversions) * 100
		}
		fmt.Fprintf(&report, "  %s: %d conversions, %.1f%% success, %d attr lost, %d content modified\n",
			nodeType, metrics.TotalConversions, successRate, metrics.AttributesLost, metrics.ContentModified)
	}

	// Strategy breakdown
	report.WriteString("\nStrategy Breakdown:\n")
	for strategy, metrics := range dvmt.strategyMetrics {
		successRate := 0.0
		if metrics.TotalConversions > 0 {
			successRate = float64(metrics.SuccessfulRounds) / float64(metrics.TotalConversions) * 100
		}
		fmt.Fprintf(&report, "  %s: %d conversions, %.1f%% success\n",
			strategy.String(), metrics.TotalConversions, successRate)
	}

	// Top errors
	report.WriteString("\nMost Common Errors:\n")
	for errorType, count := range dvmt.metrics.ErrorsByType {
		if count > 0 {
			fmt.Fprintf(&report, "  %s: %d occurrences\n", errorType, count)
		}
	}

	return report.String()
}

// Helper functions

func (dvmt *DefaultValidationMetricsTracker) updateConversionRate() {
	elapsed := time.Since(dvmt.startTime)
	if elapsed.Seconds() > 0 {
		dvmt.metrics.ConversionsPerSecond = float64(dvmt.metrics.TotalConversions) / elapsed.Seconds()
	}
}

func (dvmt *DefaultValidationMetricsTracker) recalculateFidelityScores() {
	if dvmt.metrics.TotalConversions == 0 {
		dvmt.metrics.FidelityScore = 1.0
		return
	}

	successRate := float64(dvmt.metrics.SuccessfulRounds) / float64(dvmt.metrics.TotalConversions)
	attributePreservationRate := 1.0 - (float64(dvmt.metrics.AttributesLost) / float64(dvmt.metrics.TotalConversions))
	contentPreservationRate := 1.0 - (float64(dvmt.metrics.ContentModified) / float64(dvmt.metrics.TotalConversions))

	// Weighted fidelity score
	dvmt.metrics.FidelityScore = successRate*0.5 + attributePreservationRate*0.3 + contentPreservationRate*0.2

	// Update preservation rates
	dvmt.metrics.AttributePreservationRate = attributePreservationRate
	dvmt.metrics.ContentPreservationRate = contentPreservationRate
	dvmt.metrics.StructuralFidelityRate = successRate

	// Update node-specific fidelity scores
	for _, nodeMetrics := range dvmt.nodeMetrics {
		if nodeMetrics.TotalConversions > 0 {
			nodeSuccessRate := float64(nodeMetrics.SuccessfulRounds) / float64(nodeMetrics.TotalConversions)
			nodeAttrRate := 1.0 - (float64(nodeMetrics.AttributesLost) / float64(nodeMetrics.TotalConversions))
			nodeContentRate := 1.0 - (float64(nodeMetrics.ContentModified) / float64(nodeMetrics.TotalConversions))
			nodeMetrics.FidelityScore = nodeSuccessRate*0.5 + nodeAttrRate*0.3 + nodeContentRate*0.2
		}
	}

	// Update strategy-specific fidelity scores
	for _, strategyMetrics := range dvmt.strategyMetrics {
		if strategyMetrics.TotalConversions > 0 {
			strategySuccessRate := float64(strategyMetrics.SuccessfulRounds) / float64(strategyMetrics.TotalConversions)
			strategyAttrRate := 1.0 - (float64(strategyMetrics.AttributesLost) / float64(strategyMetrics.TotalConversions))
			strategyContentRate := 1.0 - (float64(strategyMetrics.ContentModified) / float64(strategyMetrics.TotalConversions))
			strategyMetrics.FidelityScore = strategySuccessRate*0.5 + strategyAttrRate*0.3 + strategyContentRate*0.2
		}
	}
}

func (dvmt *DefaultValidationMetricsTracker) getSuccessRate() float64 {
	if dvmt.metrics.TotalConversions == 0 {
		return 0.0
	}
	return float64(dvmt.metrics.SuccessfulRounds) / float64(dvmt.metrics.TotalConversions) * 100
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
