// Package parser provides tests for admonition and callout parsing.
package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAdmonitionNote(t *testing.T) {
	source := `NOTE: This is a note admonition.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.Blocks, 1)

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok, "First block should be an AdmonitionNode")
	assert.Equal(t, ast.TypeAdmonition, admonition.Type())
	assert.Equal(t, "NOTE", admonition.Kind)
	assert.Equal(t, "This is a note admonition.", admonition.Text)
}

func TestParseAdmonitionWarning(t *testing.T) {
	source := `WARNING: This is a warning!`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "WARNING", admonition.Kind)
	assert.Equal(t, "This is a warning!", admonition.Text)
}

func TestParseAdmonitionTip(t *testing.T) {
	source := `TIP: Use Ctrl+C to copy.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "TIP", admonition.Kind)
	assert.Equal(t, "Use Ctrl+C to copy.", admonition.Text)
}

func TestParseAdmonitionCaution(t *testing.T) {
	source := `CAUTION: This action cannot be undone.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "CAUTION", admonition.Kind)
}

func TestParseAdmonitionImportant(t *testing.T) {
	source := `IMPORTANT: Read the documentation first.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "IMPORTANT", admonition.Kind)
	assert.Equal(t, "Read the documentation first.", admonition.Text)
}

func TestParseAdmonitionWithParagraphAfter(t *testing.T) {
	source := `NOTE: This is a note.

This is a regular paragraph.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, doc.Blocks, 2)

	admonition, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "NOTE", admonition.Kind)

	para, ok := doc.Blocks[1].(*ast.NodeParagraph)
	require.True(t, ok)
	assert.Equal(t, "This is a regular paragraph.", para.Text)
}

func TestParseCalloutList(t *testing.T) {
	source := `NOTE> This is a callout item
WARNING> This is another callout`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Callouts are treated as lists, so should get a list block
	assert.NotEmpty(t, doc.Blocks, "Should have at least one block")

	// Check if first block is a list
	list, ok := doc.Blocks[0].(*ast.NodeList)
	if ok {
		assert.NotEmpty(t, list.Items, "List should have items")
	}
}

func TestParseMixedAdmonitions(t *testing.T) {
	source := `NOTE: First note

Some paragraph text

WARNING: Be careful!
TIP: Pro tip`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have: NOTE, paragraph, WARNING, TIP
	require.Len(t, doc.Blocks, 4)

	admonition1, ok := doc.Blocks[0].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "NOTE", admonition1.Kind)

	admonition2, ok := doc.Blocks[2].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "WARNING", admonition2.Kind)

	admonition3, ok := doc.Blocks[3].(*ast.AdmonitionNode)
	require.True(t, ok)
	assert.Equal(t, "TIP", admonition3.Kind)
}

func TestAdmonitionHTML5Conversion(t *testing.T) {
	source := `NOTE: This is important`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf strings.Builder
	conv := converter.NewHTML5Converter()
	err = conv.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<div class="admonition-note"`)
	assert.Contains(t, output, `This is important`)
}

func TestWarningHTML5Conversion(t *testing.T) {
	source := `WARNING: Don't do this!`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf strings.Builder
	conv := converter.NewHTML5Converter()
	err = conv.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<div class="admonition-warning"`)
}
