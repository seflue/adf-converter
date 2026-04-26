package placeholder

import (
	"errors"
	"testing"

	adf "github.com/seflue/adf-converter/adf/adftypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time assertion: NewNoop must return the Manager interface.
var _ Manager = NewNoop()

func TestNewNoop_ReturnsNonNil(t *testing.T) {
	require.NotNil(t, NewNoop())
}

func TestNoop_Store(t *testing.T) {
	tests := []struct {
		name           string
		node           adf.Node
		wantID         string
		wantPreviewSub string
		wantErr        bool
	}{
		{
			name:           "code block returns empty ID and preview",
			node:           adf.Node{Type: adf.NodeTypeCodeBlock, Attrs: map[string]any{"language": "go"}},
			wantID:         "",
			wantPreviewSub: "Code Block",
		},
		{
			name:           "table returns empty ID and preview",
			node:           adf.Node{Type: adf.NodeTypeTable},
			wantID:         "",
			wantPreviewSub: "Table",
		},
		{
			name:           "mention returns empty ID and preview",
			node:           adf.Node{Type: adf.NodeTypeMention, Attrs: map[string]any{"text": "@alice"}},
			wantID:         "",
			wantPreviewSub: "Mention: @alice",
		},
		{
			name:    "empty type returns error",
			node:    adf.Node{Type: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewNoop()
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

func TestNoop_StoreDoesNotAccumulate(t *testing.T) {
	m := NewNoop()

	_, _, _ = m.Store(adf.Node{Type: adf.NodeTypeCodeBlock})
	_, _, _ = m.Store(adf.Node{Type: adf.NodeTypeTable})

	assert.Equal(t, 0, m.Count(), "noop manager should never accumulate stored nodes")
}

func TestNoop_RestoreReturnsNotFound(t *testing.T) {
	m := NewNoop()
	node, err := m.Restore("ADF_PLACEHOLDER_001")

	require.Error(t, err, "noop Restore must report not-found via sentinel")
	assert.True(t, errors.Is(err, ErrPlaceholderNotFound),
		"expected errors.Is(err, ErrPlaceholderNotFound), got %v", err)
	assert.Equal(t, adf.Node{}, node, "noop Restore returns zero-value Node")
}

func TestNoop_RestoreEmptyIDReturnsNotFound(t *testing.T) {
	m := NewNoop()
	node, err := m.Restore("")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPlaceholderNotFound))
	assert.Equal(t, adf.Node{}, node)
}

func TestNoop_GetSession(t *testing.T) {
	m := NewNoop()
	session := m.GetSession()

	require.NotNil(t, session, "GetSession must return non-nil EditSession")
	assert.NotNil(t, session.Preserved, "Preserved map must be initialized")
}

func TestNoop_Count(t *testing.T) {
	m := NewNoop()
	assert.Equal(t, 0, m.Count())
}

func TestNoop_Clear(t *testing.T) {
	m := NewNoop()
	assert.NotPanics(t, func() { m.Clear() })
}

func TestNoop_GeneratePreview(t *testing.T) {
	m := NewNoop()

	tests := []struct {
		name    string
		node    adf.Node
		wantSub string
	}{
		{
			name:    "code block",
			node:    adf.Node{Type: adf.NodeTypeCodeBlock, Attrs: map[string]any{"language": "python"}},
			wantSub: "Code Block (python",
		},
		{
			name:    "panel",
			node:    adf.Node{Type: adf.NodeTypePanel, Attrs: map[string]any{"panelType": "warning"}},
			wantSub: "Warning Panel",
		},
		{
			name:    "rule",
			node:    adf.Node{Type: adf.NodeTypeRule},
			wantSub: "Horizontal Rule",
		},
		{
			name:    "unknown type",
			node:    adf.Node{Type: "customNode"},
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
