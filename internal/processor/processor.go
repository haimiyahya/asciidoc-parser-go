// Package processor provides document processing functionality.
package processor

import (
	"regexp"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/inline"
)

// Processor processes an AsciiDoc document after parsing.
type Processor struct {
	// document is the AST being processed.
	document *ast.NodeDocument

	// attributes are document-level attributes.
	attributes map[string]string

	// predefinedAttributes are built-in AsciiDoc attributes.
	predefinedAttributes map[string]string

	// attributeRefPattern matches {attribute} references.
	attributeRefPattern *regexp.Regexp
}

// NewProcessor creates a new processor.
func NewProcessor(doc *ast.NodeDocument) *Processor {
	p := &Processor{
		document:            doc,
		attributes:          make(map[string]string),
		predefinedAttributes: initPredefinedAttributes(),
		attributeRefPattern: regexp.MustCompile(`\{([a-zA-Z0-9_-]+)\}`),
	}

	// Copy document attributes if they exist
	if doc.Attributes != nil {
		for k, v := range doc.Attributes {
			p.attributes[k] = v
		}
	}

	return p
}

// initPredefinedAttributes initializes built-in AsciiDoc attributes.
// Reference: https://docs.asciidoctor.org/asciidoc/latest/authors/builtin-attributes.html
func initPredefinedAttributes() map[string]string {
	return map[string]string{
		// Document attributes
		"toc":                      "left",
		"toclevels":                "3",
		"sectnums":                 "",
		"sectanchors":              "",
		"sectlinks":                "",
		"partnums":                 "",
		"chapter-label":            "Chapter",
		"appendix-caption":      "Appendix",
		"untitled-label":          "Untitled",
		"version-label":            "Version",
		"last-update-label":      "Last Updated",
		"example-caption":         "Example",
		"figure-caption":          "Figure",
		"table-caption":           "Table",
		"caution-caption":         "Caution",

		// Backend attributes
		"backend":                  "html5",
		"doctype":                  "",
		"imagesdir":               "images",

		// Formatting attributes
		"icons":                   "",
		"icontype":                "font",
		"iconfont-catalog":       "",
		"data-uri":                "",
		"linkattrs":               "",

		// Tab attributes
		"tabsize":                 "8",

		// Experimental
		"experimental":              "",
	}
}

// Process applies document processing (attributes, conditionals, etc.).
func (p *Processor) Process() error {
	// Process all blocks for attribute references
	p.processBlocks(p.document.Blocks)

	// Set default attributes if not already set
	p.setDefaults()

	return nil
}

// processBlocks processes blocks recursively for attribute references.
func (p *Processor) processBlocks(blocks []ast.Node) {
	for _, block := range blocks {
		switch n := block.(type) {
		case *ast.NodeParagraph:
			p.processParagraph(n)
		case *ast.NodeSection:
			p.processSection(n)
		case *ast.NodeList:
			p.processList(n)
		case *ast.NodeListItem:
			p.processListItem(n)
		case *ast.NodeLiteral:
			p.processLiteral(n)
		case *ast.NodeBlock:
			p.processBlock(n)
		case *ast.BibliographyEntryNode:
			p.processBibliographyEntry(n)
		}
	}
}

// processParagraph processes a paragraph for attribute references.
func (p *Processor) processParagraph(para *ast.NodeParagraph) {
	// Substitute attributes in the paragraph text and track position offsets
	newText, offsets := p.substituteAttributesWithOffsets(para.Text)

	// Update inline node positions based on text length changes
	p.adjustInlineNodePositions(para.InlineNodes, offsets)

	// Now do the actual substitution
	para.Text = newText

	// Substitute attributes in inline nodes
	p.processInlineNodes(para.InlineNodes)
}

// processSection processes a section for attribute references.
func (p *Processor) processSection(section *ast.NodeSection) {
	section.Title = p.substituteAttributes(section.Title)
}

// processList processes a list (no attributes in list wrapper typically).
func (p *Processor) processList(list *ast.NodeList) {
	// Process all list items
	for _, item := range list.Items {
		p.processBlocks([]ast.Node{item})
	}
}

