// Package parser provides AsciiDoc parsing functionality.
package parser

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/extension"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
	"github.com/haimiyahya/asciidoc-parser-go/internal/processor"
	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
)

// listStackEntry tracks a list in the nesting stack along with its level and type.
type listStackEntry struct {
	list      *ast.NodeList
	level     int
	blockType reader.BlockType
}

// Parser parses AsciiDoc source into an AST.
type Parser struct {
	// reader is the source reader.
	reader *reader.Reader

	// classifier classifies lines into block types.
	classifier *reader.LineClassifier

	// options configures parser behavior.
	options []ParserOption

	// Section tracking state (stack of open sections)
	sectionStack   []*ast.NodeSection
	currentSection *ast.NodeSection

	// List tracking state
	listStack         []listStackEntry
	currentList      *ast.NodeList
	currentListBlockType reader.BlockType
	currentListLevel int

	// Block anchor tracking - [[id]] before a section applies to that section
	pendingAnchorID string

	// Block style tracking - [style] before a section applies to that section
	pendingBlockStyle string
	pendingBlockStyleAttrs map[string]string

	// Bibliography tracking
	currentBibliography *ast.BibliographyNode
	pendingBiblioAnchor *reader.AnchorInfo

	// Include processor handles include::[] directives
	includeProcessor *processor.IncludeProcessor

	// Conditional tracking - stack of conditional states
	conditionalStack []conditionalState

	// Current document being parsed (for attribute access)
	currentDocument *ast.NodeDocument

	// Extension registry for custom macros and processors
	extensionRegistry *extension.Registry
}

// conditionalState represents the state of a conditional directive.
type conditionalState struct {
	// active is true if the conditional is currently including content
	active bool

	// endifTarget is the attribute name for endif::[] matching
	endifTarget string

	// depth tracks nesting depth within this conditional
	depth int
}

// ParserOption configures a parser.
type ParserOption func(*Parser)

