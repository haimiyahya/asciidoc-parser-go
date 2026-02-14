// Package reader provides a line-oriented reader for AsciiDoc source input.
//
// This file contains the LineClassifier which identifies block types
// based on line patterns, following the AsciiDoc Language Specification.
//
// Reference: https://docs.asciidoctor.org/asciidoc/latest/
package reader

import (
	"strings"
	"unicode"
)

// BlockType represents the classification of a line or block in AsciiDoc.
//
// This corresponds to the semantic categories in the AsciiDoc Language
// Specification. Each type represents a distinct structural element that
// the parser handles differently.
type BlockType int

const (
	// BlockUnknown represents a line that couldn't be classified.
	BlockUnknown BlockType = iota

	// BlockBlank is an empty or whitespace-only line.
	BlockBlank

	// BlockSection is a section header (= Title).
	BlockSection

	// BlockAttribute is an attribute entry (:name: value or :name!:).
	BlockAttribute

	// BlockComment is a single-line comment (// comment).
	BlockComment

	// BlockCommentBlock is a delimited comment block (//// ... ////).
	BlockCommentBlock

	// BlockListUnordered is an unordered list item (-, *, o).
	BlockListUnordered

	// BlockListOrdered is an ordered list item (., .., ..., ;, ::, :::).
	BlockListOrdered

	// BlockListLabeled is a labeled/definition list item (term :: def).
	BlockListLabeled

	// BlockListChecklist is a checklist item (- [x]).
	BlockListChecklist

	// BlockListCallout is a callout list item (NOTE>, WARNING>, etc.).
	BlockListCallout

	// BlockLiteral is a literal block (indented or .... delimited).
	BlockLiteral

	// BlockVerbatim is a verbatim block (---- delimited).
	BlockVerbatim

	// BlockExample is an example block (==== delimited).
	BlockExample

	// BlockQuote is a quote block (____ delimited).
	BlockQuote

	// BlockPassthrough is a passthrough block (++++ delimited).
	BlockPassthrough

	// BlockSidebar is a sidebar block (**** delimited).
	BlockSidebar

	// BlockSource is a source block (++++ specific tags).
	BlockSource

	// BlockTable is a table (|=== delimited or with | columns).
	BlockTable

	// BlockHorizontalRule is a horizontal rule (--- or '''').
	BlockHorizontalRule

	// BlockPageBreak is a page break (<<<).
	BlockPageBreak

	// BlockThematicBreak is a thematic break (single line of --- or ''').
	BlockThematicBreak

	// BlockMacro is a block macro (image::, include::, etc.).
	BlockMacro

	// BlockStyle is a block style marker ([style]).
	BlockStyle

	// BlockAnchor is a block anchor ([#anchor] or [[anchor]]).
	BlockAnchor

	// BlockAdmonition is an admonition paragraph/block (NOTE:, WARNING:, etc.).
	BlockAdmonition

	// BlockParagraph is a regular paragraph (default).
	BlockParagraph
)

// String returns the string representation of the BlockType.
func (bt BlockType) String() string {
	names := map[BlockType]string{
		BlockUnknown:        "Unknown",
		BlockBlank:          "Blank",
		BlockSection:        "Section",
		BlockAttribute:      "Attribute",
		BlockComment:        "Comment",
		BlockCommentBlock:   "CommentBlock",
		BlockListUnordered:  "ListUnordered",
		BlockListOrdered:    "ListOrdered",
		BlockListLabeled:    "ListLabeled",
		BlockListChecklist:  "ListChecklist",
		BlockListCallout:    "ListCallout",
		BlockLiteral:        "Literal",
		BlockVerbatim:       "Verbatim",
		BlockExample:        "Example",
		BlockQuote:          "Quote",
		BlockPassthrough:    "Passthrough",
		BlockSidebar:        "Sidebar",
		BlockSource:         "Source",
		BlockTable:          "Table",
		BlockHorizontalRule: "HorizontalRule",
		BlockPageBreak:      "PageBreak",
		BlockThematicBreak:  "ThematicBreak",
		BlockMacro:          "Macro",
		BlockStyle:          "Style",
		BlockAnchor:         "Anchor",
		BlockAdmonition:     "Admonition",
		BlockParagraph:      "Paragraph",
	}
	if name, ok := names[bt]; ok {
		return name
	}
	return "Unknown"
}

