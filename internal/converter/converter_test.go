// Package converter tests output converters for AsciiDoc AST.
package converter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTML5ConverterParagraph(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Hello, world!",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<p>Hello, world!</p>")
	assert.Contains(t, output, "</body>")
	assert.Contains(t, output, "</html>")
}

func TestHTML5ConverterSection(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeSection{
				Level: 1,
				Title: "Introduction",
				Pos:   ast.Position{Line: 1},
			},
			&ast.NodeSection{
				Level: 2,
				Title: "Getting Started",
				Pos:   ast.Position{Line: 2},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<h2>Introduction</h2>")
	assert.Contains(t, output, "<h3>Getting Started</h3>")
}

func TestHTML5ConverterUnorderedList(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind:  ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "Apple",
						Pos:    ast.Position{Line: 1},
					},
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "Banana",
						Pos:    ast.Position{Line: 2},
					},
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "Cherry",
						Pos:    ast.Position{Line: 3},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<ul>")
	assert.Contains(t, output, "<li>Apple</li>")
	assert.Contains(t, output, "<li>Banana</li>")
	assert.Contains(t, output, "<li>Cherry</li>")
	assert.Contains(t, output, "</ul>")
}

func TestHTML5ConverterOrderedList(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind: ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: ".",
						Text:   "First",
						Ordinal: 1,
						Pos:    ast.Position{Line: 1},
					},
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: ".",
						Text:   "Second",
						Ordinal: 2,
						Pos:    ast.Position{Line: 2},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<ol>")
	assert.Contains(t, output, "<li>First</li>")
	assert.Contains(t, output, "<li>Second</li>")
	assert.Contains(t, output, "</ol>")
}

func TestHTML5ConverterLabeledList(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind: ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:      ast.TypeListItem,
						Marker:     "::",
						Term:        "Term 1",
						Definition:  "Definition 1",
						Pos:        ast.Position{Line: 1},
					},
					&ast.NodeListItem{
						Kind:      ast.TypeListItem,
						Marker:     "::",
						Term:        "Term 2",
						Definition:  "Definition 2",
						Pos:        ast.Position{Line: 2},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<dl>")
	assert.Contains(t, output, "<dt>Term 1</dt>")
	assert.Contains(t, output, "<dd>Definition 1</dd>")
	assert.Contains(t, output, "<dt>Term 2</dt>")
	assert.Contains(t, output, "<dd>Definition 2</dd>")
	assert.Contains(t, output, "</dl>")
}

func TestHTML5ConverterLiteralBlock(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeLiteral{
				Kind:  ast.TypeLiteral,
				Lines: []string{"line 1", "line 2", "line 3"},
				Pos:   ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<pre>")
	assert.Contains(t, output, "line 1")
	assert.Contains(t, output, "line 2")
	assert.Contains(t, output, "</pre>")
}

func TestHTML5ConverterDelimitedBlock(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeBlock{
				Kind:      ast.TypeBlock,
				Delimiter:  "=",
				Lines:      []string{"This is an example block"},
				Pos:        ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<div")
	assert.Contains(t, output, `class="exampleblock"`)
	assert.Contains(t, output, "This is an example block")
	assert.Contains(t, output, "</div>")
}

func TestHTML5ConverterDocumentHeader(t *testing.T) {
	doc := &ast.NodeDocument{
		Header: &ast.DocumentHeader{
			Title: "My Document",
		},
		Blocks: []ast.Node{},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<h1>My Document</h1>")
}

func TestHTML5ConverterHTMLEscaping(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "This has <special> & characters",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "&lt;special&gt;")
	assert.Contains(t, output, "&amp; characters")
}

func TestHTML5ConverterMixedContent(t *testing.T) {
	doc := &ast.NodeDocument{
		Header: &ast.DocumentHeader{
			Title: "Test Document",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "First paragraph",
				Pos:  ast.Position{Line: 1},
			},
			&ast.NodeList{
				Kind: ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "List item",
						Pos:    ast.Position{Line: 2},
					},
				},
				Pos: ast.Position{Line: 2},
			},
			&ast.NodeParagraph{
				Text: "Second paragraph",
				Pos:  ast.Position{Line: 3},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<h1>Test Document</h1>")
	assert.Contains(t, output, "<p>First paragraph</p>")
	assert.Contains(t, output, "<ul>")
	assert.Contains(t, output, "<li>List item</li>")
	assert.Contains(t, output, "</ul>")
	assert.Contains(t, output, "<p>Second paragraph</p>")
}

func TestConverterFactory(t *testing.T) {
	factory := NewConverterFactory()

	// Get default converter
	conv := factory.GetDefault()
	require.NotNil(t, conv)

	// Get HTML5 converter
	htmlConv, ok := factory.Get(BackendHTML5)
	assert.True(t, ok)
	assert.NotNil(t, htmlConv)
}

func TestHTML5ConverterAsteriskList(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind: ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "*",
						Text:   "Item 1",
						Pos:    ast.Position{Line: 1},
					},
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "*",
						Text:   "Item 2",
						Pos:    ast.Position{Line: 2},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<ul>")
	assert.Contains(t, output, "<li>Item 1</li>")
	assert.Contains(t, output, "<li>Item 2</li>")
}

func TestHTML5ConverterQuoteBlock(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeBlock{
				Kind:      ast.TypeBlock,
				Delimiter:  "_",
				Lines:      []string{"This is a quote"},
				Pos:        ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<div")
	assert.Contains(t, output, `class="quoteblock"`)
	assert.Contains(t, output, "This is a quote")
}

func TestHTML5ConverterEmptyDocument(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<html")
	assert.Contains(t, output, "<body>")
	assert.Contains(t, output, "</body>")
	assert.Contains(t, output, "</html>")
}

func TestHTML5ConverterPrettyPrint(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Test",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	// Check that output has newlines for pretty printing
	lines := strings.Split(output, "\n")
	assert.Greater(t, len(lines), 5)

	// Check indentation
	assert.Contains(t, output, "  <p>")
}
