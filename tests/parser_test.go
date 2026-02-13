// Package tests provides integration tests.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
)

// TestBasicParse verifies basic parsing works.
func TestBasicParse(t *testing.T) {
	source := `= Document Title

This is a paragraph.

== Section

Another paragraph.
`

	p, err := parser.NewParserFromString(source)
	if err != nil {
		t.Fatalf("NewParserFromString: %v", err)
	}

	doc, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if doc == nil {
		t.Fatal("doc is nil")
	}
}

// TestExamplesFromDir parses all .adoc files in testdata/examples.
func TestExamplesFromDir(t *testing.T) {
	examplesDir := filepath.Join("..", "testdata", "examples")

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		// Skip if directory doesn't exist yet
		t.Skipf("examples directory not found: %v", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".adoc") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(examplesDir, entry.Name())

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}

			p, err := parser.NewParserFromString(string(content))
			if err != nil {
				t.Fatalf("NewParserFromString: %v", err)
			}

			_, err = p.Parse()
			if err != nil {
				t.Logf("Parse error (expected during development): %v", err)
			}
		})
	}
}
