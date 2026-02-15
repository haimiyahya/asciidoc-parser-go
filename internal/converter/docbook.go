// Package converter provides DocBook 5.1.1 converter for AsciiDoc AST.
package converter

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
)

// DocBookConverter converts an AsciiDoc AST to DocBook 5.1.1 XML.
type DocBookConverter struct {
	// indent is current indentation level for pretty-printing.
	indent string

	// pretty enables pretty-printing with indentation.
	pretty bool

	// suppressHeaderFooter omits the XML declaration.
	suppressHeaderFooter bool

	// doctype specifies the document type (article, book, etc.)
	doctype string
}

// NewDocBookConverter creates a new DocBook 5.1.1 converter.
func NewDocBookConverter() *DocBookConverter {
	return &DocBookConverter{
		indent:               "",
		pretty:               true,
		suppressHeaderFooter: false,
		doctype:              "article",
	}
}

// WithDoctype sets the document type (article, book, etc.).
func (c *DocBookConverter) WithDoctype(doctype string) *DocBookConverter {
	c.doctype = doctype
	return c
}

// WithoutHeaderFooter configures the converter to omit the XML declaration.
func (c *DocBookConverter) WithoutHeaderFooter() *DocBookConverter {
	c.suppressHeaderFooter = true
	return c
}

// escapeXML escapes special XML characters.
func (c *DocBookConverter) escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// Convert converts document to DocBook 5.1.1 XML.
func (c *DocBookConverter) Convert(doc *ast.NodeDocument, w io.Writer) error {
	var buf bytes.Buffer

	// Write XML declaration if not suppressed
	if !c.suppressHeaderFooter {
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	}

	// Determine root element based on doctype
	rootElement := "article"
	if c.doctype == "book" {
		rootElement = "book"
	}
	// Check document attributes for doctype override
	if dt, ok := doc.Attributes["doctype"]; ok {
		if dt == "book" {
			rootElement = "book"
		}
	}

	// Start root element
	buf.WriteString(c.indent)
	buf.WriteString(fmt.Sprintf("<%s xmlns=\"http://docbook.org/ns/docbook\"", rootElement))
	buf.WriteString(" xmlns:xl=\"http://www.w3.org/1999/xlink\"")
	buf.WriteString(" version=\"5.1\">\n")

	// Write document title if present
	if doc.Header != nil && doc.Header.Title != "" {
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<info>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString(fmt.Sprintf("<title>%s</title>\n", c.escapeXML(doc.Header.Title)))
		// Write author if present
		if doc.Header.Author != "" {
			buf.WriteString(c.indent)
			buf.WriteString("<author>\n")
			c.incIndent()
			buf.WriteString(c.indent)
			buf.WriteString(fmt.Sprintf("<personname>%s</personname>\n", c.escapeXML(doc.Header.Author)))
			c.decIndent()
			buf.WriteString(c.indent)
			buf.WriteString("</author>\n")
		}
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</info>\n")
		c.decIndent()
	}

	// Convert all blocks
	for _, block := range doc.Blocks {
		c.convertNode(block, &buf)
	}

	// End root element
	buf.WriteString(fmt.Sprintf("</%s>\n", rootElement))

	// Write to output
	_, err := w.Write(buf.Bytes())
	return err
}

// convertNode converts an AST node to DocBook XML.
func (c *DocBookConverter) convertNode(node ast.Node, buf *bytes.Buffer) {
	switch n := node.(type) {
	case *ast.NodeParagraph:
		c.convertParagraph(n, buf)
	case *ast.NodeSection:
		c.convertSection(n, buf)
	case *ast.NodeList:
		c.convertList(n, buf)
	case *ast.NodeListItem:
		c.convertListItem(n, buf)
	case *ast.NodeLiteral:
		c.convertLiteral(n, buf)
	case *ast.NodeBlock:
		c.convertBlock(n, buf)
	case *ast.Table:
		c.convertTable(n, buf)
	case *ast.AdmonitionNode:
		c.convertAdmonition(n, buf)
	case *ast.MacroNode:
		c.convertMacro(n, buf)
	}
}

