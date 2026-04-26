package placeholder

import (
	"crypto/rand"
	"fmt"
	"strings"

	adf "github.com/seflue/adf-converter/adf/adftypes"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Manager handles the storage and retrieval of preserved ADF content
// during the editing process
type Manager interface {
	Store(node adf.Node) (placeholderID string, preview string, err error)
	Restore(placeholderID string) (adf.Node, error)
	GeneratePreview(node adf.Node) string
	GetSession() *EditSession
	Clear()
	Count() int
}

// EditSession represents a single editing session with preserved content
type EditSession struct {
	ID        string                       `json:"id"`
	Preserved map[string]adf.Node `json:"preserved"`
	Metadata  SessionMetadata              `json:"metadata"`
}

// SessionMetadata contains metadata about the editing session
type SessionMetadata struct {
	OriginalVersion int    `json:"original_version"`
	Timestamp       int64  `json:"timestamp"`
	DocumentType    string `json:"document_type"`
}

// DefaultManager implements the Manager interface
type DefaultManager struct {
	session *EditSession
	counter int
}

// NewManager creates a new placeholder manager with a fresh session
func NewManager() Manager {
	sessionID := generateSessionID()
	return &DefaultManager{
		session: &EditSession{
			ID:        sessionID,
			Preserved: make(map[string]adf.Node),
			Metadata:  SessionMetadata{},
		},
		counter: 0,
	}
}

// NewManagerWithSession creates a manager from an existing session
func NewManagerWithSession(session *EditSession) Manager {
	return &DefaultManager{
		session: session,
		counter: len(session.Preserved),
	}
}

// Store saves an ADF node and returns a placeholder ID and preview text
func (m *DefaultManager) Store(node adf.Node) (string, string, error) {
	if node.Type == "" {
		return "", "", fmt.Errorf("cannot store node with empty type")
	}

	var placeholderID string
	if key := mediaIDKey(node); key != "" {
		placeholderID = fmt.Sprintf("ADF_PLACEHOLDER_%s", key)
	} else {
		m.counter++
		placeholderID = fmt.Sprintf("ADF_PLACEHOLDER_%03d", m.counter)
	}

	m.session.Preserved[placeholderID] = node
	preview := m.GeneratePreview(node)

	return placeholderID, preview, nil
}

// Restore retrieves the original ADF node for a given placeholder ID
func (m *DefaultManager) Restore(placeholderID string) (adf.Node, error) {
	node, exists := m.session.Preserved[placeholderID]
	if !exists {
		return adf.Node{}, fmt.Errorf("placeholder ID %s: %w", placeholderID, ErrPlaceholderNotFound)
	}

	return node, nil
}

// GeneratePreview creates a human-readable preview of an ADF node
func (m *DefaultManager) GeneratePreview(node adf.Node) string {
	return generatePreview(node)
}

// generatePreview creates a human-readable preview of an ADF node.
// Package-level function so both DefaultManager and NullManager can use it.
func generatePreview(node adf.Node) string {
	switch node.Type {
	case adf.NodeTypeCodeBlock:
		return generateCodeBlockPreview(node)
	case adf.NodeTypeTable:
		return generateTablePreview(node)
	case adf.NodeTypePanel:
		return generatePanelPreview(node)
	case adf.NodeTypeBlockquote:
		return generateBlockquotePreview(node)
	case adf.NodeTypeRule:
		return "Horizontal Rule"
	case adf.NodeTypeMediaSingle:
		return generateMediaPreview(node)
	case adf.NodeTypeMediaInline:
		return generateMediaPreview(node)
	case adf.NodeTypeMention:
		return generateMentionPreview(node)
	case adf.NodeTypeDate:
		return generateDatePreview(node)
	case adf.NodeTypeEmoji:
		return generateEmojiPreview(node)
	case adf.NodeTypeStatus:
		return generateStatusPreview(node)
	case adf.NodeTypeInlineCard:
		return "InlineCard (data object)"
	default:
		return fmt.Sprintf("%s (complex content)", cases.Title(language.English).String(string(node.Type)))
	}
}

// generateCodeBlockPreview creates a preview for code blocks
func generateCodeBlockPreview(node adf.Node) string {
	language := "text"
	if node.Attrs != nil {
		if lang, ok := node.Attrs["language"].(string); ok && lang != "" {
			language = lang
		}
	}

	// Extract text content and count lines
	text := extractTextContent(node)
	lines := strings.Split(text, "\n")
	lineCount := len(lines)

	// Get first line for preview
	firstLine := ""
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		firstLine = strings.TrimSpace(lines[0])
		if len(firstLine) > 50 {
			firstLine = firstLine[:47] + "..."
		}
	}

	if firstLine != "" {
		return fmt.Sprintf("Code Block (%s, %d lines): %s", language, lineCount, firstLine)
	}
	return fmt.Sprintf("Code Block (%s, %d lines)", language, lineCount)
}

