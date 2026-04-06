package inline

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/forPelevin/gomoji"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"

	"adf-converter/adf_types"
)

func TestParseContent_PlainText(t *testing.T) {
	nodes, err := ParseContent("plain text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Type != adf_types.NodeTypeText {
		t.Errorf("expected text node, got %s", nodes[0].Type)
	}

	if nodes[0].Text != "plain text" {
		t.Errorf("expected 'plain text', got '%s'", nodes[0].Text)
	}
}

func TestParseContent_BoldText(t *testing.T) {
	nodes, err := ParseContent("**bold**")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "bold" {
		t.Errorf("expected 'bold', got '%s'", nodes[0].Text)
	}

	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != "strong" {
		t.Errorf("expected strong mark")
	}
}

func TestParseContent_ItalicText(t *testing.T) {
	nodes, err := ParseContent("*italic*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "italic" {
		t.Errorf("expected 'italic', got '%s'", nodes[0].Text)
	}

	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != "em" {
		t.Errorf("expected em mark")
	}
}

func TestParseContent_CodeText(t *testing.T) {
	nodes, err := ParseContent("`code`")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Text != "code" {
		t.Errorf("expected 'code', got '%s'", nodes[0].Text)
	}

	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != "code" {
		t.Errorf("expected code mark")
	}
}

func TestParseContent_Links(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantText string
		wantHref string
	}{
		{
			name:     "simple link",
			input:    "[text](url)",
			wantText: "text",
			wantHref: "url",
		},
		{
			name:     "InlineCard (same text and url)",
			input:    "[https://example.com](https://example.com)",
			wantText: "",
			wantHref: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d", len(nodes))
			}

			if tt.input == "[https://example.com](https://example.com)" {
				// Should be InlineCard
				if nodes[0].Type != adf_types.NodeTypeInlineCard {
					t.Errorf("expected InlineCard node, got %s", nodes[0].Type)
				}
			} else {
				// Should be text with link mark
				if nodes[0].Text != tt.wantText {
					t.Errorf("expected text '%s', got '%s'", tt.wantText, nodes[0].Text)
				}
			}
		})
	}
}

func TestParseContent_MixedFormatting(t *testing.T) {
	nodes, err := ParseContent("plain **bold** *italic* `code` plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 7 {
		t.Fatalf("expected 7 nodes, got %d", len(nodes))
	}

	// Verify sequence
	if nodes[0].Text != "plain " {
		t.Errorf("expected 'plain ', got '%s'", nodes[0].Text)
	}
	if nodes[1].Text != "bold" || len(nodes[1].Marks) == 0 || nodes[1].Marks[0].Type != "strong" {
		t.Errorf("expected bold node")
	}
	if nodes[2].Text != " " {
		t.Errorf("expected ' ', got '%s'", nodes[2].Text)
	}
	if nodes[3].Text != "italic" || len(nodes[3].Marks) == 0 || nodes[3].Marks[0].Type != "em" {
		t.Errorf("expected italic node")
	}
	if nodes[4].Text != " " {
		t.Errorf("expected ' ', got '%s'", nodes[4].Text)
	}
	if nodes[5].Text != "code" || len(nodes[5].Marks) == 0 || nodes[5].Marks[0].Type != "code" {
		t.Errorf("expected code node")
	}
	if nodes[6].Text != " plain" {
		t.Errorf("expected ' plain', got '%s'", nodes[6].Text)
	}
}

func TestParseContent_NestedFormatting(t *testing.T) {
	// Note: Current implementation doesn't support true nesting,
	// but we should handle adjacent formatting correctly
	nodes, err := ParseContent("**bold*italic***")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should parse as bold "bold*italic*" (asterisks inside are treated as content)
	if len(nodes) == 0 {
		t.Fatal("expected at least one node")
	}
}