// convertParagraph converts a paragraph to DocBook.
func (c *DocBookConverter) convertParagraph(para *ast.NodeParagraph, buf *bytes.Buffer) {
	buf.WriteString(c.indent)
	buf.WriteString("<para>")

	// Write text content mixed with inline nodes
	if len(para.InlineNodes) == 0 {
		buf.WriteString(c.escapeXML(para.Text))
	} else {
		lastEnd := 0
		for _, node := range para.InlineNodes {
			if inlineNode, ok := node.(*inline.Node); ok {
				startPos := inlineNode.StartPos
				// Write any text before this inline node
				if startPos > lastEnd {
					text := para.Text[lastEnd:startPos]
					buf.WriteString(c.escapeXML(text))
				}
				lastEnd = inlineNode.Position

				// Render the inline node
				c.convertInlineNode(inlineNode, buf)
			}
		}

		// Write any remaining text after last inline node
		if lastEnd < len(para.Text) {
			text := para.Text[lastEnd:]
			buf.WriteString(c.escapeXML(text))
		}
	}

	buf.WriteString("</para>\n")
}

// convertInlineNode converts an inline.Node to DocBook XML.
func (c *DocBookConverter) convertInlineNode(node *inline.Node, buf *bytes.Buffer) {
	switch node.Type {
	case inline.NodeText:
		buf.WriteString(c.escapeXML(node.Text))
	case inline.NodeBold:
		buf.WriteString("<emphasis role=\"strong\">")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(c.escapeXML(node.Text))
		}
		buf.WriteString("</emphasis>")
	case inline.NodeItalic:
		buf.WriteString("<emphasis>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(c.escapeXML(node.Text))
		}
		buf.WriteString("</emphasis>")
	case inline.NodeMonospace:
		buf.WriteString("<literal>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(c.escapeXML(node.Text))
		}
		buf.WriteString("</literal>")
	case inline.NodeSubscript:
		buf.WriteString("<subscript>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(c.escapeXML(node.Text))
		}
		buf.WriteString("</subscript>")
	case inline.NodeSuperscript:
		buf.WriteString("<superscript>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(c.escapeXML(node.Text))
		}
		buf.WriteString("</superscript>")
	case inline.NodeLink:
		if node.URL != "" {
			buf.WriteString(fmt.Sprintf("<link xl:href=\"%s\">", c.escapeXML(node.URL)))
		} else {
			buf.WriteString("<link>")
		}
		c.renderInlineChildren(node, buf)
		buf.WriteString("</link>")
	case inline.NodeImage:
		alt := node.Alt
		if alt == "" {
			alt = node.Text
		}
		// DocBook uses <inlinemediaobject> for inline images
		buf.WriteString("<inlinemediaobject>")
		buf.WriteString(fmt.Sprintf("<imageobject><imagedata fileref=\"%s\"/></imageobject>", c.escapeXML(node.URL)))
		if alt != "" {
			buf.WriteString(fmt.Sprintf("<textobject><phrase>%s</phrase></textobject>", c.escapeXML(alt)))
		}
		buf.WriteString("</inlinemediaobject>")
	case inline.NodeCrossRef:
		// Cross-reference: <<section-id>> becomes <link linkend="section-id">
		buf.WriteString(fmt.Sprintf("<link linkend=\"%s\">", c.escapeXML(node.Ref)))
		c.renderInlineChildren(node, buf)
		buf.WriteString("</link>")
	}
}

