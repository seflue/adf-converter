package elements

import (
	"fmt"
	"strings"

	"adf-converter/adf_types"
	"adf-converter/converter/internal"
)

// TaskListConverter implements markdown checkbox conversion for ADF taskList nodes
type TaskListConverter struct{}

func NewTaskListConverter() *TaskListConverter {
	return &TaskListConverter{}
}

func (tc *TaskListConverter) ToMarkdown(node adf_types.ADFNode, context ConversionContext) (EnhancedConversionResult, error) {
	if node.Type != "taskList" {
		return EnhancedConversionResult{}, fmt.Errorf("task list converter can only handle taskList nodes, got: %s", node.Type)
	}

	builder := NewEnhancedConversionResultBuilder(MarkdownTaskList)

	if context.PreserveAttrs && node.Attrs != nil {
		builder.PreserveAttributes(node.Attrs)
	}

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

	if tc.shouldPreserveAttrs(context, node) {
		result.Content = tc.wrapTaskListWithXML(result.Content, node.Attrs)
	}

	return result, nil
}

func (tc *TaskListConverter) getTaskState(item adf_types.ADFNode) string {
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

func (tc *TaskListConverter) extractTaskContent(taskItem adf_types.ADFNode) string {
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

func (tc *TaskListConverter) convertParagraphToMarkdown(paragraph adf_types.ADFNode) string {
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
func (tc *TaskListConverter) FromMarkdown(lines []string, startIndex int, context ConversionContext) (adf_types.ADFNode, int, error) {
	emptyNode := adf_types.ADFNode{Type: "taskList", Attrs: map[string]interface{}{}, Content: nil}

	if startIndex >= len(lines) {
		return emptyNode, 0, nil
	}

	firstLine := strings.TrimSpace(lines[startIndex])
	attrs := make(map[string]interface{})

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

	contentLines := lines[startIndex : startIndex+consumed]
	taskItems := tc.parseTaskItems(contentLines, attrs)
	return adf_types.ADFNode{Type: "taskList", Content: taskItems, Attrs: attrs}, consumed, nil
}

// countXMLTaskListLines counts lines from startIndex to the closing </taskList> tag (inclusive).
func (tc *TaskListConverter) countXMLTaskListLines(lines []string, startIndex int) int {
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "</taskList>") {
			return i - startIndex + 1
		}
	}
	return len(lines) - startIndex // unclosed: consume all remaining
}

// countTaskListLines counts consecutive task-list lines (- [ ]/- [x] and empty lines between them).
func (tc *TaskListConverter) countTaskListLines(lines []string, startIndex int) int {
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

func (tc *TaskListConverter) hasXMLWrapper(lines []string) bool {
	return len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "<taskList")
}

func (tc *TaskListConverter) extractFromXMLWrapper(lines []string) ([]string, map[string]interface{}) {
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
func (tc *TaskListConverter) parseTaskItems(lines []string, taskListAttrs map[string]interface{}) []adf_types.ADFNode {
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
			Attrs: map[string]interface{}{
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

func (tc *TaskListConverter) CanHandle(nodeType ADFNodeType) bool {
	return nodeType == NodeTaskList
}

func (tc *TaskListConverter) GetStrategy() ConversionStrategy {
	return MarkdownTaskList
}

func (tc *TaskListConverter) ValidateInput(input interface{}) error {
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

func (tc *TaskListConverter) shouldPreserveAttrs(context ConversionContext, node adf_types.ADFNode) bool {
	return context.PreserveAttrs && len(node.Attrs) > 0
}

func (tc *TaskListConverter) wrapTaskListWithXML(markdownTaskList string, attrs map[string]interface{}) string {
	var xmlBuilder strings.Builder

	xmlBuilder.WriteString("<taskList")

	if localId, ok := attrs["localId"].(string); ok {
		xmlBuilder.WriteString(fmt.Sprintf(` localId="%s"`, localId))
	}

	completed, total := tc.countTaskStats(markdownTaskList)
	if total > 0 {
		xmlBuilder.WriteString(fmt.Sprintf(` completed="%d" total="%d"`, completed, total))
	}

	xmlBuilder.WriteString(">\n")
	xmlBuilder.WriteString(markdownTaskList)
	if !strings.HasSuffix(markdownTaskList, "\n") {
		xmlBuilder.WriteString("\n")
	}
	xmlBuilder.WriteString("</taskList>")

	return xmlBuilder.String()
}

func (tc *TaskListConverter) countTaskStats(markdown string) (completed, total int) {
	for _, line := range strings.Split(strings.TrimSpace(markdown), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [x]") {
			completed++
			total++
		} else if strings.HasPrefix(line, "- [ ]") {
			total++
		}
	}
	return
}
