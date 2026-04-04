package inline

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/forPelevin/gomoji"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"adf-converter/adf_types"
	"adf-converter/placeholder"
)

// PlaceholderPattern matches placeholder HTML comments
// Example: <!-- ADF_PLACEHOLDER_001: Emoji: :white_check_mark: -->
var PlaceholderPattern = regexp.MustCompile(`<!--\s*(ADF_PLACEHOLDER_\d+):\s*([^>]+?)\s*-->`)

// SpanColorPattern extracts color value from span style attribute
var SpanColorPattern = regexp.MustCompile(`color:\s*([^;"]+)`)

// DatePattern matches inline date syntax: [date:2025-04-04]
var DatePattern = regexp.MustCompile(`\[date:(\d{4}-\d{2}-\d{2})\]`)

// StatusPattern matches inline status syntax: [status:Text|color]
// Lenient on color (any alphabetic word) for forward compatibility
var StatusPattern = regexp.MustCompile(`\[status:([^|\]]+)\|([a-zA-Z]+)\]`)

// statusInfo holds extracted status text and color for marker restoration
type statusInfo struct {
	text  string
	color string
}

// ParseContent parses inline markdown formatting into ADF text nodes with marks
// This is the unified inline content parser used by all element converters
// Uses goldmark for CommonMark-compliant parsing
//
// If manager is provided, placeholder comments will be restored to their original ADF nodes
func ParseContent(markdown string) ([]adf_types.ADFNode, error) {
	return ParseContentWithPlaceholders(markdown, nil)
}

// ParseContentWithPlaceholders parses inline markdown and restores placeholder nodes
func ParseContentWithPlaceholders(markdown string, manager placeholder.Manager) ([]adf_types.ADFNode, error) {
	if markdown == "" {
		return []adf_types.ADFNode{}, nil
	}

	// Step 1: Extract placeholder comments before goldmark parsing
	// Goldmark treats HTML comments specially and may drop or mangle them
	placeholders, cleanedMarkdown := extractPlaceholders(markdown)

	// Step 1b: Extract date patterns before goldmark parsing
	// Goldmark may break [date:...] at bracket boundaries
	dates, cleanedMarkdown := extractDatePatterns(cleanedMarkdown)

	// Step 1c: Extract status patterns before goldmark parsing
	// Goldmark would interpret [status:...] as a link reference
	statuses, cleanedMarkdown := extractStatusPatterns(cleanedMarkdown)

	// Step 2: Parse with goldmark (now without HTML comments or date patterns)
	source := []byte(cleanedMarkdown)
	parser := goldmark.New()
	doc := parser.Parser().Parse(text.NewReader(source))

	// Extract inline content from paragraph
	para := doc.FirstChild()
	if para == nil {
		return []adf_types.ADFNode{}, nil
	}

	// Convert goldmark AST → ADF nodes directly
	result, err := convertInlineAST(para.FirstChild(), source, []adf_types.ADFMark{})
	if err != nil {
		return nil, err
	}

	// Step 3: Restore placeholder nodes from manager
	if manager != nil && len(placeholders) > 0 {
		result = restorePlaceholders(result, placeholders, manager)
	}

	// Step 3b: Restore date nodes
	if len(dates) > 0 {
		result = restoreDateNodes(result, dates)
	}

	// Step 3c: Restore status nodes
	if len(statuses) > 0 {
		result = restoreStatusNodes(result, statuses)
	}

	// Merge consecutive text nodes with identical marks
	// (goldmark splits text at special chars like ! for image syntax)
	return mergeConsecutiveTextNodes(result), nil
}