// BlockStyleInfo represents the styling/modifier information for a block.
type BlockStyleInfo struct {
	// Name is the style name (e.g., "example", "quote", "source")
	Name string

	// Kind is the kind of block style (open, delimited, etc.)
	Kind BlockStyleKind

	// Attributes are any inline attributes on the style
	Attributes map[string]string
}

// BlockStyleKind represents different ways styles are applied.
type BlockStyleKind int

const (
	// StyleDelimited is a delimited block (====, ----, etc.)
	StyleDelimited BlockStyleKind = iota

	// StyleOpen is an open block (indention-based)
	StyleOpen

	// StyleInline is an inline style ([style])
	StyleInline
)

// SectionInfo contains information about a section header.
type SectionInfo struct {
	// Level is the section level (0 for document title, 1-6 for sections)
	Level int

	// Title is the section title text
	Title string

	// ID is the optional section ID
	ID string

	// Attributes are any inline attributes
	Attributes map[string]string
}

// ListInfo contains information about a list item.
type ListInfo struct {
	// Type is the list type
	Type BlockType

	// Marker is the list marker character (., -, *, ::, etc.)
	Marker string

	// Level is the nesting level (1-based)
	Level int

	// Ordinal is the ordinal number for ordered lists (1, 2, etc.)
	Ordinal int

	// Text is the item text (excluding the marker)
	Text string

	// Term is the term text for labeled lists (text before ::)
	Term string

	// Continuation is true if this line ends with +
	Continuation bool

	// ChecklistState is the checklist state (true, false, or empty for indeterminate)
	ChecklistState ChecklistState
}

// ChecklistState represents the state of a checklist item.
type ChecklistState int

const (
	// ChecklistUnchecked is a checklist item with [ ]
	ChecklistUnchecked ChecklistState = iota

	// ChecklistChecked is a checklist item with [x] or [*]
	ChecklistChecked

	// ChecklistIndeterminate is a checklist item with [-]
	ChecklistIndeterminate
)

// AttributeInfo contains information about an attribute entry.
type AttributeInfo struct {
	// Name is the attribute name
	Name string

	// Value is the attribute value (empty for unset attributes)
	Value string

	// IsSet is false if the attribute is being unset (:name!)
	IsSet bool

	// IsEntry is true for :entry: style attributes
	IsEntry bool
}

// Classification contains the complete classification of a line.
type Classification struct {
	// Type is the primary block type
	Type BlockType

	// Section is populated for section headers
	Section *SectionInfo

	// List is populated for list items
	List *ListInfo

	// Attribute is populated for attribute entries
	Attribute *AttributeInfo

	// Style contains any block style information
	Style *BlockStyleInfo

	// Level is the nesting level (for lists, literal blocks, etc.)
	Level int

	// Indent is the leading whitespace amount (for indented blocks)
	Indent int

	// Continuation is true if the line ends with a continuation character (+)
	Continuation bool

	// Original is the original line before any processing
	Original string

	// Trimmed is the line with leading/trailing whitespace trimmed
	Trimmed string

	// ChecklistState is populated for checklist items
	ChecklistState ChecklistState
}

// LineClassifier classifies AsciiDoc lines into block types.
//
// The classifier implements "human visual perception emulation" - it makes
// decisions the way a human scanning the text would, using recognizable
// patterns and minimal lookahead.
//
// The design mirrors Asciidoctor's block parsing strategy:
// 1. Check for blank lines first (they're visually obvious)
// 2. Check for delimited block delimiters (visually distinct patterns)
// 3. Check for list markers (recognizable prefixes)
// 4. Check for sections (visually prominent leading characters)
// 5. Check for attributes and macros (specific syntax)
// 6. Default to paragraph
type LineClassifier struct {
	// compatMode enables Asciidoctor compatibility mode
	compatMode bool

	// strictMode enables stricter spec compliance
	strictMode bool
}

// ClassifierOption configures a LineClassifier.
type ClassifierOption func(*LineClassifier)

// WithCompatMode enables Asciidoctor compatibility mode.
func WithCompatMode() ClassifierOption {
	return func(lc *LineClassifier) {
		lc.compatMode = true
	}
}

// WithStrictMode enables strict spec compliance mode.
func WithStrictMode() ClassifierOption {
	return func(lc *LineClassifier) {
		lc.strictMode = true
	}
}

