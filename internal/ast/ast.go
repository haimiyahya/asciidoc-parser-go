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

	// Position is the location in the source.
	Position Position
}

func (d *Document) Type() NodeType { return NodeDocument }
func (d *Document) Position() Position { return d.Position }

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
