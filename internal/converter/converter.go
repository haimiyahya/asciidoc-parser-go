// Package converter provides output converters for AsciiDoc AST.
package converter

import (
	"io"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// Converter converts an AsciiDoc AST to an output format.
type Converter interface {
	// Convert converts the document to the target format.
	Convert(doc *ast.Document, w io.Writer) error
}

// BackendType represents the output backend type.
type BackendType string

const (
	// BackendHTML5 is the HTML5 backend.
	BackendHTML5 BackendType = "html5"

	// BackendDocBook is the DocBook 5.1 backend.
	BackendDocBook BackendType = "docbook"

	// BackendPDF is the PDF backend (via DocBook).
	BackendPDF BackendType = "pdf"

	// BackendMan is the troff/nroff man page backend.
	BackendMan BackendType = "man"
)

// ConverterFactory creates converters for a backend.
type ConverterFactory struct {
	backends map[BackendType]Converter
}

// NewConverterFactory creates a new converter factory.
func NewConverterFactory() *ConverterFactory {
	return &ConverterFactory{
		backends: make(map[BackendType]Converter),
	}
}

// Register registers a converter for a backend.
func (f *ConverterFactory) Register(backend BackendType, c Converter) {
	f.backends[backend] = c
}

// Get retrieves a converter for a backend.
func (f *ConverterFactory) Get(backend BackendType) (Converter, bool) {
	c, ok := f.backends[backend]
	return c, ok
}
