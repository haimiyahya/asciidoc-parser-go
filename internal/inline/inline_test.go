// Package inline provides tests for inline AsciiDoc parsing.
package inline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		nodeType NodeType
		expected string
	}{
		{NodeText, "Text"},
		{NodeBold, "Bold"},
		{NodeItalic, "Italic"},
		{NodeMonospace, "Monospace"},
		{NodeSubscript, "Subscript"},
		{NodeSuperscript, "Superscript"},
		{NodeLink, "Link"},
		{NodeImage, "Image"},
		{NodeKbd, "Kbd"},
		{NodeBtn, "Btn"},
		{NodeMenu, "Menu"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.nodeType.String())
		})
	}
}

func TestParseBold(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []*Node
	}{
		{
			name: "constrained bold",
			source: "**bold**",
			expected: []*Node{
				{Type: NodeBold, Text: "bold"},
			},
		},
		{
			name: "unconstrained bold",
			source: "*bold*",
			expected: []*Node{
				{Type: NodeBold, Text: "bold"},
			},
		},
		{
			name: "bold with spaces",
			source: "**bold text**",
			expected: []*Node{
				{Type: NodeBold, Text: "bold text"},
			},
		},
		{
			name: "plain text",
			source: "plain text",
			expected: []*Node{
				{Type: NodeText, Text: "plain text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Equal(t, len(tt.expected), len(nodes), "node count mismatch")
			for i, expectedNode := range tt.expected {
				actualNode := nodes[i]
				assert.Equal(t, expectedNode.Type, actualNode.Type)
				assert.Equal(t, expectedNode.Text, actualNode.Text)
			}
		})
	}
}

func TestParseItalic(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []*Node
	}{
		{
			name: "constrained italic",
			source: "__italic__",
			expected: []*Node{
				{Type: NodeItalic, Text: "italic"},
			},
		},
		{
			name: "unconstrained italic",
			source: "_italic_",
			expected: []*Node{
				{Type: NodeItalic, Text: "italic"},
			},
		},
		{
			name: "italic with spaces",
			source: "__italic text__",
			expected: []*Node{
				{Type: NodeItalic, Text: "italic text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Equal(t, len(tt.expected), len(nodes), "node count mismatch")
			for i, expectedNode := range tt.expected {
				actualNode := nodes[i]
				assert.Equal(t, expectedNode.Type, actualNode.Type)
				assert.Equal(t, expectedNode.Text, actualNode.Text)
			}
		})
	}
}

func TestParseMonospace(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []*Node
	}{
		{
			name: "backtick monospace",
			source: "`mono`",
			expected: []*Node{
				{Type: NodeMonospace, Text: "mono"},
			},
		},
		{
			name: "plusplus span",
			source: "++mono++",
			expected: []*Node{
				{Type: NodeText, Text: "mono"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Equal(t, len(tt.expected), len(nodes), "node count mismatch")
			for i, expectedNode := range tt.expected {
				actualNode := nodes[i]
				assert.Equal(t, expectedNode.Type, actualNode.Type)
				assert.Equal(t, expectedNode.Text, actualNode.Text)
			}
		})
	}
}

func TestParseLink(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []*Node
	}{
		{
			name: "macro link",
			source: "link:https://example.com[Click Here]",
			expected: []*Node{
				{Type: NodeLink, Text: "Click Here", URL: "https://example.com"},
			},
		},
		{
			name: "bare https url",
			source: "https://example.com",
			expected: []*Node{
				{Type: NodeLink, Text: "https://example.com", URL: "https://example.com"},
			},
		},
		{
			name: "bare http url",
			source: "http://example.com",
			expected: []*Node{
				{Type: NodeLink, Text: "http://example.com", URL: "http://example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Equal(t, len(tt.expected), len(nodes), "node count mismatch")
			for i, expectedNode := range tt.expected {
				actualNode := nodes[i]
				assert.Equal(t, expectedNode.Type, actualNode.Type)
				assert.Equal(t, expectedNode.Text, actualNode.Text)
				if expectedNode.URL != "" {
					assert.Equal(t, expectedNode.URL, actualNode.URL)
				}
			}
		})
	}
}

func TestParseMixed(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []*Node
	}{
		{
			name: "bold and italic",
			source: "**bold** and __italic__",
			expected: []*Node{
				{Type: NodeBold, Text: "bold"},
				{Type: NodeText, Text: " and "},
				{Type: NodeItalic, Text: "italic"},
			},
		},
		{
			name: "link with bold",
			source: "**bold** link:https://example.com[Click]",
			expected: []*Node{
				{Type: NodeBold, Text: "bold"},
				{Type: NodeText, Text: " "},
				{Type: NodeLink, Text: "Click", URL: "https://example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Equal(t, len(tt.expected), len(nodes), "node count mismatch")
			for i, expectedNode := range tt.expected {
				if i < len(nodes) {
					actualNode := nodes[i]
					assert.Equal(t, expectedNode.Type, actualNode.Type)
					assert.Equal(t, expectedNode.Text, actualNode.Text)
				}
			}
		})
	}
}

