# Roadmap

This document outlines future features and implementation priorities for asciidoc-parser-go.

**Current Status:** All core phases (0-13) are complete with 100% Asciidoctor compatibility on 32 built-in test cases. See [README.md](README.md) for completed features.

## Visual Validation Showcases

To ensure visual correctness and provide examples for users, we maintain showcase HTML files that demonstrate specific AsciiDoc features.

### Current Showcases

| Showcase | Status | Description |
|----------|--------|-------------|
| `examples/showcase.adoc` | ✅ Complete | General overview - sections, basic lists, paragraphs |
| `examples/tables-advanced.adoc` | ✅ Complete | All table features - frame, grid, stripes, alignment, etc. |

### Planned Showcases (High Priority)

| Showcase | Description | Features to Demonstrate |
|----------|-------------|-------------------------|
| `inline-formatting.adoc` | Text markup showcase | Bold, italic, monospace, sub/superscript, links (external/inter-document/xrefs), inline images, icons, keyboard shortcuts, buttons, menus |
| `lists.adoc` | All list types | Ordered (arabic, alpha, roman), unordered with nesting, labeled/description lists, checklist tasks, inline and nested combinations |
| `admonitions.adoc` | Callout blocks | Note, tip, warning, caution, important blocks with/without titles, multi-line content, icons |
| `code-blocks.adoc` | Source code | Listing blocks with syntax highlighting, literal blocks, callouts (`<1>`), different languages, line numbering |

### Medium Priority Showcases

| Showcase | Description |
|----------|-------------|
| `blocks-showcase.adoc` | Example, quote (plain vs verse), sidebar, pass-through blocks |
| `document-structure.adoc` | Multiple section levels, preamble, TOC, book vs article doctype |
| `images-and-media.adoc` | Images, diagrams, video embeds |
| `attributes-demo.adoc` | Attribute definitions, references, conditional inclusion |

**Purpose:** These showcases serve both as visual regression tests and user-facing examples. Each should be viewable in a browser to verify correct rendering.

## Implementation Priorities

### High Priority

#### 1. Attribute Substitution
**Status:** Not implemented

**Description:** Replace attribute references with their values throughout the document.

**Required Features:**
- Document attribute definitions - `:name: value`
- Attribute references - `{name}` substituted with value
- Inline attribute setters - `{set:name}value{set}`
- Predefined attributes - `{toc}`, `{sectnums}`, `{revdate}`, etc.
- Attribute unset - `:name!`
- Conditional attribute evaluation - `ifdef::name[]`
- Attribute chaining and dependencies

**Examples:**
```asciidoc
:author: John Doe
:version: 1.0

Written by {author}, version {version}
{set:product}My Product{set}
Using {product}...
```

#### 2. Enhanced LSP Features
**Status:** Basic LSP implemented (diagnostics, symbols, completions, hover, go-to-definition)

**Remaining Features:**
- Semantic tokens - Syntax highlighting for different AsciiDoc constructs
- Code lens - Section references count, attribute references
- Inlay hints - Show implicit attribute values
- Signature help - For macro parameters
- Workspace symbols - Search across all AsciiDoc files
- Rename - Rename sections and update all references
- Document link - Clickable links for includes and xrefs
- Selection range - Smart selection of sections, blocks
- Folding range - Fold sections, blocks, lists

#### 3. Table Improvements
**Current:** Basic tables with column specifications and attributes

**Remaining Features:**
- **Multi-line cells (`+` continuation)** - Lines ending with `+` should continue in same cell with newline separator
  - Example: `| Line 1 +` followed by `| Line 2` should become one cell with "Line 1\nLine 2"
  - The `+` character should be removed from output
  - This requires changes to the table row parser to detect and collect continuation lines
