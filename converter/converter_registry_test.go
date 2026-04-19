package converter

import (
	"testing"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inlineOnlyConverter implements ElementConverter but NOT BlockParser
// (no CanParseLine method). Used to exercise the type-assert error path
// in RegisterBlockParser.
type inlineOnlyConverter struct{}

func (inlineOnlyConverter) ToMarkdown(adf_types.ADFNode, ConversionContext) (EnhancedConversionResult, error) {
	return EnhancedConversionResult{}, nil
}

func (inlineOnlyConverter) FromMarkdown([]string, int, ConversionContext) (adf_types.ADFNode, int, error) {
	return adf_types.ADFNode{}, 0, nil
}

func (inlineOnlyConverter) CanHandle(ADFNodeType) bool            { return true }
func (inlineOnlyConverter) GetStrategy() ConversionStrategy       { return StandardMarkdown }
func (inlineOnlyConverter) ValidateInput(any) error               { return nil }

// blockConverter implements BlockParser (ElementConverter + CanParseLine).
type blockConverter struct{ inlineOnlyConverter }

func (blockConverter) CanParseLine(string) bool { return true }

func TestRegister_NilConverterReturnsError(t *testing.T) {
	r := NewConverterRegistry()

	err := r.Register("text", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "text")
	assert.Equal(t, 0, r.Count())
}

func TestRegister_ValidConverterReturnsNil(t *testing.T) {
	r := NewConverterRegistry()

	err := r.Register("text", blockConverter{})

	require.NoError(t, err)
	assert.Equal(t, 1, r.Count())
}

func TestRegister_DuplicateReplacesSilently(t *testing.T) {
	// Per ac-0100 decision (Option a): replace-semantics preserved.
	// Duplicate-check stays out of scope; belongs to ac-0094.
	r := NewConverterRegistry()
	require.NoError(t, r.Register("text", blockConverter{}))

	err := r.Register("text", blockConverter{})

	require.NoError(t, err)
	assert.Equal(t, 1, r.Count())
}

func TestRegisterBlockParser_UnregisteredReturnsError(t *testing.T) {
	r := NewConverterRegistry()

	err := r.RegisterBlockParser("missing")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestRegisterBlockParser_NonBlockParserReturnsError(t *testing.T) {
	r := NewConverterRegistry()
	require.NoError(t, r.Register("text", inlineOnlyConverter{}))

	err := r.RegisterBlockParser("text")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "text")
}

func TestRegisterBlockParser_ValidReturnsNil(t *testing.T) {
	r := NewConverterRegistry()
	require.NoError(t, r.Register("paragraph", blockConverter{}))

	err := r.RegisterBlockParser("paragraph")

	require.NoError(t, err)
	assert.Len(t, r.BlockParsers(), 1)
}

func TestMustRegister_NilConverterPanics(t *testing.T) {
	r := NewConverterRegistry()

	assert.Panics(t, func() { r.MustRegister("text", nil) })
}

func TestMustRegister_ValidDoesNotPanic(t *testing.T) {
	r := NewConverterRegistry()

	assert.NotPanics(t, func() { r.MustRegister("text", blockConverter{}) })
	assert.Equal(t, 1, r.Count())
}

func TestMustRegisterBlockParser_UnregisteredPanics(t *testing.T) {
	r := NewConverterRegistry()

	assert.Panics(t, func() { r.MustRegisterBlockParser("missing") })
}

func TestMustRegisterBlockParser_NonBlockParserPanics(t *testing.T) {
	r := NewConverterRegistry()
	r.MustRegister("text", inlineOnlyConverter{})

	assert.Panics(t, func() { r.MustRegisterBlockParser("text") })
}

func TestMustRegisterBlockParser_ValidDoesNotPanic(t *testing.T) {
	r := NewConverterRegistry()
	r.MustRegister("paragraph", blockConverter{})

	assert.NotPanics(t, func() { r.MustRegisterBlockParser("paragraph") })
}