// Comprehensive goldmark-based tests
func TestParseContent_Goldmark(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, nodes []adf_types.ADFNode)
	}{
		{
			name:  "nested_bold_italic",
			input: "This is **bold with *italic* inside** text",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				// Goldmark produces 5 text nodes:
				// 1. "This is " (plain)
				// 2. "bold with " (strong)
				// 3. "italic" (strong+em)
				// 4. " inside" (strong)
				// 5. " text" (plain)
				if len(nodes) != 5 {
					t.Fatalf("expected 5 nodes, got %d", len(nodes))
				}
				// Check for proper nesting
				foundNested := false
				for _, node := range nodes {
					if len(node.Marks) == 2 {
						// Should have both strong and em marks
						foundNested = true
						hasStrong := false
						hasEm := false
						for _, mark := range node.Marks {
							if mark.Type == "strong" {
								hasStrong = true
							}
							if mark.Type == "em" {
								hasEm = true
							}
						}
						if !hasStrong || !hasEm {
							t.Errorf("nested node should have both strong and em marks")
						}
					}
				}
				if !foundNested {
					t.Errorf("should have a node with nested marks")
				}
			},
		},
		{
			name:  "inline_card_detection",
			input: "Check [https://example.com](https://example.com) out",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				// Should have InlineCard node
				found := false
				for _, node := range nodes {
					if node.Type == adf_types.NodeTypeInlineCard {
						found = true
						if node.Attrs["url"] != "https://example.com" {
							t.Errorf("InlineCard should have url attribute")
						}
					}
				}
				if !found {
					t.Errorf("Should have InlineCard node")
				}
			},
		},
		{
			name:  "link_with_formatted_text",
			input: "[**bold link**](https://example.com)",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				if len(nodes) != 1 {
					t.Fatalf("expected 1 node, got %d", len(nodes))
				}
				// Should have both link and strong marks
				if len(nodes[0].Marks) != 2 {
					t.Errorf("should have 2 marks (link + strong), got %d", len(nodes[0].Marks))
				}
			},
		},
		{
			name:  "text_node_merging",
			input: "Hello, world!",
			validate: func(t *testing.T, nodes []adf_types.ADFNode) {
				// Should be single merged text node (goldmark splits at !)
				if len(nodes) != 1 {
					t.Errorf("expected 1 merged node, got %d", len(nodes))
				}
				if nodes[0].Text != "Hello, world!" {
					t.Errorf("expected 'Hello, world!', got '%s'", nodes[0].Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, nodes)
		})
	}
}

func TestParseContent_Mention(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantType     string
		wantID       string
		wantText     string
		wantAccess   string
		wantUserType string
	}{
		{
			name:     "basic mention",
			input:    "[@john.doe](accountid:abc123)",
			wantType: adf_types.NodeTypeMention,
			wantID:   "abc123",
			wantText: "@john.doe",
		},
		{
			name:         "mention with query params",
			input:        "[@john.doe](accountid:abc123?accessLevel=CONTAINER&userType=DEFAULT)",
			wantType:     adf_types.NodeTypeMention,
			wantID:       "abc123",
			wantText:     "@john.doe",
			wantAccess:   "CONTAINER",
			wantUserType: "DEFAULT",
		},
		{
			name:       "mention with only accessLevel",
			input:      "[@jane](accountid:xyz?accessLevel=APPLICATION)",
			wantType:   adf_types.NodeTypeMention,
			wantID:     "xyz",
			wantText:   "@jane",
			wantAccess: "APPLICATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
			}

			node := nodes[0]
			if node.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, node.Type)
			}

			if id, _ := node.Attrs["id"].(string); id != tt.wantID {
				t.Errorf("expected id %q, got %q", tt.wantID, id)
			}

			if text, _ := node.Attrs["text"].(string); text != tt.wantText {
				t.Errorf("expected text %q, got %q", tt.wantText, text)
			}

			if tt.wantAccess != "" {
				if al, _ := node.Attrs["accessLevel"].(string); al != tt.wantAccess {
					t.Errorf("expected accessLevel %q, got %q", tt.wantAccess, al)
				}
			}

			if tt.wantUserType != "" {
				if ut, _ := node.Attrs["userType"].(string); ut != tt.wantUserType {
					t.Errorf("expected userType %q, got %q", tt.wantUserType, ut)
				}
			}
		})
	}
}

