// Package blocks provides tests for block-level AsciiDoc parsing.
package blocks

import (
	"fmt"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockTypeString(t *testing.T) {
	tests := []struct {
		bt       BlockType
		expected string
	}{
		{BlockParagraph, "Paragraph"},
		{BlockList, "List"},
		{BlockDelimited, "Delimited"},
		{BlockSection, "Section"},
		{BlockAttribute, "Attribute"},
		{BlockComment, "Comment"},
		{BlockMacro, "Macro"},
		{BlockSidebar, "Sidebar"},
		{BlockCallout, "Callout"},
		{BlockTable, "Table"},
		{BlockThematicBreak, "ThematicBreak"},
		{BlockPageBreak, "PageBreak"},
		{BlockPassthrough, "Passthrough"},
		{BlockLiteral, "Literal"},
		{BlockVerbatim, "Verbatim"},
		{BlockExample, "Example"},
		{BlockQuote, "Quote"},
		{BlockAdmonition, "Admonition"},
		{BlockStyleType, "Style"},
		{BlockAnchor, "Anchor"},
		{BlockSeparator, "Separator"},
		{BlockTitle, "Title"},
		{BlockPreamble, "Preamble"},
		{BlockUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bt.String())
		})
	}
}

func TestIsDelimitedBlock(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "4 char delimiter",
			line: "----",
			want: true,
		},
		{
			name: "3 char not delimiter",
			line: "--",
			want: false,
		},
		{
			name: "empty line",
			line: "",
			want: false,
		},
		{
			name: "single char",
			line: "-",
			want: false,
		},
		{
			name: "mixed chars",
			line: "-_-",
			want: false,
		},
		{
			name: "verbatim block",
			line: "....",
			want: true,
		},
		{
			name: "example block",
			line: "====",
			want: true,
		},
		{
			name: "quote block",
			line: "____",
			want: true,
		},
		{
			name: "passthrough block",
			line: "++++",
			want: true,
		},
		{
			name: "sidebar block",
			line: "****",
			want: true,
		},
		{
			name: "comment block",
			line: "////",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsDelimitedBlock(tt.line))
		})
	}
}

func TestIdentifyDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		line      string
		wantType DelimiterType
	}{
		{
			name:     "literal block (4 dots)",
			line:     "....",
			wantType: DelimiterLiteral,
		},
		{
			name:     "verbatim block (4 dashes)",
			line:     "----",
			wantType: DelimiterVerbatim,
		},
		{
			name:     "example block (4 equals)",
			line:     "====",
			wantType: DelimiterExample,
		},
		{
			name:     "quote block (4 underscores)",
			line:     "____",
			wantType: DelimiterQuote,
		},
		{
			name:     "passthrough block (4 pluses)",
			line:     "++++",
			wantType: DelimiterPassthrough,
		},
		{
			name:     "sidebar block (4 asterisks)",
			line:     "****",
			wantType: DelimiterSidebar,
		},
		{
			name:     "comment block (4 slashes)",
			line:     "////",
			wantType: DelimiterComment,
		},
		{
			name:     "mixed chars",
			line:     "-_-=",
			wantType: DelimiterUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := IdentifyDelimiter(tt.line)
			assert.Equal(t, tt.wantType, info.Type)
		})
	}
}

func TestIdentifySectionHeader(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantLevel int
		wantTitle string
	}{
		{
			name:      "document title",
			line:      "= Document Title",
			wantLevel: 0,
			wantTitle: "Document Title",
		},
		{
			name:      "level 1 section",
			line:      "== Chapter Title",
			wantLevel: 1,
			wantTitle: "Chapter Title",
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
			name:      "level 4 section",
			line:      "===== Deeper",
			wantLevel: 4,
			wantTitle: "Deeper",
		},
		{
			name:      "level 5 section",
			line:      "====== Deepest",
			wantLevel: 5,
			wantTitle: "Deepest",
		},
		{
			name:      "level 6 section",
			line:      "======= Lowest",
			wantLevel: 6,
			wantTitle: "Lowest",
		},
		{
			name:      "no space after =s",
			line:      "=No",
			wantLevel: 0,
			wantTitle: "",
		},
		{
			name:      "empty line",
			line:      "",
			wantLevel: 0,
			wantTitle: "",
		},
		{
			name:      "not starting with =",
			line:      "- Not Title",
			wantLevel: 0,
			wantTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			section := IdentifySectionHeader(tt.line)
			if tt.wantTitle == "" {
				assert.Nil(t, section)
			} else {
				assert.NotNil(t, section)
				assert.Equal(t, tt.wantLevel, section.Level)
				assert.Equal(t, tt.wantTitle, section.Title)
			}
		})
	}
}

func TestIdentifyAttributeEntry(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantName  string
		wantValue string
		wantIsSet bool
	}{
		{
			name:      "simple attribute",
			line:      ":name: value",
			wantName:  "name",
			wantValue: "value",
			wantIsSet: true,
		},
		{
			name:      "attribute without value",
			line:      ":name:",
			wantName:  "name",
			wantValue: "",
			wantIsSet: true,
		},
		{
			name:      "unset attribute",
			line:      ":name!:",
			wantName:  "name!",
			wantValue: "",
			wantIsSet: false,
		},
		{
			name:      "override value",
			line:      ":name!: value",
			wantName:  "name!",
			wantValue: "value",
			wantIsSet: true,
		},
		{
			name:      "entry-style attribute",
			line:      ":entry: value",
			wantName:  "entry",
			wantValue: "value",
			wantIsSet: true,
		},
		{
			name:      "invalid - starts with digit",
			line:      ":1invalid: value",
			wantName:  "",
			wantValue: "",
			wantIsSet: true,
		},
		{
			name:      "no colon at start",
			line:      "no colon here",
			wantName:  "",
			wantValue: "",
			wantIsSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := IdentifyAttributeEntry(tt.line)
			if tt.wantName == "" {
				assert.Nil(t, attr)
			} else {
				assert.Equal(t, tt.wantName, attr.Name)
				assert.Equal(t, tt.wantValue, attr.Value)
				assert.Equal(t, tt.wantIsSet, attr.IsSet)
			}
		})
	}
}

