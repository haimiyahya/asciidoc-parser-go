// Package converter provides HTML5 converter for AsciiDoc AST.
package converter

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
)

// HTML5Converter converts an AsciiDoc AST to HTML5.
type HTML5Converter struct {
	// indent is current indentation level for pretty-printing.
	indent string

	// pretty enables pretty-printing with indentation.
	pretty bool

	// suppressHeaderFooter omits the HTML5 document shell (DOCTYPE, html, body tags).
	suppressHeaderFooter bool

	// document is the document being converted (for bibliography lookup)
	document *ast.NodeDocument
}

// NewHTML5Converter creates a new HTML5 converter.
func NewHTML5Converter() *HTML5Converter {
	return &HTML5Converter{
		indent:               "",
		pretty:               true,
		suppressHeaderFooter: false,
	}
}

// WithoutHeaderFooter configures the converter to omit the HTML5 document shell.
// When enabled, only the document content is output, not the DOCTYPE, html, or body tags.
// Also disables pretty-printing to match Asciidoctor's embedded output style.
func (c *HTML5Converter) WithoutHeaderFooter() *HTML5Converter {
	c.suppressHeaderFooter = true
	c.pretty = false // Disable pretty-printing for Asciidoctor compatibility
	return c
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
	// Store document reference for bibliography lookup
	c.document = doc

	if !c.suppressHeaderFooter {
		// Start HTML5 document
		fmt.Fprint(w, "<!DOCTYPE html>")
		if c.pretty {
			fmt.Fprintln(w)
		}
		c.writeOpenTag("html", w)
		c.writeOpenTag("body", w)
	}

	// Convert document header if present
	if doc.Header != nil {
		c.convertHeader(doc.Header, w)
	}

	// Convert all blocks
	for _, block := range doc.Blocks {
		c.convertNode(block, w)
	}

	if !c.suppressHeaderFooter {
		// Close HTML
		c.writeCloseTag("body", w)
		c.writeCloseTag("html", w)
	}

	return nil
}

// convertHeader converts a document header to HTML.
func (c *HTML5Converter) convertHeader(header *ast.DocumentHeader, w io.Writer) {
	// In embedded mode, don't output document title (Asciidoctor compatibility)
	if c.suppressHeaderFooter {
		return
	}
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
	case *ast.StyledBlockNode:
		c.convertStyledBlock(n, w)
	case *ast.SidebarNode:
		c.convertSidebar(n, w)
	case *ast.PassThroughNode:
		c.convertPassThrough(n, w)
	case *ast.VerseNode:
		c.convertVerse(n, w)
	case *ast.BibliographyNode:
		c.convertBibliography(n, w)
	default:
		// Unknown node type - skip
	}
}

// convertParagraph converts a paragraph to HTML.
func (c *HTML5Converter) convertParagraph(para *ast.NodeParagraph, w io.Writer) {
	// Always add newlines for Asciidoctor compatibility
	fmt.Fprintf(w, `<div class="paragraph">`+"\n")

	// If there are no inline nodes, use simple rendering
	if len(para.InlineNodes) == 0 {
		fmt.Fprintf(w, `<p>%s</p>`+"\n", c.escape(para.Text))
		fmt.Fprintf(w, `</div>`+"\n")
		return
	}

	// Complex paragraph with inline nodes
	fmt.Fprintf(w, `<p>`)

	// Write text content mixed with inline nodes
	lastEnd := 0
	for _, node := range para.InlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			startPos := inlineNode.StartPos
			// Write any text before this inline node
			if startPos > lastEnd {
				text := para.Text[lastEnd:startPos]
				c.writeRawString(c.escape(text), w)
			}
			lastEnd = inlineNode.Position

			// Render the inline node
			c.convertInlineNode(inlineNode, w)
		}
	}

	// Write any remaining text after last inline node
	if lastEnd < len(para.Text) {
		text := para.Text[lastEnd:]
		c.writeRawString(c.escape(text), w)
	}

	fmt.Fprintf(w, `</p>`+"\n")
	fmt.Fprintf(w, `</div>`+"\n")
}

