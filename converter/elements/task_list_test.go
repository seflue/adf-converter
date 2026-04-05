package elements

import (
	"strings"
	"testing"

	"adf-converter/adf_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskListConverter_FromMarkdown_Plain(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		expectedConsumed int
		want             adf_types.ADFNode
	}{
		{
			name:             "single unchecked task",
			lines:            []string{"- [ ] Task 1"},
			startIndex:       0,
			expectedConsumed: 1,
			want: adf_types.ADFNode{
				Type:  "taskList",
				Attrs: map[string]interface{}{},
				Content: []adf_types.ADFNode{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "TODO",
						},
						Content: []adf_types.ADFNode{
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
			want: adf_types.ADFNode{
				Type:  "taskList",
				Attrs: map[string]interface{}{},
				Content: []adf_types.ADFNode{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "DONE",
						},
						Content: []adf_types.ADFNode{
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
			want: adf_types.ADFNode{
				Type:  "taskList",
				Attrs: map[string]interface{}{},
				Content: []adf_types.ADFNode{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "DONE",
						},
						Content: []adf_types.ADFNode{
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
			want: adf_types.ADFNode{
				Type:  "taskList",
				Attrs: map[string]interface{}{},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 3"}},
					},
				},
			},
		},
		{
			name:             "task list with empty lines",
			lines:            []string{"- [ ] Task 1", "", "- [x] Task 2", "", "- [ ] Task 3"},
			startIndex:       0,
			expectedConsumed: 5,
			want: adf_types.ADFNode{
				Type:  "taskList",
				Attrs: map[string]interface{}{},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 3"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, err := converter.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConsumed, consumed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskListConverter_FromMarkdown_XMLWrapped(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		expectedConsumed int
		want             adf_types.ADFNode
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
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "abc123",
				},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO", "localId": "abc123-item-0"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE", "localId": "abc123-item-1"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
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
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId":   "xyz789",
					"completed": 2,
					"total":     3,
				},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO", "localId": "xyz789-item-0"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE", "localId": "xyz789-item-1"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE", "localId": "xyz789-item-2"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 3"}},
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
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "test",
				},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO", "localId": "test-item-0"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE", "localId": "test-item-1"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
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
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "abc",
				},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO", "localId": "abc-item-0"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task"}},
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
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "skip",
				},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE", "localId": "skip-item-0"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Done"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, err := converter.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConsumed, consumed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskListConverter_FromMarkdown_EdgeCases(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name             string
		lines            []string
		startIndex       int
		expectedConsumed int
		want             adf_types.ADFNode
	}{
		{
			name:             "empty lines slice",
			lines:            []string{},
			startIndex:       0,
			expectedConsumed: 0,
			want: adf_types.ADFNode{
				Type:    "taskList",
				Attrs:   map[string]interface{}{},
				Content: nil,
			},
		},
		{
			name:             "startIndex out of bounds",
			lines:            []string{"- [ ] Task"},
			startIndex:       5,
			expectedConsumed: 0,
			want: adf_types.ADFNode{
				Type:    "taskList",
				Attrs:   map[string]interface{}{},
				Content: nil,
			},
		},
		{
			name:             "invalid checkbox syntax stops boundary",
			lines:            []string{"- [?] Invalid task", "- [ ] Valid task"},
			startIndex:       0,
			expectedConsumed: 0,
			want: adf_types.ADFNode{
				Type:    "taskList",
				Attrs:   map[string]interface{}{},
				Content: nil,
			},
		},
		{
			name:             "boundary: stops at non-task line",
			lines:            []string{"- [ ] Task 1", "- [x] Task 2", "Regular text"},
			startIndex:       0,
			expectedConsumed: 2,
			want: adf_types.ADFNode{
				Type:  "taskList",
				Attrs: map[string]interface{}{},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, consumed, err := converter.FromMarkdown(tt.lines, tt.startIndex, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedConsumed, consumed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskListConverter_RoundTrip(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{PreserveAttrs: true}

	tests := []struct {
		name     string
		original adf_types.ADFNode
	}{
		{
			name: "simple task list",
			original: adf_types.ADFNode{
				Type: "taskList",
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "DONE"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 2"}},
					},
				},
			},
		},
		{
			name: "task list with attributes",
			original: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "test123",
				},
				Content: []adf_types.ADFNode{
					{
						Type:    "taskItem",
						Attrs:   map[string]interface{}{"state": "TODO", "localId": "test123-item-0"},
						Content: []adf_types.ADFNode{{Type: "text", Text: "Task 1"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert ADF -> Markdown
			mdResult, err := converter.ToMarkdown(tt.original, ctx)
			require.NoError(t, err)

			// Convert Markdown -> ADF
			lines := strings.Split(mdResult.Content, "\n")
			roundtrip, consumed, err := converter.FromMarkdown(lines, 0, ctx)
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

// Suppress unused import warning
var _ = strings.Split
