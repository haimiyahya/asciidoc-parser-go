// Package inline provides inline parsing for AsciiDoc text content.
//
// This handles inline markup like bold, italic, links, monospace,
// superscript, subscript, and other phrasal elements.
//
// Reference: https://docs.asciidoctor.org/asciidoc/latest/text/bold-italic.html
package inline

import (
	"fmt"
	"strings"
)

// NodeType represents the type of an inline node.
type NodeType int

const (
	// NodeText is plain text content.
	NodeText NodeType = iota

	// NodeBold is bold text (**bold** or *bold* on single words).
	NodeBold

	// NodeItalic is italic text (__italic__ or _italic_ on single words).
	NodeItalic

	// NodeMonospace is monospace text (`text` or ++text++).
	NodeMonospace

	// NodeSubscript is subscript text (~text~).
	NodeSubscript

	// NodeSuperscript is superscript text (^text^).
	NodeSuperscript

	// NodeLink is a link (link:text[url] or bare URL).
	NodeLink

	// NodeImage is an inline image (image:url[alt-text]).
	NodeImage

	// NodeCrossRef is a cross-reference (<<section-id>>).
	NodeCrossRef
)

// String returns the string representation of NodeType.
func (nt NodeType) String() string {
	names := map[NodeType]string{
		NodeText:      "Text",
		NodeBold:      "Bold",
		NodeItalic:     "Italic",
		NodeMonospace:  "Monospace",
		NodeSubscript:  "Subscript",
		NodeSuperscript: "Superscript",
		NodeLink:       "Link",
		NodeImage:      "Image",
		NodeCrossRef:   "CrossRef",
	}
	if name, ok := names[nt]; ok {
		return name
	}
	return "Unknown"
}

// Node is an inline AST node.
type Node struct {
	// Type is the node type.
	Type NodeType

	// Text is the text content (for Text, Bold, Italic, etc.).
	Text string

	// URL is the link URL (for Link nodes).
	URL string

	// Alt is the alt text (for Image nodes).
	Alt string

	// Ref is the cross-reference target ID (for CrossRef nodes).
	Ref string

	// RefText is the custom reference text (for CrossRef with <<id,text>> syntax).
	RefText string

	// StartPos is the starting position of this node's markup in source text.
	StartPos int

	// Position is the ending position of this node's markup in source text.
	Position int

	// Children are child inline nodes (for complex inline structures).
	Children []*Node
}

// String returns a string representation of node.
func (n *Node) String() string {
	switch n.Type {
	case NodeText:
		return n.Text
	case NodeBold:
		return fmt.Sprintf("**%s**", n.Text)
	case NodeItalic:
		return fmt.Sprintf("__%s__", n.Text)
	case NodeMonospace:
		return fmt.Sprintf("`%s`", n.Text)
	case NodeSubscript:
		return fmt.Sprintf("~%s~", n.Text)
	case NodeSuperscript:
		return fmt.Sprintf("^%s^", n.Text)
	case NodeLink:
		if n.URL != "" {
			return fmt.Sprintf("%s[%s]", n.Text, n.URL)
		}
		return n.Text
	case NodeImage:
		return fmt.Sprintf("image:%s[%s]", n.URL, n.Alt)
	case NodeCrossRef:
		return fmt.Sprintf("<<%s>>", n.Ref)
	default:
		return n.Text
	}
}

// Parser parses inline AsciiDoc markup into AST nodes.
type Parser struct {
	// text is the input text to parse.
	text string
	// pos is the current parsing position.
	pos int
}

// NewParser creates a new inline parser.
func NewParser(text string) *Parser {
	return &Parser{
		text: text,
		pos:  0,
	}
}

