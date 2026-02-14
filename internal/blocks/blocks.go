// Package blocks provides block-level parsing for AsciiDoc.
package blocks

import (
	"fmt"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
)

type BlockType int

const (
	BlockUnknown BlockType = iota
	BlockParagraph
	BlockDelimited
	BlockList
	BlockSection
	BlockAttribute
	BlockComment
	BlockTable
	BlockLiteral
	BlockVerbatim
	BlockExample
	BlockQuote
	BlockPassthrough
	BlockSidebar
	BlockCallout
	BlockAdmonition
	BlockStyleType
	BlockAnchor
	BlockMacro
	BlockThematicBreak
	BlockPageBreak
	BlockFrontMatter
	BlockSeparator
	BlockTitle
	BlockPreamble
	BlockManife
	BlockRule
	BlockInclude
	BlockIfDef
	BlockEndIf
	BlockMacroDef
	BlockMacroDefBody
	BlockReset
	BlockDeprecated
	BlockReplacement
	BlockTemplate
	BlockMetadata
	BlockIndexTerms
	BlockImage
	BlockVideo
	BlockAudio
	BlockTOC
	BlockTodo
	BlockDescriptionList
	BlockUlist
	BlockOlist
	BlockCalloutList
	BlockQandA
	BlockGlossary
	BlockBibliography
	BlockColophon
	BlockFloatingTitle
	BlockReproducibleUnits
	BlockSyntaxHighlighting
	BlockSourceHighlighting
	BlockPassthroughBlock
	BlockIndex
	BlockFootnote
	BlockFootnoteRef
)

func (bt BlockType) String() string {
	names := map[BlockType]string{
		BlockUnknown:           "Unknown",
		BlockParagraph:          "Paragraph",
		BlockList:               "List",
		BlockDelimited:           "Delimited",
		BlockSection:             "Section",
		BlockAttribute:           "Attribute",
		BlockComment:             "Comment",
		BlockMacro:               "Macro",
		BlockSidebar:             "Sidebar",
		BlockCallout:             "Callout",
		BlockTable:               "Table",
		BlockThematicBreak:       "ThematicBreak",
		BlockPageBreak:           "PageBreak",
		BlockPassthrough:          "Passthrough",
		BlockLiteral:             "Literal",
		BlockVerbatim:            "Verbatim",
		BlockExample:             "Example",
		BlockQuote:               "Quote",
		BlockAdmonition:          "Admonition",
		BlockStyleType:           "Style",
		BlockAnchor:              "Anchor",
		BlockFrontMatter:         "FrontMatter",
		BlockSeparator:           "Separator",
		BlockTitle:               "Title",
		BlockPreamble:            "Preamble",
		BlockManife:              "Manife",
		BlockRule:                "Rule",
		BlockInclude:             "Include",
		BlockIfDef:              "IfDef",
		BlockEndIf:              "EndIf",
		BlockMacroDef:           "MacroDef",
		BlockMacroDefBody:        "MacroDefBody",
		BlockReset:               "Reset",
		BlockDeprecated:          "Deprecated",
		BlockReplacement:         "Replacement",
		BlockTemplate:            "Template",
		BlockMetadata:            "Metadata",
		BlockIndexTerms:          "IndexTerms",
		BlockImage:               "Image",
		BlockVideo:               "Video",
		BlockAudio:               "Audio",
		BlockTOC:                 "TOC",
		BlockTodo:                "Todo",
		BlockDescriptionList:     "DescriptionList",
		BlockUlist:               "Ulist",
		BlockOlist:               "Olist",
		BlockCalloutList:         "CalloutList",
		BlockQandA:               "QandA",
		BlockGlossary:            "Glossary",
		BlockBibliography:        "Bibliography",
		BlockColophon:            "Colophon",
		BlockFloatingTitle:       "FloatingTitle",
		BlockReproducibleUnits:   "ReproducibleUnits",
		BlockSyntaxHighlighting:  "SyntaxHighlighting",
		BlockSourceHighlighting:  "SourceHighlighting",
		BlockPassthroughBlock:    "PassthroughBlock",
		BlockIndex:               "Index",
		BlockFootnote:            "Footnote",
		BlockFootnoteRef:         "FootnoteRef",
	}
	if name, ok := names[bt]; ok {
		return name
	}
	return "Unknown"
}

