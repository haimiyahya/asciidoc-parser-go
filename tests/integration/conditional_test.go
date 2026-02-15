// Package integration provides end-to-end tests for conditional directive processing.
package integration

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConditionalIfdef tests ifdef directive.
func TestConditionalIfdef(t *testing.T) {
	// Test with attribute defined
	src := `:release-phase: beta

ifdef::release-phase[]
This content should appear because the attribute is defined.
endif::release-phase[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have the content
	assert.NotEmpty(t, doc.Blocks, "document should have blocks from ifdef")
}

// TestConditionalIfdefNotDefined tests ifdef with undefined attribute.
func TestConditionalIfdefNotDefined(t *testing.T) {
	// Test with attribute NOT defined
	src := `ifdef::undefined-attr[]
This content should NOT appear because the attribute is not defined.
endif::undefined-attr[]

This content should appear.
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should only have one block (the paragraph after the conditional)
	// The ifdef content should be excluded
	assert.Equal(t, 1, len(doc.Blocks), "document should only have the final paragraph")
}

// TestConditionalIfndef tests ifndef directive.
func TestConditionalIfndef(t *testing.T) {
	// Test with attribute NOT defined
	src := `ifndef::undefined-attr[]
This content should appear because the attribute is not defined.
endif::undefined-attr[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have the content
	assert.NotEmpty(t, doc.Blocks, "document should have blocks from ifndef")
}

// TestConditionalIfndefWithDefined tests ifndef with defined attribute.
func TestConditionalIfndefWithDefined(t *testing.T) {
	// Test with attribute defined
	src := `:myattr: value

ifndef::myattr[]
This content should NOT appear because the attribute is defined.
endif::myattr[]

This content should appear.
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should only have one block (the paragraph after the conditional)
	assert.Equal(t, 1, len(doc.Blocks), "document should only have the final paragraph")
}

// TestConditionalIfeval tests ifeval directive.
func TestConditionalIfeval(t *testing.T) {
	// Test with equality - no blank lines between conditionals
	src := `:env: production
ifeval::["{env}" == "production"]
This content should appear because env equals production.
endif::[]
ifeval::["{env}" == "development"]
This content should NOT appear because env is not development.
endif::[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have one block from the first ifeval
	assert.Equal(t, 1, len(doc.Blocks), "document should have one block from ifeval")
}

// TestConditionalIfevalNotEqual tests ifeval with != operator.
func TestConditionalIfevalNotEqual(t *testing.T) {
	src := `:env: production

ifeval::["{env}" != "development"]
This content should appear because env is not development.
endif::eval[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	assert.NotEmpty(t, doc.Blocks, "document should have blocks from ifeval !=")
}

// TestConditionalNested tests nested conditionals.
func TestConditionalNested(t *testing.T) {
	src := `:outer: yes
:inner: yes

ifdef::outer[]
Outer content.
ifdef::inner[]
Inner content.
endif::inner[]
endif::outer[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have content from both conditionals
	assert.NotEmpty(t, doc.Blocks, "document should have blocks from nested conditionals")
}

// TestConditionalHTML5Conversion tests that conditional content converts properly to HTML5.
func TestConditionalHTML5Conversion(t *testing.T) {
	src := `:feature-x: enabled

ifdef::feature-x[]
Feature X is enabled!
endif::feature-x[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Convert to HTML5
	var buf strings.Builder
	c := converter.NewHTML5Converter()
	err = c.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify conditional content is present
	assert.Contains(t, output, "Feature X is enabled")
}

// TestConditionalMultipleInDocument tests multiple conditionals in one document.
func TestConditionalMultipleInDocument(t *testing.T) {
	src := `:os: linux
:arch: amd64

ifdef::os[]
Running on os.
endif::os[]

ifdef::arch[]
Architecture is arch.
endif::arch[]

ifndef::windows[]
Not Windows.
endif::windows[]
`

	p, err := parser.NewParserFromString(src)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have 3 blocks
	assert.Equal(t, 3, len(doc.Blocks), "document should have 3 blocks from conditionals")
}
