// Package reader provides tests for the AsciiDoc reader components.
package reader

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewReader tests creating new Reader instances.
func TestNewReader(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr error
		wantLen int
	}{
		{
			name:    "simple text",
			source:  "Hello World",
			wantErr: nil,
			wantLen: 1,
		},
		{
			name:    "multi-line text",
			source:  "Line 1\nLine 2\nLine 3",
			wantErr: nil,
			wantLen: 3,
		},
		{
			name:    "empty source",
			source:  "",
			wantErr: ErrEmptySource,
			wantLen: 0,
		},
		{
			name:    "text with trailing newline",
			source:  "Hello\nWorld\n",
			wantErr: nil,
			wantLen: 2, // ["Hello", "World"]
		},
		{
			name:    "preserves trailing empty lines",
			source:  "Line 1\nLine 2\n\n",
			wantErr: nil,
			wantLen: 3, // ["Line 1", "Line 2", ""]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewReader(tt.source)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, r)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, r)
				assert.Equal(t, tt.wantLen, len(r.sourceLines))
			}
		})
	}
}

// TestReaderHasMoreLines tests the HasMoreLines method.
func TestReaderHasMoreLines(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		advance  int
		wantMore bool
		skip     bool
	}{
		{
			name:     "has lines initially",
			source:   "Line 1\nLine 2",
			advance:  0,
			wantMore: true,
		},
		{
			name:     "empty after reading all",
			source:   "Line 1\nLine 2",
			advance:  2,
			wantMore: false,
		},
		{
			name:     "empty reader",
			source:   "",
			advance:  0,
			wantMore: false,
			skip:     true, // NewReader returns error for empty source
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("skipped in test data")
			}

			r, err := NewReader(tt.source)
			require.NoError(t, err)

			for i := 0; i < tt.advance; i++ {
				r.Advance()
			}

			assert.Equal(t, tt.wantMore, r.HasMoreLines())
		})
	}
}

// TestReaderNextLine tests the NextLine method.
func TestReaderNextLine(t *testing.T) {
	t.Run("reads lines in order", func(t *testing.T) {
		r, err := NewReader("Line 1\nLine 2\nLine 3")
		require.NoError(t, err)

		lines := []string{}
		for r.HasMoreLines() {
			lines = append(lines, r.NextLine())
		}

		assert.Equal(t, []string{"Line 1", "Line 2", "Line 3"}, lines)
	})

	t.Run("returns empty string at end", func(t *testing.T) {
		r, err := NewReader("Single line")
		require.NoError(t, err)

		r.NextLine() // Consume the only line
		assert.Equal(t, "", r.NextLine())
	})
}

// TestReaderPeekLine tests the PeekLine method.
func TestReaderPeekLine(t *testing.T) {
	t.Run("peeks without consuming", func(t *testing.T) {
		r, err := NewReader("Line 1\nLine 2")
		require.NoError(t, err)

		first := r.PeekLine()
		second := r.PeekLine()

		assert.Equal(t, "Line 1", first)
		assert.Equal(t, "Line 1", second) // Same line
		assert.Equal(t, 1, r.lineno)      // Line number unchanged (peek doesn't advance)
	})

	t.Run("then consumed after NextLine", func(t *testing.T) {
		r, err := NewReader("Line 1\nLine 2")
		require.NoError(t, err)

		r.PeekLine()
		r.NextLine()
		assert.Equal(t, "Line 2", r.PeekLine())
	})

	t.Run("returns empty at end", func(t *testing.T) {
		r, err := NewReader("Single line")
		require.NoError(t, err)

		assert.Equal(t, "", r.PeekLine())
	})
}

