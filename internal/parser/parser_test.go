// Package parser tests AsciiDoc parsing functionality.
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
		t.Fatalf("Parse error: %v", err)
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
