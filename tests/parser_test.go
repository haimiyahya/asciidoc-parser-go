// Package tests provides integration tests for asciidoc-parser-go.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
)

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
		path := filepath.Join(examplesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		// Skip empty files (like basic.adoc)
		if strings.TrimSpace(string(content)) == "" {
			t.Skip("empty content, skip parsing")
			continue
		}
		p, err := parser.NewParserFromString(string(content))
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
}