func TestParseContent_MentionInText(t *testing.T) {
	nodes, err := ParseContent("Hello [@john](accountid:abc) world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	if nodes[0].Type != adf_types.NodeTypeText || nodes[0].Text != "Hello " {
		t.Errorf("expected text 'Hello ', got type=%s text=%q", nodes[0].Type, nodes[0].Text)
	}
	if nodes[1].Type != adf_types.NodeTypeMention {
		t.Errorf("expected mention node, got %s", nodes[1].Type)
	}
	if nodes[2].Type != adf_types.NodeTypeText || nodes[2].Text != " world" {
		t.Errorf("expected text ' world', got type=%s text=%q", nodes[2].Type, nodes[2].Text)
	}
}

func TestParseContent_BoldWrappingUnderline(t *testing.T) {
	nodes, err := ParseContent("**<u>bold underlined</u>**")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "bold underlined" {
		t.Errorf("expected 'bold underlined', got '%s'", nodes[0].Text)
	}

	hasUnderline := false
	hasStrong := false
	for _, mark := range nodes[0].Marks {
		if mark.Type == adf_types.MarkTypeUnderline {
			hasUnderline = true
		}
		if mark.Type == adf_types.MarkTypeStrong {
			hasStrong = true
		}
	}
	if !hasUnderline || !hasStrong {
		t.Errorf("expected underline+strong marks, got %v", nodes[0].Marks)
	}
}

func TestParseContent_UnderlineWithBoldInside(t *testing.T) {
	nodes, err := ParseContent("<u>**bold underlined**</u>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "bold underlined" {
		t.Errorf("expected 'bold underlined', got '%s'", nodes[0].Text)
	}

	hasUnderline := false
	hasStrong := false
	for _, mark := range nodes[0].Marks {
		if mark.Type == adf_types.MarkTypeUnderline {
			hasUnderline = true
		}
		if mark.Type == adf_types.MarkTypeStrong {
			hasStrong = true
		}
	}
	if !hasUnderline || !hasStrong {
		t.Errorf("expected underline+strong marks, got %v", nodes[0].Marks)
	}
}

func TestParseContent_UnderlineWithSurroundingText(t *testing.T) {
	nodes, err := ParseContent("before <u>underlined</u> after")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "before " {
		t.Errorf("expected 'before ', got '%s'", nodes[0].Text)
	}
	if nodes[1].Text != "underlined" || len(nodes[1].Marks) != 1 || nodes[1].Marks[0].Type != adf_types.MarkTypeUnderline {
		t.Errorf("expected underlined text with mark, got text=%q marks=%v", nodes[1].Text, nodes[1].Marks)
	}
	if nodes[2].Text != " after" {
		t.Errorf("expected ' after', got '%s'", nodes[2].Text)
	}
}

func TestParseContent_SimpleUnderline(t *testing.T) {
	nodes, err := ParseContent("<u>underlined</u>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "underlined" {
		t.Errorf("expected 'underlined', got '%s'", nodes[0].Text)
	}
	if len(nodes[0].Marks) != 1 || nodes[0].Marks[0].Type != adf_types.MarkTypeUnderline {
		t.Errorf("expected underline mark, got %v", nodes[0].Marks)
	}
}

