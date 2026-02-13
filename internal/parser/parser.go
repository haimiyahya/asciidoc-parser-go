// Package parser provides AsciiDoc parsing functionality.
package parser

import (
	"io"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
)

// Parser parses AsciiDoc source into an AST.
type Parser struct {
	// reader is the source reader.
	reader *reader.Reader

	// options configures parser behavior.
	options []ParserOption
}

// ParserOption configures a parser.
type ParserOption func(*Parser)

// NewParser creates a new Parser.
func NewParser(r *reader.Reader, opts ...ParserOption) *Parser {
	return &Parser{
		reader:  r,
		options: opts,
	}
}

// NewParserFromReader creates a parser from an io.Reader.
func NewParserFromReader(r io.Reader, opts ...ParserOption) (*Parser, error) {
	rd, err := reader.NewReaderFromReader(r)
	if err != nil {
		return nil, err
	}
	return NewParser(rd, opts...)
}

// NewParserFromString creates a parser from a string.
func NewParserFromString(source string, opts ...ParserOption) (*Parser, error) {
	rd, err := reader.NewReader(source)
	if err != nil {
		return nil, err
	}
	return NewParser(rd, opts...)
}

// Parse parses the AsciiDoc source into a document AST.
func (p *Parser) Parse() (*ast.Document, error) {
	doc := &ast.Document{
		Attributes: make(map[string]string),
	}

	// TODO: Implement full parsing
	// For now, just read all lines
	lines := p.reader.ReadLines()

	// TODO: Convert lines to AST nodes

	return doc, nil
}
