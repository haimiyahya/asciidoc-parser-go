// Package parser tests for index term integration with full document parsing.
package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
)

func TestIndexTermInDocument(t *testing.T) {
	source := `= Document Title

This is a paragraph with a ((flow index term)) and a (((concealed, secondary))) term.

== Section

More text with ((another term)) here.`

	p, err := NewParserFromString(source)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find paragraphs with index terms
	var foundFlow, foundConcealed bool
	for _, block := range doc.Blocks {
		if para, ok := block.(*ast.NodeParagraph); ok {
			text := para.Text
			if strings.Contains(text, "flow index term") {
				foundFlow = true
			}
			if strings.Contains(text, "concealed") {
				foundConcealed = true
			}
		}
	}

	if !foundFlow {
		t.Error("Flow index term text not found in document")
	}
	if !foundConcealed {
		t.Error("Concealed index term section not found in document")
	}
}

func TestIndexTermHTML5Conversion(t *testing.T) {
	source := `= Test

Paragraph with ((visible term)) and (((hidden, term))).

== Another

More text.`

	p, _ := NewParserFromString(source)
	doc, _ := p.Parse()

	var buf strings.Builder
	htmlConverter := converter.NewHTML5Converter()
	htmlConverter.Convert(doc, &buf)

	html := buf.String()

	// Flow index term should be visible in output
	if !strings.Contains(html, "visible term") {
		t.Error("Flow index term 'visible term' not found in HTML output")
	}

	// Concealed index terms should not appear in output (Asciidoctor behavior)
	// They're used for index generation but not displayed in HTML

	// The paragraph should end with just a period (concealed term produces no output)
	if !strings.Contains(html, "<p>Paragraph with visible term and .</p>") {
		t.Error("Paragraph should not contain concealed index term in HTML output")
	}
}

func TestIndexTermInList(t *testing.T) {
	source := `= Test

* Item with ((term))
* Item with (((hidden, secondary)) term`

	p, _ := NewParserFromString(source)
	doc, _ := p.Parse()

	// Should have a list
	var hasList bool
	for _, block := range doc.Blocks {
		if list, ok := block.(*ast.NodeList); ok {
			hasList = true
			if len(list.Items) < 2 {
				t.Errorf("Expected at least 2 list items, got %d", len(list.Items))
			}
		}
	}

	if !hasList {
		t.Error("No list found in document")
	}
}

func TestIndexTermNodesInInlineParser(t *testing.T) {
	// This tests that the inline parser creates the correct node types
	text := `Text with ((flow)) and (((concealed, secondary, tertiary))) terms.`

	p := inline.NewParser(text)
	nodes := p.Parse()

	var flowCount, concealedCount int
	for _, node := range nodes {
		if node.Type == inline.NodeIndexTerm {
			if !node.IndexTermConcealed {
				flowCount++
				if node.IndexTermPrimary != "flow" {
					t.Errorf("Expected flow term 'flow', got '%s'", node.IndexTermPrimary)
				}
			} else {
				concealedCount++
				if node.IndexTermPrimary != "concealed" {
					t.Errorf("Expected concealed primary 'concealed', got '%s'", node.IndexTermPrimary)
				}
				if node.IndexTermSecondary != "secondary" {
					t.Errorf("Expected secondary 'secondary', got '%s'", node.IndexTermSecondary)
				}
				if node.IndexTermTertiary != "tertiary" {
					t.Errorf("Expected tertiary 'tertiary', got '%s'", node.IndexTermTertiary)
				}
			}
		}
	}

	if flowCount != 1 {
		t.Errorf("Expected 1 flow index term, got %d", flowCount)
	}
	if concealedCount != 1 {
		t.Errorf("Expected 1 concealed index term, got %d", concealedCount)
	}
}
