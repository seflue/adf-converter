package elements

import (
	"testing"

	"adf-converter/adf_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskListConverter_FromMarkdown_PlainTaskList(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name     string
		markdown string
		want     adf_types.ADFNode
	}{
		{
			name:     "single unchecked task",
			markdown: `- [ ] Task 1`,
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
			name:     "single checked task",
			markdown: `- [x] Task 1`,
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
			name:     "uppercase X in checkbox",
			markdown: `- [X] Task 1`,
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
			name: "mixed checked and unchecked tasks",
			markdown: `- [ ] Task 1
- [x] Task 2
- [ ] Task 3`,
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
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "DONE",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 2"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "TODO",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 3"},
						},
					},
				},
			},
		},
		{
			name: "task list with empty lines",
			markdown: `- [ ] Task 1

- [x] Task 2

- [ ] Task 3`,
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
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "DONE",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 2"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "TODO",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 3"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.FromMarkdown(tt.markdown, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskListConverter_FromMarkdown_XMLWrapped(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name     string
		markdown string
		want     adf_types.ADFNode
	}{
		{
			name: "XML wrapped with localId",
			markdown: `<taskList localId="abc123">
- [ ] Task 1
- [x] Task 2
</taskList>`,
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "abc123",
				},
				Content: []adf_types.ADFNode{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "TODO",
							"localId": "abc123-item-0",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 1"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "DONE",
							"localId": "abc123-item-1",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 2"},
						},
					},
				},
			},
		},
		{
			name: "XML wrapped with multiple attributes",
			markdown: `<taskList localId="xyz789" completed="2" total="3">
- [ ] Task 1
- [x] Task 2
- [x] Task 3
</taskList>`,
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId":   "xyz789",
					"completed": 2,
					"total":     3,
				},
				Content: []adf_types.ADFNode{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "TODO",
							"localId": "xyz789-item-0",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 1"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "DONE",
							"localId": "xyz789-item-1",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 2"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "DONE",
							"localId": "xyz789-item-2",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 3"},
						},
					},
				},
			},
		},
		{
			name: "XML wrapped with empty lines inside",
			markdown: `<taskList localId="test">
- [ ] Task 1

- [x] Task 2
</taskList>`,
			want: adf_types.ADFNode{
				Type: "taskList",
				Attrs: map[string]interface{}{
					"localId": "test",
				},
				Content: []adf_types.ADFNode{
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "TODO",
							"localId": "test-item-0",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 1"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "DONE",
							"localId": "test-item-1",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 2"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.FromMarkdown(tt.markdown, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskListConverter_FromMarkdown_EdgeCases(t *testing.T) {
	converter := NewTaskListConverter()
	ctx := ConversionContext{}

	tests := []struct {
		name     string
		markdown string
		want     adf_types.ADFNode
	}{
		{
			name:     "empty markdown",
			markdown: ``,
			want: adf_types.ADFNode{
				Type:    "taskList",
				Attrs:   map[string]interface{}{},
				Content: nil,
			},
		},
		{
			name:     "whitespace only",
			markdown: `   `,
			want: adf_types.ADFNode{
				Type:    "taskList",
				Attrs:   map[string]interface{}{},
				Content: nil,
			},
		},
		{
			name: "invalid checkbox syntax",
			markdown: `- [?] Invalid task
- [ ] Valid task`,
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
							{Type: "text", Text: "Valid task"},
						},
					},
				},
			},
		},
		{
			name: "non-task list items",
			markdown: `- Regular list item
- [ ] Task item`,
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
							{Type: "text", Text: "Task item"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.FromMarkdown(tt.markdown, ctx)
			require.NoError(t, err)
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
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "TODO",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 1"},
						},
					},
					{
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state": "DONE",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 2"},
						},
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
						Type: "taskItem",
						Attrs: map[string]interface{}{
							"state":   "TODO",
							"localId": "test123-item-0",
						},
						Content: []adf_types.ADFNode{
							{Type: "text", Text: "Task 1"},
						},
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
			roundtrip, err := converter.FromMarkdown(mdResult.Content, ctx)
			require.NoError(t, err)

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
