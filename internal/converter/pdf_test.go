package converter

import (
	"bytes"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// TestPDFConverterConfiguration tests that the PDF converter can be configured properly.
// This test does not require Chrome/Chromium to be installed.
func TestPDFConverterConfiguration(t *testing.T) {
	t.Run("NewPDFConverter", func(t *testing.T) {
		c := NewPDFConverter()
		if c == nil {
			t.Fatal("NewPDFConverter returned nil")
		}
		if c.pageWidth != 8.5 {
			t.Errorf("expected default page width 8.5, got %f", c.pageWidth)
		}
		if c.pageHeight != 11.0 {
			t.Errorf("expected default page height 11.0, got %f", c.pageHeight)
		}
	})

	t.Run("WithPageSize", func(t *testing.T) {
		c := NewPDFConverter().WithPageSize(7.5, 10.0)
		if c.pageWidth != 7.5 {
			t.Errorf("expected page width 7.5, got %f", c.pageWidth)
		}
		if c.pageHeight != 10.0 {
			t.Errorf("expected page height 10.0, got %f", c.pageHeight)
		}
	})

	t.Run("WithA4", func(t *testing.T) {
		c := NewPDFConverter().WithA4()
		if c.pageWidth != 8.27 {
			t.Errorf("expected A4 width 8.27, got %f", c.pageWidth)
		}
		if c.pageHeight != 11.69 {
			t.Errorf("expected A4 height 11.69, got %f", c.pageHeight)
		}
	})

	t.Run("WithMargins", func(t *testing.T) {
		c := NewPDFConverter().WithMargins(1.0, 1.0, 1.0, 1.0)
		if c.marginTop != 1.0 {
			t.Errorf("expected margin top 1.0, got %f", c.marginTop)
		}
		if c.marginBottom != 1.0 {
			t.Errorf("expected margin bottom 1.0, got %f", c.marginBottom)
		}
		if c.marginLeft != 1.0 {
			t.Errorf("expected margin left 1.0, got %f", c.marginLeft)
		}
		if c.marginRight != 1.0 {
			t.Errorf("expected margin right 1.0, got %f", c.marginRight)
		}
	})

	t.Run("WithTOC", func(t *testing.T) {
		c := NewPDFConverter().WithTOC(true)
		if !c.generateTOC {
			t.Error("expected TOC generation to be enabled")
		}
	})

	t.Run("WithPageNumbers", func(t *testing.T) {
		c := NewPDFConverter().WithPageNumbers(false)
		if c.showPageNumbers {
			t.Error("expected page numbers to be disabled")
		}
	})

	t.Run("WithCoverPage", func(t *testing.T) {
		c := NewPDFConverter().WithCoverPage(true).
			WithCoverTitle("Test Title").
			WithCoverAuthor("Test Author").
			WithCoverDate("2024-01-01")

		if !c.enableCoverPage {
			t.Error("expected cover page to be enabled")
		}
		if c.coverTitle != "Test Title" {
			t.Errorf("expected cover title 'Test Title', got '%s'", c.coverTitle)
		}
		if c.coverAuthor != "Test Author" {
			t.Errorf("expected cover author 'Test Author', got '%s'", c.coverAuthor)
		}
		if c.coverDate != "2024-01-01" {
			t.Errorf("expected cover date '2024-01-01', got '%s'", c.coverDate)
		}
	})

	t.Run("WithCustomCSS", func(t *testing.T) {
		customCSS := "body { font-size: 12pt; }"
		c := NewPDFConverter().WithCustomCSS(customCSS)
		if c.customCSS != customCSS {
			t.Errorf("expected custom CSS to be set")
		}
	})

	t.Run("WithPDFMetadata", func(t *testing.T) {
		c := NewPDFConverter().
			WithPDFTitle("Test Document").
			WithPDFAuthor("Test Author").
			WithPDFKeywords("test, pdf").
			WithPDFSubject("Test Subject")

		if c.pdfTitle != "Test Document" {
			t.Errorf("expected PDF title 'Test Document', got '%s'", c.pdfTitle)
		}
		if c.pdfAuthor != "Test Author" {
			t.Errorf("expected PDF author 'Test Author', got '%s'", c.pdfAuthor)
		}
		if c.pdfKeywords != "test, pdf" {
			t.Errorf("expected PDF keywords 'test, pdf', got '%s'", c.pdfKeywords)
		}
		if c.pdfSubject != "Test Subject" {
			t.Errorf("expected PDF subject 'Test Subject', got '%s'", c.pdfSubject)
		}
	})
}

// TestPDFConverterHTMLWrapping tests the HTML wrapping functionality.
func TestPDFConverterHTMLWrapping(t *testing.T) {
	c := NewPDFConverter()

	doc := &ast.NodeDocument{
		Header: &ast.DocumentHeader{
			Title:  "Test Document",
			Author: "Test Author",
		},
		Blocks: []ast.Node{
			&ast.NodeParagraph{
				Text: "This is a test paragraph.",
				Pos:  ast.Position{Line: 1},
			},
		},
	}

	html := c.wrapWithPDFStyles("<p>This is a test paragraph.</p>", doc)

	// Check for required elements
	required := []string{
		"<!DOCTYPE html>",
		"<html>",
		"<head>",
		"<style>",
		"</style>",
		"<body>",
		"</body>",
		"</html>",
	}

	for _, req := range required {
		if !bytes.Contains([]byte(html), []byte(req)) {
			t.Errorf("HTML wrapper missing required element: %s", req)
		}
	}
}

// TestSlugify tests the slugify function used for generating TOC IDs.
func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Test Document Title", "test-document-title"},
		{"Already-Slugified", "already-slugified"},
		{"With_Underscores", "with_underscores"},
		{"Multiple   Spaces", "multiple---spaces"},
		{"Special@#$Characters", "specialcharacters"},
		{"123 Numbers", "123-numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPDFConverterExtractMetadata tests metadata extraction from documents.
func TestPDFConverterExtractMetadata(t *testing.T) {
	doc := &ast.NodeDocument{
		Header: &ast.DocumentHeader{
			Title:  "Document Title",
			Author: "Document Author",
		},
		Attributes: map[string]string{
			"keywords": "test, keywords",
			"subject":  "Test Subject",
		},
	}

	c := NewPDFConverter()
	c.extractMetadata(doc)

	if c.pdfTitle != "Document Title" {
		t.Errorf("expected title 'Document Title', got '%s'", c.pdfTitle)
	}
	if c.pdfAuthor != "Document Author" {
		t.Errorf("expected author 'Document Author', got '%s'", c.pdfAuthor)
	}
	if c.pdfKeywords != "test, keywords" {
		t.Errorf("expected keywords 'test, keywords', got '%s'", c.pdfKeywords)
	}
	if c.pdfSubject != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got '%s'", c.pdfSubject)
	}
}

// TestGenerateCoverPageHTML tests cover page HTML generation.
func TestGenerateCoverPageHTML(t *testing.T) {
	c := NewPDFConverter().
		WithCoverTitle("Test Title").
		WithCoverSubtitle("Test Subtitle").
		WithCoverAuthor("Test Author").
		WithCoverDate("January 1, 2024")

	html := c.generateCoverPageHTML()

	required := []string{
		`class="cover-page"`,
		`class="cover-title"`,
		`>Test Title<`,
		`class="cover-subtitle"`,
		`>Test Subtitle<`,
		`class="cover-author"`,
		`>Test Author<`,
		`class="cover-date"`,
		`>January 1, 2024<`,
	}

	for _, req := range required {
		if !bytes.Contains([]byte(html), []byte(req)) {
			t.Errorf("Cover page HTML missing: %s", req)
		}
	}
}

// TestGenerateTOCHTML tests TOC HTML generation.
func TestGenerateTOCHTML(t *testing.T) {
	doc := &ast.NodeDocument{
		Blocks: []ast.Node{
			&ast.NodeSection{
				Level: 1,
				Title: "Section 1",
				ID:    "section-1",
				Children: []ast.Node{
					&ast.NodeSection{
						Level: 2,
						Title: "Subsection 1.1",
						ID:    "subsection-1-1",
					},
				},
			},
			&ast.NodeSection{
				Level: 1,
				Title: "Section 2",
				ID:    "section-2",
			},
		},
	}

	c := NewPDFConverter().WithTOCTitle("Contents")

	html := c.generateTOCHTML(doc)

	required := []string{
		`class="toc"`,
		`class="toc-title"`,
		`>Contents<`,
		`class="toc-list"`,
		`toc-item`,    // Note: class is combined like "toc-item toc-level-1"
		`toc-level-1`, // Check for level-1 class
		`>Section 1<`,
		`toc-level-2`, // Check for level-2 class
		`>Subsection 1.1<`,
		`>Section 2<`,
		`toc-sublist`, // Check for nested list
	}

	for _, req := range required {
		if !bytes.Contains([]byte(html), []byte(req)) {
			t.Errorf("TOC HTML missing: %s\nActual HTML:\n%s", req, html)
		}
	}
}

// TestGetPDFCSS tests PDF CSS generation.
func TestGetPDFCSS(t *testing.T) {
	c := NewPDFConverter()
	css := c.getPDFCSS()

	required := []string{
		"@page",
		"body {",
		"font-family:",
		".page-break",
		"page-break-after: always",
		".cover-page",
		".toc",
		"color:",
	}

	for _, req := range required {
		if !bytes.Contains([]byte(css), []byte(req)) {
			t.Errorf("PDF CSS missing: %s", req)
		}
	}
}