type BlockStyle struct {
	Name string
	Args []string
	Position reader.Position
}

func (bs BlockStyle) String() string {
	result := "["
	result += bs.Name
	if len(bs.Args) > 0 {
		result += " "
		result += strings.Join(bs.Args, ", ")
	}
	result += "]"
	return result
}

type Block struct {
	Type          BlockType
	Style         *BlockStyle
	Level         int
	Lines         []string
	Attributes    map[string]string
	Position      reader.Position
	SourceLocation *Position
	Caption       string
	ID            string
	Title         string
	Subtitle      string
	RefID         string
	Callouts      []*Block
	Blocks        []*Block
}

type Position struct {
	File   string
	Path   string
	Line    int
	Column  int
}

type ListType int

const (
	ListUnknown ListType = iota
	ListUnordered
	ListUnorderedBullet
	ListUnorderedDash
	ListUnorderedAlpha
	ListOrdered
	ListLabeled
	ListChecklist
	ListCallout
	ListQandA
	ListGlossary
	ListBibliography
	ListOther
)

func (lt ListType) String() string {
	names := map[ListType]string{
		ListUnknown:         "Unknown",
		ListUnordered:       "Unordered",
		ListUnorderedBullet: "Bullet",
		ListUnorderedDash:   "Dash",
		ListUnorderedAlpha:  "Alpha",
		ListOrdered:         "Ordered",
		ListLabeled:         "Labeled",
		ListChecklist:       "Checklist",
		ListCallout:         "Callout",
		ListQandA:           "QandA",
		ListGlossary:        "Glossary",
		ListBibliography:     "Bibliography",
		ListOther:           "Other",
	}
	if name, ok := names[lt]; ok {
		return name
	}
	return "Unknown"
}

type ListMarker struct {
	Type    ListType
	Marker  string
	Level   int
	Ordinal int
	Bullet  rune
	Compact bool
}

func (lm ListMarker) String() string {
	var content string
	if lm.Ordinal > 0 {
		content += strings.Repeat(string(lm.Marker), lm.Ordinal)
	}
	content += lm.Marker
	return content
}

type Section struct {
	Level      int
	Title      string
	ID         string
	Attributes map[string]string
	Blocks     []*Block
}

type Document struct {
	Title          string
	Attributes     map[string]string
	Blocks         []*Block
	SourceLocation *Position
}

type Parser struct {
	reader       *reader.Reader
	currentBlock *Block
	currentLevel map[string]int
	levelStack   []int
	blockStack   []*Block
	document     *Document
}

func NewParser(r *reader.Reader) *Parser {
	return &Parser{
		reader:       r,
		currentLevel: make(map[string]int),
		levelStack:   make([]int, 0),
		blockStack:   make([]*Block, 0),
		document:     &Document{
			Attributes: make(map[string]string),
		},
	}
}

func (p *Parser) Parse() (*Document, error) {
	p.document = &Document{
		SourceLocation: &Position{
			File: "",
			Path: "",
		},
		Attributes: make(map[string]string),
	}

	for p.reader.HasMoreLines() {
		block, err := p.parseBlock()
		if err != nil {
			return nil, err
		}

		if block != nil {
			p.document.Blocks = append(p.document.Blocks, block)
		}
	}

	return p.document, nil
}

