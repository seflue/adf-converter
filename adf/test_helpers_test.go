package adf_test

import (
	"encoding/json"
	"testing"

	"github.com/seflue/adf-converter/adf"
	"github.com/stretchr/testify/require"
)

// Helper function to parse ADF payload from JSON string
// Used across all test files for consistency
func parseTestADFPayload(t *testing.T, payload string) adf.Document {
	t.Helper()
	var parsed struct {
		Fields struct {
			Description adf.Document `json:"description"`
		} `json:"fields"`
	}

	err := json.Unmarshal([]byte(payload), &parsed)
	require.NoError(t, err, "Failed to parse ADF payload")
	return parsed.Fields.Description
}
