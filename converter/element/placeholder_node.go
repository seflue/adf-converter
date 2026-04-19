package element

import (
	"strings"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// parsePlaceholderNode restores preserved content from placeholder comments.
func parsePlaceholderNode(lines []string, manager placeholder.Manager) (*adf_types.ADFNode, int, error) {
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
	if adf_types.IsInlineNode(node.Type) {
		para := adf_types.ADFNode{
			Type:    adf_types.NodeTypeParagraph,
			Content: []adf_types.ADFNode{node},
		}
		return &para, 1, nil
	}

	return &node, 1, nil
}