// NewParser creates a new Parser.
func NewParser(r *reader.Reader, opts ...ParserOption) *Parser {
	p := &Parser{
		reader:    r,
		classifier: reader.NewLineClassifier(),
		options:   opts,
	}
	// Apply options
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// NewParserFromReader creates a parser from an io.Reader.
func NewParserFromReader(r io.Reader, opts ...ParserOption) (*Parser, error) {
	rd, err := reader.NewReaderFromReader(r)
	if err != nil {
		return nil, err
	}
	return NewParser(rd, opts...), nil
}

// NewParserFromString creates a parser from a string.
func NewParserFromString(source string, opts ...ParserOption) (*Parser, error) {
	rd, err := reader.NewReader(source)
	if err != nil {
		return nil, err
	}
	return NewParser(rd, opts...), nil
}

// WithIncludeProcessor sets an include processor for handling include::[] directives.
func WithIncludeProcessor(ip *processor.IncludeProcessor) ParserOption {
	return func(p *Parser) {
		p.includeProcessor = ip
	}
}

// WithBaseDir sets the base directory for resolving relative paths (e.g., includes).
func WithBaseDir(dir string) ParserOption {
	return func(p *Parser) {
		p.reader.SetDir(dir)
	}
}

// WithExtensionRegistry sets an extension registry for custom macros and processors.
func WithExtensionRegistry(registry *extension.Registry) ParserOption {
	return func(p *Parser) {
		p.extensionRegistry = registry
	}
}

// Parse parses the AsciiDoc source into a document AST.
func (p *Parser) Parse() (*ast.NodeDocument, error) {
	doc := &ast.NodeDocument{
		Attributes: make(map[string]string),
		Blocks:     make([]ast.Node, 0),
	}
	p.currentDocument = doc
	p.conditionalStack = nil

	// Track current paragraph lines being accumulated
	var paragraphLines []string
	var paragraphLineno int

	// Track whether we're in a delimited block
	var inDelimitedBlock bool
	var delimitedBlockType reader.BlockType
	var delimitedBlockLines []string
	var delimitedBlockLineno int

	// Track pending attribute line for delimited blocks (e.g., [options="header"] before table)
	var pendingAttributeLine string
	// Track pending caption line for tables (e.g., .Table Title before |===)
	var pendingCaptionLine string

	for p.reader.HasMoreLines() {
		line := p.reader.PeekLine()
		lineno := p.reader.GetLineno()

		// Classify the line
		classification := p.classifier.ClassifyLine(line)

		// Handle attribute lines before delimited blocks (e.g., [options="header"] before |===)
		// This must be checked before classification since we consume the line
		if strings.HasPrefix(strings.TrimSpace(line), "[") && !inDelimitedBlock && pendingAttributeLine == "" {
			// Save the line and consume it, then check if next line(s) lead to a table
			savedLine := line
			p.reader.Advance()
			if p.reader.HasMoreLines() {
				nextLine := p.reader.PeekLine()
				trimmedNext := strings.TrimSpace(nextLine)
				nextClass := p.classifier.ClassifyLine(nextLine)

				// Check if followed by table delimiter OR caption (then delimiter)
				isTableFollowed := nextClass.Type == reader.BlockTable
				// Also check if next line is a caption - look one more line ahead
				isCaptionThenTable := false
				if strings.HasPrefix(trimmedNext, ".") && !strings.HasPrefix(trimmedNext, "..") &&
					!strings.HasPrefix(trimmedNext, ".#") {
					// Might be a caption - check if line after that is table delimiter
					p.reader.Advance() // Consume the potential caption
					if p.reader.HasMoreLines() {
						thirdLine := p.reader.PeekLine()
						thirdClass := p.classifier.ClassifyLine(thirdLine)
						if thirdClass.Type == reader.BlockTable {
							isCaptionThenTable = true
							pendingCaptionLine = nextLine // Save the caption
						}
					}
					if !isCaptionThenTable {
						// Not followed by table, put the line back
						p.reader.UnshiftLine(nextLine)
					}
				}

				if isTableFollowed || isCaptionThenTable {
					// This attribute line belongs to a table
					pendingAttributeLine = savedLine
					continue
				}
			}
			// Not followed by a table - this is a standalone attribute line
			// Restore it and let the normal attribute handling code process it
			p.reader.UnshiftLine(savedLine)
			// Continue to normal processing
		}

		// Handle caption lines before tables and listing blocks (e.g., .Title before |=== or ----)
		// Must come after attribute line check since captions can follow attributes
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, ".") && !strings.HasPrefix(trimmedLine, "..") &&
			!strings.HasPrefix(trimmedLine, ".#") && !inDelimitedBlock &&
			pendingCaptionLine == "" {
			// This might be a caption - save and check if followed by a block
			savedLine := line
			p.reader.Advance()
			if p.reader.HasMoreLines() {
				nextLine := p.reader.PeekLine()
				nextClass := p.classifier.ClassifyLine(nextLine)

				// Check if directly followed by a delimited block
				isDelimitedBlockFollowed := nextClass.Type == reader.BlockTable ||
					nextClass.Type == reader.BlockLiteral ||
					nextClass.Type == reader.BlockVerbatim

				// Check if followed by styled block (which might be followed by delimited block)
				isStyledBlockFollowed := nextClass.Style != nil

				if isDelimitedBlockFollowed {
					// This caption line belongs to the following block
					pendingCaptionLine = savedLine
					continue
				}

				if isStyledBlockFollowed {
					// Check if line after styled block is a delimited block
					p.reader.Advance() // Consume the styled block line
					if p.reader.HasMoreLines() {
						thirdLine := p.reader.PeekLine()
						thirdClass := p.classifier.ClassifyLine(thirdLine)
						if thirdClass.Type == reader.BlockLiteral ||
							thirdClass.Type == reader.BlockVerbatim ||
							thirdClass.Type == reader.BlockTable {
							// This caption line belongs to the styled+delimited block
							pendingCaptionLine = savedLine
							// Put the styled block line back (it will be processed normally)
							p.reader.UnshiftLine(nextLine)
							continue
						}
					}
					// Not followed by delimited block, put both lines back
					p.reader.UnshiftLine(nextLine)
				}
			}
			// Not followed by a supported block - restore it
			p.reader.UnshiftLine(savedLine)
			// Continue to normal processing
		}

		// Handle delimited blocks
		if classification.Type.IsDelimitedBlock() {
			if !inDelimitedBlock {
				// Starting a delimited block
				inDelimitedBlock = true
				delimitedBlockType = classification.Type
				delimitedBlockLines = []string{}
				delimitedBlockLineno = lineno

				// For tables, add pending caption and attribute lines at the beginning
				// NOTE: In AsciiDoc, caption comes BEFORE attributes (e.g., .Title [cols=...] |===)
				// The table parser expects: attributes first, then caption
				if classification.Type == reader.BlockTable {
					if pendingCaptionLine != "" {
						delimitedBlockLines = append(delimitedBlockLines, pendingCaptionLine)
						pendingCaptionLine = ""
					}
					if pendingAttributeLine != "" {
						delimitedBlockLines = append(delimitedBlockLines, pendingAttributeLine)
						pendingAttributeLine = ""
					}
				}

				p.reader.Advance()
				continue
			} else {
				// Check if this is the closing delimiter
				// Same type means we're closing
				if classification.Type == delimitedBlockType {
					// Close the delimited block
					p.reader.Advance()

					// Extract caption for non-table blocks before creating the block
					var blockCaption string
					if classification.Type != reader.BlockTable && pendingCaptionLine != "" {
						blockCaption = pendingCaptionLine
						pendingCaptionLine = ""
					}

					block := p.createDelimitedBlock(delimitedBlockType, delimitedBlockLines, delimitedBlockLineno, blockCaption)
					if block != nil {
						// Add to current section if exists, otherwise to doc.Blocks
						p.addBlockToCurrentSection(doc, block)

						// Parse callout list after literal/verbatim/source blocks
						if literal, ok := block.(*ast.NodeLiteral); ok && len(literal.Callouts) > 0 {
							p.parseCalloutList(literal)
						} else if styled, ok := block.(*ast.StyledBlockNode); ok && (styled.Style == "source" || styled.Style == "listing") && len(styled.Callouts) > 0 {
							p.parseCalloutListForStyled(styled)
						}
					}
					inDelimitedBlock = false
					delimitedBlockLines = nil
					continue
				} else {
					// Different delimiter inside block - treat as content
					delimitedBlockLines = append(delimitedBlockLines, line)
					p.reader.Advance()
					continue
				}
			}
		}

		// If we're in a delimited block, accumulate lines
		if inDelimitedBlock {
			delimitedBlockLines = append(delimitedBlockLines, line)
			p.reader.Advance()
			continue
		}

		// Handle section headers
		if classification.Type == reader.BlockSection && classification.Section != nil {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph first
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			// For level 0 (document title), set the document header
			if classification.Section.Level == 0 && doc.Header == nil {
				doc.Header = &ast.DocumentHeader{Title: classification.Section.Title}
			}

			section := p.createSection(classification.Section, lineno)
			if section != nil {
				if sec, ok := section.(*ast.NodeSection); ok {
					if sec.Level > 0 {
						// Handle section nesting
						p.pushSection(doc, sec)
					}
				} else if bib, ok := section.(*ast.BibliographyNode); ok {
					// Add bibliography node to document
					p.addBlockToCurrentSection(doc, bib)
				}
			}
			p.reader.Advance()
			continue
		}

		// Handle attribute entries
		if classification.Type == reader.BlockAttribute && classification.Attribute != nil {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			attr := classification.Attribute
			if attr.IsSet {
				doc.Attributes[attr.Name] = attr.Value
			} else {
				delete(doc.Attributes, attr.Name)
			}

			// Check for document title attribute
			if attr.Name == "title" && doc.Header == nil {
				doc.Header = &ast.DocumentHeader{Title: attr.Value}
			}

			p.reader.Advance()
			continue
		}

		// Handle conditional directives (ifdef, ifndef, ifeval, endif)
		if classification.Type == reader.BlockConditionalIfdef ||
			classification.Type == reader.BlockConditionalIfndef ||
			classification.Type == reader.BlockConditionalIfeval {

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			p.handleConditionalDirective(classification, doc)
			p.reader.Advance()
			continue
		}

		// Check for endif::[] directive
		if p.isEndifDirective(line) {
			p.handleEndifDirective(line)
			p.reader.Advance()
			continue
		}

		// Check if we're currently in an inactive conditional
		if p.isInConditionalSkip() {
			// Skip all content until we find the matching endif
			p.reader.Advance()
			continue
		}

		// Handle styled blocks (pass::[], sidebar::[], verse::[], etc.)
		if classification.Type == reader.BlockStyledBlock {
			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			block := p.createStyledBlock(classification.StyleBlock, lineno)
			if block != nil {
				p.addBlockToCurrentSection(doc, block)
			}
			p.reader.Advance()
			continue
		}

		// Handle block styles ([style]) that apply to the next block
		if classification.Type == reader.BlockStyle && classification.Style != nil {
			// Store the style to apply to the next section
			p.pendingBlockStyle = classification.Style.Name
			p.pendingBlockStyleAttrs = classification.Style.Attributes
			p.reader.Advance()
			continue
		}

		// Handle list items
		if classification.Type.IsListItem() {
			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			// Check if we need to close the current list
			if p.currentList == nil {
				// No current list - need to check if this item starts a new list
				if classification.List != nil {
					// Check if this is a bibliography entry (starts with [[[)
					if strings.HasPrefix(classification.List.Text, "[[[") {
						// This is a bibliography entry - start list and add immediately
						p.startNewList(classification, lineno, doc)
						p.addListItemToList(classification, lineno)
					} else {
						// Regular list item - start list but don't add yet
						p.startNewList(classification, lineno, doc)
					}
				}
			} else {
				// We have a current list - check if this item belongs to it
				itemInfo := classification.List
				sameType := (itemInfo.Type == p.currentListBlockType)
				sameLevel := (itemInfo.Level == p.currentListLevel)

				if sameType && sameLevel {
					// Same list - add item to it
					p.addListItemToList(classification, lineno)
				} else if itemInfo.Level > p.currentListLevel {
					// Nested list - add as child of current list item
					p.addNestedList(classification, lineno, doc)
				} else if itemInfo.Level < p.currentListLevel {
					// Going back up the hierarchy - pop from stack until we find the right level
					p.popListStackToLevel(itemInfo.Level)
					// After popping, add the item to the now-current list
					p.addListItemToList(classification, lineno)
				} else {
					// Different list type at same level - close current and start new
					p.closeCurrentList(doc)
					if classification.List != nil {
						p.startNewList(classification, lineno, doc)
					}
				}
			}

			p.reader.Advance()
			continue
		}

		// Handle blank lines - they terminate paragraphs and lists
		if classification.Type == reader.BlockBlank {
			// Close any open list
			p.closeCurrentList(doc)

			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}
			p.reader.Advance()
			continue
		}

		// Handle comments (skip them)
		if classification.Type == reader.BlockComment {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}
			p.reader.Advance()
			continue
		}

		// Handle block anchors ([#id] or [[id]] or [[[id]]] bibliography)
		if classification.Type == reader.BlockAnchor && classification.Anchor != nil {
			// Check if this is a bibliography anchor ([[[id]]])
			if classification.Anchor.IsBibliography {
				// Store as pending bibliography anchor for the next list item
				p.pendingBiblioAnchor = classification.Anchor
			} else {
				// Store the anchor ID to apply to the next section
				p.pendingAnchorID = classification.Anchor.ID
			}
			p.reader.Advance()
			continue
		}

		// Handle horizontal rules, page breaks
		if classification.Type == reader.BlockHorizontalRule || classification.Type == reader.BlockPageBreak {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}
			p.reader.Advance()
			continue
		}

		// Handle block macros
		if classification.Type == reader.BlockMacro {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			// Check if this is an include directive
			if classification.Macro != nil && classification.Macro.Target == "include" && p.includeProcessor != nil {
				// Get the line content
				line := classification.Original
				directive, err := processor.ParseInclude(line, lineno)
				if err == nil {
					// Get base directory from reader or default to current directory
					baseDir := p.reader.Dir()
					if baseDir == "" {
						baseDir = "."
					}
					// Create a new processor for this include with proper base dir
					includeProc := processor.NewIncludeProcessor(baseDir)

					// Process the include
					content, err := includeProc.Process(directive)
					if err != nil {
						// Include error - consume directive and continue
						p.reader.Advance()
						continue
					}
					if content != "" {
						// Split content into lines and inject into reader
						includedLines := strings.Split(content, "\n")
						// Consume the include directive line first, then inject
						p.reader.Advance()
						p.reader.InjectLines(includedLines)
						// Continue to next iteration without advancing again
						continue
					}
				}
				// If include failed or no content, just consume and continue
				p.reader.Advance()
				continue
			}

			// Create macro node (for non-include macros)
			macro := p.createMacro(classification.Macro, lineno)
			if macro != nil {
				p.addBlockToCurrentSection(doc, macro)
			}

			p.reader.Advance()
			continue
		}

		// Handle admonitions
		if classification.Type == reader.BlockAdmonition {
			// Close any open list first
			p.closeCurrentList(doc)

			// Flush any pending paragraph
			if len(paragraphLines) > 0 {
				para := p.createParagraph(paragraphLines, paragraphLineno)
				if para != nil {
					p.addBlockToCurrentSection(doc, para)
				}
				paragraphLines = nil
			}

			// Collect continuation lines for the admonition
			// Admonitions can span multiple lines until a blank line or different block
			admonitionLines := []string{classification.Admonition.Text}
			p.reader.Advance()

			for p.reader.HasMoreLines() {
				nextLine := p.reader.PeekLine()
				nextClass := p.classifier.ClassifyLine(nextLine)

				// Stop at blank lines or non-paragraph blocks
				if nextClass.Type == reader.BlockBlank ||
					nextClass.Type != reader.BlockParagraph {
					break
				}

				// Accumulate the continuation line
				p.reader.Advance()
				admonitionLines = append(admonitionLines, strings.TrimSpace(nextLine))
			}

			// Join the lines and create the admonition
			fullText := strings.Join(admonitionLines, " ")
			admonitionInfo := &reader.AdmonitionInfo{
				Kind: classification.Admonition.Kind,
				Text: fullText,
			}
			admonition := p.createAdmonition(admonitionInfo, lineno)
			if admonition != nil {
				p.addBlockToCurrentSection(doc, admonition)
			}

			continue
		}

		// Default: accumulate paragraph lines
		if len(paragraphLines) == 0 {
			paragraphLineno = lineno
		}
		paragraphLines = append(paragraphLines, line)
		p.reader.Advance()
	}

	// Flush any remaining paragraph
	if len(paragraphLines) > 0 {
		para := p.createParagraph(paragraphLines, paragraphLineno)
		if para != nil {
			p.addBlockToCurrentSection(doc, para)
		}
	}

	// Close any remaining open list
	p.closeCurrentList(doc)

	// Run tree processors if an extension registry is configured
	if p.extensionRegistry != nil {
		if err := p.runTreeProcessors(doc); err != nil {
			return nil, err
		}
		// Run block processors
		if err := p.runBlockProcessors(doc); err != nil {
			return nil, err
		}
	}

	return doc, nil
}

