// Package converter provides PDF converter for AsciiDoc AST.
package converter

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// PDFConverter converts an AsciiDoc AST to PDF using Chrome/Chromium.
type PDFConverter struct {
	// htmlConverter is the underlying HTML5 converter used for rendering.
	htmlConverter *HTML5Converter

	// page size options
	pageWidth     float64
	pageHeight    float64
	marginTop     float64
	marginBottom  float64
	marginLeft    float64
	marginRight   float64

	// display options
	displayHeaderFooter bool
	printBackground     bool

	// custom CSS for PDF styling
	customCSS string

	// table of contents
	generateTOC      bool
	tocTitle         string
	tocMaxLevel      int

	// page numbering
	showPageNumbers bool
	pageNumberFormat string // "roman" for lowercase roman, "ROMAN" for uppercase, "arabic" for numbers

	// cover page
	enableCoverPage bool
	coverTitle      string
	coverSubtitle   string
	coverAuthor     string
	coverDate       string

	// custom headers and footers
	headerTemplate string
	footerTemplate string

	// PDF metadata
	pdfTitle       string
	pdfAuthor      string
	pdfKeywords    string
	pdfSubject     string

	// chromedp options
	chromedpOptions []chromedp.ContextOption
}

// NewPDFConverter creates a new PDF converter.
func NewPDFConverter() *PDFConverter {
	return &PDFConverter{
		htmlConverter:      NewHTML5Converter(),
		pageWidth:          8.5,   // Letter width in inches
		pageHeight:         11.0,  // Letter height in inches
		marginTop:          0.4,   // 0.4 inch margins
		marginBottom:       0.4,
		marginLeft:         0.4,
		marginRight:        0.4,
		displayHeaderFooter: false,
		printBackground:     true,
		generateTOC:         false,
		tocTitle:            "Table of Contents",
		tocMaxLevel:         3,
		showPageNumbers:     true,
		pageNumberFormat:    "arabic",
		enableCoverPage:     false,
		chromedpOptions:     []chromedp.ContextOption{},
	}
}

// WithPageSize sets the page size in inches.
func (c *PDFConverter) WithPageSize(width, height float64) *PDFConverter {
	c.pageWidth = width
	c.pageHeight = height
	return c
}

// WithMargins sets the page margins in inches.
func (c *PDFConverter) WithMargins(top, bottom, left, right float64) *PDFConverter {
	c.marginTop = top
	c.marginBottom = bottom
	c.marginLeft = left
	c.marginRight = right
	return c
}

// WithDisplayHeaderFooter enables or disables header and footer display.
func (c *PDFConverter) WithDisplayHeaderFooter(display bool) *PDFConverter {
	c.displayHeaderFooter = display
	return c
}

// WithPrintBackground enables or disables background graphics printing.
func (c *PDFConverter) WithPrintBackground(print bool) *PDFConverter {
	c.printBackground = print
	return c
}

// WithCustomCSS adds custom CSS for PDF styling.
func (c *PDFConverter) WithCustomCSS(css string) *PDFConverter {
	c.customCSS = css
	return c
}

// WithA4 sets page size to A4.
func (c *PDFConverter) WithA4() *PDFConverter {
	c.pageWidth = 8.27
	c.pageHeight = 11.69
	return c
}

// WithLetter sets page size to Letter.
func (c *PDFConverter) WithLetter() *PDFConverter {
	c.pageWidth = 8.5
	c.pageHeight = 11.0
	return c
}

// WithTOC enables table of contents generation.
func (c *PDFConverter) WithTOC(enable bool) *PDFConverter {
	c.generateTOC = enable
	return c
}

// WithTOCTitle sets the title for the table of contents.
func (c *PDFConverter) WithTOCTitle(title string) *PDFConverter {
	c.tocTitle = title
	return c
}

// WithTOCMaxLevel sets the maximum level for table of contents entries.
func (c *PDFConverter) WithTOCMaxLevel(level int) *PDFConverter {
	c.tocMaxLevel = level
	return c
}

// WithPageNumbers enables or disables page numbers.
func (c *PDFConverter) WithPageNumbers(enable bool) *PDFConverter {
	c.showPageNumbers = enable
	return c
}

// WithPageNumberFormat sets the page number format (arabic, roman, ROMAN).
func (c *PDFConverter) WithPageNumberFormat(format string) *PDFConverter {
	c.pageNumberFormat = format
	return c
}