// TestReaderSkipBlankLines tests the SkipBlankLines method.
func TestReaderSkipBlankLines(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantSkipped int
		wantNext    string
	}{
		{
			name:        "skips leading blanks",
			source:      "\n\n\nText",
			wantSkipped: 2,
			wantNext:    "Text",
		},
		{
			name:        "skips whitespace lines",
			source:      "   \n  \t\nText",
			wantSkipped: 3,
			wantNext:    "Text",
		},
		{
			name:        "no blanks to skip",
			source:      "Text",
			wantSkipped: 0,
			wantNext:    "Text",
		},
		{
			name:        "stops at non-blank",
			source:      "Line 1\nLine 2\nLine 3\n",
			wantSkipped: 2,
			wantNext:    "Line 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewReader(tt.source)
			require.NoError(t, err)

			skipped := r.SkipBlankLines()

			assert.Equal(t, tt.wantSkipped, skipped)
			if r.HasMoreLines() {
				assert.Equal(t, tt.wantNext, r.PeekLine())
			}
		})
	}
}

// TestReaderMarkAndRestore tests the Mark and Restore functionality.
func TestReaderMarkAndRestore(t *testing.T) {
	t.Run("saves and restores state", func(t *testing.T) {
		r, err := NewReader("Line 1\nLine 2\nLine 3")
		require.NoError(t, err)

		// Read first line
		r.NextLine()
		assert.Equal(t, 2, r.lineno)

		// Mark position
		r.Mark()

		// Read second line
		r.NextLine()
		assert.Equal(t, 3, r.lineno)

		// Restore to mark
		restored := r.Restore()

		assert.True(t, restored)
		assert.Equal(t, 2, r.lineno)
		assert.Equal(t, "Line 2", r.PeekLine())
	})

	t.Run("restore without mark returns false", func(t *testing.T) {
		r, err := NewReader("Line 1\nLine 2")
		require.NoError(t, err)

		assert.False(t, r.Restore())
	})
}

// TestReaderUnshiftLine tests the UnshiftLine method.
func TestReaderUnshiftLine(t *testing.T) {
	t.Run("puts line back", func(t *testing.T) {
		r, err := NewReader("Line 1\nLine 2")
		require.NoError(t, err)

		first := r.NextLine()
		assert.Equal(t, "Line 1", first)

		// Put it back
		r.UnshiftLine(first)
		assert.Equal(t, "Line 1", r.PeekLine())
		assert.Equal(t, 1, r.lineno) // Line number decremented
	})
}

// TestPosition tests the Position type.
func TestPosition(t *testing.T) {
	tests := []struct {
		pos      Position
		expected string
	}{
		{
			pos:      Position{Path: "test.adoc", Lineno: 42},
			expected: "test.adoc:42",
		},
		{
			pos:      Position{Lineno: 10},
			expected: "line 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pos.String())
		})
	}
}

// TestPrepareLines tests the prepareLines function.
func TestPrepareLines(t *testing.T) {
	t.Run("splits on newline", func(t *testing.T) {
		lines := prepareLines("a\nb\nc\n")
		assert.Equal(t, []string{"a", "b", "c", ""}, lines)
	})

	t.Run("strips trailing whitespace", func(t *testing.T) {
		lines := prepareLines("a  \n  b\t\nc")
		assert.Equal(t, []string{"a", "b", "c"}, lines)
	})

	t.Run("handles empty string", func(t *testing.T) {
		lines := prepareLines("")
		assert.Equal(t, []string{}, lines)
	})
}

// TestSkipFrontMatter tests the SkipFrontMatter function.
func TestSkipFrontMatter(t *testing.T) {
	t.Run("skips YAML front matter", func(t *testing.T) {
		lines := []string{
			"---",
			"title: Test",
			"---",
			"Content here",
		}

		front, remaining := SkipFrontMatter(lines)

		assert.Equal(t, []string{"title: Test"}, front)
		assert.Equal(t, []string{"Content here"}, remaining)
	})

	t.Run("skips TOML front matter", func(t *testing.T) {
		lines := []string{
			"+++",
			"title = \"Test\"",
			"+++",
			"Content here",
		}

		front, remaining := SkipFrontMatter(lines)

		assert.Equal(t, []string{"title = \"Test\""}, front)
		assert.Equal(t, []string{"Content here"}, remaining)
	})

	t.Run("returns nil when no front matter", func(t *testing.T) {
		lines := []string{"No front matter", "Just content"}

		front, remaining := SkipFrontMatter(lines)

		assert.Nil(t, front)
		assert.Equal(t, lines, remaining)
	})

	t.Run("handles unterminated front matter", func(t *testing.T) {
		lines := []string{
			"---",
			"Content without closing",
		}

		front, remaining := SkipFrontMatter(lines)

		assert.Nil(t, front)
		assert.Equal(t, lines, remaining)
	})
}