// NewLineClassifier creates a new LineClassifier.
func NewLineClassifier(opts ...ClassifierOption) *LineClassifier {
	return &LineClassifier{}
}

// Classify classifies a single line into a BlockType.
// For detailed classification information, use ClassifyLine.
func (lc *LineClassifier) Classify(line string) BlockType {
	classification := lc.ClassifyLine(line)
	return classification.Type
}

// ClassifyLine provides detailed classification of a line.
func (lc *LineClassifier) ClassifyLine(line string) *Classification {
	result := &Classification{
		Original: line,
		Trimmed:  strings.TrimSpace(line),
	}

	trimmed := result.Trimmed

	// Empty line check
	if trimmed == "" {
		result.Type = BlockBlank
		return result
	}

	// Store indentation level
	indent := countLeadingWhitespace(line)
	result.Indent = indent

	// Check for continuation at end (+ at end of line)
	if strings.HasSuffix(trimmed, "+") {
		// Must be a space before + for continuation (in most contexts)
		// Asciidoctor: "line ending with + continues to next line"
		if len(trimmed) > 1 {
			lastChar := trimmed[len(trimmed)-1]
			secondLast := trimmed[len(trimmed)-2]
			if lastChar == '+' && (secondLast == ' ' || secondLast == '\t') {
				result.Continuation = true
				trimmed = strings.TrimRight(trimmed, " \t+")
			}
		}
	}

	// First, check for block delimiters (they're visually distinct)
	if blockType := lc.checkDelimitedBlock(trimmed); blockType != BlockUnknown {
		result.Type = blockType
		return result
	}

	// Check for horizontal rules
	if lc.isHorizontalRule(trimmed) {
		result.Type = BlockHorizontalRule
		return result
	}

	// Check for page break
	if trimmed == "<<<" {
		result.Type = BlockPageBreak
		return result
	}

	// Check for section headers
	if section := lc.checkSection(trimmed); section != nil {
		result.Type = BlockSection
		result.Section = section
		return result
	}

	// Check for attribute entries
	if attr := lc.checkAttribute(trimmed); attr != nil {
		result.Type = BlockAttribute
		result.Attribute = attr
		return result
	}

	// Check for block macros
	if lc.isBlockMacro(trimmed) {
		result.Type = BlockMacro
		return result
	}

	// Check for list items
	if listInfo := lc.checkListItem(trimmed, indent); listInfo != nil {
		result.Type = listInfo.Type
		result.List = listInfo
		result.Level = listInfo.Level
		return result
	}

	// Check for single-line comments
	if lc.isSingleLineComment(trimmed) {
		result.Type = BlockComment
		return result
	}

	// Check for block anchors
	if lc.isBlockAnchor(trimmed) {
		result.Type = BlockAnchor
		return result
	}

	// Check for admonitions
	if admonition := lc.checkAdmonition(trimmed); admonition != "" {
		result.Type = BlockAdmonition
		return result
	}

	// Check for styled blocks (open block with style)
	if lc.isStyledBlock(trimmed) {
		result.Type = BlockStyle
		return result
	}

	// Default: regular paragraph
	result.Type = BlockParagraph
	return result
}

// checkDelimitedBlock checks if the line is a delimited block delimiter.
func (lc *LineClassifier) checkDelimitedBlock(line string) BlockType {
	trimmed := strings.TrimSpace(line)

	// All delimiters are at least 4 characters
	if len(trimmed) < 4 {
		return BlockUnknown
	}

	// Special case: table block delimiter is |=== (pipe + equals)
	// This is the only delimited block that uses two different characters
	if strings.HasPrefix(trimmed, "|") {
		// Check if rest of line is all '='
		rest := trimmed[1:]
		if len(rest) >= 3 && isAllEquals(rest) {
			return BlockTable
		}
	}

	// Check if the entire line is the same character
	// Delimiters: ----, ====, ____, +++, ////, ****, ....
	first := trimmed[0]
	if !isDelimiterChar(first) {
		return BlockUnknown
	}

	allSame := true
	for _, c := range trimmed {
		if c != rune(first) {
			allSame = false
			break
		}
	}

	if !allSame {
		return BlockUnknown
	}

	// Map delimiter character to block type
	switch first {
	case '-':
		return BlockVerbatim
	case '=':
		return BlockExample
	case '_':
		return BlockQuote
	case '+':
		return BlockPassthrough
	case '*':
		return BlockSidebar
	case '/':
		return BlockCommentBlock
	case '.':
		return BlockLiteral
	case '|':
		return BlockTable
	default:
		return BlockUnknown
	}
}

