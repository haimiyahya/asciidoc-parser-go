// Package ast defines the table-related Abstract Syntax Tree nodes for AsciiDoc documents.
package ast

// Table is a complete table with support for AsciiDoc table syntax.
type Table struct {
	// Caption is the optional table caption.
	Caption string

	// ID is the optional table ID for cross-references.
	ID string

	// Columns defines column specifications (width, alignment, etc.)
	Columns []TableColumnSpec

	// Rows contains all table rows (header + body + footer)
	Rows []TableRow

	// HeaderRowIndex is the index of the header row (-1 if no header)
	HeaderRowIndex int

	// FooterRowIndex is the index of the footer row (-1 if no footer)
	FooterRowIndex int

	// Attributes are table-level attributes
	Attributes map[string]string

	// Pos is the location in the source.
	Pos Position
}

// TableColumnSpec defines column specifications.
type TableColumnSpec struct {
	// Width is the column width (percentage or auto)
	Width string

	// HorizontalAlign is the horizontal alignment (<, >, ^, .)
	HorizontalAlign string

	// VerticalAlign is the vertical alignment (<, >, ^, .)
	VerticalAlign string

	// Style is the column style (header, strong, emphasis, monospace, etc.)
	Style string

	// AutoWidth indicates if width should be automatically determined
	AutoWidth bool
}

// TableRow represents a single table row.
type TableRow struct {
	// Cells contains the cells in this row
	Cells []TableCell

	// Kind indicates if this is a header, body, or footer row
	Kind TableRowKind

	// Attributes are row-level attributes
	Attributes map[string]string
}

// TableRowKind represents the type of table row.
type TableRowKind int

const (
	// TableRowBody is a regular body row
	TableRowBody TableRowKind = iota

	// TableRowHeader is a header row
	TableRowHeader

	// TableRowFooter is a footer row
	TableRowFooter
)

// TableCell represents a single table cell.
type TableCell struct {
	// Text is the cell content
	Text string

	// InlineNodes contains inline markup nodes within the cell
	InlineNodes []interface{}

	// Blocks contains block-level content (lists, etc.) within the cell
	Blocks []Node

	// ColSpan is the number of columns this cell spans (default: 1)
	ColSpan int

	// RowSpan is the number of rows this cell spans (default: 1)
	RowSpan int

	// HorizontalAlign is the horizontal alignment (<, >, ^, .)
	HorizontalAlign string

	// VerticalAlign is the vertical alignment (<, >, ^, .)
	VerticalAlign string

	// Style is the cell style (header, strong, emphasis, monospace, etc.)
	Style string

	// Repeat indicates this cell repeats across columns (for PSV format)
	Repeat int

	// Attributes are cell-level attributes
	Attributes map[string]string
}

// TableFormat represents the table data format.
type TableFormat int

const (
	// FormatPSV is Pipe Separated Values (default AsciiDoc format)
	FormatPSV TableFormat = iota

	// FormatCSV is Comma Separated Values
	FormatCSV

	// FormatTSV is Tab Separated Values
	FormatTSV

	// FormatDSV is Colon Separated Values
	FormatDSV
)

// TableFrame specifies the outer frame style.
type TableFrame string

const (
	// FrameAll shows borders on all sides (default)
	FrameAll TableFrame = "all"

	// FrameTopbot shows borders only on top and bottom
	FrameTopbot TableFrame = "topbot"

	// FrameSides shows borders only on left and right
	FrameSides TableFrame = "sides"

	// FrameNone shows no outer borders
	FrameNone TableFrame = "none"
)

// TableGrid specifies the internal grid style.
type TableGrid string

const (
	// GridAll shows borders between all rows and columns (default)
	GridAll TableGrid = "all"

	// GridRows shows borders only between rows
	GridRows TableGrid = "rows"

	// GridCols shows borders only between columns
	GridCols TableGrid = "cols"

	// GridNone shows no internal borders
	GridNone TableGrid = "none"
)

// GetFrame returns the table frame attribute.
func (t *Table) GetFrame() TableFrame {
	if t.Attributes != nil {
		if frame, ok := t.Attributes["frame"]; ok {
			return TableFrame(frame)
		}
	}
	return FrameAll // Default
}

// GetGrid returns the table grid attribute.
func (t *Table) GetGrid() TableGrid {
	if t.Attributes != nil {
		if grid, ok := t.Attributes["grid"]; ok {
			return TableGrid(grid)
		}
	}
	return GridAll // Default
}

// GetFormat returns the table data format.
func (t *Table) GetFormat() TableFormat {
	if t.Attributes != nil {
		if format, ok := t.Attributes["format"]; ok {
			switch format {
			case "csv", "CSV":
				return FormatCSV
			case "tsv", "TSV":
				return FormatTSV
			case "dsv", "DSV":
				return FormatDSV
			}
		}
	}
	return FormatPSV // Default
}

// GetStripes returns the table stripes attribute for row styling.
func (t *Table) GetStripes() string {
	if t.Attributes != nil {
		if stripes, ok := t.Attributes["stripes"]; ok {
			return stripes
		}
	}
	return "none" // Default
}

// GetWidth returns the table width attribute.
func (t *Table) GetWidth() string {
	if t.Attributes != nil {
		if width, ok := t.Attributes["width"]; ok {
			return width
		}
	}
	return "" // Default (auto)
}

// GetOptions returns the table options attribute.
func (t *Table) GetOptions() string {
	if t.Attributes != nil {
		if options, ok := t.Attributes["options"]; ok {
			return options
		}
	}
	return "" // Default
}

// HasHeader returns true if the table has a header row.
func (t *Table) HasHeader() bool {
	return t.HeaderRowIndex >= 0 && t.HeaderRowIndex < len(t.Rows)
}

// HeaderRow returns the header row if present.
func (t *Table) HeaderRow() *TableRow {
	if t.HasHeader() {
		return &t.Rows[t.HeaderRowIndex]
	}
	return nil
}

// HasFooter returns true if the table has a footer row.
func (t *Table) HasFooter() bool {
	return t.FooterRowIndex >= 0 && t.FooterRowIndex < len(t.Rows)
}

// FooterRow returns the footer row if present.
func (t *Table) FooterRow() *TableRow {
	if t.HasFooter() {
		return &t.Rows[t.FooterRowIndex]
	}
	return nil
}

// BodyRows returns all body rows (excluding header and footer).
func (t *Table) BodyRows() []TableRow {
	var body []TableRow
	for i, row := range t.Rows {
		if i != t.HeaderRowIndex && i != t.FooterRowIndex {
			body = append(body, row)
		}
	}
	return body
}

// ColumnCount returns the number of columns in the table.
func (t *Table) ColumnCount() int {
	if len(t.Rows) == 0 {
		return 0
	}
	maxCols := 0
	for _, row := range t.Rows {
		if len(row.Cells) > maxCols {
			maxCols = len(row.Cells)
		}
	}
	return maxCols
}

// Type returns the node type for table (satisfies Node interface).
func (t *Table) Type() NodeType {
	return TypeTable
}

// Position returns the source location for table (satisfies Node interface).
func (t *Table) Position() Position {
	return t.Pos
}