// WithCoverPage enables cover page generation.
func (c *PDFConverter) WithCoverPage(enable bool) *PDFConverter {
	c.enableCoverPage = enable
	return c
}

// WithCoverTitle sets the cover page title.
func (c *PDFConverter) WithCoverTitle(title string) *PDFConverter {
	c.coverTitle = title
	return c
}

// WithCoverSubtitle sets the cover page subtitle.
func (c *PDFConverter) WithCoverSubtitle(subtitle string) *PDFConverter {
	c.coverSubtitle = subtitle
	return c
}

// WithCoverAuthor sets the cover page author.
func (c *PDFConverter) WithCoverAuthor(author string) *PDFConverter {
	c.coverAuthor = author
	return c
}

// WithCoverDate sets the cover page date.
func (c *PDFConverter) WithCoverDate(date string) *PDFConverter {
	c.coverDate = date
	return c
}

// WithHeaderTemplate sets the custom header template.
func (c *PDFConverter) WithHeaderTemplate(template string) *PDFConverter {
	c.headerTemplate = template
	c.displayHeaderFooter = true
	return c
}

// WithFooterTemplate sets the custom footer template.
func (c *PDFConverter) WithFooterTemplate(template string) *PDFConverter {
	c.footerTemplate = template
	c.displayHeaderFooter = true
	return c
}

// WithPDFTitle sets the PDF title metadata.
func (c *PDFConverter) WithPDFTitle(title string) *PDFConverter {
	c.pdfTitle = title
	return c
}

// WithPDFAuthor sets the PDF author metadata.
func (c *PDFConverter) WithPDFAuthor(author string) *PDFConverter {
	c.pdfAuthor = author
	return c
}

// WithPDFKeywords sets the PDF keywords metadata.
func (c *PDFConverter) WithPDFKeywords(keywords string) *PDFConverter {
	c.pdfKeywords = keywords
	return c
}

// WithPDFSubject sets the PDF subject metadata.
func (c *PDFConverter) WithPDFSubject(subject string) *PDFConverter {
	c.pdfSubject = subject
	return c
}

// Convert converts the document to PDF.
func (c *PDFConverter) Convert(doc *ast.NodeDocument, w io.Writer) error {
	// Extract metadata from document
	c.extractMetadata(doc)

	// First, convert AST to HTML
	var htmlBuf strings.Builder
	if err := c.htmlConverter.Convert(doc, &htmlBuf); err != nil {
		return fmt.Errorf("failed to convert to HTML: %w", err)
	}

	html := htmlBuf.String()

	// Wrap HTML with PDF-specific styling
	html = c.wrapWithPDFStyles(html, doc)

	// Convert HTML to PDF using chromedp
	pdfData, err := c.htmlToPDF(html)
	if err != nil {
		return err
	}

	// Write PDF to output
	_, err = w.Write(pdfData)
	return err
}

// extractMetadata extracts metadata from the document for PDF.
func (c *PDFConverter) extractMetadata(doc *ast.NodeDocument) {
	// Extract title
	if doc.Header != nil && doc.Header.Title != "" {
		if c.pdfTitle == "" {
			c.pdfTitle = doc.Header.Title
		}
		if c.enableCoverPage && c.coverTitle == "" {
			c.coverTitle = doc.Header.Title
		}
	}

	// Extract author
	if doc.Header != nil && doc.Header.Author != "" {
		if c.pdfAuthor == "" {
			c.pdfAuthor = doc.Header.Author
		}
		if c.enableCoverPage && c.coverAuthor == "" {
			c.coverAuthor = doc.Header.Author
		}
	}

	// Also check attributes
	if title, ok := doc.Attributes["title"]; ok && c.pdfTitle == "" {
		c.pdfTitle = title
	}
	if author, ok := doc.Attributes["author"]; ok && c.pdfAuthor == "" {
		c.pdfAuthor = author
	}
	if keywords, ok := doc.Attributes["keywords"]; ok {
		c.pdfKeywords = keywords
	}
	if subject, ok := doc.Attributes["subject"]; ok {
		c.pdfSubject = subject
	}

	// Check for TOC attribute
	if _, ok := doc.Attributes["toc"]; ok {
		c.generateTOC = true
	}
}