// renderInlineChildren renders child inline nodes within a parent inline node.
func (c *DocBookConverter) renderInlineChildren(node *inline.Node, buf *bytes.Buffer) {
	if len(node.Children) > 0 {
		lastEnd := 0
		for _, child := range node.Children {
			end := child.Position
			// Write any text before this child node
			if end > lastEnd {
				text := node.Text[lastEnd:end]
				buf.WriteString(c.escapeXML(text))
			}
			lastEnd = end
			// Render the child node
			c.convertInlineNode(child, buf)
		}
		// Write any remaining text after last child
		if lastEnd > 0 && lastEnd < len(node.Text) {
			text := node.Text[lastEnd:]
			buf.WriteString(c.escapeXML(text))
		}
	} else {
		// No children, render the text directly
		buf.WriteString(c.escapeXML(node.Text))
	}
}

// convertSection converts a section to DocBook.
func (c *DocBookConverter) convertSection(section *ast.NodeSection, buf *bytes.Buffer) {
	c.incIndent()

	// Write section opening tag with xml:id if section has an ID
	buf.WriteString(c.indent)
	if section.ID != "" {
		buf.WriteString(fmt.Sprintf("<section xml:id=\"%s\">\n", c.escapeXML(section.ID)))
	} else {
		buf.WriteString("<section>\n")
	}

	// Write section title
	c.incIndent()
	buf.WriteString(c.indent)
	buf.WriteString(fmt.Sprintf("<title>%s</title>\n", c.escapeXML(section.Title)))
	c.decIndent()

	// Convert section children
	for _, child := range section.Children {
		c.convertNode(child, buf)
	}

	// Close section tag
	buf.WriteString(c.indent)
	buf.WriteString("</section>\n")

	c.decIndent()
}

// convertList converts a list to DocBook.
func (c *DocBookConverter) convertList(list *ast.NodeList, buf *bytes.Buffer) {
	if len(list.Items) == 0 {
		return
	}

	// Determine list type based on first item
	item := list.Items[0]
	listItem, ok := item.(*ast.NodeListItem)
	if !ok {
		return
	}

	c.incIndent()

	switch listItem.Marker {
	case "-", "*", "o":
		// Unordered list
		buf.WriteString(c.indent)
		buf.WriteString("<itemizedlist>\n")
		c.incIndent()
		for _, item := range list.Items {
			if li, ok := item.(*ast.NodeListItem); ok {
				c.convertListItemDocBook(li, buf, false)
			}
		}
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</itemizedlist>\n")
	case ".":
		// Ordered list
		buf.WriteString(c.indent)
		buf.WriteString("<orderedlist>\n")
		c.incIndent()
		for _, item := range list.Items {
			if li, ok := item.(*ast.NodeListItem); ok {
				c.convertListItemDocBook(li, buf, false)
			}
		}
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</orderedlist>\n")
	case "::":
		// Variable list (definition list)
		buf.WriteString(c.indent)
		buf.WriteString("<variablelist>\n")
		c.incIndent()
		for _, item := range list.Items {
			if li, ok := item.(*ast.NodeListItem); ok {
				c.convertListItemDocBook(li, buf, true)
			}
		}
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</variablelist>\n")
	}

	c.decIndent()
}

// convertListItem converts a list item to DocBook.
func (c *DocBookConverter) convertListItem(item *ast.NodeListItem, buf *bytes.Buffer) {
	c.convertListItemDocBook(item, buf, item.Marker == "::")
}

