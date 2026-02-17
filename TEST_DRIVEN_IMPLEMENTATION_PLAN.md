# Test-Driven Implementation Plan

This document outlines features to implement based on actual Asciidoctor test coverage. Each feature identified here has corresponding unit tests in the Asciidoctor codebase that we can use for compatibility validation.

## Asciidoctor Test Files Analyzed

| Test File | Size | Primary Coverage |
|-----------|------|------------------|
| tables_test.rb | 79KB | Table parsing, formats, cell styles |
| links_test.rb | 55KB | Links, xrefs, anchors |
| extensions_test.rb | 76KB | Extension system |
| blocks_test.rb | 143KB | Block types, delimited blocks |
| lists_test.rb | 181KB | List types, nesting |
| sections_test.rb | 116KB | Section headings |
| substitutions_test.rb | 132KB | Text substitutions |
| attributes_test.rb | 57KB | Document attributes |
| paragraphs_test.rb | 19KB | Paragraph formatting |

## Implementation Priority by Feature Area

### Phase 1: Inline Macros (Highest Priority)

**Asciidoctor Source:** `test/extensions_test.rb`, `test/links_test.rb`

| Feature | Syntax | Test Reference | Implementation Status |
|---------|--------|----------------|----------------------|
| Footnote macro | `footnote:text[]` | extensions_test.rb:1800+ | ⚠️ Partial |
| Icon macro | `icon:name[]` | extensions_test.rb | ❌ Missing |
| Inline passthrough | `+text+` | substitutions_test.rb | ⚠️ Partial |
| Verbatim passthrough | `+`+text+`+` | substitutions_test.rb | ⚠️ Partial |
| Triple-plus passthrough | `+++text+++` | substitutions_test.rb | ❌ Missing |
| Menu shortcuts | `kbd:[Ctrl+C]` | extensions_test.rb | ✅ Complete |
| Button macro | `btn:[OK]` | extensions_test.rb | ✅ Complete |
| Index terms | `((term))` | links_test.rb | ❌ Missing |
| Concealed index terms | `(((term)))` | links_test.rb | ❌ Missing |

#### Test Cases from Asciidoctor

**Inline Passthrough Tests:**
```ruby
# From substitutions_test.rb
test 'should not process inline formatting inside double-plus passthrough'
test 'should process substitutions in double-plus passthrough'
test 'should not process inline formatting inside triple-plus passthrough'
```

**Footnote Tests:**
```ruby
# From extensions_test.rb
test 'should parse footnote macro and generate footnotes'
test 'should allow footnote to be referenced multiple times'
```

### Phase 2: Table Enhancements

**Asciidoctor Source:** `test/tables_test.rb`

| Feature | Syntax | Test Reference | Implementation Status |
|---------|--------|----------------|----------------------|
| Multi-line cells | `|line 1` +<br>`+line 2` | tables_test.rb:800+ | ❌ Missing |
| Cell style: literal | `l|cell` | tables_test.rb | ❌ Missing |
| Cell style: verse | `v|cell` | tables_test.rb | ⚠️ Partial |
| Cell style: aspirational | `a|cell` | tables_test.rb | ⚠️ Partial |
| Cell style: example | `e|cell` | tables_test.rb | ⚠️ Partial |
| Cell style: monospace | `m|cell` | tables_test.rb | ❌ Missing |
| Vertical alignment | `.^`, `.>`, `.<.` | tables_test.rb | ⚠️ Partial |
| Auto-width columns | `[%autowidth]` | tables_test.rb | ⚠️ Partial |
| Repeat cells | `3*|value` | tables_test.rb | ❌ Missing |
| Cell background color | `{set:cellbgcolor:red}` | tables_test.rb | ❌ Missing |

#### Test Cases from Asciidoctor

**Multi-line Cell Tests:**
```ruby
# From tables_test.rb
test 'should handle horizontal and vertical source data with blank lines and table header'
test 'should support multi-line table cells'

# Example test input:
input = <<~'EOS'
[width="80%",cols="3,^2,^2,10",options="header"]
|===
|Date |Duration |Avg HR |Notes
|22-Aug-08 |10:24 | 157 |
Worked out MSHR (max sustainable heart rate) by going hard
for this interval.
|===
EOS
```

**Cell Style Tests:**
```ruby
# From tables_test.rb
test 'should format first cell as literal if there is no implicit header row and column has l style'
test 'should format first cell as AsciiDoc if there is no implicit header row and column has a style'
test 'should only substitute specialchars for literal table cells'
test 'should apply cell style for column to repeated content'
```

### Phase 3: Conditional Processing

**Asciidoctor Source:** `test/document_test.rb` (conditional tests)

| Feature | Syntax | Test Reference | Implementation Status |
|---------|--------|----------------|----------------------|
| ifdef directive | `ifdef::attr[]` | document_test.rb | ⚠️ Partial |
| ifndef directive | `ifndef::attr[]` | document_test.rb | ⚠️ Partial |
| ifeval directive | `ifeval::[]` | document_test.rb | ❌ Missing |
| Nested conditionals | `ifdef::attr[ifdef::nested[]]` | document_test.rb | ❌ Missing |
| elsedef directive | `ifdef::attr[content]` | document_test.rb | ❌ Missing |

### Phase 4: Block Types

**Asciidoctor Source:** `test/blocks_test.rb`

