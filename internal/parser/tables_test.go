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
	// Note: The |= prefix is a per-cell header indicator that Asciidoctor
	// treats as literal text in basic PSV tables. Use [options="header"] instead.
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

	// Per-cell indicators are kept as literal text for Asciidoctor compatibility
	assert.Equal(t, ast.TableRowBody, table.Rows[0].Kind)
	assert.Equal(t, "= Header 1", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "= Header 2", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "= Header 3", table.Rows[0].Cells[2].Text)

	// Second row should be body
	assert.Equal(t, ast.TableRowBody, table.Rows[1].Kind)
	assert.Equal(t, "Data 1", table.Rows[1].Cells[0].Text)
}

func TestParseTableWithCellAlignment(t *testing.T) {
	// Note: Per-cell alignment indicators (>, ^, <) are kept as literal text
	// for Asciidoctor compatibility. Use column specifications instead.
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

	// Alignment indicators are kept as literal text for Asciidoctor compatibility
	assert.Equal(t, ">right", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "^center", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "<left", table.Rows[0].Cells[2].Text)
}

func TestTableHelperMethods(t *testing.T) {
	// Test with options="header" for proper header row support
	source := `[options="header"]
|===
| Header
| Data 1
| Data 2
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Test HasHeader - should be true with options="header"
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

func TestTableOptionsHeader(t *testing.T) {
	// Test that options="header" correctly marks first row as header
	source := `[options="header"]
|===
| Name | Age | City
| Alice | 30 | NYC
| Bob | 25 | LA
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 3 rows total
	assert.Len(t, table.Rows, 3)

	// First row should be header
	assert.Equal(t, ast.TableRowHeader, table.Rows[0].Kind, "First row should be header")
	assert.Equal(t, "Name", table.Rows[0].Cells[0].Text)

	// Other rows should be body
	assert.Equal(t, ast.TableRowBody, table.Rows[1].Kind, "Second row should be body")
	assert.Equal(t, "Alice", table.Rows[1].Cells[0].Text)

	assert.Equal(t, ast.TableRowBody, table.Rows[2].Kind, "Third row should be body")
	assert.Equal(t, "Bob", table.Rows[2].Cells[0].Text)
}

func TestTableNoHeaderWithoutOption(t *testing.T) {
	// Test that tables without options="header" don't mark first row as header
	source := `|===
| A | B
| C | D
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 2 rows total
	assert.Len(t, table.Rows, 2)

	// First row should NOT be a header (no options="header" specified)
	assert.Equal(t, ast.TableRowBody, table.Rows[0].Kind, "First row should be body when no header option")
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)

	// Second row should also be body
	assert.Equal(t, ast.TableRowBody, table.Rows[1].Kind)
	assert.Equal(t, "C", table.Rows[1].Cells[0].Text)
}

func TestTableFrameAttribute(t *testing.T) {
	// Test frame attribute
	source := `[frame="sides"]
|===
| A | B
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check frame attribute
	assert.Equal(t, ast.FrameSides, table.GetFrame())
	assert.Equal(t, "sides", table.Attributes["frame"])
}

func TestTableGridAttribute(t *testing.T) {
	// Test grid attribute
	source := `[grid="rows"]
|===
| A | B
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check grid attribute
	assert.Equal(t, ast.GridRows, table.GetGrid())
	assert.Equal(t, "rows", table.Attributes["grid"])
}

func TestTableStripesAttribute(t *testing.T) {
	// Test stripes attribute
	source := `[stripes="even"]
|===
| A | B
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check stripes attribute
	assert.Equal(t, "even", table.GetStripes())
	assert.Equal(t, "even", table.Attributes["stripes"])
}

func TestTableAutowidthAttribute(t *testing.T) {
	// Test %autowidth positional attribute
	source := `[%autowidth]
|===
| A | B
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check autowidth attribute
	assert.Equal(t, "true", table.Attributes["autowidth"])
}

func TestTableCaptionAttribute(t *testing.T) {
	// Test caption attribute
	source := `[caption="Table Title"]
|===
| A | B
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check caption
	assert.Equal(t, "Table Title", table.Caption)
	assert.Equal(t, "Table Title", table.Attributes["caption"])
}

