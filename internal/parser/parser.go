// Package parser provides AsciiDoc parsing functionality.
package parser

import (
	"io"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
)

// Parser parses AsciiDoc source into an AST.
type Parser struct {
	// reader is the source reader.
	reader *reader.Reader

	// classifier classifies lines into block types.
	classifier *reader.LineClassifier

	// options configures parser behavior.
	options []ParserOption

	// List tracking state
	currentList      *ast.NodeList
	currentListBlockType reader.BlockType
	currentListLevel int
}

// ParserOption configures a parser.
type ParserOption func(*Parser)

// NewParser creates a new Parser.
func NewParser(r *reader.Reader, opts ...ParserOption) *Parser {
	return &Parser{
		reader:    r,
		classifier: reader.NewLineClassifier(),
		options:   opts,
	}
}

// NewParserFromReader creates a parser from an io.Reader.
func NewParserFromReader(r io.Reader, opts ...ParserOption) (*Parser, error) {
	rd, err := reader.NewReaderFromReader(r)
	if err != nil {
		return nil, err
	}
	return NewParser(rd, opts...), nil
}

// NewParserFromString creates a parser from a string.
func NewParserFromString(source string, opts ...ParserOption) (*Parser, error) {
	rd, err := reader.NewReader(source)
	if err != nil {
		return nil, err
	}
	return NewParser(rd, opts...), nil
}

// Parse parses the AsciiDoc source into a document AST.
func (p *Parser) Parse() (*ast.NodeDocument, error) {
	doc := &ast.NodeDocument{
		Attributes: make(map[string]string),
		Blocks:    make([]ast.Node, 0),
	}

	// Track current paragraph lines being accumulated
	var paragraphLines []string
	var paragraphLineno int

	// Track whether we're in a delimited block
	var inDelimitedBlock bool
	var delimitedBlockType reader.BlockType
	var delimitedBlockLines []string
	var delimitedBlockLineno int

	for p.reader.HasMoreLines() {
		line := p.reader.PeekLine()
		lineno := p.reader.GetLineno()

		// Classify the line
		classification := p.classifier.ClassifyLine(line)

		// Handle delimited blocks
		if classification.Type.IsDelimitedBlock() {
			if !inDelimitedBlock {
				// Starting a delimited block
				inDelimitedBlock = true
				delimitedBlockType = classification.Type
				delimitedBlockLines = []string{}
				delimitedBlockLineno = lineno
				p.reader.Advance()
				continue
			} else {
				// Check if this is the closing delimiter
				// Same type means we're closing
				if classification.Type == delimitedBlockType {
					// Close the delimited block
					p.reader.Advance()
					block := p.createDelimitedBlock(delimitedBlockType, delimitedBlockLines, delimitedBlockLineno)
					if block != nil {
						doc.Blocks = append(doc.Blocks, block)
					}
					inDelimitedBlock = false
					delimitedBlockLines = nil
					continue
				} else {
					// Different delimiter inside block - treat as content
					delimitedBlockLines = append(delimitedBlockLines, line)
					p.reader.Advance()
					continue
				}
			}
		}

		// If we're in a delimited block, accumulate lines
		if inDelimitedBlock {
			delimitedBlockLines = append(delimitedBlockLines, line)
			p.reader.Advance()
			continue
		}

		// Handle section headers
		if classification.Type == reader.BlockSection && classification.Section != nil {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph first
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}

			// For level 0 (document title), set the document header
			if classification.Section.Level == 0 && doc.Header == nil {
				doc.Header = &ast.DocumentHeader{Title: classification.Section.Title}
			}

			section := p.createSection(classification.Section, lineno)
			if section != nil {
				// Level 0 is the document title - don't add as a block
				// It's already been handled by setting doc.Header
				if sec, ok := section.(*ast.NodeSection); ok && sec.Level > 0 {
					doc.Blocks = append(doc.Blocks, section)
				}
			}
			p.reader.Advance()
			continue
		}

		// Handle attribute entries
		if classification.Type == reader.BlockAttribute && classification.Attribute != nil {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}

			attr := classification.Attribute
			if attr.IsSet {
				doc.Attributes[attr.Name] = attr.Value
			} else {
				delete(doc.Attributes, attr.Name)
			}

			// Check for document title attribute
			if attr.Name == "title" && doc.Header == nil {
				doc.Header = &ast.DocumentHeader{Title: attr.Value}
			}

			p.reader.Advance()
			continue
		}

		// Handle list items
		if classification.Type.IsListItem() {
			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}

			// Check if we need to close the current list
			if p.currentList == nil {
				// No current list - need to check if this item starts a new list
				if classification.List != nil {
					p.startNewList(classification, lineno, doc)
				}
			} else {
				// We have a current list - check if this item belongs to it
				itemInfo := classification.List
				sameType := (itemInfo.Type == p.currentListBlockType)
				sameLevel := (itemInfo.Level == p.currentListLevel)

				if sameType && sameLevel {
					// Same list - add item to it
					p.addListItemToList(classification, lineno)
				} else if itemInfo.Level > p.currentListLevel {
					// Nested list - add as child of current list item
					p.addNestedList(classification, lineno, doc)
				} else {
					// Different list - close current and start new
					p.closeCurrentList(doc)
					if classification.List != nil {
						p.startNewList(classification, lineno, doc)
					}
				}
			}

			p.reader.Advance()
			continue
		}

		// Handle blank lines - they terminate paragraphs and lists
		if classification.Type == reader.BlockBlank {
			// Close any open list
			p.closeCurrentList(doc)

			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}
			p.reader.Advance()
			continue
		}

		// Handle comments (skip them)
		if classification.Type == reader.BlockComment {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}
			p.reader.Advance()
			continue
		}

		// Handle horizontal rules, page breaks
		if classification.Type == reader.BlockHorizontalRule || classification.Type == reader.BlockPageBreak {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}
			p.reader.Advance()
			continue
		}

		// Handle block macros
		if classification.Type == reader.BlockMacro {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}

			// Create macro node
			macro := p.createMacro(classification.Macro, lineno)
			if macro != nil {
				doc.Blocks = append(doc.Blocks, macro)
			}

			p.reader.Advance()
			continue
		}

		// Handle admonitions
		if classification.Type == reader.BlockAdmonition {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}

			// Create admonition node
			admonition := p.createAdmonition(classification.Admonition, lineno)
			if admonition != nil {
				doc.Blocks = append(doc.Blocks, admonition)
			}

			p.reader.Advance()
			continue
		}

		// Default: accumulate paragraph lines
		if len(paragraphLines) == 0 {
			paragraphLineno = lineno
		}
		paragraphLines = append(paragraphLines, line)
		p.reader.Advance()
	}

	// Flush any remaining paragraph
	if len(paragraphLines) > 0 {
		para := p.createParagraph(paragraphLines, paragraphLineno)
		if para != nil {
			doc.Blocks = append(doc.Blocks, para)
		}
	}

	// Close any remaining open list
	p.closeCurrentList(doc)

	return doc, nil
}