// TestCountLeadingEmptyLines tests the CountLeadingEmptyLines function.
func TestCountLeadingEmptyLines(t *testing.T) {
	tests := []struct {
		lines    []string
		expected int
	}{
		{
			lines:    []string{"", "", "text"},
			expected: 2,
		},
		{
			lines:    []string{"text"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, CountLeadingEmptyLines(tt.lines))
		})
	}
}

// TestTrimLeadingEmptyLines tests the TrimLeadingEmptyLines function.
func TestTrimLeadingEmptyLines(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			input:    []string{"", "", "text"},
			expected: []string{"text"},
		},
		{
			input:    []string{"text"},
			expected: []string{"text"},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expected, TrimLeadingEmptyLines(tt.input))
		})
	}
}

// TestNewReaderWithOptions tests creating Reader with options.
func TestNewReaderWithOptions(t *testing.T) {
	t.Run("with file option", func(t *testing.T) {
		r, err := NewReader("test",
			WithFile("path/to/test.adoc"),
		)

		require.NoError(t, err)
		assert.Equal(t, "path/to/test.adoc", r.file)
		assert.Equal(t, "test.adoc", r.path)
	})

	t.Run("with starting line", func(t *testing.T) {
		r, err := NewReader("test",
			WithStartingLine(42),
		)

		require.NoError(t, err)
		assert.Equal(t, 42, r.lineno)
	})

	t.Run("without process lines", func(t *testing.T) {
		r, err := NewReader("test",
			WithProcessLines(false),
		)

		require.NoError(t, err)
		assert.False(t, r.processLines)
	})
}

// TestScanner tests the Scanner type.
func TestScanner(t *testing.T) {
	t.Run("scans lines", func(t *testing.T) {
		src := strings.NewReader("line1\nline2\nline3")
		scanner := NewScanner(src)

		lines := []string{}
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		assert.NoError(t, scanner.Err())
		assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
		assert.Equal(t, 3, scanner.Lineno())
	})

	t.Run("with file metadata", func(t *testing.T) {
		src := strings.NewReader("test")
		scanner := NewScannerFile(src, "test.adoc", "test.adoc")

		scanner.Scan()

		assert.Equal(t, "test.adoc", scanner.File())
		assert.Equal(t, 1, scanner.Lineno())

		pos := scanner.Position()
		assert.Equal(t, "test.adoc", pos.Path)
		assert.Equal(t, 1, pos.Lineno)
	})
}

// TestLineClassifierClassify tests the basic Classify method.
func TestLineClassifierClassify(t *testing.T) {
	lc := NewLineClassifier()

	tests := []struct {
		name     string
		line     string
		wantType BlockType
	}{
		{
			name:     "blank line",
			line:     "",
			wantType: BlockBlank,
		},
		{
			name:     "whitespace only",
			line:     "   \t  ",
			wantType: BlockBlank,
		},
		{
			name:     "single-line comment",
			line:     "// This is a comment",
			wantType: BlockComment,
		},
		{
			name:     "document title",
			line:     "= Document Title",
			wantType: BlockSection,
		},
		{
			name:     "section level 1",
			line:     "== Section Title",
			wantType: BlockSection,
		},
		{
			name:     "unordered list with dash",
			line:     "- Item",
			wantType: BlockListUnordered,
		},
		{
			name:     "unordered list with asterisk",
			line:     "* Item",
			wantType: BlockListUnordered,
		},
		{
			name:     "ordered list",
			line:     ". First item",
			wantType: BlockListOrdered,
		},
		{
			name:     "labeled list",
			line:     "term :: definition",
			wantType: BlockListLabeled,
		},
		{
			name:     "attribute entry",
			line:     ":author: John Doe",
			wantType: BlockAttribute,
		},
		{
			name:     "horizontal rule",
			line:     "---",
			wantType: BlockHorizontalRule,
		},
		{
			name:     "page break",
			line:     "<<<",
			wantType: BlockPageBreak,
		},
		{
			name:     "block macro",
			line:     "image::path.png[]",
			wantType: BlockMacro,
		},
		{
			name:     "admonition NOTE",
			line:     "NOTE:",
			wantType: BlockAdmonition,
		},
		{
			name:     "block anchor",
			line:     "[[id]]",
			wantType: BlockAnchor,
		},
		{
			name:     "regular paragraph",
			line:     "This is regular text.",
			wantType: BlockParagraph,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lc.Classify(tt.line)
			assert.Equal(t, tt.wantType, got)
		})
	}
}

