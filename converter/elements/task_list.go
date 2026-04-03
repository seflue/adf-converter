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

		state := "TODO"
		if item.Attrs != nil {
			if stateValue, exists := item.Attrs["state"]; exists {
				if stateStr, ok := stateValue.(string); ok {
					state = stateStr
				}
			}
		}

		var checkbox string
		switch state {
		case "DONE":
			checkbox = "- [x]"
		case "TODO":
			fallthrough
		default:
			checkbox = "- [ ]"
		}

		taskContent := tc.extractTaskContent(item)

		builder.AppendLine(fmt.Sprintf("%s %s", checkbox, taskContent))
		builder.IncrementConverted()
	}

	result := builder.Build()

	if context.PreserveAttrs && node.Attrs != nil && len(node.Attrs) > 0 {
		wrappedMarkdown, err := tc.wrapTaskListWithXML(result.Content, node.Attrs)
		if err != nil {
			return CreateErrorResult(err.Error(), MarkdownTaskList), err
		}
		result.Content = wrappedMarkdown
	}

	return result, nil
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
func (tc *TaskListConverter) FromMarkdown(markdown string, context ConversionContext) (adf_types.ADFNode, error) {
	lines := strings.Split(strings.TrimSpace(markdown), "\n")

	// Check if this is an XML-wrapped task list
	attrs := make(map[string]interface{})
	var contentLines []string

	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "<taskList") {
		// Extract attributes from XML opening tag
		openTag := strings.TrimSpace(lines[0])
		attrs = tc.parseXMLAttributes(openTag)

		// Find content lines between opening and closing tags
		for i := 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, "</taskList>") {
				break
			}
			contentLines = append(contentLines, lines[i])
		}
	} else {
		// Plain markdown task list
		contentLines = lines
	}

	// Parse task items from markdown content
	taskItems := tc.parseTaskItems(contentLines, attrs)

	return adf_types.ADFNode{
		Type:    "taskList",
		Content: taskItems,
		Attrs:   attrs,
	}, nil
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

// parseXMLAttributes extracts attributes from an XML opening tag
func (tc *TaskListConverter) parseXMLAttributes(xmlTag string) map[string]interface{} {
	return internal.ParseXMLAttributes(xmlTag)
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

//nolint:unparam // error return reserved for future use
func (tc *TaskListConverter) wrapTaskListWithXML(markdownTaskList string, attrs map[string]interface{}) (string, error) {
	var xmlBuilder strings.Builder

	xmlBuilder.WriteString("<taskList")

	if attrs != nil {
		if localId, exists := attrs["localId"]; exists {
			if localIdStr, ok := localId.(string); ok {
				xmlBuilder.WriteString(fmt.Sprintf(` localId="%s"`, localIdStr))
			}
		}

		lines := strings.Split(strings.TrimSpace(markdownTaskList), "\n")
		completed := 0
		total := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "- [x]") {
				completed++
				total++
			} else if strings.HasPrefix(line, "- [ ]") {
				total++
			}
		}

		if total > 0 {
			xmlBuilder.WriteString(fmt.Sprintf(` completed="%d" total="%d"`, completed, total))
		}
	}

	xmlBuilder.WriteString(">\n")

	xmlBuilder.WriteString(markdownTaskList)

	if !strings.HasSuffix(markdownTaskList, "\n") {
		xmlBuilder.WriteString("\n")
	}
	xmlBuilder.WriteString("</taskList>")

	return xmlBuilder.String(), nil
}