// convertInlineNode converts an inline.Node to HTML.
func (c *HTML5Converter) convertInlineNode(node *inline.Node, w io.Writer) {
	switch node.Type {
	case inline.NodeText:
		c.writeRawString(c.escape(node.Text), w)
	case inline.NodeBold:
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<strong class="`+class+`">`, w)
		} else {
			c.writeRawString("<strong>", w)
		}
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, w)
		} else {
			c.writeRawString(c.escape(node.Text), w)
		}
		c.writeRawString("</strong>", w)
	case inline.NodeItalic:
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<em class="`+class+`">`, w)
		} else {
			c.writeRawString("<em>", w)
		}
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, w)
		} else {
			c.writeRawString(c.escape(node.Text), w)
		}
		c.writeRawString("</em>", w)
	case inline.NodeMonospace:
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<code class="`+class+`">`, w)
		} else {
			c.writeRawString("<code>", w)
		}
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, w)
		} else {
			c.writeRawString(c.escape(node.Text), w)
		}
		c.writeRawString("</code>", w)
	case inline.NodeSubscript:
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<sub class="`+class+`">`, w)
		} else {
			c.writeRawString("<sub>", w)
		}
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, w)
		} else {
			c.writeRawString(c.escape(node.Text), w)
		}
		c.writeRawString("</sub>", w)
	case inline.NodeSuperscript:
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<sup class="`+class+`">`, w)
		} else {
			c.writeRawString("<sup>", w)
		}
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, w)
		} else {
			c.writeRawString(c.escape(node.Text), w)
		}
		c.writeRawString("</sup>", w)
	case inline.NodeLink:
		if node.URL != "" {
			class := c.getClassAttr(node.Roles)
			// Build attributes string
			var attrs []string
			attrs = append(attrs, fmt.Sprintf(`href="%s"`, c.escape(node.URL)))

			// Add class if present
			if class != "" {
				attrs = append(attrs, fmt.Sprintf(`class="%s"`, class))
			}

			// Add additional attributes from node.Attributes
			// Map Asciidoctor attributes to HTML attributes
			for key, val := range node.Attributes {
				// Map window -> target (Asciidoctor compatibility)
				if key == "window" {
					attrs = append(attrs, fmt.Sprintf(`target="%s"`, c.escape(val)))
					// Asciidoctor also adds rel="noopener" for target="_blank"
					if val == "_blank" {
						if _, hasRel := node.Attributes["rel"]; !hasRel {
							attrs = append(attrs, `rel="noopener"`)
						}
					}
				} else {
					attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, c.escape(val)))
				}
			}

			fmt.Fprintf(w, `<a %s>`, strings.Join(attrs, " "))
		} else {
			c.writeRawString("<a>", w)
		}
		c.renderInlineChildren(node, w)
		c.writeRawString("</a>", w)
	case inline.NodeImage:
		alt := node.Alt
		if alt == "" {
			alt = node.Text
		}
		// Wrap in span class="image" (Asciidoctor compatibility)
		fmt.Fprint(w, `<span class="image">`)
		class := c.getClassAttr(node.Roles)
		if class != "" {
			fmt.Fprintf(w, `<img src="%s" alt="%s" class="%s">`, c.escape(node.URL), c.escape(alt), class)
		} else {
			fmt.Fprintf(w, `<img src="%s" alt="%s">`, c.escape(node.URL), c.escape(alt))
		}
		fmt.Fprint(w, `</span>`)
	case inline.NodeCrossRef:
		// Check if this is a bibliography citation (references a bibliography entry)
		if c.document != nil && c.document.BibliographyEntries != nil {
			if bibEntry, ok := c.document.BibliographyEntries[node.Ref]; ok {
				// This is a bibliography citation - render as [label] or [xreftext]
				if bibEntry.XRefText != "" {
					fmt.Fprintf(w, "[%s]", c.escape(bibEntry.XRefText))
				} else {
					fmt.Fprintf(w, "[%s]", c.escape(node.Ref))
				}
				break
			}
		}
		// Regular cross-reference: <<section-id>> or <<section-id,text>> becomes <a href="#section-id">
		class := c.getClassAttr(node.Roles)
		if class != "" {
			fmt.Fprintf(w, `<a href="#%s" class="%s">`, c.escape(node.Ref), class)
		} else {
			fmt.Fprintf(w, `<a href="#%s">`, c.escape(node.Ref))
		}
		if node.RefText != "" {
			c.writeRawString(c.escape(node.RefText), w)
		} else {
			c.renderInlineChildren(node, w)
		}
		c.writeRawString("</a>", w)
	case inline.NodeKbd:
		// Keyboard macro: kbd:[Ctrl+C] -> <kbd><span class="key">Ctrl</span>+<span class="key">C</span></kbd>
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<kbd class="`+class+`">`, w)
		} else {
			c.writeRawString("<kbd>", w)
		}
		c.renderKbdKeys(node.Text, w)
		c.writeRawString("</kbd>", w)
	case inline.NodeBtn:
		// Button macro: btn:[OK] -> <b class="btn">OK</b>
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<b class="btn `+class+`">`, w)
		} else {
			c.writeRawString(`<b class="btn">`, w)
		}
		c.writeRawString(c.escape(node.Text), w)
		c.writeRawString("</b>", w)
	case inline.NodeMenu:
		// Menu macro: menu:[File > Save] -> <span class="menu"><span class="menuitem">File</span> &#10140; <span class="menuitem">Save</span></span>
		class := c.getClassAttr(node.Roles)
		if class != "" {
			c.writeRawString(`<span class="menu `+class+`">`, w)
		} else {
			c.writeRawString(`<span class="menu">`, w)
		}
		c.renderMenuPath(node.Text, w)
		c.writeRawString("</span>", w)
	case inline.NodeIndexTerm:
		// Index terms: ((term)) is flow (visible), (((term))) is concealed (hidden)
		// For Asciidoctor compatibility in HTML5:
		// - Flow index terms render as plain text (no span wrapper)
		// - Concealed index terms are completely invisible (no output)
		if !node.IndexTermConcealed {
			// Flow index term - render as visible plain text only
			c.writeRawString(c.escape(node.Text), w)
		}
		// Concealed index terms produce no output
	}
}

// getClassAttr returns a space-separated list of roles for HTML class attribute.
func (c *HTML5Converter) getClassAttr(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	// Join roles with spaces for HTML class attribute
	return strings.Join(roles, " ")
}

// renderKbdKeys renders keyboard key combinations with proper formatting.
// Splits on "+" and wraps each key in a span.key element.
func (c *HTML5Converter) renderKbdKeys(keys string, w io.Writer) {
	keyParts := strings.Split(keys, "+")
	for i, key := range keyParts {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey != "" {
			c.writeRawString(`<span class="key">`, w)
			c.writeRawString(c.escape(trimmedKey), w)
			c.writeRawString(`</span>`, w)
			if i < len(keyParts)-1 {
				c.writeRawString("+", w)
			}
		}
	}
}

// renderMenuPath renders a menu path with proper formatting.
// Splits on " > " or "," and wraps each menu item in a span.menuitem element.
func (c *HTML5Converter) renderMenuPath(path string, w io.Writer) {
	// Try to split on " > " first
	items := strings.Split(path, " > ")
	if len(items) == 1 {
		// Try comma separator
		items = strings.Split(path, ",")
	}
	for i, item := range items {
		trimmedItem := strings.TrimSpace(item)
		if trimmedItem != "" {
			c.writeRawString(`<span class="menuitem">`, w)
			c.writeRawString(c.escape(trimmedItem), w)
			c.writeRawString(`</span>`, w)
			if i < len(items)-1 {
				c.writeRawString(` &#10140; `, w) // Unicode right arrow
			}
		}
	}
}

