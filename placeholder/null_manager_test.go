package placeholder

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullManager_Store(t *testing.T) {
	tests := []struct {
		name           string
		node           adf_types.ADFNode
		wantID         string
		wantPreviewSub string
		wantErr        bool
	}{
		{
			name:           "code block returns empty ID and preview",
			node:           adf_types.ADFNode{Type: adf_types.NodeTypeCodeBlock, Attrs: map[string]any{"language": "go"}},
			wantID:         "",
			wantPreviewSub: "Code Block",
		},
		{
			name:           "table returns empty ID and preview",
			node:           adf_types.ADFNode{Type: adf_types.NodeTypeTable},
			wantID:         "",
			wantPreviewSub: "Table",
		},
		{
			name:           "mention returns empty ID and preview",
			node:           adf_types.ADFNode{Type: adf_types.NodeTypeMention, Attrs: map[string]any{"text": "@alice"}},
			wantID:         "",
			wantPreviewSub: "Mention: @alice",
		},
		{
			name:    "empty type returns error",
			node:    adf_types.ADFNode{Type: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewNullManager()
			id, preview, err := m.Store(tt.node)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			assert.Contains(t, preview, tt.wantPreviewSub)
		})
	}
}

func TestNullManager_StoreDoesNotAccumulate(t *testing.T) {
	m := NewNullManager()

	_, _, _ = m.Store(adf_types.ADFNode{Type: adf_types.NodeTypeCodeBlock})
	_, _, _ = m.Store(adf_types.ADFNode{Type: adf_types.NodeTypeTable})

	assert.Equal(t, 0, m.Count(), "NullManager should never accumulate stored nodes")
}

func TestNullManager_Restore(t *testing.T) {
	m := NewNullManager()
	_, err := m.Restore("ADF_PLACEHOLDER_001")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "display mode")
}

func TestNullManager_GetSession(t *testing.T) {
	m := NewNullManager()
	session := m.GetSession()

	require.NotNil(t, session, "GetSession must return non-nil EditSession")
	assert.NotNil(t, session.Preserved, "Preserved map must be initialized")
}

func TestNullManager_Count(t *testing.T) {
	m := NewNullManager()
	assert.Equal(t, 0, m.Count())
}

func TestNullManager_Clear(t *testing.T) {
	m := NewNullManager()
	assert.NotPanics(t, func() { m.Clear() })
}

func TestNullManager_GeneratePreview(t *testing.T) {
	m := NewNullManager()

	tests := []struct {
		name    string
		node    adf_types.ADFNode
		wantSub string
	}{
		{
			name:    "code block",
			node:    adf_types.ADFNode{Type: adf_types.NodeTypeCodeBlock, Attrs: map[string]any{"language": "python"}},
			wantSub: "Code Block (python",
		},
		{
			name:    "panel",
			node:    adf_types.ADFNode{Type: adf_types.NodeTypePanel, Attrs: map[string]any{"panelType": "warning"}},
			wantSub: "Warning Panel",
		},
		{
			name:    "rule",
			node:    adf_types.ADFNode{Type: adf_types.NodeTypeRule},
			wantSub: "Horizontal Rule",
		},
		{
			name:    "unknown type",
			node:    adf_types.ADFNode{Type: "customNode"},
			wantSub: "complex content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := m.GeneratePreview(tt.node)
			assert.Contains(t, preview, tt.wantSub)
		})
	}
}