// createParagraph creates a paragraph node from accumulated lines.
func (p *Parser) createParagraph(lines []string, lineno int) ast.Node {
	if len(lines) == 0 {
		return nil
	}

	// Join lines with spaces
	content := strings.Join(lines, " ")

	// Parse inline markup within paragraph
	inlineParser := inline.NewParser(content)
	inlineNodes := inlineParser.Parse()

	// Strip role syntax [.role] from text and adjust inline node positions
	cleanedContent, offset := stripRoleSyntaxWithOffset(content)

	// Convert inline.Node slice to []interface{} for storage
	// Filter out NodeText nodes since they're already in the Text field
	nodes := make([]interface{}, 0, len(inlineNodes))
	for _, node := range inlineNodes {
		if node.Type != inline.NodeText {
			nodes = append(nodes, node)
		}
	}

	// Adjust inline node positions based on how much text was removed
	// Process only the filtered nodes (non-Text nodes)
	var adjustedNodes []interface{}
	for _, node := range nodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			// Create a copy of the node with adjusted positions
			adjustedNode := *inlineNode
			adjustedNode.StartPos -= getOffset(offset, adjustedNode.StartPos, len(content))
			adjustedNode.Position -= getOffset(offset, adjustedNode.Position, len(content))
			adjustedNodes = append(adjustedNodes, &adjustedNode)
		}
	}

	return &ast.NodeParagraph{
		Text:        cleanedContent,
		InlineNodes: adjustedNodes,
		Pos:  ast.Position{Line: lineno},
	}
}

// getOffset gets the offset for a position, defaulting to 0 if not found
func getOffset(offset map[int]int, pos, maxPos int) int {
	if off, exists := offset[pos]; exists {
		return off
	}
	// For positions past the end of the map, find the last offset
	for i := pos; i <= maxPos; i++ {
		if off, exists := offset[i]; exists {
			return off
		}
	}
	return 0
}

