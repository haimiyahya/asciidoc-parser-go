package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

func TestParser(t *testing.T) {
	// Minimal placeholder test - will be expanded
	// This ensures the test file compiles
}

func TestCaptionOrder(t *testing.T) {
	input := `= Test

.Guidelines for Effective Tables
[cols="2*",options="header"]
|===
| Practice | Recommendation
| Headers | Use options="header"
|===
`
	p, err := NewParserFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(doc.Blocks) == 0 {
		t.Fatal("No blocks in document")
	}

	table, ok := doc.Blocks[0].(*ast.Table)
	if !ok {
		t.Fatalf("First block is not a table, got %T", doc.Blocks[0])
	}

	if table.Caption != "Guidelines for Effective Tables" {
		t.Errorf("Expected caption 'Guidelines for Effective Tables', got '%s'", table.Caption)
	}
}

func TestBlockStyleAdmonition(t *testing.T) {
	source := `[NOTE]
This is a note.`
	p, err := NewParserFromReader(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(doc.Blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d: %v", len(doc.Blocks), doc.Blocks)
	}

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	if !ok {
		t.Fatalf("First block should be an AdmonitionNode, got %T", doc.Blocks[0])
	}

	if admonition.Kind != "NOTE" {
		t.Errorf("Expected kind NOTE, got %s", admonition.Kind)
	}

	if len(admonition.Blocks) != 1 {
		t.Fatalf("Expected 1 child block, got %d", len(admonition.Blocks))
	}

	para, ok := admonition.Blocks[0].(*ast.NodeParagraph)
	if !ok {
		t.Fatalf("Child block should be a NodeParagraph, got %T", admonition.Blocks[0])
	}

	if para.Text != "This is a note." {
		t.Errorf("Expected text 'This is a note.', got '%s'", para.Text)
	}
}

func TestInlineStyleAdmonition(t *testing.T) {
	source := `NOTE: This is an inline-style note.`
	p, err := NewParserFromReader(strings.NewReader(source))
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(doc.Blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(doc.Blocks))
	}

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	if !ok {
		t.Fatalf("First block should be an AdmonitionNode, got %T", doc.Blocks[0])
	}

	if admonition.Kind != "NOTE" {
		t.Errorf("Expected kind NOTE, got %s", admonition.Kind)
	}

	if admonition.Text != "This is an inline-style note." {
		t.Errorf("Expected text 'This is an inline-style note.', got '%s'", admonition.Text)
	}

	if len(admonition.Blocks) != 0 {
		t.Errorf("Inline-style should not have child blocks, got %d", len(admonition.Blocks))
	}
}