// isDelimiterChar returns true if the character is used for block delimiters.
func isDelimiterChar(c byte) bool {
	return c == '-' || c == '=' || c == '_' || c == '+' ||
		c == '*' || c == '/' || c == '.' || c == '|'
}

// isAllEquals returns true if the string contains only '=' characters.
func isAllEquals(s string) bool {
	for _, c := range s {
		if c != '=' {
			return false
		}
	}
	return true
}

// isHorizontalRule checks for horizontal rule syntax.
// Horizontal rules are: ----, ””, or ===== (sometimes)
func (lc *LineClassifier) isHorizontalRule(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Must be at least 3 characters
	if len(trimmed) < 3 {
		return false
	}

	// Common pattern: 3 or more of the same character
	first := trimmed[0]
	if first == '-' || first == '\'' || first == '=' {
		allSame := true
		for _, c := range trimmed {
			if c != rune(first) {
				allSame = false
				break
			}
		}
		return allSame
	}

	return false
}

// checkSection checks for section header syntax (= Title).
func (lc *LineClassifier) checkSection(line string) *SectionInfo {
	trimmed := strings.TrimSpace(line)

	// Section headers start with = and have at least one space after the =s
	// Pattern: ^=+\s+.+
	if len(trimmed) < 2 {
		return nil
	}

	// Count leading = characters
	equalsCount := 0
	for _, c := range trimmed {
		if c == '=' {
			equalsCount++
		} else {
			break
		}
	}

	if equalsCount == 0 {
		return nil
	}

	// Check if there's space after the =s
	if equalsCount >= len(trimmed) || trimmed[equalsCount] != ' ' {
		// No space after =s, not a valid section header
		// Exception: single "=" followed immediately by text is NOT a section
		return nil
	}

	// Extract the title (after =s and space, with inline attributes)
	title := trimmed[equalsCount+1:]

	// Level 0 is document title (= Title)
	// Level 1-6 are sections (== to =======)
	level := equalsCount - 1
	if level > 6 {
		level = 6
	}

	// Parse out inline attributes like [#id] or [role,opts="val"]
	cleanTitle, attrs := lc.parseInlineAttributes(title)

	return &SectionInfo{
		Level:      level,
		Title:      cleanTitle,
		ID:         attrs["id"],
		Attributes: attrs,
	}
}

// parseInlineAttributes extracts attributes from inline syntax.
// Handles: [#id], [role], [opts="val"], etc.
func (lc *LineClassifier) parseInlineAttributes(text string) (string, map[string]string) {
	trimmed := strings.TrimSpace(text)

	// Check for [...] at end or at position 0
	// Pattern: Title [#id] or [role]Title or Title[opts="val"]
	if !strings.Contains(trimmed, "[") {
		return trimmed, nil
	}

	// Simple parsing: look for [...] at the end
	openBracket := strings.LastIndex(trimmed, "[")
	closeBracket := strings.LastIndex(trimmed, "]")

	if openBracket >= 0 && closeBracket > openBracket {
		// Check if ] is at the end or only followed by whitespace
		afterClose := strings.TrimSpace(trimmed[closeBracket+1:])
		if afterClose == "" {
			attrs := make(map[string]string)

			// Extract the content between [ and ]
			attrContent := trimmed[openBracket+1 : closeBracket]

			// Parse id: #id
			if strings.HasPrefix(attrContent, "#") {
				attrs["id"] = strings.TrimPrefix(attrContent, "#")
			} else {
				// Parse as style/options
				attrs["style"] = attrContent
			}

			// Return text before the attributes
			title := strings.TrimSpace(trimmed[:openBracket])
			return title, attrs
		}
	}

	return trimmed, nil
}

