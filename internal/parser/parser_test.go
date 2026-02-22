package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// ExampleNewParserFromString demonstrates basic parsing of AsciiDoc source.
func ExampleNewParserFromString() {
	source := `= Document Title

This is a paragraph with *bold* and _italic_ text.

* First item
* Second item
* Third item
`

	p, err := NewParserFromString(source)
	if err != nil {
		panic(err)
	}

	doc, err := p.Parse()
	if err != nil {
		panic(err)
	}

	// Access document header
	if doc.Header != nil {
		_ = doc.Header.Title
	}

	// Iterate through blocks
	for _, block := range doc.Blocks {
		_ = block.Type()
	}

	// Output:
}

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

func TestSourceBlockWithLineNumbers(t *testing.T) {
	// Test [source,lang,linenums] syntax - ensures language and linenums are parsed correctly
	source := `[source,go,linenums]
----
func main() {
    println("test")
}
----`
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

	block, ok := doc.Blocks[0].(*ast.StyledBlockNode)
	if !ok {
		t.Fatalf("First block should be a StyledBlockNode, got %T", doc.Blocks[0])
	}

	// Check line numbers are enabled
	if !block.LineNumbers {
		t.Error("Expected LineNumbers to be true")
	}

	// Check start line number is 1
	if block.StartLineNumber != 1 {
		t.Errorf("Expected StartLineNumber 1, got %d", block.StartLineNumber)
	}

	// Check language attribute is "go" not "go,linenums"
	if block.Attributes["language"] != "go" {
		t.Errorf("Expected language 'go', got '%s'", block.Attributes["language"])
	}

	// Check linenums attribute exists
	if block.Attributes["linenums"] != "1" {
		t.Errorf("Expected linenums '1', got '%s'", block.Attributes["linenums"])
	}
}

func TestSourceBlockWithLineNumbersCustomStart(t *testing.T) {
	// Test [source,lang,linenums=N] syntax
	source := `[source,javascript,linenums=10]
----
function greet() {
    return "hello";
}
----`
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

	block, ok := doc.Blocks[0].(*ast.StyledBlockNode)
	if !ok {
		t.Fatalf("First block should be a StyledBlockNode, got %T", doc.Blocks[0])
	}

	// Check line numbers are enabled
	if !block.LineNumbers {
		t.Error("Expected LineNumbers to be true")
	}

	// Check start line number is 10
	if block.StartLineNumber != 10 {
		t.Errorf("Expected StartLineNumber 10, got %d", block.StartLineNumber)
	}

	// Check language attribute is "javascript" not "javascript,linenums=10"
	if block.Attributes["language"] != "javascript" {
		t.Errorf("Expected language 'javascript', got '%s'", block.Attributes["language"])
	}
}

func TestSourceBlockWithMultipleOptions(t *testing.T) {
	// Test [source,lang,option1,option2] syntax to ensure proper parsing
	source := `[source,python,linenums,indent=0]
----
def hello():
    pass
----`
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

	block, ok := doc.Blocks[0].(*ast.StyledBlockNode)
	if !ok {
		t.Fatalf("First block should be a StyledBlockNode, got %T", doc.Blocks[0])
	}

	// Check language attribute is just "python"
	if block.Attributes["language"] != "python" {
		t.Errorf("Expected language 'python', got '%s'", block.Attributes["language"])
	}

	// Check both linenums and indent attributes
	if block.Attributes["linenums"] != "1" {
		t.Errorf("Expected linenums '1', got '%s'", block.Attributes["linenums"])
	}

	if block.Attributes["indent"] != "0" {
		t.Errorf("Expected indent '0', got '%s'", block.Attributes["indent"])
	}
}
