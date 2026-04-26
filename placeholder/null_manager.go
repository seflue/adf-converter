package placeholder

import (
	"fmt"

	adf "github.com/seflue/adf-converter/adf/adftypes"
)

// nullManager implements Manager for display-only mode. Store returns
// preview text without generating placeholder IDs, and Restore is a
// no-op passthrough so callers that traverse markdown containing
// placeholder comments do not blow up.
type nullManager struct {
	session *EditSession
}

// NewNoop returns a Manager that never accumulates state. Use it for
// read-only display flows where Store should yield preview text instead
// of placeholder comments and Restore is not expected to recover nodes.
func NewNoop() Manager {
	return &nullManager{
		session: &EditSession{
			Preserved: make(map[string]adf.Node),
		},
	}
}

func (m *nullManager) Store(node adf.Node) (string, string, error) {
	if node.Type == "" {
		return "", "", fmt.Errorf("cannot store node with empty type")
	}
	return "", generatePreview(node), nil
}

// Restore always returns ErrPlaceholderNotFound: the noop manager never
// stores anything, so there is nothing to recover. Callers detect this
// with errors.Is(err, ErrPlaceholderNotFound) and treat it the same as
// a user-deleted placeholder, keeping display-mode flows working.
func (m *nullManager) Restore(placeholderID string) (adf.Node, error) {
	return adf.Node{}, fmt.Errorf("placeholder ID %s: %w", placeholderID, ErrPlaceholderNotFound)
}

func (m *nullManager) GeneratePreview(node adf.Node) string {
	return generatePreview(node)
}

func (m *nullManager) GetSession() *EditSession {
	return m.session
}

func (m *nullManager) Clear() {}

func (m *nullManager) Count() int {
	return 0
}