// generateTOCHTML generates table of contents HTML from the document.
func (c *PDFConverter) generateTOCHTML(doc *ast.NodeDocument) string {
	var toc strings.Builder
	toc.WriteString(fmt.Sprintf(`<div class="toc">
<h2 class="toc-title">%s</h2>
<ul class="toc-list">`, c.tocTitle))

	c.extractSectionsForTOC(doc.Blocks, &toc, 1)

	toc.WriteString("</ul></div>")
	return toc.String()
}

// extractSectionsForTOC extracts sections from blocks for TOC.
func (c *PDFConverter) extractSectionsForTOC(blocks []ast.Node, toc *strings.Builder, level int) {
	for _, block := range blocks {
		if sec, ok := block.(*ast.NodeSection); ok {
			if sec.Level <= c.tocMaxLevel {
				// Generate ID if not present
				id := sec.ID
				if id == "" {
					id = slugify(sec.Title)
				}

				toc.WriteString(fmt.Sprintf(`<li class="toc-item toc-level-%d"><a href="#%s">%s</a>`, sec.Level, id, sec.Title))

				// Process subsections
				if len(sec.Children) > 0 {
					toc.WriteString("<ul class=\"toc-sublist\">")
					c.extractSectionsForTOC(sec.Children, toc, level+1)
					toc.WriteString("</ul>")
				}

				toc.WriteString("</li>")
			}
		} else if doc, ok := block.(*ast.NodeDocument); ok {
			c.extractSectionsForTOC(doc.Blocks, toc, level)
		}
	}
}

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		} else if r == ' ' {
			result.WriteRune('-')
		}
	}
	return result.String()
}

// generateCoverPageHTML generates cover page HTML.
func (c *PDFConverter) generateCoverPageHTML() string {
	title := c.coverTitle
	subtitle := c.coverSubtitle
	author := c.coverAuthor
	date := c.coverDate

	if date == "" {
		date = time.Now().Format("January 2, 2006")
	}

	var html strings.Builder
	html.WriteString(`<div class="cover-page">`)
	html.WriteString(`<div class="cover-content">`)

	if title != "" {
		html.WriteString(fmt.Sprintf(`<h1 class="cover-title">%s</h1>`, htmlEscape(title)))
	}

	if subtitle != "" {
		html.WriteString(fmt.Sprintf(`<h2 class="cover-subtitle">%s</h2>`, htmlEscape(subtitle)))
	}

	html.WriteString(`<div class="cover-spacer"></div>`)

	if author != "" {
		html.WriteString(fmt.Sprintf(`<p class="cover-author">%s</p>`, htmlEscape(author)))
	}

	if date != "" {
		html.WriteString(fmt.Sprintf(`<p class="cover-date">%s</p>`, htmlEscape(date)))
	}

	html.WriteString(`</div>`)
	html.WriteString(`</div>`)

	return html.String()
}

