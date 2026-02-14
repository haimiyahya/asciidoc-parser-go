// Package parser provides tests for table parsing.
package parser

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSimpleTable(t *testing.T) {
	source := `|===
| A | B | C |
| D | E | F |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.Blocks, 1)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok, "First block should be a Table")
	assert.Equal(t, ast.TypeTable, table.Type())

	// Check header (first non-empty row)
	assert.Len(t, table.Header, 3)
	assert.Equal(t, "A", table.Header[0])
	assert.Equal(t, "B", table.Header[1])
	assert.Equal(t, "C", table.Header[2])

	// Check body
	assert.Len(t, table.Body, 1)
	assert.Len(t, table.Body[0], 3)
	assert.Equal(t, "D", table.Body[0][0])
	assert.Equal(t, "E", table.Body[0][1])
	assert.Equal(t, "F", table.Body[0][2])
}

func TestParseTableWithMultipleRows(t *testing.T) {
	source := `|===
| Name | Age | City |
| Alice | 30 | NYC |
| Bob | 25 | LA |
| Charlie | 35 | Chicago |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Header
	assert.Len(t, table.Header, 3)
	assert.Equal(t, "Name", table.Header[0])
	assert.Equal(t, "Age", table.Header[1])
	assert.Equal(t, "City", table.Header[2])

	// Body - 3 rows
	assert.Len(t, table.Body, 3)
	assert.Equal(t, "Alice", table.Body[0][0])
	assert.Equal(t, "Bob", table.Body[1][0])
	assert.Equal(t, "Charlie", table.Body[2][0])
}

func TestParseTableWithEmptyCells(t *testing.T) {
	source := `|===
| A | | C |
| D | F |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Header with empty middle cell
	assert.Len(t, table.Header, 3)
	assert.Equal(t, "A", table.Header[0])
	assert.Equal(t, "", table.Header[1])
	assert.Equal(t, "C", table.Header[2])

	// Body
	assert.Len(t, table.Body, 1)
	assert.Len(t, table.Body[0], 2) // Body row has only 2 non-empty cells
	assert.Equal(t, "D", table.Body[0][0])
	assert.Equal(t, "F", table.Body[0][1])
}

func TestParseTableWithLeadingPipe(t *testing.T) {
	source := `|===
| A | B | C |
| D | E | F |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should handle leading/trailing pipes correctly
	assert.Len(t, table.Header, 3)
	assert.Equal(t, "A", table.Header[0])
	assert.Equal(t, "B", table.Header[1])
	assert.Equal(t, "C", table.Header[2])
}

func TestParseTableEmptyLines(t *testing.T) {
	source := `|===
| A | B | C |

| D | E | F |

| G | H | I |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Empty lines in tables should be skipped
	assert.Len(t, table.Header, 3)
	assert.Len(t, table.Body, 2) // Only 2 non-empty rows
}
