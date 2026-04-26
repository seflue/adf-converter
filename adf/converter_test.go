package adf_test

import (
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConverter_EmptyRegistry verifies that NewConverter without a
// populated registry returns an error instead of silently yielding an
// unusable instance (ac-0115 Foot-Gun-Fix).
func TestNewConverter_EmptyRegistry(t *testing.T) {
	c, err := adf.NewConverter()

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "empty registry")
}

// TestNewConverter_EmptyRegistryViaWithRegistry guards against callers
// passing a freshly constructed empty registry explicitly.
func TestNewConverter_EmptyRegistryViaWithRegistry(t *testing.T) {
	c, err := adf.NewConverter(
		adf.WithRegistry(adf.NewConverterRegistry()),
	)

	require.Error(t, err)
	assert.Nil(t, c)
}
