// Package converter provides HTML5 converter for AsciiDoc AST.
package converter

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// HTML5Converter converts an AsciiDoc AST to HTML5.
type HTML5Converter struct {
	// indent is current indentation level for pretty-printing.
	indent string

	// pretty enables pretty-printing with indentation.
	pretty bool
}

// NewHTML5Converter creates a new HTML5 converter.
func NewHTML5Converter() *HTML5Converter {
	return &HTML5Converter{
		indent: "",
		pretty: true,
	}
}

// escape escapes HTML special characters.
func (c *HTML5Converter) escape(s string) string {
	return html.EscapeString(s)
}

// writeElement writes a simple HTML element with content (inline).
func (c *HTML5Converter) writeElement(tag string, content string, w io.Writer) {
	c.writeElementWithClass(tag, "", content, w)
}

// writeElementWithClass writes an HTML element with class attribute and content (inline).
func (c *HTML5Converter) writeElementWithClass(tag, class, content string, w io.Writer) {
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "<")
	fmt.Fprint(w, tag)
	if class != "" {
		fmt.Fprintf(w, " class=\"%s\"", class)
	}
	fmt.Fprint(w, ">")
	fmt.Fprint(w, content)
	fmt.Fprintf(w, "</%s>", tag)
	if c.pretty {
		fmt.Fprintln(w)
	}
}

// writeOpenTag writes an opening HTML tag.
func (c *HTML5Converter) writeOpenTag(tag string, w io.Writer) {
	c.writeOpenTagWithClass(tag, "", w)
}

// writeOpenTagWithClass writes an opening HTML tag with class attribute.
func (c *HTML5Converter) writeOpenTagWithClass(tag, class string, w io.Writer) {
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "<")
	fmt.Fprint(w, tag)
	if class != "" {
		fmt.Fprintf(w, " class=\"%s\"", class)
	}
	fmt.Fprint(w, ">")
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}
}

// writeCloseTag writes a closing HTML tag.
func (c *HTML5Converter) writeCloseTag(tag string, w io.Writer) {
	if c.pretty {
		c.indent = strings.TrimSuffix(c.indent, "  ")
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, "</%s>", tag)
	if c.pretty {
		fmt.Fprintln(w)
	}
}

// Convert converts document to HTML5.
func (c *HTML5Converter) Convert(doc *ast.NodeDocument, w io.Writer) error {
	// Start HTML5 document
	fmt.Fprint(w, "<!DOCTYPE html>")
	if c.pretty {
		fmt.Fprintln(w)
	}
	c.writeOpenTag("html", w)
	c.writeOpenTag("body", w)

	// Convert document header if present
	if doc.Header != nil {
		c.convertHeader(doc.Header, w)
	}

	// Convert all blocks
	for _, block := range doc.Blocks {
		c.convertNode(block, w)
	}

	// Close HTML
	c.writeCloseTag("body", w)
	c.writeCloseTag("html", w)

	return nil
}

// convertHeader converts a document header to HTML.
func (c *HTML5Converter) convertHeader(header *ast.DocumentHeader, w io.Writer) {
	if header.Title != "" {
		c.writeElement("h1", c.escape(header.Title), w)
	}
}

// convertNode converts an AST node to HTML.
func (c *HTML5Converter) convertNode(node ast.Node, w io.Writer) {
	switch n := node.(type) {
	case *ast.NodeParagraph:
		c.convertParagraph(n, w)
	case *ast.NodeSection:
		c.convertSection(n, w)
	case *ast.NodeList:
		c.convertList(n, w)
	case *ast.NodeListItem:
		c.convertListItem(n, w)
	case *ast.NodeLiteral:
		c.convertLiteral(n, w)
	case *ast.NodeBlock:
		c.convertBlock(n, w)
	case *ast.Table:
		c.convertTable(n, w)
	case *ast.AdmonitionNode:
		c.convertAdmonition(n, w)
	case *ast.MacroNode:
		c.convertMacro(n, w)
	default:
		// Unknown node type - skip
	}
}

// convertParagraph converts a paragraph to HTML.
func (c *HTML5Converter) convertParagraph(para *ast.NodeParagraph, w io.Writer) {
	c.writeElement("p", c.escape(para.Text), w)
}

// writeRawString writes raw text without indentation or newlines.
func (c *HTML5Converter) writeRawString(s string, w io.Writer) {
	fmt.Fprint(w, s)
	if c.pretty {
		fmt.Fprintln(w)
	}
}

// convertSection converts a section to HTML.
func (c *HTML5Converter) convertSection(section *ast.NodeSection, w io.Writer) {
	// Determine heading tag based on level
	tag := c.headingTag(section.Level)
	c.writeElement(tag, c.escape(section.Title), w)

	// TODO: Process section content (children)
}

// headingTag returns appropriate HTML tag for a section level.
// Level 1 maps to h2 (h1 is for document title), level 2 -> h3, etc.
func (c *HTML5Converter) headingTag(level int) string {
	if level >= 1 && level <= 5 {
		return fmt.Sprintf("h%d", level+1)
	}
	return "h6" // Default to h6 for levels >= 6
}