// extractPlaceholders finds and removes placeholder comments from markdown
// Returns: map of marker → placeholderID, and cleaned markdown with markers
func extractPlaceholders(markdown string) (map[string]string, string) {
	placeholders := make(map[string]string)
	counter := 0

	cleaned := PlaceholderPattern.ReplaceAllStringFunc(markdown, func(match string) string {
		// Extract placeholder ID from comment
		submatches := PlaceholderPattern.FindStringSubmatch(match)
		if len(submatches) >= 2 {
			placeholderID := submatches[1]
			// Create a unique marker that goldmark won't interpret as formatting
			// Use a marker with characters that goldmark won't treat as emphasis/strong
			marker := fmt.Sprintf("{{PLACEHOLDER-%d}}", counter)
			placeholders[marker] = placeholderID
			counter++
			return marker
		}
		return match
	})

	return placeholders, cleaned
}

// restorePlaceholders replaces marker text nodes with actual ADF nodes from manager
func restorePlaceholders(nodes []adf_types.ADFNode, placeholders map[string]string, manager placeholder.Manager) []adf_types.ADFNode {
	result := make([]adf_types.ADFNode, 0, len(nodes))

	for _, node := range nodes {
		if node.Type == adf_types.NodeTypeText {
			// Check if text contains a placeholder marker
			restored := restoreTextWithPlaceholders(node, placeholders, manager)
			result = append(result, restored...)
		} else {
			result = append(result, node)
		}
	}

	return result
}

// detectAndConvertEmojis detects unicode emojis in text and converts to ADF emoji nodes
// Returns a slice of nodes containing text nodes and emoji nodes
func detectAndConvertEmojis(text string, parentMarks []adf_types.ADFMark) []adf_types.ADFNode {
	// Check if text contains any emojis
	if !gomoji.ContainsEmoji(text) {
		// No emojis - return single text node
		textNode := adf_types.NewTextNode(text)
		textNode.Marks = append(textNode.Marks, parentMarks...)
		return []adf_types.ADFNode{textNode}
	}

	// Text contains emojis - split into text and emoji nodes
	var nodes []adf_types.ADFNode
	remaining := text

	for len(remaining) > 0 {
		// Find all emojis in remaining text
		emojis := gomoji.FindAll(remaining)
		if len(emojis) == 0 {
			// No more emojis - add remaining text
			if len(remaining) > 0 {
				textNode := adf_types.NewTextNode(remaining)
				textNode.Marks = append(textNode.Marks, parentMarks...)
				nodes = append(nodes, textNode)
			}
			break
		}

		// Find position of first emoji character
		firstEmoji := emojis[0]
		emojiIndex := strings.Index(remaining, firstEmoji.Character)

		// Add text before emoji (if any)
		if emojiIndex > 0 {
			beforeText := remaining[:emojiIndex]
			textNode := adf_types.NewTextNode(beforeText)
			textNode.Marks = append(textNode.Marks, parentMarks...)
			nodes = append(nodes, textNode)
		}

		// Create emoji node
		emojiNode := createEmojiNode(firstEmoji)
		nodes = append(nodes, emojiNode)

		// Update remaining text (skip past the emoji)
		endIndex := emojiIndex + len(firstEmoji.Character)
		if endIndex < len(remaining) {
			remaining = remaining[endIndex:]
		} else {
			remaining = ""
		}
	}

	return nodes
}

// createEmojiNode creates an ADF emoji node from gomoji emoji info
func createEmojiNode(emoji gomoji.Emoji) adf_types.ADFNode {
	attrs := map[string]interface{}{
		"text": emoji.Character,
	}

	// Add shortName if available
	if emoji.Slug != "" {
		// gomoji uses slug like "thumbs-up", convert to shortName like ":thumbs_up:"
		shortName := ":" + strings.ReplaceAll(emoji.Slug, "-", "_") + ":"
		attrs["shortName"] = shortName
	}

	// Add id (code point) if available
	if emoji.CodePoint != "" {
		// gomoji provides code point like "U+1F44D", extract hex without "U+"
		codePoint := strings.TrimPrefix(emoji.CodePoint, "U+")
		attrs["id"] = strings.ToLower(codePoint)
	}

	return adf_types.ADFNode{
		Type:  adf_types.NodeTypeEmoji,
		Attrs: attrs,
	}
}

