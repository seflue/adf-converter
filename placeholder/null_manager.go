package placeholder

import (
	"fmt"

	"adf-converter/adf_types"
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
			Preserved: make(map[string]adf_types.ADFNode),
		},
	}
}

func (m *NullManager) Store(node adf_types.ADFNode) (string, string, error) {
	if node.Type == "" {
		return "", "", fmt.Errorf("cannot store node with empty type")
	}
	return "", generatePreview(node), nil
}

func (m *NullManager) Restore(_ string) (adf_types.ADFNode, error) {
	return adf_types.ADFNode{}, fmt.Errorf("display mode: restore not supported")
}

func (m *NullManager) GeneratePreview(node adf_types.ADFNode) string {
	return generatePreview(node)
}

func (m *NullManager) GetSession() *EditSession {
	return m.session
}

func (m *NullManager) Clear() {}

func (m *NullManager) Count() int {
	return 0
}
