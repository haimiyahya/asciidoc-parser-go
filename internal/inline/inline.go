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

	// NodeMonospace is monospace text (`text` with backticks).
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

	// NodeKbd is a keyboard key combination (kbd:[Ctrl+C]).
	NodeKbd

	// NodeBtn is a button label (btn:[OK]).
	NodeBtn

	// NodeMenu is a menu path (menu:[File > Save]).
	NodeMenu

	// NodeIndexTerm is an index term entry.
	NodeIndexTerm
)

// String returns the string representation of NodeType.
func (nt NodeType) String() string {
	names := map[NodeType]string{
		NodeText:       "Text",
		NodeBold:       "Bold",
		NodeItalic:     "Italic",
		NodeMonospace:  "Monospace",
		NodeSubscript:  "Subscript",
		NodeSuperscript: "Superscript",
		NodeLink:       "Link",
		NodeImage:      "Image",
		NodeCrossRef:   "CrossRef",
		NodeKbd:        "Kbd",
		NodeBtn:        "Btn",
		NodeMenu:       "Menu",
		NodeIndexTerm:  "IndexTerm",
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

	// Roles are CSS role classes applied to this node (from [.role] syntax).
	Roles []string

	// Attributes are additional attributes (for links with window="_blank", etc.)
	Attributes map[string]string

	// StartPos is the starting position of this node's markup in source text.
	StartPos int

	// Position is the ending position of this node's markup in source text.
	Position int

	// Children are child inline nodes (for complex inline structures).
	Children []*Node

	// Index term fields (for NodeIndexTerm)
	IndexTermPrimary   string // Primary index term
	IndexTermSecondary string // Secondary index term (optional)
	IndexTermTertiary  string // Tertiary index term (optional)
	IndexTermConcealed  bool   // True for concealed index terms (((...))) - hidden from text
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
	case NodeIndexTerm:
		if n.IndexTermConcealed {
			// Concealed index term: (((primary, secondary, tertiary)))
			terms := n.IndexTermPrimary
			if n.IndexTermSecondary != "" {
				terms += ", " + n.IndexTermSecondary
			}
			if n.IndexTermTertiary != "" {
				terms += ", " + n.IndexTermTertiary
			}
			return fmt.Sprintf("((%s))", terms)
		}
		// Flow index term: ((primary))
		return fmt.Sprintf("(%s)", n.IndexTermPrimary)
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
	var pendingRoles []string
	skipRole := false // Skip role parsing if we just emitted unattached roles

	for p.pos < len(p.text) {
		// Check for role syntax: [.role] or [.role1.role2]
		// Only try to parse role if we didn't just emit unattached roles
		if !skipRole && len(pendingRoles) == 0 {
			if roles, endPos := p.tryRole(); roles != nil {
				pendingRoles = roles
				p.pos = endPos
				continue
			}
		}
		skipRole = false // Reset the skip flag

		// Try to match inline constructs
		// Check for index terms first (before other constructs that might use parentheses)
		if node, newPos := p.tryIndexTerm(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryKbd(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryBtn(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryMenu(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryCrossRef(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryImage(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryLink(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryBold(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryItalic(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.trySpan(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.tryMonospace(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.trySubscript(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		if node, newPos := p.trySuperscript(); node != nil {
			p.applyRoles(node, pendingRoles)
			pendingRoles = nil
			nodes = append(nodes, node)
			p.pos = newPos
			continue
		}

		// If we have pending roles but no inline element matched, emit the role as plain text
		if len(pendingRoles) > 0 {
			// Roles were not consumed - emit as text and consume at least one character
			// Emit the role markup as plain text
			roleText := p.text[p.pos:min(p.pos+1, len(p.text))]
			nodes = append(nodes, &Node{Type: NodeText, Text: roleText})
			p.pos = min(p.pos+1, len(p.text))
			pendingRoles = nil
			skipRole = true // Skip role parsing in next iteration to avoid infinite loop
			continue
		}

		// If no inline construct matched, look ahead to find where next inline construct starts
		// Consume text up to that point as a single text node
		nextInlinePos := len(p.text) // Default to consuming all remaining text

		// Look for next inline construct marker
		for i := p.pos; i < len(p.text); i++ {
			remaining := p.text[i:]
			if strings.HasPrefix(remaining, "[.") ||
				strings.HasPrefix(remaining, "<<") ||
				strings.HasPrefix(remaining, "image:") ||
				strings.HasPrefix(remaining, "link:") ||
				strings.HasPrefix(remaining, "kbd:") ||
				strings.HasPrefix(remaining, "btn:") ||
				strings.HasPrefix(remaining, "menu:") ||
				strings.HasPrefix(remaining, "**") ||
				strings.HasPrefix(remaining, "*") ||
				strings.HasPrefix(remaining, "__") ||
				strings.HasPrefix(remaining, "_") ||
				strings.HasPrefix(remaining, "`") ||
				strings.HasPrefix(remaining, "++") ||
				strings.HasPrefix(remaining, "http://") ||
				strings.HasPrefix(remaining, "https://") ||
				strings.HasPrefix(remaining, "~") ||
				strings.HasPrefix(remaining, "^") ||
				strings.HasPrefix(remaining, "(((") || // Concealed index term
				strings.HasPrefix(remaining, "((") { // Flow index term
				nextInlinePos = i
				break
			}
		}

		if p.pos < nextInlinePos {
			nodes = append(nodes, &Node{Type: NodeText, Text: p.text[p.pos:nextInlinePos]})
			p.pos = nextInlinePos
		} else {
			// No more inline constructs found, consume remaining text and exit
			if p.pos < len(p.text) {
				nodes = append(nodes, &Node{Type: NodeText, Text: p.text[p.pos:]})
				p.pos = len(p.text)
			}
		}
	}

	return nodes
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
// Supports: link:url[text], bare URLs, and https:// URLs.
func (p *Parser) tryLink() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for macro link: link:url[text, attrs...]
	if strings.HasPrefix(remaining, "link:") {
		// Find the closing ]
		closeBracket := strings.Index(remaining, "]")
		if closeBracket == -1 {
			return nil, p.pos
		}
		// Extract link:url[text, attrs...]
		linkSpec := remaining[5:closeBracket]
		// Find the [ to split url from text
		openBracket := strings.Index(linkSpec, "[")
		if openBracket == -1 {
			return nil, p.pos
		}
		// url is before [, text+attrs are after it
		url := linkSpec[:openBracket]
		textAndAttrs := linkSpec[openBracket+1 : len(linkSpec)]

		// Parse text and attributes
		// Format: text, attr1=value1, attr2=value2, ...
		text := textAndAttrs
		attrs := make(map[string]string)

		// Check if there are attributes (comma followed by space and attribute)
		commaIdx := strings.Index(textAndAttrs, ", ")
		if commaIdx != -1 {
			text = textAndAttrs[:commaIdx]
			attrStr := textAndAttrs[commaIdx+2:]

			// Parse attributes: window="_blank", rel="noopener"
			// Simple regex-based parsing
			attrParts := strings.Split(attrStr, ", ")
			for _, attr := range attrParts {
				// Split on "=" to get key-value
				if eqIdx := strings.Index(attr, "="); eqIdx != -1 {
					key := strings.TrimSpace(attr[:eqIdx])
					// Remove quotes from value
					val := attr[eqIdx+1:]
					if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') {
						val = val[1 : len(val)-1]
					}
					attrs[key] = val
				}
			}
		}

		if url == "" {
			return nil, p.pos
		}

		return &Node{
			Type:       NodeLink,
			Text:       text,
			URL:        url,
			Attributes: attrs,
			StartPos:   p.pos,
			Position:   p.pos + closeBracket + 1,
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
			Type:  NodeLink,
			Text:  url,
			URL:   url,
			Roles: []string{"bare"}, // Add "bare" role for Asciidoctor compatibility
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

	// Check for unconstrained bold: *bold* (single word or phrase)
	if strings.HasPrefix(remaining, "*") && len(remaining) > 1 {
		// Must be followed by a non-space character
		if remaining[1] == ' ' {
			return nil, p.pos
		}
		closeIndex := strings.Index(remaining[1:], "*")
		if closeIndex != -1 && closeIndex > 0 {
			// Asciidoctor compatibility: allow multi-word unconstrained bold
			text := remaining[1 : closeIndex+1]
			return &Node{
				Type:     NodeBold,
				Text:     text,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 2,
			}, p.pos + closeIndex + 2
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

	// Check for unconstrained italic: _italic_ (single word or phrase)
	if strings.HasPrefix(remaining, "_") && len(remaining) > 1 {
		// Must be followed by a non-space character
		if remaining[1] == ' ' {
			return nil, p.pos
		}
		closeIndex := strings.Index(remaining[1:], "_")
		if closeIndex != -1 && closeIndex > 0 {
			// Asciidoctor compatibility: allow multi-word unconstrained italic
			text := remaining[1 : closeIndex+1]
			return &Node{
				Type:     NodeItalic,
				Text:     text,
				StartPos:  p.pos,
				Position:  p.pos + closeIndex + 2,
			}, p.pos + closeIndex + 2
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

	// Note: ++text++ is NOT monospace in Asciidoctor by default
	// It's a span that just groups text. Only backticks produce monospace.

	return nil, p.pos
}

// trySpan attempts to parse a span: ++text++.
// This just removes the delimiters and outputs the text as-is.
func (p *Parser) trySpan() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for double plus span: ++text++
	if strings.HasPrefix(remaining, "++") {
		closeIndex := strings.Index(remaining[2:], "++")
		if closeIndex != -1 && closeIndex > 0 {
			text := remaining[2 : closeIndex+2]
			// Return as text node (no special formatting)
			return &Node{
				Type:     NodeText,
				Text:     text,
				StartPos: p.pos,
				Position: p.pos + closeIndex + 4,
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

// tryRole attempts to parse a role attribute: [.role] or [.role1.role2].
// Returns the list of role names and the new position, or nil if not a role.
func (p *Parser) tryRole() ([]string, int) {
	remaining := p.text[p.pos:]

	// Check for role syntax: [.role] or [.role1.role2]
	if strings.HasPrefix(remaining, "[.") {
		// Find the closing ]
		closeIndex := strings.Index(remaining, "]")
		if closeIndex == -1 {
			return nil, p.pos
		}

		// Extract the role specification (without [. and ])
		roleSpec := remaining[2:closeIndex]
		if roleSpec == "" {
			return nil, p.pos
		}

		// Split by dots to get individual roles
		roles := strings.Split(roleSpec, ".")

		// Filter out empty strings
		var validRoles []string
		for _, role := range roles {
			if role != "" {
				validRoles = append(validRoles, role)
			}
		}

		if len(validRoles) == 0 {
			return nil, p.pos
		}

		return validRoles, p.pos + closeIndex + 1
	}

	return nil, p.pos
}

// applyRoles applies pending roles to a node.
func (p *Parser) applyRoles(node *Node, roles []string) {
	if len(roles) > 0 {
		node.Roles = append(node.Roles, roles...)
	}
}

// tryKbd attempts to parse a keyboard macro: kbd:[keys] or kbd:[keys+keys].
// Supports key combinations like Ctrl+C, Alt+Shift+Del.
func (p *Parser) tryKbd() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for kbd macro: kbd:[keys]
	if strings.HasPrefix(remaining, "kbd:[") {
		// Find the closing ]
		closeIndex := strings.Index(remaining[4:], "]")
		if closeIndex == -1 {
			return nil, p.pos
		}

		// Extract the keys
		keys := remaining[5 : closeIndex+4]
		if keys == "" {
			return nil, p.pos
		}

		return &Node{
			Type:     NodeKbd,
			Text:     keys,
			StartPos: p.pos,
			Position: p.pos + closeIndex + 5,
		}, p.pos + closeIndex + 5
	}

	return nil, p.pos
}

// tryBtn attempts to parse a button macro: btn:[label].
func (p *Parser) tryBtn() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for btn macro: btn:[label]
	if strings.HasPrefix(remaining, "btn:[") {
		// Find the closing ]
		closeIndex := strings.Index(remaining[4:], "]")
		if closeIndex == -1 {
			return nil, p.pos
		}

		// Extract the label
		label := remaining[5 : closeIndex+4]
		if label == "" {
			return nil, p.pos
		}

		return &Node{
			Type:     NodeBtn,
			Text:     label,
			StartPos: p.pos,
			Position: p.pos + closeIndex + 5,
		}, p.pos + closeIndex + 5
	}

	return nil, p.pos
}

// tryMenu attempts to parse a menu macro: menu:[File > Save] or menu:[File,Save].
// Supports menu paths with > or , as separator.
func (p *Parser) tryMenu() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for menu macro: menu:[path]
	if strings.HasPrefix(remaining, "menu:[") {
		// Find the closing ]
		closeIndex := strings.Index(remaining[5:], "]")
		if closeIndex == -1 {
			return nil, p.pos
		}

		// Extract the menu path
		path := remaining[6 : closeIndex+5]
		if path == "" {
			return nil, p.pos
		}

		return &Node{
			Type:     NodeMenu,
			Text:     path,
			StartPos: p.pos,
			Position: p.pos + closeIndex + 6,
		}, p.pos + closeIndex + 6
	}

	return nil, p.pos
}

// tryIndexTerm attempts to parse an index term: ((term)) or (((term, secondary, tertiary))).
// Flow index terms: ((term)) - visible in text
// Concealed index terms: (((term, secondary, tertiary))) - hidden from text
func (p *Parser) tryIndexTerm() (*Node, int) {
	remaining := p.text[p.pos:]

	// Check for concealed index term: (((term)))
	if strings.HasPrefix(remaining, "(((") {
		closeIndex := strings.Index(remaining[3:], ")))")
		if closeIndex != -1 {
			// Extract content between ((( and )))
			content := remaining[3 : closeIndex+3]

			// Parse comma-separated terms, handling quoted segments
			terms := p.parseIndexTerms(content)
			if len(terms) > 0 && terms[0] != "" {
				node := &Node{
					Type:              NodeIndexTerm,
					IndexTermPrimary:   terms[0],
					IndexTermConcealed: true,
					StartPos:           p.pos,
					Position:           p.pos + closeIndex + 6,
				}
				if len(terms) > 1 {
					node.IndexTermSecondary = terms[1]
				}
				if len(terms) > 2 {
					node.IndexTermTertiary = terms[2]
				}
				return node, p.pos + closeIndex + 6
			}
		}
	}

	// Check for flow index term: ((term))
	if strings.HasPrefix(remaining, "((") {
		closeIndex := strings.Index(remaining[2:], "))")
		if closeIndex != -1 {
			// Extract content between (( and ))
			content := remaining[2 : closeIndex+2]
			content = strings.TrimSpace(content)

			if content != "" {
				return &Node{
					Type:              NodeIndexTerm,
					IndexTermPrimary:   content,
					IndexTermConcealed: false,
					Text:               content, // Flow terms show the text
					StartPos:           p.pos,
					Position:           p.pos + closeIndex + 4,
				}, p.pos + closeIndex + 4
			}
		}
	}

	return nil, p.pos
}

// parseIndexTerms parses comma-separated index terms, handling quoted segments.
// For example: "knight, \"Arthur, King\"" returns ["knight", "Arthur, King"]
func (p *Parser) parseIndexTerms(s string) []string {
	var terms []string
	var currentTerm strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == '"' && (i == 0 || s[i-1] != '\\') {
			inQuotes = !inQuotes
			continue
		}

		if c == ',' && !inQuotes {
			terms = append(terms, strings.TrimSpace(currentTerm.String()))
			currentTerm.Reset()
			continue
		}

		currentTerm.WriteByte(c)
	}

	// Add the last term
	if currentTerm.Len() > 0 {
		terms = append(terms, strings.TrimSpace(currentTerm.String()))
	}

	return terms
}
