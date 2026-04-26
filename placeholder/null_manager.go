package placeholder

import (
	"fmt"

	adf "github.com/seflue/adf-converter/adf/adftypes"
)

// NullManager implements the Manager interface for display-only mode.
// Store() returns preview text without generating placeholder IDs,
// so callers can render readable markdown without comment wrappers.
type NullManager struct {
	session *EditSession
}

// NewNullManager creates a NullManager with an empty but non-nil session.
func NewNullManager() Manager {
	return &NullManager{
		session: &EditSession{
			Preserved: make(map[string]adf.Node),
		},
	}
}

func (m *NullManager) Store(node adf.Node) (string, string, error) {
	if node.Type == "" {
		return "", "", fmt.Errorf("cannot store node with empty type")
	}
	return "", generatePreview(node), nil
}

func (m *NullManager) Restore(_ string) (adf.Node, error) {
	return adf.Node{}, fmt.Errorf("display mode: restore not supported")
}

func (m *NullManager) GeneratePreview(node adf.Node) string {
	return generatePreview(node)
}

func (m *NullManager) GetSession() *EditSession {
	return m.session
}

func (m *NullManager) Clear() {}

func (m *NullManager) Count() int {
	return 0
}
