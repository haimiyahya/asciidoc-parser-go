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
	assert.Contains(t, output, "<li>")
	assert.Contains(t, output, "Apple")
	assert.Contains(t, output, "Banana")
	assert.Contains(t, output, "Cherry")
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
	assert.Contains(t, output, "<ol")
	assert.Contains(t, output, "First")
	assert.Contains(t, output, "Second")
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
	assert.Contains(t, output, "<dt")
	assert.Contains(t, output, "Term 1")
	assert.Contains(t, output, "Term 2")
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
	assert.Contains(t, output, "List item")
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
	assert.Contains(t, output, "Item 1")
	assert.Contains(t, output, "Item 2")
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

	// Check indentation - paragraph may be wrapped in div
	assert.Contains(t, output, "<p>")
}

func TestHTML5ConverterTableWithHeader(t *testing.T) {
	// Test table with options="header" renders thead and th
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.Table{
				Rows: []ast.TableRow{
					{
						Kind: ast.TableRowHeader,
						Cells: []ast.TableCell{
							{Text: "Name"},
							{Text: "Age"},
						},
					},
					{
						Kind: ast.TableRowBody,
						Cells: []ast.TableCell{
							{Text: "Alice"},
							{Text: "30"},
						},
					},
				},
				Attributes: map[string]string{
					"options": "header",
				},
				HeaderRowIndex: 0,
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<thead>")
	assert.Contains(t, output, "<th")
	assert.Contains(t, output, "<tbody>")
	assert.Contains(t, output, "<td")
}

func TestHTML5ConverterTableWithoutHeader(t *testing.T) {
	// Test table without options="header" doesn't render thead
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.Table{
				Rows: []ast.TableRow{
					{
						Kind: ast.TableRowBody,
						Cells: []ast.TableCell{
							{Text: "A"},
							{Text: "B"},
						},
					},
				},
				HeaderRowIndex: -1,
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "<thead>")
	assert.Contains(t, output, "<tbody>")
}

func TestHTML5ConverterTableFrameClasses(t *testing.T) {
	// Test that frame attribute is rendered as CSS class
	testCases := []struct {
		frame     ast.TableFrame
		classSubstr string
	}{
		{ast.FrameAll, "frame-all"},
		{ast.FrameSides, "frame-sides"},
		{ast.FrameTopbot, "frame-topbot"},
		{ast.FrameNone, "frame-none"},
	}

	for _, tc := range testCases {
		doc := &ast.NodeDocument{
			Blocks: []ast.Node{
				&ast.Table{
					Rows: []ast.TableRow{
						{
							Cells: []ast.TableCell{{Text: "A"}},
						},
					},
					Attributes: map[string]string{},
				},
			},
		}
		// Set the frame by accessing internal state
		if table, ok := doc.Blocks[0].(*ast.Table); ok {
			table.Attributes["frame"] = string(tc.frame)
		}

		converter := NewHTML5Converter()
		var buf bytes.Buffer
		err := converter.Convert(doc, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, tc.classSubstr, "Frame class not found for "+string(tc.frame))
	}
}

func TestHTML5ConverterTableGridClasses(t *testing.T) {
	// Test that grid attribute is rendered as CSS class
	testCases := []struct {
		grid       ast.TableGrid
		classSubstr string
	}{
		{ast.GridAll, "grid-all"},
		{ast.GridRows, "grid-rows"},
		{ast.GridCols, "grid-cols"},
		{ast.GridNone, "grid-none"},
	}

	for _, tc := range testCases {
		doc := &ast.NodeDocument{
			Blocks: []ast.Node{
				&ast.Table{
					Rows: []ast.TableRow{
						{
							Cells: []ast.TableCell{{Text: "A"}},
						},
					},
					Attributes: map[string]string{},
				},
			},
		}
		if table, ok := doc.Blocks[0].(*ast.Table); ok {
			table.Attributes["grid"] = string(tc.grid)
		}

		converter := NewHTML5Converter()
		var buf bytes.Buffer
		err := converter.Convert(doc, &buf)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, tc.classSubstr, "Grid class not found for "+string(tc.grid))
	}
}

func TestHTML5ConverterTableStripes(t *testing.T) {
	// Test that stripes attribute is rendered as CSS class
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.Table{
				Rows: []ast.TableRow{
					{
						Cells: []ast.TableCell{{Text: "A"}},
					},
				},
				Attributes: map[string]string{
					"stripes": "even",
				},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "stripes-even")
}

func TestHTML5ConverterTableCaption(t *testing.T) {
	// Test that caption is rendered
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.Table{
				Rows: []ast.TableRow{
					{
						Cells: []ast.TableCell{{Text: "A"}},
					},
				},
				Caption: "My Caption",
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<caption>My Caption</caption>")
}

func TestHTML5ConverterTableAutowidth(t *testing.T) {
	// Test that %autowidth suppresses colgroup
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.Table{
				Rows: []ast.TableRow{
					{
						Cells: []ast.TableCell{{Text: "A"}},
					},
				},
				Attributes: map[string]string{
					"autowidth": "true",
				},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "<colgroup>", "colgroup should not be present with autowidth")
}

func TestHTML5ConverterTableNoAutowidth(t *testing.T) {
	// Test that colgroup is rendered when autowidth is not set
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.Table{
				Rows: []ast.TableRow{
					{
						Cells: []ast.TableCell{{Text: "A"}},
					},
				},
				Attributes: map[string]string{},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<colgroup>", "colgroup should be present without autowidth")
}

func TestHTML5ConverterExplicitStyleOrderedList(t *testing.T) {
	// Test that explicit style markers [a], [A], [i], [I] render as <ol> with correct class
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeList{
				Items: []ast.Node{
					&ast.NodeListItem{
						Marker: "[a]",
						Text:   "First item",
					},
					&ast.NodeListItem{
						Marker: "[a]",
						Text:   "Second item",
					},
				},
			},
		},
	}

	converter := NewHTML5Converter()
	var buf bytes.Buffer
	err := converter.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	t.Logf("Output:\n%s", output)
	assert.Contains(t, output, `<ol class="loweralpha">`, "Should be ordered list with loweralpha class")
	assert.Contains(t, output, `<div class="olist loweralpha">`, "Should have olist wrapper with loweralpha class")
}