| Feature | Syntax | Test Reference | Implementation Status |
|---------|--------|----------------|----------------------|
| Verse block | `[verse]` `----` | blocks_test.rb | ⚠️ Partial |
| Quote block with attribution | `[quote,author,source]` | blocks_test.rb | ⚠️ Partial |
| Sidebar block | `sidebar::[]` | blocks_test.rb | ⚠️ Partial |
| Example block with caption | `.Title` `====` | blocks_test.rb | ⚠️ Partial |
| Literal block variants | `[literal]`, `[verse]` | blocks_test.rb | ⚠️ Partial |
| Open block | `--` delimiter | blocks_test.rb | ⚠️ Partial |

#### Test Cases from Asciidoctor

**Verse Block Tests:**
```ruby
# From blocks_test.rb
test 'should parse verse block'
test 'should preserve newlines in verse block'
test 'should apply substitutions to verse block content'

# Example test input:
input = <<~'EOS'
[verse]
We shall go on till the end.
We shall fight in France.
We shall fight on the seas and oceans.
— Winston Churchill
[quote, author, source]
====
To be or not to be.
====
EOF
```

### Phase 5: Advanced Link Features

**Asciidoctor Source:** `test/links_test.rb`

| Feature | Syntax | Test Reference | Implementation Status |
|---------|--------|----------------|----------------------|
| Inter-document xref | `<<file.adoc#section>>` | links_test.rb | ⚠️ Partial |
| Xref with custom text | `<<id,text>>` | links_test.rb | ✅ Complete |
| Bibliography anchor | `[[[label,text]]]` | links_test.rb | ❌ Missing |
| Link with rel=nofollow | `opts=nofollow` | links_test.rb | ❌ Missing |
| Link attributes | `id`, `title`, `role` | links_test.rb | ❌ Missing |
| Window target | `window=_blank`, `^` | links_test.rb | ❌ Missing |

## Compatibility Testing Strategy

For each feature implementation:

1. **Extract Test Case** from Asciidoctor source
2. **Create Input File** in `tests/compatibility/fixtures/`
3. **Generate Golden File** using Asciidoctor (when available)
4. **Implement Feature** in Go parser
5. **Validate** against golden file output

### Example: Implementing Multi-line Table Cells

**Step 1: Extract Asciidoctor Test**
```ruby
# From tables_test.rb
test 'supports horizontal and vertical source data with blank lines and table header' do
  input = <<~'EOS'
.Horizontal and vertical source data
[width="80%",cols="3,^2,^2,10",options="header"]
|===
|Date |Duration |Avg HR |Notes
|22-Aug-08 |10:24 | 157 |
Worked out MSHR (max sustainable heart rate) by going hard
for this interval.
|===
EOS
  # ... assertions
end
```

**Step 2: Create Compatibility Test**
```go
// In tests/compatibility/compatibility_test.go
func TestCompatibility_TableMultilineCells(t *testing.T) {
  testName := "table_multiline_cells"
  runCompatibilityTest(t, testName)
}
```

**Step 3: Add Input Fixture**
```
In: tests/compatibility/fixtures/table_multiline_cells.adoc
```

**Step 4: Generate Golden (with Asciidoctor installed)**
```bash
GENERATE_GOLDEN=1 go test ./tests/compatibility/...
```

**Step 5: Implement Feature**
```go
// In internal/parser/table.go
func (p *Parser) parseTableContinuation(line string) bool {
    // Detect + continuation
    // Append to current cell
}
```

## Current Project Status Summary

| Component | Status | Next Action |
|-----------|--------|-------------|
| Core Parser | ✅ Complete | - |
| Inline Parsing | ✅ Complete | Add passthrough macros |
| Table Parsing | ⚠️ Basic | Add multi-line cells, styles |
| Link Parsing | ⚠️ Basic | Add attributes, window target |
| Block Parsing | ⚠️ Basic | Enhance verse, quote blocks |
| Conditional Processing | ⚠️ Basic | Add ifeval, nested |
| Compatibility Tests | ✅ Framework | Add test cases |

## Implementation Order Recommendation

1. **Inline Passthrough Macros** (foundational for other features)
   - Triple-plus passthrough `+++text+++`
   - Double-plus passthrough `++text++`
   - Verbatim passthrough

2. **Table Cell Styles** (high-value, well-tested)
   - Literal cells (`l|`)
   - Monospace cells (`m|`)
   - Multi-line cells with `+` continuation

3. **Link Attributes** (commonly used)
   - Window target (`^` suffix)
   - Link attributes (id, title, rel)
   - Bibliography anchors

4. **Enhanced Blocks** (content structure)
   - Verse block improvements
   - Quote with attribution
   - Sidebar styling

5. **Advanced Conditionals** (preprocessing)
   - ifeval expressions
   - Nested conditionals

## Test File Locations

```
tests/compatibility/
├── compatibility_test.go      # Main test runner
├── fixtures/
│   ├── inline_passthrough.adoc
│   ├── table_multiline.adoc
│   ├── table_cell_styles.adoc
│   ├── link_attributes.adoc
│   ├── verse_block.adoc
│   └── ...
└── golden/
    ├── inline_passthrough.html
    ├── table_multiline.html
    └── ...
```

---

**Last Updated:** 2026-02-17
**Based on Asciidoctor Version:** main branch (2026-02)