// checkAttribute checks for attribute entry syntax.
// Patterns: :name:, :name: value, :name!: (unset), :name!: value (override)
func (lc *LineClassifier) checkAttribute(line string) *AttributeInfo {
	trimmed := strings.TrimSpace(line)

	// Must start with :
	if len(trimmed) < 2 || trimmed[0] != ':' {
		return nil
	}

	// Find the closing : or space after the name
	colonPos := strings.Index(trimmed[1:], ":")
	if colonPos == -1 {
		return nil
	}
	colonPos += 1 // Adjust for the initial :

	name := trimmed[1:colonPos]

	// Validate attribute name (letters, digits, dash, underscore)
	if !lc.isValidAttributeName(name) {
		return nil
	}

	info := &AttributeInfo{
		Name:  name,
		IsSet: true,
	}

	// Check for ! (unset or set within attribute list)
	if len(trimmed) > colonPos+1 {
		remaining := strings.TrimSpace(trimmed[colonPos+1:])

		// Check for ! after :name
		if remaining == "!" {
			info.IsSet = false
		} else if strings.HasPrefix(remaining, "!") {
			// Override syntax: :name!: value
			info.IsSet = true
			info.Value = strings.TrimSpace(remaining[1:])
		} else {
			// Regular value assignment
			info.Value = remaining
			info.IsEntry = strings.HasPrefix(info.Value, "@")
		}
	}

	return info
}

// isValidAttributeName checks if the attribute name is valid.
func (lc *LineClassifier) isValidAttributeName(name string) bool {
	if name == "" {
		return false
	}

	for _, c := range name {
		isValid := unicode.IsLetter(c) || unicode.IsDigit(c) ||
			c == '-' || c == '_' || c == '.'
		if !isValid {
			return false
		}
	}

	// Can't be all digits
	allDigits := true
	for _, c := range name {
		if !unicode.IsDigit(c) {
			allDigits = false
			break
		}
	}

	return !allDigits
}

// isBlockMacro checks for block macro syntax.
// Pattern: target::path[attrlist] or just target::path
func (lc *LineClassifier) isBlockMacro(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Must contain :: (double colon)
	if !strings.Contains(trimmed, "::") {
		return false
	}

	// Find the position of ::
	doubleColon := strings.Index(trimmed, "::")
	if doubleColon == 0 {
		return false
	}

	// The target must be before ::
	target := trimmed[:doubleColon]

	// Target must be valid (letters, digits, underscore, dash)
	if !lc.isValidMacroTarget(target) {
		return false
	}

	// Optional: check for attribute list [...]
	afterMacro := trimmed[doubleColon+2:]
	if strings.HasPrefix(afterMacro, "[") {
		// Has attribute list, validate closing ]
		closeBracket := strings.Index(afterMacro, "]")
		if closeBracket == -1 {
			// Invalid macro
			return false
		}
	}

	return true
}

// isValidMacroTarget checks if the macro target is valid.
func (lc *LineClassifier) isValidMacroTarget(target string) bool {
	if target == "" {
		return false
	}

	for _, c := range target {
		isValid := unicode.IsLetter(c) || unicode.IsDigit(c) ||
			c == '_' || c == '-' || c == ':' || c == '/' || c == '.'
		if !isValid {
			return false
		}
	}

	return true
}

// checkListItem checks if the line is a list item.
func (lc *LineClassifier) checkListItem(line string, indent int) *ListInfo {
	trimmed := strings.TrimSpace(line)

	if len(trimmed) < 1 {
		return nil
	}

	// Check for specific list markers
	info := &ListInfo{
		Level: 1,
	}

	// Determine level from indentation (1 space = 1 level, typically)
	// Asciidoctor: 1 space per nesting level for ordered, 1+ for unordered
	if indent > 0 {
		info.Level = indent
	}

	// Unordered lists: -, *, o (with space after)
	if lc.isUnorderedListItem(trimmed, info) {
		return info
	}

	// Ordered lists: ., .., ..., ;, ::, ::: (dot-style)
	if lc.isOrderedListItem(trimmed, info) {
		return info
	}

	// Labeled lists: term :: or term ;; or term :::: (with space after)
	if lc.isLabeledListItem(trimmed, info) {
		return info
	}

	// Checklist: - [x] or - [ ]
	if lc.isChecklistItem(trimmed, info) {
		return info
	}

	// Callout: NOTE>, WARNING>, etc.
	if lc.isCalloutItem(trimmed, info) {
		return info
	}

	return nil
}