// renderInlineChildren renders child inline nodes within a parent inline node.
func (c *HTML5Converter) renderInlineChildren(node *inline.Node, w io.Writer) {
	// If there are children, render them; otherwise fall back to rendering text
	if len(node.Children) > 0 {
		lastEnd := 0
		for _, child := range node.Children {
			end := child.Position
			// Write any text before this child node (only if child is after lastEnd)
			if end > lastEnd {
				text := node.Text[lastEnd:end]
				c.writeRawString(c.escape(text), w)
			}
			lastEnd = end
			// Render the child node
			c.convertInlineNode(child, w)
		}
		// Write any remaining text after last child
		if lastEnd > 0 && lastEnd < len(node.Text) {
			text := node.Text[lastEnd:]
			c.writeRawString(c.escape(text), w)
		}
	} else {
		// No children, render the text directly
		c.writeRawString(c.escape(node.Text), w)
	}
}

// writeRawString writes raw text without indentation or newlines.
func (c *HTML5Converter) writeRawString(s string, w io.Writer) {
	fmt.Fprint(w, s)
}

// convertSection converts a section to HTML.
func (c *HTML5Converter) convertSection(section *ast.NodeSection, w io.Writer) {
	// Determine section class based on level (.sect1, .sect2, etc.)
	sectionClass := fmt.Sprintf("sect%d", section.Level)

	// Always add newlines for Asciidoctor compatibility
	fmt.Fprintf(w, `<div class="%s">`+"\n", sectionClass)

	// Determine heading tag based on level
	tag := c.headingTag(section.Level)

	// Write opening heading tag with id attribute if section has an ID
	if section.ID != "" {
		fmt.Fprintf(w, `<%s id="%s">%s</%s>`+"\n", tag, c.escape(section.ID), c.escape(section.Title), tag)
	} else {
		fmt.Fprintf(w, `<%s>%s</%s>`+"\n", tag, c.escape(section.Title), tag)
	}

	// Wrap section content in sectionbody div (only for level 1 sections, Asciidoctor compatibility)
	if section.Level == 1 {
		fmt.Fprintf(w, `<div class="sectionbody">`+"\n")
	}

	// Convert section children (paragraphs, lists, subsections, etc.)
	for _, child := range section.Children {
		c.convertNode(child, w)
	}

	if section.Level == 1 {
		fmt.Fprintf(w, `</div>`+"\n") // Close sectionbody
	}

	fmt.Fprintf(w, `</div>`+"\n") // Close section wrapper (sectN)
}

