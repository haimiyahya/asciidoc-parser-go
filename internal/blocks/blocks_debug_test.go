package blocks

import (
	"fmt"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
	"github.com/stretchr/testify/require"
)

func TestDebugParseSimpleDocument(t *testing.T) {
	source := `= Document Title

This is a paragraph.

== Section 1

This is a second paragraph.`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	fmt.Printf("=== DEBUG: Reader created, testing initial state ===\n")
	fmt.Printf("HasMoreLines: %v\n", r.HasMoreLines())
	
	parser := NewParser(r)
	
	fmt.Printf("=== DEBUG: About to call Parse ===\n")
	doc, err := parser.Parse()
	
	fmt.Printf("=== DEBUG: Parse returned ===\n")
	fmt.Printf("doc: %v\n", doc)
	fmt.Printf("err: %v\n", err)

	require.NoError(t, err)
	require.NotNil(t, doc)

	fmt.Printf("=== DEBUG: Doc has %d blocks ===\n", len(doc.Blocks))
	for i, b := range doc.Blocks {
		fmt.Printf("  Block %d: Type=%v\n", i, b.Type)
	}
}
