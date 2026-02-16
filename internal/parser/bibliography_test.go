// Package parser tests for bibliography functionality.
package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

func TestBibliographySection(t *testing.T) {
	source := `[bibliography]
== Bibliography

* [[[pp]]] Andy Hunt & Dave Thomas. The Pragmatic Programmer.
* [[[gof,gang]]] Erich Gamma et al. Design Patterns.
`

	p, err := NewParserFromString(source)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find the bibliography node
	var bib *ast.BibliographyNode
	for _, block := range doc.Blocks {
		if b, ok := block.(*ast.BibliographyNode); ok {
			bib = b
			break
		}
	}

	if bib == nil {
		t.Fatal("No bibliography node found")
	}

	if bib.Title != "Bibliography" {
		t.Errorf("Expected title 'Bibliography', got '%s'", bib.Title)
	}

	if len(bib.Entries) != 2 {
		t.Fatalf("Expected 2 bibliography entries, got %d", len(bib.Entries))
	}

	// Check first entry
	entry1 := bib.Entries[0]
	if entry1.Label != "pp" {
		t.Errorf("Expected label 'pp', got '%s'", entry1.Label)
	}
	if entry1.XRefText != "" {
		t.Errorf("Expected empty XRefText, got '%s'", entry1.XRefText)
	}
	if !strings.Contains(entry1.Text, "The Pragmatic Programmer") {
		t.Errorf("Entry text missing expected content: %s", entry1.Text)
	}

	// Check second entry with xreftext
	entry2 := bib.Entries[1]
	if entry2.Label != "gof" {
		t.Errorf("Expected label 'gof', got '%s'", entry2.Label)
	}
	if entry2.XRefText != "gang" {
		t.Errorf("Expected xreftext 'gang', got '%s'", entry2.XRefText)
	}

	// Check document bibliography entries map
	if doc.BibliographyEntries == nil {
		t.Fatal("Document BibliographyEntries map is nil")
	}

	if len(doc.BibliographyEntries) != 2 {
		t.Errorf("Expected 2 entries in BibliographyEntries map, got %d", len(doc.BibliographyEntries))
	}

	if doc.BibliographyEntries["pp"] == nil {
		t.Error("'pp' entry not found in BibliographyEntries map")
	}

	if doc.BibliographyEntries["gof"] == nil {
		t.Error("'gof' entry not found in BibliographyEntries map")
	}
}

func TestBibliographyWithCitation(t *testing.T) {
	source := `= Document Title

This book <<pp>> is essential reading.

[bibliography]
== Bibliography

* [[[pp]]] Andy Hunt & Dave Thomas. The Pragmatic Programmer.
`

	p, err := NewParserFromString(source)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find the bibliography node
	var bib *ast.BibliographyNode
	for _, block := range doc.Blocks {
		if b, ok := block.(*ast.BibliographyNode); ok {
			bib = b
			break
		}
	}

	if bib == nil {
		t.Fatal("No bibliography node found")
	}

	if len(bib.Entries) != 1 {
		t.Fatalf("Expected 1 bibliography entry, got %d", len(bib.Entries))
	}

	// Find the paragraph with citation
	var para *ast.NodeParagraph
	for _, block := range doc.Blocks {
		if p, ok := block.(*ast.NodeParagraph); ok {
			para = p
			break
		}
	}

	if para == nil {
		t.Fatal("No paragraph found")
	}

	// Check that the paragraph contains a cross-reference to 'pp'
	if !strings.Contains(para.Text, "pp") {
		t.Error("Paragraph text missing 'pp' reference")
	}
}

func TestBibliographyEmptyEntries(t *testing.T) {
	source := `[bibliography]
== Bibliography
`

	p, err := NewParserFromString(source)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find the bibliography node
	var bib *ast.BibliographyNode
	for _, block := range doc.Blocks {
		if b, ok := block.(*ast.BibliographyNode); ok {
			bib = b
			break
		}
	}

	if bib == nil {
		t.Fatal("No bibliography node found")
	}

	if len(bib.Entries) != 0 {
		t.Errorf("Expected 0 bibliography entries, got %d", len(bib.Entries))
	}
}

func TestBibliographyEntryWithInlineMarkup(t *testing.T) {
	source := `[bibliography]
== Bibliography

* [[[pp]]] **The Pragmatic Programmer** by __Andy Hunt__.
`

	p, err := NewParserFromString(source)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find the bibliography node
	var bib *ast.BibliographyNode
	for _, block := range doc.Blocks {
		if b, ok := block.(*ast.BibliographyNode); ok {
			bib = b
			break
		}
	}

	if bib == nil {
		t.Fatal("No bibliography node found")
	}

	if len(bib.Entries) != 1 {
		t.Fatalf("Expected 1 bibliography entry, got %d", len(bib.Entries))
	}

	entry := bib.Entries[0]
	if len(entry.InlineNodes) == 0 {
		t.Error("Expected inline nodes in bibliography entry")
	}
}