// stripInlineMarkupWithOffset removes inline markup delimiters from text and returns
// both the cleaned text and a map of position offsets.
// Handles: [.role] role syntax and ++text++ span delimiters
func stripInlineMarkupWithOffset(text string) (string, map[int]int) {
	// Pattern to match both [.role] and ++text++
	// This is a simplified approach - we process them sequentially

	// First, handle ++text++ spans
	spanRe := regexp.MustCompile(`\+\+[^+]+\+\+`)
	spanMatches := spanRe.FindAllStringIndex(text, -1)

	// Then handle [.role] syntax
	roleRe := regexp.MustCompile(`\[\.([^\]]+)\]`)
	roleMatches := roleRe.FindAllStringIndex(text, -1)

	// Combine all matches and sort by position
	allMatches := append(spanMatches, roleMatches...)
	// Sort matches by start position
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i][0] < allMatches[j][0]
	})

	// If no matches, return original text with no offsets
	if len(allMatches) == 0 {
		return text, make(map[int]int)
	}

	// Create offset map: for each position in original text, how much to subtract
	// to get the position in cleaned text
	offset := make(map[int]int)
	totalRemoved := 0
	lastEnd := 0
	var cleaned strings.Builder

	for _, match := range allMatches {
		start, end := match[0], match[1]

		// Copy text before this match
		cleaned.WriteString(text[lastEnd:start])

		// For ++text++ spans, we need to keep the inner text
		if strings.HasPrefix(text[start:], "++") {
			// Extract inner text (remove the ++ delimiters)
			innerText := text[start+2 : end-2]
			cleaned.WriteString(innerText)
			// The delimiters are 4 characters total (++ at start, ++ at end)
			totalRemoved += 4
		} else {
			// For [.role], remove the entire thing
			totalRemoved += (end - start)
		}

		// Update offsets for positions within the removed section
		for i := start; i < end; i++ {
			offset[i] = totalRemoved - (end - i) // Offset to account for removed chars
		}
		lastEnd = end
	}

	// Copy remaining text
	cleaned.WriteString(text[lastEnd:])

	// For positions after the last match, set the offset to totalRemoved
	for i := lastEnd; i <= len(text); i++ {
		if _, exists := offset[i]; !exists {
			offset[i] = totalRemoved
		}
	}

	// For positions before the first match, offset is 0
	if len(allMatches) > 0 {
		for i := 0; i < allMatches[0][0]; i++ {
			offset[i] = 0
		}
	}

	return cleaned.String(), offset
}

// stripRoleSyntaxWithOffset removes [.role] patterns from text and returns
// both the cleaned text and a map of position offsets.
// Deprecated: Use stripInlineMarkupWithOffset instead.
func stripRoleSyntaxWithOffset(text string) (string, map[int]int) {
	return stripInlineMarkupWithOffset(text)
}

// generateSectionID creates a section ID from a section title.
// This follows Asciidoctor conventions:
// - Simple words (no spaces/special chars): no underscore prefix (e.g., "details")
// - Multi-word or special chars: underscore prefix added (e.g., "_section_one")
// - Lowercase, underscores for spaces, removes special characters
func generateSectionID(title string) string {
	// Convert to lowercase
	id := strings.ToLower(title)

	// Replace spaces and special chars with underscores
	// Keep only alphanumeric and underscores
	reg := regexp.MustCompile(`[^a-z0-9_]+`)
	id = reg.ReplaceAllString(id, "_")

	// Trim underscores from start/end
	id = strings.Trim(id, "_")

	// Collapse multiple underscores
	reg = regexp.MustCompile(`_+`)
	id = reg.ReplaceAllString(id, "_")

	// Add underscore prefix ONLY if the ID contains underscores
	// This matches Asciidoctor behavior: simple words get no prefix,
	// multi-word titles get underscore prefix
	if strings.Contains(id, "_") {
		id = "_" + id
	}

	// Fallback if ID is empty
	if id == "" {
		id = "_"
	}

	return id
}

// createSection creates a section node from section info.
func (p *Parser) createSection(info *reader.SectionInfo, lineno int) ast.Node {
	if info == nil {
		return nil
	}

	// Create a copy of attributes to avoid modifying the original
	attrs := make(map[string]string)
	for k, v := range info.Attributes {
		attrs[k] = v
	}

	// Check if there's a pending block style from a previous [style] line
	// If no style attribute is set, use the pending block style
	if _, hasStyle := attrs["style"]; !hasStyle && p.pendingBlockStyle != "" {
		attrs["style"] = p.pendingBlockStyle
		p.pendingBlockStyle = "" // Consume the pending style
	} else {
		// Clear pending style even if not used (section has its own style)
		p.pendingBlockStyle = ""
	}

	// Check if this is a bibliography section (has style="bibliography" attribute)
	isBibliography := false
	if style, ok := attrs["style"]; ok && style == "bibliography" {
		isBibliography = true
	}

	// Use explicit ID if provided, otherwise use pending anchor ID, then auto-generate from title
	sectionID := info.ID
	if sectionID == "" {
		// Check if there's a pending block anchor ID
		if p.pendingAnchorID != "" {
			sectionID = p.pendingAnchorID
			p.pendingAnchorID = "" // Consume the pending anchor
		} else {
			sectionID = generateSectionID(info.Title)
		}
	} else {
		// Explicit ID was provided, clear any pending anchor
		p.pendingAnchorID = ""
	}

	// For level 0 (document title), set the document header
	if info.Level == 0 {
		return &ast.NodeSection{
			Level:      0,
			Title:       info.Title,
			ID:          sectionID,
			Attributes:  attrs,
			Pos:         ast.Position{Line: lineno},
		}
	}

	// If this is a bibliography section, create a BibliographyNode instead
	if isBibliography {
		// Ensure bibliography section ID has underscore prefix (Asciidoctor compatibility)
		if sectionID != "" && !strings.HasPrefix(sectionID, "_") {
			sectionID = "_" + sectionID
		}
		p.currentBibliography = &ast.BibliographyNode{
			Title:      info.Title,
			ID:         sectionID,
			Entries:    make([]*ast.BibliographyEntryNode, 0),
			Attributes: attrs,
			Pos:        ast.Position{Line: lineno},
		}
		return p.currentBibliography
	}

	return &ast.NodeSection{
		Level:      info.Level,
		Title:       info.Title,
		ID:          sectionID,
		Attributes:  attrs,
		Pos:         ast.Position{Line: lineno},
	}
}

// createListItem creates a list item node from list info.
func (p *Parser) createListItem(info *reader.ListInfo, lineno int) ast.Node {
	if info == nil {
		return nil
	}

	// For labeled lists, Text contains the term (before ::), and Definition contains the definition
	text := info.Text
	if info.Type == reader.BlockListLabeled && info.Term != "" {
		text = info.Term
	}

	// Parse inline markup within list item text
	inlineParser := inline.NewParser(text)
	inlineNodes := inlineParser.Parse()

	// Convert inline.Node slice to []interface{} for storage
	// Filter out NodeText nodes since they're already in the Text field
	nodes := make([]interface{}, 0, len(inlineNodes))
	for _, node := range inlineNodes {
		if node.Type != inline.NodeText {
			nodes = append(nodes, node)
		}
	}

	// For labeled lists, also parse inline markup in the definition
	var definitionNodes []interface{}
	if info.Type == reader.BlockListLabeled && info.Text != "" {
		defParser := inline.NewParser(info.Text)
		defNodes := defParser.Parse()
		definitionNodes = make([]interface{}, 0, len(defNodes))
		for _, node := range defNodes {
			if node.Type != inline.NodeText {
				definitionNodes = append(definitionNodes, node)
			}
		}
	}

	return &ast.NodeListItem{
		Kind:           ast.TypeListItem,
		Marker:          info.Marker,
		Level:           info.Level,
		Ordinal:         info.Ordinal,
		Text:            text,
		Term:            info.Term,
		Definition:      info.Text, // For labeled lists, Text contains the definition
		DefinitionNodes: definitionNodes,
		InlineNodes:     nodes,
		Pos:             ast.Position{Line: lineno},
	}
}

