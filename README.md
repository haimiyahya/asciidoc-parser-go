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
- **LSP Server** - Language Server Protocol for editor integration (VS Code, Vim, Emacs, etc.)

## Installation

### CLI Tool

```bash
go install github.com/haimiyahya/asciidoc-parser-go/cmd/asciidoc@latest
```

Or build from source:

```bash
git clone https://github.com/haimiyahya/asciidoc-parser-go
cd asciidoc-parser-go
go build ./cmd/asciidoc
```

### LSP Server

```bash
go install github.com/haimiyahya/asciidoc-parser-go/cmd/asciidoc-lsp@latest
```

Or build from source:

```bash
cd asciidoc-parser-go
go build ./cmd/asciidoc-lsp
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

### Using the LSP Server

The Language Server Protocol (LSP) server provides editor integration features like:

- **Diagnostics** - Real-time error checking as you type
- **Document Symbols** - Outline view of sections
- **Completions** - Context-aware suggestions for AsciiDoc syntax
- **Hover** - Information about AsciiDoc constructs
- **Go-to-Definition** - Navigate to section references
- **Document Highlights** - Highlight all occurrences of a word

**VS Code Configuration:**

```json
{
  "asciidoc.languageServerPath": "/path/to/asciidoc-lsp",
  "asciidoc.trace.server": "verbose",
  "files.associations": {
    "*.adoc": "asciidoc",
    "*.asciidoc": "asciidoc"
  }
}
```

**Vim/Neovim (with coc.nvim):**

```vim
" coc-settings.json
{
  "languageserver": {
    "asciidoc": {
      "command": "/path/to/asciidoc-lsp",
      "filetypes": ["asciidoc"],
      "rootPatterns": [".git"]
    }
  }
}
```

**Emacs (lsp-mode):**

```elisp
(lsp-register-client
  (make-lsp-client :new-connection (lsp-stdio-connection "/path/to/asciidoc-lsp")
                   :major-modes '(asciidoc-mode)
                   :server-id 'asciidoc-lsp))
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
[cols="2*l"]                             Column specifications with styles
[caption="Title"]                        Table caption
|===|                                    Cell separator (empty cell)
2+                                       Colspan indicator
```

Tables support:
- **Column specifications** with styles (`[cols="2*l"]` for literal columns)
- **Table attributes** (frame, grid, stripes, width, caption)
- **Multiple data formats** (PSV, CSV, TSV, DSV)

**Note:** In basic PSV tables, per-cell indicators like `l|` (cell style), `3*` (repeat), or `|+` (continuation) are treated as literal cell content by Asciidoctor. Use column specifications instead for applying styles to entire columns.

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
| `+text+` | `text` | Inline passthrough |
| `++text++` | `text` | Inline passthrough (alternative) |
| `+++text+++` | `text` | Raw passthrough |
| `link:text[url]` | <a href="url">text</a> | Macro link |
| `https://url` | <a href="url">url</a> | Bare URL |
| `image:path[alt]` | <img src="path" alt="alt"> | Inline image |
| `<<section-id>>` | <a href="#section-id">section-id</a> | Cross-reference |
| `<<id,text>>` | <a href="#id">text</a> | Cross-reference with custom text |
| `[.role]**text**` | <span class="role">text</span> | Role/CSS class |

**Custom UI Macros** (extensions, not in AsciiDoc spec):
| `kbd:[Ctrl+C]` | <kbd>...</kbd> | Keyboard shortcut |
| `btn:[OK]` | <b class="btn">OK</b> | Button label |
| `menu:[File > Save]` | <span class="menu">...</span> | Menu path |

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

### Index Terms

Index terms allow you to create searchable indexes for printed output (PDF, DocBook):

```
This paragraph contains a ((visible term)) that appears in both text and index.

This paragraph has a (((hidden, term))) that only appears in the index.

(((primary, secondary, tertiary))) multi-level terms are supported.
```

Syntax:
- `((term))` - Flow index term (visible in text and index)
- `(((term)))` - Concealed index term (only in index)
- `(((primary, secondary, tertiary)))` - Multi-level index entry
- `(((term, "comma, in term")))` - Quoted terms with commas

Note: The HTML5 converter marks index terms with `data-indexterm` attributes but does not generate an index. Use PDF or DocBook converters for full index generation.

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
├── lsp/              # Language Server Protocol implementation
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
- [x] Phase 13: LSP Server

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

### Asciidoctor Compatibility Testing

This project includes a compatibility testing framework that compares the parser's output against reference Asciidoctor HTML to ensure compatibility:

```bash
# Run compatibility tests
go test ./tests/compatibility/...

# Check if Asciidoctor is available (for generating expected output)
go test ./tests/compatibility/... -run TestCompatibility_AsciidoctorAvailable -v
```

**Current Status: 29/32 tests passing (90.6% compatibility)**

The compatibility framework supports:
- **Golden file testing**: Pre-captured expected HTML files for regression testing
- **Asciidoctor integration**: When `asciidoctor` is installed, generates expected HTML on-the-fly
- **Detailed diff reporting**: Shows exact differences between expected and actual output
- **32 built-in test cases**: Covering basic syntax, lists, inline markup, admonitions, blocks, tables, index terms, bibliography, roles, and passthrough

**Known Differences (Custom Extensions):**
- `kbd:[...]`, `btn:[...]`, `menu:[...]` UI macros are custom extensions not supported by default Asciidoctor

#### Using Asciidoctor for Expected Output

For the most accurate compatibility testing, install Asciidoctor:

```bash
# Install Ruby (if not already installed)
# On macOS: brew install ruby
# On Ubuntu/Debian: apt install ruby

# Install Asciidoctor
gem install asciidoctor
```

When Asciidoctor is available in your PATH, the compatibility tests will automatically use it to generate expected HTML output.

#### Regenerating Golden Files

To update golden files (e.g., after implementing a new feature):

```bash
GENERATE_GOLDEN=1 go test ./tests/compatibility/... -run TestCompatibility_GenerateGoldenFiles -v
```

### Project Status

| Component | Status | Notes |
|-----------|--------|-------|
| Reader | ✅ Complete | Line-oriented with lookahead |
| Block Classifier | ✅ Complete | All AsciiDoc block types |
| List Parsing | ✅ Complete | Nested, ordered, unordered, labeled |
| Section Parsing | ✅ Complete | Multi-level headings |
| Table Parsing | ✅ Complete | Column specifications, attributes, formats (PSV/CSV/TSV/DSV) |
| Inline Parsing | ✅ Complete | Bold, italic, monospace, links, images, superscript, subscript, passthrough, roles |
| Admonitions | ✅ Complete | All 5 types |
| Block Macros | ✅ Complete | Image, video, audio, include |
| Delimited Blocks | ✅ Complete | Example, quote, literal, styled blocks (pass::[], sidebar::[], verse::[]) |
| AST Builder | ✅ Complete | Rich node hierarchy |
| HTML5 Converter | ✅ Complete | Semantic HTML5, Asciidoctor-compatible formatting |
| PDF Converter | ✅ Complete | With TOC, cover page, metadata |
| DocBook Converter | ✅ Complete | DocBook 5.1.1 |
| Man Page Converter | ✅ Complete | troff/nroff format |
| EPUB Converter | ✅ Complete | EPUB 2.0.1 |
| Attribute Processor | ✅ Complete | Document attribute handling |
| Include Processor | ✅ Complete | Include directive with tag filtering |
| Conditional Processing | ✅ Complete | ifdef, ifndef, ifeval |
| Bibliography | ✅ Complete | Citation processing with [[[label]]] syntax |
| Index Terms | ✅ Complete | (((term))) indexing with flow and concealed terms |
| Compatibility Testing | ✅ Complete | 29/32 tests passing (90.6%) vs Asciidoctor |
| CLI | ✅ Complete | Full Asciidoctor-compatible options |
| LSP Server | ✅ Complete | Diagnostics, symbols, completion, hover, go-to-definition |

### Missing/Incomplete Features (Post-MVP)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Extensions System** | Custom block/inline macros, tree processors | Medium |

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