// generateTablePreview creates a preview for tables
func generateTablePreview(node adf.Node) string {
	rows := 0
	cols := 0

	// Count rows and columns
	for _, row := range node.Content {
		if row.Type == adf.NodeTypeTableRow {
			rows++
			if len(row.Content) > cols {
				cols = len(row.Content)
			}
		}
	}

	// Try to extract first cell content for preview
	preview := ""
	if len(node.Content) > 0 && len(node.Content[0].Content) > 0 {
		firstCell := node.Content[0].Content[0]
		cellText := extractTextContent(firstCell)
		if cellText != "" {
			cellText = strings.ReplaceAll(cellText, "\n", " ")
			if len(cellText) > 30 {
				cellText = cellText[:27] + "..."
			}
			preview = fmt.Sprintf(": %s", cellText)
		}
	}

	return fmt.Sprintf("Table (%dx%d%s)", rows, cols, preview)
}

// generatePanelPreview creates a preview for info/warning/error panels
func generatePanelPreview(node adf.Node) string {
	panelType := "info"
	if node.Attrs != nil {
		if pType, ok := node.Attrs["panelType"].(string); ok && pType != "" {
			panelType = pType
		}
	}

	text := extractTextContent(node)
	if text != "" {
		text = strings.ReplaceAll(text, "\n", " ")
		if len(text) > 50 {
			text = text[:47] + "..."
		}
		return fmt.Sprintf("%s Panel: %s", cases.Title(language.English).String(panelType), text)
	}

	return fmt.Sprintf("%s Panel", cases.Title(language.English).String(panelType))
}

// generateBlockquotePreview creates a preview for blockquotes
func generateBlockquotePreview(node adf.Node) string {
	text := extractTextContent(node)
	if text != "" {
		text = strings.ReplaceAll(text, "\n", " ")
		if len(text) > 50 {
			text = text[:47] + "..."
		}
		return fmt.Sprintf("Quote: %s", text)
	}
	return "Quote"
}

// mediaIDKey returns the first 5 characters of the media node's id attr,
// or an empty string if the id is unavailable or shorter than 5 characters.
func mediaIDKey(node adf.Node) string {
	var id string
	switch node.Type {
	case adf.NodeTypeMediaInline:
		if node.Attrs != nil {
			id, _ = node.Attrs["id"].(string)
		}
	case adf.NodeTypeMediaSingle:
		if len(node.Content) > 0 && node.Content[0].Attrs != nil {
			id, _ = node.Content[0].Attrs["id"].(string)
		}
	}
	if len(id) >= 5 {
		return id[:5]
	}
	return ""
}