func TestParseContent_SimpleTextColor(t *testing.T) {
	nodes, err := ParseContent(`<span style="color: #ff0000">red text</span>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "red text" {
		t.Errorf("expected 'red text', got %q", nodes[0].Text)
	}
	if len(nodes[0].Marks) != 1 {
		t.Fatalf("expected 1 mark, got %d", len(nodes[0].Marks))
	}
	if nodes[0].Marks[0].Type != adf_types.MarkTypeTextColor {
		t.Errorf("expected textColor mark, got %s", nodes[0].Marks[0].Type)
	}
	color, ok := nodes[0].Marks[0].Attrs["color"].(string)
	if !ok || color != "#ff0000" {
		t.Errorf("expected color #ff0000, got %v", nodes[0].Marks[0].Attrs["color"])
	}
}

func TestParseContent_TextColorWithSurroundingText(t *testing.T) {
	nodes, err := ParseContent(`before <span style="color: #00ff00">green</span> after`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "before " {
		t.Errorf("expected 'before ', got %q", nodes[0].Text)
	}
	if nodes[1].Text != "green" || len(nodes[1].Marks) != 1 || nodes[1].Marks[0].Type != adf_types.MarkTypeTextColor {
		t.Errorf("expected green text with textColor mark, got text=%q marks=%v", nodes[1].Text, nodes[1].Marks)
	}
	if nodes[2].Text != " after" {
		t.Errorf("expected ' after', got %q", nodes[2].Text)
	}
}

func TestParseContent_TextColorWithBoldInside(t *testing.T) {
	nodes, err := ParseContent(`<span style="color: #ff0000">**bold red**</span>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "bold red" {
		t.Errorf("expected 'bold red', got %q", nodes[0].Text)
	}

	markTypes := make(map[string]bool)
	for _, mark := range nodes[0].Marks {
		markTypes[mark.Type] = true
	}
	if !markTypes[adf_types.MarkTypeTextColor] || !markTypes[adf_types.MarkTypeStrong] {
		t.Errorf("expected textColor+strong marks, got %v", nodes[0].Marks)
	}
}

func TestParseContent_BoldWrappingTextColor(t *testing.T) {
	nodes, err := ParseContent(`**<span style="color: #ff0000">bold red</span>**`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Text != "bold red" {
		t.Errorf("expected 'bold red', got %q", nodes[0].Text)
	}

	markTypes := make(map[string]bool)
	for _, mark := range nodes[0].Marks {
		markTypes[mark.Type] = true
	}
	if !markTypes[adf_types.MarkTypeTextColor] || !markTypes[adf_types.MarkTypeStrong] {
		t.Errorf("expected textColor+strong marks, got %v", nodes[0].Marks)
	}
}

func TestParseContent_DateNode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDate string
	}{
		{
			name:     "basic date",
			input:    "[date:2025-04-04]",
			wantDate: "2025-04-04",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
			}

			node := nodes[0]
			if node.Type != adf_types.NodeTypeDate {
				t.Errorf("expected date node, got %s", node.Type)
			}

			timestamp, ok := node.Attrs["timestamp"].(string)
			if !ok {
				t.Fatal("expected timestamp attribute")
			}

			ms, _ := strconv.ParseInt(timestamp, 10, 64)
			gotDate := time.Unix(ms/1000, 0).UTC().Format("2006-01-02")
			if gotDate != tt.wantDate {
				t.Errorf("expected date %q, got %q (timestamp: %s)", tt.wantDate, gotDate, timestamp)
			}
		})
	}
}

func TestParseContent_DateMixedWithText(t *testing.T) {
	nodes, err := ParseContent("Due: [date:2025-04-04] done")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	if nodes[0].Type != adf_types.NodeTypeText || nodes[0].Text != "Due: " {
		t.Errorf("expected text 'Due: ', got type=%s text=%q", nodes[0].Type, nodes[0].Text)
	}
	if nodes[1].Type != adf_types.NodeTypeDate {
		t.Errorf("expected date node, got %s", nodes[1].Type)
	}
	if nodes[2].Type != adf_types.NodeTypeText || nodes[2].Text != " done" {
		t.Errorf("expected text ' done', got type=%s text=%q", nodes[2].Type, nodes[2].Text)
	}
}