// headingTag returns appropriate HTML tag for a section level.
// Level 1 maps to h2 (h1 is for document title), level 2 -> h3, etc.
func (c *HTML5Converter) headingTag(level int) string {
	if level >= 1 && level <= 5 {
		return fmt.Sprintf("h%d", level+1)
	}
	return "h6" // Default to h6 for levels >= 6
}

// convertBibliography converts a bibliography section to HTML.
func (c *HTML5Converter) convertBibliography(bib *ast.BibliographyNode, w io.Writer) {
	// Asciidoctor compatibility: bibliography is wrapped in sect1/sectionbody divs
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `<div class="sect1">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	// Write bibliography section heading
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	if bib.ID != "" {
		fmt.Fprintf(w, `<h2 id="%s">`, c.escape(bib.ID))
	} else {
		fmt.Fprint(w, "<h2>")
	}
	fmt.Fprint(w, c.escape(bib.Title))
	fmt.Fprint(w, "</h2>")
	if c.pretty {
		fmt.Fprintln(w)
	}

	// Write sectionbody div
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `<div class="sectionbody">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	// Write ulist bibliography div
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `<div class="ulist bibliography">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	// Write ul.bibliography
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `<ul class="bibliography">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	for _, entry := range bib.Entries {
		c.convertBibliographyEntry(entry, w)
	}

	c.indent = strings.TrimSuffix(c.indent, "  ")
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "</ul>")
	if c.pretty {
		fmt.Fprintln(w)
	}

	c.indent = strings.TrimSuffix(c.indent, "  ")
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "</div>")
	if c.pretty {
		fmt.Fprintln(w)
	}

	c.indent = strings.TrimSuffix(c.indent, "  ")
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "</div>")
	if c.pretty {
		fmt.Fprintln(w)
	}

	c.indent = strings.TrimSuffix(c.indent, "  ")
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "</div>")
	if c.pretty {
		fmt.Fprintln(w)
	}
}