// processListItem processes a list item for attribute references.
func (p *Processor) processListItem(item *ast.NodeListItem) {
	item.Text = p.substituteAttributes(item.Text)
	if item.Definition != "" {
		item.Definition = p.substituteAttributes(item.Definition)
	}
	if item.Term != "" {
		item.Term = p.substituteAttributes(item.Term)
	}

	// Substitute attributes in inline nodes
	p.processInlineNodes(item.InlineNodes)
	p.processInlineNodes(item.DefinitionNodes)
}

// processLiteral processes a literal block (verbatim, no substitution).
func (p *Processor) processLiteral(literal *ast.NodeLiteral) {
	// Literal blocks are verbatim - no substitution
}

// processBlock processes a generic block for attribute references.
func (p *Processor) processBlock(block *ast.NodeBlock) {
	for i, line := range block.Lines {
		block.Lines[i] = p.substituteAttributes(line)
	}
}

// processBibliographyEntry processes a bibliography entry for attribute references.
func (p *Processor) processBibliographyEntry(entry *ast.BibliographyEntryNode) {
	entry.Text = p.substituteAttributes(entry.Text)
	entry.XRefText = p.substituteAttributes(entry.XRefText)

	// Substitute attributes in inline nodes
	p.processInlineNodes(entry.InlineNodes)
}

// substituteAttributes replaces {attribute} references with their values.
func (p *Processor) substituteAttributes(text string) string {
	return p.attributeRefPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract attribute name (remove braces)
		attrName := match[1 : len(match)-1]
		attrName = strings.TrimSpace(attrName)

		// Check for special attribute syntax
		if strings.HasPrefix(attrName, "set:") {
			// {set:name!} - unset attribute
			name := strings.TrimPrefix(attrName, "set:")
			name = strings.TrimSuffix(name, "!")
			delete(p.attributes, name)
			if p.document.Attributes != nil {
				delete(p.document.Attributes, name)
			}
			return ""
		}

		if strings.HasSuffix(attrName, "!") {
			// {name!} - undefined check
			name := strings.TrimSuffix(attrName, "!")
			if _, exists := p.getAttribute(name); !exists {
				return "{undefined}"
			}
			return ""
		}

		// Regular attribute reference
		if val, exists := p.getAttribute(attrName); exists {
			return val
		}

		// Return original if attribute not found
		return ""
	})
}

// getAttribute retrieves an attribute value with precedence.
func (p *Processor) getAttribute(name string) (string, bool) {
	// Check document attributes first
	if val, ok := p.attributes[name]; ok {
		return val, true
	}

	// Check predefined attributes
	if val, ok := p.predefinedAttributes[name]; ok {
		return val, true
	}

	// Check environment variables (optional, for future)
	// Not implemented for now

	return "", false
}

// setDefaults sets default attributes if not already set.
func (p *Processor) setDefaults() {
	defaults := map[string]string{
		"toc":       "left",
		"toclevels": "3",
		"backend":    "html5",
		"icons":      "",
		"sectnums":  "",
	}

	for name, defaultValue := range defaults {
		if _, exists := p.attributes[name]; !exists {
			p.attributes[name] = defaultValue
			if p.document.Attributes == nil {
				p.document.Attributes = make(map[string]string)
			}
			p.document.Attributes[name] = defaultValue
		}
	}
}

// GetAttribute returns an attribute value (public method).
func (p *Processor) GetAttribute(name string) (string, bool) {
	return p.getAttribute(name)
}

// SetAttribute sets an attribute value (public method).
func (p *Processor) SetAttribute(name, value string) {
	p.attributes[name] = value
	if p.document.Attributes == nil {
		p.document.Attributes = make(map[string]string)
	}
	p.document.Attributes[name] = value
}

// GetAllAttributes returns all attributes (document + predefined).
func (p *Processor) GetAllAttributes() map[string]string {
	result := make(map[string]string)

	// Add predefined attributes
	for k, v := range p.predefinedAttributes {
		result[k] = v
	}

	// Document attributes override predefined
	for k, v := range p.attributes {
		result[k] = v
	}

	return result
}

// processInlineNodes processes inline nodes for attribute references.
// This handles attribute substitution in inline markup like links, images, etc.
func (p *Processor) processInlineNodes(nodes []interface{}) {
	for _, node := range nodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			p.processInlineNode(inlineNode)
		}
	}
}

