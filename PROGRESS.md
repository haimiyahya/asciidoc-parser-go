# AsciiDoc Parser - Progress Tracking

This file tracks resolved issues and remaining work for Asciidoctor compatibility.

## Date: 2026-02-16 (Session 3)

### ✅ Newly Resolved Issues

#### 8. Table Structure and Parsing (2 tests)
**Problem**:
- Table parser was incorrectly treating `|=` as a header row indicator
- Table converter was missing `frame-all grid-all stretch` classes
- Empty colgroup was being output instead of proper col tags with widths
- Using `thead`/`th` when basic tables should use `tbody`/`td`

**Solution**:
- Removed `|=` header row detection from parser (it's per-cell, not per-row)
- Updated converter to always include `frame-all grid-all stretch` classes
- Fixed colgroup to use actual column count from rows
- Changed to use `tbody` with `td` for all rows in basic tables

**Files Modified**:
- `internal/parser/table.go` - Removed `|=` header row detection
- `internal/converter/html5.go` - Fixed `convertTable()` function

#### 9. Index Terms Visibility (2 tests)
**Problem**:
- Flow index terms `((text))` were being wrapped in `<span>` elements
- Concealed index terms `(((text)))` were outputting empty `<span>` elements
- Asciidoctor outputs plain text for flow and nothing for concealed

**Solution**:
- Changed flow index terms to output plain text only (no span wrapper)
- Changed concealed index terms to produce no output at all

**Files Modified**:
- `internal/converter/html5.go` - Fixed `convertInlineNode()` for `NodeIndexTerm`

#### 10. Bibliography Structure (1 test)
**Problem**:
- Section ID not getting underscore prefix
- Missing `sect1`/`sectionbody` wrapper divs
- Using simple divs instead of `ul.bibliography` with `li` elements
- Missing `<a id="">` anchors
- Not using xref text for display label

**Solution**:
- Added special handling for bibliography section IDs to always add underscore prefix
- Rewrote `convertBibliography()` to output proper Asciidoctor structure
- Added `sect1`/`sectionbody` wrapper divs
- Changed to use `ul.bibliography` with `li` elements
- Added `<a id="">` anchors and xref text support

**Files Modified**:
- `internal/parser/parser.go` - Added underscore prefix for bibliography IDs
- `internal/converter/html5.go` - Rewrote `convertBibliography()` and `convertBibliographyEntry()`

#### 11. Unconstrained Bold/Italic Multi-Word Support
**Problem**:
- Unconstrained bold `*text*` and italic `_text_` only worked for single words
- Asciidoctor allows multi-word phrases with unconstrained syntax

**Solution**:
- Removed `isWord()` constraint from unconstrained bold and italic parsing
- Now supports `*multi word phrase*` and `_multi word phrase_`

**Files Modified**:
- `internal/inline/inline.go` - Updated `tryBold()` and `tryItalic()` functions

### 📊 Test Status

**Integration Tests**: 28/28 PASSING ✅

**Compatibility Tests**: 32/32 PASSING ✅

- **PASSING**:
  - basic (paragraphs, document-title, sections, text-formatting)
  - lists (unordered, ordered, labeled)
  - inline (links, images, monospace, superscript-subscript)
  - admonitions (note, tip, warning)
  - blocks (literal, example, quote, sidebar)
  - tables (basic, with-header)
  - indexterms (flow, concealed)
  - bibliography (basic)
  - ui (keyboard, button, menu)
  - roles (basic)

### 🔧 Remaining Issues

**None! All compatibility tests are passing.** 🎉

### 📝 Notes

- All 32 compatibility tests now match Asciidoctor output exactly
- The parser now supports all major AsciiDoc features including:
  - Block types (literal, example, quote, sidebar)
  - Lists (unordered, ordered, labeled)
  - Tables (basic, with headers)
  - Inline markup (bold, italic, monospace, links, images, superscript, subscript)
  - Index terms (flow and concealed)
  - Bibliography sections
  - UI macros (keyboard, button, menu)
  - Admonitions
  - Cross-references
  - Conditional includes

