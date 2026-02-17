// Package parser provides table parsing functionality for AsciiDoc.
package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
)

// TableParser handles parsing of AsciiDoc tables.
type TableParser struct {
	// No state needed - inline parser is created per cell
}

// NewTableParser creates a new table parser.
func NewTableParser() *TableParser {
	return &TableParser{}
}

// ParseTable parses table content from lines.
// The lines should include the opening delimiter and content, but not the closing delimiter.
func (p *TableParser) ParseTable(lines []string, lineno int) *ast.Table {
	if len(lines) == 0 {
		return nil
	}

	table := &ast.Table{
		Rows:       make([]ast.TableRow, 0),
		Attributes: make(map[string]string),
		Pos:        ast.Position{Line: lineno},
	}

	// First line might contain table attributes
	lineIdx := 0
	firstLine := strings.TrimSpace(lines[lineIdx])

	// Check if first line contains attributes (e.g., [cols="1,2,3"])
	if strings.HasPrefix(firstLine, "[") {
		attrs := p.parseTableAttributes(firstLine)
		for k, v := range attrs {
			table.Attributes[k] = v
		}
		lineIdx++
	}

	// Parse column specs if provided
	if colsSpec, ok := table.Attributes["cols"]; ok {
		table.Columns = p.parseColumnSpecs(colsSpec)
	}

	// Parse rows
	for i := lineIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for row separator
		if strings.HasPrefix(line, "|===") || strings.HasPrefix(line, "|+++") || strings.HasPrefix(line, "|---") {
			continue
		}

		// Check for cell continuation: |+
		// This continues the last cell of the previous row
		if strings.HasPrefix(line, "|+") || line == "|+" {
			p.handleCellContinuation(table, lines[i], i)
			continue
		}

		// Determine row kind for this specific row
		rowKind := ast.TableRowBody

		// Note: In AsciiDoc, |= is a per-cell header indicator, not a row indicator.
		// For basic table compatibility, we treat all rows as body rows.
		// The |= prefix is kept as literal text in the cell.

		// Check for footer row specification
		if strings.HasPrefix(line, "|_") {
			rowKind = ast.TableRowFooter
			line = strings.TrimPrefix(line, "|_")
			line = strings.TrimLeft(line, "_")
		}

		isHeaderRow := false // Always false for basic tables

		// Parse the row
		format := table.GetFormat()
		row := p.parseRow(line, format, isHeaderRow)

		// Set row kind
		row.Kind = rowKind

		// Track header row index
		if row.Kind == ast.TableRowHeader && table.HeaderRowIndex < 0 {
			table.HeaderRowIndex = len(table.Rows)
		}

		// Track footer row index
		if row.Kind == ast.TableRowFooter {
			table.FooterRowIndex = len(table.Rows)
		}

		table.Rows = append(table.Rows, row)
	}

	// If no explicit header but columns specified, first row might be header
	if table.HeaderRowIndex < 0 && len(table.Rows) > 0 && len(table.Columns) > 0 {
		// Check if first row should be header based on column specs
		hasHeaderStyle := false
		for _, col := range table.Columns {
			if col.Style == "header" {
				hasHeaderStyle = true
				break
			}
		}
		if hasHeaderStyle {
			table.HeaderRowIndex = 0
			table.Rows[0].Kind = ast.TableRowHeader
		}
	}

	// Parse caption if provided
	if caption, ok := table.Attributes["caption"]; ok {
		table.Caption = caption
	}

	return table
}

// parseTableAttributes parses attributes from the first line of a table.
// Example: [cols="1,2,3",frame="all",grid="rows"]
func (p *TableParser) parseTableAttributes(line string) map[string]string {
	attrs := make(map[string]string)

	// Remove opening [ and closing ]
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return attrs
	}

	line = line[1 : len(line)-1]

	// Split by commas that are not inside quotes
	parts := p.splitByCommas(line)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by =
		if idx := strings.Index(part, "="); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])

			// Remove quotes from value
			value = strings.Trim(value, "\"'")
			attrs[key] = value
		} else {
			// Boolean attribute
			attrs[part] = "true"
		}
	}

	return attrs
}

