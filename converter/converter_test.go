package converter_test

import (
	"testing"

	"github.com/seflue/adf-converter/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConverter_EmptyRegistry verifies that NewConverter without a
// populated registry returns an error instead of silently yielding an
// unusable instance (ac-0115 Foot-Gun-Fix).
func TestNewConverter_EmptyRegistry(t *testing.T) {
	c, err := converter.NewConverter()

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "empty registry")
}

// TestNewConverter_EmptyRegistryViaWithRegistry guards against callers
// passing a freshly constructed empty registry explicitly.
func TestNewConverter_EmptyRegistryViaWithRegistry(t *testing.T) {
	c, err := converter.NewConverter(
		converter.WithRegistry(converter.NewConverterRegistry()),
	)

	require.Error(t, err)
	assert.Nil(t, c)
}
