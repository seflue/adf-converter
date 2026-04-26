package adf

import (
	"fmt"
	"strings"

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
	ToMarkdown(doc Document) (markdown string, session *placeholder.EditSession, err error)

	// FromMarkdown converts edited Markdown back to ADF, restoring preserved content from placeholders.
	// Returns the converted document, a deletion summary tracking which placeholders the user removed,
	// and any conversion error.
	FromMarkdown(markdown string, session *placeholder.EditSession) (Document, *DeletionSummary, error)
}

// DefaultConverter uses the classifier and placeholder manager for conversion.
// Each instance holds its own registry — there is no global registry.
type DefaultConverter struct {
	classifier ContentClassifier
	manager    placeholder.Manager
	registry   *ConverterRegistry
}

// Option configures a DefaultConverter at construction time.
type Option func(*DefaultConverter)

// WithRegistry overrides the converter registry used for element dispatch.
func WithRegistry(r *ConverterRegistry) Option {
	return func(c *DefaultConverter) {
		c.registry = r
	}
}

// WithClassifier overrides the content classifier.
func WithClassifier(cl ContentClassifier) Option {
	return func(c *DefaultConverter) {
		c.classifier = cl
	}
}

// WithPlaceholderManager overrides the placeholder manager.
func WithPlaceholderManager(m placeholder.Manager) Option {
	return func(c *DefaultConverter) {
		c.manager = m
	}
}

// NewConverter creates a new DefaultConverter.
//
// Defaults: DefaultClassifier, placeholder.NewManager, empty registry.
// Use WithRegistry to supply a populated registry (see adf/defaults).
//
// Returns an error if the registry is empty after applying options, because a
// converter without registered element converters cannot process any content
// and every FromMarkdown/ToMarkdown call would fail at dispatch time (ac-0115).
func NewConverter(opts ...Option) (*DefaultConverter, error) {
	c := &DefaultConverter{
		classifier: NewDefaultClassifier(),
		manager:    placeholder.NewManager(),
		registry:   NewConverterRegistry(),
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.registry == nil || c.registry.Count() == 0 {
		return nil, fmt.Errorf("converter: empty registry; use defaults.NewDefaultConverter() or pass WithRegistry(...)")
	}
	return c, nil
}

// ToMarkdown converts an ADF document to editable Markdown with placeholders for complex content
func (c *DefaultConverter) ToMarkdown(doc Document) (string, *placeholder.EditSession, error) {
	return toMarkdown(doc, c.classifier, c.manager, c.registry)
}

// FromMarkdown converts edited Markdown back to ADF with deletion tracking.
func (c *DefaultConverter) FromMarkdown(markdown string, session *placeholder.EditSession) (Document, *DeletionSummary, error) {
	if session == nil {
		return Document{}, nil, fmt.Errorf("session cannot be nil")
	}

	deletionTracker := newDeletionTracker(session, c.manager)

	lines := strings.Split(markdown, "\n")

	parser := NewMarkdownParser(session, c.manager, c.registry)
	nodes, err := parser.ParseMarkdownToADFNodes(lines)
	if err != nil {
		return Document{}, nil, fmt.Errorf("failed to parse markdown: %w", err)
	}

	doc := Document{
		Version: 1,
		Type:    "doc",
		Content: nodes,
	}

	return doc, deletionTracker.GetSummary(), nil
}

