// Package ast defines the Abstract Syntax Tree for AsciiDoc documents.
//
// The AST represents the hierarchical structure of a parsed AsciiDoc document.
// All nodes implement the Node interface, providing Type() and Position()
// methods for introspection.
//
// # Node Hierarchy
//
// The document root is a NodeDocument, which contains:
//   - A DocumentHeader with title, author, and metadata
//   - Top-level blocks (sections, paragraphs, lists, tables, etc.)
//   - Document-level attributes
//
// Sections (NodeSection) can contain nested blocks, including subsections.
// Lists (NodeList) contain ListItems, which may have nested lists.
//
// # Node Types
//
// The NodeType enum identifies all node types:
//   - TypeDocument: Root document node
//   - TypeSection: Section heading
//   - TypeParagraph: Text paragraph
//   - TypeList: Ordered/unordered/labeled list
//   - TypeListItem: Single list item
//   - TypeAdmonition: NOTE/TIP/WARNING/CAUTION/IMPORTANT
//   - TypeLiteral: Verbatim/literal block
//   - TypeBlock: Generic delimited block
//   - TypeTable: Data table
//   - TypeStyledBlock: Styled block (pass, sidebar, verse, etc.)
//   - TypeMacro: Block macro (image, include, etc.)
//   - And more specialized types...
//
// # Position Tracking
//
// Each node carries a Position field indicating its location in the
// source document for error reporting and debugging.
package ast

// Node is the interface implemented by all AST nodes.
//
// The Node interface provides type identification and source location
// tracking for all nodes in the AST.
type Node interface {
	// Type returns the node type for type switching and identification.
	Type() NodeType

	// Position returns the source location for error reporting.
	Position() Position
}

// DocumentHeader contains the document header information.
//
// The header stores document-level metadata including the title,
// author, version, and other front-matter attributes.
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
	Number string
	Date   string
	Remark string
}

// NodeType represents the type of an AST node.
//
// NodeType is an enumeration that identifies the kind of node
// for type switching and pattern matching.
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
	// TypeBibliography is a bibliography section node type.
	TypeBibliography
	// TypeBibliographyEntry is a single bibliography entry node type.
	TypeBibliographyEntry
)

// Position represents a location in the source document.
//
// Position tracks the file and line number for error reporting
// and source mapping.
type Position struct {
	File string
	Line int
}

// NodeDocument is the root document node.
//
// The NodeDocument represents the entire parsed AsciiDoc document,
// including the header, all blocks, and document-level attributes.
type NodeDocument struct {
	// Kind is the node type.
	Kind NodeType
	// Header contains the document header (title, attributes, etc).
	Header *DocumentHeader
	// Blocks are the top-level blocks in the document.
	Blocks []Node
	// Attributes are document-level attributes.
	Attributes map[string]string
	// BibliographyEntries maps citation labels to their entries for citations lookup.
	BibliographyEntries map[string]*BibliographyEntryNode
	// Pos is the location in the source.
	Pos Position
}

// NodeSection is a section heading.
//
// Sections have a level (0-6, where 0 is the document title),
// a title, an optional ID for cross-references, and child blocks.
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
//
// Paragraphs contain text with optional inline markup nodes
// (bold, italic, links, etc.) for formatting.
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
//
// Lists contain items and may have an optional style (e.g., "qanda")
// and list-level attributes.
type NodeList struct {
	// Kind is the node type.
	Kind NodeType
	// Items are the list items.
	Items []Node
	// Style is the list style (e.g., "qanda", "bibliography").
	// Empty string for default styling.
	Style string
	// Attributes are list-level attributes (e.g., "start" for ordered lists).
	Attributes map[string]string
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
	// Checked indicates if this is a checked checklist item.
	// For checklists: true for "[x]", false for "[ ]"
	Checked bool
	// Text is the item text content.
	Text string
	// Definition is the definition text for labeled lists (item.Definition).
	Definition string
	// DefinitionNodes contains inline markup nodes for the definition text.
	DefinitionNodes []interface{}
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
	// Caption is the block title/caption (e.g., from ".Title" before the block).
	Caption string
	// LineNumbers enables line number display.
	LineNumbers bool
	// StartLineNumber is the starting line number (default 1).
	StartLineNumber int
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
	// Caption is the block title/caption (e.g., from ".Title" before the block).
	Caption string
	// Pos is the location in the source.
	Pos Position
}

