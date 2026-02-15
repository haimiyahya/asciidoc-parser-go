// Package ast provides Abstract Syntax Tree nodes for AsciiDoc, including attribute support.
package ast

// AttributeNode represents an attribute entry in the AST.
type AttributeNode struct {
	// Kind is the attribute kind (e.g., "role", "id", "style").
	Kind string
	// Name is the attribute name (e.g., "role", "toc", "title").
	Name string

	// Value is the attribute value (for key-value pairs).
	Value string

	// Pos is the position in source text.
	Pos Position
}

// String returns a string representation of node.
func (n *AttributeNode) String() string {
	return "[" + n.Kind + "=" + n.Name + "=" + n.Value + "]"
}
