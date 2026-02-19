// Package inline tests for pass macro.
package inline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassMacroPositions(t *testing.T) {
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
	// Positions are relative to the content, not including macro delimiters
	assert.Equal(t, 12, nodes[1].StartPos, "StartPos should be at content start")
	assert.Equal(t, 30, nodes[1].Position, "Position should be at content end")
	assert.Equal(t, 30, nodes[2].StartPos, "Next node should start after content")
}
