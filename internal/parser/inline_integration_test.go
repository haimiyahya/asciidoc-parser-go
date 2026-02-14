// Package parser provides integration tests for inline parsing.
package parser

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/stretchr/testify/require"
)

func TestParseInlineBold(t *testing.T) {
	source := "**bold** text"

	p := inline.NewParser(source)
	nodes := p.Parse()

	require.NotEmpty(t, nodes)
	require.Len(t, 1)

	node := nodes[0]
	require.Equal(t, ast.TypeBold, node.Type())
	require.Equal(t, "bold text", node.Text)
}

func TestParseInlineItalic(t *testing.T) {
	source := "__italic__ text"

	p := inline.NewParser(source)
	nodes := p.Parse()

	require.NotEmpty(t, nodes)
	require.Len(t, 1)

	node := nodes[0]
	require.Equal(t, ast.TypeItalic, node.Type())
	require.Equal(t, "italic text", node.Text)
}

func TestParseInlineLink(t *testing.T) {
	source := "link:text[url]"

	p := inline.NewParser(source)
	nodes := p.Parse()

	require.NotEmpty(t, nodes)
	require.Len(t, 1)

	node := nodes[0]
	require.Equal(t, ast.TypeLink, node.Type())
	require.Equal(t, "text", node.Text)
	require.Equal(t, "url", node.URL)
}
