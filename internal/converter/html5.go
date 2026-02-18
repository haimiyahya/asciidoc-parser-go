// Package converter provides HTML5 converter for AsciiDoc AST.
package converter

import (
	"fmt"
	htmlstd "html"
	"io"
	"math"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/extension"
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

	// extensionRegistry holds custom macro processors
	extensionRegistry *extension.Registry
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

// WithExtensionRegistry sets an extension registry for custom macros.
func (c *HTML5Converter) WithExtensionRegistry(registry *extension.Registry) *HTML5Converter {
	c.extensionRegistry = registry
	return c
}

// escape escapes HTML special characters.
func (c *HTML5Converter) escape(s string) string {
	return htmlstd.EscapeString(s)
}

// writeCSS writes CSS styles to the output.
func (c *HTML5Converter) writeCSS(w io.Writer) {
	css := `
<style>
body {
	font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
	line-height: 1.6;
	max-width: 900px;
	margin: 0 auto;
	padding: 1rem;
	color: #333;
}
h1, h2, h3, h4, h5, h6 {
	margin-top: 1.5em;
	margin-bottom: 0.5em;
	font-weight: 600;
}
p {
	margin: 0.5em 0;
}
ul, ol {
	margin: 0.5em 0;
	padding-left: 2em;
}
code {
	background-color: #f4f4f4;
	padding: 0.2em 0.4em;
	border-radius: 3px;
	font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
	font-size: 0.9em;
}
.listingblock, .literalblock {
	background-color: #f4f4f4;
	border: 1px solid #ddd;
	border-radius: 4px;
	padding: 1em;
	margin: 1em 0;
	overflow-x: auto;
}
.listingblock pre, .literalblock pre {
	margin: 0;
	padding: 0;
}
.tableblock {
	margin: 1em 0;
}
table.tableblock {
	border-collapse: collapse;
	width: 100%;
}
table.tableblock th,
table.tableblock td {
	border: 1px solid #ddd;
	padding: 0.5em 0.75em;
	text-align: left;
}
table.tableblock thead th {
	background-color: #f2f2f2;
	font-weight: 600;
	text-align: center;
}
table.tableblock.frame-all {
	border: 1px solid #ddd;
}
table.tableblock.grid-all td,
table.tableblock.grid-all th {
	border-left: 1px solid #ddd;
}
table.tableblock.grid-all td:first-child,
table.tableblock.grid-all th:first-child {
	border-left: none;
}
.admonitionblock {
	margin: 1em 0;
	padding: 1em;
	border-left: 4px solid #4CAF50;
	background-color: #f8f8f8;
}
.admonitionblock.note {
	border-left-color: #2196F3;
}
.admonitionblock.tip {
	border-left-color: #4CAF50;
}
.admonitionblock.warning {
	border-left-color: #FF9800;
}
.admonitionblock.caution {
	border-left-color: #FF5722;
}
.admonitionblock.important {
	border-left-color: #9C27B0;
}
.admonitionblock td.icon {
	padding-right: 1em;
	font-weight: bold;
}
.quoteblock {
	margin: 1em 0;
	padding-left: 1em;
	border-left: 4px solid #ddd;
	color: #555;
}
.quoteblock blockquote {
	margin: 0;
}
.sidebarblock {
	background-color: #f0f7ff;
	border-left: 4px solid #2196F3;
	border-radius: 4px;
	padding: 1em;
	margin: 1.5em 0;
}
.exampleblock {
	background-color: #f8f8f8;
	border: 1px solid #ddd;
	border-radius: 4px;
	padding: 1em;
	margin: 1em 0;
}
.colist {
	margin-top: 1em;
	margin-bottom: 1em;
}
.colist.arabic {
	list-style-type: none;
}
.colist.arabic .li {
	padding-left: 2em;
	position: relative;
}
.colist.arabic .conum {
	display: inline-block;
	width: 1.8em;
	height: 1.8em;
	line-height: 1.8em;
	text-align: center;
	border-radius: 50%;
	background-color: #2196F3;
	color: white;
	font-weight: bold;
	font-size: 0.9em;
	position: absolute;
	left: 0;
}
.colist.arabic .conum:before {
	content: attr(data-value);
}
.sect1 {
	margin-bottom: 2em;
}
.sect2 {
	margin-top: 1.5em;
}
a {
	color: #2196F3;
	text-decoration: none;
}
a:hover {
	text-decoration: underline;
}
strong, b {
	font-weight: 600;
}
em, i {
	font-style: italic;
}
/* Chroma Syntax Highlighting (GitHub theme) */
.highlight {
	margin: 1em 0;
}
.highlight pre {
	background-color: #f6f8fa;
	border-radius: 6px;
	padding: 16px;
	overflow: auto;
	font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
	font-size: 13px;
	line-height: 1.45;
}
.highlight code {
	background: none;
	padding: 0;
	font-family: inherit;
}
/* Chroma token colors (GitHub theme) */
.chroma . { color: #24292e; } /* Plain text */
.chroma .k { color: #cf222e; } /* Keyword */
.chroma .kd { color: #cf222e; } /* Keyword declaration */
.chroma .kn { color: #cf222e; } /* Keyword namespace */
.chroma .kt { color: #cf222e; } /* Keyword type */
.chroma .n { color: #6f42c1; } /* Name */
.chroma .na { color: #6f42c1; } /* Name attribute */
.chroma .nb { color: #959da5; } /* Name builtin */
.chroma .bp { color: #959da5; } /* Name pseudo builtin */
.chroma .nc { color: #6f42c1; } /* Name class */
.chroma .no { color: #cf222e; } /* Name constant */
.chroma .nd { color: #6f42c1; } /* Name decorator */
.chroma .ni { color: #0550ae; } /* Name entity */
.chroma .ne { color: #cf222e; } /* Name exception */
.chroma .nf { color: #8250df; } /* Name function */
.chroma .nl { color: #959da5; } /* Name label */
.chroma .nn { color: #6f42c1; } /* Name namespace */
.chroma .nx { color: #24292e; } /* Name other */
.chroma .py { color: #959da5; } /* Name property */
.chroma .nt { color: #22863a; } /* Name tag */
.chroma .nv { color: #e36209; } /* Name variable */
.chroma .vc { color: #e36209; } /* Name variable class */
.chroma .vg { color: #e36209; } /* Name variable global */
.chroma .vi { color: #e36209; } /* Name variable instance */
.chroma .vm { color: #e36209; } /* Name variable magic */
.chroma .l { color: #0550ae; } /* Literal */
.chroma .s { color: #0a3069; } /* Literal string */
.chroma .sa { color: #0a3069; } /* Literal string affine */
.chroma .sb { color: #0a3069; } /* Literal string backtick */
.chroma .sc { color: #0a3069; } /* Literal string char */
.chroma .dl { color: #0a3069; } /* Literal string delimiter */
.chroma .sd { color: #0a3069; } /* Literal string doc */
.chroma .s2 { color: #0a3069; } /* Literal string double */
.chroma .se { color: #cf222e; } /* Literal string escape */
.chroma .sh { color: #0a3069; } /* Literal string heredoc */
.chroma .si { color: #0a3069; } /* Literal string interpol */
.chroma .sx { color: #0a3069; } /* Literal string other */
.chroma .sr { color: #0550ae; } /* Literal string regex */
.chroma .s1 { color: #0a3069; } /* Literal string single */
.chroma .ss { color: #cf222e; } /* Literal string symbol */
.chroma .m { color: #0550ae; } /* Literal number */
.chroma .mb { color: #0550ae; } /* Literal number bin */
.chroma .mf { color: #0550ae; } /* Literal number float */
.chroma .mh { color: #0550ae; } /* Literal number hex */
.chroma .mi { color: #0550ae; } /* Literal number integer */
.chroma .il { color: #0550ae; } /* Literal number long */
.chroma .mo { color: #0550ae; } /* Literal number oct */
.chroma .o { color: #cf222e; } /* Operator */
.chroma .ow { color: #cf222e; } /* Operator word */
.chroma .p { color: #24292e; } /* Punctuation */
.chroma .c { color: #6e7781; font-style: italic; } /* Comment */
.chroma .ch { color: #6e7781; font-style: italic; } /* Comment hashbang */
.chroma .cm { color: #6e7781; font-style: italic; } /* Comment multiline */
.chroma .cp { color: #6e7781; } /* Comment preproc */
.chroma .cpf { color: #6e7781; } /* Comment preproc file */
.chroma .c1 { color: #6e7781; font-style: italic; } /* Comment single */
.chroma .cs { color: #6e7781; font-style: italic; } /* Comment special */
.chroma .g { color: #24292e; } /* Generic */
.chroma .gd { color: #cf222e; background-color: #ffebe9; } /* Generic deleted */
.chroma .ge { color: #24292e; font-style: italic; } /* Generic emphasis */
.chroma .gh { color: #0550ae; font-weight: bold; } /* Generic heading */
.chroma .gs { color: #24292e; font-weight: bold; } /* Generic strong */
.chroma .gu { color: #959da5; font-weight: bold; } /* Generic subheading */
.chroma .gi { color: #22863a; background-color: #f0fff4; } /* Generic inserted */
.chroma .go { color: #959da5; } /* Generic output */
.chroma .gp { color: #959da5; } /* Generic prompt */
.chroma .gr { color: #cf222e; } /* Generic error */
.chroma .gt { color: #cf222e; } /* Generic traceback */
.chroma .gl { color: #24292e; text-decoration: underline; } /* Generic link */
</style>
`
	if c.pretty {
		fmt.Fprintln(w)
	}
	fmt.Fprint(w, css)
	if c.pretty {
		fmt.Fprintln(w)
	}
}

// convertSourceBlock converts a source/listing block with syntax highlighting.
func (c *HTML5Converter) convertSourceBlock(block *ast.StyledBlockNode, w io.Writer) {
	// Get the language from attributes
	language := "text" // default
	if lang, ok := block.Attributes["language"]; ok && lang != "" {
		language = lang
	}

	// Split content into lines
	lines := strings.Split(block.Content, "\n")
	if len(lines) == 0 {
		return
	}

	// Find the lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Tokenize the code
	content := strings.Join(lines, "\n")
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		// Fallback to plain pre/code if tokenization fails
		c.writeOpenTagWithClass("div", "listingblock", w)
		fmt.Fprintf(w, `<pre><code class="language-%s">%s</code></pre>`, language, c.escape(content))
		c.writeCloseTag("div", w)
		return
	}

	// Use the github style for syntax highlighting (light theme)
	style := styles.GitHub

	// Format the HTML - use functional options
	formatter := html.New(
		html.Standalone(false),
		html.WithClasses(true),
	)

	// Build the HTML
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, `<div class="listingblock">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, `<div class="content">`)
	if c.pretty {
		fmt.Fprintln(w)
		c.indent += "  "
	}

	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, `<pre class="highlight"><code class="language-%s chroma" data-lang="%s">`, language, language)
	if c.pretty {
		fmt.Fprintln(w)
	}

	// Write the highlighted code
	var buf strings.Builder
	if err := formatter.Format(&buf, style, iterator); err == nil {
		c.writeRawString(buf.String(), w)
	} else {
		// Fallback on error
		c.writeRawString(c.escape(content), w)
	}

	fmt.Fprintf(w, `</code></pre>`)

	if c.pretty {
		fmt.Fprintln(w)
		c.indent = strings.TrimSuffix(c.indent, "  ")
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, `</div>`)

	if c.pretty {
		fmt.Fprintln(w)
		c.indent = strings.TrimSuffix(c.indent, "  ")
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprintf(w, `</div>`)
	if c.pretty {
		fmt.Fprintln(w)
	}

	// Render callout descriptions if present
	if len(block.Callouts) > 0 {
		c.renderCalloutListForStyled(block, w)
	}
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

		// Write head section with meta tags for proper Reader Mode support
		fmt.Fprint(w, "<head>")
		if c.pretty {
			fmt.Fprintln(w)
		}
		fmt.Fprint(w, `<meta charset="utf-8">`)
		if c.pretty {
			fmt.Fprintln(w)
		}
		fmt.Fprint(w, `<meta name="viewport" content="width=device-width, initial-scale=1.0">`)
		if c.pretty {
			fmt.Fprintln(w)
		}
		// Write title
		title := "AsciiDoc Document"
		if doc.Header != nil && doc.Header.Title != "" {
			title = doc.Header.Title
		}
		fmt.Fprintf(w, "<title>%s</title>", c.escape(title))
		if c.pretty {
			fmt.Fprintln(w)
		}
		// Write CSS
		c.writeCSS(w)
		fmt.Fprint(w, "</head>")
		if c.pretty {
			fmt.Fprintln(w)
		}

		c.writeOpenTag("body", w)
		// Wrap main content in <article> for Reader Mode support
		fmt.Fprint(w, "<article>")
		if c.pretty {
			fmt.Fprintln(w)
		}
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
		fmt.Fprint(w, "</article>")
		if c.pretty {
			fmt.Fprintln(w)
		}
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
	case inline.NodePassThrough:
		// Inline passthrough: +text+ - content is passed through with substitutions
		// Apply special characters substitution only
		c.writeRawString(c.escape(node.Text), w)
	case inline.NodeRawPassThrough:
		// Raw inline passthrough: +++text+++ - content is passed through WITHOUT any escaping
		// This is used for HTML fragments, etc.
		c.writeRawString(node.Text, w)
	case inline.NodeCustomMacro:
		// Custom inline macro - check extension registry
		if c.extensionRegistry != nil {
			if processor, ok := c.extensionRegistry.GetInlineMacro(node.MacroName); ok {
				html, err := processor.Process(node.MacroTarget, node.MacroAttrs)
				if err == nil {
					c.writeRawString(html, w)
				} else {
					// On error, output the raw macro text
					c.writeRawString(c.escape(node.Text), w)
				}
			} else {
				// Unknown macro - output as-is
				c.writeRawString(c.escape(node.Text), w)
			}
		} else {
			// No registry - output as-is
			c.writeRawString(c.escape(node.Text), w)
		}
	case inline.NodeMenu:
		// UI menu path: menu:[File > New > Document]
		// Split by " > " and create spans with proper separators
		class := c.getClassAttr(node.Roles)
		classAttr := ""
		if class != "" {
			classAttr = fmt.Sprintf(` class="%s"`, class)
		}
		fmt.Fprintf(w, `<span class="menuseq"%s>`, classAttr)

		// Split the menu path by " > "
		parts := strings.Split(node.Text, " > ")
		for i, part := range parts {
			if i > 0 {
				fmt.Fprintf(w, `<span class="menusep">&#160;&#9656;</span>`)
			}
			fmt.Fprintf(w, `<span class="menu">%s</span>`, c.escape(part))
		}
		fmt.Fprintf(w, `</span>`)
	case inline.NodeButton:
		// UI button: btn:[OK]
		class := c.getClassAttr(node.Roles)
		classAttr := "button"
		if class != "" {
			classAttr = fmt.Sprintf("%s %s", class, "button")
		}
		fmt.Fprintf(w, `<b class="%s">%s</b>`, classAttr, c.escape(node.Text))
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

