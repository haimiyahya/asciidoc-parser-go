package converter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

func TestAllConvertersBasic(t *testing.T) {
	// Create test AST directly
	doc := &ast.NodeDocument{
		Header: &ast.DocumentHeader{Title: "Test Document"},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "This is a paragraph with bold text.",
				Pos:  ast.Position{Line: 2},
			},
			&ast.NodeList{
				Items: []ast.Node{
					&ast.NodeListItem{Marker: "-", Text: "Item 1", Pos: ast.Position{Line: 4}},
					&ast.NodeListItem{Marker: "-", Text: "Item 2", Pos: ast.Position{Line: 5}},
				},
				Pos: ast.Position{Line: 4},
			},
		},
	}

	// Test HTML5
	t.Run("HTML5", func(t *testing.T) {
		var buf bytes.Buffer
		c := NewHTML5Converter()
		err := c.Convert(doc, &buf)
		if err != nil {
			t.Fatalf("HTML5 converter error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "<h1>Test Document</h1>") {
			t.Errorf("HTML5 missing title")
		}
		if !strings.Contains(output, "<p>") {
			t.Errorf("HTML5 missing paragraph")
		}
		if !strings.Contains(output, "<ul>") {
			t.Errorf("HTML5 missing list")
		}
		t.Logf("HTML5 output:\n%s", output)
	})

	// Test DocBook
	t.Run("DocBook", func(t *testing.T) {
		var buf bytes.Buffer
		c := NewDocBookConverter()
		err := c.Convert(doc, &buf)
		if err != nil {
			t.Fatalf("DocBook converter error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "<?xml version=\"1.0\"") {
			t.Errorf("DocBook missing XML declaration")
		}
		if !strings.Contains(output, "<article") {
			t.Errorf("DocBook missing article element")
		}
		if !strings.Contains(output, "<title>Test Document</title>") {
			t.Errorf("DocBook missing title")
		}
		t.Logf("DocBook output:\n%s", output)
	})

	// Test ManPage
	t.Run("ManPage", func(t *testing.T) {
		var buf bytes.Buffer
		c := NewManPageConverter().WithManualName("TEST")
		err := c.Convert(doc, &buf)
		if err != nil {
			t.Fatalf("ManPage converter error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, ".TH") {
			t.Errorf("ManPage missing TH header")
		}
		if !strings.Contains(output, "TEST") {
			t.Errorf("ManPage missing manual name")
		}
		t.Logf("ManPage output:\n%s", output)
	})
}
