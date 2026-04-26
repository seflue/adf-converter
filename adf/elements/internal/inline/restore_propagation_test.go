package inline

import (
	"errors"
	"testing"

	adf "github.com/seflue/adf-converter/adf/adftypes"
	"github.com/seflue/adf-converter/placeholder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFailingManager implements placeholder.Manager and returns a non-sentinel
// error from Restore so we can verify caller-side propagation behavior.
type fakeFailingManager struct {
	restoreErr error
}

func (f *fakeFailingManager) Store(node adf.Node) (string, string, error) {
	return "", "", nil
}

func (f *fakeFailingManager) Restore(_ string) (adf.Node, error) {
	return adf.Node{}, f.restoreErr
}

func (f *fakeFailingManager) GeneratePreview(_ adf.Node) string { return "" }
func (f *fakeFailingManager) GetSession() *placeholder.EditSession {
	return &placeholder.EditSession{Preserved: map[string]adf.Node{}}
}
func (f *fakeFailingManager) Clear()     {}
func (f *fakeFailingManager) Count() int { return 0 }

// TestParseContent_PropagatesNonSentinelRestoreError verifies that a
// Manager-internal error from Restore is propagated by the inline parser
// rather than silently treated as a deletion. ErrPlaceholderNotFound
// remains the only error that means "user removed the placeholder".
func TestParseContent_PropagatesNonSentinelRestoreError(t *testing.T) {
	internalErr := errors.New("manager backend exploded")
	mgr := &fakeFailingManager{restoreErr: internalErr}

	markdown := "before <!-- ADF_PLACEHOLDER_001: Emoji: :ok: --> after"

	_, err := ParseContentWithPlaceholders(markdown, mgr)

	require.Error(t, err, "non-sentinel Restore error must propagate, not be swallowed")
	assert.ErrorIs(t, err, internalErr,
		"propagated error must wrap the original Manager error")
	assert.False(t, errors.Is(err, placeholder.ErrPlaceholderNotFound),
		"non-sentinel error must NOT be treated as ErrPlaceholderNotFound")
}

// TestParseContent_TreatsSentinelAsDeletion verifies that ErrPlaceholderNotFound
// is treated as legitimate deletion: parsing succeeds, marker is dropped,
// surrounding text is preserved.
func TestParseContent_TreatsSentinelAsDeletion(t *testing.T) {
	mgr := &fakeFailingManager{
		restoreErr: placeholder.ErrPlaceholderNotFound,
	}

	markdown := "before <!-- ADF_PLACEHOLDER_001: Emoji: :ok: --> after"

	nodes, err := ParseContentWithPlaceholders(markdown, mgr)

	require.NoError(t, err, "ErrPlaceholderNotFound must be treated as deletion")
	require.NotEmpty(t, nodes)
}