// renderInlineChildren renders child inline nodes within a parent inline node.
func (c *HTML5Converter) renderInlineChildren(node *inline.Node, w io.Writer) {
	// If there are children, render them; otherwise fall back to rendering text
	if len(node.Children) > 0 {
		lastEnd := 0
		for _, child := range node.Children {
			start := child.StartPos
			// Write any text before this child node (only if child is after lastEnd)
			if start > lastEnd {
				text := node.Text[lastEnd:start]
				c.writeRawString(c.escape(text), w)
			}
			lastEnd = child.Position
			// Render the child node
			c.convertInlineNode(child, w)
		}
		// Write any remaining text after last child
		if lastEnd < len(node.Text) {
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

	// Use semantic <section> tag for better Reader Mode compatibility
	// Keep the class attribute for styling compatibility
	fmt.Fprintf(w, `<section class="%s">`+"\n", sectionClass)

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

	fmt.Fprintf(w, `</section>`+"\n") // Close section wrapper (sectN)
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
	// Use semantic HTML5 <section> tag for better Reader Mode compatibility
	if c.pretty {
		fmt.Fprint(w, c.indent)
	}
	fmt.Fprint(w, `<section class="sect1">`)
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
	fmt.Fprint(w, "</section>")
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
	var orderedListClass string
	switch tag {
	case "ul":
		wrapperClass = "ulist"
	case "ol":
		// Determine ordered list class based on marker count
		orderedListClass = c.orderedListClass(list.Items[0])
		wrapperClass = "olist " + orderedListClass
	case "dl":
		wrapperClass = "dlist"
	}

	// Always add newlines for Asciidoctor compatibility
	fmt.Fprintf(w, `<div class="%s">`+"\n", wrapperClass)

	// For ordered lists, add class attribute
	if tag == "ol" {
		fmt.Fprintf(w, `<ol class="%s">`+"\n", orderedListClass)
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
		case ".", "..", "...", "....", ".....", "......", ".......":
			return "ol" // Ordered list (including nested)
		case ";", ";;", ";;;", ";;;;":
			return "ol" // Ordered list variant (semicolon-style)
		case "::", ":::", "::::", ";;;;:":
			return "dl" // Labeled/definition list
		}
	}
	// Check for multi-character ordered list markers (e.g. "..", ";;", "::")
	if li, ok := item.(*ast.NodeListItem); ok {
		marker := li.Marker
		if len(marker) > 0 && (marker[0] == '.' || marker[0] == ';') {
			return "ol" // Multi-dot or multi-semicolon is ordered list
		}
	}
	return "ul" // Default
}

// orderedListClass returns the CSS class for an ordered list based on marker count.
// Matches Asciidoctor's behavior:
// - 1 dot (.) = arabic (1, 2, 3)
// - 2 dots (..) = lowerroman (i, ii, iii)
// - 3 dots (...) = loweralpha (a, b, c)
// - 4 dots (....) = upperalpha (A, B, C)
// - 5+ dots cycles through the above
func (c *HTML5Converter) orderedListClass(item ast.Node) string {
	if li, ok := item.(*ast.NodeListItem); ok {
		markerLen := len(li.Marker)
		// Handle both dot (.) and semicolon (;) markers
		if markerLen > 0 && (li.Marker[0] == '.' || li.Marker[0] == ';') {
			// Cycle through styles based on marker count
			switch markerLen {
			case 1:
				return "arabic"
			case 2:
				return "lowerroman"
			case 3:
				return "loweralpha"
			case 4:
				return "upperalpha"
			case 5:
				return "lowerroman"
			case 6:
				return "loweralpha"
			default:
				return "arabic"
			}
		}
	}
	return "arabic" // Default
}

// convertListItem converts a list item to HTML.
func (c *HTML5Converter) convertListItem(item *ast.NodeListItem, w io.Writer) {
	// Determine tag based on list type
	if item.Marker == "::" {
		// Labeled list: dt and dd (Asciidoctor compatibility)
		fmt.Fprint(w, `<dt class="hdlist1">`)
		// Check if term has inline nodes
		if len(item.InlineNodes) == 0 {
			fmt.Fprint(w, c.escape(item.Term))
		} else {
			// Render term with inline formatting
			c.renderInlineText(item.Term, item.InlineNodes, w)
		}
		fmt.Fprint(w, `</dt>
`)
		// Write dd with p-wrapped content
		fmt.Fprint(w, `<dd>
`)
		fmt.Fprint(w, `<p>`)
		// Check if definition has inline nodes
		if len(item.DefinitionNodes) == 0 {
			fmt.Fprint(w, c.escape(item.Definition))
		} else {
			// Render definition with inline formatting
			c.renderInlineText(item.Definition, item.DefinitionNodes, w)
		}
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

// renderCalloutListForStyled renders callout descriptions as an HTML list for styled blocks.
func (c *HTML5Converter) renderCalloutListForStyled(styled *ast.StyledBlockNode, w io.Writer) {
	// Collect callouts that have descriptions
	var numberedCallouts []*ast.CalloutNode
	maxNumber := 0
	for _, co := range styled.Callouts {
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

// convertBlockContent converts block content to HTML, detecting and rendering lists.
// Supports unordered (*) and ordered (.) lists within block content.
func (c *HTML5Converter) convertBlockContent(lines []string, w io.Writer) {
	type segmentType int
	const (
		segParagraph segmentType = iota
		segList
	)

	type segment struct {
		segType segmentType
		content string
		marker  string // for lists
	}

	var segments []segment
	var currentPara []string
	var currentList []string
	var inList bool
	var listMarker string

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		isEmpty := strings.TrimRight(line, " \t") == ""

		// Check if line starts a list item
		isListItem := strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "- ") ||
			strings.HasPrefix(trimmed, ". ")

		if isListItem {
			// Flush any pending paragraph
			if len(currentPara) > 0 {
				segments = append(segments, segment{
					segType: segParagraph,
					content: strings.Join(currentPara, " "),
				})
				currentPara = nil
			}

			marker := string(trimmed[0])
			if !inList {
				// Start a new list
				inList = true
				listMarker = marker
				currentList = []string{line}
			} else if marker == listMarker {
				// Continuation of same list
				currentList = append(currentList, line)
			} else {
				// Different list type - flush current list and start new
				segments = append(segments, segment{
					segType: segList,
					content: c.convertListToHTML(currentList, listMarker),
				})
				listMarker = marker
				currentList = []string{line}
			}
		} else if inList {
			// Check if this is a continuation (indentation) or end of list
			if isEmpty || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				// Continuation of current list item
				if len(currentList) > 0 {
					currentList[len(currentList)-1] += "\n" + line
				}
			} else {
				// End of list - flush and start new paragraph
				segments = append(segments, segment{
					segType: segList,
					content: c.convertListToHTML(currentList, listMarker),
				})
				currentList = nil
				inList = false
				currentPara = []string{line}
			}
		} else {
			// Regular text line
			if isEmpty {
				// Flush any pending paragraph on blank line
				if len(currentPara) > 0 {
					segments = append(segments, segment{
						segType: segParagraph,
						content: strings.Join(currentPara, " "),
					})
					currentPara = nil
				}
			} else {
				currentPara = append(currentPara, line)
			}
		}
	}

	// Flush any pending paragraph
	if len(currentPara) > 0 {
		segments = append(segments, segment{
			segType: segParagraph,
			content: strings.Join(currentPara, " "),
		})
	}

	// Flush any remaining list
	if inList && len(currentList) > 0 {
		segments = append(segments, segment{
			segType: segList,
			content: c.convertListToHTML(currentList, listMarker),
		})
	}

	// Convert segments to HTML
	for _, seg := range segments {
		if seg.segType == segList {
			fmt.Fprint(w, seg.content)
		} else if seg.content != "" {
			fmt.Fprintf(w, `<div class="paragraph">`+"\n")
			fmt.Fprintf(w, `<p>%s</p>`+"\n", c.escape(seg.content))
			fmt.Fprintf(w, `</div>`+"\n")
		}
	}
}

// convertListToHTML converts a list of raw list items to HTML.
func (c *HTML5Converter) convertListToHTML(items []string, marker string) string {
	if len(items) == 0 {
		return ""
	}

	var tag string
	switch marker {
	case "*", "-":
		tag = "ul"
	case ".":
		tag = "ol"
	default:
		tag = "ul"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, `<div class="ulist">`+"\n")
	fmt.Fprintf(&sb, `<%s>`+"\n", tag)

	for _, item := range items {
		// Extract the content after the marker
		trimmed := strings.TrimLeft(item, " \t")
		if len(trimmed) >= 2 && (trimmed[1] == ' ' || trimmed[1] == '\t') {
			trimmed = trimmed[2:]
		}
		// Remove any trailing newline content
		trimmed = strings.ReplaceAll(trimmed, "\n", " ")
		trimmed = strings.TrimSpace(trimmed)

		fmt.Fprintf(&sb, `<li>`+"\n")
		fmt.Fprintf(&sb, `<p>%s</p>`, c.escape(trimmed))
		fmt.Fprintf(&sb, `</li>`+"\n")
	}

	fmt.Fprintf(&sb, `</%s>`+"\n", tag)
	fmt.Fprintf(&sb, `</div>`+"\n")

	return sb.String()
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
		// Example block - support nested content (lists, paragraphs)
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="content">`)
		fmt.Fprint(w, "\n")
		c.convertBlockContent(block.Lines, w)
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `</div>`)
		fmt.Fprint(w, "\n")
	} else if block.Delimiter == "*" {
		// Sidebar block - support nested content (lists, paragraphs)
		fmt.Fprintf(w, `<div class="%s">`, class)
		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, `<div class="content">`)
		fmt.Fprint(w, "\n")
		c.convertBlockContent(block.Lines, w)
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

	// Handle source and listing blocks with syntax highlighting
	if block.Style == "source" || block.Style == "listing" {
		c.convertSourceBlock(block, w)
		return
	}

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
	// Wrap in passblock div for Asciidoctor compatibility
	fmt.Fprintf(w, `<div class="passblock">`+"\n")
	fmt.Fprintf(w, `<div class="content">`+"\n")
	// Passthrough content is written directly without HTML escaping
	c.writeRawString(pass.Content, w)
	fmt.Fprintf(w, `</div>`+"\n")
	fmt.Fprintf(w, `</div>`+"\n")
}

// convertVerse converts a verse block to HTML.
// Verse blocks preserve line breaks and formatting.
func (c *HTML5Converter) convertVerse(verse *ast.VerseNode, w io.Writer) {
	c.writeOpenTagWithClass("div", "verseblock", w)

	// Use <pre class="content"> like Asciidoctor
	// This preserves whitespace and line breaks
	fmt.Fprint(w, c.indent)
	fmt.Fprint(w, `<pre class="content">`)
	fmt.Fprint(w, c.escape(verse.Content))
	fmt.Fprint(w, `</pre>`)
	if c.pretty {
		fmt.Fprintln(w)
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
	// Skip colgroup if %autowidth is set (columns auto-size to content)
	if table.Attributes["autowidth"] != "true" {
		c.writeOpenTag("colgroup", w)
		numCols := table.ColumnCount()
		if numCols > 0 {
			colWidth := 100.0 / float64(numCols)
			// Track the sum of floored widths for accurate last column calculation
			sumFloored := 0.0
			for i := 0; i < numCols; i++ {
				var width float64
				if i == numCols-1 {
					// Last column gets the remainder based on actual floored values
					width = 100.0 - sumFloored
				} else {
					width = colWidth
				}
				if c.pretty {
					fmt.Fprint(w, c.indent)
				}
				// Asciidoctor uses integer for whole numbers, 4 decimals for fractions
				// For fractional values, Asciidoctor floors to 4 decimals (not rounds)
				var displayWidth float64
				if width == math.Trunc(width) {
					displayWidth = width
					fmt.Fprintf(w, `<col style="width: %.0f%%;">`, displayWidth)
				} else {
					// Floor to 4 decimals like Asciidoctor
					displayWidth = math.Floor(width*10000) / 10000
					// Format with variable precision to trim trailing zeros
					widthStr := fmt.Sprintf("%.4f", displayWidth)
					widthStr = strings.TrimRight(widthStr, "0")
					widthStr = strings.TrimRight(widthStr, ".")
					fmt.Fprintf(w, `<col style="width: %s%%;">`, widthStr)
				}
				sumFloored += displayWidth
				if c.pretty {
					fmt.Fprintln(w)
				}
			}
		}
		c.writeCloseTag("colgroup", w)
	}

	// Write caption if present
	if table.Caption != "" {
		c.writeElement("caption", c.escape(table.Caption), w)
	}

	// Write thead if table has a header row
	if table.HasHeader() {
		c.writeOpenTag("thead", w)
		headerRow := table.HeaderRow()
		c.writeTableRow(headerRow, "th", w)
		c.writeCloseTag("thead", w)
	}

	// Write tbody for body rows
	c.writeOpenTag("tbody", w)
	for _, row := range table.Rows {
		// Skip header row if it was already rendered in thead
		if table.HasHeader() && row.Kind == ast.TableRowHeader {
			continue
		}
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

	// Determine content wrapper based on cell style
	style := cell.Style
	var contentWrapper string

	switch style {
	case "literal":
		contentWrapper = "code"  // <code> for literal
	case "monospace":
		contentWrapper = "code"  // <code> for monospace
	case "verse":
		contentWrapper = "div"   // <div> for verse (preserves line breaks)
	case "emphasis":
		contentWrapper = "em"    // <em> for emphasis
	case "strong", "header":
		contentWrapper = "strong" // <strong> for strong/header
	default:
		contentWrapper = "p"     // <p> for normal content
	}

	// Check if cell has block content (lists, etc.)
	if len(cell.Blocks) > 0 {
		// Render block content (no content wrapper needed)
		for _, block := range cell.Blocks {
			c.convertNode(block, w)
		}
	} else {
		// Write opening content wrapper tag
		if contentWrapper != "" {
			fmt.Fprintf(w, `<%s class="tableblock">`, contentWrapper)
		}

		// Check if cell has multi-line content
		cellText := cell.Text
		hasMultiLine := strings.Contains(cellText, "\n")

		if len(cell.InlineNodes) == 0 {
			if hasMultiLine {
				// For multi-line cells, replace newlines with <br> tags
				lines := strings.Split(cellText, "\n")
				for i, line := range lines {
					c.writeRawString(c.escape(line), w)
					if i < len(lines)-1 {
						c.writeRawString(`<br>`, w)
					}
				}
			} else {
				c.writeRawString(c.escape(cell.Text), w)
			}
		} else {
			// Render inline nodes
			lastEnd := 0
			for _, node := range cell.InlineNodes {
				if inlineNode, ok := node.(*inline.Node); ok {
					startPos := inlineNode.StartPos
					// Write any text before this inline node
					if startPos > lastEnd {
						text := cell.Text[lastEnd:startPos]
						// Handle multi-line text between inline nodes
						if strings.Contains(text, "\n") {
							lines := strings.Split(text, "\n")
							for i, line := range lines {
								c.writeRawString(c.escape(line), w)
								if i < len(lines)-1 {
									c.writeRawString(`<br>`, w)
								}
							}
						} else {
							c.writeRawString(c.escape(text), w)
						}
					}
					lastEnd = inlineNode.Position
					// Render the inline node
					c.convertInlineNode(inlineNode, w)
				}
			}
			// Write any remaining text
			if lastEnd < len(cell.Text) {
				text := cell.Text[lastEnd:]
				if strings.Contains(text, "\n") {
					lines := strings.Split(text, "\n")
					for i, line := range lines {
						c.writeRawString(c.escape(line), w)
						if i < len(lines)-1 {
							c.writeRawString(`<br>`, w)
						}
					}
				} else {
					c.writeRawString(c.escape(text), w)
				}
			}
		}

		// Close the content wrapper tag
		if contentWrapper != "" {
			fmt.Fprintf(w, `</%s>`, contentWrapper)
		}
	}
	fmt.Fprintf(w, "</%s>", tag)
	if c.pretty {
		fmt.Fprintln(w)
	}
}
