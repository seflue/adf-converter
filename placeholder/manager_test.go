package placeholder

import (
	"strings"
	"testing"

	adf "github.com/seflue/adf-converter/adf/adftypes"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.Count() != 0 {
		t.Errorf("NewManager() should start with 0 preserved nodes, got %d", manager.Count())
	}

	// Verify session was created
	defaultManager := manager.(*DefaultManager)
	if defaultManager.session == nil {
		t.Error("NewManager() should create a session")
	}

	if defaultManager.session.ID == "" {
		t.Error("NewManager() should create a session with an ID")
	}
}

func TestManager_Store(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		node        adf.Node
		expectError bool
	}{
		{
			name: "valid code block",
			node: adf.Node{
				Type: adf.NodeTypeCodeBlock,
				Attrs: map[string]any{
					"language": "javascript",
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "console.log('hello');"},
				},
			},
			expectError: false,
		},
		{
			name: "valid table",
			node: adf.Node{
				Type: adf.NodeTypeTable,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeTableRow,
						Content: []adf.Node{
							{Type: adf.NodeTypeTableCell},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:        "invalid empty node type",
			node:        adf.Node{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placeholderID, preview, err := manager.Store(tt.node)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if placeholderID == "" {
				t.Error("Expected non-empty placeholder ID")
			}

			if !strings.HasPrefix(placeholderID, "ADF_PLACEHOLDER_") {
				t.Errorf("Placeholder ID should start with ADF_PLACEHOLDER_, got %s", placeholderID)
			}

			if preview == "" {
				t.Error("Expected non-empty preview")
			}

			// Verify we can restore the node
			restored, err := manager.Restore(placeholderID)
			if err != nil {
				t.Errorf("Failed to restore node: %v", err)
			}

			if restored.Type != tt.node.Type {
				t.Errorf("Restored node type %s, want %s", restored.Type, tt.node.Type)
			}
		})
	}
}

func TestManager_Restore(t *testing.T) {
	manager := NewManager()

	// Store a node first
	node := adf.Node{
		Type: adf.NodeTypeCodeBlock,
		Attrs: map[string]any{
			"language": "go",
		},
		Content: []adf.Node{
			{Type: adf.NodeTypeText, Text: "fmt.Println(\"Hello\")"},
		},
	}

	placeholderID, _, err := manager.Store(node)
	if err != nil {
		t.Fatalf("Failed to store node: %v", err)
	}

	// Test restoration
	restored, err := manager.Restore(placeholderID)
	if err != nil {
		t.Errorf("Failed to restore node: %v", err)
	}

	if restored.Type != node.Type {
		t.Errorf("Restored node type %s, want %s", restored.Type, node.Type)
	}

	// Test restoring non-existent placeholder
	_, err = manager.Restore("NONEXISTENT")
	if err == nil {
		t.Error("Expected error when restoring non-existent placeholder")
	}
}