func (p *Parser) parseBlock() (*Block, error) {
	line := p.reader.PeekLine()

	if strings.TrimSpace(line) == "" {
		p.reader.NextLine()
		return nil, nil
	}

	blockType := identifyBlockType(line)

	switch blockType {
	case BlockLiteral:
		return p.parseLiteralBlock()
	case BlockVerbatim:
		return p.parseVerbatimBlock()
	case BlockExample:
		return p.parseExampleBlock()
	case BlockQuote:
		return p.parseQuoteBlock()
	case BlockPassthrough:
		return p.parsePassthroughBlock()
	case BlockSidebar:
		return p.parseSidebarBlock()
	case BlockList:
		return p.parseListBlock()
	case BlockSection:
		line := p.reader.PeekLine()
		section := IdentifySectionHeader(line)
		p.reader.NextLine()
		if section != nil {
			return &Block{
				Type:  BlockSection,
				Level: section.Level,
				Title: section.Title,
				Lines: []string{line},
			}, nil
		}
		return nil, nil
	case BlockThematicBreak:
		p.reader.NextLine()
		return &Block{Type: blockType}, nil
	case BlockSeparator:
		p.reader.NextLine()
		return &Block{Type: blockType}, nil
	case BlockPageBreak:
		p.reader.NextLine()
		return &Block{Type: blockType}, nil
	case BlockComment:
		p.reader.NextLine()
		return &Block{Type: blockType}, nil
	case BlockAttribute:
		p.reader.NextLine()
		return &Block{Type: blockType}, nil
	case BlockParagraph:
		return p.parseParagraphBlock()
	default:
		return p.parseParagraphBlock()
	}
}

func identifyBlockType(line string) BlockType {
	trimmed := strings.TrimSpace(line)

	if len(trimmed) >= 4 && isDelimiterBlock(trimmed) {
		return getDelimiterType(trimmed)
	}

	if strings.HasPrefix(trimmed, "//") {
		return BlockComment
	}

	// Check for section headers: = Title, == Section, === Subsection, etc.
	// Must start with = followed by at least one space after the = symbols
	if strings.HasPrefix(trimmed, "=") && len(trimmed) > 1 {
		// Count leading = characters
		equalsCount := 0
		for _, c := range trimmed {
			if c == '=' {
				equalsCount++
			} else {
				break
			}
		}
		// After the = characters, there must be a space
		if equalsCount < len(trimmed) && trimmed[equalsCount] == ' ' {
			return BlockSection
		}
	}

	if strings.HasPrefix(trimmed, ":") && len(trimmed) > 1 {
		return BlockAttribute
	}

	if len(trimmed) >= 3 && strings.HasPrefix(trimmed, "---") {
		return BlockSeparator
	}

	if trimmed == "<<<" {
		return BlockPageBreak
	}

	if IsListLine(trimmed) {
		return BlockList
	}

	return BlockParagraph
}

func isDelimiterBlock(line string) bool {
	if len(line) < 4 {
		return false
	}

	first := line[0]
	for i := 1; i < len(line); i++ {
		if line[i] != first {
			return false
		}
	}

	return isDelimiterChar(first)
}

func isDelimiterChar(c byte) bool {
	return c == '-' || c == '=' || c == '_' || c == '+' || c == '*' || c == '.' || c == '/'
}

func getDelimiterType(line string) BlockType {
	if len(line) == 0 {
		return BlockUnknown
	}

	first := line[0]
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
	case '.':
		return BlockLiteral
	case '/':
		return BlockComment
	default:
		return BlockUnknown
	}
}