// createBibliographyEntry creates a bibliography entry node.
func (p *Parser) createBibliographyEntry(info *reader.ListInfo, lineno int) *ast.BibliographyEntryNode {
	if info == nil {
		return nil
	}

	// Parse bibliography anchor from the text
	// Format: [[[label]]] or [[[label,xreftext]]]
	text := info.Text
	var label, xrefText, entryText string

	// Debug output
	// fmt.Printf("[DEBUG] createBibliographyEntry: text=%q\n", text)

	// Check if text starts with [[[ and ends with ]]]
	if strings.HasPrefix(text, "[[[") && strings.Contains(text, "]]]") {
		// Find the closing ]]]
		endIdx := strings.Index(text, "]]]")
		if endIdx != -1 {
			// Extract content between [[[ and ]]]
			anchorContent := text[3:endIdx]
			restOfText := strings.TrimSpace(text[endIdx+3:])

			// Split on comma to get label and optional xreftext
			parts := strings.SplitN(anchorContent, ",", 2)
			label = parts[0]
			if len(parts) == 2 {
				xrefText = strings.TrimSpace(parts[1])
			}
			entryText = restOfText
		} else {
			// Invalid format, treat as regular text
			entryText = text
		}
	} else {
		// Not a bibliography entry format, but we're in a bibliography section
		// Treat the entire text as the entry with no label
		entryText = text
	}

	// If no label was found, this isn't a valid bibliography entry
	if label == "" {
		return nil
	}

	// Parse inline markup within the entry text
	inlineParser := inline.NewParser(entryText)
	inlineNodes := inlineParser.Parse()

	// Convert inline.Node slice to []interface{} for storage
	nodes := make([]interface{}, 0, len(inlineNodes))
	for _, node := range inlineNodes {
		if node.Type != inline.NodeText {
			nodes = append(nodes, node)
		}
	}

	return &ast.BibliographyEntryNode{
		Label:        label,
		XRefText:     xrefText,
		Text:         entryText,
		InlineNodes:  nodes,
		Pos:         ast.Position{Line: lineno},
	}
}

// createDelimitedBlock creates a delimited block node.
// createAdmonition creates an admonition node.
func (p *Parser) createAdmonition(admonition *reader.AdmonitionInfo, lineno int) ast.Node {
	if admonition == nil {
		return nil
	}

	return &ast.AdmonitionNode{
		Kind: admonition.Kind,
		Text: admonition.Text,
		Pos:  ast.Position{Line: lineno},
	}
}

// createMacro creates a block macro node.
func (p *Parser) createMacro(macro *reader.MacroInfo, lineno int) ast.Node {
	if macro == nil {
		return nil
	}

	// Check if a custom block macro processor is registered
	if p.extensionRegistry != nil {
		if processor, ok := p.extensionRegistry.GetBlockMacro(macro.Target); ok {
			// Convert attributes map to a format ParseMacroAttributes can handle
			attrString := ""
			if rawAttrs, ok := macro.Attributes["raw"]; ok {
				attrString = "[" + rawAttrs + "]"
			}
			attrs := extension.ParseMacroAttributes(attrString)
			attrMap := make(map[string]string)
			for k, v := range attrs.Named {
				attrMap[k] = v
			}
			for i, v := range attrs.Positional {
				attrMap[fmt.Sprintf("pos%d", i)] = v
			}
			// Copy all existing attributes
			for k, v := range macro.Attributes {
				attrMap[k] = v
			}

			// Call the custom processor
			// Note: macro.Path is the target/path after :: (e.g., "file.txt" in include::file.txt[])
			result, err := processor.Process(macro.Path, attrMap, []string{}, ast.Position{Line: lineno})
			if err != nil {
				// On error, fall back to standard macro node
			} else if result != nil {
				return result
			}
		}
	}

	// Default: create standard macro node
	return &ast.MacroNode{
		Kind:       ast.TypeMacro,
		Target:      macro.Target,
		Path:        macro.Path,
		Attributes:  macro.Attributes,
		Pos:         ast.Position{Line: lineno},
	}
}

// createStyledBlock creates a styled block node (pass::[], sidebar::[], verse::[], etc.).
func (p *Parser) createStyledBlock(styleBlock *reader.StyleBlockInfo, lineno int) ast.Node {
	if styleBlock == nil {
		return nil
	}

	switch styleBlock.Style {
	case "pass":
		return &ast.PassThroughNode{
			Content:    styleBlock.Content,
			Attributes: styleBlock.Attributes,
			Pos:        ast.Position{Line: lineno},
		}
	case "sidebar":
		return &ast.SidebarNode{
			Title:      "", // Can be parsed from content if needed
			Content:    styleBlock.Content,
			Attributes: styleBlock.Attributes,
			Pos:        ast.Position{Line: lineno},
		}
	case "verse":
		return &ast.VerseNode{
			Content:    styleBlock.Content,
			Attributes: styleBlock.Attributes,
			Pos:        ast.Position{Line: lineno},
		}
	default:
		// For other styles (quote, example, listing, literal, source),
		// return a generic StyledBlockNode
		return &ast.StyledBlockNode{
			Style:      styleBlock.Style,
			Content:    styleBlock.Content,
			Attributes: styleBlock.Attributes,
			Pos:        ast.Position{Line: lineno},
		}
	}
}

