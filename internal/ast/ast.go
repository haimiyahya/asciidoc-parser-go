// Package ast defines the Abstract Syntax Tree for AsciiDoc documents.
package ast

// Node is the interface for all AST nodes.
type Node interface {
	// Type returns the node type.
	Type() NodeType

	// Position returns the source location.
	Position() Position
}

// DocumentHeader contains the document header information.
type DocumentHeader struct {
	// Title is the document title.
	Title string
	// Author is the document author.
	Author string
	// Version is the document version.
	Version string
	// Attributes are header-level attributes.
	Attributes map[string]string
	// RevisionInfo contains revision information.
	RevisionInfo *RevisionInfo
}

// RevisionInfo contains document revision information.
type RevisionInfo struct {
	Number    string
	Date      string
	Remark    string
}

// NodeType represents the type of an AST node.
type NodeType int

const (
	// TypeDocument is the root document node type.
	TypeDocument NodeType = iota

	// TypeSection is a section heading node type.
	TypeSection
	// TypeParagraph is a paragraph node type.
	TypeParagraph
	// TypeList is a list node type.
	TypeList
	// TypeListItem is an item in a list node type.
	TypeListItem
	// TypeLiteral is a literal/verbatim block node type.
	TypeLiteral
	// TypeBlock is a delimited block node type.
	TypeBlock
	// TypeTable is a table node type.
	TypeTable
	// TypeAttribute is an attribute entry node type.
	TypeAttribute
	// TypeInline is inline content type (text, formatting, links, etc.).
	TypeInline
)

// Position represents a location in the source.
type Position struct {
	File   string
	Line   int
}

// NodeDocument is the root document node.
type NodeDocument struct {
	// Kind is the node type.
	Kind NodeType
	// Header contains the document header (title, attributes, etc).
	Header *DocumentHeader
	// Blocks are the top-level blocks in the document.
	Blocks []Node
	// Attributes are document-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// NodeSection is a section heading.
type NodeSection struct {
	// Level is the section level (1-6, where 1 is document title).
	Level int
	// Title is the section title text.
	Title string
	// ID is the section ID (for anchors).
	ID string
	// Attributes are section-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// NodeParagraph is a paragraph.
type NodeParagraph struct {
	// Kind is the node type.
	Kind NodeType
	// Text is the paragraph text content.
	Text string
	// InlineNodes contains inline markup nodes found within the paragraph (future).
	// InlineNodes []Node // TODO: Implement inline parsing
	// Pos is the location in the source.
	Pos Position
}

// NodeList is a list (ordered, unordered, labeled, checklist).
type NodeList struct {
	// Kind is the node type.
	Kind NodeType
	// Items are the list items.
	Items []Node
	// Pos is the location in the source.
	Pos Position
}

// NodeListItem is an item in a list.
type NodeListItem struct {
	// Kind is the node type.
	Kind NodeType
	// Marker is the list marker character (-, *, ., etc).
	Marker string
	// Level is the nesting level (1-based for ordered lists).
	Level int
	// Ordinal is the ordinal number for ordered lists (1, 2, etc.).
	Ordinal int
	// Text is the item text content.
	Text string
	// Definition is the definition text for labeled lists (item.Definition).
	Definition string
	// Term is the term text for labeled lists (item.Term).
	Term string
	// InlineNodes contains inline markup nodes found within the list item (future).
	// InlineNodes []Node // TODO: Implement inline parsing
	// Pos is the location in the source.
	Pos Position
}

// NodeLiteral is a literal/verbatim block.
type NodeLiteral struct {
	// Kind is the node type.
	Kind NodeType
	// Lines are the literal block lines.
	Lines []string
	// Pos is the location in the source.
	Pos Position
}

// NodeBlock is a delimited block (example, quote, etc.).
type NodeBlock struct {
	// Kind is the node type.
	Kind NodeType
	// Delimiter is the block delimiter character.
	Delimiter string
	// Lines are the block lines.
	Lines []string
	// Style is the block style (optional, in brackets).
	Style string
	// Attributes are block-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// NodeTable is a table.
type Table struct {
	// Header is the table header.
	Header []string
	// Body is the table body (rows).
	Body [][]string
	// Pos is the location in the source.
	Pos Position
}

// Node interface methods for each type.

func (n *NodeDocument) Type() NodeType   { return n.Kind }
func (n *NodeDocument) Position() Position { return n.Pos }

func (n *NodeSection) Type() NodeType   { return TypeSection }
func (n *NodeSection) Position() Position { return n.Pos }

func (n *NodeParagraph) Type() NodeType   { return TypeParagraph }
func (n *NodeParagraph) Position() Position { return n.Pos }

func (n *NodeList) Type() NodeType   { return TypeList }
func (n *NodeList) Position() Position { return n.Pos }

func (n *NodeListItem) Type() NodeType   { return TypeListItem }
func (n *NodeListItem) Position() Position { return n.Pos }

func (n *NodeLiteral) Type() NodeType   { return TypeLiteral }
func (n *NodeLiteral) Position() Position { return n.Pos }

func (n *NodeBlock) Type() NodeType   { return TypeBlock }
func (n *NodeBlock) Position() Position { return n.Pos }

func (n *Table) Type() NodeType   { return TypeTable }
func (n *Table) Position() Position { return n.Pos }