// restoreTextWithPlaceholders splits text node on markers and restores placeholders
func restoreTextWithPlaceholders(textNode adf_types.ADFNode, placeholders map[string]string, manager placeholder.Manager) []adf_types.ADFNode {
	text := textNode.Text
	var result []adf_types.ADFNode

	// Find all markers in the text
	for marker, placeholderID := range placeholders {
		if strings.Contains(text, marker) {
			// Split text on marker
			parts := strings.Split(text, marker)

			// Add text before marker
			if len(parts[0]) > 0 {
				beforeNode := adf_types.NewTextNode(parts[0])
				beforeNode.Marks = textNode.Marks
				result = append(result, beforeNode)
			}

			// Restore original node from placeholder
			if originalNode, err := manager.Restore(placeholderID); err == nil {
				result = append(result, originalNode)
			}

			// Handle remaining text after marker
			if len(parts) > 1 {
				remaining := strings.Join(parts[1:], marker)
				if len(remaining) > 0 {
					// Recursively process remaining text (might have more markers)
					afterNode := adf_types.NewTextNode(remaining)
					afterNode.Marks = textNode.Marks
					moreRestored := restoreTextWithPlaceholders(afterNode, placeholders, manager)
					result = append(result, moreRestored...)
				}
			}

			// Return early - we processed this marker
			return result
		}
	}

	// No markers found, return original node
	return []adf_types.ADFNode{textNode}
}

// extractColorFromSpan parses the color value from a <span style="color: ..."> tag
func extractColorFromSpan(content string) (string, bool) {
	matches := SpanColorPattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return "", false
	}
	return strings.TrimSpace(matches[1]), true
}

// rawHTMLContent extracts the text content from a RawHTML node
func rawHTMLContent(n *ast.RawHTML, source []byte) string {
	var buf strings.Builder
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		buf.Write(source[seg.Start:seg.Stop])
	}
	return buf.String()
}

// isOpeningHTMLTag checks if a RawHTML node contains an opening tag for the given element
func isOpeningHTMLTag(n *ast.RawHTML, source []byte, tagName string) bool {
	content := rawHTMLContent(n, source)
	return content == "<"+tagName+">" || strings.HasPrefix(content, "<"+tagName+" ")
}

// isClosingHTMLTag checks if a RawHTML node contains a closing tag for the given element
func isClosingHTMLTag(n *ast.RawHTML, source []byte, tagName string) bool {
	return rawHTMLContent(n, source) == "</"+tagName+">"
}

// collectNodesUntilClosingTag walks sibling nodes between opening and closing HTML tags,
// applies the given mark to all collected content, and returns the node after the closing tag.
func collectNodesUntilClosingTag(start ast.Node, source []byte, tagName string, mark adf_types.ADFMark, parentMarks []adf_types.ADFMark) ([]adf_types.ADFNode, ast.Node, error) {
	allMarks := append(parentMarks, mark)
	var nodes []adf_types.ADFNode

	current := start
	for current != nil {
		if rawHTML, ok := current.(*ast.RawHTML); ok && isClosingHTMLTag(rawHTML, source, tagName) {
			return nodes, current.NextSibling(), nil
		}

		innerNodes, err := convertSingleInlineNode(current, source, allMarks)
		if err != nil {
			return nil, nil, err
		}
		nodes = append(nodes, innerNodes...)
		current = current.NextSibling()
	}

	return nil, nil, fmt.Errorf("no closing </%s> tag found", tagName)
}

