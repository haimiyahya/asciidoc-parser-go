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

// Convert converts the document to PDF.
func (c *PDFConverter) Convert(doc *ast.NodeDocument, w io.Writer) error {
	// First, convert AST to HTML
	var htmlBuf strings.Builder
	if err := c.htmlConverter.Convert(doc, &htmlBuf); err != nil {
		return fmt.Errorf("failed to convert to HTML: %w", err)
	}

	html := htmlBuf.String()

	// Wrap HTML with PDF-specific styling
	html = c.wrapWithPDFStyles(html)

	// Convert HTML to PDF using chromedp
	pdfData, err := c.htmlToPDF(html)
	if err != nil {
		return err
	}

	// Write PDF to output
	_, err = w.Write(pdfData)
	return err
}

// wrapWithPDFStyles wraps HTML with appropriate styles for PDF output.
func (c *PDFConverter) wrapWithPDFStyles(html string) string {
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
</html>`, css, bodyContent)
}

// getPDFCSS returns CSS for PDF output.
func (c *PDFConverter) getPDFCSS() string {
	return `
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
		chromedpOptions:     []chromedp.ContextOption{},
	}
}