func TestParseContent_StatusNode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantText  string
		wantColor string
	}{
		{
			name:      "basic status",
			input:     "[status:In Progress|blue]",
			wantText:  "In Progress",
			wantColor: "blue",
		},
		{
			name:      "status neutral",
			input:     "[status:TODO|neutral]",
			wantText:  "TODO",
			wantColor: "neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d: %+v", len(nodes), nodes)
			}

			node := nodes[0]
			if node.Type != adf_types.NodeTypeStatus {
				t.Errorf("expected status node, got %s", node.Type)
			}

			gotText, _ := node.Attrs["text"].(string)
			if gotText != tt.wantText {
				t.Errorf("expected text %q, got %q", tt.wantText, gotText)
			}

			gotColor, _ := node.Attrs["color"].(string)
			if gotColor != tt.wantColor {
				t.Errorf("expected color %q, got %q", tt.wantColor, gotColor)
			}
		})
	}
}

func TestParseContent_StatusMixedWithText(t *testing.T) {
	nodes, err := ParseContent("Task [status:In Progress|blue] is active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %+v", len(nodes), nodes)
	}

	if nodes[0].Type != adf_types.NodeTypeText || nodes[0].Text != "Task " {
		t.Errorf("expected text 'Task ', got type=%s text=%q", nodes[0].Type, nodes[0].Text)
	}
	if nodes[1].Type != adf_types.NodeTypeStatus {
		t.Errorf("expected status node, got %s", nodes[1].Type)
	}
	if nodes[2].Type != adf_types.NodeTypeText || nodes[2].Text != " is active" {
		t.Errorf("expected text ' is active', got type=%s text=%q", nodes[2].Type, nodes[2].Text)
	}
}

func TestParseContent_Strikethrough(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantText  string
		wantMarks []string
	}{
		{
			name:      "simple strikethrough",
			input:     "~~deleted~~",
			wantText:  "deleted",
			wantMarks: []string{adf_types.MarkTypeStrike},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(nodes) != 1 {
				t.Fatalf("expected 1 node, got %d", len(nodes))
			}

			if nodes[0].Text != tt.wantText {
				t.Errorf("expected text %q, got %q", tt.wantText, nodes[0].Text)
			}

			if len(nodes[0].Marks) != len(tt.wantMarks) {
				t.Fatalf("expected %d marks, got %d: %v", len(tt.wantMarks), len(nodes[0].Marks), nodes[0].Marks)
			}

			for i, wantMark := range tt.wantMarks {
				if nodes[0].Marks[i].Type != wantMark {
					t.Errorf("mark[%d]: expected %q, got %q", i, wantMark, nodes[0].Marks[i].Type)
				}
			}
		})
	}
}

func TestParseContent_StrikethroughWithSurroundingText(t *testing.T) {
	nodes, err := ParseContent("before ~~deleted~~ after")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	if nodes[0].Type != adf_types.NodeTypeText || nodes[0].Text != "before " {
		t.Errorf("expected text 'before ', got type=%s text=%q", nodes[0].Type, nodes[0].Text)
	}

	if nodes[1].Text != "deleted" {
		t.Errorf("expected 'deleted', got %q", nodes[1].Text)
	}
	if len(nodes[1].Marks) != 1 || nodes[1].Marks[0].Type != adf_types.MarkTypeStrike {
		t.Errorf("expected strike mark, got %v", nodes[1].Marks)
	}

	if nodes[2].Type != adf_types.NodeTypeText || nodes[2].Text != " after" {
		t.Errorf("expected text ' after', got type=%s text=%q", nodes[2].Type, nodes[2].Text)
	}
}

func TestParseContent_SubScript(t *testing.T) {
	nodes, err := ParseContent("H<sub>2</sub>O")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	if nodes[0].Text != "H" {
		t.Errorf("expected 'H', got %q", nodes[0].Text)
	}

	if nodes[1].Text != "2" {
		t.Errorf("expected '2', got %q", nodes[1].Text)
	}
	if len(nodes[1].Marks) != 1 || nodes[1].Marks[0].Type != adf_types.MarkTypeSubsup {
		t.Fatalf("expected subsup mark, got %v", nodes[1].Marks)
	}
	if nodes[1].Marks[0].Attrs["type"] != "sub" {
		t.Errorf("expected subsup type 'sub', got %v", nodes[1].Marks[0].Attrs["type"])
	}

	if nodes[2].Text != "O" {
		t.Errorf("expected 'O', got %q", nodes[2].Text)
	}
}

