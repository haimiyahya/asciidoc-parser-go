// Package blocks provides tests for block delimiter detection.
package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDelimitedBlock(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantType DelimiterType
		wantChar byte
	}{
		{
			name:     "verbatim block (4 dashes)",
			line:     "----",
			wantType: DelimiterVerbatim,
			wantChar: '-',
		},
		{
			name:     "example block (4 equals)",
			line:     "====",
			wantType: DelimiterExample,
			wantChar: '=',
		},
		{
			name:     "quote block (4 underscores)",
			line:     "____",
			wantType: DelimiterQuote,
			wantChar: '_',
		},
		{
			name:     "passthrough block (4 pluses)",
			line:     "++++",
			wantType: DelimiterPassthrough,
			wantChar: '+',
		},
		{
			name:     "sidebar block (4 asterisks)",
			line:     "****",
			wantType: DelimiterSidebar,
			wantChar: '*',
		},
		{
			name:     "comment block (4 slashes)",
			line:     "////",
			wantType: DelimiterComment,
			wantChar: '/',
		},
		{
			name:     "literal block (4 dots)",
			line:     "....",
			wantType: DelimiterLiteral,
			wantChar: '.',
		},
		{
			name:     "longer verbatim block",
			line:     "--------",
			wantType: DelimiterVerbatim,
			wantChar: '-',
		},
		{
			name:     "table separator",
			line:     "|===",
			wantType: DelimiterLiteral,
			wantChar: '|',
		},
		{
			name:     "3 char not delimiter",
			line:     "---",
			wantType: DelimiterUnknown,
			wantChar: 0,
		},
		{
			name:     "empty line",
			line:     "",
			wantType: DelimiterUnknown,
			wantChar: 0,
		},
		{
			name:     "mixed chars",
			line:     "-_-=",
			wantType: DelimiterUnknown,
			wantChar: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDelimitedBlock(tt.line)
			assert.Equal(t, tt.wantType, got.Type)
			if tt.wantChar != 0 {
				assert.Equal(t, tt.wantChar, got.Char)
			}
		})
	}
}
