# Asciidoctor Feature Support Analysis

This document compares Asciidoctor's feature support with the asciidoc-parser-go implementation, to guide development of features that have test coverage in Asciidoctor's codebase.

## Asciidoctor Feature Support Summary

| Feature | Asciidoctor Support | Go Parser Status | Priority |
|---------|---------------------|------------------|----------|
| **LSP Server** | ❌ No (community only) | ✅ Complete | - |
| **Extension System** | ✅ Yes | ⚠️ Partial | Medium |
| **Table: Multiline cells** | ✅ Yes | ❌ Missing | High |
| **Table: Vertical alignment** | ✅ Yes | ❌ Missing | High |
| **Table: Cell styles** | ✅ Yes | ❌ Missing | High |
| **Passthrough macros** | ✅ Yes | ⚠️ Partial | High |
| **Footnote macro** | ✅ Yes | ⚠️ Partial | High |
| **Icon fonts** | ✅ Yes | ❌ Missing | Medium |
| **Conditional processing** | ✅ Yes | ⚠️ Partial | High |
| **Custom backends** | ✅ Yes | ✅ Yes (5 backends) | - |
| **Verse blocks** | ✅ Yes | ⚠️ Partial | Medium |
| **Quote blocks** | ✅ Yes | ⚠️ Partial | Medium |
| **Sidebar blocks** | ✅ Yes | ⚠️ Partial | Medium |
| **Callout lists** | ✅ Yes | ⚠️ Partial | High |
| **Validation/Linting** | ⚠️ External tools | ❌ Missing | Low |
| **Watch mode** | ⚠️ External tools | ❌ Missing | Low |

## Features Supported by Asciidoctor (with Test Coverage)

### Inline Macros
Asciidoctor has comprehensive inline macro support:
- **Footnotes** - [footnote:[] macro](https://docs.asciidoctor.org/asciidoc/latest/macros/footnote/)
- **Icon fonts** - [icon:[] macro](https://docs.asciidoctor.org/asciidoc/latest/macros/icon-macro/) with Font Awesome
- **Inline passthroughs** - [+pass+ macro](https://docs.asciidoctor.org/asciidoc/latest/pass/pass-macro/)
- **XRef (cross-reference)** - `<<id>>` syntax
- **Index terms** - `((term))` and `(((term)))` syntax

### Conditional Processing
Full preprocessing directive support:
- **ifdef** - Include if attribute IS set
- **ifndef** - Include if attribute is NOT set
- **ifeval** - Evaluate complex conditions
- [Documentation](https://docs.asciidoctor.org/asciidoc/latest/directives/conditionals/)

### Table Features
Advanced table formatting:
- [Format content by cell](https://docs.asciidoctor.org/asciidoc/latest/tables/format-cell-content/)
- [Align by cell](https://docs.asciidoctor.org/asciidoc/latest/tables/align-by-cell/) (horizontal and vertical)
- [Span columns and rows](https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/)
- Cell styles: `a` (aspirational), `e` (example), `m` (monospace), `l` (literal), `v` (verse)
- Multi-line cells with `+` continuation

### Block Types
- **Verse blocks** - Poetry with line breaks preserved
- **Quote blocks** - With attribution and cite title
- **Sidebar blocks** - Collapsible sidebar content
- **Example blocks** - With caption and title
- **Literal blocks** - Multiple syntax options

### Extensions
[Full extension API](https://docs.asciidoctor.org/asciidoc/latest/extensions/):
- Block macro processors
- Inline macro processors
- Block processors
- Tree processors
- Postprocessors
- Docinfo processors
- Custom converters

## Asciidoctor's LSP Status

Asciidoctor does **NOT** provide an official LSP server:
- [GitHub Issue #3630](https://github.com/asciidoctor/asciidoctor/issues/3630) - Discussion about creating an LSP
- The official VS Code extension uses its own parsing, not LSP
- Community projects exist (e.g., [ViToni/asciidoc-lsp](https://github.com/ViToni/asciidoc-lsp))

**This is an area where asciidoc-parser-go is AHEAD of Asciidoctor!**

## Implementation Priority Based on Asciidoctor Test Coverage

### Phase 1: Inline Macros (Highest Priority)
Asciidoctor has extensive tests for inline macros in:
- `test/inline_macro_test.rb`
- `test/paragraphs_test.rb`
- `test/links_test.rb`

Features to implement:
1. Footnote macro with `footnote:[]` syntax
2. Icon macro with `icon:name[]` syntax
3. Inline passthrough `+text+`
4. Verbatim passthrough `+`+text+`+`
5. Single-line passthrough `$$text$$`

### Phase 2: Table Enhancements
Asciidoctor tests in:
- `test/tables_test.rb`

Features to implement:
1. Multi-line cells with `+` continuation
2. Cell styles (a, e, m, l, v, h, d)
3. Vertical alignment (`<top`, `<bottom`, `<middle`)
4. Automatic column width

### Phase 3: Conditional Processing
Asciidoctor tests in:
- `test/conditionals_test.rb`
- `test/attribute_entries_test.rb`

Features to implement:
1. `ifdef::[]` directive
2. `ifndef::[]` directive
3. `ifeval::[]` directive with expressions
4. Nested conditionals
5. `elsifdef` and `elsedef` (if supported)

### Phase 4: Block Types
Asciidoctor tests in:
- `test/blocks_test.rb`

Features to implement:
1. Verse block formatting
2. Quote block with attribution
3. Sidebar block styling
4. Example block with caption

## Development Strategy

For each feature:
1. Find corresponding Asciidoctor test file
2. Identify test cases covering the feature
3. Run Asciidoctor to get expected output
4. Implement feature in Go parser
5. Use compatibility test framework to verify

## Asciidoctor Test Files Reference

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

## Sources

- [Asciidoctor Extensions](https://docs.asciidoctor.org/asciidoc/latest/extensions/)
- [Inline Passthroughs](https://docs.asciidoctor.org/asciidoc/latest/pass/pass-macro/)
- [Font Icons Mode](https://docs.asciidoctor.org/asciidoc/latest/macros/icons-font/)
- [Footnotes](https://docs.asciidoctor.org/asciidoc/latest/macros/footnote/)
- [Conditionals](https://docs.asciidoctor.org/asciidoc/latest/directives/conditionals/)
- [ifdef/ifnedef Directives](https://docs.asciidoctor.org/asciidoc/latest/directives/ifdef-ifndef/)
- [ifeval Directive](https://docs.asciidoctor.org/asciidoc/latest/directives/ifeval/)
- [Table Formatting](https://docs.asciidoctor.org/asciidoc/latest/tables/format-cell-content/)
- [Table Alignment](https://docs.asciidoctor.org/asciidoc/latest/tables/align-by-cell/)
- [Table Cell Spanning](https://docs.asciidoctor.org/asciidoc/latest/tables/span-cells/)
- [LSP Discussion #3630](https://github.com/asciidoctor/asciidoctor/issues/3630)
- [ViToni/asciidoc-lsp](https://github.com/ViToni/asciidoc-lsp)

---

**Last Updated:** 2026-02-17
