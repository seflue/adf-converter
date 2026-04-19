package converter

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// deletionTracker tracks which placeholders are requested vs successfully restored during conversion
type deletionTracker struct {
	originalCount         int
	requestedPlaceholders map[string]bool
	restoredPlaceholders  map[string]bool
	session               *placeholder.EditSession
	manager               placeholder.Manager
}

// newDeletionTracker creates a new deletion tracker
func newDeletionTracker(session *placeholder.EditSession, manager placeholder.Manager) *deletionTracker {
	return &deletionTracker{
		originalCount:         len(session.Preserved),
		requestedPlaceholders: make(map[string]bool),
		restoredPlaceholders:  make(map[string]bool),
		session:               session,
		manager:               manager,
	}
}

// RecordPlaceholderRequest tracks that a placeholder was requested from the markdown
func (dt *deletionTracker) RecordPlaceholderRequest(placeholderID string) {
	dt.requestedPlaceholders[placeholderID] = true
}

// RecordPlaceholderRestored tracks that a placeholder was successfully restored
func (dt *deletionTracker) RecordPlaceholderRestored(placeholderID string) {
	dt.restoredPlaceholders[placeholderID] = true
}

// GetSummary generates the final deletion summary
func (dt *deletionTracker) GetSummary() *DeletionSummary {
	var deletedIDs []string

	// Find placeholders that were in the original session but never restored
	// This correctly identifies deletions regardless of whether they were requested from markdown
	for placeholderID := range dt.session.Preserved {
		if !dt.restoredPlaceholders[placeholderID] {
			deletedIDs = append(deletedIDs, placeholderID)
		}
	}

	preservedCount := len(dt.restoredPlaceholders)
	deletedCount := len(deletedIDs)

	return &DeletionSummary{
		DeletedPlaceholderIDs: deletedIDs,
		DeletedCount:          deletedCount,
		PreservedCount:        preservedCount,
		OriginalCount:         dt.originalCount,
	}
}

// FromMarkdown converts edited Markdown back to ADF, restoring preserved content from placeholders
func FromMarkdown(markdown string, session *placeholder.EditSession, manager placeholder.Manager, registry *ConverterRegistry) (adf_types.ADFDocument, error) {
	// PHASE 5: Comprehensive error handling with recovery
	defer func() {
		if r := recover(); r != nil {
			slog.Error("FromMarkdown: critical error recovered", "error", r, "markdownLength", len(markdown))
		}
	}()

	if session == nil {
		return adf_types.ADFDocument{}, fmt.Errorf("session cannot be nil")
	}

	// PHASE 5: Additional input validation
	if len(markdown) > 1000000 { // 1MB limit
		return adf_types.ADFDocument{}, fmt.Errorf("markdown input too large: %d bytes (max 1MB)", len(markdown))
	}

	// Split markdown into lines for processing
	lines := strings.Split(markdown, "\n")

	// PHASE 5: Validate line count
	if len(lines) > 10000 {
		slog.Warn("FromMarkdown: processing extremely large document", "lineCount", len(lines))
	}

	// Parse the markdown into ADF nodes
	parser := NewMarkdownParser(session, manager, registry)
	nodes, err := parser.ParseMarkdownToADFNodes(lines)
	if err != nil {
		return adf_types.ADFDocument{}, fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Create the ADF document
	doc := adf_types.ADFDocument{
		Version: 1,
		Type:    "doc",
		Content: nodes,
	}

	return doc, nil
}