// splitByCommas splits a string by commas, respecting quotes.
func (p *TableParser) splitByCommas(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range s {
		switch r {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = rune(0)
			} else {
				current.WriteRune(r)
			}
		case ',':
			if !inQuotes {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseColumnSpecs parses the cols attribute into column specifications.
// Example: "1,2,3" or "<,>,^" or "h,l,l"
func (p *TableParser) parseColumnSpecs(spec string) []ast.TableColumnSpec {
	// Remove any surrounding quotes
	spec = strings.Trim(spec, "\"'")

	// Split by comma
	parts := strings.Split(spec, ",")
	cols := make([]ast.TableColumnSpec, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		col := ast.TableColumnSpec{
			AutoWidth: true,
		}

		// Parse each part
		for _, r := range part {
			switch r {
			case '<':
				col.HorizontalAlign = "left"
			case '>':
				col.HorizontalAlign = "right"
			case '^':
				col.HorizontalAlign = "center"
			case '.':
				col.HorizontalAlign = "left" // default
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// Width specifier - for now just mark as not auto
				col.AutoWidth = false
				col.Width = string(r)
			case 'h', 'H':
				col.Style = "header"
			case 's', 'S':
				col.Style = "strong"
			case 'e', 'E':
				col.Style = "emphasis"
			case 'm', 'M':
				col.Style = "monospace"
			case 'a', 'A':
				col.Style = "asciidoc"
			case 'l', 'L':
				col.Style = "literal"
			case 'v', 'V':
				col.Style = "verse"
			}
		}

		cols = append(cols, col)
	}

	return cols
}

// parseRow parses a single table row.
func (p *TableParser) parseRow(line string, format ast.TableFormat, isHeaderRow bool) ast.TableRow {
	var cells []ast.TableCell

	switch format {
	case ast.FormatCSV:
		cells = p.parseCSVRow(line)
	case ast.FormatTSV:
		cells = p.parseTSVRow(line)
	case ast.FormatDSV:
		cells = p.parseDSVRow(line)
	default: // FormatPSV
		cells = p.parsePSVRow(line, isHeaderRow)
	}

	return ast.TableRow{
		Cells:      cells,
		Attributes: make(map[string]string),
	}
}

// parsePSVRow parses a Pipe Separated Values row.
// Supports: | Cell 1 | Cell 2 | Cell 3 |
//           | Cell 1 | Cell 2 || Cell 3 |
//           | >right | ^center | <left |
//           |= Header 1 |= Header 2 (header row)
//           |l|Literal cell |m|Monospace cell (cell styles)
func (p *TableParser) parsePSVRow(line string, isHeaderRow bool) []ast.TableCell {
	var cells []ast.TableCell

	// Trim leading |
	line = strings.TrimLeft(line, "|")

	// Split by |, but handle || as empty cell with colspan
	parts := strings.Split(line, "|")

	// Remove trailing empty string if line ends with |
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	// Current style to apply to next cell
	var currentStyle string

	for i, part := range parts {
		trimmed := strings.TrimSpace(part)

		// Check for single-letter style indicator (l, m, v, a, e, h, s, d)
		// These indicate cell style for the following cell ONLY when:
		// - It's a single letter that IS a valid style indicator
		// - It's followed by actual cell content (not another style indicator)
		isStyle := false
		if len(trimmed) == 1 && p.isCellStyleIndicator(trimmed) {
			// Check if this might be a style indicator by looking ahead
			// A style indicator must be followed by actual cell content
			if i+1 < len(parts) {
				nextPart := strings.TrimSpace(parts[i+1])
				// If next part has content and is NOT another style indicator, this is a style indicator
				if nextPart != "" && !p.isCellStyleIndicator(nextPart) {
					isStyle = true
				}
				// Also if next part is empty but there's more content after that,
				// it could be like: |l||content| which would mean empty literal cell
				if nextPart == "" && i+2 < len(parts) {
					isStyle = true
				}
			}
		}

		if isStyle {
			currentStyle = p.normalizeCellStyle(trimmed)
			continue
		}

		cell := ast.TableCell{
			ColSpan:    1,
			RowSpan:    1,
			Attributes: make(map[string]string),
		}

		// Apply current style if set
		if currentStyle != "" {
			cell.Style = currentStyle
			currentStyle = "" // Reset after use
		}

		// Check for leading alignment indicators
		// Note: The = character is kept as literal text (|= Header becomes "= Header")
		// because it's part of the cell content in basic AsciiDoc tables.
		if len(trimmed) > 0 {
			switch trimmed[0] {
			case '<':
				cell.HorizontalAlign = "left"
				trimmed = strings.TrimSpace(trimmed[1:])
			case '>':
				cell.HorizontalAlign = "right"
				trimmed = strings.TrimSpace(trimmed[1:])
			case '^':
				cell.HorizontalAlign = "center"
				trimmed = strings.TrimSpace(trimmed[1:])
			case '.':
				// Default alignment
				trimmed = strings.TrimSpace(trimmed[1:])
			}
		}

		// Check for repeat indicator (e.g., 3* for repeating cell)
		// Syntax: 3*value means "value" appears 3 times in the row
		if match := regexp.MustCompile(`^(\d+)\*(.*)`).FindStringSubmatch(trimmed); len(match) > 0 {
			if repeat, err := strconv.Atoi(match[1]); err == nil {
				cell.Repeat = repeat
				trimmed = match[2]
			}
		}

		// Check for colspan indicator (e.g., 2+ or +)
		if match := regexp.MustCompile(`^(\d+)?\+(.*)`).FindStringSubmatch(trimmed); len(match) > 0 && match[1] != "" {
			if colspan, err := strconv.Atoi(match[1]); err == nil {
				cell.ColSpan = colspan
				trimmed = match[2]
			}
		}

		// Parse inline content based on cell style
		cell.Text = trimmed
		if trimmed != "" {
			// For literal cells, don't parse inline content
			if cell.Style != "literal" {
				inlineParser := inline.NewParser(trimmed)
				inlineNodes := inlineParser.Parse()
				cell.InlineNodes = make([]interface{}, 0)
				for _, node := range inlineNodes {
					if node.Type != inline.NodeText {
						cell.InlineNodes = append(cell.InlineNodes, node)
					}
				}
			}
			// For literal cells, InlineNodes remain empty
		}

		cells = append(cells, cell)

		// Handle repeat - duplicate cells
		// The repeat syntax is: 3*|value which means "value" appears 3 times
		if cell.Repeat > 1 {
			repeatCount := cell.Repeat
			cell.Repeat = 0 // Reset so we don't repeat again
			for i := 1; i < repeatCount; i++ {
				cells = append(cells, cell)
			}
		}
	}

	return cells
}

// parseCSVRow parses a Comma Separated Values row.
func (p *TableParser) parseCSVRow(line string) []ast.TableCell {
	var cells []ast.TableCell

	// Simple CSV parsing - split by comma, respecting quotes
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range line {
		switch r {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = rune(0)
			} else {
				current.WriteRune(r)
			}
		case ',':
			if !inQuotes {
				cells = append(cells, p.createCell(current.String()))
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		cells = append(cells, p.createCell(current.String()))
	}

	return cells
}

// parseTSVRow parses a Tab Separated Values row.
func (p *TableParser) parseTSVRow(line string) []ast.TableCell {
	var cells []ast.TableCell

	parts := strings.Split(line, "\t")
	for _, part := range parts {
		cells = append(cells, p.createCell(part))
	}

	return cells
}

// parseDSVRow parses a Colon Separated Values row.
func (p *TableParser) parseDSVRow(line string) []ast.TableCell {
	var cells []ast.TableCell

	// Split by colon, respecting quotes
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range line {
		switch r {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = rune(0)
			} else {
				current.WriteRune(r)
			}
		case ':':
			if !inQuotes {
				cells = append(cells, p.createCell(current.String()))
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		cells = append(cells, p.createCell(current.String()))
	}

	return cells
}

// createCell creates a table cell from text content.
func (p *TableParser) createCell(text string) ast.TableCell {
	text = strings.TrimSpace(text)

	cell := ast.TableCell{
		Text:       text,
		ColSpan:    1,
		RowSpan:    1,
		Attributes: make(map[string]string),
	}

	if text != "" {
		inlineParser := inline.NewParser(text)
		inlineNodes := inlineParser.Parse()
		cell.InlineNodes = make([]interface{}, 0)
		for _, node := range inlineNodes {
			if node.Type != inline.NodeText {
				cell.InlineNodes = append(cell.InlineNodes, node)
			}
		}
	}

	return cell
}

// DetectTableFormat detects the table format from the format attribute.
func DetectTableFormat(formatAttr string) ast.TableFormat {
	switch strings.ToLower(formatAttr) {
	case "csv":
		return ast.FormatCSV
	case "tsv":
		return ast.FormatTSV
	case "dsv":
		return ast.FormatDSV
	default:
		return ast.FormatPSV
	}
}

// IsTableDelimiter checks if a line is a table delimiter.
func IsTableDelimiter(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "|===") ||
		strings.HasPrefix(line, "|+++") ||
		strings.HasPrefix(line, "|---")
}

// IsTableStart checks if a line starts a table block.
func IsTableStart(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "|===") ||
		(line == "|") ||
		strings.HasPrefix(line, "[") && IsTableLine(line)
}

// IsTableLine checks if a line looks like a table content line.
func IsTableLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "|") ||
		strings.HasPrefix(line, "[") ||
		strings.HasPrefix(line, "|=") ||
		strings.HasPrefix(line, "|_")
}

// IsCellContinuation checks if a line is a cell continuation marker.
func IsCellContinuation(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "|+") || line == "|+"
}

