// Package converter provides output converters for AsciiDoc AST.
//
// The converter package transforms parsed AsciiDoc documents (represented as AST)
// into various output formats. The primary output is HTML5, with support for
// DocBook 5.1, PDF (via DocBook), man pages, and EPUB.
//
// # Basic Usage
//
//	factory := converter.NewConverterFactory()
//	c, _ := factory.Get(converter.BackendHTML5)
//
//	doc := parseAsciidoc(source) // *ast.NodeDocument
//	var buf bytes.Buffer
//	err := c.Convert(doc, &buf)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	html := buf.String()
//
// # Converter Interface
//
// The Converter interface defines the contract for all backends:
//
//	type Converter interface {
//	    Convert(doc *ast.NodeDocument, w io.Writer) error
//	}
//
// Custom converters can be registered with the factory using Register().
//
// # Supported Backends
//
//   - BackendHTML5: HTML5 output (default)
//   - BackendDocBook: DocBook 5.1 XML
//   - BackendPDF: PDF via DocBook transformation
//   - BackendMan: troff/nroff man page format
package converter

import (
	"io"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// Converter converts an AsciiDoc AST to an output format.
//
// The Converter interface is implemented by all output backends.
// Each backend traverses the AST and writes formatted output to
// the provided io.Writer.
type Converter interface {
	// Convert converts the document to the target format.
	//
	// Convert writes the formatted output to w and returns any
	// error encountered during conversion.
	Convert(doc *ast.NodeDocument, w io.Writer) error
}

// BackendType represents the output backend type.
//
// BackendType identifies which converter to use for document conversion.
type BackendType string

const (
	// BackendHTML5 is the HTML5 backend.
	//
	// The HTML5 backend generates semantic HTML5 output with
	// appropriate classes for styling.
	BackendHTML5 BackendType = "html5"

	// BackendDocBook is the DocBook 5.1 backend.
	//
	// The DocBook backend generates DocBook 5.1 XML, which can
	// be further transformed into other formats like PDF.
	BackendDocBook BackendType = "docbook"

	// BackendPDF is the PDF backend (via DocBook).
	//
	// The PDF backend generates output through DocBook-to-PDF
	// transformation.
	BackendPDF BackendType = "pdf"

	// BackendMan is the troff/nroff man page backend.
	//
	// The man page backend generates troff/nroff source suitable
	// for UNIX manual pages.
	BackendMan BackendType = "man"
)

// ConverterFactory creates converters for a backend.
//
// The factory manages registered converters and provides lookup
// by backend type. The HTML5 converter is registered by default.
type ConverterFactory struct {
	backends map[BackendType]Converter
}

// NewConverterFactory creates a new ConverterFactory with HTML5 registered.
//
// The factory is initialized with the HTML5 converter as the default.
// Additional converters can be registered using Register().
func NewConverterFactory() *ConverterFactory {
	f := &ConverterFactory{
		backends: make(map[BackendType]Converter),
	}
	// Register HTML5 as default backend
	f.Register(BackendHTML5, NewHTML5Converter())
	return f
}

// Register registers a converter for a backend.
//
// Register adds a new converter to the factory, associating it
// with the specified backend type. If a converter already exists
// for the backend, it is replaced.
func (f *ConverterFactory) Register(backend BackendType, c Converter) {
	f.backends[backend] = c
}

// Get retrieves a converter for a backend.
//
// Get returns the converter registered for the given backend type
// and a boolean indicating whether it was found.
func (f *ConverterFactory) Get(backend BackendType) (Converter, bool) {
	c, ok := f.backends[backend]
	return c, ok
}

// GetDefault returns the default HTML5 converter.
//
// GetDefault is a convenience method for obtaining the HTML5 converter
// without explicitly calling Get(BackendHTML5).
func (f *ConverterFactory) GetDefault() Converter {
	c, ok := f.Get(BackendHTML5)
	if !ok {
		return NewHTML5Converter()
	}
	return c
}