// convertList converts a list to HTML.
func (c *HTML5Converter) convertList(list *ast.NodeList, w io.Writer) {
	// Determine list type based on first item
	if len(list.Items) == 0 {
		return
	}

	tag := c.listTag(list.Items[0])
	if tag == "" {
		return
	}

	c.writeOpenTag(tag, w)

	// Convert all list items
	for _, item := range list.Items {
		c.convertNode(item, w)
	}

	c.writeCloseTag(tag, w)
}

// listTag returns appropriate HTML tag for a list.
func (c *HTML5Converter) listTag(item ast.Node) string {
	if li, ok := item.(*ast.NodeListItem); ok {
		switch li.Marker {
		case "-", "*", "o":
			return "ul" // Unordered list
		case ".":
			return "ol" // Ordered list
		case "::":
			return "dl" // Labeled/definition list
		}
	}
	return "ul" // Default
}

// convertListItem converts a list item to HTML.
func (c *HTML5Converter) convertListItem(item *ast.NodeListItem, w io.Writer) {
	// Determine tag based on list type
	if item.Marker == "::" {
		// Labeled list: dt and dd
		c.writeElement("dt", c.escape(item.Term), w)
		c.writeElement("dd", c.escape(item.Definition), w)
	} else if item.NestedList != nil {
		// Has nested list - open li, render text, nested list, close li
		c.writeOpenTag("li", w)
		c.writeElement("span", c.escape(item.Text), w)
		c.convertNode(item.NestedList, w)
		c.writeCloseTag("li", w)
	} else {
		// Regular list item without nested list
		c.writeElement("li", c.escape(item.Text), w)
	}
}

// convertLiteral converts a literal block to HTML.
func (c *HTML5Converter) convertLiteral(literal *ast.NodeLiteral, w io.Writer) {
	tag := "pre"
	content := c.escape(strings.Join(literal.Lines, "\n"))
	c.writeElement(tag, content, w)
}

// convertBlock converts a delimited block to HTML.
func (c *HTML5Converter) convertBlock(block *ast.NodeBlock, w io.Writer) {
	// Determine block type from delimiter
	tag := "div"
	class := c.blockClass(block.Delimiter)
	content := c.escape(strings.Join(block.Lines, "\n"))

	c.writeElementWithClass(tag, class, content, w)
}

// convertAdmonition converts an admonition to HTML.
func (c *HTML5Converter) convertAdmonition(admonition *ast.AdmonitionNode, w io.Writer) {
	// Determine class based on admonition kind
	class := "admonition-" + strings.ToLower(admonition.Kind)
	c.writeElementWithClass("div", class, c.escape(admonition.Text), w)
}

// convertMacro converts a block macro to HTML.
func (c *HTML5Converter) convertMacro(macro *ast.MacroNode, w io.Writer) {
	// For now, render as a comment showing the macro
	// In a full implementation, this would handle specific macro types
	// (image -> <img>, video -> <video>, etc.)
	switch macro.Target {
	case "image":
		// <img src="path" />
		fmt.Fprintf(w, `<img src="%s" alt="%s">`, c.escape(macro.Path), c.escape(macro.Path))
		if c.pretty {
			fmt.Fprintln(w)
		}
	case "video":
		// <video src="path"></video>
		fmt.Fprintf(w, `<video src="%s">`, c.escape(macro.Path))
		if c.pretty {
			fmt.Fprintln(w)
		}
		fmt.Fprint(w, `</video>`)
		if c.pretty {
			fmt.Fprintln(w)
		}
	case "audio":
		// <audio src="path"></audio>
		fmt.Fprintf(w, `<audio src="%s">`, c.escape(macro.Path))
		if c.pretty {
			fmt.Fprintln(w)
		}
		fmt.Fprint(w, `</audio>`)
		if c.pretty {
			fmt.Fprintln(w)
		}
	default:
		// Unknown macro - render as a comment for now
		fmt.Fprintf(w, `<!-- %s::%s -->`, macro.Target, c.escape(macro.Path))
		if c.pretty {
			fmt.Fprintln(w)
		}
	}
}

// blockClass returns CSS class for a delimited block.
func (c *HTML5Converter) blockClass(delimiter string) string {
	switch delimiter {
	case "=":
		return "exampleblock"
	case "_":
		return "quoteblock"
	case "-":
		return "verbatimblock"
	case "+":
		return "passblock"
	case "*":
		return "sidebarblock"
	case "/":
		return "literalblock"
	default:
		return ""
	}
}

// convertTable converts a table to HTML.
func (c *HTML5Converter) convertTable(table *ast.Table, w io.Writer) {
	c.writeOpenTag("table", w)

	// Write header if present
	if len(table.Header) > 0 {
		c.writeOpenTag("thead", w)
		c.writeOpenTag("tr", w)
		for _, cell := range table.Header {
			c.writeElement("th", c.escape(cell), w)
		}
		c.writeCloseTag("tr", w)
		c.writeCloseTag("thead", w)
	}

	// Write body
	if len(table.Body) > 0 {
		c.writeOpenTag("tbody", w)
		for _, row := range table.Body {
			c.writeOpenTag("tr", w)
			for _, cell := range row {
				c.writeElement("td", c.escape(cell), w)
			}
			c.writeCloseTag("tr", w)
		}
		c.writeCloseTag("tbody", w)
	}

	c.writeCloseTag("table", w)
}
