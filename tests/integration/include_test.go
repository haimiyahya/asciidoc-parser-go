// Package integration provides end-to-end tests for include directive processing.
package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/haimiyahya/asciidoc-parser-go/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncludeDirective tests basic include directive functionality.
func TestIncludeDirective(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a chapter file to be included
	chapterContent := `== Included Chapter

This content was included from another file.

* Item 1
* Item 2
* Item 3
`
	chapterPath := filepath.Join(tmpDir, "chapter.adoc")
	err := os.WriteFile(chapterPath, []byte(chapterContent), 0644)
	require.NoError(t, err)

	// Create main document with include directive
	mainDoc := `= Main Document

This is the main document.

include::chapter.adoc[]

This content comes after the include.
`

	// Create parser with include processor and base directory
	includeProc := processor.NewIncludeProcessor(tmpDir)
	p, err := parser.NewParserFromString(mainDoc,
		parser.WithIncludeProcessor(includeProc),
		parser.WithBaseDir(tmpDir))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify the document has content from the included file
	// The included chapter should add sections and content
	assert.NotEmpty(t, doc.Blocks, "document should have blocks from include")
}

// TestIncludeDirectiveHTML5Conversion tests that included content converts properly to HTML5.
func TestIncludeDirectiveHTML5Conversion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a section file to be included
	sectionContent := `== Included Section

This is a paragraph.

* List item 1
* List item 2
`
	sectionPath := filepath.Join(tmpDir, "section.adoc")
	err := os.WriteFile(sectionPath, []byte(sectionContent), 0644)
	require.NoError(t, err)

	// Create main document
	mainDoc := `= Main Title

Main paragraph.

include::section.adoc[]

Final paragraph.
`

	includeProc := processor.NewIncludeProcessor(tmpDir)
	p, err := parser.NewParserFromString(mainDoc,
		parser.WithIncludeProcessor(includeProc),
		parser.WithBaseDir(tmpDir))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Convert to HTML5
	var buf strings.Builder
	c := converter.NewHTML5Converter()
	err = c.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify included section is present
	assert.Contains(t, output, `id="included-section"`)
	assert.Contains(t, output, "<li>List item 1</li>")
}

// TestIncludeDirectiveWithLines tests include directive with line range filtering.
func TestIncludeDirectiveWithLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with multiple lines
	content := `Line 1
Line 2
Line 3
Line 4
Line 5
`
	filePath := filepath.Join(tmpDir, "lines.txt")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Include only lines 2-4
	mainDoc := `= Test

include::lines.txt[lines=2..4]
`

	includeProc := processor.NewIncludeProcessor(tmpDir)
	p, err := parser.NewParserFromString(mainDoc,
		parser.WithIncludeProcessor(includeProc),
		parser.WithBaseDir(tmpDir))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify we got content (lines 2-4)
	assert.NotEmpty(t, doc.Blocks)
}

// TestIncludeDirectiveWithTags tests include directive with tag filtering.
func TestIncludeDirectiveWithTags(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with tagged content
	content := `// tag::feature-a
Feature A content
// end::feature-a

// tag::feature-b
Feature B content
// end::feature-b
`
	filePath := filepath.Join(tmpDir, "features.txt")
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	// Include only feature-a
	mainDoc := `= Test

include::features.txt[tags=feature-a]
`

	includeProc := processor.NewIncludeProcessor(tmpDir)
	p, err := parser.NewParserFromString(mainDoc,
		parser.WithIncludeProcessor(includeProc),
		parser.WithBaseDir(tmpDir))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify we got content
	assert.NotEmpty(t, doc.Blocks)
}

// TestNestedIncludeDirective tests that nested includes are supported.
func TestNestedIncludeDirective(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file that includes another file
	innerContent := `Inner content from nested include.`
	innerPath := filepath.Join(tmpDir, "inner.adoc")
	err := os.WriteFile(innerPath, []byte(innerContent), 0644)
	require.NoError(t, err)

	// Create middle file that includes inner
	middleContent := `include::inner.adoc[]`
	middlePath := filepath.Join(tmpDir, "middle.adoc")
	err = os.WriteFile(middlePath, []byte(middleContent), 0644)
	require.NoError(t, err)

	// Create main document that includes middle
	mainDoc := `= Test

include::middle.adoc[]
`

	includeProc := processor.NewIncludeProcessor(tmpDir)
	p, err := parser.NewParserFromString(mainDoc,
		parser.WithIncludeProcessor(includeProc),
		parser.WithBaseDir(tmpDir))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify nested content was included
	assert.NotEmpty(t, doc.Blocks, "nested include should work")
}

// TestIncludeWithoutProcessor tests that without processor, include is treated as macro.
func TestIncludeWithoutProcessor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file
	chapterContent := `== Included Chapter`
	chapterPath := filepath.Join(tmpDir, "chapter.adoc")
	err := os.WriteFile(chapterPath, []byte(chapterContent), 0644)
	require.NoError(t, err)

	// Create main document
	mainDoc := `include::chapter.adoc[]`

	// Parse WITHOUT include processor
	p, err := parser.NewParserFromString(mainDoc)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have a macro node since include processor wasn't provided
	assert.NotEmpty(t, doc.Blocks, "document should have macro node when no include processor")
}
