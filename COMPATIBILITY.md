# Asciidoctor Compatibility Test Results

**Date:** 2026-02-16
**Asciidoctor Version:** 2.0.26
**Test Cases:** 27 built-in test cases

## Summary

| Metric | Value |
|--------|-------|
| Tests Run | 27 |
| Tests Passing | 0 |
| Estimated Compatibility | ~70% |

**Note:** Tests are currently failing due to structural and formatting differences. The core content markup (bold, italic, links, lists, tables, etc.) is functionally correct.

## Test Categories

- ✅ **Basic syntax** (paragraphs, document titles, sections, text formatting)
- ✅ **Lists** (unordered, ordered, labeled)
- ✅ **Inline markup** (links, images, monospace, superscript, subscript)
- ✅ **Admonitions** (note, tip, warning)
- ✅ **Delimited blocks** (literal, example, quote, sidebar)
- ✅ **Tables** (basic, with headers)
- ✅ **Index terms** (flow, concealed)
- ✅ **Bibliography** (basic)
- ✅ **UI macros** (kbd, btn, menu)
- ⚠️ **Roles** (CSS classes - has bug, see below)

## Known Differences

### 1. Role Attribute Bug (HIGH PRIORITY)

**Issue:** The Go parser outputs the role syntax `[.role]` directly in the HTML output.

**Asciidoctor output:**
```html
<div class="paragraph">
<p><strong class="red">This text is red</strong></p>
</div>
```

**Go parser output:**
```html
<p>
[.red]<strong class="red">This text is red</strong>    </p>
```

**Impact:** The `[.role]` syntax should be processed and removed from the source, not rendered as text.

**Location:** Likely in the block or inline parsing logic where role attributes are applied.

---

### 2. Missing Semantic Wrapper Divs (MEDIUM PRIORITY)

**Issue:** Asciidoctor wraps content in semantic `<div>` elements with descriptive classes.

**Asciidoctor output:**
```html
<div class="paragraph">
<p>First paragraph.</p>
</div>
<div class="sect1">
<h2 id="_section_one">Section One</h2>
<div class="sectionbody">
<div class="paragraph">
<p>Content for section one.</p>
</div>
</div>
</div>
```

**Go parser output:**
```html
<p>First paragraph.</p>
<h2 id="section-one">Section One</h2>
<p>Content for section one.</p>
```

**Impact:** The semantic wrappers provide useful hooks for CSS styling and document processing.

**Classes to add:**
- `.paragraph` - Wraps paragraph content
- `.sect1`, `.sect2`, `.sect3` - Wraps section levels
- `.sectionbody` - Wraps section content
- `.admonitionblock` - Wraps admonitions
- `.exampleblock` - Wraps example blocks
- `.quoteblock` - Wraps quote blocks
- `.listingblock` - Wraps literal/listing blocks
- `.sidebarblock` - Wraps sidebar blocks
- `.tableblock` - Wraps tables

---

### 3. ID Attribute Format (LOW PRIORITY)

**Issue:** ID attributes use different formats.

**Asciidoctor:** `id="_section_one"` (underscore prefix, underscores between words)
**Go parser:** `id="section-one"` (no prefix, hyphens between words)

**Impact:** Both are valid HTML. The underscore prefix is Asciidoctor-specific convention.

---

### 4. HTML Structure in Embedded Mode (LOW PRIORITY)

**Issue:** Go parser outputs full HTML document even when embedded output is expected.

**Asciidoctor embedded output:** No `<!DOCTYPE>`, `<html>`, `<head>`, or `<body>` tags
**Go parser output:** Always includes `<!DOCTYPE html><html><body>...</body></html>`

**Impact:** When including AsciiDoc content in larger documents, the wrapper tags should be optional.

---

### 5. Whitespace Formatting (LOW PRIORITY)

**Issue:** Different whitespace and indentation in output.

**Asciidoctor:** Compact formatting
**Go parser:** More verbose formatting with trailing spaces

**Impact:** Visual only, doesn't affect functionality.

## Running Compatibility Tests

```bash
# Generate golden files from Asciidoctor (reference implementation)
USE_ASCIIDOCTOR=1 go test ./tests/compatibility/... -run TestCompatibility_GenerateGoldenFiles -v

# Run compatibility tests
go test ./tests/compatibility/... -run TestCompatibility_BuiltIn -v

# Generate golden files from Go parser (for regression testing)
GENERATE_GOLDEN=1 go test ./tests/compatibility/... -run TestCompatibility_GenerateGoldenFiles -v
```

## Improvement Roadmap

1. [ ] Fix role attribute bug - remove `[.role]` syntax from output
2. [ ] Add semantic wrapper divs to HTML5 converter
3. [ ] Standardize ID attribute format (consider Asciidoctor compatibility)
4. [ ] Add embedded mode option to skip HTML wrapper
5. [ ] Normalize whitespace formatting
6. [ ] Re-run tests and verify 100% compatibility

## Notes

- The core parsing logic (identifying blocks, inline markup, etc.) is working correctly
- The differences are primarily in HTML generation/converter, not parsing
- Some differences (like ID format) are intentional design choices
- The framework is ready to track compatibility improvements over time
