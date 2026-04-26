package adf

import (
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
