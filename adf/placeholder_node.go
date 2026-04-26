package adf

import (
	"strings"

	"github.com/seflue/adf-converter/placeholder"
)

// parsePlaceholderNode restores preserved content from placeholder comments.
func parsePlaceholderNode(lines []string, manager placeholder.Manager) (*Node, int, error) {
	if len(lines) == 0 {
		return nil, 1, nil
	}

	line := strings.TrimSpace(lines[0])
	placeholderID, found := placeholder.ParsePlaceholderComment(line)
	if !found {
		return nil, 1, nil
	}

	node, err := manager.Restore(placeholderID)
	if err != nil {
		// Placeholder was deleted from markdown - skip it (allows intentional deletion)
		return nil, 1, nil
	}

	// Inline nodes live inside paragraphs; wrap them to restore the original structure
	if IsInlineNode(node.Type) {
		para := Node{
			Type:    NodeTypeParagraph,
			Content: []Node{node},
		}
		return &para, 1, nil
	}

	return &node, 1, nil
}
