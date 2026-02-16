// Package converter provides EPUB converter for AsciiDoc AST.
package converter

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
)

// EPUBConverter converts an AsciiDoc AST to EPUB format.
type EPUBConverter struct {
	// book title
	title string

	// book author
	author string

	// book language
	language string

	// book identifier (UUID)
	identifier string

	// publisher
	publisher string

	// cover image
	coverImage string

	// css for styling
	customCSS string
}

// NewEPUBConverter creates a new EPUB converter.
func NewEPUBConverter() *EPUBConverter {
	return &EPUBConverter{
		title:     "Untitled",
		language:  "en",
		identifier: generateID(),
	}
}

// generateID generates a unique identifier for the EPUB.
func generateID() string {
	return fmt.Sprintf("urn:uuid:%d", time.Now().UnixNano())
}

// WithTitle sets the book title.
func (c *EPUBConverter) WithTitle(title string) *EPUBConverter {
	c.title = title
	return c
}

// WithAuthor sets the book author.
func (c *EPUBConverter) WithAuthor(author string) *EPUBConverter {
	c.author = author
	return c
}

// WithLanguage sets the book language.
func (c *EPUBConverter) WithLanguage(lang string) *EPUBConverter {
	c.language = lang
	return c
}

// WithIdentifier sets the book identifier.
func (c *EPUBConverter) WithIdentifier(id string) *EPUBConverter {
	c.identifier = id
	return c
}

// WithPublisher sets the publisher.
func (c *EPUBConverter) WithPublisher(publisher string) *EPUBConverter {
	c.publisher = publisher
	return c
}

// WithCustomCSS adds custom CSS for styling.
func (c *EPUBConverter) WithCustomCSS(css string) *EPUBConverter {
	c.customCSS = css
	return c
}

// Convert converts document to EPUB format (ZIP archive).
func (c *EPUBConverter) Convert(doc *ast.NodeDocument, w io.Writer) error {
	// Create ZIP writer
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Extract metadata from document
	if doc.Header != nil {
		if doc.Header.Title != "" {
			c.title = doc.Header.Title
		}
		if doc.Header.Author != "" {
			c.author = doc.Header.Author
		}
	}

	// Override with attributes if present
	if title, ok := doc.Attributes["title"]; ok {
		c.title = title
	}
	if author, ok := doc.Attributes["author"]; ok {
		c.author = author
	}

	// 1. Write mimetype (must be first, uncompressed)
	mimetype := "application/epub+zip"
	err := c.writeUncompressedFile(zipWriter, "mimetype", mimetype)
	if err != nil {
		return fmt.Errorf("failed to write mimetype: %w", err)
	}

	// 2. Write META-INF/container.xml
	containerXML := c.generateContainerXML()
	err = c.writeTextFile(zipWriter, "META-INF/container.xml", containerXML)
	if err != nil {
		return fmt.Errorf("failed to write container.xml: %w", err)
	}

	// 3. Write OEBPS/content.opf (metadata and manifest)
	contentOPF := c.generateContentOPF(doc)
	err = c.writeTextFile(zipWriter, "OEBPS/content.opf", contentOPF)
	if err != nil {
		return fmt.Errorf("failed to write content.opf: %w", err)
	}

	// 4. Write OEBPS/toc.ncx (navigation)
	tocNCX := c.generateTOCNCX(doc)
	err = c.writeTextFile(zipWriter, "OEBPS/toc.ncx", tocNCX)
	if err != nil {
		return fmt.Errorf("failed to write toc.ncx: %w", err)
	}

	// 5. Write OEBPS/stylesheet.css
	css := c.getEPUBCSS()
	if c.customCSS != "" {
		css += "\n" + c.customCSS
	}
	err = c.writeTextFile(zipWriter, "OEBPS/stylesheet.css", css)
	if err != nil {
		return fmt.Errorf("failed to write stylesheet.css: %w", err)
	}

	// 6. Write content as XHTML
	chapterIndex := 0
	var tocEntries []tocEntry

	// Generate table of contents from sections
	for _, block := range doc.Blocks {
		if section, ok := block.(*ast.NodeSection); ok {
			tocEntries = append(tocEntries, tocEntry{
				ID:    sectionID(section, chapterIndex),
				Title: section.Title,
				Order: chapterIndex + 1,
			})
			chapterIndex++
		}
	}

	// Write content XHTML
	chapterIndex = 0
	contentXHTML := c.generateContentXHTML(doc)
	err = c.writeTextFile(zipWriter, "OEBPS/content.xhtml", contentXHTML)
	if err != nil {
		return fmt.Errorf("failed to write content.xhtml: %w", err)
	}

	// Close ZIP writer
	return zipWriter.Close()
}

