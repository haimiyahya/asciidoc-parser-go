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

func TestBareURLWithPunctuation(t *testing.T) {
	// Test that bare URLs don't include trailing punctuation
	source := `Visit https://example.com.`

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
	// Should include full URL in href, but period is separate text
	assert.Contains(t, output, `href="https://example.com">https://example.com</a>.`, "HTML should contain link with period after")
}

func TestBareURLWithMultiplePunctuation(t *testing.T) {
	// Test various punctuation marks - they should not be part of URL
	testCases := []string{
		`Check https://site.com, then`,
		`Visit https://site.com!`,
		`Go to https://site.com?`,
		`See https://site.com)`,
		`Open https://site.com]`,
	}

	for _, source := range testCases {
		p, err := NewParserFromString(source)
		require.NoError(t, err)

		doc, err := p.Parse()
		require.NoError(t, err)

		htmlConverter := converter.NewHTML5Converter()
		var buf bytes.Buffer
		err = htmlConverter.Convert(doc, &buf)
		require.NoError(t, err)

		output := buf.String()
		// URL should be clean without punctuation
		assert.Contains(t, output, `href="https://site.com">`, "URL should be clean")
	}
}

func TestNestedInlineMarkup(t *testing.T) {
	source := "This is **bold with _italic_ inside** text."

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes - should have a bold node
	require.NotEmpty(t, para.InlineNodes)

	// Find the bold node and verify it has an italic child
	var foundBold, foundItalicChild bool
	for _, node := range para.InlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			if inlineNode.Type == inline.NodeBold {
				foundBold = true
				// Check if bold node has children
				for _, child := range inlineNode.Children {
					if child.Type == inline.NodeItalic {
						foundItalicChild = true
						assert.Equal(t, "italic", child.Text)
					}
				}
			}
		}
	}
	assert.True(t, foundBold, "Should have found a bold node")
	assert.True(t, foundItalicChild, "Bold node should have italic child")

	// Verify HTML output - should have <strong><em>italic</em></strong>
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<strong>", "HTML should contain strong tag")
	assert.Contains(t, output, "<em>italic</em>", "HTML should contain em tag inside strong")
}

func TestInlineSubscriptParsing(t *testing.T) {
	source := `This is ~subscript~ text.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes
	require.NotEmpty(t, para.InlineNodes)

	// Find the subscript node
	var foundSubscript bool
	for _, node := range para.InlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			if inlineNode.Type == inline.NodeSubscript {
				foundSubscript = true
				assert.Equal(t, "subscript", inlineNode.Text)
			}
		}
	}
	assert.True(t, foundSubscript, "Should have found a subscript node")

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<sub>subscript</sub>", "HTML should contain sub tag")
}

func TestInlineSuperscriptParsing(t *testing.T) {
	source := `This is ^superscript^ text.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	para, ok := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok)

	// Check inline nodes
	require.NotEmpty(t, para.InlineNodes)

	// Find the superscript node
	var foundSuperscript bool
	for _, node := range para.InlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			if inlineNode.Type == inline.NodeSuperscript {
				foundSuperscript = true
				assert.Equal(t, "superscript", inlineNode.Text)
			}
		}
	}
	assert.True(t, foundSuperscript, "Should have found a superscript node")

	// Verify HTML output
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<sup>superscript</sup>", "HTML should contain sup tag")
}

func TestInlineSubscriptSuperscriptMixed(t *testing.T) {
	source := `H~2~O is water and x^2^ + y^2^ = r^2^`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Verify HTML output contains both sub and sup tags
	htmlConverter := converter.NewHTML5Converter()
	var buf bytes.Buffer
	err = htmlConverter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<sub>2</sub>", "HTML should contain subscript for 2")
	assert.Contains(t, output, "<sup>2</sup>", "HTML should contain superscript for 2")
}
