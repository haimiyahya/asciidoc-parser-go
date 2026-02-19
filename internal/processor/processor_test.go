// Package processor tests document processing functionality.
package processor

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessorAttributeSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"firstname": "John",
			"lastname":  "Doe",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Hello, {firstname} {lastname}!",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	assert.Equal(t, "Hello, John Doe!", para.Text)
}

func TestProcessorUndefinedAttribute(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Hello, {firstname}!",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	assert.Equal(t, "Hello, !", para.Text)
}

func TestProcessorPredefinedAttributes(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"toc": "right",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "TOC is {toc}",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	// Document attribute overrides predefined
	val, ok := processor.GetAttribute("toc")
	assert.True(t, ok)
	assert.Equal(t, "right", val)

	// Predefined attribute still accessible
	val, ok = processor.GetAttribute("toclevels")
	assert.True(t, ok)
	assert.Equal(t, "3", val)
}

func TestProcessorSetAttribute(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{},
		Blocks:    []ast.Node{},
	}

	processor := NewProcessor(doc)
	processor.SetAttribute("newattr", "newvalue")

	val, ok := processor.GetAttribute("newattr")
	assert.True(t, ok)
	assert.Equal(t, "newvalue", val)
	assert.Equal(t, "newvalue", doc.Attributes["newattr"])
}

func TestProcessorListSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"fruit": "Apple",
		},
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind:  ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "{fruit}",
						Pos:    ast.Position{Line: 1},
					},
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "Banana",
						Pos:    ast.Position{Line: 2},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	list := doc.Blocks[0].(*ast.NodeList)
	item0 := list.Items[0].(*ast.NodeListItem)
	assert.Equal(t, "Apple", item0.Text)
}

func TestProcessorLabeledListSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"term1": "First Term",
			"def1":  "First Definition",
		},
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind: ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:      ast.TypeListItem,
						Marker:     "::",
						Term:       "{term1}",
						Definition:  "{def1}",
						Pos:        ast.Position{Line: 1},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	list := doc.Blocks[0].(*ast.NodeList)
	item := list.Items[0].(*ast.NodeListItem)
	assert.Equal(t, "First Term", item.Term)
	assert.Equal(t, "First Definition", item.Definition)
}

func TestProcessorSectionSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"section-title": "My Section",
		},
		Blocks: []ast.Node{
			&ast.NodeSection{
				Level: 1,
				Title: "{section-title}",
				Pos:   ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	section := doc.Blocks[0].(*ast.NodeSection)
	assert.Equal(t, "My Section", section.Title)
}

func TestProcessorLiteralBlockNoSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"code": "value",
		},
		Blocks: []ast.Node{
			&ast.NodeLiteral{
				Kind: ast.TypeLiteral,
				Lines: []string{"This is {code} text", "Another line"},
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	literal := doc.Blocks[0].(*ast.NodeLiteral)
	// Literal blocks should NOT have substitutions
	assert.Equal(t, "This is {code} text", literal.Lines[0])
}

func TestProcessorDefaultAttributes(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{},
		Blocks:    []ast.Node{},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	// Check that defaults are set
	val, ok := processor.GetAttribute("toc")
	assert.True(t, ok)
	assert.Equal(t, "left", val)

	val, ok = processor.GetAttribute("toclevels")
	assert.True(t, ok)
	assert.Equal(t, "3", val)
}

func TestProcessorGetAllAttributes(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"custom": "value",
			"toc":    "right",
		},
		Blocks: []ast.Node{},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	allAttrs := processor.GetAllAttributes()

	// Custom attribute overrides predefined
	assert.Equal(t, "right", allAttrs["toc"])
	assert.Equal(t, "value", allAttrs["custom"])

	// Other predefined attributes still present
	assert.Equal(t, "3", allAttrs["toclevels"])
}

func TestProcessorMultipleReferences(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"name": "Alice",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Hello {name}, welcome {name}!",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	assert.Equal(t, "Hello Alice, welcome Alice!", para.Text)
}

func TestProcessorAttributeWithUnderscore(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"my_attr": "value",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Test: {my_attr}",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	assert.Equal(t, "Test: value", para.Text)
}

func TestProcessorConsecutiveReferences(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"a": "1",
			"b": "2",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "{a}{b}",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	assert.Equal(t, "12", para.Text)
}

// Tests for inline node attribute substitution

func TestProcessorInlineLinkSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"url":  "https://example.com",
			"text": "Click Here",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Visit our site",
				InlineNodes: []interface{}{
					&inline.Node{
						Type: inline.NodeLink,
						URL:  "{url}",
						Text: "{text}",
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	require.Len(t, para.InlineNodes, 1)
	link := para.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "https://example.com", link.URL)
	assert.Equal(t, "Click Here", link.Text)
}

func TestProcessorInlineBoldSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"name": "John Doe",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Hello, **bold** text",
				InlineNodes: []interface{}{
					&inline.Node{
						Type: inline.NodeBold,
						Text: "{name}",
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	require.Len(t, para.InlineNodes, 1)
	bold := para.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "John Doe", bold.Text)
}

func TestProcessorInlineImageSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"img-path": "images/photo.png",
			"img-alt":  "A photo",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "See the image",
				InlineNodes: []interface{}{
					&inline.Node{
						Type: inline.NodeImage,
						URL:  "{img-path}",
						Alt:  "{img-alt}",
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	require.Len(t, para.InlineNodes, 1)
	img := para.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "images/photo.png", img.URL)
	assert.Equal(t, "A photo", img.Alt)
}

func TestProcessorInlineCrossRefSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"section-id": "introduction",
			"link-text":  "Introduction Section",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "See the section",
				InlineNodes: []interface{}{
					&inline.Node{
						Type:    inline.NodeCrossRef,
						Ref:     "{section-id}",
						RefText: "{link-text}",
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	require.Len(t, para.InlineNodes, 1)
	xref := para.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "introduction", xref.Ref)
	assert.Equal(t, "Introduction Section", xref.RefText)
}

func TestProcessorInlineNestedChildrenSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"inner": "nested text",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Bold italic text",
				InlineNodes: []interface{}{
					&inline.Node{
						Type: inline.NodeBoldItalic,
						Text: "outer",
						Children: []*inline.Node{
							{
								Type: inline.NodeText,
								Text: "{inner}",
							},
						},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	require.Len(t, para.InlineNodes, 1)
	boldItalic := para.InlineNodes[0].(*inline.Node)
	require.Len(t, boldItalic.Children, 1)
	child := boldItalic.Children[0]
	assert.Equal(t, "nested text", child.Text)
}

func TestProcessorInlineMacroAttributesSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"icon-name": "github",
			"size":      "2x",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "Icon",
				InlineNodes: []interface{}{
					&inline.Node{
						Type:       inline.NodeCustomMacro,
						MacroName:  "icon",
						MacroTarget: "{icon-name}",
						MacroAttrs: map[string]string{
							"size": "{size}",
						},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	para := doc.Blocks[0].(*ast.NodeParagraph)
	require.Len(t, para.InlineNodes, 1)
	macro := para.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "github", macro.MacroTarget)
	assert.Equal(t, "2x", macro.MacroAttrs["size"])
}

func TestProcessorListItemInlineSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"item-text": "First Item",
		},
		Blocks: []ast.Node{
			&ast.NodeList{
				Kind: ast.TypeList,
				Items: []ast.Node{
					&ast.NodeListItem{
						Kind:   ast.TypeListItem,
						Marker: "-",
						Text:   "{item-text}",
						InlineNodes: []interface{}{
							&inline.Node{
								Type: inline.NodeBold,
								Text: "{item-text}",
							},
						},
						Pos: ast.Position{Line: 1},
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	list := doc.Blocks[0].(*ast.NodeList)
	item := list.Items[0].(*ast.NodeListItem)
	assert.Equal(t, "First Item", item.Text)
	require.Len(t, item.InlineNodes, 1)
	bold := item.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "First Item", bold.Text)
}

func TestProcessorBibliographyEntryInlineSubstitution(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"author": "John Doe",
		},
		Blocks: []ast.Node{
			&ast.BibliographyEntryNode{
				Label:    "pp",
				Text:     "Book by {author}",
				InlineNodes: []interface{}{
					&inline.Node{
						Type: inline.NodeItalic,
						Text: "{author}",
					},
				},
				Pos: ast.Position{Line: 1},
			},
		},
	}

	processor := NewProcessor(doc)
	err := processor.Process()
	require.NoError(t, err)

	entry := doc.Blocks[0].(*ast.BibliographyEntryNode)
	assert.Equal(t, "Book by John Doe", entry.Text)
	require.Len(t, entry.InlineNodes, 1)
	italic := entry.InlineNodes[0].(*inline.Node)
	assert.Equal(t, "John Doe", italic.Text)
}