// convertSingleInlineNode converts a single goldmark inline node to ADF nodes.
// Unlike convertInlineAST, this does NOT iterate over siblings.
func convertSingleInlineNode(node ast.Node, source []byte, parentMarks []adf_types.ADFMark) ([]adf_types.ADFNode, error) {
	switch n := node.(type) {
	case *ast.Text:
		segment := n.Segment
		txt := string(source[segment.Start:segment.Stop])
		return detectAndConvertEmojis(txt, parentMarks), nil

	case *ast.Emphasis:
		mark := getEmphasisMark(n)
		childMarks := append(parentMarks, mark)
		return convertInlineAST(n.FirstChild(), source, childMarks)

	case *ast.CodeSpan:
		if n.FirstChild() != nil {
			if txtNode, ok := n.FirstChild().(*ast.Text); ok {
				segment := txtNode.Segment
				txt := string(source[segment.Start:segment.Stop])
				codeNode := adf_types.NewTextNode(txt)
				codeMark := adf_types.ADFMark{Type: adf_types.MarkTypeCode}
				codeNode.Marks = append([]adf_types.ADFMark{codeMark}, parentMarks...)
				return []adf_types.ADFNode{codeNode}, nil
			}
		}
		return nil, nil

	case *ast.Link:
		href := string(n.Destination)
		linkText := extractLinkText(n.FirstChild(), source)

		// Mention: accountid: scheme → ADF mention node
		if mentionNode, ok := parseMentionLink(href, linkText); ok {
			return []adf_types.ADFNode{mentionNode}, nil
		}

		if linkText == href {
			inlineCardNode := adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"url": href,
				},
			}
			return []adf_types.ADFNode{inlineCardNode}, nil
		}

		attrs := map[string]interface{}{
			"href": href,
		}
		if len(n.Title) > 0 {
			attrs["title"] = string(n.Title)
		}
		linkMark := adf_types.NewMark(adf_types.MarkTypeLink, attrs)
		childMarks := append(parentMarks, linkMark)
		return convertInlineAST(n.FirstChild(), source, childMarks)

	case *ast.RawHTML:
		if isOpeningHTMLTag(n, source, "u") {
			underlineMark := adf_types.ADFMark{Type: adf_types.MarkTypeUnderline}
			collected, nextNode, err := collectNodesUntilClosingTag(node.NextSibling(), source, "u", underlineMark, parentMarks)
			if err == nil {
				if nextNode != nil {
					remaining, err := convertInlineAST(nextNode, source, parentMarks)
					if err != nil {
						return nil, err
					}
					collected = append(collected, remaining...)
				}
				return collected, nil
			}
		}
		if isOpeningHTMLTag(n, source, "span") {
			content := rawHTMLContent(n, source)
			if color, ok := extractColorFromSpan(content); ok {
				colorMark := adf_types.NewMark(adf_types.MarkTypeTextColor, map[string]interface{}{
					"color": color,
				})
				collected, nextNode, err := collectNodesUntilClosingTag(node.NextSibling(), source, "span", colorMark, parentMarks)
				if err == nil {
					if nextNode != nil {
						remaining, err := convertInlineAST(nextNode, source, parentMarks)
						if err != nil {
							return nil, err
						}
						collected = append(collected, remaining...)
					}
					return collected, nil
				}
			}
		}
		return nil, nil

	default:
		return nil, nil
	}
}

// convertInlineAST walks goldmark inline nodes and converts to ADF
// marks parameter accumulates marks from parent nodes (for nested formatting)
func convertInlineAST(node ast.Node, source []byte, parentMarks []adf_types.ADFMark) ([]adf_types.ADFNode, error) {
	var nodes []adf_types.ADFNode

	for current := node; current != nil; current = current.NextSibling() {
		result, err := convertSingleInlineNode(current, source, parentMarks)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, result...)

		// RawHTML with tag-pairing already processed remaining siblings via recursion
		if rawHTML, ok := current.(*ast.RawHTML); ok {
			isPairedHTMLTag := isOpeningHTMLTag(rawHTML, source, "u") || isOpeningHTMLTag(rawHTML, source, "span")
			if isPairedHTMLTag {
				return nodes, nil
			}
		}
	}

	return nodes, nil
}

func getEmphasisMark(n *ast.Emphasis) adf_types.ADFMark {
	switch n.Level {
	case 1:
		return adf_types.ADFMark{Type: adf_types.MarkTypeEm}
	case 2:
		return adf_types.ADFMark{Type: adf_types.MarkTypeStrong}
	default:
		return adf_types.ADFMark{Type: adf_types.MarkTypeEm}
	}
}

