package converter

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// DeletionSummary tracks which placeholders were deleted during markdown → ADF conversion
type DeletionSummary struct {
	// DeletedPlaceholderIDs contains the IDs of placeholders that were deleted by the user
	DeletedPlaceholderIDs []string

	// DeletedCount is the number of placeholders that were deleted
	DeletedCount int

	// PreservedCount is the number of placeholders that were successfully restored
	PreservedCount int

	// OriginalCount is the total number of placeholders that were originally stored
	OriginalCount int
}

// ConversionResult contains both the converted document and deletion tracking information
type ConversionResult struct {
	// Document is the converted ADF document
	Document adf_types.ADFDocument

	// Deletions contains information about which placeholders were deleted during conversion
	Deletions *DeletionSummary
}

// HasDeletions returns true if any placeholders were deleted during conversion
func (ds *DeletionSummary) HasDeletions() bool {
	return ds.DeletedCount > 0
}

// GetPreservationRatio returns the ratio of preserved to original placeholders (0.0 to 1.0)
func (ds *DeletionSummary) GetPreservationRatio() float64 {
	if ds.OriginalCount == 0 {
		return 1.0
	}
	return float64(ds.PreservedCount) / float64(ds.OriginalCount)
}

// FormatUserMessage returns a user-friendly message about the conversion results
func (ds *DeletionSummary) FormatUserMessage() string {
	if ds.OriginalCount == 0 {
		return "No complex formatting blocks to preserve"
	}

	if ds.DeletedCount == 0 {
		return "All complex formatting preserved"
	}

	if ds.PreservedCount == 0 {
		return fmt.Sprintf("All %d complex formatting block(s) deleted by user", ds.OriginalCount)
	}

	return fmt.Sprintf("%d of %d complex formatting preserved (%d deleted by user)",
		ds.PreservedCount, ds.OriginalCount, ds.DeletedCount)
}

// Converter provides bidirectional ADF ↔ Markdown conversion
type Converter interface {
	// ToMarkdown converts an ADF document to editable Markdown with placeholders for complex content
	ToMarkdown(doc adf_types.ADFDocument) (markdown string, session *placeholder.EditSession, err error)

	// FromMarkdown converts edited Markdown back to ADF, restoring preserved content from placeholders
	FromMarkdown(markdown string, session *placeholder.EditSession) (ConversionResult, error)
}

// DefaultConverter uses the classifier and placeholder manager for conversion
type DefaultConverter struct {
	classifier ContentClassifier
	manager    placeholder.Manager
}

// NewConverter creates a new DefaultConverter with the provided classifier and manager
func NewConverter(classifier ContentClassifier, manager placeholder.Manager) Converter {
	return &DefaultConverter{
		classifier: classifier,
		manager:    manager,
	}
}

// NewDefaultConverter creates a DefaultConverter with default implementations
func NewDefaultConverter() Converter {
	return &DefaultConverter{
		classifier: NewDefaultClassifier(),
		manager:    placeholder.NewManager(),
	}
}

// NewDisplayConverter creates a converter for read-only display mode.
// Uses a NullManager that produces preview text instead of placeholder comments.
// FromMarkdown is still available but not useful in display context.
func NewDisplayConverter() Converter {
	return &DefaultConverter{
		classifier: NewDefaultClassifier(),
		manager:    placeholder.NewNullManager(),
	}
}

// ToMarkdown converts an ADF document to editable Markdown with placeholders for complex content
func (c *DefaultConverter) ToMarkdown(doc adf_types.ADFDocument) (string, *placeholder.EditSession, error) {
	return ToMarkdown(doc, c.classifier, c.manager)
}

// FromMarkdown converts edited Markdown back to ADF with deletion tracking
func (c *DefaultConverter) FromMarkdown(markdown string, session *placeholder.EditSession) (ConversionResult, error) {
	if session == nil {
		return ConversionResult{}, fmt.Errorf("session cannot be nil")
	}

	deletionTracker := newDeletionTracker(session, c.manager)

	lines := strings.Split(markdown, "\n")

	parser := NewMarkdownParser(session, c.manager)
	nodes, err := parser.ParseMarkdownToADFNodes(lines)
	if err != nil {
		return ConversionResult{}, fmt.Errorf("failed to parse markdown: %w", err)
	}

	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: nodes,
	}

	return ConversionResult{
		Document:  doc,
		Deletions: deletionTracker.GetSummary(),
	}, nil
}

// GetClassifier returns the content classifier used by this converter
func (c *DefaultConverter) GetClassifier() ContentClassifier {
	return c.classifier
}

// GetManager returns the placeholder manager used by this converter
func (c *DefaultConverter) GetManager() placeholder.Manager {
	return c.manager
}
