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
			name: "plusplus monospace",
			source: "++mono++",
			expected: []*Node{
				{Type: NodeMonospace, Text: "mono"},
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
			source: "link:text[url]",
			expected: []*Node{
				{Type: NodeLink, Text: "text", URL: "url"},
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
			source: "**bold** link:text[url]",
			expected: []*Node{
				{Type: NodeBold, Text: "bold"},
				{Type: NodeText, Text: " "},
				{Type: NodeLink, Text: "text", URL: "url"},
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
