package adf_types

// compatibility.go provides compatibility types and wrappers for go-atlassian integration

// Type aliases for go-atlassian compatibility
// These allow our types to be used directly with go-atlassian APIs while maintaining
// our superior design (value types, document structure, semantic naming)

// CommentNodeScheme is a type alias for ADFNode to provide go-atlassian compatibility
// This allows our ADFNode to be used wherever go-atlassian expects CommentNodeScheme
type CommentNodeScheme = ADFNode

// MarkScheme is a type alias for ADFMark to provide go-atlassian compatibility
type MarkScheme = ADFMark

// Convenience wrapper types for API integration

// CommentPayload represents a comment payload structure compatible with go-atlassian
// while using our superior ADFDocument for the body content
type CommentPayload struct {
	Visibility *CommentVisibility `json:"visibility,omitempty"` // The visibility of the comment
	Body       *ADFDocument       `json:"body,omitempty"`       // The body using our document structure
}

// CommentVisibility represents the visibility settings for a comment
type CommentVisibility struct {
	Type  string `json:"type,omitempty"`  // The type of visibility (e.g., "group", "role")
	Value string `json:"value,omitempty"` // The value (group name, role name, etc.)
}

// IssueComment represents a complete issue comment with metadata
type IssueComment struct {
	Self         string             `json:"self,omitempty"`         // The self link of the comment
	ID           string             `json:"id,omitempty"`           // The ID of the comment
	Author       *User              `json:"author,omitempty"`       // The author of the comment
	RenderedBody string             `json:"renderedBody,omitempty"` // The rendered body (useful for debugging)
	Body         *ADFDocument       `json:"body,omitempty"`         // The body using our document structure
	JSDPublic    bool               `json:"jsdPublic,omitempty"`    // Whether the comment is public
	UpdateAuthor *User              `json:"updateAuthor,omitempty"` // The author of the last update
	Created      string             `json:"created,omitempty"`      // The creation time
	Updated      string             `json:"updated,omitempty"`      // The last update time
	Visibility   *CommentVisibility `json:"visibility,omitempty"`   // The visibility of the comment
}

// User represents a user in the Jira system
type User struct {
	AccountID    string `json:"accountId,omitempty"`    // The account ID
	DisplayName  string `json:"displayName,omitempty"`  // The display name
	EmailAddress string `json:"emailAddress,omitempty"` // The email address
}

// CommentPage represents a paginated list of comments
type CommentPage struct {
	StartAt    int             `json:"startAt,omitempty"`    // The start index
	MaxResults int             `json:"maxResults,omitempty"` // The maximum results per page
	Total      int             `json:"total,omitempty"`      // The total number of comments
	Comments   []*IssueComment `json:"comments,omitempty"`   // The comments
}

// Conversion helpers

// ToCommentNodeScheme converts an ADFDocument to a CommentNodeScheme (ADFNode)
// This is useful when go-atlassian APIs expect a single node rather than a document
func (d ADFDocument) ToCommentNodeScheme() CommentNodeScheme {
	return CommentNodeScheme{
		Type:    d.Type,
		Content: d.Content,
		Attrs: map[string]interface{}{
			"version": d.Version,
		},
	}
}

// FromCommentNodeScheme converts a CommentNodeScheme (ADFNode) to an ADFDocument
// This handles the reverse conversion when receiving data from go-atlassian
func FromCommentNodeScheme(node CommentNodeScheme) ADFDocument {
	doc := ADFDocument{
		Type:    NodeTypeDoc,
		Version: 1,
		Content: []ADFNode{},
	}

	// If it's already a document node, use its content directly
	if node.Type == NodeTypeDoc {
		doc.Content = node.Content
		if version, ok := node.GetAttribute("version"); ok {
			if v, ok := version.(int); ok {
				doc.Version = v
			}
		}
	} else {
		// If it's any other node, wrap it in the document
		doc.Content = []ADFNode{node}
	}

	return doc
}

// NewCommentPayload creates a new comment payload with the given body
func NewCommentPayload(body ADFDocument) *CommentPayload {
	return &CommentPayload{
		Body: &body,
	}
}

// NewCommentPayloadWithVisibility creates a comment payload with visibility settings
func NewCommentPayloadWithVisibility(body ADFDocument, visibilityType, visibilityValue string) *CommentPayload {
	return &CommentPayload{
		Body: &body,
		Visibility: &CommentVisibility{
			Type:  visibilityType,
			Value: visibilityValue,
		},
	}
}
