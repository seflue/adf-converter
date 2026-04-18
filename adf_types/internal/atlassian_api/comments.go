// Package atlassian_api holds go-atlassian compatibility wrappers around
// adf_types. Moved here to keep the public adf_types surface minimal.
package atlassian_api

import "github.com/seflue/adf-converter/adf_types"

// CommentNodeScheme is a type alias for adf_types.ADFNode to provide
// go-atlassian compatibility.
type CommentNodeScheme = adf_types.ADFNode

// MarkScheme is a type alias for adf_types.ADFMark to provide
// go-atlassian compatibility.
type MarkScheme = adf_types.ADFMark

// CommentPayload represents a comment payload structure compatible with go-atlassian
// while using our ADFDocument for the body content.
type CommentPayload struct {
	Visibility *CommentVisibility    `json:"visibility,omitempty"`
	Body       *adf_types.ADFDocument `json:"body,omitempty"`
}

// CommentVisibility represents the visibility settings for a comment.
type CommentVisibility struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// IssueComment represents a complete issue comment with metadata.
type IssueComment struct {
	Self         string                 `json:"self,omitempty"`
	ID           string                 `json:"id,omitempty"`
	Author       *User                  `json:"author,omitempty"`
	RenderedBody string                 `json:"renderedBody,omitempty"`
	Body         *adf_types.ADFDocument `json:"body,omitempty"`
	JSDPublic    bool                   `json:"jsdPublic,omitempty"`
	UpdateAuthor *User                  `json:"updateAuthor,omitempty"`
	Created      string                 `json:"created,omitempty"`
	Updated      string                 `json:"updated,omitempty"`
	Visibility   *CommentVisibility     `json:"visibility,omitempty"`
}

// User represents a user in the Jira system.
type User struct {
	AccountID    string `json:"accountId,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// CommentPage represents a paginated list of comments.
type CommentPage struct {
	StartAt    int             `json:"startAt,omitempty"`
	MaxResults int             `json:"maxResults,omitempty"`
	Total      int             `json:"total,omitempty"`
	Comments   []*IssueComment `json:"comments,omitempty"`
}

// ToCommentNodeScheme converts an ADFDocument to a CommentNodeScheme (ADFNode).
func ToCommentNodeScheme(d adf_types.ADFDocument) CommentNodeScheme {
	return adf_types.ADFNode{
		Type:    d.Type,
		Content: d.Content,
		Attrs: map[string]any{
			"version": d.Version,
		},
	}
}

// FromCommentNodeScheme converts a CommentNodeScheme (ADFNode) to an ADFDocument.
func FromCommentNodeScheme(node CommentNodeScheme) adf_types.ADFDocument {
	doc := adf_types.ADFDocument{
		Type:    adf_types.NodeTypeDoc,
		Version: 1,
		Content: []adf_types.ADFNode{},
	}

	if node.Type == adf_types.NodeTypeDoc {
		doc.Content = node.Content
		if version, ok := node.GetAttribute("version"); ok {
			if v, ok := version.(int); ok {
				doc.Version = v
			}
		}
	} else {
		doc.Content = []adf_types.ADFNode{node}
	}

	return doc
}

// NewCommentPayload creates a new comment payload with the given body.
func NewCommentPayload(body adf_types.ADFDocument) *CommentPayload {
	return &CommentPayload{
		Body: &body,
	}
}

// NewCommentPayloadWithVisibility creates a comment payload with visibility settings.
func NewCommentPayloadWithVisibility(body adf_types.ADFDocument, visibilityType, visibilityValue string) *CommentPayload {
	return &CommentPayload{
		Body: &body,
		Visibility: &CommentVisibility{
			Type:  visibilityType,
			Value: visibilityValue,
		},
	}
}