// isUnorderedListItem checks for unordered list markers.
func (lc *LineClassifier) isUnorderedListItem(line string, info *ListInfo) bool {
	// Patterns: " - ", " * ", " o " (with space after)
	if len(line) < 3 {
		return false
	}

	markers := map[rune]bool{'-': true, '*': true, 'o': true}

	first := rune(line[0])
	second := line[1]

	// Must have space after marker
	if second != ' ' && second != '\t' {
		return false
	}

	if markers[first] {
		info.Type = BlockListUnordered
		info.Marker = string(first)
		info.Text = strings.TrimSpace(line[2:])
		return true
	}

	return false
}

// isOrderedListItem checks for ordered list markers.
func (lc *LineClassifier) isOrderedListItem(line string, info *ListInfo) bool {
	// Patterns: ". ", ".. ", "... ", "; ", ":: ", "::: "
	// The count of dots (or semicolons) determines the nesting level

	if len(line) < 2 {
		return false
	}

	// Check for dot-style or semicolon-style
	first := line[0]
	if first != '.' && first != ';' {
		return false
	}

	// Count the markers
	markerCount := 0
	for _, c := range line {
		if c == rune(first) {
			markerCount++
		} else {
			break
		}
	}

	// Must have space after markers (or be end of line for some cases)
	if markerCount >= len(line) || line[markerCount] != ' ' && line[markerCount] != '\t' {
		// Edge case: single "." at end of line is valid for simple ordered
		if markerCount != len(line) {
			return false
		}
	}

	if first == '.' {
		info.Type = BlockListOrdered
	} else {
		info.Type = BlockListOrdered // Asciidoctor treats ; as ordered variant
	}

	info.Marker = strings.Repeat(string(first), markerCount)
	info.Ordinal = markerCount
	info.Level = markerCount
	info.Text = strings.TrimSpace(line[markerCount:])
	return true
}

// isLabeledListItem checks for labeled/definition list markers.
func (lc *LineClassifier) isLabeledListItem(line string, info *ListInfo) bool {
	// Patterns: " :: ", " ;; ", " ::: " (space before ::)
	// Or: "term:: " (no space before :: if it's a multi-line term)

	// Check for :: (without space before)
	doubleColon := strings.Index(line, "::")
	if doubleColon == -1 {
		return false
	}

	// Check for space before ::
	if doubleColon > 0 {
		before := line[doubleColon-1]
		if before != ' ' && before != '\t' {
			// Check for "term::" pattern (single :: with no space before)
			// This is a labeled list term
			term := strings.TrimSpace(line[:doubleColon])
			definition := ""
			if doubleColon+2 < len(line) {
				definition = strings.TrimSpace(line[doubleColon+2:])
			}
			info.Type = BlockListLabeled
			info.Marker = "::"
			info.Term = term
			info.Text = definition
			return true
		}
	}

	// Space before :: means it's a labeled list
	if doubleColon > 0 {
		before := line[doubleColon-1]
		if before == ' ' || before == '\t' {
			// Extract term (before ::) and definition (after ::)
			term := strings.TrimSpace(line[:doubleColon-1])
			definition := ""
			if doubleColon+1 < len(line) {
				definition = strings.TrimSpace(line[doubleColon+2:])
			}
			info.Type = BlockListLabeled
			info.Marker = "::"
			info.Term = term
			info.Text = definition
			return true
		}
	}

	return false
}

// isChecklistItem checks for checklist syntax.
func (lc *LineClassifier) isChecklistItem(line string, info *ListInfo) bool {
	// Patterns: "- [x] ", "- [*] ", "- [ ] ", "- [-] "

	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "-") {
		return false
	}

	remaining := strings.TrimSpace(trimmed[1:])
	if len(remaining) < 3 || remaining[0] != '[' {
		return false
	}

	closeBracket := strings.Index(remaining, "]")
	if closeBracket == -1 {
		return false
	}

	content := remaining[1:closeBracket]

	// Determine checklist state
	switch content {
	case "x", "*":
		info.ChecklistState = ChecklistChecked
	case " ", "-":
		info.ChecklistState = ChecklistUnchecked
	default:
		info.ChecklistState = ChecklistIndeterminate
	}

	// Get text after ]
	text := ""
	if closeBracket+1 < len(remaining) {
		text = strings.TrimSpace(remaining[closeBracket+1:])
	}

	info.Type = BlockListChecklist
	info.Marker = "-"
	info.Text = text
	return true
}

