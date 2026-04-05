package placeholder

import (
	"strings"
	"testing"

	"adf-converter/adf_types"
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
		node        adf_types.ADFNode
		expectError bool
	}{
		{
			name: "valid code block",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeCodeBlock,
				Attrs: map[string]interface{}{
					"language": "javascript",
				},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "console.log('hello');"},
				},
			},
			expectError: false,
		},
		{
			name: "valid table",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeTable,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeTableRow,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeTableCell},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:        "invalid empty node type",
			node:        adf_types.ADFNode{},
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
	node := adf_types.ADFNode{
		Type: adf_types.NodeTypeCodeBlock,
		Attrs: map[string]interface{}{
			"language": "go",
		},
		Content: []adf_types.ADFNode{
			{Type: adf_types.NodeTypeText, Text: "fmt.Println(\"Hello\")"},
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
		node     adf_types.ADFNode
		contains string
	}{
		{
			name: "code block with language",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeCodeBlock,
				Attrs: map[string]interface{}{
					"language": "javascript",
				},
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "function test() {\n  return 42;\n}"},
				},
			},
			contains: "Code Block (javascript, 3 lines)",
		},
		{
			name: "table with content",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeTable,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeTableRow,
						Content: []adf_types.ADFNode{
							{
								Type: adf_types.NodeTypeTableCell,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeParagraph,
										Content: []adf_types.ADFNode{
											{Type: adf_types.NodeTypeText, Text: "Header 1"},
										},
									},
								},
							},
							{
								Type: adf_types.NodeTypeTableCell,
								Content: []adf_types.ADFNode{
									{
										Type: adf_types.NodeTypeParagraph,
										Content: []adf_types.ADFNode{
											{Type: adf_types.NodeTypeText, Text: "Header 2"},
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
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypePanel,
				Attrs: map[string]interface{}{
					"panelType": "warning",
				},
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "This is a warning message"},
						},
					},
				},
			},
			contains: "Warning Panel",
		},
		{
			name: "mention",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeMention,
				Attrs: map[string]interface{}{
					"text": "@john.doe",
				},
			},
			contains: "Mention: @john.doe",
		},
		{
			name: "rule",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeRule,
			},
			contains: "Horizontal Rule",
		},
		{
			name: "inlineCard_with_data",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeInlineCard,
				Attrs: map[string]interface{}{
					"data": map[string]interface{}{
						"@type": "Document",
						"name":  "My Document",
					},
				},
			},
			contains: "InlineCard (data object)",
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
	node1 := adf_types.ADFNode{Type: adf_types.NodeTypeCodeBlock}
	node2 := adf_types.ADFNode{Type: adf_types.NodeTypeTable}

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
		node     adf_types.ADFNode
		expected string
	}{
		{
			name:     "simple text node",
			node:     adf_types.ADFNode{Type: adf_types.NodeTypeText, Text: "Hello, world!"},
			expected: "Hello, world!",
		},
		{
			name: "paragraph with multiple text nodes",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeParagraph,
				Content: []adf_types.ADFNode{
					{Type: adf_types.NodeTypeText, Text: "Hello, "},
					{Type: adf_types.NodeTypeText, Text: "world!"},
				},
			},
			expected: "Hello, world!",
		},
		{
			name: "nested structure",
			node: adf_types.ADFNode{
				Type: adf_types.NodeTypeTableCell,
				Content: []adf_types.ADFNode{
					{
						Type: adf_types.NodeTypeParagraph,
						Content: []adf_types.ADFNode{
							{Type: adf_types.NodeTypeText, Text: "Cell content"},
						},
					},
				},
			},
			expected: "Cell content",
		},
		{
			name:     "non-text node with no content",
			node:     adf_types.ADFNode{Type: adf_types.NodeTypeRule},
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

func TestNewManagerWithSession(t *testing.T) {
	// Create a session with some preserved content
	session := &EditSession{
		ID: "test-session-123",
		Preserved: map[string]adf_types.ADFNode{
			"ADF_PLACEHOLDER_001": {Type: adf_types.NodeTypeCodeBlock},
			"ADF_PLACEHOLDER_002": {Type: adf_types.NodeTypeTable},
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

	if node.Type != adf_types.NodeTypeCodeBlock {
		t.Errorf("Expected codeBlock, got %s", node.Type)
	}

	// Verify counter continues from existing count
	newNode := adf_types.ADFNode{Type: adf_types.NodeTypePanel}
	placeholderID, _, err := manager.Store(newNode)
	if err != nil {
		t.Errorf("Failed to store new node: %v", err)
	}

	if placeholderID != "ADF_PLACEHOLDER_003" {
		t.Errorf("Expected ADF_PLACEHOLDER_003, got %s", placeholderID)
	}
}