// handleCellContinuation handles the |+ continuation syntax for multi-line cells.
// In Asciidoctor, |+ at the start creates a new row where cells with content
// are added to the table. The |+ prefix is preserved as literal text in cells.
func (p *TableParser) handleCellContinuation(table *ast.Table, line string, lineno int) {
	// Trim the leading | but keep the + as literal text
	// |+ content becomes a row with first cell being "+ content"
	line = strings.TrimLeft(line, "| \t")

	// Parse this as a regular row - the + will be part of the cell content
	format := table.GetFormat()
	newRow := p.parseRow("|"+line, format, false)

	// Add the new row to the table
	table.Rows = append(table.Rows, newRow)
}

// isCellStyleIndicator checks if a single character is a cell style indicator.
// Cell style indicators are always lowercase single letters.
func (p *TableParser) isCellStyleIndicator(s string) bool {
	if len(s) != 1 {
		return false
	}
	// Must be lowercase letter (a-z)
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}
	switch s {
	case "l": // literal
		return true
	case "m": // monospace
		return true
	case "v": // verse
		return true
	case "a": // asciidoc
		return true
	case "e": // emphasis
		return true
	case "h": // header
		return true
	case "s": // strong
		return true
	case "d": // disk
		return true
	default:
		return false
	}
}

// normalizeCellStyle converts a single-letter style indicator to a normalized style name.
func (p *TableParser) normalizeCellStyle(s string) string {
	switch strings.ToLower(s) {
	case "l":
		return "literal"
	case "m":
		return "monospace"
	case "v":
		return "verse"
	case "a":
		return "asciidoc"
	case "e":
		return "emphasis"
	case "h":
		return "header"
	case "s":
		return "strong"
	case "d":
		return "disk"
	default:
		return ""
	}
}