// htmlEscape escapes HTML special characters.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// wrapWithPDFStyles wraps HTML with appropriate styles for PDF output.
func (c *PDFConverter) wrapWithPDFStyles(html string, doc *ast.NodeDocument) string {
	// Extract body content if DOCTYPE is present
	bodyContent := html
	if strings.HasPrefix(html, "<!DOCTYPE html>") || strings.HasPrefix(html, "<html") {
		// Simple extraction - find content between <body> tags
		start := strings.Index(html, "<body>")
		end := strings.Index(html, "</body>")
		if start != -1 && end != -1 {
			start += len("<body>")
			bodyContent = html[start:end]
		}
	}

	// Build content with cover page and TOC if enabled
	var contentBuilder strings.Builder

	// Add cover page
	if c.enableCoverPage {
		contentBuilder.WriteString(c.generateCoverPageHTML())
		contentBuilder.WriteString(`<div class="page-break"></div>`)
	}

	// Add TOC
	if c.generateTOC {
		contentBuilder.WriteString(c.generateTOCHTML(doc))
		contentBuilder.WriteString(`<div class="page-break"></div>`)
	}

	// Add main content
	contentBuilder.WriteString(bodyContent)

	css := c.getPDFCSS()
	if c.customCSS != "" {
		css += "\n" + c.customCSS
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
%s
</style>
</head>
<body>
%s
</body>
</html>`, css, contentBuilder.String())
}

// getPDFCSS returns CSS for PDF output.
func (c *PDFConverter) getPDFCSS() string {
	css := `
@page {
	margin: 0;
}

body {
	font-family: "Noto Serif", Georgia, "Times New Roman", Times, serif;
	font-size: 11pt;
	line-height: 1.6;
	color: #333;
	max-width: 100%;
	margin: 0;
	padding: 0;
}

/* Page break */
.page-break {
	page-break-after: always;
}

/* Cover page styles */
.cover-page {
	height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	page-break-after: always;
}

.cover-content {
	text-align: center;
	padding-top: 20vh;
}

.cover-title {
	font-size: 36pt;
	font-weight: 600;
	color: #000;
	margin-bottom: 0.5em;
	page-break-after: avoid;
}

.cover-subtitle {
	font-size: 18pt;
	font-weight: 400;
	color: #666;
	margin-bottom: 2em;
	page-break-after: avoid;
}

.cover-spacer {
	flex-grow: 1;
	min-height: 20vh;
}

.cover-author {
	font-size: 14pt;
	color: #333;
	margin: 0.5em 0;
}

.cover-date {
	font-size: 12pt;
	color: #666;
	margin: 0.5em 0;
}

/* Table of contents styles */
.toc {
	page-break-after: always;
	padding: 2em 0;
}

.toc-title {
	font-size: 24pt;
	font-weight: 600;
	color: #000;
	border-bottom: 2px solid #333;
	padding-bottom: 0.5em;
	margin-bottom: 1em;
}

.toc-list {
	list-style: none;
	padding-left: 0;
}

.toc-item {
	margin: 0.5em 0;
}

.toc-item a {
	color: #333;
	text-decoration: none;
	display: block;
}

.toc-item a:hover {
	text-decoration: underline;
}

.toc-level-1 {
	font-size: 14pt;
	font-weight: 600;
	padding-left: 0;
}

.toc-level-2 {
	font-size: 12pt;
	font-weight: 400;
	padding-left: 2em;
}

.toc-level-3 {
	font-size: 11pt;
	font-weight: 400;
	padding-left: 4em;
}

.toc-level-4 {
	font-size: 10pt;
	font-weight: 400;
	padding-left: 6em;
}

.toc-sublist {
	list-style: none;
	padding-left: 0;
	margin-top: 0.3em;
}

/* Page number styles */
@media print {
	@page {
		@bottom-right {
			content: counter(page);
			font-size: 10pt;
			color: #666;
		}
	}
`

	// Add page number CSS based on format
	if c.showPageNumbers {
		css += `
/* Page numbering */
.pagenum {
	display: none;
}
`
	}

	// Add header/footer template CSS
	if c.displayHeaderFooter {
		css += `
/* Header and footer */
@page {
	margin-top: 1in;
	margin-bottom: 1in;
}

@page {
	@top-center {
		content: "` + c.headerTemplate + `";
		font-size: 10pt;
		color: #666;
	}
	@bottom-center {
		content: "` + c.footerTemplate + `";
		font-size: 10pt;
		color: #666;
	}
}
`
	}

	css += `
h1, h2, h3, h4, h5, h6 {
	font-family: "Noto Sans", Helvetica, Arial, sans-serif;
	font-weight: 600;
	page-break-after: avoid;
	margin-top: 1.5em;
	margin-bottom: 0.5em;
}

h1 { font-size: 24pt; color: #000; border-bottom: 2px solid #ddd; padding-bottom: 0.3em; }
h2 { font-size: 18pt; color: #333; border-bottom: 1px solid #eee; padding-bottom: 0.2em; }
h3 { font-size: 14pt; color: #444; }
h4 { font-size: 12pt; color: #555; }

p {
	margin: 0.5em 0;
	text-align: justify;
}

a {
	color: #0066cc;
	text-decoration: none;
}

a[href^="http"] {
	word-break: break-all;
}

code {
	font-family: "Courier New", Courier, monospace;
	background-color: #f5f5f5;
	padding: 0.2em 0.4em;
	border-radius: 3px;
	font-size: 0.9em;
}

pre {
	background-color: #f5f5f5;
	border: 1px solid #ddd;
	border-radius: 4px;
	padding: 1em;
	overflow-x: auto;
	page-break-inside: avoid;
}

pre code {
	background-color: transparent;
	padding: 0;
	border-radius: 0;
}

ul, ol {
	padding-left: 2em;
}

li {
	margin: 0.3em 0;
}

table {
	border-collapse: collapse;
	width: 100%;
	margin: 1em 0;
	page-break-inside: avoid;
}

th, td {
	border: 1px solid #ddd;
	padding: 0.5em 0.8em;
	text-align: left;
}

th {
	background-color: #f5f5f5;
	font-weight: 600;
}

blockquote {
	border-left: 4px solid #ddd;
	margin: 1em 0;
	padding-left: 1em;
	color: #666;
	font-style: italic;
}

hr {
	border: none;
	border-top: 1px solid #ddd;
	margin: 2em 0;
}

.admonition-note, .admonition-tip, .admonition-warning, .admonition-caution, .admonition-important {
	border-left: 4px solid;
	padding: 0.8em 1em;
	margin: 1em 0;
	background-color: #f9f9f9;
	page-break-inside: avoid;
}

.admonition-note { border-left-color: #0066cc; }
.admonition-tip { border-left-color: #009900; }
.admonition-warning { border-left-color: #ff9900; }
.admonition-caution { border-left-color: #cc0000; }
.admonition-important { border-left-color: #000; }
`
	return css
}

// htmlToPDF converts HTML to PDF using chromedp.
func (c *PDFConverter) htmlToPDF(html string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create chromedp context
	opts := c.chromedpOptions
	ctx, cancel = chromedp.NewContext(ctx, opts...)
	defer cancel()

	var pdfBuf []byte

	// Run chromedp tasks
	err := chromedp.Run(ctx,
		// Navigate to blank page first
		chromedp.Navigate("about:blank"),

		// Set the HTML content
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`document.write(%q)`, html), nil).Do(ctx)
		}),

		// Wait for page to stabilize
		chromedp.Sleep(100*time.Millisecond),

		// Set viewport (affects rendering)
		emulation.SetDeviceMetricsOverride(1200, 800, 1.0, false),

		// Print to PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			printToPDFOpts := page.PrintToPDF().
				WithPaperWidth(c.pageWidth).
				WithPaperHeight(c.pageHeight).
				WithMarginTop(c.marginTop).
				WithMarginBottom(c.marginBottom).
				WithMarginLeft(c.marginLeft).
				WithMarginRight(c.marginRight).
				WithDisplayHeaderFooter(c.displayHeaderFooter).
				WithPrintBackground(c.printBackground).
				WithLandscape(false).
				WithPreferCSSPageSize(true)

			// Add header template if specified
			if c.headerTemplate != "" {
				printToPDFOpts = printToPDFOpts.WithHeaderTemplate(c.headerTemplate)
			} else if c.displayHeaderFooter {
				// Default header template with title
				headerTpl := `<span class="title"></span>`
				printToPDFOpts = printToPDFOpts.WithHeaderTemplate(headerTpl)
			}

			// Add footer template if specified
			if c.footerTemplate != "" {
				printToPDFOpts = printToPDFOpts.WithFooterTemplate(c.footerTemplate)
			} else if c.displayHeaderFooter && c.showPageNumbers {
				// Default footer template with page numbers
				footerTpl := `<span class="pageNumber"></span> of <span class="totalPages"></span>`
				printToPDFOpts = printToPDFOpts.WithFooterTemplate(footerTpl)
			}

			var err error
			pdfBuf, _, err = printToPDFOpts.Do(ctx)
			// Note: stream return value is ignored (empty when not using ReturnAsStream)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to PDF: %w", err)
	}

	return pdfBuf, nil
}

// NewPDFConverterWithHTML creates a new PDF converter with a custom HTML converter.
func NewPDFConverterWithHTML(htmlConverter *HTML5Converter) *PDFConverter {
	return &PDFConverter{
		htmlConverter:      htmlConverter,
		pageWidth:          8.5,
		pageHeight:         11.0,
		marginTop:          0.4,
		marginBottom:       0.4,
		marginLeft:         0.4,
		marginRight:        0.4,
		displayHeaderFooter: false,
		printBackground:     true,
		generateTOC:         false,
		tocTitle:            "Table of Contents",
		tocMaxLevel:         3,
		showPageNumbers:     true,
		pageNumberFormat:    "arabic",
		enableCoverPage:     false,
		chromedpOptions:     []chromedp.ContextOption{},
	}
}