- Cell styles via column specs - `[cols="2*l"]` for literal columns already works
- Vertical table support
- Auto-fit column widths (`[%autowidth]` implemented)`

#### 4. Advanced Inline Parsing
**Current:** Basic inline formatting implemented

**Remaining Features:**
- Triple-plus passthrough - `+++passthrough+++`
- Footnotes - `footnote:id[text]` with note list at bottom
- Icon fonts - `icon:github[]` with font-awesome support
- Index terms - `((term))` and `(((term)))` (partial)

**Footnotes Details:**
- Footnote references - `footnote:note-id[Note text]` or `footnote:note-id[]`
- Numbered footnotes - Automatic numbering and reference list
- Footnote reuse - `footnote:note-id[]` to reference existing note
- Footnote list rendering at document bottom

### Medium Priority

#### 5. More Block Types
**Current:** Example, quote, literal, sidebar blocks implemented

**Remaining Features:**
- **Admonition blocks with custom content** - Block-style syntax `[NOTE]`, `[TIP]`, etc. with title `.Title` and multi-paragraph content
  - Current: Only inline syntax works (`NOTE: text here`)
  - Needed: Support `[NOTE]` block style with title, lists, code blocks inside
- **Source code line numbers** - `linenums` option for listing blocks to show line numbers
  - Syntax: `[source,python,linenums]` or `%linenums` start number
  - Needed: Parse `linenums` attribute and render line numbers in output
- Verse blocks - Poetry/verse with enhanced line break handling
- Quote attribution - Enhanced attribution, cite title support
- Callout lists - Coordinated callouts across code blocks
- Open blocks - `--` delimiter variant

#### 5. Document Preprocessing
**Current:** Basic `ifdef`, `ifndef`, `ifeval` implemented

**Remaining Features:**
- Nested conditionals - `ifdef::attr[ifdef::nested[]]`
- `elsedef` directive
- Document attribute substitution enhancements

#### 6. Extension System Enhancements
**Status:** Basic extension system in `internal/extension/`

**Potential additions:**
- More bundled extensions (diagrams, charts, plantuml, mermaid)
- Extension registry API
- Extension configuration via attributes
- Extension discovery and loading

See [EXTENSIONS.md](EXTENSIONS.md) for current extension capabilities.

### Lower Priority

#### 7. Additional Backends
**Current:** HTML5, PDF, DocBook 5.1.1, Man Page, EPUB

**Potential additions:**
- Text output - Plain text converter
- Markdown - AsciiDoc to Markdown converter
- Slide decks - Reveal.js/remark.js output

#### 8. CLI Enhancements
**Current:** Full CLI with Asciidoctor-compatible options

**Potential additions:**
- Watch mode for auto-regeneration on file changes
- Server mode for multiple document conversion
- Progress bars for large documents
- Dry-run mode (parse without output)
- Config file support (`~/.config/asciidoc/config.toml`)

#### 9. Validation & Linting
- Document structure validation
- Link checking (internal and external)
- Style checking (recommended practices)
- Attribute validation

#### 10. Developer Tools
- AST inspector CLI (`asciidoc ast --format=json document.adoc`)
- Diff tool (`asciidoc diff old.adoc new.adoc`)
- Attribute explorer (`asciidoc attributes document.adoc`)
- Symbol browser (`asciidoc symbols document.adoc`)

#### 11. Performance
- Parallel rendering for large documents
- Streaming output for very large files
- Caching for incremental builds

## Asciidoctor Feature Comparison

| Feature | Asciidoctor | Go Parser | Priority |
|---------|-------------|-----------|----------|
| **LSP Server** | Community only | Basic implemented | Medium |
| **Extension System** | Full API | Partial | Medium |
| **Admonition: Block-style [NOTE]** | Yes | No (inline only) | Medium |
| **Source blocks: Line numbers (linenums)** | Yes | No | Medium |
| **Table: Multiline cells (+ continuation)** | Yes | No | High |
| **Table: Column repeat (3*)** | Yes | Yes | - |
| **Table: Vertical alignment** | As literal | As literal | - |
| **Table: Cell styles** | Via column specs | Via column specs | - |
| **Passthrough macros** | Full | Partial | High |
| **Footnote macro** | Yes | Partial | High |
| **Icon fonts** | Yes | No | Medium |
| **Conditional processing** | Full | Basic | Medium |
| **Custom backends** | Yes | 5 backends | - |
| **Verse blocks** | Yes | Partial | Medium |
| **Quote blocks** | Yes | Partial | Medium |
| **Sidebar blocks** | Yes | Yes | - |
| **Callout lists** | Yes | Partial | High |

## Development Strategy

For each feature implementation:

1. **Find corresponding Asciidoctor test** in `test/` directory
2. **Create input fixture** in `tests/compatibility/fixtures/`
3. **Generate golden file** using Asciidoctor (when available)
4. **Implement feature** in Go parser
5. **Validate** against golden file output

### Asciidoctor Test Files Reference

```
asciidoctor/test/
├── attribute_entries_test.rb    # Attribute processing
├── basic_document_test.rb        # Basic document structure
├── blocks_test.rb                # Block-level elements
├── conditionals_test.rb          # Conditional processing
├── document_test.rb              # Document model
├── examples_test.rb              # Example blocks
├── inline_macro_test.rb          # Inline macros
├── links_test.rb                 # Links and cross-references
├── lists_test.rb                 # Lists (ordered, unordered, labeled)
├── paragraphs_test.rb            # Paragraph formatting
├── parser_test.rb                # Parser core
├── sections_test.rb              # Section headings
├── tables_test.rb                # Table processing
└── template_converter_test.rb   # Converter tests
```

## Notes

- Asciidoctor does **NOT** provide an official LSP server ([GitHub Issue #3630](https://github.com/asciidoctor/asciidoctor/issues/3630))
- The Go parser's LSP implementation is **ahead** of Asciidoctor in this area
- All 32 compatibility tests pass (100% compatibility with Asciidoctor 2.0.26)
- UI macros (`kbd:[...]`, `btn:[...]`, `menu:[...]`) are Asciidoctor extensions, available as custom extensions in this implementation

---

**Last Updated:** 2026-02-19
**Asciidoctor Version:** 2.0.26
**Compatibility:** 32/32 tests passing (100%)
