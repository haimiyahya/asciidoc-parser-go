// Package ast defines the Abstract Syntax Tree for AsciiDoc documents.
package ast

// Node is the interface for all AST nodes.
type Node interface {
	// Type returns the node type.
	Type() NodeType

	// Position returns the source position.
	Position() Position
}

// NodeType represents the type of an AST node.
type NodeType int

const (
	// NodeDocument is the root document node.
	NodeDocument NodeType = iota

	// NodeSection is a section heading.
	NodeSection

	// NodeParagraph is a paragraph.
	NodeParagraph

	// NodeList is a list (ordered, unordered, labeled, checklist).
	NodeList

	// NodeListItem is an item in a list.
	NodeListItem

	// NodeLiteral is a literal block (verbatim, source, etc.).
	NodeLiteral

	// NodeBlock is a delimited block (example, quote, etc.).
	NodeBlock

	// NodeTable is a table.
	NodeTable

	// NodeAttribute is an attribute entry.
	NodeAttribute

	// NodeInline is inline content (text, formatting, links, etc.).
	NodeInline
)

// Position represents a location in the source.
type Position struct {
	File   string
	Line    int
	Column   int
}

// Document is the root node of an AsciiDoc document.
type Document struct {
	// Header contains the document header (title, attributes, etc.).
	Header *DocumentHeader

	// Blocks are the top-level blocks in the document.
	Blocks []Node

	// Attributes are document-level attributes.
	Attributes map[string]string

	// Pos is the location in the source.
	Pos Position
}

func (d *Document) Type() NodeType { return NodeDocument }
func (d *Document) Position() Position { return d.Pos }

// DocumentHeader contains the document header.
type DocumentHeader struct {
	// Title is the document title.
	Title string

	// Authors are the document authors.
	Authors []string

	// Revision information (version, date, etc.).
	Revision *RevisionInfo
}

// RevisionInfo contains document revision information.
type RevisionInfo struct {
	Version string
	Date    string
	Remark  string
}

// Section is a section heading (== Title, === Title, etc.).
type Section struct {
	// Level is the section level (1-6, where 1 is highest).
	Level int

	// Title is the section title text.
	Title string

	// ID is the optional section ID.
	ID string

	// Pos is the location in the source.
	Pos Position
}

func (s *Section) Type() NodeType { return NodeSection }
func (s *Section) Position() Position { return s.Pos }

// Title is a document title (= Title).
type Title struct {
	// Text is the title text.
	Text string

	// ID is the optional title ID.
	ID string

	// Pos is the location in the source.
	Pos Position
}

func (t *Title) Type() NodeType { return NodeSection }
func (t *Title) Position() Position { return t.Pos }

// Paragraph is a paragraph block.
type Paragraph struct {
	// Text is the paragraph text content.
	Text string

	// Pos is the location in the source.
	Pos Position
}

func (p *Paragraph) Type() NodeType { return NodeParagraph }
func (p *Paragraph) Position() Position { return p.Pos }

// ListItem is an item in a list.
type ListItem struct {
	// NodeType is the list item type (unordered, ordered, labeled, etc.).
	NodeType NodeType

	// Marker is the list marker character (-, *, ., etc.).
	Marker string

	// Level is the nesting level (1-based).
	Level int

	// Ordinal is the ordinal number for ordered lists.
	Ordinal int

	// Text is the item text content.
	Text string

	// Pos is the location in the source.
	Pos Position
}

func (l *ListItem) Type() NodeType { return l.NodeType }
func (l *ListItem) Position() Position { return l.Pos }

// Literal is a literal or verbatim block.
type Literal struct {
	// Text is the literal text content.
	Text string

	// Pos is the location in the source.
	Pos Position
}

func (l *Literal) Type() NodeType { return NodeLiteral }
func (l *Literal) Position() Position { return l.Pos }

// Block is a delimited block (example, quote, sidebar, etc.).
type Block struct {
	// NodeType is the block node type.
	NodeType NodeType

	// Style is the block style (example, quote, sidebar, etc.).
	Style string

	// Content is the block content.
	Content string

	// Pos is the location in the source.
	Pos Position
}

func (b *Block) Type() NodeType { return b.NodeType }
func (b *Block) Position() Position { return b.Pos }