// extractLinkText extracts plain text from link children for InlineCard detection
func extractLinkText(node ast.Node, source []byte) string {
	var buf strings.Builder
	for n := node; n != nil; n = n.NextSibling() {
		if txt, ok := n.(*ast.Text); ok {
			segment := txt.Segment
			buf.Write(source[segment.Start:segment.Stop])
		}
	}
	return buf.String()
}

// mergeConsecutiveTextNodes combines adjacent text nodes with identical marks
// This is needed because goldmark splits text at special characters like ! and [
func mergeConsecutiveTextNodes(nodes []adf_types.ADFNode) []adf_types.ADFNode {
	if len(nodes) <= 1 {
		return nodes
	}

	merged := make([]adf_types.ADFNode, 0, len(nodes))
	current := nodes[0]

	for i := 1; i < len(nodes); i++ {
		next := nodes[i]

		// Can merge if both are text nodes with identical marks
		if current.Type == adf_types.NodeTypeText && next.Type == adf_types.NodeTypeText && marksEqual(current.Marks, next.Marks) {
			// Merge text content
			current.Text += next.Text
		} else {
			// Can't merge, append current and start new
			merged = append(merged, current)
			current = next
		}
	}

	// Append final node
	merged = append(merged, current)
	return merged
}

// marksEqual checks if two mark slices are identical
func marksEqual(a, b []adf_types.ADFMark) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type {
			return false
		}
		// For marks with attributes (like links), check attrs equality
		if !attrsEqual(a[i].Attrs, b[i].Attrs) {
			return false
		}
	}
	return true
}

// attrsEqual checks if two attribute maps are equal
func attrsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// parseMentionLink checks if a link destination is an accountid: mention
// and returns the corresponding ADF mention node
func parseMentionLink(href, linkText string) (adf_types.ADFNode, bool) {
	const prefix = "accountid:"
	if !strings.HasPrefix(href, prefix) {
		return adf_types.ADFNode{}, false
	}

	remainder := strings.TrimPrefix(href, prefix)

	// Split id from query parameters
	id := remainder
	var queryString string
	if idx := strings.Index(remainder, "?"); idx >= 0 {
		id = remainder[:idx]
		queryString = remainder[idx+1:]
	}

	if id == "" {
		return adf_types.ADFNode{}, false
	}

	attrs := map[string]interface{}{
		"id":   id,
		"text": linkText,
	}

	// Parse query parameters for optional attrs
	if queryString != "" {
		params, err := url.ParseQuery(queryString)
		if err == nil {
			if v := params.Get("accessLevel"); v != "" {
				attrs["accessLevel"] = v
			}
			if v := params.Get("userType"); v != "" {
				attrs["userType"] = v
			}
		}
	}

	return adf_types.ADFNode{
		Type:  adf_types.NodeTypeMention,
		Attrs: attrs,
	}, true
}

// extractDatePatterns finds and removes [date:YYYY-MM-DD] patterns from markdown
// Returns: map of marker → ISO date string, and cleaned markdown
func extractDatePatterns(markdown string) (map[string]string, string) {
	dates := make(map[string]string)
	counter := 0

	cleaned := DatePattern.ReplaceAllStringFunc(markdown, func(match string) string {
		submatches := DatePattern.FindStringSubmatch(match)
		if len(submatches) >= 2 {
			dateStr := submatches[1]
			marker := fmt.Sprintf("{{DATE-%d}}", counter)
			dates[marker] = dateStr
			counter++
			return marker
		}
		return match
	})

	return dates, cleaned
}

// restoreDateNodes replaces date marker text nodes with ADF date nodes
func restoreDateNodes(nodes []adf_types.ADFNode, dates map[string]string) []adf_types.ADFNode {
	result := make([]adf_types.ADFNode, 0, len(nodes))

	for _, node := range nodes {
		if node.Type == adf_types.NodeTypeText {
			restored := restoreTextWithDates(node, dates)
			result = append(result, restored...)
		} else {
			result = append(result, node)
		}
	}

	return result
}