// writeUncompressedFile writes a file to the ZIP without compression (required for mimetype).
func (c *EPUBConverter) writeUncompressedFile(zipWriter *zip.Writer, path, content string) error {
	writer, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   path,
		Method: zip.Store, // No compression
	})
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(content))
	return err
}

// writeTextFile writes a text file to the ZIP with compression.
func (c *EPUBConverter) writeTextFile(zipWriter *zip.Writer, path, content string) error {
	writer, err := zipWriter.Create(path)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(content))
	return err
}

// generateContainerXML generates the META-INF/container.xml file.
func (c *EPUBConverter) generateContainerXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf"/>
  </rootfiles>
</container>`
}

// generateContentOPF generates the OEBPS/content.opf file (metadata and manifest).
func (c *EPUBConverter) generateContentOPF(doc *ast.NodeDocument) string {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookID" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:identifier id="BookID">`)
	buf.WriteString(c.identifier)
	buf.WriteString(`</dc:identifier>
    <dc:title>`)
	buf.WriteString(escapeXML(c.title))
	buf.WriteString(`</dc:title>
    <dc:language>`)
	buf.WriteString(c.language)
	buf.WriteString(`</dc:language>
    <dc:date>`)
	buf.WriteString(time.Now().Format("2006-01-02"))
	buf.WriteString(`</dc:date>`)

	if c.author != "" {
		buf.WriteString(`    <dc:creator>`)
		buf.WriteString(escapeXML(c.author))
		buf.WriteString(`</dc:creator>
`)
	}

	if c.publisher != "" {
		buf.WriteString(`    <dc:publisher>`)
		buf.WriteString(escapeXML(c.publisher))
		buf.WriteString(`</dc:publisher>
`)
	}

	buf.WriteString(`  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="content" href="content.xhtml" media-type="application/xhtml+xml"/>
    <item id="stylesheet" href="stylesheet.css" media-type="text/css"/>
  </manifest>
  <spine>
    <itemref idref="content"/>
  </spine>
</package>`)

	return buf.String()
}

// tocEntry represents an entry in the table of contents.
type tocEntry struct {
	ID    string
	Title string
	Order int
}

// generateTOCNCX generates the OEBPS/toc.ncx file (navigation).
func (c *EPUBConverter) generateTOCNCX(doc *ast.NodeDocument) string {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="`)
	buf.WriteString(c.identifier)
	buf.WriteString(`"/>
  </head>
  <docTitle>
    <text>`)
	buf.WriteString(escapeXML(c.title))
	buf.WriteString(`</text>
  </docTitle>
  <navMap>
    <navPoint id="navpoint-1" playOrder="1">
      <navLabel>
        <text>`)
	buf.WriteString(escapeXML(c.title))
	buf.WriteString(`</text>
      </navLabel>
      <content src="content.xhtml"/>
    </navPoint>
  </navMap>
</ncx>`)

	return buf.String()
}