func TestManager_GeneratePreview(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		node     adf.Node
		contains string
	}{
		{
			name: "code block with language",
			node: adf.Node{
				Type: adf.NodeTypeCodeBlock,
				Attrs: map[string]any{
					"language": "javascript",
				},
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "function test() {\n  return 42;\n}"},
				},
			},
			contains: "Code Block (javascript, 3 lines)",
		},
		{
			name: "table with content",
			node: adf.Node{
				Type: adf.NodeTypeTable,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeTableRow,
						Content: []adf.Node{
							{
								Type: adf.NodeTypeTableCell,
								Content: []adf.Node{
									{
										Type: adf.NodeTypeParagraph,
										Content: []adf.Node{
											{Type: adf.NodeTypeText, Text: "Header 1"},
										},
									},
								},
							},
							{
								Type: adf.NodeTypeTableCell,
								Content: []adf.Node{
									{
										Type: adf.NodeTypeParagraph,
										Content: []adf.Node{
											{Type: adf.NodeTypeText, Text: "Header 2"},
										},
									},
								},
							},
						},
					},
				},
			},
			contains: "Table (1x2",
		},
		{
			name: "panel with type",
			node: adf.Node{
				Type: adf.NodeTypePanel,
				Attrs: map[string]any{
					"panelType": "warning",
				},
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "This is a warning message"},
						},
					},
				},
			},
			contains: "Warning Panel",
		},
		{
			name: "mention",
			node: adf.Node{
				Type: adf.NodeTypeMention,
				Attrs: map[string]any{
					"text": "@john.doe",
				},
			},
			contains: "Mention: @john.doe",
		},
		{
			name: "rule",
			node: adf.Node{
				Type: adf.NodeTypeRule,
			},
			contains: "Horizontal Rule",
		},
		{
			name: "inlineCard_with_data",
			node: adf.Node{
				Type: adf.NodeTypeInlineCard,
				Attrs: map[string]any{
					"data": map[string]any{
						"@type": "Document",
						"name":  "My Document",
					},
				},
			},
			contains: "InlineCard (data object)",
		},
		{
			name: "mediaSingle with image type, dimensions, and layout",
			node: adf.Node{
				Type: adf.NodeTypeMediaSingle,
				Attrs: map[string]any{
					"layout": "wide",
				},
				Content: []adf.Node{
					{
						Type: "media",
						Attrs: map[string]any{
							"id":     "a1b2c3d4",
							"type":   "image",
							"width":  float64(1920),
							"height": float64(1080),
						},
					},
				},
			},
			contains: "Image (1920x1080, wide)",
		},
		{
			name: "mediaSingle with file type and no dimensions",
			node: adf.Node{
				Type: adf.NodeTypeMediaSingle,
				Content: []adf.Node{
					{
						Type: "media",
						Attrs: map[string]any{
							"id":   "b2c3d4",
							"type": "file",
						},
					},
				},
			},
			contains: "File",
		},
		{
			name: "mediaInline with image type and dimensions",
			node: adf.Node{
				Type: adf.NodeTypeMediaInline,
				Attrs: map[string]any{
					"id":     "c3d4e5",
					"type":   "image",
					"width":  float64(200),
					"height": float64(150),
				},
			},
			contains: "Inline Image (200x150)",
		},
		{
			name: "mediaInline without attrs",
			node: adf.Node{
				Type: adf.NodeTypeMediaInline,
			},
			contains: "Inline Media",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := manager.GeneratePreview(tt.node)

			if preview == "" {
				t.Error("Expected non-empty preview")
			}

			if !strings.Contains(preview, tt.contains) {
				t.Errorf("Preview %q should contain %q", preview, tt.contains)
			}
		})
	}
}

func TestManager_Clear(t *testing.T) {
	manager := NewManager()

	// Store some nodes
	node1 := adf.Node{Type: adf.NodeTypeCodeBlock}
	node2 := adf.Node{Type: adf.NodeTypeTable}

	_, _, err := manager.Store(node1)
	if err != nil {
		t.Fatalf("Failed to store node1: %v", err)
	}

	_, _, err = manager.Store(node2)
	if err != nil {
		t.Fatalf("Failed to store node2: %v", err)
	}

	if manager.Count() != 2 {
		t.Errorf("Expected 2 stored nodes, got %d", manager.Count())
	}

	// Clear and verify
	manager.Clear()

	if manager.Count() != 0 {
		t.Errorf("Expected 0 stored nodes after clear, got %d", manager.Count())
	}
}