func (p *Parser) parseLiteralBlock() (*Block, error) {
	line := p.reader.NextLine()

	block := &Block{
		Type:  BlockLiteral,
		Lines: []string{line},
		Level: p.currentLevel["literal"],
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()
		if strings.TrimSpace(nextLine) == "...." {
			p.reader.NextLine()
			break
		}
		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parseVerbatimBlock() (*Block, error) {
	line := p.reader.NextLine()

	block := &Block{
		Type:  BlockVerbatim,
		Lines: []string{line},
		Level: p.currentLevel["verbatim"],
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()
		if strings.TrimSpace(nextLine) == "----" {
			p.reader.NextLine()
			break
		}
		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parseExampleBlock() (*Block, error) {
	line := p.reader.NextLine()

	block := &Block{
		Type:  BlockExample,
		Lines: []string{line},
		Level: p.currentLevel["example"],
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()
		if strings.TrimSpace(nextLine) == "====" {
			p.reader.NextLine()
			break
		}
		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parseQuoteBlock() (*Block, error) {
	line := p.reader.NextLine()

	block := &Block{
		Type:  BlockQuote,
		Lines: []string{line},
		Level: p.currentLevel["quote"],
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()
		if strings.TrimSpace(nextLine) == "____" {
			p.reader.NextLine()
			break
		}
		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parsePassthroughBlock() (*Block, error) {
	line := p.reader.NextLine()

	block := &Block{
		Type:  BlockPassthrough,
		Lines: []string{line},
		Level: p.currentLevel["passthrough"],
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()
		if strings.TrimSpace(nextLine) == "++++" {
			p.reader.NextLine()
			break
		}
		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parseSidebarBlock() (*Block, error) {
	line := p.reader.NextLine()

	block := &Block{
		Type:  BlockSidebar,
		Lines: []string{line},
		Level: p.currentLevel["sidebar"],
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()
		if strings.TrimSpace(nextLine) == "****" {
			p.reader.NextLine()
			break
		}
		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parseListBlock() (*Block, error) {
	block := &Block{
		Type:  BlockList,
		Level: p.currentLevel["list"] + 1,
		Lines: []string{},
		Blocks: []*Block{},
	}

	for p.reader.HasMoreLines() {
		nextLine := p.reader.PeekLine()

		if strings.TrimSpace(nextLine) == "" {
			p.reader.NextLine()
			continue
		}

		if !IsListLine(nextLine) {
			break
		}

		marker := IdentifyListMarker(nextLine)
		if marker == nil {
			break
		}

		indent := getIndent(nextLine)
		if indent > block.Level {
			block.Level = indent
		}

		block.Lines = append(block.Lines, p.reader.NextLine())
	}

	return block, nil
}

func (p *Parser) parseParagraphBlock() (*Block, error) {
	lines := []string{}

	for p.reader.HasMoreLines() {
		line := p.reader.PeekLine()
		trimmed := strings.TrimSpace(line)

		// Debug

		// Check for stopping conditions first
		if trimmed == "" {
			break
		}
		if isDelimiterBlock(trimmed) {
			break
		}
		if strings.HasPrefix(trimmed, "=") && len(trimmed) > 1 {
			break
		}

		// Line is valid - consume it and add to paragraph
		lines = append(lines, p.reader.NextLine())
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty paragraph")
	}

	return &Block{
		Type:  BlockParagraph,
		Lines: lines,
	}, nil
}

func IsListLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}

	first := trimmed[0]
	if first == '-' || first == '*' {
		return true
	}

	if first == '.' {
		return true
	}

	if strings.Contains(trimmed, "::") {
		return true
	}

	return false
}

func getIndent(line string) int {
	indent := 0
	for _, c := range line {
		if c == ' ' || c == '\t' {
			indent++
		} else {
			break
		}
	}
	return (indent / 2) + 1
}

func IsDelimitedBlock(line string) bool {
	trimmed := strings.TrimSpace(line)
	return len(trimmed) >= 4 && isDelimiterBlock(trimmed)
}

type DelimiterInfo struct {
	Char byte
	Type DelimiterType
}

type DelimiterType int

const (
	DelimiterUnknown DelimiterType = iota
	DelimiterLiteral
	DelimiterVerbatim
	DelimiterExample
	DelimiterQuote
	DelimiterPassthrough
	DelimiterSidebar
	DelimiterComment
)

func IdentifyDelimiter(line string) DelimiterInfo {
	trimmed := strings.TrimSpace(line)

	if len(trimmed) < 4 {
		return DelimiterInfo{Type: DelimiterUnknown}
	}

	first := trimmed[0]
	for i := 1; i < len(trimmed); i++ {
		if trimmed[i] != first {
			return DelimiterInfo{Type: DelimiterUnknown}
		}
	}

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

func IdentifySectionHeader(line string) *Section {
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "=") {
		return nil
	}

	equalsCount := 0
	for _, c := range trimmed {
		if c == '=' {
			equalsCount++
		} else {
			break
		}
	}

	if equalsCount < 1 || equalsCount > len(trimmed) {
		return nil
	}

	if equalsCount >= len(trimmed) || trimmed[equalsCount] != ' ' {
		return nil
	}

	level := equalsCount - 1
	if level > 6 {
		level = 6
	}

	title := strings.TrimSpace(trimmed[equalsCount:])

	return &Section{
		Level:      level,
		Title:      title,
		ID:          "",
		Attributes:   make(map[string]string),
		Blocks:      []*Block{},
	}
}

type AttributeEntry struct {
	Name  string
	Value string
	IsSet bool
}

func IdentifyAttributeEntry(line string) *AttributeEntry {
	trimmed := strings.TrimSpace(line)

	if len(trimmed) < 2 {
		return nil
	}

	if trimmed[0] != 58 {
		return nil
	}

	remaining := trimmed[1:]

	unsetIndex := strings.Index(remaining, "!")
	setIndex := strings.Index(remaining, ":")

	nameEnd := -1
	isSet := true
	var delimiterPos int

	if unsetIndex >= 0 && setIndex == unsetIndex+1 {
		// `!:` pattern - name ends at !, could be unset or override
		// setIndex is relative to remaining, so add 1 for position in trimmed
		delimiterPos = setIndex + 1 // Position of : in trimmed
		nameEnd = unsetIndex + 1 // Include ! in name
		// Check if there's a value after the delimiter
		if delimiterPos+1 < len(trimmed) && trimmed[delimiterPos+1] == 32 {
			// `!: value` pattern - override, value exists
			isSet = true
		} else {
			// `!:` pattern - unset, no value after
			isSet = false
		}
	} else if setIndex >= 0 {
		// Normal `:` pattern (no ! involved)
		delimiterPos = setIndex + 1
		nameEnd = setIndex
	}

	if nameEnd < 0 {
		return nil
	}

	name := remaining[:nameEnd]

	if len(name) == 0 {
		return nil
	}

	firstChar := name[0]
	isLower := firstChar >= 97 && firstChar <= 122
	isUpper := firstChar >= 65 && firstChar <= 90
	isUnderscore := firstChar == 95

	if !isLower && !isUpper && !isUnderscore {
		return nil
	}

	var value string
	valueStart := delimiterPos + 2 // Skip past ": " or "!: "
	if valueStart < len(trimmed) {
		value = strings.TrimSpace(trimmed[valueStart:])
	}

	return &AttributeEntry{
		Name:  name,
		Value: value,
		IsSet: isSet,
	}
}

func IdentifyListMarker(line string) *ListMarker {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return nil
	}

	first := trimmed[0]

	switch first {
	case '-':
		return &ListMarker{
			Type:   ListUnordered,
			Marker: "-",
			Level:   1,
			Bullet:  '-',
		}
	case '*':
		return &ListMarker{
			Type:   ListUnordered,
			Marker: "*",
			Level:   1,
			Bullet:  '*',
		}
	}

	if len(trimmed) > 1 && first == 111 && trimmed[1] == 32 {
		return &ListMarker{
			Type:   ListUnordered,
			Marker: "o",
			Level:   1,
			Bullet: 111,
		}
	}

	if first == 46 {
		dotCount := 0
		for i := 0; i < len(trimmed); i++ {
			if trimmed[i] == 46 {
				dotCount++
			} else {
				break
			}
		}
		if dotCount <= len(trimmed) && dotCount < len(trimmed) && trimmed[dotCount] == 32 {
			return &ListMarker{
				Type:    ListOrdered,
				Marker:  strings.Repeat(".", dotCount),
				Level:   dotCount,
				Ordinal: dotCount,
				Compact: true,
			}
		}
	}

	if strings.Contains(trimmed, "::") {
		doubleColonIndex := strings.Index(trimmed, "::")
		if doubleColonIndex > 0 {
			return &ListMarker{
				Type:   ListLabeled,
				Marker: "::",
				Level:   1,
			}
		}
	}

	return nil
}