// processInlineNode processes a single inline node for attribute references.
func (p *Processor) processInlineNode(node *inline.Node) {
	// Substitute attributes in Text field
	node.Text = p.substituteAttributes(node.Text)

	// Substitute attributes in URL field (for links)
	node.URL = p.substituteAttributes(node.URL)

	// Substitute attributes in Alt field (for images)
	node.Alt = p.substituteAttributes(node.Alt)

	// Substitute attributes in Ref and RefText fields (for cross-references)
	node.Ref = p.substituteAttributes(node.Ref)
	node.RefText = p.substituteAttributes(node.RefText)

	// Recursively process child nodes
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			p.processInlineNode(child)
		}
	}

	// Substitute attributes in macro-specific fields
	if node.MacroName != "" {
		node.MacroTarget = p.substituteAttributes(node.MacroTarget)
		if node.MacroAttrs != nil {
			for k, v := range node.MacroAttrs {
				node.MacroAttrs[k] = p.substituteAttributes(v)
			}
		}
	}
}

// positionOffset represents a change in text position at a specific index.
type positionOffset struct {
	index  int   // Position in the original text
	offset int   // Cumulative offset change at this position
}

// substituteAttributesWithOffsets substitutes attributes and returns position offsets.
// This is used to adjust inline node positions after text length changes.
func (p *Processor) substituteAttributesWithOffsets(text string) (string, []positionOffset) {
	var offsets []positionOffset
	currentOffset := 0

	result := p.attributeRefPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Get the position of this match in the original text
		matchStart := strings.Index(text[currentOffset:], match)
		if matchStart == -1 {
			matchStart = 0
		} else {
			matchStart += currentOffset
		}

		// Extract attribute name (remove braces)
		attrName := match[1 : len(match)-1]
		attrName = strings.TrimSpace(attrName)

		var replacement string
		var originalLenMatch = len(match)

		// Check for special attribute syntax
		if strings.HasPrefix(attrName, "set:") {
			// {set:name!} - unset attribute
			name := strings.TrimPrefix(attrName, "set:")
			name = strings.TrimSuffix(name, "!")
			delete(p.attributes, name)
			if p.document.Attributes != nil {
				delete(p.document.Attributes, name)
			}
			replacement = ""
		} else if strings.HasSuffix(attrName, "!") {
			// {name!} - undefined check
			name := strings.TrimSuffix(attrName, "!")
			if _, exists := p.getAttribute(name); !exists {
				replacement = "{undefined}"
			} else {
				replacement = ""
			}
		} else {
			// Regular attribute reference
			if val, exists := p.getAttribute(attrName); exists {
				replacement = val
			} else {
				replacement = ""
			}
		}

		// Calculate the length change
		lenChange := len(replacement) - originalLenMatch
		currentOffset += lenChange

		// Record the offset at this position
		if lenChange != 0 {
			offsets = append(offsets, positionOffset{
				index:  matchStart + originalLenMatch,
				offset: currentOffset,
			})
		}

		// Update currentOffset for next match search
		currentOffset = matchStart + len(replacement)

		return replacement
	})

	return result, offsets
}

// adjustInlineNodePositions adjusts inline node positions based on text length changes.
func (p *Processor) adjustInlineNodePositions(nodes []interface{}, offsets []positionOffset) {
	for _, node := range nodes {
		if inlineNode, ok := node.(*inline.Node); ok {
			p.adjustSingleNodePosition(inlineNode, offsets)
		}
	}
}

// adjustSingleNodePosition adjusts a single inline node's positions.
func (p *Processor) adjustSingleNodePosition(node *inline.Node, offsets []positionOffset) {
	// Find the applicable offset for this node's position
	getOffsetAt := func(pos int) int {
		cumulative := 0
		for _, off := range offsets {
			if pos >= off.index {
				cumulative = off.offset
			} else {
				break
			}
		}
		return cumulative
	}

	// Adjust StartPos and Position
	if node.StartPos > 0 {
		node.StartPos += getOffsetAt(node.StartPos)
	}
	if node.Position > 0 {
		node.Position += getOffsetAt(node.Position)
	}

	// Recursively adjust child nodes
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			p.adjustSingleNodePosition(child, offsets)
		}
	}
}