func TestParseCrossRef(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []*Node
	}{
		{
			name:   "simple cross-reference",
			source: "<<section-id>>",
			expected: []*Node{
				{Type: NodeCrossRef, Ref: "section-id", Text: "section-id"},
			},
		},
		{
			name:   "cross-reference with hyphen",
			source: "<<my-section>>",
			expected: []*Node{
				{Type: NodeCrossRef, Ref: "my-section", Text: "my-section"},
			},
		},
		{
			name:   "cross-reference with text before",
			source: "See <<intro>> for details",
			expected: []*Node{
				{Type: NodeText, Text: "See "},
				{Type: NodeCrossRef, Ref: "intro", Text: "intro"},
				{Type: NodeText, Text: " for details"},
			},
		},
		{
			name:   "cross-reference with underscore",
			source: "<<_section_id>>",
			expected: []*Node{
				{Type: NodeCrossRef, Ref: "_section_id", Text: "_section_id"},
			},
		},
		{
			name:   "multiple cross-references",
			source: "<<section1>> and <<section2>>",
			expected: []*Node{
				{Type: NodeCrossRef, Ref: "section1", Text: "section1"},
				{Type: NodeText, Text: " and "},
				{Type: NodeCrossRef, Ref: "section2", Text: "section2"},
			},
		},
		{
			name: "cross-reference with whitespace",
			source: "<< section-id >>",
			expected: []*Node{
				{Type: NodeCrossRef, Ref: "section-id", Text: "section-id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Equal(t, len(tt.expected), len(nodes), "node count mismatch")
			for i, expectedNode := range tt.expected {
				if i < len(nodes) {
					actualNode := nodes[i]
					assert.Equal(t, expectedNode.Type, actualNode.Type, "node type mismatch at index %d", i)
					assert.Equal(t, expectedNode.Text, actualNode.Text, "text mismatch at index %d", i)
					if expectedNode.Type == NodeCrossRef && expectedNode.Ref != "" {
						assert.Equal(t, expectedNode.Ref, actualNode.Ref, "ref mismatch at index %d", i)
					}
				}
			}
		})
	}
}

func TestNodeTypeStringWithCrossRef(t *testing.T) {
	assert.Equal(t, "CrossRef", NodeCrossRef.String())
}

func TestParseCrossRefWithCustomText(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected struct {
			Ref     string
			RefText string
		}
	}{
		{
			name: "cross-reference with custom text",
			source: "<<section-id,See Section>>",
			expected: struct {
				Ref     string
				RefText string
			}{
				Ref:     "section-id",
				RefText: "See Section",
			},
		},
		{
			name: "cross-reference with custom text and spaces",
			source: "<<intro,Introduction Chapter>>",
			expected: struct {
				Ref     string
				RefText string
			}{
				Ref:     "intro",
				RefText: "Introduction Chapter",
			},
		},
		{
			name: "cross-reference with custom text and spaces around comma",
			source: "<<intro , Introduction Chapter>>",
			expected: struct {
				Ref     string
				RefText string
			}{
				Ref:     "intro",
				RefText: "Introduction Chapter",
			},
		},
		{
			name: "cross-reference with underscores in ref and custom text",
			source: "<<_section_id,Custom Text>>",
			expected: struct {
				Ref     string
				RefText string
			}{
				Ref:     "_section_id",
				RefText: "Custom Text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Greater(t, len(nodes), 0, "should have at least one node")
			actualNode := nodes[0]
			assert.Equal(t, NodeCrossRef, actualNode.Type)
			assert.Equal(t, tt.expected.Ref, actualNode.Ref)
			assert.Equal(t, tt.expected.RefText, actualNode.RefText)
			assert.Equal(t, tt.expected.RefText, actualNode.Text)
		})
	}
}

func TestParseKbd(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected struct {
			Type NodeType
			Text string
		}
	}{
		{
			name: "single key",
			source: "kbd:[Ctrl]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeKbd,
				Text: "Ctrl",
			},
		},
		{
			name: "key combination",
			source: "kbd:[Ctrl+C]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeKbd,
				Text: "Ctrl+C",
			},
		},
		{
			name: "three key combination",
			source: "kbd:[Ctrl+Shift+Del]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeKbd,
				Text: "Ctrl+Shift+Del",
			},
		},
		{
			name: "kbd with surrounding text",
			source: "Press kbd:[Enter] to continue",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeKbd,
				Text: "Enter",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			// Find the kbd node
			var kbdNode *Node
			for _, node := range nodes {
				if node.Type == NodeKbd {
					kbdNode = node
					break
				}
			}

			require.NotNil(t, kbdNode, "should have a kbd node")
			assert.Equal(t, tt.expected.Type, kbdNode.Type)
			assert.Equal(t, tt.expected.Text, kbdNode.Text)
		})
	}
}