// convertBibliographyEntry converts a single bibliography entry to HTML.
func (c *HTML5Converter) convertBibliographyEntry(entry *ast.BibliographyEntryNode, w io.Writer) {
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}

	// Each entry is an <li> with a <p> inside
	fmt.Fprint(w, "<li>")
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "<p>")

	// Write anchor with id
	if entry.Label != "" {
		fmt.Fprintf(w, `<a id="%s"></a>`, c.escape(entry.Label))
	}

	// Write the citation reference [label] at the start
	// Use XRefText if available, otherwise use Label
	displayLabel := entry.Label
	if entry.XRefText != "" {
		displayLabel = entry.XRefText
	}
	fmt.Fprintf(w, "[%s] ", c.escape(displayLabel))

	// Write the entry text
	if len(entry.InlineNodes) == 0 {
		// Simple text entry
		fmt.Fprint(w, c.escape(entry.Text))
	} else {
		// Entry with inline markup
		lastEnd := 0
		for _, node := range entry.InlineNodes {
			if inlineNode, ok := node.(*inline.Node); ok {
				startPos := inlineNode.StartPos
				// Write any text before this inline node
				if startPos > lastEnd {
					text := entry.Text[lastEnd:startPos]
					fmt.Fprint(w, c.escape(text))
				}
				lastEnd = inlineNode.Position

				// Render the inline node
				c.convertInlineNode(inlineNode, w)
			}
		}

		// Write any remaining text after last inline node
		if lastEnd < len(entry.Text) {
			text := entry.Text[lastEnd:]
			fmt.Fprint(w, c.escape(text))
		}
	}

	fmt.Fprint(w, "</p>")
	if c.pretty {
		fmt.Fprintln(w)
	}

	c.indent = strings.TrimSuffix(c.indent, "  ")
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "</li>")
	if c.pretty {
		fmt.Fprintln(w)
	}
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

	// Determine wrapper class based on list type (Asciidoctor compatibility)
	var wrapperClass string
	switch tag {
	case "ul":
		wrapperClass = "ulist"
	case "ol":
		wrapperClass = "olist arabic" // Add arabic for ordered lists
	case "dl":
		wrapperClass = "dlist"
	}

	// Always add newlines for Asciidoctor compatibility
	fmt.Fprintf(w, `<div class="%s">`+"\n", wrapperClass)

	// For ordered lists, add class attribute
	if tag == "ol" {
		fmt.Fprintf(w, `<ol class="arabic">`+"\n")
	} else {
		fmt.Fprintf(w, `<%s>`+"\n", tag)
	}

	// Convert all list items
	for _, item := range list.Items {
		c.convertNode(item, w)
	}

	fmt.Fprintf(w, `</%s>`+"\n", tag)
	fmt.Fprintf(w, `</div>`+"\n")
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
		// Labeled list: dt and dd (Asciidoctor compatibility)
		fmt.Fprint(w, `<dt class="hdlist1">`)
		fmt.Fprint(w, c.escape(item.Term))
		fmt.Fprint(w, `</dt>
`)
		// Write dd with p-wrapped content
		fmt.Fprint(w, `<dd>
`)
		fmt.Fprint(w, `<p>`)
		fmt.Fprint(w, c.escape(item.Definition))
		fmt.Fprint(w, `</p>
`)
		fmt.Fprint(w, `</dd>
`)
	} else if item.NestedList != nil {
		// Has nested list - open li, render text in p, nested list, close li
		fmt.Fprint(w, `<li>
`)
		fmt.Fprint(w, `<p>`)
		if len(item.InlineNodes) == 0 {
			c.writeRawString(c.escape(item.Text), w)
		} else {
			c.renderInlineText(item.Text, item.InlineNodes, w)
		}
		fmt.Fprint(w, `</p>
`)
		c.convertNode(item.NestedList, w)
		fmt.Fprint(w, `</li>
`)
	} else {
		// Regular list item without nested list - wrap content in p tag (Asciidoctor compatibility)
		fmt.Fprint(w, `<li>
`)
		fmt.Fprint(w, `<p>`)
		if len(item.InlineNodes) == 0 {
			c.writeRawString(c.escape(item.Text), w)
		} else {
			c.renderInlineText(item.Text, item.InlineNodes, w)
		}
		fmt.Fprint(w, `</p>
`)
		fmt.Fprint(w, `</li>
`)
	}
}

// renderInlineText renders text with inline nodes.
func (c *HTML5Converter) renderInlineText(text string, inlineNodes []interface{}, w io.Writer) {
	lastEnd := 0
	for _, node := range inlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			startPos := inlineNode.StartPos
			// Write any text before this inline node
			if startPos > lastEnd {
				plainText := text[lastEnd:startPos]
				c.writeRawString(c.escape(plainText), w)
			}
			lastEnd = inlineNode.Position

			// Render the inline node
			c.convertInlineNode(inlineNode, w)
		}
	}

	// Write any remaining text after last inline node
	if lastEnd < len(text) {
		plainText := text[lastEnd:]
		c.writeRawString(c.escape(plainText), w)
	}
}

