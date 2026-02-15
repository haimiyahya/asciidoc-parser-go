package parser

import (
	"bytes"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInlineParsingInParagraph(t *testing.T) {
	source := `This is **bold** and __italic__ text.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have 1 block (paragraph)
	assert.Len(t, doc.Blocks, 1)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok, "First block should be a paragraph")
	assert.Equal(t, "This is **bold** and __italic__ text.", para.Text)

	// Check inline nodes were parsed
	assert.NotEmpty(t, para.InlineNodes, "Paragraph should have inline nodes")
}

func TestInlineParsingHTML5Conversion(t *testing.T) {
	source := `This is **bold** and __italic__ text.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Convert to HTML5
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify HTML contains the expected inline markup
	assert.Contains(t, output, "<strong>bold</strong>", "HTML should contain <strong>bold</strong>")
	assert.Contains(t, output, "<em>italic</em>", "HTML should contain <em>italic</em>")
}

func TestInlineMonospaceParsing(t *testing.T) {
	source := `This is ` + "`monospace`" + ` text.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes
	require.NotEmpty(t, para.InlineNodes)

	// Find the monospace node
	var foundMonospace bool
	for _, node := range para.InlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			if inlineNode.Type == inline.NodeMonospace {
				foundMonospace = true
				assert.Equal(t, "monospace", inlineNode.Text)
			}
		}
	}
	assert.True(t, foundMonospace, "Should have found a monospace node")

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<code>monospace</code>", "HTML should contain <code>monospace</code>")
}

func TestInlineLinkParsing(t *testing.T) {
	// Test with macro-style link
	source := `Check out link:AsciiDoc[https://www.asciidoctor.org].`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes - should have link nodes
	require.NotEmpty(t, para.InlineNodes)

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<a href="`, "HTML should contain an anchor tag with href")
}

func TestInlineMixedFormatting(t *testing.T) {
	source := `This has **bold**, __italic__, and ` + "`monospace`" + ` text.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify HTML output contains all formatting
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<strong>", "HTML should contain <strong>")
	assert.Contains(t, output, "<em>", "HTML should contain <em>")
	assert.Contains(t, output, "<code>", "HTML should contain <code>")
}

func TestInlineInListItems(t *testing.T) {
	source := `- **Bold** item
- Normal item
- __Italic__ item`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have 1 list block
	require.Len(t, doc.Blocks, 1)

	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok, "Block should be a list")
	require.Len(t, list.Items, 3, "Should have 3 list items")

	// First item should have inline nodes
	firstItem := list.Items[0].(*ast.NodeListItem)
	assert.NotEmpty(t, firstItem.InlineNodes, "First item should have inline nodes")

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<strong>Bold</strong>", "HTML should contain <strong>Bold</strong>")
	assert.Contains(t, output, "<em>Italic</em>", "HTML should contain <em>Italic</em>")
	assert.Contains(t, output, "Normal item", "HTML should contain normal list item text")
}

func TestInlineImageParsing(t *testing.T) {
	// Test image:url[alt-text] syntax
	source := `See image:logo.png[Our Logo] for details.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes - should have an image node
	require.NotEmpty(t, para.InlineNodes)

	// Find the image node
	var foundImage bool
	for _, node := range para.InlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			if inlineNode.Type == inline.NodeImage {
				foundImage = true
				assert.Equal(t, "logo.png", inlineNode.URL)
				assert.Equal(t, "Our Logo", inlineNode.Alt)
			}
		}
	}
	assert.True(t, foundImage, "Should have found an image node")

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<img src="logo.png" alt="Our Logo">`, "HTML should contain img tag")
}

func TestInlineImageWithoutAlt(t *testing.T) {
	// Test image:url without alt text
	source := `See image:photo.jpg`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes
	require.NotEmpty(t, para.InlineNodes)

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<img src="photo.jpg"`, "HTML should contain img tag")
}

func TestInlineImageWithFullPath(t *testing.T) {
	// Test image with full path
	source := `External image:https://example.com/pic.png[Alt text]`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<img src="https://example.com/pic.png" alt="Alt text">`, "HTML should contain img with full URL")
}
