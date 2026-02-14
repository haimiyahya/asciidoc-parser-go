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
func (p *Parser) Parse() (*ast.Document, error) {
	doc := &ast.Document{
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
			// Flush any pending paragraph first
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}

			section := p.createSection(classification.Section, lineno)
			if section != nil {
				doc.Blocks = append(doc.Blocks, section)
			}
			p.reader.Advance()
			continue
		}

		// Handle attribute entries
		if classification.Type == reader.BlockAttribute && classification.Attribute != nil {
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

			listItem := p.createListItem(classification.List, lineno)
			if listItem != nil {
				doc.Blocks = append(doc.Blocks, listItem)
			}
			p.reader.Advance()
			continue
		}

		// Handle blank lines - they terminate paragraphs
		if classification.Type == reader.BlockBlank {
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
			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					doc.Blocks = append(doc.Blocks, para)
				}
				paragraphLines = nil
			}
			// For now, skip macros
			p.reader.Advance()
			continue
		}

		// Handle admonitions
		if classification.Type == reader.BlockAdmonition {
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

	return doc, nil
}

// createParagraph creates a paragraph node from accumulated lines.
func (p *Parser) createParagraph(lines []string, lineno int) ast.Node {
	if len(lines) == 0 {
		return nil
	}

	// Join lines with spaces
	content := strings.Join(lines, " ")

	return &ast.Paragraph{
		Text: strings.TrimSpace(content),
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
		return &ast.Title{
			Text: info.Title,
			Pos:  ast.Position{Line: lineno},
		}
	}

	return &ast.Section{
		Level: info.Level,
		Title: info.Title,
		ID:    info.ID,
		Pos:   ast.Position{Line: lineno},
	}
}

// createListItem creates a list item node from list info.
func (p *Parser) createListItem(info *reader.ListInfo, lineno int) ast.Node {
	if info == nil {
		return nil
	}

	return &ast.ListItem{
		NodeType: ast.NodeType(info.Type),
		Marker:   info.Marker,
		Level:    info.Level,
		Ordinal:  info.Ordinal,
		Text:     info.Text,
		Pos:      ast.Position{Line: lineno},
	}
}

// createDelimitedBlock creates a delimited block node.
func (p *Parser) createDelimitedBlock(blockType reader.BlockType, lines []string, lineno int) ast.Node {
	content := strings.Join(lines, "\n")

	switch blockType {
	case reader.BlockLiteral:
		return &ast.Literal{
			Text: content,
			Pos:  ast.Position{Line: lineno},
		}
	case reader.BlockVerbatim:
		return &ast.Literal{
			Text: content,
			Pos:  ast.Position{Line: lineno},
		}
	case reader.BlockExample:
		return &ast.Block{
			NodeType: ast.NodeBlock,
			Pos:  ast.Position{Line: lineno},
		}
	case reader.BlockQuote:
		return &ast.Block{
			NodeType: ast.NodeBlock,
			Pos:  ast.Position{Line: lineno},
		}
	default:
		return &ast.Block{
			NodeType: ast.NodeBlock,
			Pos:  ast.Position{Line: lineno},
		}
	}
}

// Advance is a helper that consumes the next line.
func (p *Parser) Advance() bool {
	return p.reader.Advance()
}