// Parse parses the inline markup into a slice of nodes.
func (p *Parser) Parse() []*Node {
	var nodes []*Node

	for p.pos < len(p.text) {
		// Try to match inline constructs
		if node, newPos := p.tryCrossRef(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryImage(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryLink(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryBold(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryItalic(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryMonospace(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.trySubscript(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.trySuperscript(); node != nil {
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		// If no inline construct matched, look ahead to find where next inline construct starts
		// Consume text up to that point as a single text node
		nextInlinePos := len(p.text) // Default to consuming all remaining text

		// Look for next inline construct marker
		for i := p.pos; i < len(p.text); i++ {
			remaining := p.text[i:]
			if strings.HasPrefix(remaining, "<<") ||
				strings.HasPrefix(remaining, "image:") ||
				strings.HasPrefix(remaining, "link:") ||
				strings.HasPrefix(remaining, "**") ||
				strings.HasPrefix(remaining, "*") ||
				strings.HasPrefix(remaining, "__") ||
				strings.HasPrefix(remaining, "_") ||
				strings.HasPrefix(remaining, "`") ||
				strings.HasPrefix(remaining, "++") ||
				strings.HasPrefix(remaining, "http://") ||
				strings.HasPrefix(remaining, "https://") ||
				strings.HasPrefix(remaining, "~") ||
				strings.HasPrefix(remaining, "^") {
				nextInlinePos = i
				break
			}
		}

		if p.pos < nextInlinePos {
			nodes = append(nodes, &Node{Type: NodeText, Text: p.text[p.pos:nextInlinePos]})
			p.pos = nextInlinePos
		}
	}

	return nodes
}

// tryImage attempts to parse an inline image at current position.
// Supports: image:url[alt-text] where alt-text is optional.
func (p *Parser) tryImage() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for image macro: image:url[alt-text]
	if strings.HasPrefix(remaining, "image:") {
		// Find the closing ] or end of line
		closeBracket := strings.Index(remaining, "]")
		if closeBracket == -1 {
			// No alt-text bracket, use rest of line as URL
			url := remaining[6:]
			if url == "" {
				return nil, p.pos
			}
			// Trim any trailing whitespace
			url = strings.TrimRight(url, " \t\n")
			return &Node{
				Type: NodeImage,
				URL:  url,
				Alt:  "",
				Text: "",
				StartPos: p.pos,
				Position: p.pos + len(url) + 6,
			}, p.pos + len(url) + 6
		}

		// Extract image:url[alt-text]
		imageSpec := remaining[6:closeBracket]
		// Find the [ to split url from alt-text
		openBracket := strings.Index(imageSpec, "[")
		if openBracket == -1 {
			// No [ found, treat entire spec as URL without alt
			url := strings.TrimRight(imageSpec, " \t\n")
			return &Node{
				Type: NodeImage,
				URL:  url,
				Alt:  "",
				Text: "",
				StartPos: p.pos,
				Position: p.pos + closeBracket + 1,
			}, p.pos + closeBracket + 1
		}

		// URL is before [, alt-text is after it
		url := strings.TrimRight(imageSpec[:openBracket], " \t\n")
		alt := imageSpec[openBracket+1:]

		if url == "" {
			return nil, p.pos
		}

		return &Node{
			Type: NodeImage,
			URL:  url,
			Alt:  alt,
			Text: alt, // Use alt as display text
			StartPos: p.pos,
			Position: p.pos + closeBracket + 1,
		}, p.pos + closeBracket + 1
	}

	return nil, p.pos
}

// tryLink attempts to parse a link at the current position.
// Supports: link:text[url], bare URLs, and https:// URLs.
func (p *Parser) tryLink() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for macro link: link:text[url]
	if strings.HasPrefix(remaining, "link:") {
		// Find the closing ]
		closeBracket := strings.Index(remaining, "]")
		if closeBracket == -1 {
			return nil, p.pos
		}
		// Extract link:text[url]
		linkSpec := remaining[5:closeBracket]
		// Find the last [ to split text from url
		openBracket := strings.LastIndex(linkSpec, "[")
		if openBracket == -1 {
			return nil, p.pos
		}
		// text is before last [, url is after it
		text := linkSpec[:openBracket]
		url := linkSpec[openBracket+1 : len(linkSpec)]

		if text == "" || url == "" {
			return nil, p.pos
		}

		return &Node{
			Type: NodeLink,
			Text: text,
			URL:  url,
			StartPos: p.pos,
			Position: p.pos + closeBracket + 1,
		}, p.pos + closeBracket + 1
	}

	// Check for bare URL
	// Simple heuristic: starts with http:// or https://
	if strings.HasPrefix(remaining, "https://") || strings.HasPrefix(remaining, "http://") {
		// Find the end of the URL (stop at space or punctuation)
		// Note: periods are valid in URLs (domain names), but trailing periods should be excluded
		end := len(remaining)
		for i, c := range remaining {
			if c == ' ' || c == '\t' || c == '\n' ||
				c == ',' || c == '!' || c == '?' ||
				c == ')' || c == ']' || c == ';' {
				end = i
				break
			}
			// Handle trailing period: stop at period only if followed by space/end
			if c == '.' && (i+1 >= len(remaining) || remaining[i+1] == ' ') {
				end = i
				break
			}
		}
		url := remaining[:end]
		return &Node{
			Type: NodeLink,
			Text: url,
			URL:  url,
			StartPos: p.pos,
			Position: p.pos + end,
		}, p.pos + end
	}

	return nil, p.pos
}

// tryBold attempts to parse bold text: **bold** or *bold* (on single words).
func (p *Parser) tryBold() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for constrained bold: **bold**
	if strings.HasPrefix(remaining, "**") {
		closeIndex := strings.Index(remaining[2:], "**")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[2 : closeIndex+2]
			// Recursively parse inner content for nested inline markup
			children := NewParser(text).Parse()
			return &Node{
				Type:     NodeBold,
				Text:     text,
				Children:  children,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 4,
			}, p.pos + closeIndex + 4
		}
	}

	// Check for unconstrained bold: *bold* (single word only)
	if strings.HasPrefix(remaining, "*") && len(remaining) > 1 {
		// Must be followed by a non-space character
		if remaining[1] == ' ' {
			return nil, p.pos
		}
		closeIndex := strings.Index(remaining[1:], "*")
		if closeIndex != -1 && closeIndex > 0 {
			// Check if constrained by word boundaries
			text := remaining[1 : closeIndex+1]
			if p.isWord(text) {
				return &Node{
					Type:     NodeBold,
					Text:     text,
					StartPos:  p.pos,
					Position:  p.pos + closeIndex + 2,
				}, p.pos + closeIndex + 2
			}
		}
	}

	return nil, p.pos
}

// tryItalic attempts to parse italic text: __italic__ or _italic_ (on single words).
func (p *Parser) tryItalic() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for constrained italic: __italic__
	if strings.HasPrefix(remaining, "__") {
		closeIndex := strings.Index(remaining[2:], "__")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[2 : closeIndex+2]
			// Recursively parse inner content for nested inline markup
			children := NewParser(text).Parse()
			return &Node{
				Type:     NodeItalic,
				Text:     text,
				Children:  children,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 4,
			}, p.pos + closeIndex + 4
		}
	}

	// Check for unconstrained italic: _italic_ (single word only)
	if strings.HasPrefix(remaining, "_") && len(remaining) > 1 {
		// Must be followed by a non-space character
		if remaining[1] == ' ' {
			return nil, p.pos
		}
		closeIndex := strings.Index(remaining[1:], "_")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[1 : closeIndex+1]
			if p.isWord(text) {
				return &Node{
					Type:     NodeItalic,
					Text:     text,
					StartPos:  p.pos,
					Position:  p.pos + closeIndex + 2,
				}, p.pos + closeIndex + 2
			}
		}
	}

	return nil, p.pos
}

// tryMonospace attempts to parse monospace text: `text` or ++text++.
func (p *Parser) tryMonospace() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for backtick monospace: `text`
	if strings.HasPrefix(remaining, "`") {
		closeIndex := strings.Index(remaining[1:], "`")
		if closeIndex != -1 {
			text := remaining[1 : closeIndex+1]
			// Recursively parse inner content for nested inline markup
			children := NewParser(text).Parse()
			return &Node{
				Type:     NodeMonospace,
				Text:     text,
				Children:  children,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 2,
			}, p.pos + closeIndex + 2
		}
	}

	// Check for pluses monospace: ++text++
	if strings.HasPrefix(remaining, "++") {
		closeIndex := strings.Index(remaining[2:], "++")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[2 : closeIndex+2]
			return &Node{
				Type:     NodeMonospace,
				Text:     text,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 4,
			}, p.pos + closeIndex + 4
		}
	}

	return nil, p.pos
}

