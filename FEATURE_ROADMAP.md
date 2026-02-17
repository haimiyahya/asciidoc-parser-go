# Future Features Roadmap

This document outlines potential future features for the asciidoc-parser-go project.

## High Priority (Core Functionality)

### 1. Extension System Enhancements
**Status:** Partially implemented in `internal/extension/`

Potential additions:
- More bundled extensions (diagrams, charts, plantuml, mermaid)
- Extension registry API
- Extension configuration via attributes
- Extension discovery and loading

### 2. Enhanced LSP Features
**Status:** Basic LSP implemented

Remaining LSP features:
- **Semantic tokens** - Syntax highlighting for different AsciiDoc constructs
- **Code lens** - Section references count, attribute references
- **Inlay hints** - Show implicit attribute values
- **Signature help** - For macro parameters
- **Workspace symbols** - Search across all AsciiDoc files
- **Rename** - Rename sections and update all references
- **Document link** - Clickable links for includes and xrefs
- **Selection range** - Smart selection of sections, blocks
- **Folding range** - Fold sections, blocks, lists

### 3. Table Improvements
- Multi-line cells (`+` continuation)
- Cell styles (`a` for aspirational, `e` for example, `m` for monospace)
- Vertical table support
- Auto-fit column widths
- Table column width calculations

## Medium Priority (Enhanced Features)

### 4. Advanced Inline Parsing
- **Passthrough macros** - `+++passthrough+++`, `$$passthrough$$`
- **Inline macros** - `footnote:[]`, `xref:[]`, `indexterm:[]`
- **Icon fonts** - `icon:github[]` with font-awesome support
- **Menu shortcuts** - `kbd:[Ctrl+C]` variations
- **Inline images** - More image macro variations

### 5. More Block Types
- **Verse blocks** - Poetry/verse with line breaks preserved
- **Example blocks** - More formatting options
- **Quote blocks** - Attribution, cite title
- **Sidebar blocks** - Collapsible sidebar content
- **Callout lists** - Coordinated callouts across code blocks

### 6. Document Preprocessing
- **Macro expansion** - User-defined macros
- **Conditional inclusion** - `ifdef::[]`, `ifndef::[]`, `ifeval::[]`
- **Document attributes** - `{set:}`, `{counter:}`
- **Attribute substitution** - More attribute types

### 7. Additional Backends
- **Text output** - Plain text converter
- **Markdown** - AsciiDoc to Markdown converter
- **Slide decks** - Reveal.js/remark.js output

## Lower Priority (Nice to Have)

### 8. CLI Enhancements
- Watch mode for auto-regeneration on file changes
- Server mode for multiple document conversion
- Progress bars for large documents
- Dry-run mode (parse without output)

### 9. Validation & Linting
- Document structure validation
- Link checking (internal and external)
- Style checking (recommended practices)
- Attribute validation

### 10. Developer Tools
- AST inspector CLI (`asciidoc ast --format=json document.adoc`)
- Diff tool (`asciidoc diff old.adoc new.adoc`)
- Attribute explorer (`asciidoc attributes document.adoc`)
- Symbol browser (`asciidoc symbols document.adoc`)

### 11. Performance
- Parallel rendering for large documents
- Streaming output for very large files
- Caching for incremental builds

### 12. Internationalization
- Document attribute localization
- Multi-language document support
- Localized admonition labels

---

**Last Updated:** 2026-02-17
**Current Version:** 0.1.0