func TestExtractTextContent(t *testing.T) {
	tests := []struct {
		name     string
		node     adf.Node
		expected string
	}{
		{
			name:     "simple text node",
			node:     adf.Node{Type: adf.NodeTypeText, Text: "Hello, world!"},
			expected: "Hello, world!",
		},
		{
			name: "paragraph with multiple text nodes",
			node: adf.Node{
				Type: adf.NodeTypeParagraph,
				Content: []adf.Node{
					{Type: adf.NodeTypeText, Text: "Hello, "},
					{Type: adf.NodeTypeText, Text: "world!"},
				},
			},
			expected: "Hello, world!",
		},
		{
			name: "nested structure",
			node: adf.Node{
				Type: adf.NodeTypeTableCell,
				Content: []adf.Node{
					{
						Type: adf.NodeTypeParagraph,
						Content: []adf.Node{
							{Type: adf.NodeTypeText, Text: "Cell content"},
						},
					},
				},
			},
			expected: "Cell content",
		},
		{
			name:     "non-text node with no content",
			node:     adf.Node{Type: adf.NodeTypeRule},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTextContent(tt.node)
			if result != tt.expected {
				t.Errorf("extractTextContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGeneratePlaceholderComment(t *testing.T) {
	placeholderID := "ADF_PLACEHOLDER_001"
	preview := "Code Block (javascript, 5 lines)"

	comment := GeneratePlaceholderComment(placeholderID, preview)
	expected := "<!-- ADF_PLACEHOLDER_001: Code Block (javascript, 5 lines) -->"

	if comment != expected {
		t.Errorf("GeneratePlaceholderComment() = %q, want %q", comment, expected)
	}
}

func TestParsePlaceholderComment(t *testing.T) {
	tests := []struct {
		name       string
		comment    string
		expectedID string
		expectedOK bool
	}{
		{
			name:       "valid placeholder comment",
			comment:    "<!-- ADF_PLACEHOLDER_001: Code Block (javascript, 5 lines) -->",
			expectedID: "ADF_PLACEHOLDER_001",
			expectedOK: true,
		},
		{
			name:       "valid placeholder comment with extra spaces",
			comment:    "<!--   ADF_PLACEHOLDER_123  :  Table (2x3)   -->",
			expectedID: "ADF_PLACEHOLDER_123",
			expectedOK: true,
		},
		{
			name:       "invalid - not a comment",
			comment:    "ADF_PLACEHOLDER_001: Code Block",
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "invalid - wrong prefix",
			comment:    "<!-- WRONG_PREFIX_001: Code Block -->",
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "invalid - no colon",
			comment:    "<!-- ADF_PLACEHOLDER_001 Code Block -->",
			expectedID: "",
			expectedOK: false,
		},
		{
			name:       "empty comment",
			comment:    "<!-- -->",
			expectedID: "",
			expectedOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := ParsePlaceholderComment(tt.comment)

			if ok != tt.expectedOK {
				t.Errorf("ParsePlaceholderComment() ok = %v, want %v", ok, tt.expectedOK)
			}

			if id != tt.expectedID {
				t.Errorf("ParsePlaceholderComment() id = %q, want %q", id, tt.expectedID)
			}
		})
	}
}

func TestManager_Store_MediaKeys(t *testing.T) {
	tests := []struct {
		name              string
		node              adf.Node
		expectedKeyPrefix string
	}{
		{
			name: "mediaSingle uses first 5 chars of media child id",
			node: adf.Node{
				Type: adf.NodeTypeMediaSingle,
				Attrs: map[string]any{
					"layout": "center",
				},
				Content: []adf.Node{
					{
						Type: "media",
						Attrs: map[string]any{
							"id":         "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
							"type":       "image",
							"collection": "col-1",
						},
					},
				},
			},
			expectedKeyPrefix: "ADF_PLACEHOLDER_a1b2c",
		},
		{
			name: "mediaInline uses first 5 chars of id attr",
			node: adf.Node{
				Type: adf.NodeTypeMediaInline,
				Attrs: map[string]any{
					"id":         "xyz99-abcd-ef12-3456-7890abcdef12",
					"type":       "file",
					"collection": "col-2",
				},
			},
			expectedKeyPrefix: "ADF_PLACEHOLDER_xyz99",
		},
		{
			name: "mediaSingle without id falls back to counter",
			node: adf.Node{
				Type: adf.NodeTypeMediaSingle,
				Content: []adf.Node{
					{
						Type:  "media",
						Attrs: map[string]any{"type": "image"},
					},
				},
			},
			expectedKeyPrefix: "ADF_PLACEHOLDER_001",
		},
		{
			name: "mediaInline without id falls back to counter",
			node: adf.Node{
				Type: adf.NodeTypeMediaInline,
			},
			expectedKeyPrefix: "ADF_PLACEHOLDER_002",
		},
	}

	manager := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placeholderID, _, err := manager.Store(tt.node)
			if err != nil {
				t.Fatalf("Store() unexpected error: %v", err)
			}
			if placeholderID != tt.expectedKeyPrefix {
				t.Errorf("Store() placeholderID = %q, want %q", placeholderID, tt.expectedKeyPrefix)
			}
		})
	}
}

func TestNewManagerWithSession(t *testing.T) {
	// Create a session with some preserved content
	session := &EditSession{
		ID: "test-session-123",
		Preserved: map[string]adf.Node{
			"ADF_PLACEHOLDER_001": {Type: adf.NodeTypeCodeBlock},
			"ADF_PLACEHOLDER_002": {Type: adf.NodeTypeTable},
		},
		Metadata: SessionMetadata{
			OriginalVersion: 1,
			DocumentType:    "doc",
		},
	}

	manager := NewManagerWithSession(session)

	if manager.Count() != 2 {
		t.Errorf("Expected 2 preserved nodes, got %d", manager.Count())
	}

	// Verify we can restore existing placeholders
	node, err := manager.Restore("ADF_PLACEHOLDER_001")
	if err != nil {
		t.Errorf("Failed to restore existing placeholder: %v", err)
	}

	if node.Type != adf.NodeTypeCodeBlock {
		t.Errorf("Expected codeBlock, got %s", node.Type)
	}

	// Verify counter continues from existing count
	newNode := adf.Node{Type: adf.NodeTypePanel}
	placeholderID, _, err := manager.Store(newNode)
	if err != nil {
		t.Errorf("Failed to store new node: %v", err)
	}

	if placeholderID != "ADF_PLACEHOLDER_003" {
		t.Errorf("Expected ADF_PLACEHOLDER_003, got %s", placeholderID)
	}
}