// TestLineClassifierDelimitedBlocks tests delimited block detection.
func TestLineClassifierDelimitedBlocks(t *testing.T) {
	lc := NewLineClassifier()

	tests := []struct {
		name     string
		line     string
		wantType BlockType
	}{
		{
			name:     "verbatim block",
			line:     "----",
			wantType: BlockVerbatim,
		},
		{
			name:     "example block",
			line:     "====",
			wantType: BlockExample,
		},
		{
			name:     "quote block",
			line:     "____",
			wantType: BlockQuote,
		},
		{
			name:     "passthrough block",
			line:     "++++",
			wantType: BlockPassthrough,
		},
		{
			name:     "sidebar block",
			line:     "****",
			wantType: BlockSidebar,
		},
		{
			name:     "comment block",
			line:     "////",
			wantType: BlockCommentBlock,
		},
		{
			name:     "literal block",
			line:     "....",
			wantType: BlockLiteral,
		},
		{
			name:     "table block",
			line:     "|===",
			wantType: BlockTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lc.Classify(tt.line)
			assert.Equal(t, tt.wantType, got)
		})
	}
}

// TestLineClassifierSections tests section header parsing.
func TestLineClassifierSections(t *testing.T) {
	lc := NewLineClassifier()

	tests := []struct {
		name      string
		line      string
		wantLevel int
		wantTitle string
	}{
		{
			name:      "document title",
			line:      "= My Document",
			wantLevel: 0,
			wantTitle: "My Document",
		},
		{
			name:      "level 1 section",
			line:      "== Section One",
			wantLevel: 1,
			wantTitle: "Section One",
		},
		{
			name:      "level 2 section",
			line:      "=== Subsection",
			wantLevel: 2,
			wantTitle: "Subsection",
		},
		{
			name:      "level 3 section",
			line:      "==== Sub-sub",
			wantLevel: 3,
			wantTitle: "Sub-sub",
		},
		{
			name:      "max level caps at 6",
			line:      "========= Deep",
			wantLevel: 6,
			wantTitle: "Deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := lc.ClassifyLine(tt.line)

			assert.Equal(t, BlockSection, classification.Type)
			assert.NotNil(t, classification.Section)
			assert.Equal(t, tt.wantLevel, classification.Section.Level)
			assert.Equal(t, tt.wantTitle, classification.Section.Title)
		})
	}
}

// TestLineClassifierAttributes tests attribute entry parsing.
func TestLineClassifierAttributes(t *testing.T) {
	lc := NewLineClassifier()

	t.Run("simple attribute", func(t *testing.T) {
		classification := lc.ClassifyLine(":name: value")

		assert.Equal(t, BlockAttribute, classification.Type)
		assert.NotNil(t, classification.Attribute)
		assert.Equal(t, "name", classification.Attribute.Name)
		assert.Equal(t, "value", classification.Attribute.Value)
		assert.True(t, classification.Attribute.IsSet)
	})

	t.Run("attribute without value", func(t *testing.T) {
		classification := lc.ClassifyLine(":name:")

		assert.Equal(t, BlockAttribute, classification.Type)
		assert.Equal(t, "name", classification.Attribute.Name)
		assert.Equal(t, "", classification.Attribute.Value)
	})
}