// convertLiteral converts a literal block to HTML.
func (c *HTML5Converter) convertLiteral(literal *ast.NodeLiteral, w io.Writer) {
	// Always add newlines for Asciidoctor compatibility
	fmt.Fprintf(w, `<div class="literalblock">`)
	fmt.Fprint(w, "\n")
	fmt.Fprintf(w, `<div class="content">`)
	fmt.Fprint(w, "\n")

	if len(literal.Callouts) == 0 {
		// No callouts, simple rendering
		fmt.Fprintf(w, `<pre>%s</pre>`, c.escape(strings.Join(literal.Lines, "\n")))
		fmt.Fprint(w, "\n")
	} else {
		// Render with callouts
		fmt.Fprintf(w, `<pre>`)

		// Render each line with callout markers
		for lineIdx, line := range literal.Lines {
			// Find callouts on this line
			var lineCallouts []*ast.CalloutNode
			for _, co := range literal.Callouts {
				if co.LineIndex == lineIdx {
					lineCallouts = append(lineCallouts, co)
				}
			}

			// Sort callouts by column position
			// For now, just append them at the end
			fmt.Fprint(w, c.escape(line))

			// Add callout markers at end of line
			for _, co := range lineCallouts {
				fmt.Fprintf(w, ` <b class="conum" data-value="%d"></b>`, co.Number)
			}

			if lineIdx < len(literal.Lines)-1 {
				fmt.Fprint(w, "\n")
			}
		}

		fmt.Fprintf(w, `</pre>`)
		fmt.Fprint(w, "\n")

		// Render callout descriptions
		c.renderCalloutList(literal, w)
	}

	fmt.Fprintf(w, `</div>`)
	fmt.Fprint(w, "\n")
	fmt.Fprintf(w, `</div>`)
	fmt.Fprint(w, "\n")
}

// renderCalloutList renders callout descriptions as an HTML list.
func (c *HTML5Converter) renderCalloutList(literal *ast.NodeLiteral, w io.Writer) {
	// Collect callouts that have descriptions
	var numberedCallouts []*ast.CalloutNode
	maxNumber := 0
	for _, co := range literal.Callouts {
		if co.Description != "" {
			numberedCallouts = append(numberedCallouts, co)
			if co.Number > maxNumber {
				maxNumber = co.Number
			}
		}
	}

	if len(numberedCallouts) == 0 {
		return
	}

	// Create a map for easy lookup
	calloutMap := make(map[int]*ast.CalloutNode)
	for _, co := range numberedCallouts {
		calloutMap[co.Number] = co
	}

	c.writeOpenTagWithClass("div", "colist arabic", w)

	// Render callouts in numerical order
	for i := 1; i <= maxNumber; i++ {
		if co, exists := calloutMap[i]; exists {
			if c.pretty {
				fmt.Fprint(w, c.indent)
			}
			fmt.Fprintf(w, `<div class="li">`)
			fmt.Fprintf(w, `<b class="conum" data-value="%d"></b> `, i)
			fmt.Fprintf(w, `<span>%s</span>`, c.escape(co.Description))
			fmt.Fprintf(w, `</div>`)
			if c.pretty {
				fmt.Fprintln(w)
			}
		}
	}

	c.writeCloseTag("div", w)
}

