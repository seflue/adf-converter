package elements

import (
	"strings"
	"testing"

	"github.com/seflue/adf-converter/adf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stripLocalIds returns a deep copy of node with all localId attrs removed.
// Used in tests that check structural correctness without caring about generated IDs.
func stripLocalIds(node adf.Node) adf.Node {
	cleaned := node
	if cleaned.Attrs != nil {
		newAttrs := make(map[string]any, len(cleaned.Attrs))
		for k, v := range cleaned.Attrs {
			if k != "localId" {
				newAttrs[k] = v
			}
		}
		cleaned.Attrs = newAttrs
	}
	if cleaned.Content != nil {
		newContent := make([]adf.Node, len(cleaned.Content))
		for i, child := range cleaned.Content {
			newContent[i] = stripLocalIds(child)
		}
		cleaned.Content = newContent
	}
	return cleaned
}

func TestTaskListConverter_FromMarkdown_Plain(t *testing.T) {
	conv := NewTaskListConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		expectedConsumed int
		want             adf.Node
	}{
		{
			name:             "single unchecked task",
			lines:            []string{"- [ ] Task 1"},
			startIndex:       0,
			expectedConsumed: 1,
			want: adf.Node{
				Type:  "taskList",
				Attrs: map[string]any{},
				Content: []adf.Node{
					{
						Type: "taskItem",
						Attrs: map[string]any{
							"state": "TODO",
						},
						Content: []adf.Node{
							{Type: "text", Text: "Task 1"},
						},
					},
				},
			},
		},
		{
			name:             "single checked task",
			lines:            []string{"- [x] Task 1"},
			startIndex:       0,
			expectedConsumed: 1,
			want: adf.Node{
				Type:  "taskList",
				Attrs: map[string]any{},
				Content: []adf.Node{
					{
						Type: "taskItem",
						Attrs: map[string]any{
							"state": "DONE",
						},
						Content: []adf.Node{
							{Type: "text", Text: "Task 1"},
						},
					},
				},
			},
		},
		{
			name:             "uppercase X in checkbox",
			lines:            []string{"- [X] Task 1"},
			startIndex:       0,
			expectedConsumed: 1,
			want: adf.Node{
				Type:  "taskList",
				Attrs: map[string]any{},
				Content: []adf.Node{
					{
						Type: "taskItem",
						Attrs: map[string]any{
							"state": "DONE",
						},
						Content: []adf.Node{
							{Type: "text", Text: "Task 1"},
						},
					},
				},
			},
		},
		{
			name:             "mixed checked and unchecked tasks",
			lines:            []string{"- [ ] Task 1", "- [x] Task 2", "- [ ] Task 3"},
			startIndex:       0,
			expectedConsumed: 3,
			want: adf.Node{
				Type:  "taskList",
				Attrs: map[string]any{},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO"},
						Content: []adf.Node{{Type: "text", Text: "Task 3"}},
					},
				},
			},
		},
		{
			name:             "task list with empty lines",
			lines:            []string{"- [ ] Task 1", "", "- [x] Task 2", "", "- [ ] Task 3"},
			startIndex:       0,
			expectedConsumed: 5,
			want: adf.Node{
				Type:  "taskList",
				Attrs: map[string]any{},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO"},
						Content: []adf.Node{{Type: "text", Text: "Task 3"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, err := conv.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConsumed, consumed)
			assert.Equal(t, tt.want, stripLocalIds(got))
		})
	}
}

