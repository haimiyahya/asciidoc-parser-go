# asciidoc-parser-go

A native Go implementation of an [AsciiDoc](https://asciidoc.org/) parser and processor, targeting near-full compliance with the [Eclipse AsciiDoc Language Specification](https://gitlab.eclipse.org/asciidoc-lang/asciidoc-lang) while maintaining compatibility with [Asciidoctor](https://asciidoctor.org/).

## Features

- **Line-oriented Reader** - Mimics human visual scanning of AsciiDoc markup
- **Block Classifier** - Identifies paragraphs, lists, sections, delimited blocks, tables, etc.
- **Inline Parser** - Handles bold, italic, monospace, links, images, superscript, subscript
- **AST Builder** - Rich Abstract Syntax Tree preserving source locations
- **Attribute Processor** - Document attribute management and substitution
- **Include Processor** - `include::[]` directive with tag filtering and line ranges
- **Conditional Processing** - `ifdef`, `ifndef`, `ifeval` directives
- **Multiple Output Backends**:
  - HTML5 converter
  - PDF converter (with TOC, cover page, metadata)
  - DocBook 5.1.1 converter
  - Man Page (troff/nroff) converter
  - EPUB converter
- **CLI Tool** - Full-featured command-line interface compatible with Asciidoctor options

## Installation

```bash
go install github.com/haimiyahya/asciidoc-parser-go/cmd/asciidoc@latest
```

Or build from source:

```bash
git clone https://github.com/haimiyahya/asciidoc-parser-go
cd asciidoc-parser-go
go build ./cmd/asciidoc
```

## Quick Start

### Using the CLI

```bash
# Convert to HTML (default)
asciidoc document.adoc

# Convert to PDF
asciidoc -b pdf document.adoc

# Convert to PDF with TOC and cover page
asciidoc -b pdf --pdf-toc --pdf-cover-page document.adoc

# Convert to DocBook
asciidoc -b docbook document.adoc

# Convert to Man Page
asciidoc -b manpage mycommand.adoc

# Convert to EPUB
asciidoc -b epub document.adoc
```

### Parsing a Document (Library)

```go
package main

import (
    "fmt"
    "os"

    "github.com/haimiyahya/asciidoc-parser-go/internal/parser"
)

func main() {
    source := `= Document Title

    == Section One

    This is a *bold* paragraph.

    - Item 1
    - Item 2
    `

    p, err := parser.NewParser(source)
    if err != nil {
        panic(err)
    }

    doc, err := p.Parse()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Parsed %d blocks\n", len(doc.Blocks))
}
```

### Converting to HTML5 (Library)

```go
package main

import (
    "os"

    "github.com/haimiyahya/asciidoc-parser-go/internal/converter"
    "github.com/haimiyahya/asciidoc-parser-go/internal/parser"
)

func main() {
    source := `= My Document

    This is **bold** and __italic__ text.

    image:logo.png[Alt Text]
    `

    p, _ := parser.NewParser(source)
    doc, _ := p.Parse()

    htmlConverter := converter.NewHTML5Converter()
    htmlConverter.Convert(doc, os.Stdout)
}
```

**Output:**
```html
<!DOCTYPE html>
<html>
  <body>
    <h1>My Document</h1>
    <p>This is <strong>bold</strong> and <em>italic</em> text.</p>
    <img src="logo.png" alt="Alt Text">
  </body>
</html>
```

### Converting to PDF (Library)

```go
pdfConverter := converter.NewPDFConverter()

// Enable PDF enhancements
pdfConverter.WithTOC(true)              // Table of contents
pdfConverter.WithCoverPage(true)        // Cover page
pdfConverter.WithA4()                   // A4 page size

// Set PDF metadata
pdfConverter.WithPDFTitle("My Document")
pdfConverter.WithPDFAuthor("John Doe")

pdfConverter.Convert(doc, &buf)
```

## Supported Syntax

### Blocks

| Syntax | Description |
|--------|-------------|
| `= Title` | Level 0 (document) heading |
| `== Section` | Level 1 section heading |
| `=== Subsection` | Level 2 subsection heading |
| Plain text | Paragraph |
| `* Item` | Unordered list item |
| `. Item` | Ordered list item |
| `Term :: Definition` | Labeled/description list |
| `----` | Delimited block (example/quote/etc) |
| `====` | Example block |
| `____` | Quote block |
| `////` | Literal block |
| `|===` | Table block |

### Tables

```
|===                                    Table delimiter (start/end)
| Cell 1 | Cell 2 |                     Table row
|= Header |                              Header row
| >right | ^center | <left |             Cell alignment
[cols="1,2,3"]                          Column specifications
[caption="Title"]                        Table caption
2+                                      Colspan
3*                                      Cell repeat
```

Tables support:
- Header rows with `|=`
- Cell alignment (`<` left, `>` right, `^` center)
- Column specifications (width, style)
- Table attributes (frame, grid, stripes, width)
- Multiple data formats (PSV, CSV, TSV, DSV)

### Inline Markup

| Syntax | Output | Notes |
|--------|--------|-------|
| `**bold**` | **bold** | Constrained bold |
| `*word*` | *word* | Unconstrained bold (single word) |
| `__italic__` | __italic__ | Constrained italic |
| `_word_` | _word_ | Unconstrained italic (single word) |
| `` `code` `` | `code` | Monospace |
| `++code++` | `++code++` | Monospace alternative |
| `^sup^` | sup | Superscript |
| `~sub~` | sub | Subscript |
| `link:text[url]` | <a href="url">text</a> | Macro link |
| `https://url` | <a href="url">url</a> | Bare URL |
| `image:path[alt]` | <img src="path" alt="alt"> | Inline image |
| `<<section-id>>` | <a href="#section-id">section-id</a> | Cross-reference |
| `<<id,text>>` | <a href="#id">text</a> | Cross-reference with custom text |
| `[.role]**text**` | <span class="role">text</span> | Role/CSS class |
| `kbd:[Ctrl+C]` | <kbd><span class="key">Ctrl</span>+<span class="key">C</span></kbd> | Keyboard shortcut |
| `btn:[OK]` | <b class="btn">OK</b> | Button label |
| `menu:[File > Save]` | <span class="menu">...</span> | Menu path |

### Roles

Roles allow you to add CSS classes to inline elements:

```
[.red]**This text is red**
[.role1.role2]__This has two classes__
[.highlight]`code with highlight`
```

### UI Macros

UI macros help document user interface elements:

```
Press kbd:[Ctrl+C] to copy.
Click btn:[Submit] to continue.
Go to menu:[File > Save As].
```

### Admonitions

```
NOTE: This is important.
TIP: Try this approach.
WARNING: Be careful!
CAUTION: Watch out!
IMPORTANT: Pay attention.
```

### Callouts

Callouts allow you to annotate lines in code blocks or literal content:

```
----
line of code // <1>
another line # <2>
----
<1> Description for the first callout
<2> Description for the second callout
```

Supported line comment prefixes for callouts:
- `//` for C-style languages (C, C++, Go, Java, JavaScript)
- `#` for shell, Ruby, Python, Perl
- `;;` for Clojure
- `<!--1-->` for XML/HTML

### Bibliography

Bibliography sections allow you to define reference entries that can be cited throughout your document:

```
[bibliography]
== Bibliography

* [[[pp]]] Andy Hunt & Dave Thomas. **The Pragmatic Programmer**.
* [[[gof,gang]]] Erich Gamma et al. __Design Patterns__.
```

Citations in text use cross-reference syntax:

```
As discussed in <<pp>>, the Pragmatic Programmer approach...
See <<gof>> for design patterns.
```

Syntax:
- `* [[[label]]]` - Bibliography entry with label
- `* [[[label,xreftext]]]` - Entry with custom reference text
- `<<label>>` - Citation that renders as `[label]` or `[xreftext]`

## CLI Options

```
Options:
  -b, --backend BACKEND           Set backend output format: [html5, pdf, docbook, manpage, epub]
  -d, --doctype DOCTYPE           Document type: [article, book, manpage, inline]
  -e, --embedded                  Suppress enclosing document structure
  -o, --out-file FILE             Output file (default: based on input file; use - for STDOUT)
  -n, --section-numbers           Auto-number section titles
  -a, --attribute name[=value]    Define a document attribute
  -B, --base-dir DIR              Base directory containing the document and resources
  -D, --destination-dir DIR       Destination output directory
  -q, --quiet                     Silence application log messages
  -v, --verbose                   Enable verbose output
  -h, --help                      Show help message

PDF Options:
  --pdf-page-size SIZE            PDF page size: [letter, a4]
  --pdf-toc                       Generate table of contents
  --pdf-cover-page                Generate cover page
  --pdf-no-page-numbers           Disable page numbers
  --pdf-header TEMPLATE           PDF header template
  --pdf-footer TEMPLATE           PDF footer template
```

## Architecture

```
internal/
├── ast/              # Abstract Syntax Tree node definitions
├── attribute_parser/ # Attribute parsing and substitution
├── blocks/           # Block-level AST node types
├── converter/        # Output converters (HTML5, PDF, DocBook, ManPage, EPUB)
├── inline/           # Inline markup parser
├── parser/           # Main parser orchestrator
├── processor/        # Attribute, include, conditional processors
└── reader/           # Line-oriented reader with classifier
```

### Data Flow

```
Source Text
    ↓
[Reader] → Lines with Position
    ↓
[Classifier] → Block Types
    ↓
[Parser] → AST (Document + Blocks + Inline)
    ↓
[Processor] → Attribute substitution, Include expansion
    ↓
[Converter] → HTML5, PDF, DocBook, Man Page, EPUB
```

## Roadmap

- [x] Phase 0: Line-Oriented Reader + Basic Block Classifier
- [x] Phase 1: Core Parser Block Detection
- [x] Phase 2: Inline Parser
- [x] Phase 3: AST Builder
- [x] Phase 4: Attribute Processor
- [x] Phase 5: Include Processor
- [x] Phase 6: HTML5 Converter
- [x] Phase 7: Testing and Validation
- [x] Phase 8: CLI and Tooling
- [x] Phase 9: PDF Backend
- [x] Phase 10: DocBook Backend
- [x] Phase 11: Man Page Backend
- [x] Phase 12: EPUB Backend

See [ROADMAP.md](ROADMAP.md) for details.

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/parser/...

# Run with verbose output
go test ./internal/inline/... -v
```

### Project Status

| Component | Status | Notes |
|-----------|--------|-------|
| Reader | ✅ Complete | Line-oriented with lookahead |
| Block Classifier | ✅ Complete | All AsciiDoc block types |
| List Parsing | ✅ Complete | Nested, ordered, unordered, labeled |
| Section Parsing | ✅ Complete | Multi-level headings |
| Table Parsing | ✅ Complete | Delimited block syntax, attributes, alignment, colspan, rowspan |
| Inline Parsing | ✅ Complete | Bold, italic, monospace, links, images, superscript, subscript, roles, kbd, btn, menu |
| Admonitions | ✅ Complete | All 5 types |
| Block Macros | ✅ Complete | Image, video, audio, include |
| Delimited Blocks | ✅ Complete | Example, quote, literal, styled blocks (pass::[], sidebar::[], verse::[]) |
| AST Builder | ✅ Complete | Rich node hierarchy |
| HTML5 Converter | ✅ Complete | Semantic HTML5 |
| PDF Converter | ✅ Complete | With TOC, cover page, metadata |
| DocBook Converter | ✅ Complete | DocBook 5.1.1 |
| Man Page Converter | ✅ Complete | troff/nroff format |
| EPUB Converter | ✅ Complete | EPUB 2.0.1 |
| Attribute Processor | ✅ Complete | Document attribute handling |
| Include Processor | ✅ Complete | Include directive with tag filtering |
| Conditional Processing | ✅ Complete | ifdef, ifndef, ifeval |
| Bibliography | ✅ Complete | Citation processing with [[[label]]] syntax |
| CLI | ✅ Complete | Full Asciidoctor-compatible options |

### Missing/Incomplete Features (Post-MVP)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Extensions System** | Custom block/inline macros, tree processors | Medium |
| **Index Terms** | `(((term)))` indexing | Low |
| **LSP Support** | Language Server Protocol for editor integration | Medium |

## Contributing

This project follows [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation
- `refactor:` for code refactoring
- `test:` for test changes

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Acknowledgments

- [AsciiDoc](https://asciidoc.org/) - The lightweight markup language
- [Asciidoctor](https://asciidoctor.org/) - Reference implementation for behavior
- [Eclipse AsciiDoc Language Project](https://gitlab.eclipse.org/asciidoc-lang/asciidoc-lang) - Language specification

## Resources

- [AsciiDoc Language Specification](https://gitlab.eclipse.org/asciidoc-lang/asciidoc-lang)
- [Asciidoctor Documentation](https://docs.asciidoctor.org/asciidoc/latest/)
- [AsciiDoc Quick Reference](https://asciidoc.org/)