func TestParseContent_SuperScript(t *testing.T) {
	nodes, err := ParseContent("x<sup>2</sup>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0].Text != "x" {
		t.Errorf("expected 'x', got %q", nodes[0].Text)
	}

	if nodes[1].Text != "2" {
		t.Errorf("expected '2', got %q", nodes[1].Text)
	}
	if len(nodes[1].Marks) != 1 || nodes[1].Marks[0].Type != adf_types.MarkTypeSubsup {
		t.Fatalf("expected subsup mark, got %v", nodes[1].Marks)
	}
	if nodes[1].Marks[0].Attrs["type"] != "sup" {
		t.Errorf("expected subsup type 'sup', got %v", nodes[1].Marks[0].Attrs["type"])
	}
}

func TestParseContent_StatusNotMatched(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "missing pipe", input: "[status:NoPipe]"},
		{name: "empty text", input: "[status:|blue]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := ParseContent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, node := range nodes {
				if node.Type == adf_types.NodeTypeStatus {
					t.Errorf("should not produce status node for input %q", tt.input)
				}
			}
		})
	}
}

func TestConvertInlineAST_DoesNotMutateParentMarksBackingArray(t *testing.T) {
	// Regression test for the slice-aliasing bug in mark propagation (ac-0017).
	// append(parentMarks, mark) silently mutates the backing array of the caller's
	// slice when cap > len. Fix: three-index slice parentMarks[:n:n] caps capacity.
	parentMarks := make([]adf_types.ADFMark, 1, 2)
	parentMarks[0] = adf_types.ADFMark{Type: adf_types.MarkTypeStrong}

	// Place a sentinel at the extra-capacity slot to detect mutation.
	backing := parentMarks[:2]
	backing[1] = adf_types.ADFMark{Type: "sentinel"}

	// Parse "*italic*": the Emphasis case does append(parentMarks, emMark).
	// Without the fix, emMark is written to backing[1], overwriting the sentinel.
	source := []byte("*italic*")
	parser := goldmark.New(goldmark.WithExtensions(extension.Strikethrough))
	doc := parser.Parser().Parse(text.NewReader(source))
	para := doc.FirstChild()
	if para == nil {
		t.Fatal("expected paragraph node")
	}

	if _, err := convertInlineAST(para.FirstChild(), source, parentMarks); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backing[1].Type != "sentinel" {
		t.Errorf("backing array mutated: got %q, want %q — append(parentMarks, mark) aliased the caller's backing array",
			backing[1].Type, "sentinel")
	}
}

func TestCreateEmojiNode_ShortNameFromSlug(t *testing.T) {
	emoji := gomoji.Emoji{
		Character: "✅",
		Slug:      "white-check-mark",
		CodePoint: "U+2705",
	}
	node := createEmojiNode(emoji)
	shortName, ok := node.Attrs["shortName"].(string)
	if !ok || shortName == "" {
		t.Fatalf("expected shortName from slug, got %v", node.Attrs["shortName"])
	}
	if shortName != ":white_check_mark:" {
		t.Errorf("expected ':white_check_mark:', got %q", shortName)
	}
}

func TestCreateEmojiNode_ShortNameFallbackToUnicodeName(t *testing.T) {
	emoji := gomoji.Emoji{
		Character:   "✅",
		Slug:        "",
		UnicodeName: "white check mark",
		CodePoint:   "U+2705",
	}
	node := createEmojiNode(emoji)
	shortName, ok := node.Attrs["shortName"].(string)
	if !ok || shortName == "" {
		t.Fatalf("expected shortName fallback from UnicodeName, got %v", node.Attrs["shortName"])
	}
	if !strings.HasPrefix(shortName, ":") || !strings.HasSuffix(shortName, ":") {
		t.Errorf("shortName should be colon-wrapped, got %q", shortName)
	}
}
