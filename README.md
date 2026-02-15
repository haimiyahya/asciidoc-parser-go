# asciidoc-parser-go

A native Go implementation of an [AsciiDoc](https://asciidoc.org/) parser and processor, targeting near-full compliance with the [Eclipse AsciiDoc Language Specification](https://gitlab.eclipse.org/asciidoc-lang/asciidoc-lang) while maintaining compatibility with [Asciidoctor](https://asciidoctor.org/).

## Features

- **Line-oriented Reader** - Mimics human visual scanning of AsciiDoc markup
- **Block Classifier** - Identifies paragraphs, lists, sections, delimited blocks, tables, etc.
- **Inline Parser** - Handles bold, italic, monospace, links, images, etc.
- **AST Builder** - Rich Abstract Syntax Tree preserving source locations
- **HTML5 Converter** - Produces semantic HTML5 via visitor pattern
- **Extensible Design** - Interface-based architecture for custom converters/processors

## Installation

```bash
go get github.com/haimiyahya/asciidoc-parser-go
```

## Quick Start

### Parsing a Document

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

### Converting to HTML5

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

### Inline Markup

| Syntax | Output | Notes |
|--------|--------|-------|
| `**bold**` | **bold** | Constrained bold |
| `*word*` | *word* | Unconstrained bold (single word) |
| `__italic__` | __italic__ | Constrained italic |
| `_word_` | _word_ | Unconstrained italic (single word) |
| `` `code` ` | `code` | Monospace |
| `++code++` | `++code++` | Monospace alternative |
| `link:text[url]` | <a href="url">text</a> | Macro link |
| `https://url` | <a href="url">url</a> | Bare URL |
| `image:path[alt]` | <img src="path" alt="alt"> | Inline image |

### Admonitions

```
NOTE: This is important.
TIP: Try this approach.
WARNING: Be careful!
CAUTION: Watch out!
IMPORTANT: Pay attention.
```

### Tables

```
| Col 1 | Col 2 |
| Cell 1 | Cell 2 |
```

## Architecture

```
internal/
├── ast/          # Abstract Syntax Tree node definitions
├── blocks/       # Block-level AST node types
├── converter/    # Output converters (HTML5, extensible)
├── inline/       # Inline markup parser
├── parser/       # Main parser orchestrator
├── processor/    # Attribute & conditional processor
└── reader/       # Line-oriented reader with classifier
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
[Converter] → HTML5 (or other format)
```

## Roadmap

- [x] Phase 0: Line-Oriented Reader + Basic Block Classifier
- [x] Phase 1: Core Parser Block Detection
- [x] Phase 2: Inline Parser
- [x] Phase 3: AST Builder
- [x] Phase 4: Attribute Processor
- [x] Phase 5: Include Processor
- [x] Phase 6: HTML5 Converter
- [ ] Phase 7: Testing and Validation
- [ ] Phase 8: CLI and Tooling

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
| Table Parsing | ✅ Complete | Basic tables |
| Inline Parsing | ✅ Complete | Bold, italic, monospace, links, images |
| Admonitions | ✅ Complete | All 5 types |
| Block Macros | ✅ Complete | Image, video, audio |
| Delimited Blocks | ✅ Complete | Example, quote, literal, etc. |
| AST Builder | ✅ Complete | Rich node hierarchy |
| HTML5 Converter | ✅ Complete | Semantic HTML5 |
| Attribute Processor | 🚧 In Progress | Basic attribute handling |
| Include Processor | 🚧 In Progress | Basic include directive |
| CLI | 📝 Planned | Phase 8 |

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
