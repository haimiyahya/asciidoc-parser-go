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

	// Should create 1 parent list block with nested list attached to first item
	assert.Len(t, doc.Blocks, 1)

	// First block is the parent list
	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok, "First block should be a NodeList")
	assert.Len(t, list.Items, 2, "Parent list should have 2 items")

	// First parent item should have a nested list
	parentItem1, ok1 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok1, "First item should be a NodeListItem")
	assert.Equal(t, "Parent item 1", parentItem1.Text)
	assert.NotNil(t, parentItem1.NestedList, "First item should have a NestedList")

	// Verify nested list structure
	nestedList := parentItem1.NestedList
	assert.Len(t, nestedList.Items, 2, "Nested list should have 2 items")

	nestedItem1, ok2 := nestedList.Items[0].(*ast.NodeListItem)
	require.True(t, ok2, "Nested item should be a NodeListItem")
	assert.Equal(t, "Nested item 1", nestedItem1.Text)

	nestedItem2, ok3 := nestedList.Items[1].(*ast.NodeListItem)
	require.True(t, ok3, "Nested item should be a NodeListItem")
	assert.Equal(t, "Nested item 2", nestedItem2.Text)

	// Second parent item should not have a nested list
	parentItem2, ok4 := list.Items[1].(*ast.NodeListItem)
	require.True(t, ok4, "Second item should be a NodeListItem")
	assert.Equal(t, "Parent item 2", parentItem2.Text)
	assert.Nil(t, parentItem2.NestedList, "Second item should not have a NestedList")
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

func TestParseNestedOrderedList(t *testing.T) {
	source := `. Parent item 1
  .. Nested item 1
  .. Nested item 2
. Parent item 2
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

	parentItem1, ok1 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok1)
	assert.NotNil(t, parentItem1.NestedList)
	assert.Len(t, parentItem1.NestedList.Items, 2)
}

func TestParseMixedNestedListTypes(t *testing.T) {
	source := `- Unordered parent
  .. Nested ordered item 1
  .. Nested ordered item 2
- Another unordered item
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

	parentItem1, ok1 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok1)
	assert.Equal(t, "-", parentItem1.Marker)
	assert.NotNil(t, parentItem1.NestedList)

	// Nested list should be ordered (using ".." marker)
	nestedList := parentItem1.NestedList
	assert.Len(t, nestedList.Items, 2)

	nestedItem1, ok2 := nestedList.Items[0].(*ast.NodeListItem)
	require.True(t, ok2)
	assert.Equal(t, "..", nestedItem1.Marker)
	assert.Equal(t, "Nested ordered item 1", nestedItem1.Text)
}

func TestParseTripleNestedList(t *testing.T) {
	t.Skip("TODO: Parser needs to track nested list stack for deep nesting")

	source := `- Level 1 item
  - Level 2 item
    - Level 3 item
      - Level 4 item
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have one parent list
	assert.Len(t, doc.Blocks, 1)

	list, ok := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok)
	assert.Len(t, list.Items, 1)

	// Level 1 item
	level1Item, ok1 := list.Items[0].(*ast.NodeListItem)
	require.True(t, ok1)
	assert.Equal(t, "Level 1 item", level1Item.Text)
	assert.NotNil(t, level1Item.NestedList)

	// Level 2 list (nested in level 1 item)
	level2List := level1Item.NestedList
	assert.Len(t, level2List.Items, 1)

	level2Item, ok2 := level2List.Items[0].(*ast.NodeListItem)
	require.True(t, ok2)
	assert.Equal(t, "Level 2 item", level2Item.Text)
	assert.NotNil(t, level2Item.NestedList)

	// Level 3 list (nested in level 2 item)
	level3List := level2Item.NestedList
	assert.Len(t, level3List.Items, 1)

	level3Item, ok3 := level3List.Items[0].(*ast.NodeListItem)
	require.True(t, ok3)
	assert.Equal(t, "Level 3 item", level3Item.Text)
	assert.NotNil(t, level3Item.NestedList)

	// Level 4 list (nested in level 3 item)
	level4List := level3Item.NestedList
	assert.Len(t, level4List.Items, 1)
}

func TestParseNestedListWithBlankLine(t *testing.T) {
	source := `- Parent 1
  - Child 1
  - Child 2

- Parent 2
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have 2 separate lists (blank line terminates the first)
	assert.Len(t, doc.Blocks, 2)

	// First list has nested children
	list1, ok1 := doc.Blocks[0].(*ast.NodeList)
	require.True(t, ok1)
	assert.Len(t, list1.Items, 1)

	item1, ok2 := list1.Items[0].(*ast.NodeListItem)
	require.True(t, ok2)
	assert.NotNil(t, item1.NestedList)
	assert.Len(t, item1.NestedList.Items, 2)
}
