// Package inline tests for passthrough macros.
package inline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawPassthroughMacro(t *testing.T) {
	test := "raw:[<b>test</b>]"
	parser := NewParser(test)
	nodes := parser.Parse()

	assert.Len(t, nodes, 1, "Should parse one node")
	assert.Equal(t, NodeRawPassThrough, nodes[0].Type)
	assert.Equal(t, "<b>test</b>", nodes[0].Text)
}

func TestPassPassthroughMacro(t *testing.T) {
	test := "pass:[<b>test</b>]"
	parser := NewParser(test)
	nodes := parser.Parse()

	assert.Len(t, nodes, 1, "Should parse one node")
	assert.Equal(t, NodePassThrough, nodes[0].Type)
	assert.Equal(t, "<b>test</b>", nodes[0].Text)
}

func TestTriplePlusPassthrough(t *testing.T) {
	test := "+++<b>test</b>+++"
	parser := NewParser(test)
	nodes := parser.Parse()

	assert.Len(t, nodes, 1, "Should parse one node")
	assert.Equal(t, NodeRawPassThrough, nodes[0].Type)
	assert.Equal(t, "<b>test</b>", nodes[0].Text)
}

func TestSinglePlusPassthrough(t *testing.T) {
	test := "+<b>test</b>+"
	parser := NewParser(test)
	nodes := parser.Parse()

	assert.Len(t, nodes, 1, "Should parse one node")
	assert.Equal(t, NodePassThrough, nodes[0].Type)
	assert.Equal(t, "<b>test</b>", nodes[0].Text)
}
