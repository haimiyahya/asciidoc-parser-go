// Package processor tests document processing functionality.
package processor

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
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