// convertListItemDocBook converts a list item to DocBook with format awareness.
func (c *DocBookConverter) convertListItemDocBook(item *ast.NodeListItem, buf *bytes.Buffer, isVariableList bool) {
	c.incIndent()

	if isVariableList {
		// Variable list item
		buf.WriteString(c.indent)
		buf.WriteString("<varlistentry>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString(fmt.Sprintf("<term>%s</term>\n", c.escapeXML(item.Term)))
		c.decIndent()
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<listitem>\n")
		c.incIndent()
		c.renderInlineTextInPara(item.Text, item.InlineNodes, buf)
		c.decIndent()
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</listitem>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</varlistentry>\n")
	} else if item.NestedList != nil {
		// Has nested list
		buf.WriteString(c.indent)
		buf.WriteString("<listitem>\n")
		c.incIndent()
		c.renderInlineTextInPara(item.Text, item.InlineNodes, buf)
		c.convertNode(item.NestedList, buf)
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</listitem>\n")
	} else {
		// Regular list item
		buf.WriteString(c.indent)
		buf.WriteString("<listitem>\n")
		c.incIndent()
		c.renderInlineTextInPara(item.Text, item.InlineNodes, buf)
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</listitem>\n")
	}

	c.decIndent()
}

// renderInlineTextInPara renders text with inline nodes wrapped in a para element.
func (c *DocBookConverter) renderInlineTextInPara(text string, inlineNodes []interface{}, buf *bytes.Buffer) {
	buf.WriteString(c.indent)
	buf.WriteString("<para>")
	lastEnd := 0
	for _, node := range inlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			startPos := inlineNode.StartPos
			// Write any text before this inline node
			if startPos > lastEnd {
				plainText := text[lastEnd:startPos]
				buf.WriteString(c.escapeXML(plainText))
			}
			lastEnd = inlineNode.Position

			// Render the inline node
			c.convertInlineNode(inlineNode, buf)
		}
	}

	// Write any remaining text after last inline node
	if lastEnd < len(text) {
		plainText := text[lastEnd:]
		buf.WriteString(c.escapeXML(plainText))
	}

	buf.WriteString("</para>\n")
}

// convertLiteral converts a literal block to DocBook.
func (c *DocBookConverter) convertLiteral(literal *ast.NodeLiteral, buf *bytes.Buffer) {
	c.incIndent()
	buf.WriteString(c.indent)
	buf.WriteString("<programlisting>")

	// Preserve whitespace in literal blocks
	content := strings.Join(literal.Lines, "\n")
	buf.WriteString(c.escapeXML(content))

	buf.WriteString("</programlisting>\n")
	c.decIndent()
}