// convertBlock converts a delimited block to HTML.
func (c *HTML5Converter) convertBlock(block *ast.NodeBlock, w io.Writer) {
	// Determine block class from delimiter (Asciidoctor compatibility)
	class := c.blockClass(block.Delimiter)
	if class == "" {
		class = "listingblock" // Default
	}

	// Write content
	content := strings.Join(block.Lines, "\n")

	if block.Delimiter == "_" {
		// Quote block - special structure
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<blockquote>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="paragraph">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<p>%s</p>`, c.escape(content))
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</blockquote>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
	} else if block.Delimiter == "=" {
		// Example block - wrap in paragraph div
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="content">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="paragraph">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<p>%s</p>`, c.escape(content))
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
	} else if block.Delimiter == "*" {
		// Sidebar block - wrap in paragraph div
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="content">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="paragraph">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<p>%s</p>`, c.escape(content))
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
	} else if block.Delimiter == "/" || block.Delimiter == "-" {
		// Literal/verbatim blocks
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="content">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<pre>%s</pre>`, c.escape(content))
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
	} else {
		// Other blocks - generic structure
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="content">`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `%s`, c.escape(content))
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
	}
}

// convertAdmonition converts an admonition to HTML.
func (c *HTML5Converter) convertAdmonition(admonition *ast.AdmonitionNode, w io.Writer) {
	// Determine class based on admonition kind (Asciidoctor compatibility)
	kind := strings.ToLower(admonition.Kind)
	class := "admonitionblock " + kind

	// Always add newlines for Asciidoctor compatibility (even in embedded mode)
	fmt.Fprintf(w, `<div class="%s">`+"\n", class)
	fmt.Fprintf(w, `<table>`+"\n")
	fmt.Fprintf(w, `<tr>`+"\n")
	fmt.Fprintf(w, `<td class="icon">`+"\n")
	fmt.Fprintf(w, `<div class="title">%s</div>`+"\n", strings.ToUpper(kind[:1])+kind[1:])
	fmt.Fprintf(w, `</td>`+"\n")
	fmt.Fprintf(w, `<td class="content">`+"\n")
	fmt.Fprintf(w, `%s`+"\n", c.escape(admonition.Text))
	fmt.Fprintf(w, `</td>`+"\n")
	fmt.Fprintf(w, `</tr>`+"\n")
	fmt.Fprintf(w, `</table>`+"\n")
	fmt.Fprintf(w, `</div>`+"\n")
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

// convertStyledBlock converts a styled block to HTML.
func (c *HTML5Converter) convertStyledBlock(block *ast.StyledBlockNode, w io.Writer) {
	class := block.Style + "block"

	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, `<div class="%s">`, class)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	// Write content
	content := strings.TrimSpace(block.Content)
	if content != "" {
		c.writeRawString(content, w)
	}

	if c.pretty {
		c.indent = strings.TrimSuffix(c.indent, "  ")
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `</div>`)
	if c.pretty {
		fmt.Fprintln(w)
	}
}

// convertSidebar converts a sidebar block to HTML.
func (c *HTML5Converter) convertSidebar(sidebar *ast.SidebarNode, w io.Writer) {
	c.writeOpenTagWithClass("div", "sidebarblock", w)

	// Write title if present
	if sidebar.Title != "" {
		c.writeElementWithClass("div", "title", c.escape(sidebar.Title), w)
	}

	// Write content
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `<div class="content">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	content := strings.TrimSpace(sidebar.Content)
	if content != "" {
		c.writeRawString(content, w)
	}

	if c.pretty {
		c.indent = strings.TrimSuffix(c.indent, "  ")
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `</div>`)
	if c.pretty {
		fmt.Fprintln(w)
	}

	c.writeCloseTag("div", w)
}

// convertPassThrough converts a passthrough block to HTML.
// Passthrough content is output as-is without escaping.
func (c *HTML5Converter) convertPassThrough(pass *ast.PassThroughNode, w io.Writer) {
	// Passthrough content is written directly without HTML escaping
	c.writeRawString(pass.Content, w)
	if c.pretty && !strings.HasSuffix(pass.Content, "\n") {
		fmt.Fprintln(w)
	}
}

// convertVerse converts a verse block to HTML.
// Verse blocks preserve line breaks and formatting.
func (c *HTML5Converter) convertVerse(verse *ast.VerseNode, w io.Writer) {
	c.writeOpenTagWithClass("div", "verseblock", w)

	// Write content with line breaks preserved
	lines := strings.Split(strings.TrimSpace(verse.Content), "\n")
	for i, line := range lines {
		if c.pretty {
			fmt.Fprint(w, c.indent)
		}
		fmt.Fprint(w, c.escape(line))
		if i < len(lines)-1 {
			fmt.Fprint(w, "<br>")
		}
		if c.pretty {
			fmt.Fprintln(w)
		}
	}

	c.writeCloseTag("div", w)
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
	// Build table classes (Asciidoctor compatibility)
	// Always include frame-all, grid-all, and stretch for basic tables
	classes := []string{"tableblock", "frame-all", "grid-all", "stretch"}

	// Override defaults if explicitly set
	frame := table.GetFrame()
	if frame != ast.FrameAll {
		// Replace frame-all with custom frame
		for i, cls := range classes {
			if cls == "frame-all" {
				classes[i] = "frame-" + string(frame)
				break
			}
		}
	}
	grid := table.GetGrid()
	if grid != ast.GridAll {
		// Replace grid-all with custom grid
		for i, cls := range classes {
			if cls == "grid-all" {
				classes[i] = "grid-" + string(grid)
				break
			}
		}
	}
	if stripes := table.GetStripes(); stripes != "none" {
		classes = append(classes, "stripes-"+stripes)
	}

	// Write opening table tag with classes and attributes
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, "<table")
	fmt.Fprintf(w, ` class="%s"`, strings.Join(classes, " "))
	if table.ID != "" {
		fmt.Fprintf(w, ` id="%s"`, c.escape(table.ID))
	}
	if width := table.GetWidth(); width != "" {
		fmt.Fprintf(w, ` style="width:%s"`, c.escape(width))
	}
	fmt.Fprint(w, ">")
	if c.pretty {
		fmt.Fprintln(w)
	}

	// Write colgroup for column widths (Asciidoctor compatibility)
	c.writeOpenTag("colgroup", w)
	numCols := table.ColumnCount()
	if numCols > 0 {
		colWidth := 100.0 / float64(numCols)
		for i := 0; i < numCols; i++ {
			// Last column may have slightly different width for rounding
			width := colWidth
			if i == numCols-1 {
				width = 100.0 - (colWidth*float64(numCols-1)) + 0.0001
			}
			if c.pretty {
				fmt.Fprint(w, c.indent)
			}
			fmt.Fprintf(w, `<col style="width: %.4f%%;">`, width)
			if c.pretty {
				fmt.Fprintln(w)
			}
		}
	}
	c.writeCloseTag("colgroup", w)

	// Write caption if present
	if table.Caption != "" {
		c.writeElement("caption", c.escape(table.Caption), w)
	}

	// For basic tables, all rows go in tbody
	// (Header row detection is disabled for basic AsciiDoc compatibility)
	c.writeOpenTag("tbody", w)
	for _, row := range table.Rows {
		c.writeTableRow(&row, "td", w)
	}
	c.writeCloseTag("tbody", w)

	c.writeCloseTag("table", w)
}