// createParagraph creates a paragraph node from accumulated lines.
func (p *Parser) createParagraph(lines []string, lineno int) ast.Node {
	if len(lines) == 0 {
		return nil
	}

	// Join lines with spaces
	content := strings.Join(lines, " ")

	return &ast.NodeParagraph{
		Text: content,
		Pos:  ast.Position{Line: lineno},
	}
}

// createSection creates a section node from section info.
func (p *Parser) createSection(info *reader.SectionInfo, lineno int) ast.Node {
	if info == nil {
		return nil
	}

	// For level 0 (document title), set the document header
	if info.Level == 0 {
		return &ast.NodeSection{
			Level: 0,
			Title: info.Title,
			Pos:   ast.Position{Line: lineno},
		}
	}

	return &ast.NodeSection{
		Level: info.Level,
		Title: info.Title,
		Pos:   ast.Position{Line: lineno},
	}
}

// createListItem creates a list item node from list info.
func (p *Parser) createListItem(info *reader.ListInfo, lineno int) ast.Node {
	if info == nil {
		return nil
	}

	// For labeled lists, Text contains the term (before ::), and Definition contains the definition
	text := info.Text
	if info.Type == reader.BlockListLabeled && info.Term != "" {
		text = info.Term
	}

	return &ast.NodeListItem{
		Kind:       ast.TypeListItem,
		Marker:      info.Marker,
		Level:       info.Level,
		Ordinal:     info.Ordinal,
		Text:         text,
		Term:          info.Term,
		Definition:    info.Text, // For labeled lists, Text contains the definition
		Pos:          ast.Position{Line: lineno},
	}
}

// createDelimitedBlock creates a delimited block node.
// createAdmonition creates an admonition node.
func (p *Parser) createAdmonition(admonition *reader.AdmonitionInfo, lineno int) ast.Node {
	if admonition == nil {
		return nil
	}

	return &ast.AdmonitionNode{
		Kind: admonition.Kind,
		Text: admonition.Text,
		Pos:  ast.Position{Line: lineno},
	}
}

// createMacro creates a block macro node.
func (p *Parser) createMacro(macro *reader.MacroInfo, lineno int) ast.Node {
	if macro == nil {
		return nil
	}

	return &ast.MacroNode{
		Kind:       ast.TypeMacro,
		Target:      macro.Target,
		Path:        macro.Path,
		Attributes:  macro.Attributes,
		Pos:         ast.Position{Line: lineno},
	}
}