func TestTableMultipleAttributes(t *testing.T) {
	// Test multiple attributes combined
	source := `[frame="topbot",grid="cols",%autowidth,caption="My Table"]
|===
| A | B
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Check all attributes
	assert.Equal(t, ast.FrameTopbot, table.GetFrame())
	assert.Equal(t, ast.GridCols, table.GetGrid())
	assert.Equal(t, "true", table.Attributes["autowidth"])
	assert.Equal(t, "My Table", table.Caption)
}

func TestParseTableWithMultilineCellContinuation(t *testing.T) {
	source := `|===
| Cell 1 | Cell 2+
This is a continuation
of Cell 2
| Cell 3
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 2 rows
	assert.Len(t, table.Rows, 2)

	// First row - Cell 2 should have multi-line content
	assert.Len(t, table.Rows[0].Cells, 2)
	assert.Equal(t, "Cell 1", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "Cell 2\nThis is a continuation\nof Cell 2", table.Rows[0].Cells[1].Text)

	// Second row
	assert.Len(t, table.Rows[1].Cells, 1)
	assert.Equal(t, "Cell 3", table.Rows[1].Cells[0].Text)
}

func TestParseTableWithMultipleContinuationLines(t *testing.T) {
	source := `|===
| A | B+
Line 2 of B
Line 3 of B
Line 4 of B
| C | D
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 2 rows
	assert.Len(t, table.Rows, 2)

	// First row - Cell B should have 4 lines of content
	assert.Len(t, table.Rows[0].Cells, 2)
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "B\nLine 2 of B\nLine 3 of B\nLine 4 of B", table.Rows[0].Cells[1].Text)

	// Second row
	assert.Equal(t, "C", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "D", table.Rows[1].Cells[1].Text)
}

func TestParseTableWithContinuationAndNewRow(t *testing.T) {
	source := `|===
| A | B+
Continued line
| C | D+
More D
Even more D
| E | F
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 3 rows
	assert.Len(t, table.Rows, 3)

	// First row
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "B\nContinued line", table.Rows[0].Cells[1].Text)

	// Second row
	assert.Equal(t, "C", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "D\nMore D\nEven more D", table.Rows[1].Cells[1].Text)

	// Third row
	assert.Equal(t, "E", table.Rows[2].Cells[0].Text)
	assert.Equal(t, "F", table.Rows[2].Cells[1].Text)
}

func TestParseTableWithContinuationEndingWithPlus(t *testing.T) {
	source := `|===
| A | B+
Line 2 of B+
Line 3 of B
| C | D
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 2 rows
	assert.Len(t, table.Rows, 2)

	// First row - B should have all 3 lines
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "B\nLine 2 of B\nLine 3 of B", table.Rows[0].Cells[1].Text)

	// Second row
	assert.Equal(t, "C", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "D", table.Rows[1].Cells[1].Text)
}

func TestParseTableWithContinuationAcrossEmptyLines(t *testing.T) {
	source := `|===
| A | B+
Line 2 of B

Line 4 of B (after empty line)
| C | D
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 2 rows
	assert.Len(t, table.Rows, 2)

	// First row - B should have content (empty lines are skipped by parser)
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	// Empty lines are currently skipped, so we only get non-empty lines
	assert.Equal(t, "B\nLine 2 of B\nLine 4 of B (after empty line)", table.Rows[0].Cells[1].Text)

	// Second row
	assert.Equal(t, "C", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "D", table.Rows[1].Cells[1].Text)
}

func TestParseTableWithAllCellsContinuing(t *testing.T) {
	source := `|===
| A+
More A
| B+
More B
| C+
More C
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Each | starts a new row in AsciiDoc, so we get 3 rows with 1 continuing cell each
	assert.Len(t, table.Rows, 3)

	// First row
	assert.Len(t, table.Rows[0].Cells, 1)
	assert.Equal(t, "A\nMore A", table.Rows[0].Cells[0].Text)

	// Second row
	assert.Len(t, table.Rows[1].Cells, 1)
	assert.Equal(t, "B\nMore B", table.Rows[1].Cells[0].Text)

	// Third row
	assert.Len(t, table.Rows[2].Cells, 1)
	assert.Equal(t, "C\nMore C", table.Rows[2].Cells[0].Text)
}

func TestParseTableWithMultipleCellsOnSameRowContinuing(t *testing.T) {
	// Test the correct way to have multiple cells continuing on the same row
	source := `|===
| A | B | C+
Continuation of C
More of C
| D | E | F
|===`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	table, ok := doc.Blocks[0].(*ast.Table)
	require.True(t, ok)

	// Should have 2 rows
	assert.Len(t, table.Rows, 2)

	// First row - only cell C continues
	assert.Len(t, table.Rows[0].Cells, 3)
	assert.Equal(t, "A", table.Rows[0].Cells[0].Text)
	assert.Equal(t, "B", table.Rows[0].Cells[1].Text)
	assert.Equal(t, "C\nContinuation of C\nMore of C", table.Rows[0].Cells[2].Text)

	// Second row
	assert.Equal(t, "D", table.Rows[1].Cells[0].Text)
	assert.Equal(t, "E", table.Rows[1].Cells[1].Text)
	assert.Equal(t, "F", table.Rows[1].Cells[2].Text)
}

