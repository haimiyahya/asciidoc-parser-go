// Package parser provides callout parsing for AsciiDoc literal blocks.
package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// Default line comment prefixes supported by AsciiDoc for callouts.
var defaultLineComments = []string{
	"//",   // C-style (C, C++, Java, JavaScript, Go, etc.)
	"#",    // Shell, Ruby, Python, Perl, etc.
	";;",   // Clojure
	"<!--", // XML, HTML (special handling)
	"--",   // AsciiDoc open block delimiter (when line-comment disabled)
}

// CalloutParser handles parsing of callout annotations.
type CalloutParser struct {
	// lineComment is the custom line comment prefix.
	lineComment string
}

// NewCalloutParser creates a new callout parser.
func NewCalloutParser() *CalloutParser {
	return &CalloutParser{}
}

// SetLineComment sets a custom line comment prefix.
func (p *CalloutParser) SetLineComment(comment string) {
	p.lineComment = comment
}

// calloutRegex matches callout indicators like <1>, <12>, etc.
var calloutRegex = regexp.MustCompile(`<(\d+)>`)

// xmlCommentCalloutRegex matches XML comment callouts like <!--1-->
var xmlCommentCalloutRegex = regexp.MustCompile(`<!--(\d+)-->`)

// ParseCalloutsInLiteral parses callout indicators from literal block lines.
// Returns the cleaned lines (with callout indicators removed) and the callouts found.
func (p *CalloutParser) ParseCalloutsInLiteral(lines []string) ([]string, []*ast.CalloutNode) {
	var cleanedLines []string
	var callouts []*ast.CalloutNode

	for lineIdx, line := range lines {
		cleanedLine, lineCallouts := p.extractCallouts(line, lineIdx)
		cleanedLines = append(cleanedLines, cleanedLine)
		callouts = append(callouts, lineCallouts...)
	}

	return cleanedLines, callouts
}

// extractCallouts extracts callout indicators from a single line.
// Returns the cleaned line and any callouts found.
func (p *CalloutParser) extractCallouts(line string, lineIdx int) (string, []*ast.CalloutNode) {
	var callouts []*ast.CalloutNode
	cleanedLine := line

	// Determine which line comment prefix to use
	commentPrefix := p.lineComment
	if commentPrefix == "" {
		// Try to find a matching default comment prefix
		for _, prefix := range defaultLineComments {
			if strings.Contains(line, prefix) {
				commentPrefix = prefix
				break
			}
		}
	}

	// Handle XML comment style callouts specially
	if commentPrefix == "<!--" || (commentPrefix == "" && strings.Contains(line, "<!--")) {
		matches := xmlCommentCalloutRegex.FindAllStringSubmatchIndex(line, -1)
		for _, match := range matches {
			if len(match) >= 4 {
				numStr := line[match[2]:match[3]]
				if num, err := strconv.Atoi(numStr); err == nil {
					callout := &ast.CalloutNode{
						Number:     num,
						LineIndex:  lineIdx,
						Column:     match[0],
						Pos:        ast.Position{Line: lineIdx + 1},
					}
					callouts = append(callouts, callout)
				}
				// Remove the XML comment callout from the cleaned line
				cleanedLine = cleanedLine[:match[0]] + cleanedLine[match[1]:]
			}
		}
		// Remove the <!-- --> part if it's now empty (just the comment delimiters)
		cleanedLine = strings.ReplaceAll(cleanedLine, "<!--->", "")
		cleanedLine = strings.ReplaceAll(cleanedLine, "<!---->", "")
		// Clean up any trailing whitespace
		cleanedLine = strings.TrimRight(cleanedLine, " \t")
		return cleanedLine, callouts
	}

	// Handle standard line comment style callouts
	// Look for patterns like: comment <1>, comment<1>, or just <1>
	matches := calloutRegex.FindAllStringSubmatchIndex(line, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			numStr := line[match[2]:match[3]]
			if num, err := strconv.Atoi(numStr); err == nil {
				callout := &ast.CalloutNode{
					Number:     num,
					LineIndex:  lineIdx,
					Column:     match[0],
					Pos:        ast.Position{Line: lineIdx + 1},
				}
				callouts = append(callouts, callout)
			}

			// Check if there's a line comment prefix before the callout
			beforeCallout := line[:match[0]]
			afterCallout := line[match[1]:]

			if commentPrefix != "" {
				// Remove the comment prefix and space if present before the callout
				trimmedBefore := strings.TrimSpace(beforeCallout)
				if strings.HasSuffix(trimmedBefore, commentPrefix) ||
					strings.HasSuffix(beforeCallout, commentPrefix+" ") ||
					strings.HasSuffix(beforeCallout, commentPrefix) {
					// Find where the comment starts
					commentStart := strings.LastIndex(beforeCallout, commentPrefix)
					if commentStart >= 0 {
						// Rebuild the line without the comment and callout
						prefix := line[:commentStart]
						// Check if there's content before the comment
						prefix = strings.TrimRight(prefix, " \t")
						cleanedLine = prefix + afterCallout
					}
				} else {
					// Just remove the callout, keep the rest
					cleanedLine = beforeCallout + afterCallout
				}
			} else {
				// No comment prefix, just remove the callout
				cleanedLine = beforeCallout + afterCallout
			}
		}
	}

	// Clean up any trailing whitespace
	cleanedLine = strings.TrimRight(cleanedLine, " \t")

	return cleanedLine, callouts
}

// ParseCalloutList parses a callout list line like "<1> Description here".
// Returns the callout number and description, or (-1, "") if not a callout list item.
func (p *CalloutParser) ParseCalloutList(line string) (int, string) {
	line = strings.TrimSpace(line)

	// Match pattern: <number> description
	matches := calloutRegex.FindStringSubmatchIndex(line)
	if len(matches) < 4 {
		return -1, ""
	}

	numStr := line[matches[2]:matches[3]]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return -1, ""
	}

	// Extract description after the callout number
	var descStart int
	if matches[1] < len(line) {
		// Skip the callout and any whitespace after it
		descStart = matches[1]
		desc := strings.TrimSpace(line[descStart:])
		return num, desc
	}

	return num, ""
}

// IsCalloutListLine checks if a line is a callout list item.
func (p *CalloutParser) IsCalloutListLine(line string) bool {
	num, _ := p.ParseCalloutList(line)
	return num > 0
}

// MergeCalloutDescriptions merges callout descriptions into the literal block's callouts.
func (p *CalloutParser) MergeCalloutDescriptions(literal *ast.NodeLiteral, calloutList *ast.CalloutListNode) {
	if calloutList == nil || literal.Callouts == nil {
		return
	}

	// Create a map of callouts by number for easy lookup
	calloutMap := make(map[int]*ast.CalloutNode)
	for _, co := range literal.Callouts {
		calloutMap[co.Number] = co
	}

	// Merge descriptions from the callout list
	for num, descCallout := range calloutList.Items {
		if co, exists := calloutMap[num]; exists {
			co.Description = descCallout.Description
		}
	}
}
