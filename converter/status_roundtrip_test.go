package converter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"adf-converter/adf_types"
	"adf-converter/converter"
	"adf-converter/converter/elements"
	"adf-converter/placeholder"
)

func setupStatusTestRegistry() {
	converter.GetGlobalRegistry().Clear()
	converter.RegisterDefaultConverters(
		elements.NewTextConverter(),
		elements.NewHardBreakConverter(),
		elements.NewParagraphConverter(),
		elements.NewHeadingConverter(),
		elements.NewListItemConverter(),
		elements.NewBulletListConverter(),
		elements.NewOrderedListConverter(),
		elements.NewExpandConverter(),
		elements.NewInlineCardConverter(),
		elements.NewEmojiConverter(),
		elements.NewCodeBlockConverter(),
		elements.NewRuleConverter(),
		elements.NewMentionConverter(),
		elements.NewTableConverter(),
		elements.NewPanelConverter(),
		elements.NewDateConverter(),
		elements.NewStatusConverter(),
	)
}

func TestStatusRoundTrip(t *testing.T) {
	setupStatusTestRegistry()
	classifier := converter.NewDefaultClassifier()
	mgr := placeholder.NewManager()

	tests := []struct {
		name      string
		adfJSON   string
		wantMD    string
		wantText  string
		wantColor string
	}{
		{
			name: "basic status blue",
			adfJSON: `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "In Progress", "color": "blue", "localId": "abc-123", "style": ""}}
				]}]
			}`,
			wantMD:    "[status:In Progress|blue]",
			wantText:  "In Progress",
			wantColor: "blue",
		},
		{
			name: "status green",
			adfJSON: `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "Done", "color": "green"}}
				]}]
			}`,
			wantMD:    "[status:Done|green]",
			wantText:  "Done",
			wantColor: "green",
		},
		{
			name: "status neutral",
			adfJSON: `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "TODO", "color": "neutral"}}
				]}]
			}`,
			wantMD:    "[status:TODO|neutral]",
			wantText:  "TODO",
			wantColor: "neutral",
		},
		{
			name: "status purple",
			adfJSON: `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "Review", "color": "purple"}}
				]}]
			}`,
			wantMD:    "[status:Review|purple]",
			wantText:  "Review",
			wantColor: "purple",
		},
		{
			name: "status red",
			adfJSON: `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "Blocked", "color": "red"}}
				]}]
			}`,
			wantMD:    "[status:Blocked|red]",
			wantText:  "Blocked",
			wantColor: "red",
		},
		{
			name: "status yellow",
			adfJSON: `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "Waiting", "color": "yellow"}}
				]}]
			}`,
			wantMD:    "[status:Waiting|yellow]",
			wantText:  "Waiting",
			wantColor: "yellow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse ADF
			var doc adf_types.ADFDocument
			require.NoError(t, json.Unmarshal([]byte(tt.adfJSON), &doc))

			// ADF → Markdown
			md, session, err := converter.ToMarkdown(doc, classifier, mgr)
			require.NoError(t, err)
			assert.Contains(t, md, tt.wantMD)

			// Markdown → ADF
			result, err := converter.FromMarkdown(md, session, mgr)
			require.NoError(t, err)

			// Find status node in result
			require.NotEmpty(t, result.Content, "expected at least one block")
			para := result.Content[0]
			require.NotEmpty(t, para.Content, "expected inline content")

			var statusNode *adf_types.ADFNode
			for i := range para.Content {
				if para.Content[i].Type == adf_types.NodeTypeStatus {
					statusNode = &para.Content[i]
					break
				}
			}
			require.NotNil(t, statusNode, "expected status node in roundtrip result")

			gotText, _ := statusNode.Attrs["text"].(string)
			gotColor, _ := statusNode.Attrs["color"].(string)
			assert.Equal(t, tt.wantText, gotText)
			assert.Equal(t, tt.wantColor, gotColor)

			// localId and style should NOT survive roundtrip
			_, hasLocalId := statusNode.Attrs["localId"]
			_, hasStyle := statusNode.Attrs["style"]
			assert.False(t, hasLocalId, "localId should not survive roundtrip")
			assert.False(t, hasStyle, "style should not survive roundtrip")
		})
	}
}

