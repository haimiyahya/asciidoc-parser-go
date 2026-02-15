# Roadmap for asciidoc-parser-go

This document outlines the phased implementation plan for building a native Go implementation of a complete AsciiDoc parser and processor.

## Phase 0: Line-Oriented Reader + Basic Block Classifier ✅

**Status**: Complete

**Objectives**:
- Implement a line-oriented Reader that mimics human visual scanning of text
- Implement Block Classifier to identify block types based on line patterns
- Establish position tracking for error reporting

**Implementation**:
- `internal/reader/reader.go`: Core Reader with lookahead, mark/restore
- `internal/reader/classifier.go`: LineClassifier for all block types
- `internal/reader/reader_test.go`: Comprehensive tests

**Key Design Decisions**:
- Lines stored in reverse order for efficient popping (matches Asciidoctor)
- 1-based line numbers (human convention)
- Position tracking (file, path, lineno)
- Human-like visual pattern matching with minimal backtracking

## Phase 1: Core Parser Block Detection ✅

**Objectives**:
- Implement block boundary detection (delimited blocks)
- Implement list detection and nesting logic
- Implement attribute entry parsing
- Implement section header parsing
- Implement inline markup detection (bold, italic, etc.)

**Implementation**:
- Create `internal/parser/block.go` for block-level parsing
- Extend LineClassifier with block-level context
- Add support for:
  - Delimited blocks: `----`, `====`, `____`, `++++`, `////`, `****`, `....`
  - Lists: ordered (`.`), unordered (`-`, `*`, `o`), labeled (`::`)
  - Attributes: `:name: value`, `:name!`
  - Sections: `= Title` (level based on `=` count)

## Phase 2: Inline Parser ✅

**Objectives**:
- Implement inline parsing for:
  - Text formatting: `**bold**`, `*italic*`, ``_monospace_``, `^superscript^`, `~subscript~`
  - Links: `link:text[url]`, `https://url` or `url`
  - Images: `image:path[alt]`
  - Attributes references: `{attr}` (inline substitution)
  - Passthrough: `+++pass+++` (raw output)
  - Macros: `macro::target[attrs]` (block/inline)

**Implementation**:
- Create `internal/parser/inline.go` for inline element parsing
- State machine for nested inline formatting (e.g., `**bold with _italic_**`)
- Proper attribute reference resolution
- Link and image URL parsing

## Phase 3: AST Builder ✅

**Objectives**:
- Construct rich Abstract Syntax Tree from parsed content
- Preserve source location for error reporting
- Store block-level and inline-level structure
- Support document metadata (header attributes)

**Implementation**:
- Extend `internal/ast/ast.go` with:
  - `Document` node with header, blocks, attributes
  - `Section` nodes with level, title, id, attributes
  - `Block` nodes: Paragraph, List, ListItem, Literal, Verbatim, etc.
  - `Inline` nodes: Text, Bold, Italic, Link, Image, etc.
  - Position tracking for all nodes

## Phase 4: Attribute Processor ✅

**Objectives**:
- Implement attribute substitution/replacement
- Implement conditional processing (`ifdef`, `ifndef`, `ifeval`)
- Implement document-level attribute defaults

**Implementation**:
- `internal/processor/attributes.go` for attribute management
- `internal/processor/conditional.go` for conditional directives
- Attribute inheritance and scoping
- Predefined attributes (e.g., `toc`, `sectnums`)

## Phase 5: Include Processor ✅

**Objectives**:
- Implement `include::[]` directive
- Support tag filtering (`tag=` and `tags=`)
- Handle `lines=` attribute for line ranges
- Control include depth and circular reference detection

**Implementation**:
- `internal/processor/include.go` for include directive handling
- Relative path resolution
- Safe mode enforcement
- Include stack tracking for error reporting

## Phase 6: HTML5 Converter ✅

**Objectives**:
- Convert AST to HTML5 output
- Implement semantic HTML elements
- Support syntax highlighting (optional)
- Handle all block and inline types

**Implementation**:
- `internal/converter/html5.go` for HTML5 backend
- Visitor pattern for AST traversal
- Proper escaping and safe HTML generation
- Support for converter extensions (DocBook, PDF)

## Phase 7: Testing and Validation ✅

**Objectives**:
- Validate against Asciidoctor test suite
- Create comprehensive test coverage
- Performance benchmarking

**Implementation**:
- Integration tests comparing with Asciidoctor output
- Table-driven tests for all components
- Examples from AsciiDoc Language Specification
- Benchmarking for optimization opportunities

## Phase 8: CLI and Tooling ✅

**Objectives**:
- Command-line interface for common operations
- Watch mode for auto-processing on file changes
- Stdin/stdout support for pipeline usage
- Library mode for embedding

**Implementation**:
- `cmd/asciidoc/main.go` CLI implementation
- Full flag support: input/output files, backend selection, attributes
- Stdin/stdout support for pipeline usage
- Config file support (`~/.config/asciidoc/config.toml`) - TODO
- Plugin/extension system - TODO
- Watch mode (fsnotify or polling) - TODO

## Phase 9: PDF Backend ✅

**Status**: Complete

**Objectives**:
- Convert AsciiDoc to PDF using Chrome/Chromium headless
- Support page size and margin configuration
- Professional PDF styling with custom CSS

**Implementation**:
- `internal/converter/pdf.go`: PDF converter using chromedp
- Converts AST → HTML → PDF via headless Chrome
- Supports Letter, A4, and custom page sizes
- Configurable margins and print options
- Professional CSS styling for PDF output

**Note**: Requires Chrome or Chromium to be installed for PDF generation.

## Phase 10: DocBook Backend ✅

**Status**: Complete

**Objectives**:
- Convert AsciiDoc to DocBook 5.1.1 XML
- Support article and book document types
- Enable technical documentation pipelines

**Implementation**:
- `internal/converter/docbook.go`: DocBook 5.1.1 converter
- Full support for sections, paragraphs, lists, inline elements
- Admonitions, tables, code blocks, quotes
- Media objects (images, video, audio)
- Proper XML namespaces and DocBook 5.1.1 compliance

## Future Enhancements (Post-MVP)

### Additional Backends
- DocBook 5.1.1 ✅
- Man page via groff/troff
- EPUB via asciidoc-epub3

### PDF Enhancements
- Table of contents generation
- Page numbering options
- Cover page support
- Custom headers and footers
- PDF metadata (author, title, keywords)

### Extension Points
- Custom block macros
- Custom inline macros (xref, btn, etc.)
- Tree processors (AST transformations)
- Postprocessors (output manipulation)

### Editor Integration
- Language Server Protocol (LSP) support
- VS Code extension
- Vim/Neovim plugin
- Emacs major mode

## Notes

- This roadmap follows the principle of "incremental and verifiable"
- Each phase should produce passing tests before starting the next
- The architecture prioritizes "spec-first, compatibility-second"
- Performance optimization should not compromise correctness