// createDelimitedBlock creates a delimited block node.
func (p *Parser) createDelimitedBlock(blockType reader.BlockType, lines []string, lineno int) ast.Node {
	content := strings.Join(lines, "\n")

	switch blockType {
	case reader.BlockLiteral:
		return &ast.NodeLiteral{
			Lines: strings.Split(content, "\n"),
			Pos:   ast.Position{Line: lineno},
		}
	case reader.BlockVerbatim:
		return &ast.NodeLiteral{
			Lines: strings.Split(content, "\n"),
			Pos:   ast.Position{Line: lineno},
		}
	case reader.BlockExample:
		return &ast.NodeBlock{
			Delimiter: "=",
			Lines:    strings.Split(content, "\n"),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockQuote:
		return &ast.NodeBlock{
			Delimiter: "_",
			Lines:    strings.Split(content, "\n"),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockTable:
		return p.createTable(lines, lineno)
	default:
		return &ast.NodeBlock{
			Lines: strings.Split(content, "\n"),
			Pos:   ast.Position{Line: lineno},
		}
	}
}

// createTable parses table content into a Table node.
// Tables use | as column separator.
// First row after |=== is typically the header.
func (p *Parser) createTable(lines []string, lineno int) ast.Node {
	if len(lines) == 0 {
		return nil
	}

	table := &ast.Table{
		Pos: ast.Position{Line: lineno},
	}

	// Process each row
	for i, line := range lines {
		// Skip empty lines in tables
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split row by | separator
		cells := p.splitTableRow(line)
		if i == 0 {
			// First non-empty row is header
			table.Header = cells
		} else {
			// Body rows
			table.Body = append(table.Body, cells)
		}
	}

	return table
}

// splitTableRow splits a table row by | separator.
// Handles | as column separator, preserving cell content.
func (p *Parser) splitTableRow(row string) []string {
	// Split by | character
	cells := strings.Split(row, "|")

	// Trim whitespace from each cell
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}

	// Remove empty first/last cells if caused by leading/trailing |
	if len(cells) > 0 && cells[0] == "" {
		cells = cells[1:]
	}
	if len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}

	return cells
}

// closeCurrentList closes the current open list if any.
func (p *Parser) closeCurrentList(doc *ast.NodeDocument) {
	if p.currentList != nil {
		doc.Blocks = append(doc.Blocks, p.currentList)
		p.currentList = nil
		p.currentListBlockType = 0
		p.currentListLevel = 0
	}
}

// startNewList starts a new list with the given item as its first element.
func (p *Parser) startNewList(classification *reader.Classification, lineno int, doc *ast.NodeDocument) {
	info := classification.List
	if info == nil {
		return
	}

	// Create a new list with this item as its first element
	listItem := p.createListItem(info, lineno)
	if listItem == nil {
		return
	}

	// Create the list node - all lists use TypeList as the Kind
	p.currentList = &ast.NodeList{
		Kind:  ast.TypeList,
		Items: []ast.Node{listItem},
		Pos:   ast.Position{Line: lineno},
	}
	p.currentListBlockType = info.Type
	p.currentListLevel = info.Level

	// Note: Don't add to doc.Blocks yet - wait until list is closed
}

// addListItemToList adds a list item to the current list.
func (p *Parser) addListItemToList(classification *reader.Classification, lineno int) {
	info := classification.List
	if info == nil || p.currentList == nil {
		return
	}

	listItem := p.createListItem(info, lineno)
	if listItem != nil {
		p.currentList.Items = append(p.currentList.Items, listItem)
	}
}

// addNestedList adds a nested list as a child of the current list's last item.
func (p *Parser) addNestedList(classification *reader.Classification, lineno int, doc *ast.NodeDocument) {
	info := classification.List
	if info == nil || p.currentList == nil {
		return
	}

	// Get the last item in the current list
	if len(p.currentList.Items) == 0 {
		return
	}

	lastItemIdx := len(p.currentList.Items) - 1
	lastItem := p.currentList.Items[lastItemIdx]

	// Check if last item already has a nested list of the same type
	// If so, add to that nested list instead of creating a new one
	item, ok := lastItem.(*ast.NodeListItem)
	if !ok {
		return
	}

	if item.NestedList != nil {
		// Add new item to existing nested list
		listItem := p.createListItem(info, lineno)
		if listItem != nil {
			item.NestedList.Items = append(item.NestedList.Items, listItem)
		}
		return
	}

	// Create a new nested list
	nestedList := &ast.NodeList{
		Kind:  ast.TypeList,
		Items: []ast.Node{},
		Pos:   ast.Position{Line: lineno},
	}

	// Add the new list item to the nested list
	listItem := p.createListItem(info, lineno)
	if listItem != nil {
		nestedList.Items = append(nestedList.Items, listItem)
	}

	// Attach the nested list to the parent list item
	item.NestedList = nestedList
}

// Advance is a helper that consumes the next line.
func (p *Parser) Advance() bool {
	return p.reader.Advance()
}