// generateContentXHTML generates the OEBPS/content.xhtml file (main content).
func (c *EPUBConverter) generateContentXHTML(doc *ast.NodeDocument) string {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>`)
	buf.WriteString(escapeXML(c.title))
	buf.WriteString(`</title>
  <link rel="stylesheet" type="text/css" href="stylesheet.css"/>
</head>
<body>
`)

	// Write document title as heading
	if c.title != "" {
		buf.WriteString("<h1>")
		buf.WriteString(escapeXML(c.title))
		buf.WriteString("</h1>\n\n")
	}

	// Convert all blocks
	for _, block := range doc.Blocks {
		c.convertNode(block, &buf)
	}

	buf.WriteString(`</body>
</html>`)

	return buf.String()
}

// convertNode converts an AST node to XHTML.
func (c *EPUBConverter) convertNode(node ast.Node, buf *bytes.Buffer) {
	switch n := node.(type) {
	case *ast.NodeParagraph:
		c.convertParagraph(n, buf)
	case *ast.NodeSection:
		c.convertSection(n, buf)
	case *ast.NodeList:
		c.convertList(n, buf)
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
	case *ast.StyledBlockNode:
		c.convertStyledBlock(n, buf)
	case *ast.SidebarNode:
		c.convertSidebar(n, buf)
	case *ast.PassThroughNode:
		c.convertPassThrough(n, buf)
	case *ast.VerseNode:
		c.convertVerse(n, buf)
	}
}

// convertParagraph converts a paragraph to XHTML.
func (c *EPUBConverter) convertParagraph(para *ast.NodeParagraph, buf *bytes.Buffer) {
	buf.WriteString("<p>")

	if len(para.InlineNodes) == 0 {
		buf.WriteString(escapeXML(para.Text))
	} else {
		lastEnd := 0
		for _, node := range para.InlineNodes {
			if inlineNode, ok := node.(*inline.Node); ok {
				startPos := inlineNode.StartPos
				if startPos > lastEnd {
					text := para.Text[lastEnd:startPos]
					buf.WriteString(escapeXML(text))
				}
				lastEnd = inlineNode.Position
				c.convertInlineNode(inlineNode, buf)
			}
		}
		if lastEnd < len(para.Text) {
			text := para.Text[lastEnd:]
			buf.WriteString(escapeXML(text))
		}
	}

	buf.WriteString("</p>\n")
}

// convertInlineNode converts an inline node to XHTML.
func (c *EPUBConverter) convertInlineNode(node *inline.Node, buf *bytes.Buffer) {
	switch node.Type {
	case inline.NodeText:
		buf.WriteString(escapeXML(node.Text))
	case inline.NodeBold:
		buf.WriteString("<strong>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(escapeXML(node.Text))
		}
		buf.WriteString("</strong>")
	case inline.NodeItalic:
		buf.WriteString("<em>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(escapeXML(node.Text))
		}
		buf.WriteString("</em>")
	case inline.NodeMonospace:
		buf.WriteString("<code>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(escapeXML(node.Text))
		}
		buf.WriteString("</code>")
	case inline.NodeSubscript:
		buf.WriteString("<sub>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(escapeXML(node.Text))
		}
		buf.WriteString("</sub>")
	case inline.NodeSuperscript:
		buf.WriteString("<sup>")
		if len(node.Children) > 0 {
			c.renderInlineChildren(node, buf)
		} else {
			buf.WriteString(escapeXML(node.Text))
		}
		buf.WriteString("</sup>")
	case inline.NodeLink:
		if node.URL != "" {
			buf.WriteString(fmt.Sprintf("<a href=\"%s\">", escapeXMLAttr(node.URL)))
		} else {
			buf.WriteString("<a>")
		}
		c.renderInlineChildren(node, buf)
		buf.WriteString("</a>")
	case inline.NodeImage:
		alt := node.Alt
		if alt == "" {
			alt = node.Text
		}
		buf.WriteString(fmt.Sprintf("<img src=\"%s\" alt=\"%s\"/>", escapeXMLAttr(node.URL), escapeXMLAttr(alt)))
	case inline.NodeCrossRef:
		buf.WriteString(fmt.Sprintf("<a href=\"#%s\">", escapeXMLAttr(node.Ref)))
		if node.RefText != "" {
			buf.WriteString(escapeXML(node.RefText))
		} else {
			c.renderInlineChildren(node, buf)
		}
		buf.WriteString("</a>")
	}
}

// renderInlineChildren renders child inline nodes.
func (c *EPUBConverter) renderInlineChildren(node *inline.Node, buf *bytes.Buffer) {
	if len(node.Children) > 0 {
		lastEnd := 0
		for _, child := range node.Children {
			end := child.Position
			if end > lastEnd {
				text := node.Text[lastEnd:end]
				buf.WriteString(escapeXML(text))
			}
			lastEnd = end
			c.convertInlineNode(child, buf)
		}
		if lastEnd < len(node.Text) {
			text := node.Text[lastEnd:]
			buf.WriteString(escapeXML(text))
		}
	} else {
		buf.WriteString(escapeXML(node.Text))
	}
}

// convertSection converts a section to XHTML.
func (c *EPUBConverter) convertSection(section *ast.NodeSection, buf *bytes.Buffer) {
	level := section.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	tag := fmt.Sprintf("h%d", level)
	buf.WriteString(fmt.Sprintf("<%s>", tag))
	buf.WriteString(escapeXML(section.Title))
	buf.WriteString(fmt.Sprintf("</%s>\n", tag))

	for _, child := range section.Children {
		c.convertNode(child, buf)
	}
}

// convertList converts a list to XHTML.
func (c *EPUBConverter) convertList(list *ast.NodeList, buf *bytes.Buffer) {
	if len(list.Items) == 0 {
		return
	}

	item := list.Items[0]
	listItem, ok := item.(*ast.NodeListItem)
	if !ok {
		return
	}

	switch listItem.Marker {
	case "-", "*", "o":
		buf.WriteString("<ul>\n")
		for _, item := range list.Items {
			if li, ok := item.(*ast.NodeListItem); ok {
				buf.WriteString("<li>")
				if li.Text != "" || len(li.InlineNodes) > 0 {
					c.renderInlineText(li.Text, li.InlineNodes, buf)
				}
				if li.NestedList != nil {
					buf.WriteString("\n")
					c.convertNode(li.NestedList, buf)
				}
				buf.WriteString("</li>\n")
			}
		}
		buf.WriteString("</ul>\n")
	case ".":
		buf.WriteString("<ol>\n")
		for _, item := range list.Items {
			if li, ok := item.(*ast.NodeListItem); ok {
				buf.WriteString("<li>")
				if li.Text != "" || len(li.InlineNodes) > 0 {
					c.renderInlineText(li.Text, li.InlineNodes, buf)
				}
				if li.NestedList != nil {
					buf.WriteString("\n")
					c.convertNode(li.NestedList, buf)
				}
				buf.WriteString("</li>\n")
			}
		}
		buf.WriteString("</ol>\n")
	case "::":
		buf.WriteString("<dl>\n")
		for _, item := range list.Items {
			if li, ok := item.(*ast.NodeListItem); ok {
				buf.WriteString("<dt>")
				buf.WriteString(escapeXML(li.Term))
				buf.WriteString("</dt>\n")
				buf.WriteString("<dd>")
				if li.Text != "" || len(li.InlineNodes) > 0 {
					c.renderInlineText(li.Text, li.InlineNodes, buf)
				}
				buf.WriteString("</dd>\n")
			}
		}
		buf.WriteString("</dl>\n")
	}
}

// renderInlineText renders text with inline nodes.
func (c *EPUBConverter) renderInlineText(text string, inlineNodes []interface{}, buf *bytes.Buffer) {
	lastEnd := 0
	for _, node := range inlineNodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			startPos := inlineNode.StartPos
			if startPos > lastEnd {
				plainText := text[lastEnd:startPos]
				buf.WriteString(escapeXML(plainText))
			}
			lastEnd = inlineNode.Position
			c.convertInlineNode(inlineNode, buf)
		}
	}
	if lastEnd < len(text) {
		plainText := text[lastEnd:]
		buf.WriteString(escapeXML(plainText))
	}
}

// convertLiteral converts a literal block to XHTML.
func (c *EPUBConverter) convertLiteral(literal *ast.NodeLiteral, buf *bytes.Buffer) {
	buf.WriteString("<pre><code>")
	for _, line := range literal.Lines {
		buf.WriteString(escapeXML(line))
		buf.WriteString("\n")
	}
	buf.WriteString("</code></pre>\n")
}

// convertBlock converts a delimited block to XHTML.
func (c *EPUBConverter) convertBlock(block *ast.NodeBlock, buf *bytes.Buffer) {
	switch block.Delimiter {
	case "-":
		// Verbatim block
		buf.WriteString("<pre><code>")
		for _, line := range block.Lines {
			buf.WriteString(escapeXML(line))
			buf.WriteString("\n")
		}
		buf.WriteString("</code></pre>\n")
	case "_":
		// Quote block
		buf.WriteString("<blockquote><p>")
		buf.WriteString(escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</p></blockquote>\n")
	case "=":
		// Example block
		buf.WriteString("<div class=\"exampleblock\">")
		buf.WriteString("<pre><code>")
		for _, line := range block.Lines {
			buf.WriteString(escapeXML(line))
			buf.WriteString("\n")
		}
		buf.WriteString("</code></pre>")
		buf.WriteString("</div>\n")
	default:
		// Default: treat as para
		buf.WriteString("<p>")
		buf.WriteString(escapeXML(strings.Join(block.Lines, "\n")))
		buf.WriteString("</p>\n")
	}
}

// convertAdmonition converts an admonition to XHTML.
func (c *EPUBConverter) convertAdmonition(admonition *ast.AdmonitionNode, buf *bytes.Buffer) {
	class := "note"
	switch strings.ToLower(admonition.Kind) {
	case "tip":
		class = "tip"
	case "warning":
		class = "warning"
	case "caution":
		class = "caution"
	case "important":
		class = "important"
	}

	buf.WriteString(fmt.Sprintf("<div class=\"%s\">\n", class))
	buf.WriteString("<p>")
	buf.WriteString(escapeXML(admonition.Text))
	buf.WriteString("</p>")
	buf.WriteString("</div>\n")
}

// convertMacro converts a block macro to XHTML.
func (c *EPUBConverter) convertMacro(macro *ast.MacroNode, buf *bytes.Buffer) {
	switch macro.Target {
	case "image":
		buf.WriteString(fmt.Sprintf("<div class=\"imageblock\"><img src=\"%s\" alt=\"\"/></div>\n", escapeXMLAttr(macro.Path)))
	default:
		// Unknown macro
		buf.WriteString(fmt.Sprintf("<!-- %s::%s -->\n", macro.Target, macro.Path))
	}
}

// convertStyledBlock converts a styled block to XHTML.
func (c *EPUBConverter) convertStyledBlock(block *ast.StyledBlockNode, buf *bytes.Buffer) {
	class := block.Style + "block"
	buf.WriteString(fmt.Sprintf("<div class=\"%s\">%s</div>\n", class, escapeXML(block.Content)))
}

// convertSidebar converts a sidebar block to XHTML.
func (c *EPUBConverter) convertSidebar(sidebar *ast.SidebarNode, buf *bytes.Buffer) {
	buf.WriteString("<div class=\"sidebar\">\n")

	// Write title if present
	if sidebar.Title != "" {
		buf.WriteString(fmt.Sprintf("<div class=\"title\">%s</div>\n", escapeXML(sidebar.Title)))
	}

	// Write content
	buf.WriteString("<div class=\"content\">")
	buf.WriteString(escapeXML(sidebar.Content))
	buf.WriteString("</div>\n")
	buf.WriteString("</div>\n")
}

// convertPassThrough converts a passthrough block to XHTML.
// Passthrough content is output as-is without escaping.
func (c *EPUBConverter) convertPassThrough(pass *ast.PassThroughNode, buf *bytes.Buffer) {
	// Passthrough content is written directly without XML escaping
	buf.WriteString(pass.Content)
	if !strings.HasSuffix(pass.Content, "\n") {
		buf.WriteString("\n")
	}
}

// convertVerse converts a verse block to XHTML.
// Verse blocks preserve line breaks and formatting.
func (c *EPUBConverter) convertVerse(verse *ast.VerseNode, buf *bytes.Buffer) {
	buf.WriteString("<div class=\"verseblock\">\n")

	// Write content with line breaks preserved
	lines := strings.Split(strings.TrimSpace(verse.Content), "\n")
	for i, line := range lines {
		buf.WriteString(escapeXML(line))
		if i < len(lines)-1 {
			buf.WriteString("<br/>")
		}
		buf.WriteString("\n")
	}

	buf.WriteString("</div>\n")
}

// convertTable converts a table to XHTML.
func (c *EPUBConverter) convertTable(table *ast.Table, buf *bytes.Buffer) {
	// Build table classes
	classes := []string{"table"}
	frame := table.GetFrame()
	grid := table.GetGrid()

	if frame != ast.FrameAll {
		classes = append(classes, "frame-"+string(frame))
	}
	if grid != ast.GridAll {
		classes = append(classes, "grid-"+string(grid))
	}
	if stripes := table.GetStripes(); stripes != "none" {
		classes = append(classes, "stripes-"+stripes)
	}

	// Write opening table tag
	buf.WriteString(fmt.Sprintf("<table class=\"%s\">\n", strings.Join(classes, " ")))

	// Write caption if present
	if table.Caption != "" {
		buf.WriteString(fmt.Sprintf("<caption>%s</caption>\n", escapeXML(table.Caption)))
	}

	// Write header if present
	if table.HasHeader() {
		headerRow := table.HeaderRow()
		buf.WriteString("<thead><tr>\n")
		for _, cell := range headerRow.Cells {
			c.writeEPUBTableCell(&cell, "th", buf)
		}
		buf.WriteString("</tr></thead>\n")
	}

	// Write body rows
	bodyRows := table.BodyRows()
	if len(bodyRows) > 0 {
		buf.WriteString("<tbody>\n")
		for _, row := range bodyRows {
			buf.WriteString("<tr>\n")
			for _, cell := range row.Cells {
				c.writeEPUBTableCell(&cell, "td", buf)
			}
			buf.WriteString("</tr>\n")
		}
		buf.WriteString("</tbody>\n")
	}

	// Write footer if present
	if table.HasFooter() {
		footerRow := table.FooterRow()
		buf.WriteString("<tfoot><tr>\n")
		for _, cell := range footerRow.Cells {
			c.writeEPUBTableCell(&cell, "td", buf)
		}
		buf.WriteString("</tr></tfoot>\n")
	}

	buf.WriteString("</table>\n")
}

// writeEPUBTableCell writes a single table cell in XHTML format.
func (c *EPUBConverter) writeEPUBTableCell(cell *ast.TableCell, tag string, buf *bytes.Buffer) {
	buf.WriteString("<" + tag)

	// Add colspan if > 1
	if cell.ColSpan > 1 {
		buf.WriteString(fmt.Sprintf(" colspan=\"%d\"", cell.ColSpan))
	}

	// Add rowspan if > 1
	if cell.RowSpan > 1 {
		buf.WriteString(fmt.Sprintf(" rowspan=\"%d\"", cell.RowSpan))
	}

	// Add alignment
	if cell.HorizontalAlign != "" {
		var cssAlign string
		switch cell.HorizontalAlign {
		case "left":
			cssAlign = "left"
		case "right":
			cssAlign = "right"
		case "center":
			cssAlign = "center"
		default:
			cssAlign = "left"
		}
		buf.WriteString(fmt.Sprintf(" style=\"text-align:%s\"", cssAlign))
	}

	buf.WriteString(">")

	// Write cell content
	if len(cell.InlineNodes) == 0 {
		buf.WriteString(escapeXML(cell.Text))
	} else {
		// Render inline nodes
		lastEnd := 0
		for _, node := range cell.InlineNodes {
			if inlineNode, ok := node.(*inline.Node); ok {
				startPos := inlineNode.StartPos
				// Write any text before this inline node
				if startPos > lastEnd {
					text := cell.Text[lastEnd:startPos]
					buf.WriteString(escapeXML(text))
				}
				lastEnd = inlineNode.Position
				// Render the inline node
				c.convertInlineNode(inlineNode, buf)
			}
		}
		// Write any remaining text
		if lastEnd < len(cell.Text) {
			text := cell.Text[lastEnd:]
			buf.WriteString(escapeXML(text))
		}
	}

	buf.WriteString(fmt.Sprintf("</%s>\n", tag))
}

// sectionID generates an ID for a section.
func sectionID(section *ast.NodeSection, index int) string {
	if section.ID != "" {
		return section.ID
	}
	return fmt.Sprintf("section-%d", index)
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// escapeXMLAttr escapes special XML characters for attributes.
func escapeXMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// getEPUBCSS returns CSS for EPUB styling.
func (c *EPUBConverter) getEPUBCSS() string {
	return `
body {
  font-family: serif;
  line-height: 1.6;
  margin: 1em;
}

h1, h2, h3, h4, h5, h6 {
  font-family: sans-serif;
  font-weight: bold;
  margin-top: 1.5em;
  margin-bottom: 0.5em;
}

h1 {
  font-size: 2em;
  border-bottom: 2px solid #ccc;
  padding-bottom: 0.3em;
}

h2 {
  font-size: 1.5em;
  border-bottom: 1px solid #eee;
  padding-bottom: 0.2em;
}

h3 {
  font-size: 1.25em;
}

p {
  margin: 0.5em 0;
  text-align: justify;
}

a {
  color: #0066cc;
  text-decoration: none;
}

code {
  font-family: monospace;
  background-color: #f5f5f5;
  padding: 0.1em 0.3em;
  font-size: 0.9em;
}

pre {
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  padding: 1em;
  overflow-x: auto;
}

pre code {
  background-color: transparent;
  padding: 0;
  border: none;
}

ul, ol {
  padding-left: 2em;
}

li {
  margin: 0.3em 0;
}

dl dt {
  font-weight: bold;
}

dl dd {
  margin-left: 2em;
}

table {
  border-collapse: collapse;
  width: 100%;
  margin: 1em 0;
}

th, td {
  border: 1px solid #ddd;
  padding: 0.5em;
  text-align: left;
}

th {
  background-color: #f5f5f5;
  font-weight: bold;
}

blockquote {
  border-left: 4px solid #ddd;
  margin: 1em 0;
  padding-left: 1em;
  color: #666;
  font-style: italic;
}

.exampleblock {
  margin: 1em 0;
  padding: 1em;
  background-color: #f9f9f9;
  border: 1px solid #ddd;
}

.note, .tip, .warning, .caution, .important {
  margin: 1em 0;
  padding: 0.8em 1em;
  border-left: 4px solid;
  background-color: #f9f9f9;
}

.note { border-left-color: #0066cc; }
.tip { border-left-color: #009900; }
.warning { border-left-color: #ff9900; }
.caution { border-left-color: #cc0000; }
.important { border-left-color: #000; }
`
}
