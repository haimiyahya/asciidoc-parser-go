// Package integration provides end-to-end tests for cross-references and section IDs.
package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCrossReference verifies cross-reference parsing and HTML5 conversion.
func TestCrossReference(t *testing.T) {
	src := `= Document Title

== Introduction

This is the intro section. See <<details>> for more info.

== Details

More details here.

Back to <<introduction>>.

== Conclusion

The end.
`

	p, err := parser.NewParserFromReader(strings.NewReader(src))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	var buf bytes.Buffer
	c := converter.NewHTML5Converter()
	err = c.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify cross-references are converted to anchor tags
	t.Logf("Output:\n%s", output)

	// Cross-reference should produce <a href="#section-id"> for simple titles
	// (Asciidoctor only adds underscore prefix for multi-word titles)
	assert.Contains(t, output, `<a href="#details">`)
	assert.Contains(t, output, `<a href="#introduction">`)

	// Sections should have id attributes (no underscore prefix for simple titles)
	assert.Contains(t, output, `id="details"`)
	assert.Contains(t, output, `id="introduction"`)
}

// TestSectionWithIDAttribute verifies explicit section ID attributes.
func TestSectionWithIDAttribute(t *testing.T) {
	src := `= Document Title

[[custom-id]]
== Section Title

Content here.

== Another Section

See <<custom-id>> for reference.
`

	p, err := parser.NewParserFromReader(strings.NewReader(src))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf bytes.Buffer
	c := converter.NewHTML5Converter()
	err = c.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	t.Logf("Output:\n%s", output)

	// Verify custom ID is used
	assert.Contains(t, output, `id="custom-id"`)
	assert.Contains(t, output, `<a href="#custom-id">`)
}

// TestCrossReferenceInParagraph verifies cross-references within paragraphs.
func TestCrossReferenceInParagraph(t *testing.T) {
	src := `= Document Title

== Introduction

This paragraph has a <<section-two>> cross-reference in the middle.

== Section Two

Target section.
`

	p, err := parser.NewParserFromReader(strings.NewReader(src))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf bytes.Buffer
	c := converter.NewHTML5Converter()
	err = c.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	t.Logf("Output:\n%s", output)

	// Verify cross-reference in paragraph
	assert.Contains(t, output, `<a href="#section-two">`)
}

// TestMultipleCrossReferences verifies multiple cross-references in one document.
func TestMultipleCrossReferences(t *testing.T) {
	src := `= Document Title

== First Section

Content for first section.

== Second Section

Content for second section.

== Third Section

See <<first-section>> and <<second-section>>.
`

	p, err := parser.NewParserFromReader(strings.NewReader(src))
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf bytes.Buffer
	c := converter.NewHTML5Converter()
	err = c.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	t.Logf("Output:\n%s", output)

	// Verify all cross-references are rendered
	assert.Contains(t, output, `<a href="#first-section">`)
	assert.Contains(t, output, `<a href="#second-section">`)
}
