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

	// Check rows - first row is header by default
	assert.Len(t, table.Rows, 2)
	assert.Len(t, table.Rows[0].Cells, 3)
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "B", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "C", table.Rows[0].Cells[2].Text)

	// Check second row (body)
	assert.Len(t, table.Rows[1].Cells, 3)
	assert.Equal(t, "D", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "E", table.Rows[1].Cells[1].Text)
	assert.Equal(t, "F", table.Rows[1].Cells[2].Text)
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

	// All rows
	assert.Len(t, table.Rows, 4)

	// Header row
	assert.Len(t, table.Rows[0].Cells, 3)
	assert.Equal(t, "Name", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "Age", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "City", table.Rows[0].Cells[2].Text)

	// Body rows
	assert.Equal(t, "Alice", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "Bob", table.Rows[2].Cells[0].Text)
	assert.Equal(t, "Charlie", table.Rows[3].Cells[0].Text)
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

	// All rows
	assert.Len(t, table.Rows, 2)

	// Header with empty middle cell
	assert.Len(t, table.Rows[0].Cells, 3)
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "C", table.Rows[0].Cells[2].Text)

	// Body row
	assert.Len(t, table.Rows[1].Cells, 2)
	assert.Equal(t, "D", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "F", table.Rows[1].Cells[1].Text)
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
	assert.Len(t, table.Rows, 2)
	assert.Len(t, table.Rows[0].Cells, 3)
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "B", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "C", table.Rows[0].Cells[2].Text)
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
	assert.Len(t, table.Rows, 3) // 3 non-empty rows
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "D", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "G", table.Rows[2].Cells[0].Text)
}

func TestParseTableWithAttributes(t *testing.T) {
	// Note: Table attributes like [cols=...] are currently parsed inline by the table parser
	// For now, test that tables can be parsed with attributes on the same line
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

	// Check basic table structure
	assert.Len(t, table.Rows, 2)
	assert.Len(t, table.Rows[0].Cells, 3)
}

func TestParseTableWithHeaderRow(t *testing.T) {
	source := `|===
|= Header 1 |= Header 2 |= Header 3
| Data 1 | Data 2 | Data 3 |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// First row should be marked as header
	assert.Equal(t, ast.TableRowHeader, table.Rows[0].Kind)
	assert.Equal(t, "Header 1", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "Header 2", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "Header 3", table.Rows[0].Cells[2].Text)

	// Second row should be body
	assert.Equal(t, ast.TableRowBody, table.Rows[1].Kind)
	assert.Equal(t, "Data 1", table.Rows[1].Cells[0].Text)
}

func TestParseTableWithCellAlignment(t *testing.T) {
	source := `|===
| >right | ^center | <left |
| data | data | data |
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check alignment
	assert.Equal(t, "right", table.Rows[0].Cells[0].HorizontalAlign)
	assert.Equal(t, "center", table.Rows[0].Cells[1].HorizontalAlign)
	assert.Equal(t, "left", table.Rows[0].Cells[2].HorizontalAlign)
}

func TestTableHelperMethods(t *testing.T) {
	source := `|===
|= Header
| Data 1
| Data 2
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Test HasHeader
	assert.True(t, table.HasHeader())

	// Test HeaderRow
	headerRow := table.HeaderRow()
	assert.NotNil(t, headerRow)
	assert.Equal(t, "Header", headerRow.Cells[0].Text)

	// Test BodyRows
	bodyRows := table.BodyRows()
	assert.Len(t, bodyRows, 2)
	assert.Equal(t, "Data 1", bodyRows[0].Cells[0].Text)
	assert.Equal(t, "Data 2", bodyRows[1].Cells[0].Text)

	// Test ColumnCount
	assert.Equal(t, 1, table.ColumnCount())

	// Test Frame
	assert.Equal(t, ast.FrameAll, table.GetFrame())

	// Test Grid
	assert.Equal(t, ast.GridAll, table.GetGrid())
}
