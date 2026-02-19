// Package inline tests for raw passthrough macro.
package inline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawMacroEmbeddedInText(t *testing.T) {
	test := "Text before raw:[<b>test</b>] text after"
	parser := NewParser(test)
	nodes := parser.Parse()

	assert.Len(t, nodes, 3, "Should parse three nodes")
	assert.Equal(t, NodeText, nodes[0].Type)
	assert.Equal(t, "Text before ", nodes[0].Text)
	assert.Equal(t, NodeRawPassThrough, nodes[1].Type)
	assert.Equal(t, "<b>test</b>", nodes[1].Text)
	assert.Equal(t, NodeText, nodes[2].Type)
	assert.Equal(t, " text after", nodes[2].Text)
}

func TestPassMacroEmbeddedInText(t *testing.T) {
	test := "Text before pass:[<b>test</b>] text after"
	parser := NewParser(test)
	nodes := parser.Parse()

	assert.Len(t, nodes, 3, "Should parse three nodes")
	assert.Equal(t, NodeText, nodes[0].Type)
	assert.Equal(t, "Text before ", nodes[0].Text)
	assert.Equal(t, NodePassThrough, nodes[1].Type)
	assert.Equal(t, "<b>test</b>", nodes[1].Text)
	assert.Equal(t, NodeText, nodes[2].Type)
	assert.Equal(t, " text after", nodes[2].Text)
}