func TestParseBtn(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected struct {
			Type NodeType
			Text string
		}
	}{
		{
			name: "simple button",
			source: "btn:[OK]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeBtn,
				Text: "OK",
			},
		},
		{
			name: "button with spaces",
			source: "btn:[Cancel]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeBtn,
				Text: "Cancel",
			},
		},
		{
			name: "button with surrounding text",
			source: "Click btn:[Submit] to continue",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeBtn,
				Text: "Submit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			// Find the btn node
			var btnNode *Node
			for _, node := range nodes {
				if node.Type == NodeBtn {
					btnNode = node
					break
				}
			}

			require.NotNil(t, btnNode, "should have a btn node")
			assert.Equal(t, tt.expected.Type, btnNode.Type)
			assert.Equal(t, tt.expected.Text, btnNode.Text)
		})
	}
}

func TestParseMenu(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected struct {
			Type NodeType
			Text string
		}
	}{
		{
			name: "simple menu path",
			source: "menu:[File > Save]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeMenu,
				Text: "File > Save",
			},
		},
		{
			name: "nested menu path",
			source: "menu:[File > New > Document]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeMenu,
				Text: "File > New > Document",
			},
		},
		{
			name: "menu with comma separator",
			source: "menu:[File,Save]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeMenu,
				Text: "File,Save",
			},
		},
		{
			name: "menu with surrounding text",
			source: "Go to menu:[Edit > Find]",
			expected: struct {
				Type NodeType
				Text string
			}{
				Type: NodeMenu,
				Text: "Edit > Find",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			// Find the menu node
			var menuNode *Node
			for _, node := range nodes {
				if node.Type == NodeMenu {
					menuNode = node
					break
				}
			}

			require.NotNil(t, menuNode, "should have a menu node")
			assert.Equal(t, tt.expected.Type, menuNode.Type)
			assert.Equal(t, tt.expected.Text, menuNode.Text)
		})
	}
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected struct {
			Type  NodeType
			Roles []string
		}
	}{
		{
			name: "single role on bold text",
			source: "[.red]**bold text**",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeBold,
				Roles: []string{"red"},
			},
		},
		{
			name: "multiple roles on text",
			source: "[.role1.role2]**text**",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeBold,
				Roles: []string{"role1", "role2"},
			},
		},
		{
			name: "role on italic text",
			source: "[.emphasis]__italic__",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeItalic,
				Roles: []string{"emphasis"},
			},
		},
		{
			name: "role on link",
			source: "[.external]link:text[url]",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeLink,
				Roles: []string{"external"},
			},
		},
		{
			name: "role on kbd",
			source: "[.key-combo]kbd:[Ctrl+C]",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeKbd,
				Roles: []string{"key-combo"},
			},
		},
		{
			name: "role on btn",
			source: "[.primary]btn:[OK]",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeBtn,
				Roles: []string{"primary"},
			},
		},
		{
			name: "role on menu",
			source: "[.nav]menu:[File > Save]",
			expected: struct {
				Type  NodeType
				Roles []string
			}{
				Type:  NodeMenu,
				Roles: []string{"nav"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			require.Greater(t, len(nodes), 0, "should have at least one node")
			actualNode := nodes[0]
			assert.Equal(t, tt.expected.Type, actualNode.Type)
			assert.Equal(t, tt.expected.Roles, actualNode.Roles)
		})
	}
}

func TestParseRoleNoMatch(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "unclosed role bracket",
			source: "[.role**text**",
		},
		{
			name:   "empty role",
			source: "[.]*text*",
		},
		{
			name:   "role without element",
			source: "[.role] plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.source)
			nodes := parser.Parse()

			// The role should not be parsed, or the plain text should be returned
			// In our implementation, roles without a following inline element are treated as text
			assert.NotEmpty(t, nodes, "should have nodes")
		})
	}
}

func TestParseUIMacrosMixed(t *testing.T) {
	source := "Press kbd:[Ctrl+C] to copy, then click btn:[Paste] or go to menu:[Edit > Paste]"
	parser := NewParser(source)
	nodes := parser.Parse()

	// Should have: text, kbd, text, btn, text, menu
	var kbdCount, btnCount, menuCount int
	for _, node := range nodes {
		if node.Type == NodeKbd {
			kbdCount++
		}
		if node.Type == NodeBtn {
			btnCount++
		}
		if node.Type == NodeMenu {
			menuCount++
		}
	}

	assert.Equal(t, 1, kbdCount, "should have 1 kbd node")
	assert.Equal(t, 1, btnCount, "should have 1 btn node")
	assert.Equal(t, 1, menuCount, "should have 1 menu node")
}

