// Package parser provides table parsing functionality for AsciiDoc.
package parser

import (
	"fmt"
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
		Rows:           make([]ast.TableRow, 0),
		Attributes:     make(map[string]string),
		Pos:            ast.Position{Line: lineno},
		HeaderRowIndex: -1, // Initialize to -1 to indicate no header
		FooterRowIndex: -1, // Initialize to -1 to indicate no footer
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

	// Check for header option in attributes
	hasHeaderOption := false
	if options, ok := table.Attributes["options"]; ok {
		// Options can be comma-separated like "header,footer" or just "header"
		optParts := strings.Split(options, ",")
		for _, opt := range optParts {
			if strings.TrimSpace(opt) == "header" {
				hasHeaderOption = true
				break
			}
		}
	}

	// Check for caption (line starting with single .)
	// Caption comes after attributes but before table content
	if lineIdx < len(lines) {
		captionLine := strings.TrimSpace(lines[lineIdx])
		if strings.HasPrefix(captionLine, ".") && !strings.HasPrefix(captionLine, "..") && !strings.HasPrefix(captionLine, ".#") {
			// This is a caption line (single . not followed by . or #)
			// Remove the leading . and trim
			table.Caption = strings.TrimLeft(captionLine, ".")
			table.Caption = strings.TrimSpace(table.Caption)
			lineIdx++
		}
	}

	// Parse rows with support for multi-line cell content
	// In AsciiDoc, lines that don't start with | are continuation of the previous cell
	firstRow := true
	var pendingRowIndex int = -1
	var pendingCellIndex int = -1
	var pendingContent []string

	for i := lineIdx; i < len(lines); i++ {
		originalLine := lines[i]
		line := strings.TrimSpace(originalLine)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for row separator
		if strings.HasPrefix(line, "|===") || strings.HasPrefix(line, "|+++") || strings.HasPrefix(line, "|---") {
			// Flush any pending content before separator
			if pendingRowIndex >= 0 && pendingCellIndex >= 0 && len(pendingContent) > 0 {
				p.flushPendingContent(&table.Rows[pendingRowIndex], pendingCellIndex, pendingContent)
				pendingContent = nil
				pendingCellIndex = -1
			}
			pendingRowIndex = -1
			continue
		}

		// Check if this is a row line (starts with |) or continuation (doesn't start with |)
		if strings.HasPrefix(line, "|") || strings.HasPrefix(line, "|_") {
			// This is a new row - flush any pending content from previous row
			if pendingRowIndex >= 0 && pendingCellIndex >= 0 && len(pendingContent) > 0 {
				p.flushPendingContent(&table.Rows[pendingRowIndex], pendingCellIndex, pendingContent)
				pendingContent = nil
				pendingCellIndex = -1
			}

			// Determine row kind for this specific row
			rowKind := ast.TableRowBody

			// Check for footer row specification
			if strings.HasPrefix(line, "|_") {
				rowKind = ast.TableRowFooter
				line = strings.TrimPrefix(line, "|_")
				line = strings.TrimLeft(line, "_")
			}

			// Mark first row as header if option is set
			isHeaderRow := hasHeaderOption && firstRow

			// Parse the row
			format := table.GetFormat()
			row := p.parseRow(line, format, isHeaderRow)

			// Set row kind (footer takes precedence, otherwise use parsed header status or default to body)
			if rowKind == ast.TableRowFooter {
				row.Kind = rowKind
			} else if !isHeaderRow {
				row.Kind = rowKind
			}

			// Track header row index
			if row.Kind == ast.TableRowHeader && table.HeaderRowIndex < 0 {
				table.HeaderRowIndex = len(table.Rows)
			}

			// Track footer row index
			if row.Kind == ast.TableRowFooter {
				table.FooterRowIndex = len(table.Rows)
			}

			table.Rows = append(table.Rows, row)
			pendingRowIndex = len(table.Rows) - 1

			// If the last cell in this row is empty, it can accept continuation content
			if len(row.Cells) > 0 {
				lastCell := &row.Cells[len(row.Cells)-1]
				if lastCell.Text == "" {
					pendingCellIndex = len(row.Cells) - 1
				}
			}

			// Clear first row flag after processing
			firstRow = false
		} else {
			// This is a continuation line - add to pending content
			// Preserve the original line (without trimming leading whitespace for code)
			if pendingRowIndex >= 0 && pendingCellIndex >= 0 {
				pendingContent = append(pendingContent, originalLine)
			} else {
				// No pending cell to accept this content - create a new single-cell row
				format := table.GetFormat()
				row := p.parseRow("|", format, false)
				if len(row.Cells) > 0 {
					// Set the content directly
					row.Cells[0].Text = originalLine
					row.Cells[0].Blocks = p.parseListContent([]string{originalLine})
				}
				table.Rows = append(table.Rows, row)
			}
		}
	}

	// Flush any remaining pending content at end of table
	if pendingRowIndex >= 0 && pendingCellIndex >= 0 && len(pendingContent) > 0 {
		p.flushPendingContent(&table.Rows[pendingRowIndex], pendingCellIndex, pendingContent)
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
// Example: [cols="1,2,3",frame="all",grid="rows",%autowidth,options="header"]
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

		// Handle positional attributes (starting with %)
		if strings.HasPrefix(part, "%") {
			attrName := strings.TrimPrefix(part, "%")
			// For positional attributes like %autowidth, convert to key="true" format
			attrs[attrName] = "true"
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
// Example: "1,2,3" or "<,>,^" or "h,l,l" or "3*" (3 equal columns) or "2*,1*" (2 equal + 1)
func (p *TableParser) parseColumnSpecs(spec string) []ast.TableColumnSpec {
	// Remove any surrounding quotes
	spec = strings.Trim(spec, "\"'")

	// Split by comma
	parts := strings.Split(spec, ",")
	allCols := make([]ast.TableColumnSpec, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for repeat operator (e.g., "3*" means 3 columns, "2*" means 2 columns)
		repeatCount := 1
		specPart := part
		if strings.HasSuffix(part, "*") {
			// Extract the number before *
			starIndex := strings.LastIndex(part, "*")
			numStr := part[:starIndex]
			if numStr != "" {
				// Parse the repeat count
				var count int
				fmt.Sscanf(numStr, "%d", &count)
				if count > 0 {
					repeatCount = count
				}
			}
			// For "N*" or just "*", create auto-width columns
			specPart = "" // Empty spec means auto-width with default alignment
		}

		// Create the column spec
		col := ast.TableColumnSpec{
			AutoWidth: true,
		}

		// Parse the spec characters (if any)
		for _, r := range specPart {
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

		// Add the column spec repeatCount times
		for i := 0; i < repeatCount; i++ {
			allCols = append(allCols, col)
		}
	}

	return allCols
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

	row := ast.TableRow{
		Cells:      cells,
		Attributes: make(map[string]string),
	}

	// Set row kind based on isHeaderRow parameter
	if isHeaderRow {
		row.Kind = ast.TableRowHeader
	}

	return row
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

	// Special case: if line is empty after trimming |, create one empty cell
	// This handles the case of "|" (single pipe) which should create an empty cell
	if line == "" {
		cell := ast.TableCell{
			Text:       "",
			ColSpan:    1,
			RowSpan:    1,
			Attributes: make(map[string]string),
			InlineNodes: make([]interface{}, 0),
		}
		return []ast.TableCell{cell}
	}

	// Split by |, but handle || as empty cell with colspan
	parts := strings.Split(line, "|")

	// Remove trailing empty string if line ends with |
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	// If all parts were removed, we still need at least one cell
	if len(parts) == 0 {
		cell := ast.TableCell{
			Text:       "",
			ColSpan:    1,
			RowSpan:    1,
			Attributes: make(map[string]string),
			InlineNodes: make([]interface{}, 0),
		}
		return []ast.TableCell{cell}
	}

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)

		cell := ast.TableCell{
			ColSpan:    1,
			RowSpan:    1,
			Attributes: make(map[string]string),
		}

		// Check for colspan indicator (e.g., 2+ or +)
		if match := regexp.MustCompile(`^(\d+)?\+(.*)`).FindStringSubmatch(trimmed); len(match) > 0 && match[1] != "" {
			if colspan, err := strconv.Atoi(match[1]); err == nil {
				cell.ColSpan = colspan
				trimmed = match[2]
			}
		}

		// Parse inline content
		cell.Text = trimmed
		if trimmed != "" {
			inlineParser := inline.NewParser(trimmed)
			inlineNodes := inlineParser.Parse()
			cell.InlineNodes = make([]interface{}, 0)
			for _, node := range inlineNodes {
				if node.Type != inline.NodeText {
					cell.InlineNodes = append(cell.InlineNodes, node)
				}
			}
		} else {
			cell.InlineNodes = make([]interface{}, 0)
		}

		cells = append(cells, cell)
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

// flushPendingContent flushes accumulated content lines to a table cell.
func (p *TableParser) flushPendingContent(row *ast.TableRow, cellIndex int, content []string) {
	if row == nil || cellIndex < 0 || cellIndex >= len(row.Cells) {
		return
	}

	cell := &row.Cells[cellIndex]
	fullContent := strings.Join(content, "\n")

	// Set the text content
	cell.Text = fullContent

	// Parse and set any block content (lists, etc.)
	cell.Blocks = p.parseListContent(content)
}

// parseListContent checks if content lines contain a list and parses it.
// Returns a slice of block nodes if list(s) are found, nil otherwise.
func (p *TableParser) parseListContent(lines []string) []ast.Node {
	if len(lines) == 0 {
		return nil
	}

	// Check if content starts with list markers
	hasListMarkers := false
	listMarkerPatterns := []string{
		"* ",   // Unordered
		"- ",   // Unordered alt
		"** ",  // Nested
		"*** ", // Triple nested
		". ",   // Ordered
		".. ",  // Nested ordered
	}

	// Check first non-empty line
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}
		for _, pattern := range listMarkerPatterns {
			if strings.HasPrefix(trimmed, pattern) {
				hasListMarkers = true
				break
			}
		}
		if hasListMarkers {
			break
		}
	}

	if !hasListMarkers {
		return nil
	}

	// Parse the content as a list
	return p.parseSimpleList(lines)
}

// parseSimpleList parses simple list content from lines.
// Handles unordered lists with nested items.
func (p *TableParser) parseSimpleList(lines []string) []ast.Node {
	var lists []ast.Node
	var currentList *ast.NodeList
	var lastItem *ast.NodeListItem
	var lastLevel int

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}

		// Check for list item marker
		level := 0
		restOfLine := trimmed

		// Count nesting level by * prefix (only * followed by space or end of line)
		// We need to be careful not to consume ** from bold markup
		for strings.HasPrefix(restOfLine, "*") && (len(restOfLine) == 1 || restOfLine[1] == ' ' || restOfLine[1] == '\t') {
			level++
			restOfLine = strings.TrimPrefix(restOfLine, "*")
			restOfLine = strings.TrimLeft(restOfLine, " \t")
		}

		// Check for - marker as alternative
		if level == 0 && strings.HasPrefix(trimmed, "- ") {
			level = 1
			restOfLine = strings.TrimLeft(trimmed[1:], " \t")
		}

		// Check for . ordered marker
		if level == 0 && strings.HasPrefix(trimmed, ". ") {
			level = 1
			restOfLine = strings.TrimLeft(trimmed[1:], " \t")
		}

		if level > 0 && restOfLine != trimmed {
			// This is a list item
			if currentList == nil {
				currentList = &ast.NodeList{
					Kind: ast.TypeList,
					Pos:  ast.Position{Line: 1},
				}
				lists = append(lists, currentList)
			}

			// Parse inline content for the item
			item := &ast.NodeListItem{
				Kind:   ast.TypeListItem,
				Marker: "*",
				Level:  level,
				Text:   restOfLine,
				Pos:    ast.Position{Line: 1},
			}

			// Parse inline formatting
			if restOfLine != "" {
				inlineParser := inline.NewParser(restOfLine)
				inlineNodes := inlineParser.Parse()
				item.InlineNodes = make([]interface{}, 0)
				for _, node := range inlineNodes {
					if node.Type != inline.NodeText {
						item.InlineNodes = append(item.InlineNodes, node)
					}
				}
			}

			// Handle nesting
			if level > lastLevel && lastItem != nil {
				// Create nested list
				nestedList := &ast.NodeList{
					Kind: ast.TypeList,
					Pos:  ast.Position{Line: 1},
				}
				lastItem.NestedList = nestedList
				currentList = nestedList
			} else if level < lastLevel {
				// Need to go back up - for simplicity, just add to main list
				// A proper implementation would track the list hierarchy
				currentList = lists[len(lists)-1].(*ast.NodeList)
			}

			currentList.Items = append(currentList.Items, item)
			lastItem = item
			lastLevel = level
		}
	}

	return lists
}