// createDelimitedBlock creates a delimited block node.
// The caption parameter is the optional block title (e.g., from ".Title" before the block).
func (p *Parser) createDelimitedBlock(blockType reader.BlockType, lines []string, lineno int, caption string) ast.Node {
	content := strings.Join(lines, "\n")

	// Check if this is a source/listing block with pending style
	isSourceBlock := (p.pendingBlockStyle == "source" || p.pendingBlockStyle == "listing")

	if isSourceBlock && (blockType == reader.BlockLiteral || blockType == reader.BlockVerbatim) {
		// Parse callouts in source blocks
		calloutParser := NewCalloutParser()
		cleanedLines, callouts := calloutParser.ParseCalloutsInLiteral(lines)

		// Use cleaned content (callouts removed from comments)
		cleanedContent := strings.Join(cleanedLines, "\n")

		// Create a StyledBlockNode for syntax highlighting
		attrs := make(map[string]string)
		if p.pendingBlockStyleAttrs != nil {
			for k, v := range p.pendingBlockStyleAttrs {
				attrs[k] = v
			}
		}
		p.pendingBlockStyle = ""       // Consume the pending style
		p.pendingBlockStyleAttrs = nil // Clear attributes

		return &ast.StyledBlockNode{
			Style:      "source",
			Content:    cleanedContent,
			Attributes: attrs,
			Callouts:   callouts,
			Caption:    p.parseCaption(caption),
			Pos:        ast.Position{Line: lineno},
		}
	}

	switch blockType {
	case reader.BlockLiteral:
		// Parse callouts in literal blocks
		calloutParser := NewCalloutParser()
		cleanedLines, callouts := calloutParser.ParseCalloutsInLiteral(lines)
		p.pendingBlockStyle = ""       // Consume the pending style
		p.pendingBlockStyleAttrs = nil // Clear attributes
		return &ast.NodeLiteral{
			Lines:     cleanedLines,
			Callouts:  callouts,
			Caption:   p.parseCaption(caption),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockVerbatim:
		// Parse callouts in verbatim blocks
		calloutParser := NewCalloutParser()
		cleanedLines, callouts := calloutParser.ParseCalloutsInLiteral(lines)
		p.pendingBlockStyle = ""       // Consume the pending style
		p.pendingBlockStyleAttrs = nil // Clear attributes
		return &ast.NodeLiteral{
			Lines:     cleanedLines,
			Callouts:  callouts,
			Caption:   p.parseCaption(caption),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockExample:
		return &ast.NodeBlock{
			Delimiter: "=",
			Lines:    strings.Split(content, "\n"),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockQuote:
		// Check if this quote block has a verse style
		if p.pendingBlockStyle == "verse" {
			p.pendingBlockStyle = ""       // Consume the pending style
			p.pendingBlockStyleAttrs = nil // Clear attributes
			return &ast.VerseNode{
				Content: content,
				Pos:     ast.Position{Line: lineno},
			}
		}
		return &ast.NodeBlock{
			Delimiter: "_",
			Lines:    strings.Split(content, "\n"),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockSidebar:
		return &ast.NodeBlock{
			Delimiter: "*",
			Lines:    strings.Split(content, "\n"),
			Pos:       ast.Position{Line: lineno},
		}
	case reader.BlockPassthrough:
		p.pendingBlockStyle = ""       // Consume the pending style
		p.pendingBlockStyleAttrs = nil // Clear attributes
		return &ast.PassThroughNode{
			Content: content,
			Pos:     ast.Position{Line: lineno},
		}
	case reader.BlockTable:
		return p.createTable(lines, lineno)
	default:
		return &ast.NodeBlock{
			Lines: strings.Split(content, "\n"),
			Pos:   ast.Position{Line: lineno},
		}
	}
}

// createTable parses table content into a Table node.
// Tables use | as column separator.
// Supports delimited block syntax |===...|===.
func (p *Parser) createTable(lines []string, lineno int) ast.Node {
	if len(lines) == 0 {
		return nil
	}

	tableParser := NewTableParser()
	return tableParser.ParseTable(lines, lineno)
}

// closeCurrentList closes the current open list if any.
func (p *Parser) closeCurrentList(doc *ast.NodeDocument) {
	// If there are items on the stack, we have nested lists.
	// Add the outermost (first) list on the stack to the document.
	// The nested lists are already attached via NestedList pointers.
	if len(p.listStack) > 0 {
		// Get the outermost list (first item pushed)
		outermostEntry := p.listStack[0]
		p.addBlockToCurrentSection(doc, outermostEntry.list)

		// Clear pending block style if this was a styled list (e.g., qanda)
		if outermostEntry.list.Style != "" {
			p.pendingBlockStyle = ""
			p.pendingBlockStyleAttrs = nil
		}

		p.currentList = nil
		p.currentListBlockType = 0
		p.currentListLevel = 0
		p.listStack = nil
	} else if p.currentList != nil {
		// No nested lists, just add the current list
		p.addBlockToCurrentSection(doc, p.currentList)

		// Clear pending block style if this was a styled list (e.g., qanda)
		if p.currentList.Style != "" {
			p.pendingBlockStyle = ""
			p.pendingBlockStyleAttrs = nil
		}

		p.currentList = nil
		p.currentListBlockType = 0
		p.currentListLevel = 0
		p.listStack = nil
	}
}

// popListStackToLevel pops lists from the stack until we reach the target level.
// This is used when we encounter an item at a higher level (closer to root) than
// the current nested list.
func (p *Parser) popListStackToLevel(targetLevel int) {
	// Pop from stack until we find a list at or above the target level
	for len(p.listStack) > 0 {
		// Peek at the top of the stack
		topEntry := p.listStack[len(p.listStack)-1]

		if topEntry.level <= targetLevel {
			// Found a list at or above the target level
			// Restore the context from this entry
			p.currentList = topEntry.list
			p.currentListLevel = topEntry.level
			p.currentListBlockType = topEntry.blockType
			// Don't pop yet - we'll pop after processing
			return
		}

		// This list is deeper than target, pop it
		p.listStack = p.listStack[:len(p.listStack)-1]
	}

	// If we emptied the stack but didn't find a matching level,
	// we're back at the outermost list (level 1 or similar)
	// The currentList should already be set by the last pop
	if len(p.listStack) == 0 && p.currentList == nil {
		// This shouldn't happen in valid markup, but handle gracefully
		p.currentListLevel = 1
	}
}

// parseCalloutList parses callout descriptions that follow a literal block.
// Callout list items have the format: <1> Description here
func (p *Parser) parseCalloutList(literal *ast.NodeLiteral) {
	calloutParser := NewCalloutParser()
	calloutList := &ast.CalloutListNode{
		Items: make(map[int]*ast.CalloutNode),
	}

	// Look ahead for callout list items
	// They must immediately follow the literal block
	for p.reader.HasMoreLines() {
		line := p.reader.PeekLine()
		lineno := p.reader.GetLineno()

		// Stop if we hit a blank line or non-callout content
		if strings.TrimSpace(line) == "" {
			break
		}

		// Check if this is a callout list item
		if !calloutParser.IsCalloutListLine(line) {
			// Not a callout list, stop parsing
			// But don't consume the line - it belongs to the next block
			break
		}

		// Parse the callout list item
		num, desc := calloutParser.ParseCalloutList(line)
		if num > 0 {
			calloutList.Items[num] = &ast.CalloutNode{
				Number:      num,
				Description: desc,
				Pos:         ast.Position{Line: lineno},
			}
		}

		// Consume the callout list line
		p.reader.Advance()
	}

	// Merge descriptions into the literal block's callouts
	calloutParser.MergeCalloutDescriptions(literal, calloutList)
}

// parseCalloutListForStyled parses callout descriptions that follow a styled source block.
// Callout list items have the format: <1> Description here
func (p *Parser) parseCalloutListForStyled(styled *ast.StyledBlockNode) {
	calloutParser := NewCalloutParser()
	calloutList := &ast.CalloutListNode{
		Items: make(map[int]*ast.CalloutNode),
	}

	// Look ahead for callout list items
	// They must immediately follow the styled block
	for p.reader.HasMoreLines() {
		line := p.reader.PeekLine()
		lineno := p.reader.GetLineno()

		// Stop if we hit a blank line or non-callout content
		if strings.TrimSpace(line) == "" {
			break
		}

		// Check if this is a callout list item
		if !calloutParser.IsCalloutListLine(line) {
			// Not a callout list, stop parsing
			// But don't consume the line - it belongs to the next block
			break
		}

		// Parse the callout list item
		num, desc := calloutParser.ParseCalloutList(line)
		if num > 0 {
			calloutList.Items[num] = &ast.CalloutNode{
				Number:      num,
				Description: desc,
				Pos:         ast.Position{Line: lineno},
			}
		}

		// Consume the callout list line
		p.reader.Advance()
	}

	// Merge descriptions into the styled block's callouts
	calloutParser.MergeCalloutDescriptionsForStyled(styled, calloutList)
}

// startNewList starts a new list with the given item as its first element.
func (p *Parser) startNewList(classification *reader.Classification, lineno int, doc *ast.NodeDocument) {
	info := classification.List
	if info == nil {
		return
	}

	// Create a new list with this item as its first element
	listItem := p.createListItem(info, lineno)
	if listItem == nil {
		return
	}

	// Check if there's a pending block style (e.g., [qanda])
	listStyle := ""
	if p.pendingBlockStyle != "" {
		listStyle = p.pendingBlockStyle
		// Don't consume the style yet - let the list items consume it
		// The style will be cleared after the list is closed
	}

	// Create the list node - all lists use TypeList as the Kind
	p.currentList = &ast.NodeList{
		Kind:  ast.TypeList,
		Items: []ast.Node{listItem},
		Style: listStyle,
		Pos:   ast.Position{Line: lineno},
	}
	p.currentListBlockType = info.Type
	p.currentListLevel = info.Level

	// Note: Don't add to doc.Blocks yet - wait until list is closed
}

// addListItemToList adds a list item to the current list.
func (p *Parser) addListItemToList(classification *reader.Classification, lineno int) {
	info := classification.List
	if info == nil || p.currentList == nil {
		return
	}

	// Try to create a bibliography entry from the list item text
	// Bibliography entries are identified by [[[label]]] or [[[label,xreftext]]] syntax
	biblioEntry := p.createBibliographyEntry(info, lineno)
	if biblioEntry != nil {
		// This is a bibliography entry - ensure we have a bibliography node
		if p.currentBibliography == nil {
			// Create an implicit bibliography section
			p.currentBibliography = &ast.BibliographyNode{
				Title:      "Bibliography",
				ID:         "bibliography",
				Entries:    make([]*ast.BibliographyEntryNode, 0),
				Attributes: make(map[string]string),
				Pos:        ast.Position{Line: lineno},
			}
			// Add to document blocks if not already present
			p.addBlockToCurrentSection(p.currentDocument, p.currentBibliography)
		}

		p.currentBibliography.Entries = append(p.currentBibliography.Entries, biblioEntry)

		// Add entry to document's bibliography map for citation lookup
		if p.currentDocument != nil && p.currentDocument.BibliographyEntries == nil {
			p.currentDocument.BibliographyEntries = make(map[string]*ast.BibliographyEntryNode)
		}
		if p.currentDocument != nil {
			p.currentDocument.BibliographyEntries[biblioEntry.Label] = biblioEntry
		}

		// Close the current list since this item belongs to the bibliography, not the list
		// This prevents the empty/partial list from being added to the document
		p.currentList = nil
		p.currentListBlockType = 0
		p.currentListLevel = 0

		return // Don't add as a regular list item
	}

	listItem := p.createListItem(info, lineno)
	if listItem != nil {
		p.currentList.Items = append(p.currentList.Items, listItem)
	}
}

// addNestedList adds a nested list as a child of the current list's last item.
// It also updates the current list context so subsequent items are added to the nested list.
func (p *Parser) addNestedList(classification *reader.Classification, lineno int, doc *ast.NodeDocument) {
	info := classification.List
	if info == nil || p.currentList == nil {
		return
	}

	// Get the last item in the current list
	if len(p.currentList.Items) == 0 {
		return
	}

	lastItemIdx := len(p.currentList.Items) - 1
	lastItem := p.currentList.Items[lastItemIdx]

	// Check if last item already has a nested list of the same type
	// If so, add to that nested list instead of creating a new one
	item, ok := lastItem.(*ast.NodeListItem)
	if !ok {
		return
	}

	if item.NestedList != nil {
		// Add new item to existing nested list
		// Also update the current list context to point to this nested list
		listItem := p.createListItem(info, lineno)
		if listItem != nil {
			item.NestedList.Items = append(item.NestedList.Items, listItem)

			// Push current list onto stack and update to nested list
			p.listStack = append(p.listStack, listStackEntry{
				list:      p.currentList,
				level:     p.currentListLevel,
				blockType: p.currentListBlockType,
			})
			p.currentList = item.NestedList
			p.currentListLevel = info.Level
			p.currentListBlockType = info.Type
		}
		return
	}

	// Create a new nested list
	nestedList := &ast.NodeList{
		Kind:  ast.TypeList,
		Items: []ast.Node{},
		Pos:   ast.Position{Line: lineno},
	}

	// Add the new list item to the nested list
	listItem := p.createListItem(info, lineno)
	if listItem != nil {
		nestedList.Items = append(nestedList.Items, listItem)
	}

	// Attach the nested list to the parent list item
	item.NestedList = nestedList

	// Push current list onto stack and update to nested list
	p.listStack = append(p.listStack, listStackEntry{
		list:      p.currentList,
		level:     p.currentListLevel,
		blockType: p.currentListBlockType,
	})
	p.currentList = nestedList
	p.currentListLevel = info.Level
	p.currentListBlockType = info.Type
}

// Advance is a helper that consumes the next line.
func (p *Parser) Advance() bool {
	return p.reader.Advance()
}

// addBlockToCurrentSection adds a block to either the current section's children
// or to doc.Blocks if there's no current section.
func (p *Parser) addBlockToCurrentSection(doc *ast.NodeDocument, block ast.Node) {
	if p.currentSection != nil {
		p.currentSection.Children = append(p.currentSection.Children, block)
	} else {
		doc.Blocks = append(doc.Blocks, block)
	}
}

// pushSection adds a new section to the document, handling section nesting.
// When we see a new section:
// - Same level as current: close current section, start new one
// - Deeper level than current: it's a subsection of current
// - Shallower level than current: close sections until we're at the right level
func (p *Parser) pushSection(doc *ast.NodeDocument, section *ast.NodeSection) {
	// Close any open sections at or above this level
	p.closeSectionsToLevel(doc, section.Level)

	// Add this section to the appropriate parent
	if len(p.sectionStack) == 0 {
		// No current section - add to doc.Blocks
		doc.Blocks = append(doc.Blocks, section)
	} else {
		// Add as child of the current (parent) section
		parent := p.sectionStack[len(p.sectionStack)-1]
		parent.Children = append(parent.Children, section)
	}

	// This section becomes the current section
	p.sectionStack = append(p.sectionStack, section)
	p.currentSection = section
}

// closeSectionsToLevel closes all sections at or above the given level.
// When we encounter a new section at level L, we need to close all sections
// at level >= L (they're siblings or parents, not ancestors).
func (p *Parser) closeSectionsToLevel(doc *ast.NodeDocument, level int) {
	for len(p.sectionStack) > 0 && p.sectionStack[len(p.sectionStack)-1].Level >= level {
		// Pop the section from stack
		closed := p.sectionStack[len(p.sectionStack)-1]
		p.sectionStack = p.sectionStack[:len(p.sectionStack)-1]

		// If this was the current section, update currentSection
		if closed == p.currentSection {
			if len(p.sectionStack) > 0 {
				p.currentSection = p.sectionStack[len(p.sectionStack)-1]
			} else {
				p.currentSection = nil
			}
		}
	}
}

// isEndifDirective checks if the line is an endif::[] directive.
func (p *Parser) isEndifDirective(line string) bool {
	trimmed := strings.TrimSpace(line)
	// Pattern: endif::attribute[]
	return strings.HasPrefix(trimmed, "endif::") && strings.HasSuffix(trimmed, "[]")
}

// handleEndifDirective processes an endif::[] directive.
func (p *Parser) handleEndifDirective(line string) {
	trimmed := strings.TrimSpace(line)
	// Extract the attribute name from endif::attribute[] or endif::[]
	if !strings.HasPrefix(trimmed, "endif::") || !strings.HasSuffix(trimmed, "[]") {
		return
	}

	between := trimmed[7 : len(trimmed)-2] // Strip "endif::" and "[]"
	target := strings.TrimSpace(between)

	// If target is empty, pop the most recent conditional
	if target == "" {
		if len(p.conditionalStack) > 0 {
			p.conditionalStack = p.conditionalStack[:len(p.conditionalStack)-1]
		}
		return
	}

	// Pop from stack until we find matching conditional
	for len(p.conditionalStack) > 0 {
		top := p.conditionalStack[len(p.conditionalStack)-1]
		if top.endifTarget == target {
			p.conditionalStack = p.conditionalStack[:len(p.conditionalStack)-1]
			return
		}
		// Mismatch - pop and continue (could be nested conditionals)
		p.conditionalStack = p.conditionalStack[:len(p.conditionalStack)-1]
	}
}

// isInConditionalSkip returns true if we're currently in an inactive conditional.
func (p *Parser) isInConditionalSkip() bool {
	if len(p.conditionalStack) == 0 {
		return false
	}
	// We're skipping if any conditional in the stack is inactive
	for _, cond := range p.conditionalStack {
		if !cond.active {
			return true
		}
	}
	return false
}

// handleConditionalDirective processes ifdef, ifndef, and ifeval directives.
func (p *Parser) handleConditionalDirective(classification *reader.Classification, doc *ast.NodeDocument) {
	cond := classification.Conditional
	if cond == nil {
		return
	}

	var active bool
	var endifTarget string

	switch cond.Type {
	case "ifdef":
		// Include content if attribute is defined and non-empty
		endifTarget = cond.Attribute
		value, exists := doc.Attributes[cond.Attribute]
		active = exists && value != ""

	case "ifndef":
		// Include content if attribute is NOT defined or is empty
		endifTarget = cond.Attribute
		value, exists := doc.Attributes[cond.Attribute]
		active = !exists || value == ""

	case "ifeval":
		// For now, treat ifeval as inactive (full expression evaluation to be implemented)
		// Expression format: "{attribute} == 'value'" or similar
		endifTarget = "" // Use empty target for ifeval (endif::[] closes it)
		active = p.evaluateIfeval(cond.Expression, doc)
	default:
		return
	}

	// Push the conditional state onto the stack
	p.conditionalStack = append(p.conditionalStack, conditionalState{
		active:      active,
		endifTarget: endifTarget,
		depth:       0,
	})
}

// evaluateIfeval evaluates an ifeval expression.
// Supported formats:
// - {attr} == "value"
// - {attr} != "value"
// - {attr} =~ "regex" (basic regex match)
// - "{attr}" == "value" (quoted attribute reference)
func (p *Parser) evaluateIfeval(expr string, doc *ast.NodeDocument) bool {
	// Simple expression parsing
	trimmed := strings.TrimSpace(expr)

	// Check for == operator
	if strings.Contains(trimmed, "==") {
		parts := strings.SplitN(trimmed, "==", 2)
		if len(parts) != 2 {
			return false
		}
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		// Extract attribute name from {attr} or "{attr}"
		attrName := p.extractAttrName(left)
		if attrName == "" {
			return false
		}

		// Extract value from "value" or 'value'
		rightVal := right
		if strings.HasPrefix(right, "\"") && strings.HasSuffix(right, "\"") {
			rightVal = right[1 : len(right)-1]
		} else if strings.HasPrefix(right, "'") && strings.HasSuffix(right, "'") {
			rightVal = right[1 : len(right)-1]
		}

		attrValue, exists := doc.Attributes[attrName]
		if !exists {
			return false
		}
		return attrValue == rightVal
	}

	// Check for != operator
	if strings.Contains(trimmed, "!=") {
		parts := strings.SplitN(trimmed, "!=", 2)
		if len(parts) != 2 {
			return false
		}
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		// Extract attribute name from {attr} or "{attr}"
		attrName := p.extractAttrName(left)
		if attrName == "" {
			return false
		}

		// Extract value from "value" or 'value'
		rightVal := right
		if strings.HasPrefix(right, "\"") && strings.HasSuffix(right, "\"") {
			rightVal = right[1 : len(right)-1]
		} else if strings.HasPrefix(right, "'") && strings.HasSuffix(right, "'") {
			rightVal = right[1 : len(right)-1]
		}

		attrValue, exists := doc.Attributes[attrName]
		if !exists {
			return true
		}
		return attrValue != rightVal
	}

	// Check for =~ operator (basic regex match)
	if strings.Contains(trimmed, "=~") {
		parts := strings.SplitN(trimmed, "=~", 2)
		if len(parts) != 2 {
			return false
		}
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		// Extract attribute name from {attr} or "{attr}"
		attrName := p.extractAttrName(left)
		if attrName == "" {
			return false
		}

		// Extract pattern from "pattern" or 'pattern'
		pattern := right
		if strings.HasPrefix(right, "\"") && strings.HasSuffix(right, "\"") {
			pattern = right[1 : len(right)-1]
		} else if strings.HasPrefix(right, "'") && strings.HasSuffix(right, "'") {
			pattern = right[1 : len(right)-1]
		}

		attrValue, exists := doc.Attributes[attrName]
		if !exists {
			return false
		}

		// Simple contains check for now (full regex support to be added)
		return strings.Contains(attrValue, pattern)
	}

	return false
}

// parseCaption extracts the caption text from a caption line (e.g., ".Title" -> "Title").
func (p *Parser) parseCaption(captionLine string) string {
	if captionLine == "" {
		return ""
	}
	trimmed := strings.TrimSpace(captionLine)
	if strings.HasPrefix(trimmed, ".") {
		return strings.TrimSpace(trimmed[1:])
	}
	return trimmed
}

// extractAttrName extracts the attribute name from expressions like {attr} or "{attr}"
func (p *Parser) extractAttrName(s string) string {
	s = strings.TrimSpace(s)

	// Try "{attr}" format first
	if strings.HasPrefix(s, "\"{") && strings.HasSuffix(s, "}\"") {
		return s[2 : len(s)-2]
	}

	// Try '{attr}' format
	if strings.HasPrefix(s, "'{") && strings.HasSuffix(s, "}'") {
		return s[2 : len(s)-2]
	}

	// Try {attr} format
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return s[1 : len(s)-1]
	}

	return ""
}

// runTreeProcessors executes all registered tree processors.
func (p *Parser) runTreeProcessors(doc *ast.NodeDocument) error {
	processors := p.extensionRegistry.GetTreeProcessors()

	// Sort by priority (lower first)
	sort.Slice(processors, func(i, j int) bool {
		return processors[i].Priority() < processors[j].Priority()
	})

	// Execute each processor
	for _, processor := range processors {
		if err := processor.Process(doc); err != nil {
			return fmt.Errorf("tree processor %T failed: %w", processor, err)
		}
	}

	return nil
}

// runBlockProcessors executes all registered block processors on document blocks.
func (p *Parser) runBlockProcessors(doc *ast.NodeDocument) error {
	processors := p.extensionRegistry.GetBlockProcessors()

	// Sort by priority (lower first)
	sort.Slice(processors, func(i, j int) bool {
		return processors[i].Priority() < processors[j].Priority()
	})

	// Process all blocks recursively
	p.processBlocksForExtensions(doc.Blocks, processors)

	return nil
}

// processBlocksForExtensions recursively processes blocks with block processors.
func (p *Parser) processBlocksForExtensions(blocks []ast.Node, processors []extension.BlockProcessor) {
	for i, block := range blocks {
		switch n := block.(type) {
		case *ast.NodeBlock:
			// Try each processor on this block
			for _, processor := range processors {
				if processor.Match(n) {
					result, err := processor.Process(n)
					if err == nil && result != nil {
						blocks[i] = result
					}
				}
			}
			// NodeBlock doesn't have nested blocks, so skip recursion

		case *ast.NodeSection:
			// Process blocks within section (Children field)
			for j, child := range n.Children {
				if childBlock, ok := child.(*ast.NodeBlock); ok {
					for _, processor := range processors {
						if processor.Match(childBlock) {
							result, err := processor.Process(childBlock)
							if err == nil && result != nil {
								n.Children[j] = result
							}
						}
					}
				}
			}
		}
	}
}
