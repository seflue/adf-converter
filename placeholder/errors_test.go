package placeholder

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultManager_Restore_UnknownReturnsSentinel verifies that the
// DefaultManager wraps its "not found" error with ErrPlaceholderNotFound
// so callers can distinguish legitimate user-deletions from real bugs.
func TestDefaultManager_Restore_UnknownReturnsSentinel(t *testing.T) {
	m := NewManager()

	_, err := m.Restore("ADF_PLACEHOLDER_DOES_NOT_EXIST")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPlaceholderNotFound),
		"expected errors.Is(err, ErrPlaceholderNotFound), got %v", err)
}

// TestNullManager_Restore_ReturnsSentinel verifies that the noop manager
// always reports "not found" via the sentinel — it never stores anything.
// This lets display-mode callers treat any restore call as a deletion
// using the same errors.Is check as production callers.
func TestNullManager_Restore_ReturnsSentinel(t *testing.T) {
	m := NewNoop()

	_, err := m.Restore("ADF_PLACEHOLDER_001")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPlaceholderNotFound),
		"expected errors.Is(err, ErrPlaceholderNotFound), got %v", err)
}