// writeTableRow writes a single table row.
func (c *HTML5Converter) writeTableRow(row *ast.TableRow, cellTag string, w io.Writer) {
	if row == nil {
		return
	}

	c.writeOpenTag("tr", w)

	for _, cell := range row.Cells {
		c.writeTableCell(&cell, cellTag, w)
	}

	c.writeCloseTag("tr", w)
}

// writeTableCell writes a single table cell.
func (c *HTML5Converter) writeTableCell(cell *ast.TableCell, tag string, w io.Writer) {
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}

	fmt.Fprint(w, "<"+tag)

	// Add tableblock class and alignment classes (Asciidoctor compatibility)
	fmt.Fprint(w, ` class="tableblock`)

	// Add horizontal alignment class
	align := cell.HorizontalAlign
	if align == "" && tag == "th" {
		align = "center" // Default for headers
	}
	switch align {
	case "left":
		fmt.Fprint(w, ` halign-left`)
	case "right":
		fmt.Fprint(w, ` halign-right`)
	case "center":
		fmt.Fprint(w, ` halign-center`)
	default:
		fmt.Fprint(w, ` halign-left`)
	}

	// Add vertical alignment class
	valign := cell.VerticalAlign
	if valign == "" {
		valign = "top" // Default
	}
	switch valign {
	case "top":
		fmt.Fprint(w, ` valign-top`)
	case "middle":
		fmt.Fprint(w, ` valign-middle`)
	case "bottom":
		fmt.Fprint(w, ` valign-bottom`)
	default:
		fmt.Fprint(w, ` valign-top`)
	}

	fmt.Fprint(w, `"`)

	// Add colspan if > 1
	if cell.ColSpan > 1 {
		fmt.Fprintf(w, ` colspan="%d"`, cell.ColSpan)
	}

	// Add rowspan if > 1
	if cell.RowSpan > 1 {
		fmt.Fprintf(w, ` rowspan="%d"`, cell.RowSpan)
	}

	fmt.Fprint(w, ">")

	// Write cell content wrapped in p.tableblock (Asciidoctor compatibility)
	fmt.Fprint(w, `<p class="tableblock">`)

	if len(cell.InlineNodes) == 0 {
		c.writeRawString(c.escape(cell.Text), w)
	} else {
		// Render inline nodes
		lastEnd := 0
		for _, node := range cell.InlineNodes {
			if inlineNode, ok := node.(*inline.Node); ok {
				startPos := inlineNode.StartPos
				// Write any text before this inline node
				if startPos > lastEnd {
					text := cell.Text[lastEnd:startPos]
					c.writeRawString(c.escape(text), w)
				}
				lastEnd = inlineNode.Position
				// Render the inline node
				c.convertInlineNode(inlineNode, w)
			}
		}
		// Write any remaining text
		if lastEnd < len(cell.Text) {
			text := cell.Text[lastEnd:]
			c.writeRawString(c.escape(text), w)
		}
	}

	fmt.Fprint(w, `</p>`)
	fmt.Fprintf(w, "</%s>", tag)
	if c.pretty {
		fmt.Fprintln(w)
	}
}
