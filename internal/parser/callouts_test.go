// Package parser provides tests for callout parsing.
package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCalloutsInLiteral(t *testing.T) {
	p := NewCalloutParser()

	lines := []string{
		"line of code // <1>",
		"another line # <2>",
		"third line",
		"clojure line ;; <3>",
	}

	cleanedLines, callouts := p.ParseCalloutsInLiteral(lines)

	// Check that callouts were extracted
	require.Len(t, callouts, 3)

	// Check first callout (C-style)
	assert.Equal(t, 1, callouts[0].Number)
	assert.Equal(t, 0, callouts[0].LineIndex)

	// Check second callout (shell/ruby style)
	assert.Equal(t, 2, callouts[1].Number)
	assert.Equal(t, 1, callouts[1].LineIndex)

	// Check third callout (clojure style)
	assert.Equal(t, 3, callouts[2].Number)
	assert.Equal(t, 3, callouts[2].LineIndex)

	// Check that lines were cleaned (callout markers removed)
	assert.Equal(t, "line of code", cleanedLines[0])
	assert.Equal(t, "another line", cleanedLines[1])
	assert.Equal(t, "third line", cleanedLines[2])
	assert.Equal(t, "clojure line", cleanedLines[3])
}

func TestParseCalloutsInLiteral_XMLComment(t *testing.T) {
	p := NewCalloutParser()

	lines := []string{
		"<section>",
		"  <title>Title</title> <!--1-->",
		"</section>",
	}

	cleanedLines, callouts := p.ParseCalloutsInLiteral(lines)

	// Check that callout was extracted
	require.Len(t, callouts, 1)
	assert.Equal(t, 1, callouts[0].Number)
	assert.Equal(t, 1, callouts[0].LineIndex)

	// Check that line was cleaned
	assert.Equal(t, "  <title>Title</title>", cleanedLines[1])
}

func TestParseCalloutListLine(t *testing.T) {
	p := NewCalloutParser()

	tests := []struct {
		name     string
		line     string
		expected int
		desc     string
	}{
		{
			name:     "simple callout",
			line:     "<1> First description",
			expected: 1,
			desc:     "First description",
		},
		{
			name:     "callout with extra spaces",
			line:     "  <2>  Second description  ",
			expected: 2,
			desc:     "Second description",
		},
		{
			name:     "multi-digit callout",
			line:     "<10> Tenth description",
			expected: 10,
			desc:     "Tenth description",
		},
		{
			name:     "not a callout",
			line:     "regular text",
			expected: -1,
			desc:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, desc := p.ParseCalloutList(tt.line)
			assert.Equal(t, tt.expected, num)
			assert.Equal(t, tt.desc, desc)
		})
	}
}

func TestIsCalloutListLine(t *testing.T) {
	p := NewCalloutParser()

	assert.True(t, p.IsCalloutListLine("<1> Description"))
	assert.True(t, p.IsCalloutListLine("<10> Another description"))
	assert.False(t, p.IsCalloutListLine("Regular text"))
	assert.False(t, p.IsCalloutListLine(""))
	assert.True(t, p.IsCalloutListLine("  <1>")) // A callout indicator is still valid
	assert.True(t, p.IsCalloutListLine("<1>"))  // Even without description
}

func TestParseCalloutList_Document(t *testing.T) {
	source := `----
line one // <1>
line two # <2>
----
<1> First callout description
<2> Second callout description
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have a literal block
	require.Len(t, doc.Blocks, 1)
	literal, ok := doc.Blocks[0].(*ast.NodeLiteral)
	require.True(t, ok, "First block should be a literal block")

	// Should have callouts
	require.Len(t, literal.Callouts, 2)

	// Check first callout
	assert.Equal(t, 1, literal.Callouts[0].Number)
	assert.Equal(t, 0, literal.Callouts[0].LineIndex)
	assert.Equal(t, "First callout description", literal.Callouts[0].Description)

	// Check second callout
	assert.Equal(t, 2, literal.Callouts[1].Number)
	assert.Equal(t, 1, literal.Callouts[1].LineIndex)
	assert.Equal(t, "Second callout description", literal.Callouts[1].Description)

	// Lines should be cleaned
	assert.Equal(t, "line one", literal.Lines[0])
	assert.Equal(t, "line two", literal.Lines[1])
}

func TestParseCalloutList_OnlyCalloutIndicators(t *testing.T) {
	source := `----
line one // <1>
line two // <2>
----
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	require.Len(t, doc.Blocks, 1)
	literal, ok := doc.Blocks[0].(*ast.NodeLiteral)
	require.True(t, ok)

	// Should have callouts but no descriptions
	require.Len(t, literal.Callouts, 2)
	assert.Empty(t, literal.Callouts[0].Description)
	assert.Empty(t, literal.Callouts[1].Description)
}

func TestParseCalloutList_MultiLineDescription(t *testing.T) {
	source := `----
code // <1>
----
<1> This is a longer description that spans
but still continues
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	literal, ok := doc.Blocks[0].(*ast.NodeLiteral)
	require.True(t, ok)

	require.Len(t, literal.Callouts, 1)
	// Note: multi-line descriptions are not currently supported
	// Only the first line is captured
	assert.Equal(t, "This is a longer description that spans", literal.Callouts[0].Description)
}

func TestParseCalloutList_XMLStyle(t *testing.T) {
	source := `----
<section>
  <title>Section</title> <!--1-->
</section>
----
<1> The section title is required.
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	literal, ok := doc.Blocks[0].(*ast.NodeLiteral)
	require.True(t, ok)

	require.Len(t, literal.Callouts, 1)
	assert.Equal(t, 1, literal.Callouts[0].Number)
	assert.Equal(t, "The section title is required.", literal.Callouts[0].Description)
}

func TestCalloutHTML5Conversion(t *testing.T) {
	source := `----
line one // <1>
line two # <2>
----
<1> First description
<2> Second description
`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf strings.Builder
	converter := converter.NewHTML5Converter()
	converter.WithoutHeaderFooter()
	err = converter.Convert(doc, &buf)
	require.NoError(t, err)

	html := buf.String()

	// Check that callout markers are present
	assert.Contains(t, html, `class="conum"`)
	assert.Contains(t, html, `data-value="1"`)
	assert.Contains(t, html, `data-value="2"`)

	// Check that descriptions are present
	assert.Contains(t, html, "First description")
	assert.Contains(t, html, "Second description")

	// Check that colist class is present
	assert.Contains(t, html, `class="colist arabic"`)
}