// convertBlock converts a delimited block to DocBook.
func (c *DocBookConverter) convertBlock(block *ast.NodeBlock, buf *bytes.Buffer) {
	// Map delimiters to DocBlock elements
	c.incIndent()

	switch block.Delimiter {
	case "=":
		// Example block
		buf.WriteString(c.indent)
		buf.WriteString("<example>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<title>Example</title>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<programlisting>")
		buf.WriteString(c.escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</programlisting>\n")
		buf.WriteString(c.indent)
		buf.WriteString("</example>\n")
	case "_":
		// Quote block
		buf.WriteString(c.indent)
		buf.WriteString("<blockquote>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<para>")
		buf.WriteString(c.escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</para>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</blockquote>\n")
	case "-":
		// Verbatim block
		buf.WriteString(c.indent)
		buf.WriteString("<screen>")
		buf.WriteString(c.escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</screen>\n")
	case "+":
		// Pass block - output as-is (in CDATA)
		buf.WriteString(c.indent)
		buf.WriteString("<![CDATA[")
		buf.WriteString(strings.Join(block.Lines, "\n"))
		buf.WriteString("]]>\n")
	case "*":
		// Sidebar block
		buf.WriteString(c.indent)
		buf.WriteString("<sidebar>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<para>")
		buf.WriteString(c.escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</para>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</sidebar>\n")
	case "/":
		// Literal block
		buf.WriteString(c.indent)
		buf.WriteString("<programlisting>")
		buf.WriteString(c.escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</programlisting>\n")
	default:
		// Default: treat as para
		buf.WriteString(c.indent)
		buf.WriteString("<para>")
		buf.WriteString(c.escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</para>\n")
	}

	c.decIndent()
}

// convertAdmonition converts an admonition to DocBook.
func (c *DocBookConverter) convertAdmonition(admonition *ast.AdmonitionNode, buf *bytes.Buffer) {
	c.incIndent()

	// Map admonition kinds to DocBook elements
	element := "note"
	switch strings.ToLower(admonition.Kind) {
	case "tip":
		element = "tip"
	case "warning":
		element = "warning"
	case "caution":
		element = "caution"
	case "important":
		element = "important"
	}

	buf.WriteString(c.indent)
	buf.WriteString(fmt.Sprintf("<%s>\n", element))
	c.incIndent()
	buf.WriteString(c.indent)
	buf.WriteString("<para>")
	buf.WriteString(c.escapeXML(admonition.Text))
	buf.WriteString("</para>\n")
	c.decIndent()
	buf.WriteString(c.indent)
	buf.WriteString(fmt.Sprintf("</%s>\n", element))

	c.decIndent()
}

// convertMacro converts a block macro to DocBook.
func (c *DocBookConverter) convertMacro(macro *ast.MacroNode, buf *bytes.Buffer) {
	// For now, render macros as comments
	// In a full implementation, this would handle specific macro types
	switch macro.Target {
	case "image":
		// <mediaobject> with <imageobject>
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<mediaobject>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<imageobject>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString(fmt.Sprintf("<imagedata fileref=\"%s\"/>\n", c.escapeXML(macro.Path)))
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</imageobject>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</mediaobject>\n")
		c.decIndent()
	case "video":
		// <mediaobject> with <videoobject>
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<mediaobject>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<videoobject>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString(fmt.Sprintf("<videodata fileref=\"%s\"/>\n", c.escapeXML(macro.Path)))
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</videoobject>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</mediaobject>\n")
		c.decIndent()
	case "audio":
		// <mediaobject> with <audioobject>
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<mediaobject>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<audioobject>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString(fmt.Sprintf("<audiodata fileref=\"%s\"/>\n", c.escapeXML(macro.Path)))
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</audioobject>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</mediaobject>\n")
		c.decIndent()
	default:
		// Unknown macro - render as a comment for now
		buf.WriteString(c.indent)
		buf.WriteString(fmt.Sprintf("<!-- %s::%s -->\n", c.escapeXML(macro.Target), c.escapeXML(macro.Path)))
	}
}

// convertTable converts a table to DocBook.
func (c *DocBookConverter) convertTable(table *ast.Table, buf *bytes.Buffer) {
	c.incIndent()

	buf.WriteString(c.indent)
	buf.WriteString("<table>\n")

	// Write header if present
	if len(table.Header) > 0 {
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<thead>\n")
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<row>\n")
		c.incIndent()
		for _, cell := range table.Header {
			buf.WriteString(c.indent)
			buf.WriteString(fmt.Sprintf("<entry>%s</entry>\n", c.escapeXML(cell)))
		}
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</row>\n")
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</thead>\n")
		c.decIndent()
	}

	// Write body
	if len(table.Body) > 0 {
		c.incIndent()
		buf.WriteString(c.indent)
		buf.WriteString("<tbody>\n")
		for _, row := range table.Body {
			c.incIndent()
			buf.WriteString(c.indent)
			buf.WriteString("<row>\n")
			c.incIndent()
			for _, cell := range row {
				buf.WriteString(c.indent)
				buf.WriteString(fmt.Sprintf("<entry>%s</entry>\n", c.escapeXML(cell)))
			}
			c.decIndent()
			buf.WriteString(c.indent)
			buf.WriteString("</row>\n")
			c.decIndent()
		}
		c.decIndent()
		buf.WriteString(c.indent)
		buf.WriteString("</tbody>\n")
		c.decIndent()
	}

	buf.WriteString(c.indent)
	buf.WriteString("</table>\n")

	c.decIndent()
}

// incIndent increases the indentation level.
func (c *DocBookConverter) incIndent() {
	c.indent += "  "
}

// decIndent decreases the indentation level.
func (c *DocBookConverter) decIndent() {
	if len(c.indent) >= 2 {
		c.indent = c.indent[:len(c.indent)-2]
	}
}

// MarshalXML implements xml.Marshaler for DocBook output.
func (c *DocBookConverter) MarshalXML(doc *ast.NodeDocument) ([]byte, error) {
	var buf bytes.Buffer
	err := c.Convert(doc, &buf)
	return buf.Bytes(), err
}
