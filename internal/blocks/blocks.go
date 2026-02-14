// Package blocks provides block type classification and parsing utilities for AsciiDoc.

package blocks

import (
	"strings"
)

// DelimiterType represents the type of a delimited block delimiter.
type DelimiterType int

const (
	// DelimiterUnknown is an unknown delimiter type.
	DelimiterUnknown DelimiterType = iota
	// DelimiterVerbatim is a verbatim block delimiter (----).
	DelimiterVerbatim
	// DelimiterExample is an example block delimiter (====).
	DelimiterExample
	// DelimiterQuote is a quote block delimiter (____).
	DelimiterQuote
	// DelimiterPassthrough is a passthrough block delimiter (++++).
	DelimiterPassthrough
	// DelimiterSidebar is a sidebar block delimiter (****).
	DelimiterSidebar
	// DelimiterComment is a comment block delimiter (////).
	DelimiterComment
	// DelimiterLiteral is a literal block delimiter (....).
	DelimiterLiteral
)

// DelimiterInfo contains information about a delimited block delimiter.
type DelimiterInfo struct {
	// Char is the delimiter character.
	Char byte
	// Type is the delimiter type.
	Type DelimiterType
}

// IsDelimitedBlock checks if a line is a delimited block delimiter.
// Returns DelimiterInfo with information about the delimiter, or a zero value
// if the line is not a delimited block delimiter.
func IsDelimitedBlock(line string) DelimiterInfo {
	trimmed := strings.TrimSpace(line)

	// Check for table format first: |===
	// Table syntax is different from other delimited blocks
	if strings.HasPrefix(trimmed, "|===") {
		return DelimiterInfo{Char: '|', Type: DelimiterLiteral} // Tables use | delimiter
	}

	// Must be at least 4 characters
	if len(trimmed) < 4 {
		return DelimiterInfo{Type: DelimiterUnknown}
	}

	first := trimmed[0]

	// Check if entire line is same delimiter character
	for i := 1; i < len(trimmed); i++ {
		if trimmed[i] != first {
			return DelimiterInfo{Type: DelimiterUnknown}
		}
	}

	// Map delimiter character to block type
	switch first {
	case '-':
		return DelimiterInfo{Char: '-', Type: DelimiterVerbatim}
	case '=':
		return DelimiterInfo{Char: '=', Type: DelimiterExample}
	case '_':
		return DelimiterInfo{Char: '_', Type: DelimiterQuote}
	case '+':
		return DelimiterInfo{Char: '+', Type: DelimiterPassthrough}
	case '*':
		return DelimiterInfo{Char: '*', Type: DelimiterSidebar}
	case '/':
		return DelimiterInfo{Char: '/', Type: DelimiterComment}
	case '.':
		return DelimiterInfo{Char: '.', Type: DelimiterLiteral}
	default:
		return DelimiterInfo{Type: DelimiterUnknown}
	}
}

// String returns the string representation of the delimiter type.
func (dt DelimiterType) String() string {
	names := map[DelimiterType]string{
		DelimiterUnknown:    "Unknown",
		DelimiterVerbatim:   "Verbatim",
		DelimiterExample:    "Example",
		DelimiterQuote:      "Quote",
		DelimiterPassthrough: "Passthrough",
		DelimiterSidebar:    "Sidebar",
		DelimiterComment:    "Comment",
		DelimiterLiteral:    "Literal",
	}
	if name, ok := names[dt]; ok {
		return name
	}
	return "Unknown"
}