func TestIdentifyListMarker(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantType  ListType
		wantLevel int
	}{
		{
			name:      "unordered dash",
			line:      "- Item",
			wantType:  ListUnordered,
			wantLevel: 1,
		},
		{
			name:      "unordered asterisk",
			line:      "* Item",
			wantType:  ListUnordered,
			wantLevel: 1,
		},
		{
			name:      "unordered letter o",
			line:      "o Item",
			wantType:  ListUnordered,
			wantLevel: 1,
		},
		{
			name:      "ordered single dot",
			line:      ". First item",
			wantType:  ListOrdered,
			wantLevel: 1,
		},
		{
			name:      "ordered double dot",
			line:      ".. Second item",
			wantType:  ListOrdered,
			wantLevel: 2,
		},
		{
			name:      "ordered triple dot",
			line:      "... Third item",
			wantType:  ListOrdered,
			wantLevel: 3,
		},
		{
			name:      "labeled list single colon",
			line:      "term :: definition",
			wantType:  ListLabeled,
			wantLevel: 1,
		},
		{
			name:      "labeled list double colon",
			line:      "term1 :: definition",
			wantType:  ListLabeled,
			wantLevel: 1,
		},
		{
			name:      "not a list item",
			line:      "Just text",
			wantType:  ListUnknown,
			wantLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marker := IdentifyListMarker(tt.line)
			if tt.wantType == ListUnknown {
				assert.Nil(t, marker)
			} else {
				assert.NotNil(t, marker)
				assert.Equal(t, tt.wantType, marker.Type)
				assert.Equal(t, tt.wantLevel, marker.Level)
			}
		})
	}
}

func TestIsListLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "unordered dash",
			line: "- Item",
			want: true,
		},
		{
			name: "unordered asterisk",
			line: "* Item",
			want: true,
		},
		{
			name: "ordered dot",
			line: ". Item",
			want: true,
		},
		{
			name: "labeled list",
			line: "term :: def",
			want: true,
		},
		{
			name: "regular text",
			line: "Just text",
			want: false,
		},
		{
			name: "empty line",
			line: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsListLine(tt.line))
		})
	}
}

func TestParseSimpleDocument(t *testing.T) {
	source := `= Document Title

This is a paragraph.

== Section 1

This is a second paragraph.`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	parser := NewParser(r)
	doc, err := parser.Parse()

	require.NoError(t, err)
	require.NotNil(t, doc)

	for i, b := range doc.Blocks {
		fmt.Printf("  Block %d: Type=%v, Lines=%q\\n", i, b.Type, b.Lines)
	}

	assert.Greater(t, len(doc.Blocks), 0)
}

func TestParseVerbatimBlock(t *testing.T) {
	source := `----
Some verbatim content
----`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	p := NewParser(r)
	doc, err := p.Parse()
	require.NoError(t, err)

	require.NotNil(t, doc)
	assert.Len(t, doc.Blocks, 1)
	assert.Equal(t, BlockVerbatim, doc.Blocks[0].Type)
}

func TestParseLiteralBlock(t *testing.T) {
	source := `....
Some literal content
....`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	p := NewParser(r)
	doc, err := p.Parse()
	require.NoError(t, err)

	require.NotNil(t, doc)
	assert.Len(t, doc.Blocks, 1)
	assert.Equal(t, BlockLiteral, doc.Blocks[0].Type)
}

func TestParseExampleBlock(t *testing.T) {
	source := `====
Some example content
====`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	p := NewParser(r)
	doc, err := p.Parse()
	require.NoError(t, err)

	require.NotNil(t, doc)
	assert.Len(t, doc.Blocks, 1)
	assert.Equal(t, BlockExample, doc.Blocks[0].Type)
}

func TestParseQuoteBlock(t *testing.T) {
	source := `____
Some quote content
____`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	p := NewParser(r)
	doc, err := p.Parse()
	require.NoError(t, err)

	require.NotNil(t, doc)
	assert.Len(t, doc.Blocks, 1)
	assert.Equal(t, BlockQuote, doc.Blocks[0].Type)
}

func TestParseSidebarBlock(t *testing.T) {
	source := `****
Some sidebar content
****`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	p := NewParser(r)
	doc, err := p.Parse()
	require.NoError(t, err)

	require.NotNil(t, doc)
	assert.Len(t, doc.Blocks, 1)
	assert.Equal(t, BlockSidebar, doc.Blocks[0].Type)
}

func TestParseListBlock(t *testing.T) {
	source := `- Item 1
- Item 2
- Item 3`

	r, err := reader.NewReader(source)
	require.NoError(t, err)

	p := NewParser(r)
	doc, err := p.Parse()
	require.NoError(t, err)

	require.NotNil(t, doc)
	assert.Len(t, doc.Blocks, 1)
	assert.Equal(t, BlockList, doc.Blocks[0].Type)
	assert.Len(t, doc.Blocks[0].Lines, 3)
}