// isCalloutItem checks for callout list syntax.
func (lc *LineClassifier) isCalloutItem(line string, info *ListInfo) bool {
	// Patterns: "NOTE>:", "WARNING>:", "TIP>:", "CAUTION>:", "IMPORTANT>:", etc.
	trimmed := strings.TrimSpace(line)

	// Callouts use >, not :
	callouts := []string{
		"NOTE>", "WARNING>", "TIP>", "CAUTION>", "IMPORTANT>",
	}

	for _, callout := range callouts {
		if strings.HasPrefix(trimmed, callout) {
			info.Type = BlockListCallout
			info.Marker = callout
			info.Text = strings.TrimSpace(trimmed[len(callout):])
			return true
		}
	}

	return false
}

// isSingleLineComment checks for single-line comment syntax.
func (lc *LineClassifier) isSingleLineComment(line string) bool {
	// Pattern: // comment
	// But NOT: //// (which is a block delimiter)
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "//") {
		return false
	}

	// Check if it's all slashes (block delimiter)
	if isAllSlashes(trimmed) {
		return false
	}

	// Must have space after // or be more than just //
	if len(trimmed) > 2 {
		next := trimmed[2]
		if next == ' ' || next == '\t' {
			return true
		}
	}

	// Single // is a comment
	return trimmed == "//"
}

// isBlockAnchor checks for block anchor syntax.
func (lc *LineClassifier) isBlockAnchor(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Patterns: [[id]] or [#id]
	// Or: [id] when followed by anchor or alone

	// Check for [[id]]
	if strings.HasPrefix(trimmed, "[[") && strings.HasSuffix(trimmed, "]]") {
		return true
	}

	// Check for [#id]
	if strings.HasPrefix(trimmed, "[#") && strings.HasSuffix(trimmed, "]") {
		return true
	}

	return false
}

// checkAdmonition checks for admonition syntax.
func (lc *LineClassifier) checkAdmonition(line string) string {
	trimmed := strings.TrimSpace(line)

	// Patterns: NOTE:, WARNING:, TIP:, CAUTION:, IMPORTANT:
	// Must be at start and end with :

	admonitions := []string{
		"NOTE:", "WARNING:", "TIP:", "CAUTION:", "IMPORTANT:",
	}

	for _, adm := range admonitions {
		if trimmed == adm {
			// DEBUG: This should match for "NOTE:" == "NOTE:"
			return adm
		}
		// Check for admonition with space after :
		if len(trimmed) >= len(adm)-1 && strings.HasPrefix(trimmed, adm[:len(adm)-1]) {
			if len(trimmed) >= len(adm) && trimmed[len(adm)-1] == ' ' {
				// Also valid: "NOTE: text"
				return adm
			}
		}
	}

	return ""
}

// isStyledBlock checks for styled block syntax.
// Patterns: [style], [style,opts="val"]
func (lc *LineClassifier) isStyledBlock(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Must start with [ and end with ]
	if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
		return false
	}

	// Extract content between brackets
	content := trimmed[1 : len(trimmed)-1]

	// Can't be empty
	if content == "" {
		return false
	}

	// Check if it's an anchor [#id]
	if strings.HasPrefix(content, "#") {
		return false // That's an anchor, not a styled block
	}

	return true
}

// countLeadingWhitespace counts leading whitespace characters.
func countLeadingWhitespace(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' || c == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

// IsListItem returns true if the block type is any list item.
func (bt BlockType) IsListItem() bool {
	switch bt {
	case BlockListUnordered, BlockListOrdered, BlockListLabeled,
		BlockListChecklist, BlockListCallout:
		return true
	default:
		return false
	}
}

// IsDelimitedBlock returns true if the block type is a delimited block.
func (bt BlockType) IsDelimitedBlock() bool {
	switch bt {
	case BlockLiteral, BlockVerbatim, BlockExample, BlockQuote,
		BlockPassthrough, BlockSidebar, BlockCommentBlock, BlockTable:
		return true
	default:
		return false
	}
}

// IsAdmonition returns true if the block type is an admonition.
func (bt BlockType) IsAdmonition() bool {
	return bt == BlockAdmonition || bt == BlockListCallout
}

// IsCallout returns true if the block type is a callout.
func (bt BlockType) IsCallout() bool {
	if bt == BlockListCallout {
		return true
	}
	// Check if it's an admonition callout
	// (handled in caller)
	return false
}