// generateMediaPreview creates a preview for media content
func generateMediaPreview(node adf.Node) string {
	var mediaAttrs map[string]any
	var layout string

	switch node.Type {
	case adf.NodeTypeMediaSingle:
		if node.Attrs != nil {
			layout, _ = node.Attrs["layout"].(string)
		}
		if len(node.Content) > 0 {
			mediaAttrs = node.Content[0].Attrs
		}
	case adf.NodeTypeMediaInline:
		mediaAttrs = node.Attrs
	}

	prefix := ""
	if node.Type == adf.NodeTypeMediaInline {
		prefix = "Inline "
	}

	mediaType := "Media"
	var parts []string
	if mediaAttrs != nil {
		if t, ok := mediaAttrs["type"].(string); ok && t != "" {
			mediaType = cases.Title(language.English).String(t)
		}
		w, hasW := mediaAttrs["width"].(float64)
		h, hasH := mediaAttrs["height"].(float64)
		if hasW && hasH {
			parts = append(parts, fmt.Sprintf("%dx%d", int(w), int(h)))
		}
	}
	if layout != "" {
		parts = append(parts, layout)
	}

	if len(parts) > 0 {
		return fmt.Sprintf("%s%s (%s)", prefix, mediaType, strings.Join(parts, ", "))
	}
	return fmt.Sprintf("%s%s", prefix, mediaType)
}

// generateMentionPreview creates a preview for user mentions
func generateMentionPreview(node adf.Node) string {
	if node.Attrs != nil {
		if text, ok := node.Attrs["text"].(string); ok && text != "" {
			return fmt.Sprintf("Mention: %s", text)
		}
	}
	return "User Mention"
}

// generateDatePreview creates a preview for date nodes
func generateDatePreview(node adf.Node) string {
	if node.Attrs != nil {
		if timestamp, ok := node.Attrs["timestamp"].(string); ok && timestamp != "" {
			return fmt.Sprintf("Date: %s", timestamp)
		}
	}
	return "Date"
}

// generateStatusPreview creates a preview for status nodes
func generateStatusPreview(node adf.Node) string {
	if node.Attrs != nil {
		if text, ok := node.Attrs["text"].(string); ok && text != "" {
			return fmt.Sprintf("Status: %s", text)
		}
	}
	return "Status"
}

// generateEmojiPreview creates a preview for emoji nodes
func generateEmojiPreview(node adf.Node) string {
	if node.Attrs != nil {
		if shortName, ok := node.Attrs["shortName"].(string); ok && shortName != "" {
			return fmt.Sprintf("Emoji: %s", shortName)
		}
		if text, ok := node.Attrs["text"].(string); ok && text != "" {
			return fmt.Sprintf("Emoji: %s", text)
		}
	}
	return "Emoji"
}

// Clear removes all preserved content from the session
func (m *DefaultManager) Clear() {
	m.session.Preserved = make(map[string]adf.Node)
	m.counter = 0
}

// Count returns the number of preserved nodes
func (m *DefaultManager) Count() int {
	return len(m.session.Preserved)
}

// GetSession returns the current editing session
func (m *DefaultManager) GetSession() *EditSession {
	return m.session
}

// extractTextContent recursively extracts all text content from an ADF node
func extractTextContent(node adf.Node) string {
	if node.Type == adf.NodeTypeText {
		return node.Text
	}

	var texts []string
	for _, child := range node.Content {
		if childText := extractTextContent(child); childText != "" {
			texts = append(texts, childText)
		}
	}

	return strings.Join(texts, "")
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple counter-based ID if random fails
		return fmt.Sprintf("session_%d", len(bytes))
	}
	return fmt.Sprintf("session_%x", bytes)
}

// GeneratePlaceholderComment creates a Markdown comment for a placeholder
func GeneratePlaceholderComment(placeholderID, preview string) string {
	return fmt.Sprintf("<!-- %s: %s -->", placeholderID, preview)
}

// ParsePlaceholderComment extracts placeholder ID from a Markdown comment
func ParsePlaceholderComment(comment string) (placeholderID string, found bool) {
	// Look for <!-- ADF_PLACEHOLDER_XXX: ... -->
	if !strings.HasPrefix(comment, "<!--") || !strings.HasSuffix(comment, "-->") {
		return "", false
	}

	content := strings.TrimSpace(comment[4 : len(comment)-3])
	parts := strings.SplitN(content, ":", 2)
	if len(parts) < 2 {
		return "", false
	}

	id := strings.TrimSpace(parts[0])
	if strings.HasPrefix(id, "ADF_PLACEHOLDER_") {
		return id, true
	}

	return "", false
}