func TestTaskListConverter_FromMarkdown_XMLWrapped(t *testing.T) {
	conv := NewTaskListConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		expectedConsumed int
		want             adf.Node
	}{
		{
			name: "XML wrapped with localId",
			lines: []string{
				`<taskList localId="abc123">`,
				"- [ ] Task 1",
				"- [x] Task 2",
				"</taskList>",
			},
			startIndex:       0,
			expectedConsumed: 4,
			want: adf.Node{
				Type: "taskList",
				Attrs: map[string]any{
					"localId": "abc123",
				},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO", "localId": "abc123-item-0"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE", "localId": "abc123-item-1"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
				},
			},
		},
		{
			name: "XML wrapped with multiple attributes",
			lines: []string{
				`<taskList localId="xyz789" completed="2" total="3">`,
				"- [ ] Task 1",
				"- [x] Task 2",
				"- [x] Task 3",
				"</taskList>",
			},
			startIndex:       0,
			expectedConsumed: 5,
			want: adf.Node{
				Type: "taskList",
				Attrs: map[string]any{
					"localId":   "xyz789",
					"completed": 2,
					"total":     3,
				},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO", "localId": "xyz789-item-0"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE", "localId": "xyz789-item-1"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE", "localId": "xyz789-item-2"},
						Content: []adf.Node{{Type: "text", Text: "Task 3"}},
					},
				},
			},
		},
		{
			name: "XML wrapped with empty lines inside",
			lines: []string{
				`<taskList localId="test">`,
				"- [ ] Task 1",
				"",
				"- [x] Task 2",
				"</taskList>",
			},
			startIndex:       0,
			expectedConsumed: 5,
			want: adf.Node{
				Type: "taskList",
				Attrs: map[string]any{
					"localId": "test",
				},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO", "localId": "test-item-0"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE", "localId": "test-item-1"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
				},
			},
		},
		{
			name: "XML wrapped with trailing lines",
			lines: []string{
				`<taskList localId="abc">`,
				"- [ ] Task",
				"</taskList>",
				"trailing content",
			},
			startIndex:       0,
			expectedConsumed: 3,
			want: adf.Node{
				Type: "taskList",
				Attrs: map[string]any{
					"localId": "abc",
				},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO", "localId": "abc-item-0"},
						Content: []adf.Node{{Type: "text", Text: "Task"}},
					},
				},
			},
		},
		{
			name: "XML wrapped with startIndex",
			lines: []string{
				"prefix line",
				`<taskList localId="skip">`,
				"- [x] Done",
				"</taskList>",
			},
			startIndex:       1,
			expectedConsumed: 3,
			want: adf.Node{
				Type: "taskList",
				Attrs: map[string]any{
					"localId": "skip",
				},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE", "localId": "skip-item-0"},
						Content: []adf.Node{{Type: "text", Text: "Done"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, err := conv.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConsumed, consumed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskListConverter_FromMarkdown_EdgeCases(t *testing.T) {
	conv := NewTaskListConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		expectedConsumed int
		want             adf.Node
	}{
		{
			name:             "empty lines slice",
			lines:            []string{},
			startIndex:       0,
			expectedConsumed: 0,
			want: adf.Node{
				Type:    "taskList",
				Attrs:   map[string]any{},
				Content: nil,
			},
		},
		{
			name:             "startIndex out of bounds",
			lines:            []string{"- [ ] Task"},
			startIndex:       5,
			expectedConsumed: 0,
			want: adf.Node{
				Type:    "taskList",
				Attrs:   map[string]any{},
				Content: nil,
			},
		},
		{
			name:             "invalid checkbox syntax stops boundary",
			lines:            []string{"- [?] Invalid task", "- [ ] Valid task"},
			startIndex:       0,
			expectedConsumed: 0,
			want: adf.Node{
				Type:    "taskList",
				Attrs:   map[string]any{},
				Content: nil,
			},
		},
		{
			name:             "boundary: stops at non-task line",
			lines:            []string{"- [ ] Task 1", "- [x] Task 2", "Regular text"},
			startIndex:       0,
			expectedConsumed: 2,
			want: adf.Node{
				Type:  "taskList",
				Attrs: map[string]any{},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, err := conv.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConsumed, consumed)
			assert.Equal(t, tt.want, stripLocalIds(got))
		})
	}
}

func TestTaskListConverter_RoundTrip(t *testing.T) {
	conv := NewTaskListConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	tests := []struct {
		name     string
		original adf.Node
	}{
		{
			name: "simple task list",
			original: adf.Node{
				Type: "taskList",
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "DONE"},
						Content: []adf.Node{{Type: "text", Text: "Task 2"}},
					},
				},
			},
		},
		{
			name: "task list with attributes",
			original: adf.Node{
				Type: "taskList",
				Attrs: map[string]any{
					"localId": "test123",
				},
				Content: []adf.Node{
					{
						Type:    "taskItem",
						Attrs:   map[string]any{"state": "TODO", "localId": "test123-item-0"},
						Content: []adf.Node{{Type: "text", Text: "Task 1"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert ADF -> Markdown
			mdResult, err := conv.ToMarkdown(tt.original, ctx)
			require.NoError(t, err)

			// Convert Markdown -> ADF
			lines := strings.Split(mdResult.Content, "\n")
			roundtrip, consumed, err := conv.FromMarkdown(lines, 0, ctx)
			require.NoError(t, err)
			assert.Greater(t, consumed, 0)

			// Verify structure matches
			assert.Equal(t, tt.original.Type, roundtrip.Type)
			assert.Equal(t, len(tt.original.Content), len(roundtrip.Content))

			// Verify each task item state is preserved
			for i, origItem := range tt.original.Content {
				rtItem := roundtrip.Content[i]
				assert.Equal(t, origItem.Type, rtItem.Type)
				assert.Equal(t, origItem.Attrs["state"], rtItem.Attrs["state"])
			}
		})
	}
}

func TestTaskListConverter_FromMarkdown_Plain_GeneratesLocalId(t *testing.T) {
	// Jira requires localId on taskList and taskItem nodes.
	// Plain markdown (no XML wrapper) must auto-generate UUIDs.
	conv := NewTaskListConverter()
	ctx := adf.ConversionContext{Registry: newTestRegistry()}

	lines := []string{"- [ ] Task 1", "- [x] Task 2"}
	node, consumed, err := conv.FromMarkdown(lines, 0, ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, consumed)

	localId, ok := node.Attrs["localId"].(string)
	assert.True(t, ok, "taskList.attrs.localId must be a non-empty string")
	assert.NotEmpty(t, localId, "taskList.attrs.localId must not be empty")

	require.Len(t, node.Content, 2)
	for _, item := range node.Content {
		itemLocalId, ok := item.Attrs["localId"].(string)
		assert.True(t, ok, "taskItem.attrs.localId must be a non-empty string")
		assert.NotEmpty(t, itemLocalId, "taskItem.attrs.localId must not be empty")
	}
}

func TestTaskListConverter_ToMarkdown_TrailingNewline(t *testing.T) {
	// Block-level elements must end with \n\n so the next element starts on its own line.
	// Without this, the heading following a taskList ends up on the same line as </taskList>.
	conv := NewTaskListConverter()

	node := adf.Node{
		Type: "taskList",
		Attrs: map[string]any{
			"localId": "test-id",
		},
		Content: []adf.Node{
			{Type: "taskItem", Attrs: map[string]any{"state": "TODO"}, Content: []adf.Node{{Type: "text", Text: "Task"}}},
		},
	}

	t.Run("plain (no preserve attrs)", func(t *testing.T) {
		result, err := conv.ToMarkdown(node, adf.ConversionContext{Registry: newTestRegistry(), PreserveAttrs: false})
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(result.Content, "\n\n"),
			"plain task list output must end with \\n\\n, got: %q", result.Content)
	})

}

// Suppress unused import warning
var _ = strings.Split