// AdmonitionNode is an admonition block.
type AdmonitionNode struct {
	// Kind is the admonition kind (NOTE, WARNING, TIP, CAUTION, IMPORTANT).
	Kind string
	// Text is the admonition content (for inline-style admonitions like "NOTE: text").
	Text string
	// Title is the optional admonition title (for block-style admonitions).
	Title string
	// Blocks are child blocks within the admonition (for block-style with multi-paragraph content).
	Blocks []Node
	// Attributes are additional admonition attributes.
	Attributes map[string]string
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
	// Callouts are callout annotations found in this block (for source/listing blocks).
	Callouts []*CalloutNode
	// Caption is the block title/caption (e.g., from ".Title" before the block).
	Caption string
	// LineNumbers enables line number display (for source/listing blocks).
	LineNumbers bool
	// StartLineNumber is the starting line number (default 1).
	StartLineNumber int
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

// BibliographyNode represents a bibliography section.
type BibliographyNode struct {
	// Title is the bibliography section title.
	Title string
	// ID is the section ID (for anchors).
	ID string
	// Entries are the bibliography entries.
	Entries []*BibliographyEntryNode
	// Attributes are section-level attributes.
	Attributes map[string]string
	// Pos is the location in the source.
	Pos Position
}

// BibliographyEntryNode represents a single bibliography entry with triple-bracket anchor.
type BibliographyEntryNode struct {
	// Label is the citation identifier (e.g., "pp" for [[[pp]]]).
	Label string
	// XRefText is the optional custom reference text (e.g., "gang" for [[[gof,gang]]]).
	XRefText string
	// Text is the entry text content.
	Text string
	// InlineNodes contains inline markup nodes found within the entry text.
	InlineNodes []interface{}
	// Pos is the location in the source.
	Pos Position
}

// Node interface methods for each type.

func (n *NodeDocument) Type() NodeType     { return n.Kind }
func (n *NodeDocument) Position() Position { return n.Pos }

func (n *NodeSection) Type() NodeType     { return TypeSection }
func (n *NodeSection) Position() Position { return n.Pos }

func (n *NodeParagraph) Type() NodeType     { return TypeParagraph }
func (n *NodeParagraph) Position() Position { return n.Pos }

func (n *NodeList) Type() NodeType     { return TypeList }
func (n *NodeList) Position() Position { return n.Pos }

func (n *NodeListItem) Type() NodeType     { return TypeListItem }
func (n *NodeListItem) Position() Position { return n.Pos }

func (n *NodeLiteral) Type() NodeType     { return TypeLiteral }
func (n *NodeLiteral) Position() Position { return n.Pos }

func (n *NodeBlock) Type() NodeType     { return TypeBlock }
func (n *NodeBlock) Position() Position { return n.Pos }

func (n *AdmonitionNode) Type() NodeType     { return TypeAdmonition }
func (n *AdmonitionNode) Position() Position { return n.Pos }

func (n *MacroNode) Type() NodeType     { return TypeMacro }
func (n *MacroNode) Position() Position { return n.Pos }

func (n *StyledBlockNode) Type() NodeType     { return TypeStyledBlock }
func (n *StyledBlockNode) Position() Position { return n.Pos }

func (n *SidebarNode) Type() NodeType     { return TypeSidebar }
func (n *SidebarNode) Position() Position { return n.Pos }

func (n *PassThroughNode) Type() NodeType     { return TypePassThrough }
func (n *PassThroughNode) Position() Position { return n.Pos }

func (n *VerseNode) Type() NodeType     { return TypeVerse }
func (n *VerseNode) Position() Position { return n.Pos }

func (n *CalloutNode) Type() NodeType     { return TypeCallout }
func (n *CalloutNode) Position() Position { return n.Pos }

func (n *CalloutListNode) Type() NodeType     { return TypeCalloutList }
func (n *CalloutListNode) Position() Position { return n.Pos }

func (n *BibliographyNode) Type() NodeType     { return TypeBibliography }
func (n *BibliographyNode) Position() Position { return n.Pos }

func (n *BibliographyEntryNode) Type() NodeType     { return TypeBibliographyEntry }
func (n *BibliographyEntryNode) Position() Position { return n.Pos }