// trySubscript attempts to parse subscript text: ~text~.
func (p *Parser) trySubscript() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for tilde subscript: ~text~
	if strings.HasPrefix(remaining, "~") && len(remaining) > 1 {
		// Must be followed by a non-space character
		if remaining[1] == ' ' {
			return nil, p.pos
		}
		closeIndex := strings.Index(remaining[1:], "~")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[1 : closeIndex+1]
			return &Node{
				Type:     NodeSubscript,
				Text:     text,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 2,
			}, p.pos + closeIndex + 2
		}
	}

	return nil, p.pos
}

// trySuperscript attempts to parse superscript text: ^text^.
func (p *Parser) trySuperscript() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for caret superscript: ^text^
	if strings.HasPrefix(remaining, "^") && len(remaining) > 1 {
		// Must be followed by a non-space character
		if remaining[1] == ' ' {
			return nil, p.pos
		}
		closeIndex := strings.Index(remaining[1:], "^")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[1 : closeIndex+1]
			return &Node{
				Type:     NodeSuperscript,
				Text:     text,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 2,
			}, p.pos + closeIndex + 2
		}
	}

	return nil, p.pos
}

// tryCrossRef attempts to parse a cross-reference: <<section-id>> or <<section-id,link text>>.
func (p *Parser) tryCrossRef() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for cross-reference: <<section-id>> or <<section-id,link text>>
	if strings.HasPrefix(remaining, "<<") {
		// Find the closing >>
		closeIndex := strings.Index(remaining[2:], ">>")
		if closeIndex != -1 && closeIndex > 0 {
			refSpec := remaining[2 : closeIndex+2]
			// Check if there's a comma separating id from link text
			commaIndex := strings.Index(refSpec, ",")
			var refID, refText string
			if commaIndex != -1 {
				// <<id,text>> syntax
				refID = strings.TrimSpace(refSpec[:commaIndex])
				refText = strings.TrimSpace(refSpec[commaIndex+1:])
			} else {
				// <<id>> syntax - use id as display text
				refID = strings.TrimSpace(refSpec)
				refText = refID
			}

			if refID != "" {
				return &Node{
					Type:     NodeCrossRef,
					Ref:      refID,
					Text:     refText,
					RefText:  refText,
					StartPos: p.pos,
					Position: p.pos + closeIndex + 4,
				}, p.pos + closeIndex + 4
			}
		}
	}

	return nil, p.pos
}

// isWord returns true if the text appears to be a single word.
// A word contains no spaces and is not empty.
func (p *Parser) isWord(s string) bool {
	if s == "" {
		return false
	}
	return !strings.Contains(s, " ")
}