// TestLineClassifierLists tests list item detection.
func TestLineClassifierLists(t *testing.T) {
	lc := NewLineClassifier()

	t.Run("unordered dash", func(t *testing.T) {
		classification := lc.ClassifyLine("- Item text")

		assert.Equal(t, BlockListUnordered, classification.Type)
		assert.Equal(t, "-", classification.List.Marker)
		assert.Equal(t, "Item text", classification.List.Text)
	})

	t.Run("unordered asterisk", func(t *testing.T) {
		classification := lc.ClassifyLine("* Item")

		assert.Equal(t, BlockListUnordered, classification.Type)
		assert.Equal(t, "*", classification.List.Marker)
		assert.Equal(t, "Item", classification.List.Text)
	})

	t.Run("ordered single dot", func(t *testing.T) {
		classification := lc.ClassifyLine(". First item")

		assert.Equal(t, BlockListOrdered, classification.Type)
		assert.Equal(t, ".", classification.List.Marker)
		assert.Equal(t, 1, classification.List.Ordinal)
		assert.Equal(t, "First item", classification.List.Text)
	})

	t.Run("ordered double dot", func(t *testing.T) {
		classification := lc.ClassifyLine(".. Nested item")

		assert.Equal(t, BlockListOrdered, classification.Type)
		assert.Equal(t, 2, classification.List.Ordinal)
		assert.Equal(t, "Nested item", classification.List.Text)
	})

	t.Run("labeled list", func(t *testing.T) {
		classification := lc.ClassifyLine("term :: definition")

		assert.Equal(t, BlockListLabeled, classification.Type)
		assert.Equal(t, "::", classification.List.Marker)
		assert.Equal(t, "definition", classification.List.Text)
	})
}

// TestLineClassifierAdmonitions tests admonition detection.
func TestLineClassifierAdmonitions(t *testing.T) {
	lc := NewLineClassifier()

	admonitions := []string{
		"NOTE:", "WARNING:", "TIP:", "CAUTION:", "IMPORTANT:",
	}

	for _, adm := range admonitions {
		t.Run(adm, func(t *testing.T) {
			classification := lc.ClassifyLine(adm)

			assert.Equal(t, BlockAdmonition, classification.Type)
		})
	}
}

// TestBlockTypeString tests the String() method on BlockType.
func TestBlockTypeString(t *testing.T) {
	tests := []struct {
		bt       BlockType
		expected string
	}{
		{BlockParagraph, "Paragraph"},
		{BlockSection, "Section"},
		{BlockListUnordered, "ListUnordered"},
		{BlockVerbatim, "Verbatim"},
		{BlockComment, "Comment"},
		{BlockBlank, "Blank"},
		{BlockUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bt.String())
		})
	}
}

// TestBlockTypeIsListItem tests the IsListItem() method.
func TestBlockTypeIsListItem(t *testing.T) {
	listTypes := []BlockType{
		BlockListUnordered,
		BlockListOrdered,
		BlockListLabeled,
		BlockListChecklist,
		BlockListCallout,
	}

	for _, bt := range listTypes {
		t.Run(bt.String(), func(t *testing.T) {
			assert.True(t, bt.IsListItem(),
				"BlockType %v should be a list item", bt)
		})
	}

	nonListTypes := []BlockType{
		BlockParagraph,
		BlockSection,
		BlockAttribute,
	}

	for _, bt := range nonListTypes {
		t.Run(bt.String(), func(t *testing.T) {
			assert.False(t, bt.IsListItem(),
				"BlockType %v should not be a list item", bt)
		})
	}
}

// TestBlockTypeIsDelimitedBlock tests the IsDelimitedBlock() method.
func TestBlockTypeIsDelimitedBlock(t *testing.T) {
	delimitedTypes := []BlockType{
		BlockLiteral,
		BlockVerbatim,
		BlockExample,
		BlockQuote,
		BlockPassthrough,
		BlockSidebar,
		BlockCommentBlock,
	}

	for _, bt := range delimitedTypes {
		t.Run(bt.String(), func(t *testing.T) {
			assert.True(t, bt.IsDelimitedBlock(),
				"BlockType %v should be a delimited block", bt)
		})
	}

	nonDelimitedTypes := []BlockType{
		BlockParagraph,
		BlockSection,
		BlockListUnordered,
	}

	for _, bt := range nonDelimitedTypes {
		t.Run(bt.String(), func(t *testing.T) {
			assert.False(t, bt.IsDelimitedBlock(),
				"BlockType %v should not be a delimited block", bt)
		})
	}
}
