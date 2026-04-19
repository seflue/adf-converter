package elements

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter"
	"github.com/seflue/adf-converter/converter/internal"
	"github.com/seflue/adf-converter/converter/internal/convresult"
)

// taskListConverter implements markdown checkbox conversion for ADF taskList nodes
type taskListConverter struct{}

func NewTaskListConverter() converter.ElementConverter {
	return &taskListConverter{}
}

func (tc *taskListConverter) ToMarkdown(node adf_types.ADFNode, context converter.ConversionContext) (converter.EnhancedConversionResult, error) {
	if node.Type != "taskList" {
		return converter.EnhancedConversionResult{}, fmt.Errorf("task list converter can only handle taskList nodes, got: %s", node.Type)
	}

	builder := convresult.NewEnhancedConversionResultBuilder(converter.MarkdownTaskList)

	for _, item := range node.Content {
		if item.Type != "taskItem" {
			builder.AddWarningf("Skipping non-taskItem element: %s", item.Type)
			continue
		}

		checkbox := "- [ ]"
		if tc.getTaskState(item) == "DONE" {
			checkbox = "- [x]"
		}

		taskContent := tc.extractTaskContent(item)

		builder.AppendLine(fmt.Sprintf("%s %s", checkbox, taskContent))
		builder.IncrementConverted()
	}

	result := builder.Build()

	if !strings.HasSuffix(result.Content, "\n\n") {
		result.Content += "\n\n"
	}

	return result, nil
}

func (tc *taskListConverter) getTaskState(item adf_types.ADFNode) string {
	if item.Attrs == nil {
		return "TODO"
	}
	stateValue, exists := item.Attrs["state"]
	if !exists {
		return "TODO"
	}
	stateStr, ok := stateValue.(string)
	if !ok {
		return "TODO"
	}
	return stateStr
}

func (tc *taskListConverter) extractTaskContent(taskItem adf_types.ADFNode) string {
	var content strings.Builder

	for _, contentNode := range taskItem.Content {
		switch contentNode.Type {
		case "paragraph":
			content.WriteString(tc.convertParagraphToMarkdown(contentNode))
		case "text":
			content.WriteString(contentNode.Text)
		}
	}

	return strings.TrimSpace(content.String())
}

func (tc *taskListConverter) convertParagraphToMarkdown(paragraph adf_types.ADFNode) string {
	var result strings.Builder

	for _, textNode := range paragraph.Content {
		if textNode.Type == "text" {
			text := textNode.Text

			for _, mark := range textNode.Marks {
				switch mark.Type {
				case "strong":
					text = "**" + text + "**"
				case "em":
					text = "*" + text + "*"
				case "code":
					text = "`" + text + "`"
				}
			}

			result.WriteString(text)
		}
	}

	return result.String()
}

// FromMarkdown parses markdown task list content into ADF taskList node.
// Supports both plain markdown task lists and XML-wrapped task lists with attributes.
//
// Plain markdown format:
//   - [ ] Task text
//   - [x] Completed task
//
// XML-wrapped format:
//
//	<taskList localId="123" completed="1" total="2">
//	- [ ] Task text
//	- [x] Completed task
//	</taskList>
func (tc *taskListConverter) FromMarkdown(lines []string, startIndex int, context converter.ConversionContext) (adf_types.ADFNode, int, error) {
	emptyNode := adf_types.ADFNode{Type: "taskList", Attrs: map[string]any{}, Content: nil}

	if startIndex >= len(lines) {
		return emptyNode, 0, nil
	}

	firstLine := strings.TrimSpace(lines[startIndex])
	attrs := make(map[string]any)

	// XML-wrapped taskList
	if strings.HasPrefix(firstLine, "<taskList") {
		consumed := tc.countXMLTaskListLines(lines, startIndex)
		contentLines, xmlAttrs := tc.extractFromXMLWrapper(lines[startIndex : startIndex+consumed])
		attrs = xmlAttrs
		taskItems := tc.parseTaskItems(contentLines, attrs)
		return adf_types.ADFNode{Type: "taskList", Content: taskItems, Attrs: attrs}, consumed, nil
	}

	// Plain markdown taskList
	consumed := tc.countTaskListLines(lines, startIndex)
	if consumed == 0 {
		return emptyNode, 0, nil
	}

	attrs["localId"] = generateLocalId()
	contentLines := lines[startIndex : startIndex+consumed]
	taskItems := tc.parseTaskItems(contentLines, attrs)
	return adf_types.ADFNode{Type: "taskList", Content: taskItems, Attrs: attrs}, consumed, nil
}

