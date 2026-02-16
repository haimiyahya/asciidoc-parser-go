# AsciiDoc Parser - Progress Tracking

This file tracks resolved issues and remaining work for Asciidoctor compatibility.

## Date: 2026-02-16 (Session 2)

### ✅ Newly Resolved Issues

#### 5. Block Delimiter Parsing Bug
**Problem**: Sidebar blocks (`****`) were being parsed as `listingblock` instead of `sidebarblock` because `createDelimitedBlock()` didn't have a case for `BlockSidebar`.
**Solution**: Added `BlockSidebar` case to `createDelimitedBlock()` function, creating `*ast.NodeBlock` with delimiter `*`.
**Files Modified**:
- `internal/parser/parser.go` - Added sidebar case to `createDelimitedBlock()`

#### 6. Block Converter Structure Fix
**Problem**: Blocks were using full delimiter strings (`====`, `____`) for comparison, but the parser stores single characters (`=`, `_`).
**Solution**: Updated `convertBlock()` to use single character comparisons (`=`, `_`, `*`, `/`, `-`).
**Files Modified**:
- `internal/converter/html5.go` - Fixed delimiter comparisons in `convertBlock()`

#### 7. Newline Escape Sequence Bug
**Problem**: Using backtick strings with actual newlines was outputting literal `\n` escape sequences instead of newlines.
**Solution**: Changed to use `fmt.Fprint(w, "\n")` for explicit newlines instead of backtick strings.
**Files Modified**:
- `internal/converter/html5.go` - Fixed `convertBlock()`, `convertLiteral()`, `convertParagraph()`, `convertSection()`, `convertList()`, `convertAdmonition()`

### 📊 Test Status

**Integration Tests**: 28/28 PASSING ✅

**Compatibility Tests**: 27/32 PASSING (was 23/32, +4 this session)
- **NEWLY PASSING**: `blocks/literal`, `blocks/example`, `blocks/quote`, `blocks/sidebar`
- **PASSING**: basic formatting, lists, inline markup, admonitions, UI macros, roles
- **FAILING (5)**: See "Remaining Issues" below

### 🔧 Remaining Issues (5 tests)

These features need more implementation work to match Asciidoctor's exact output:

1. **Tables** (2 failures)
   - `tables/basic` - Table structure needs work (frame-all grid-all classes, colgroup, tbody vs thead)
   - `tables/with-header` - Header row parsing and structure

2. **Index Terms** (2 failures)
   - `indexterms/flow` - `((flow index term))` should be invisible (text removed from output)
   - `indexterms/concealed` - `(((hidden, term)))` should be invisible (empty span, text removed)

3. **Bibliography** (1 failure)
   - `bibliography/basic` - Bibliography structure needs work

### 📝 Notes

- All block types now render correctly with proper structure and newlines
- Table converter needs significant refactoring to match Asciidoctor's complex table structure
- Index terms require inline parser changes to make the text invisible while keeping the HTML markers

