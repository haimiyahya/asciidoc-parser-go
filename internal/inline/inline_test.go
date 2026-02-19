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
		{NodeCrossRef, "CrossRef"},
		{NodeIndexTerm, "IndexTerm"},
		{NodeCustomMacro, "CustomMacro"},
		{NodePassThrough, "PassThrough"},
		{NodeRawPassThrough, "RawPassThrough"},
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


func TestParseFootnote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType NodeType
		wantID   string
		wantText string
	}{
		{
			name:     "footnote with text",
			input:    "footnote:note1[This is a footnote]",
			wantType: NodeFootnote,
			wantID:   "note1",
			wantText: "This is a footnote",
		},
		{
			name:     "footnote reference (empty text)",
			input:    "footnote:note1[]",
			wantType: NodeFootnote,
			wantID:   "note1",
			wantText: "",
		},
		{
			name:     "footnote with complex text",
			input:    "footnote:fn2[Text with **bold** inside]",
			wantType: NodeFootnote,
			wantID:   "fn2",
			wantText: "Text with **bold** inside",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			nodes := p.Parse()

			if len(nodes) == 0 {
				t.Fatalf("no nodes returned")
			}

			node := nodes[0]
			if node.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", node.Type, tt.wantType)
			}
			if node.FootnoteID != tt.wantID {
				t.Errorf("FootnoteID = %q, want %q", node.FootnoteID, tt.wantID)
			}
			if node.FootnoteText != tt.wantText {
				t.Errorf("FootnoteText = %q, want %q", node.FootnoteText, tt.wantText)
			}
		})
	}
}

func TestParseFootnoteInText(t *testing.T) {
	input := "This is text with a footnote:note1[ref] and more text"
	p := NewParser(input)
	nodes := p.Parse()

	if len(nodes) != 3 {
		t.Fatalf("got %d nodes, want 3", len(nodes))
	}

	// First node should be text
	if nodes[0].Type != NodeText {
		t.Errorf("First node Type = %v, want NodeText", nodes[0].Type)
	}

	// Second node should be footnote
	if nodes[1].Type != NodeFootnote {
		t.Errorf("Second node Type = %v, want NodeFootnote", nodes[1].Type)
	}
	if nodes[1].FootnoteID != "note1" {
		t.Errorf("FootnoteID = %q, want \"note1\"", nodes[1].FootnoteID)
	}

	// Third node should be text (after footnote)
	if nodes[2].Type != NodeText {
		t.Errorf("Third node Type = %v, want NodeText", nodes[2].Type)
	}
}