// restoreTextWithDates splits text node on date markers and creates ADF date nodes
func restoreTextWithDates(textNode adf_types.ADFNode, dates map[string]string) []adf_types.ADFNode {
	nodeText := textNode.Text
	var result []adf_types.ADFNode

	for marker, dateStr := range dates {
		if strings.Contains(nodeText, marker) {
			parts := strings.Split(nodeText, marker)

			if len(parts[0]) > 0 {
				beforeNode := adf_types.NewTextNode(parts[0])
				beforeNode.Marks = textNode.Marks
				result = append(result, beforeNode)
			}

			millis := dateToMillisUnchecked(dateStr)
			dateNode := adf_types.ADFNode{
				Type: adf_types.NodeTypeDate,
				Attrs: map[string]interface{}{
					"timestamp": millis,
				},
			}
			result = append(result, dateNode)

			if len(parts) > 1 {
				remaining := strings.Join(parts[1:], marker)
				if len(remaining) > 0 {
					afterNode := adf_types.NewTextNode(remaining)
					afterNode.Marks = textNode.Marks
					moreRestored := restoreTextWithDates(afterNode, dates)
					result = append(result, moreRestored...)
				}
			}

			return result
		}
	}

	return []adf_types.ADFNode{textNode}
}

// dateToMillisUnchecked converts ISO date to millis string (pattern already validated by regex)
func dateToMillisUnchecked(dateStr string) string {
	t, _ := time.Parse("2006-01-02", dateStr)
	ms := t.UTC().Unix() * 1000
	return strconv.FormatInt(ms, 10)
}

// extractStatusPatterns replaces [status:Text|color] with markers before goldmark parsing
func extractStatusPatterns(markdown string) (map[string]statusInfo, string) {
	statuses := make(map[string]statusInfo)
	counter := 0

	cleaned := StatusPattern.ReplaceAllStringFunc(markdown, func(match string) string {
		submatches := StatusPattern.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			marker := fmt.Sprintf("{{STATUS-%d}}", counter)
			statuses[marker] = statusInfo{text: submatches[1], color: submatches[2]}
			counter++
			return marker
		}
		return match
	})

	return statuses, cleaned
}

// restoreStatusNodes replaces status marker text nodes with ADF status nodes
func restoreStatusNodes(nodes []adf_types.ADFNode, statuses map[string]statusInfo) []adf_types.ADFNode {
	result := make([]adf_types.ADFNode, 0, len(nodes))

	for _, node := range nodes {
		if node.Type == adf_types.NodeTypeText {
			restored := restoreTextWithStatuses(node, statuses)
			result = append(result, restored...)
		} else {
			result = append(result, node)
		}
	}

	return result
}

// restoreTextWithStatuses splits text node on status markers and creates ADF status nodes
func restoreTextWithStatuses(textNode adf_types.ADFNode, statuses map[string]statusInfo) []adf_types.ADFNode {
	nodeText := textNode.Text
	var result []adf_types.ADFNode

	for marker, info := range statuses {
		if strings.Contains(nodeText, marker) {
			parts := strings.Split(nodeText, marker)

			if len(parts[0]) > 0 {
				beforeNode := adf_types.NewTextNode(parts[0])
				beforeNode.Marks = textNode.Marks
				result = append(result, beforeNode)
			}

			statusNode := adf_types.ADFNode{
				Type: adf_types.NodeTypeStatus,
				Attrs: map[string]interface{}{
					"text":  info.text,
					"color": info.color,
				},
			}
			result = append(result, statusNode)

			if len(parts) > 1 {
				remaining := strings.Join(parts[1:], marker)
				if len(remaining) > 0 {
					afterNode := adf_types.NewTextNode(remaining)
					afterNode.Marks = textNode.Marks
					moreRestored := restoreTextWithStatuses(afterNode, statuses)
					result = append(result, moreRestored...)
				}
			}

			return result
		}
	}

	return []adf_types.ADFNode{textNode}
}
