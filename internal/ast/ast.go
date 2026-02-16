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

	// TypeAdmonition is an admonition node type.
	TypeAdmonition
	// TypeMacro is a block macro node type.
	TypeMacro
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
	// TypeStyledBlock is a styled block node type (pass::[], sidebar::[], etc.)
	TypeStyledBlock
	// TypeSidebar is a sidebar block node type.
	TypeSidebar
	// TypePassThrough is a passthrough block node type.
	TypePassThrough
	// TypeVerse is a verse block node type.
	TypeVerse
	// TypeAttribute is an attribute entry node type.
	TypeAttribute
	// TypeInline is inline content type (text, formatting, links, etc.).
	TypeInline
	// TypeCallout is a callout annotation node type.
	TypeCallout
	// TypeCalloutList is a list of callout descriptions.
	TypeCalloutList
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
	// Children are blocks that belong to this section.
	Children []Node
	// Pos is the location in the source.
	Pos Position
}

// NodeParagraph is a paragraph.
type NodeParagraph struct {
	// Kind is the node type.
	Kind NodeType
	// Text is the paragraph text content.
	Text string
	// InlineNodes contains inline markup nodes found within the paragraph.
	InlineNodes []interface{}
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
	// NestedList contains a nested list within this list item.
	NestedList *NodeList
	// InlineNodes contains inline markup nodes found within the list item.
	InlineNodes []interface{}
	// Pos is the location in the source.
	Pos Position
}

// NodeLiteral is a literal/verbatim block.
type NodeLiteral struct {
	// Kind is the node type.
	Kind NodeType
	// Lines are the literal block lines.
	Lines []string
	// Callouts are callout annotations found in this block.
	Callouts []*CalloutNode
	// LineComment is the custom line comment prefix (e.g., "%", "//", etc.)
	LineComment string
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

// AdmonitionNode is an admonition block.
type AdmonitionNode struct {
	// Kind is the admonition kind (NOTE, WARNING, TIP, CAUTION, IMPORTANT).
	Kind string
	// Text is the admonition content.
	Text string
	// Pos is the location in the source.
	Pos Position
}

// MacroNode is a block macro (image, video, audio, include, etc.).
type MacroNode struct {
	// Kind is the node type.
	Kind NodeType
	// Target is the macro target (image, video, audio, include, etc.).
	Target string
	// Path is the macro path or reference.
	Path string
	// Attributes are macro-specific attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// Table is defined in ast_table.go with full AsciiDoc table support.
// The Table type is imported from ast_table.go.

// StyledBlockNode is a styled block (pass::[], sidebar::[], verse::[], etc.).
type StyledBlockNode struct {
	// Style is the block style (pass, sidebar, verse, quote, example, etc.).
	Style string
	// Content is the block content.
	Content string
	// Attributes are block-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// SidebarNode is a sidebar block.
type SidebarNode struct {
	// Title is the optional sidebar title.
	Title string
	// Content is the sidebar content.
	Content string
	// Attributes are block-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// PassThroughNode is a passthrough block (content is output as-is).
type PassThroughNode struct {
	// Content is the raw content to pass through.
	Content string
	// Attributes are block-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// VerseNode is a verse block (preserves line breaks).
type VerseNode struct {
	// Content is the verse content (with line breaks preserved).
	Content string
	// Attributes are block-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// CalloutNode represents a single callout annotation within a literal block.
type CalloutNode struct {
	// Number is the callout number (e.g., 1, 2, 3).
	Number int
	// LineIndex is the index of the line in the literal block (0-based).
	LineIndex int
	// Column is the column position where the callout appears.
	Column int
	// Description is the callout description text (parsed from callout list).
	Description string
	// Pos is the location in the source.
	Pos Position
}

// CalloutListNode represents a list of callout descriptions after a literal block.
type CalloutListNode struct {
	// Items are the callout descriptions keyed by number.
	Items map[int]*CalloutNode
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

func (n *AdmonitionNode) Type() NodeType   { return TypeAdmonition }
func (n *AdmonitionNode) Position() Position { return n.Pos }

func (n *MacroNode) Type() NodeType   { return TypeMacro }
func (n *MacroNode) Position() Position { return n.Pos }

func (n *StyledBlockNode) Type() NodeType   { return TypeStyledBlock }
func (n *StyledBlockNode) Position() Position { return n.Pos }

func (n *SidebarNode) Type() NodeType   { return TypeSidebar }
func (n *SidebarNode) Position() Position { return n.Pos }

func (n *PassThroughNode) Type() NodeType   { return TypePassThrough }
func (n *PassThroughNode) Position() Position { return n.Pos }

func (n *VerseNode) Type() NodeType   { return TypeVerse }
func (n *VerseNode) Position() Position { return n.Pos }

func (n *CalloutNode) Type() NodeType   { return TypeCallout }
func (n *CalloutNode) Position() Position { return n.Pos }

func (n *CalloutListNode) Type() NodeType   { return TypeCalloutList }
func (n *CalloutListNode) Position() Position { return n.Pos }