func TestStatusRoundTrip_MixedParagraph(t *testing.T) {
	setupStatusTestRegistry()
	classifier := converter.NewDefaultClassifier()
	mgr := placeholder.NewManager()

	adfJSON := `{
		"version": 1, "type": "doc",
		"content": [{"type": "paragraph", "content": [
			{"type": "text", "text": "Task "},
			{"type": "status", "attrs": {"text": "In Progress", "color": "blue"}},
			{"type": "text", "text": " is active"}
		]}]
	}`

	var doc adf_types.ADFDocument
	require.NoError(t, json.Unmarshal([]byte(adfJSON), &doc))

	md, session, err := converter.ToMarkdown(doc, classifier, mgr)
	require.NoError(t, err)
	assert.Contains(t, md, "Task [status:In Progress|blue] is active")

	result, err := converter.FromMarkdown(md, session, mgr)
	require.NoError(t, err)

	para := result.Content[0]
	require.Len(t, para.Content, 3, "expected 3 inline nodes")

	assert.Equal(t, adf_types.NodeTypeText, para.Content[0].Type)
	assert.Equal(t, "Task ", para.Content[0].Text)
	assert.Equal(t, adf_types.NodeTypeStatus, para.Content[1].Type)
	assert.Equal(t, adf_types.NodeTypeText, para.Content[2].Type)
	assert.Equal(t, " is active", para.Content[2].Text)
}

func TestStatusRoundTrip_MultipleStatuses(t *testing.T) {
	setupStatusTestRegistry()
	classifier := converter.NewDefaultClassifier()
	mgr := placeholder.NewManager()

	adfJSON := `{
		"version": 1, "type": "doc",
		"content": [{"type": "paragraph", "content": [
			{"type": "status", "attrs": {"text": "Done", "color": "green"}},
			{"type": "text", "text": " and "},
			{"type": "status", "attrs": {"text": "Blocked", "color": "red"}}
		]}]
	}`

	var doc adf_types.ADFDocument
	require.NoError(t, json.Unmarshal([]byte(adfJSON), &doc))

	md, session, err := converter.ToMarkdown(doc, classifier, mgr)
	require.NoError(t, err)
	assert.Contains(t, md, "[status:Done|green]")
	assert.Contains(t, md, "[status:Blocked|red]")

	result, err := converter.FromMarkdown(md, session, mgr)
	require.NoError(t, err)

	para := result.Content[0]
	statusCount := 0
	for _, node := range para.Content {
		if node.Type == adf_types.NodeTypeStatus {
			statusCount++
		}
	}
	assert.Equal(t, 2, statusCount, "expected 2 status nodes")
}

func TestStatusRoundTrip_SpecialChars(t *testing.T) {
	setupStatusTestRegistry()
	classifier := converter.NewDefaultClassifier()
	mgr := placeholder.NewManager()

	tests := []struct {
		name     string
		text     string
		wantText string
	}{
		{name: "space in text", text: "In Progress", wantText: "In Progress"},
		{name: "apostrophe", text: "Won't Fix", wantText: "Won't Fix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adfJSON := `{
				"version": 1, "type": "doc",
				"content": [{"type": "paragraph", "content": [
					{"type": "status", "attrs": {"text": "` + tt.text + `", "color": "blue"}}
				]}]
			}`

			var doc adf_types.ADFDocument
			require.NoError(t, json.Unmarshal([]byte(adfJSON), &doc))

			md, session, err := converter.ToMarkdown(doc, classifier, mgr)
			require.NoError(t, err)

			result, err := converter.FromMarkdown(md, session, mgr)
			require.NoError(t, err)

			para := result.Content[0]
			var statusNode *adf_types.ADFNode
			for i := range para.Content {
				if para.Content[i].Type == adf_types.NodeTypeStatus {
					statusNode = &para.Content[i]
					break
				}
			}
			require.NotNil(t, statusNode)
			gotText, _ := statusNode.Attrs["text"].(string)
			assert.Equal(t, tt.wantText, gotText)
		})
	}
}