// countXMLTaskListLines counts lines from startIndex to the closing </taskList> tag (inclusive).
func (tc *taskListConverter) countXMLTaskListLines(lines []string, startIndex int) int {
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "</taskList>") {
			return i - startIndex + 1
		}
	}
	return len(lines) - startIndex // unclosed: consume all remaining
}

// countTaskListLines counts consecutive task-list lines (- [ ]/- [x] and empty lines between them).
func (tc *taskListConverter) countTaskListLines(lines []string, startIndex int) int {
	lastTaskLine := -1
	for i := startIndex; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		isTask := strings.HasPrefix(trimmed, "- [ ]") ||
			strings.HasPrefix(trimmed, "- [x]") ||
			strings.HasPrefix(trimmed, "- [X]")
		if isTask {
			lastTaskLine = i - startIndex + 1
		} else if trimmed == "" && lastTaskLine > 0 {
			// Empty line between tasks — keep scanning
			continue
		} else {
			break
		}
	}
	if lastTaskLine < 0 {
		return 0
	}
	return lastTaskLine
}

func (tc *taskListConverter) extractFromXMLWrapper(lines []string) ([]string, map[string]any) {
	attrs := internal.ParseXMLAttributes(strings.TrimSpace(lines[0]))
	var contentLines []string
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "</taskList>") {
			break
		}
		contentLines = append(contentLines, lines[i])
	}
	return contentLines, attrs
}

// parseTaskItems parses markdown task list lines into ADF taskItem nodes
func (tc *taskListConverter) parseTaskItems(lines []string, taskListAttrs map[string]any) []adf_types.ADFNode {
	var taskItems []adf_types.ADFNode

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Parse checkbox syntax: - [x] or - [ ]
		var state string
		var content string

		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			state = "DONE"
			content = strings.TrimSpace(trimmed[5:])
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			state = "TODO"
			content = strings.TrimSpace(trimmed[5:])
		} else {
			continue // Skip non-task lines
		}

		// Create task item node
		taskItem := adf_types.ADFNode{
			Type: "taskItem",
			Attrs: map[string]any{
				"state": state,
			},
			Content: []adf_types.ADFNode{
				{Type: "text", Text: content},
			},
		}

		// Add localId if available in task list attributes
		if localId, exists := taskListAttrs["localId"]; exists {
			// Generate unique localId for each task item
			taskItem.Attrs["localId"] = fmt.Sprintf("%s-item-%d", localId, len(taskItems))
		}

		taskItems = append(taskItems, taskItem)
	}

	return taskItems
}

func (tc *taskListConverter) CanParseLine(line string) bool {
	return strings.HasPrefix(line, "<taskList") ||
		strings.HasPrefix(line, "- [ ]") ||
		strings.HasPrefix(line, "- [x]") ||
		strings.HasPrefix(line, "- [X]")
}

func (tc *taskListConverter) CanHandle(nodeType converter.ADFNodeType) bool {
	return nodeType == converter.NodeTaskList
}

func (tc *taskListConverter) GetStrategy() converter.ConversionStrategy {
	return converter.MarkdownTaskList
}

func (tc *taskListConverter) ValidateInput(input any) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	switch v := input.(type) {
	case adf_types.ADFNode:
		if v.Type != "taskList" {
			return fmt.Errorf("ADF node must be of type 'taskList', got: %s", v.Type)
		}
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("markdown input cannot be empty")
		}
		return nil
	default:
		return fmt.Errorf("input must be adf_types.ADFNode or string, got: %T", input)
	}
}

func generateLocalId() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
