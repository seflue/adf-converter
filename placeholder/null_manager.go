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

// Restore is a no-op: the noop manager never stores anything, so there
// is nothing to recover. Returning a zero Node with nil error keeps the
// display converter from blowing up on placeholder comments embedded in
// the markdown.
func (m *nullManager) Restore(_ string) (adf.Node, error) {
	return adf.Node{}, nil
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
