package dedent

import (
	"strings"
	"testing"
)

func TestDedentLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "basic uniform indentation",
			input: []string{
				"  line 1",
				"  line 2",
				"  line 3",
			},
			expected: []string{
				"line 1",
				"line 2",
				"line 3",
			},
		},
		{
			name: "nested structure with relative indentation",
			input: []string{
				"  - first",
				"    - sub 1",
				"    - sub 2",
				"  - second",
			},
			expected: []string{
				"- first",
				"  - sub 1",
				"  - sub 2",
				"- second",
			},
		},
		{
			name: "user's actual failing case",
			input: []string{
				"  - first",
				"    - sub 1",
				"    - sub 2",
				"  - second",
				"    - sub 3",
				"    - sub 4",
				"  - multiline",
			},
			expected: []string{
				"- first",
				"  - sub 1",
				"  - sub 2",
				"- second",
				"  - sub 3",
				"  - sub 4",
				"- multiline",
			},
		},
		{
			name: "no indentation needed",
			input: []string{
				"line 1",
				"line 2",
			},
			expected: []string{
				"line 1",
				"line 2",
			},
		},
		{
			name: "empty lines preserved",
			input: []string{
				"  line 1",
				"",
				"  line 2",
			},
			expected: []string{
				"line 1",
				"",
				"line 2",
			},
		},
		{
			name: "mixed indentation levels",
			input: []string{
				"    deeply indented",
				"  less indented",
				"      even more",
			},
			expected: []string{
				"  deeply indented",
				"less indented",
				"    even more",
			},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "all empty lines",
			input: []string{
				"",
				"   ",
				"",
			},
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name: "tabs and spaces",
			input: []string{
				"\t\tline 1",
				"\t\tline 2",
			},
			expected: []string{
				"line 1",
				"line 2",
			},
		},
		{
			name: "line shorter than minIndent with content",
			input: []string{
				"    line 1",
				"  X", // Only 2 spaces but has content
				"    line 2",
			},
			expected: []string{
				"  line 1",
				"X",
				"  line 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DedentLines(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("DedentLines() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("DedentLines() line %d = %q, want %q", i, result[i], tt.expected[i])
					t.Logf("Full result:\n%s", strings.Join(result, "\n"))
					t.Logf("Expected:\n%s", strings.Join(tt.expected, "\n"))
					return
				}
			}
		})
	}
}

// TestDedentLines_Debug helps debug the actual failing case
func TestDedentLines_Debug(t *testing.T) {
	// This is what the user SHOULD be seeing before dedentation
	input := []string{
		"  - first",
		"    - sub 1",
		"    - sub 2",
		"  - second",
		"    - sub 3",
		"    - sub 4",
		"  - multiline",
	}

	result := DedentLines(input)

	t.Logf("Input lines:")
	for i, line := range input {
		t.Logf("  [%d] %q", i, line)
	}

	t.Logf("\nOutput lines:")
	for i, line := range result {
		t.Logf("  [%d] %q", i, line)
	}

	// Expected output preserves the 2-space nesting
	expected := []string{
		"- first",
		"  - sub 1",
		"  - sub 2",
		"- second",
		"  - sub 3",
		"  - sub 4",
		"- multiline",
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("Line %d mismatch: got %q, want %q", i, result[i], expected[i])
		}
	}
}
