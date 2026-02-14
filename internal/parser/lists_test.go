// Package parser provides tests for list parsing.
package parser

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUnorderedList(t *testing.T) {
	source := `- Item 1
- Item 2
- Item 3
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have 1 block (the list)
	assert.Len(t, doc.Blocks, 1)

	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok, "First block should be a NodeList")
	assert.Equal(t, ast.TypeList, list.Kind)
	assert.Len(t, list.Items, 3)

	// Check each item
	item0, ok0 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok0)
	assert.Equal(t, "-", item0.Marker)
	assert.Equal(t, "Item 1", item0.Text)

	item1, ok1 := list.Items[1].(*ast.NodeListItem)
	require.True(t, ok1)
	assert.Equal(t, "-", item1.Marker)
	assert.Equal(t, "Item 2", item1.Text)

	item2, ok2 := list.Items[2].(*ast.NodeListItem)
	require.True(t, ok2)
	assert.Equal(t, "-", item2.Marker)
	assert.Equal(t, "Item 3", item2.Text)
}

func TestParseAsteriskUnorderedList(t *testing.T) {
	source := `* First item
* Second item
* Third item
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.Blocks, 1)

	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok)
	assert.Equal(t, ast.TypeList, list.Kind)
	assert.Len(t, list.Items, 3)

	item0, ok0 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok0)
	assert.Equal(t, "*", item0.Marker)
}

func TestParseOrderedList(t *testing.T) {
	source := `. First item
. Second item
. Third item
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.Blocks, 1)

	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok)
	assert.Equal(t, ast.TypeList, list.Kind)
	assert.Len(t, list.Items, 3)

	item0, ok0 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok0)
	assert.Equal(t, ".", item0.Marker)
	assert.Equal(t, 1, item0.Ordinal)
}

func TestParseMixedListsSeparate(t *testing.T) {
	source := `- Unordered item 1
- Unordered item 2

. Ordered item 1
. Ordered item 2
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have 2 separate lists
	assert.Len(t, doc.Blocks, 2)

	list0, ok0 := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok0)
	assert.Len(t, list0.Items, 2)

	list1, ok1 := doc.Blocks[1].(*ast.NodeList)
	require.True(t, ok1)
	assert.Len(t, list1.Items, 2)
}

func TestParseParagraphsAroundList(t *testing.T) {
	source := `This is a paragraph.

- List item 1
- List item 2

This is another paragraph.
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have: paragraph, list, paragraph
	assert.Len(t, doc.Blocks, 3)

	// First is a paragraph
	_, ok0 := doc.Blocks[0].(*ast.NodeParagraph)
	require.True(t, ok0, "First block should be a paragraph")

	// Second is a list
	list, ok1 := doc.Blocks[1].(*ast.NodeList)
	require.True(t, ok1, "Second block should be a list")
	assert.Len(t, list.Items, 2)

	// Third is a paragraph
	_, ok2 := doc.Blocks[2].(*ast.NodeParagraph)
	require.True(t, ok2, "Third block should be a paragraph")
}

func TestParseNestedList(t *testing.T) {
	source := `- Parent item 1
  - Nested item 1
  - Nested item 2
- Parent item 2
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should create at least 2 blocks (parent list + nested list)
	assert.GreaterOrEqual(t, len(doc.Blocks), 2)

	// First block is the parent list
	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok)
	assert.Len(t, list.Items, 2)
}

func TestParseLabeledList(t *testing.T) {
	source := `Term 1 :: Definition 1
Term 2 :: Definition 2
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.Blocks, 1)

	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok)
	assert.Len(t, list.Items, 2)

	item0, ok0 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok0)
	assert.Equal(t, "::", item0.Marker)
	assert.Equal(t, "Term 1", item0.Text)
}

func TestListBlankLineTerminates(t *testing.T) {
	source := `- Item 1
- Item 2

- Item 3
- Item 4
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should create 2 separate lists (blank line terminates first)
	assert.Len(t, doc.Blocks, 2)

	list0, ok0 := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok0)
	assert.Len(t, list0.Items, 2)

	list1, ok1 := doc.Blocks[1].(*ast.NodeList)
	require.True(t, ok1)
	assert.Len(t, list1.Items, 2)
}

func TestComplexDocumentWithLists(t *testing.T) {
	source := `= Document Title

This is a paragraph.

== Section One

This is a paragraph after the section.

- List item 1
- List item 2
- List item 3

Another paragraph.

* Another list
* With items
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Count blocks (title is section, lists are grouped)
	assert.GreaterOrEqual(t, len(doc.Blocks), 6)
}
