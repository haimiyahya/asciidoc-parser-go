// Package lsp provides Language Server Protocol implementation for AsciiDoc.
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/haimiyahya/asciidoc-parser-go/internal/reader"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// Server implements the LSP server for AsciiDoc.
type Server struct {
	// conn is the JSON-RPC connection to the client.
	conn *jsonrpc2.Conn

	// documents tracks open documents and their contents.
	documents map[string]*Document
	docMu      sync.RWMutex

	// parser is the AsciiDoc parser instance.
	parser *parser.Parser

	// capabilities of this server.
	capabilities *lsp.ServerCapabilities

	// clientCapabilities are the capabilities reported by the client.
	clientCapabilities *lsp.ClientCapabilities
}

// Document represents an open document in the editor.
type Document struct {
	// URI is the document URI.
	URI string

	// LanguageID is the language identifier (e.g., "asciidoc").
	LanguageID string

	// Version is the document version.
	Version int

	// Content is the full document content.
	Content string

	// AST is the parsed AST (nil if not yet parsed).
	AST *ast.NodeDocument

	// Diagnostics are the current diagnostics for this document.
	Diagnostics []lsp.Diagnostic

	// Symbols are the document symbols.
	Symbols []lsp.SymbolInformation
}

// NewServer creates a new LSP server.
func NewServer() *Server {
	return &Server{
		documents: make(map[string]*Document),
		capabilities: &lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
				},
			},
			DocumentSymbolProvider:       true,
			DocumentFormattingProvider:  true,
			CompletionProvider: &lsp.CompletionOptions{
				ResolveProvider:   false,
				TriggerCharacters: []string{"[", ":", "{", "=", "*", "_", "^", "~", "`", "#"},
			},
			HoverProvider:              true,
			DefinitionProvider:          true,
			CodeActionProvider:         true,
			ReferencesProvider:         true,
			RenameProvider:              true,
			DocumentHighlightProvider:  true,
		},
	}
}

// DidOpen handles the textDocument/didOpen notification.
func (s *Server) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) error {
	uri := string(params.TextDocument.URI)
	content := params.TextDocument.Text

	doc := &Document{
		URI:        uri,
		LanguageID: params.TextDocument.LanguageID,
		Version:    int(params.TextDocument.Version),
		Content:    content,
	}

	s.docMu.Lock()
	s.documents[uri] = doc
	s.docMu.Unlock()

	// Parse the document and compute diagnostics
	s.parseDocument(uri)
	s.publishDiagnostics(ctx, uri)

	return nil
}

// DidChange handles the textDocument/didChange notification.
func (s *Server) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) error {
	uri := string(params.TextDocument.URI)

	s.docMu.Lock()
	defer s.docMu.Unlock()

	doc, ok := s.documents[uri]
	if !ok {
		return nil
	}

	// Handle full document sync
	for _, change := range params.ContentChanges {
		doc.Content = change.Text
	}
	doc.Version = int(params.TextDocument.Version)

	// Re-parse and update diagnostics
	s.parseDocument(uri)
	s.publishDiagnostics(ctx, uri)

	return nil
}

// DidClose handles the textDocument/didClose notification.
func (s *Server) DidClose(ctx context.Context, params *lsp.DidCloseTextDocumentParams) error {
	uri := string(params.TextDocument.URI)

	s.docMu.Lock()
	delete(s.documents, uri)
	s.docMu.Unlock()

	// Clear diagnostics
	if s.conn != nil {
		s.conn.Notify(ctx, "textDocument/publishDiagnostics", &lsp.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: []lsp.Diagnostic{},
		})
	}

	return nil
}

// parseDocument parses the document and updates its AST and diagnostics.
func (s *Server) parseDocument(uri string) {
	s.docMu.Lock()
	doc := s.documents[uri]
	s.docMu.Unlock()

	if doc == nil {
		return
	}

	// Parse the document
	reader, err := reader.NewReader(doc.Content)
	if err != nil {
		doc.Diagnostics = []lsp.Diagnostic{
			{
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: 0},
					End:   lsp.Position{Line: 0, Character: 0},
				},
				Message:  fmt.Sprintf("Failed to create reader: %v", err),
				Severity: lsp.Error,
			},
		}
		doc.AST = nil
		return
	}

	p := parser.NewParser(reader)
	astDoc, err := p.Parse()
	if err != nil {
		doc.Diagnostics = []lsp.Diagnostic{
			{
				Range: lsp.Range{
					Start: lsp.Position{Line: 0, Character: 0},
					End:   lsp.Position{Line: 0, Character: 0},
				},
				Message:  fmt.Sprintf("Parse error: %v", err),
				Severity: lsp.Error,
			},
		}
		doc.AST = nil
		return
	}

	doc.AST = astDoc
	doc.Diagnostics = nil // Clear errors on successful parse

	// Extract symbols
	doc.Symbols = s.extractSymbols(astDoc, uri)
}

// publishDiagnostics sends diagnostics to the client.
func (s *Server) publishDiagnostics(ctx context.Context, uri string) {
	s.docMu.RLock()
	doc := s.documents[uri]
	diagnostics := doc.Diagnostics
	s.docMu.RUnlock()

	if s.conn != nil {
		s.conn.Notify(ctx, "textDocument/publishDiagnostics", &lsp.PublishDiagnosticsParams{
			URI:         lsp.DocumentURI(uri),
			Diagnostics: diagnostics,
		})
	}
}

// extractSymbols extracts document symbols from the AST.
func (s *Server) extractSymbols(document *ast.NodeDocument, uri string) []lsp.SymbolInformation {
	var symbols []lsp.SymbolInformation

	for _, block := range document.Blocks {
		if section, ok := block.(*ast.NodeSection); ok {
			symbols = append(symbols, s.extractSectionSymbol(section, uri)...)
		}
	}

	return symbols
}

// extractSectionSymbol extracts a symbol from a section.
func (s *Server) extractSectionSymbol(section *ast.NodeSection, uri string) []lsp.SymbolInformation {
	symbols := []lsp.SymbolInformation{
		{
			Name:     section.Title,
			Kind:     lsp.SKArray, // Using Array as placeholder for section
			Location: lsp.Location{
				URI:   lsp.DocumentURI(uri),
				Range: s.sectionRange(section),
			},
			ContainerName: "",
		},
	}

	// Extract child sections
	for _, child := range section.Children {
		if childSection, ok := child.(*ast.NodeSection); ok {
			symbols = append(symbols, s.extractSectionSymbol(childSection, uri)...)
		}
	}

	return symbols
}

// sectionRange computes the LSP range for a section.
func (s *Server) sectionRange(section *ast.NodeSection) lsp.Range {
	line := section.Pos.Line
	if line <= 0 {
		line = 1
	}
	return lsp.Range{
		Start: lsp.Position{Line: line - 1, Character: 0},
		End:   lsp.Position{Line: line - 1, Character: len(section.Title)},
	}
}

// DocumentSymbol handles the textDocument/documentSymbol request.
func (s *Server) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) ([]lsp.SymbolInformation, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil {
		return nil, nil
	}

	return doc.Symbols, nil
}

// Completion handles the textDocument/completion request.
func (s *Server) Completion(ctx context.Context, params *lsp.CompletionParams) (*lsp.CompletionList, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil {
		return &lsp.CompletionList{
			IsIncomplete: false,
			Items:        []lsp.CompletionItem{},
		}, nil
	}

	// Get line content to determine context
	line := s.getLine(doc, int(params.Position.Line))
	char := int(params.Position.Character)

	items := s.completionForLine(line, char, params.Position)

	return &lsp.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// getLine gets a specific line from the document.
func (s *Server) getLine(doc *Document, lineNum int) string {
	lines := splitLines(doc.Content)
	if lineNum >= 0 && lineNum < len(lines) {
		return lines[lineNum]
	}
	return ""
}

// splitLines splits content into lines.
func splitLines(content string) []string {
	lines := []string{}
	current := ""
	for _, ch := range content {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else if ch != '\r' {
			current += string(ch)
		}
	}
	lines = append(lines, current)
	return lines
}

// completionForLine returns completion items based on line context.
func (s *Server) completionForLine(line string, char int, pos lsp.Position) []lsp.CompletionItem {
	var items []lsp.CompletionItem

	// Check for inline markup triggers
	if char > 0 {
		prefix := line[:char]

		// Bold completion
		if strings.HasSuffix(prefix, "*") {
			items = append(items, lsp.CompletionItem{
				Label:            "**bold**",
				Kind:             lsp.CIKFunction,
				Detail:           "Bold text",
				InsertTextFormat: lsp.ITFPlainText,
			})
		}

		// Italic completion
		if strings.HasSuffix(prefix, "_") {
			items = append(items, lsp.CompletionItem{
				Label:            "__italic__",
				Kind:             lsp.CIKFunction,
				Detail:           "Italic text",
				InsertTextFormat: lsp.ITFPlainText,
			})
		}

		// Monospace completion
		if strings.HasSuffix(prefix, "`") {
			items = append(items, lsp.CompletionItem{
				Label:            "`code`",
				Kind:             lsp.CIKFunction,
				Detail:           "Monospace text",
				InsertTextFormat: lsp.ITFPlainText,
			})
		}

		// Link completion
		if strings.HasSuffix(prefix, "link:") {
			items = append(items, lsp.CompletionItem{
				Label:            "link:text[url]",
				Kind:             lsp.CIKFunction,
				Detail:           "Link",
				InsertTextFormat: lsp.ITFPlainText,
			})
		}

		// Image completion
		if strings.HasSuffix(prefix, "image:") {
			items = append(items, lsp.CompletionItem{
				Label:            "image:path[alt]",
				Kind:             lsp.CIKFunction,
				Detail:           "Image",
				InsertTextFormat: lsp.ITFPlainText,
			})
		}
	}

	// Section/completion for =
	if strings.HasPrefix(line, "=") || strings.Contains(line, "==") {
		items = append(items, s.sectionCompletions()...)
	}

	// List completions
	if strings.HasPrefix(line, "*") {
		items = append(items, lsp.CompletionItem{
			Label:            "* Unordered list item",
			Kind:             lsp.CIKSnippet,
			Detail:           "Unordered list",
			InsertTextFormat: lsp.ITFPlainText,
		})
	}

	if strings.HasPrefix(line, ".") {
		items = append(items, lsp.CompletionItem{
			Label:            ". Ordered list item",
			Kind:             lsp.CIKSnippet,
			Detail:           "Ordered list",
			InsertTextFormat: lsp.ITFPlainText,
		})
	}

	// Block completions
	if strings.HasPrefix(line, "----") {
		items = append(items, lsp.CompletionItem{
			Label:            "----\nLiteral block\n----",
			Kind:             lsp.CIKSnippet,
			Detail:           "Literal block",
			InsertTextFormat: lsp.ITFPlainText,
		})
	}

	if strings.HasPrefix(line, "====") {
		items = append(items, lsp.CompletionItem{
			Label:            "====\nExample block\n====",
			Kind:             lsp.CIKSnippet,
			Detail:           "Example block",
			InsertTextFormat: lsp.ITFPlainText,
		})
	}

	// Admonition completions
	if strings.HasPrefix(strings.ToLower(line), "note:") {
		items = append(items, lsp.CompletionItem{
			Label:            "NOTE:",
			Kind:             lsp.CIKSnippet,
			Detail:           "Note admonition",
			InsertTextFormat: lsp.ITFPlainText,
		})
	}

	// Attribute reference completion
	if char > 0 && line[char-1] == '{' {
		items = append(items, s.attributeCompletions()...)
	}

	return items
}

// sectionCompletions returns section-level completions.
func (s *Server) sectionCompletions() []lsp.CompletionItem {
	return []lsp.CompletionItem{
		{Label: "= ", Kind: lsp.CIKSnippet, Detail: "Level 1 section", InsertTextFormat: lsp.ITFPlainText},
		{Label: "== ", Kind: lsp.CIKSnippet, Detail: "Level 2 section", InsertTextFormat: lsp.ITFPlainText},
		{Label: "=== ", Kind: lsp.CIKSnippet, Detail: "Level 3 section", InsertTextFormat: lsp.ITFPlainText},
		{Label: "==== ", Kind: lsp.CIKSnippet, Detail: "Level 4 section", InsertTextFormat: lsp.ITFPlainText},
		{Label: "===== ", Kind: lsp.CIKSnippet, Detail: "Level 5 section", InsertTextFormat: lsp.ITFPlainText},
		{Label: "====== ", Kind: lsp.CIKSnippet, Detail: "Level 6 section", InsertTextFormat: lsp.ITFPlainText},
	}
}

// attributeCompletions returns attribute reference completions.
func (s *Server) attributeCompletions() []lsp.CompletionItem {
	attrs := []string{
		"toc", "toclevels", "sectnums", "sectanchors", "sectlinks",
		"imagesdir", "icons", "data-uri", "linkattrs",
		"backend", "doctype", "author", "email",
	}

	items := make([]lsp.CompletionItem, 0, len(attrs))
	for _, attr := range attrs {
		items = append(items, lsp.CompletionItem{
			Label:            attr,
			Kind:             lsp.CIKVariable,
			Detail:           "Document attribute",
			InsertText:       attr + "}",
			InsertTextFormat: lsp.ITFPlainText,
		})
	}
	return items
}

// Hover handles the textDocument/hover request.
func (s *Server) Hover(ctx context.Context, params *lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil || doc.AST == nil {
		return nil, nil
	}

	line := s.getLine(doc, params.Position.Line)
	char := params.Position.Character

	// Provide hover info for various constructs
	content := s.hoverForLine(line, char, doc)
	if content == "" {
		return nil, nil
	}

	return &lsp.Hover{
		Contents: []lsp.MarkedString{lsp.RawMarkedString(content)},
	}, nil
}

// hoverForLine returns hover content based on line context.
func (s *Server) hoverForLine(line string, char int, doc *Document) string {
	// Check for inline markup
	if char > 0 {
		prefix := line[:char]

		if strings.Contains(prefix, "**") || strings.Contains(prefix, "__") {
			return "**Bold/Italic**: Inline formatting makes text **bold** or __italic__."
		}
		if strings.Contains(prefix, "`") {
			return "**Monospace**: `code` renders text in a monospace font."
		}
		if strings.Contains(prefix, "link:") {
			return "**Link**: `link:text[url]` creates a hyperlink."
		}
		if strings.Contains(prefix, "image:") {
			return "**Image**: `image:path[alt]` embeds an image."
		}
		if strings.Contains(prefix, "<<") {
			return "**Cross-Reference**: `<<section-id>>` creates a link to another section."
		}
		if strings.Contains(prefix, "{") && strings.Contains(line[char:], "}") {
			return "**Attribute Reference**: `{attrname}` substitutes the value of an attribute."
		}
	}

	// Section info
	if strings.HasPrefix(line, "=") {
		level := strings.Count(strings.TrimLeft(line, "="), "=")
		return fmt.Sprintf("**Section Level %d**: Lines starting with `%s` create section headings. Add an ID with `[[id]]` before the section.", level, strings.Repeat("=", level))
	}

	// List info
	if strings.HasPrefix(line, "*") {
		return "**Unordered List**: Lines starting with `*` create bullet list items."
	}
	if strings.HasPrefix(line, ".") {
		return "**Ordered List**: Lines starting with `.` create numbered list items."
	}
	if strings.HasPrefix(line, ";") {
		return "**Labeled List**: `; Term :: Definition` creates definition lists."
	}

	// Block info
	if strings.HasPrefix(line, "----") {
		return "**Literal Block**: Delimited by `----`, preserves formatting."
	}
	if strings.HasPrefix(line, "====") {
		return "**Example Block**: Delimited by `====`, for examples and figures."
	}
	if strings.HasPrefix(line, "____") {
		return "**Quote Block**: Delimited by `____`, for quoted text."
	}
	if strings.HasPrefix(line, "****") {
		return "**Sidebar Block**: Delimited by `****`, for sidebar content."
	}

	// Admonitions
	if strings.HasPrefix(strings.ToLower(line), "note:") {
		return "**Note**: A note admonition block."
	}
	if strings.HasPrefix(strings.ToLower(line), "tip:") {
		return "**Tip**: A tip admonition block."
	}
	if strings.HasPrefix(strings.ToLower(line), "warning:") {
		return "**Warning**: A warning admonition block."
	}
	if strings.HasPrefix(strings.ToLower(line), "caution:") {
		return "**Caution**: A caution admonition block."
	}
	if strings.HasPrefix(strings.ToLower(line), "important:") {
		return "**Important**: An important admonition block."
	}

	return ""
}

// Definition handles the textDocument/definition request.
func (s *Server) Definition(ctx context.Context, params *lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil || doc.AST == nil {
		return nil, nil
	}

	// Find cross-references and provide locations
	line := s.getLine(doc, int(params.Position.Line))

	// Check for cross-reference syntax <<id>>
	if strings.Contains(line, "<<") && strings.Contains(line, ">>") {
		// Extract the reference ID
		start := strings.Index(line, "<<")
		end := strings.Index(line, ">>")
		if start >= 0 && end > start {
			refID := line[start+2 : end]

			// Search for the section with this ID
			location := s.findSectionByID(doc.AST, refID, uri)
			if location != nil {
				return []lsp.Location{*location}, nil
			}
		}
	}

	return nil, nil
}

// findSectionByID finds a section by ID and returns its location.
func (s *Server) findSectionByID(document *ast.NodeDocument, id, uri string) *lsp.Location {
	for _, block := range document.Blocks {
		if section, ok := block.(*ast.NodeSection); ok {
			if section.ID == id {
				return &lsp.Location{
					URI: lsp.DocumentURI(uri),
					Range: lsp.Range{
						Start: lsp.Position{Line: section.Pos.Line - 1, Character: 0},
						End:   lsp.Position{Line: section.Pos.Line - 1, Character: len(section.Title)},
					},
				}
			}
			// Search recursively in children
			for _, child := range section.Children {
				if childSection, ok := child.(*ast.NodeSection); ok {
					if loc := s.findSectionByIDInNode(childSection, id, uri); loc != nil {
						return loc
					}
				}
			}
		}
	}
	return nil
}

// findSectionByIDInNode searches for a section ID in a node.
func (s *Server) findSectionByIDInNode(section *ast.NodeSection, id, uri string) *lsp.Location {
	if section.ID == id {
		return &lsp.Location{
			URI: lsp.DocumentURI(uri),
			Range: lsp.Range{
				Start: lsp.Position{Line: section.Pos.Line - 1, Character: 0},
				End:   lsp.Position{Line: section.Pos.Line - 1, Character: len(section.Title)},
			},
		}
	}
	for _, child := range section.Children {
		if childSection, ok := child.(*ast.NodeSection); ok {
			if loc := s.findSectionByIDInNode(childSection, id, uri); loc != nil {
				return loc
			}
		}
	}
	return nil
}

// Formatting handles the textDocument/formatting request.
func (s *Server) Formatting(ctx context.Context, params *lsp.DocumentFormattingParams) ([]lsp.TextEdit, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil {
		return nil, nil
	}

	// For now, return no edits (no formatting)
	// Could implement AsciiDoc formatting in the future
	return []lsp.TextEdit{}, nil
}

// Initialize handles the initialize request.
func (s *Server) Initialize(ctx context.Context, params *lsp.InitializeParams) (*lsp.InitializeResult, error) {
	s.clientCapabilities = &params.Capabilities

	return &lsp.InitializeResult{
		Capabilities: *s.capabilities,
	}, nil
}

// Shutdown handles the shutdown request.
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Exit handles the exit notification.
func (s *Server) Exit(ctx context.Context) error {
	os.Exit(0)
	return nil
}

// CodeAction handles the textDocument/codeAction request.
func (s *Server) CodeAction(ctx context.Context, params *lsp.CodeActionParams) ([]lsp.Command, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil {
		return nil, nil
	}

	var commands []lsp.Command

	// Add quick fixes for diagnostics
	for _, diag := range doc.Diagnostics {
		commands = append(commands, lsp.Command{
			Title: diag.Message,
			Command: "asciidoc.showDiagnostic",
		})
	}

	return commands, nil
}

// References handles the textDocument/references request.
func (s *Server) References(ctx context.Context, params *lsp.ReferenceParams) ([]lsp.Location, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil || doc.AST == nil {
		return nil, nil
	}

	// Find references to the symbol at the current position
	line := s.getLine(doc, params.Position.Line)

	var locations []lsp.Location

	// Find all sections if we're on a line that looks like a section definition
	if strings.HasPrefix(line, "=") {
		for _, block := range doc.AST.Blocks {
			if section, ok := block.(*ast.NodeSection); ok {
				locations = append(locations, lsp.Location{
					URI: lsp.DocumentURI(uri),
					Range: lsp.Range{
						Start: lsp.Position{Line: section.Pos.Line - 1, Character: 0},
						End:   lsp.Position{Line: section.Pos.Line - 1, Character: len(section.Title)},
					},
				})
			}
		}
	}

	return locations, nil
}

// Rename handles the textDocument/rename request.
func (s *Server) Rename(ctx context.Context, params *lsp.RenameParams) (*lsp.WorkspaceEdit, error) {
	// Basic rename implementation
	// In a full implementation, this would find all references and rename them
	return nil, fmt.Errorf("rename not yet implemented")
}

// DocumentHighlight handles the textDocument/documentHighlight request.
func (s *Server) DocumentHighlight(ctx context.Context, params *lsp.TextDocumentPositionParams) ([]lsp.DocumentHighlight, error) {
	uri := string(params.TextDocument.URI)

	s.docMu.RLock()
	doc := s.documents[uri]
	s.docMu.RUnlock()

	if doc == nil {
		return nil, nil
	}

	// Highlight the word at the current position
	line := s.getLine(doc, int(params.Position.Line))
	char := int(params.Position.Character)

	// Find word boundaries
	start, end := s.findWordBoundaries(line, char)
	if start < 0 {
		return nil, nil
	}

	word := line[start:end]
	var highlights []lsp.DocumentHighlight

	// Find all occurrences of this word in the document
	lines := splitLines(doc.Content)
	for lineNum, l := range lines {
		if idx := strings.Index(l, word); idx != -1 {
			highlights = append(highlights, lsp.DocumentHighlight{
				Range: lsp.Range{
					Start: lsp.Position{Line: lineNum, Character: idx},
					End:   lsp.Position{Line: lineNum, Character: idx + len(word)},
				},
			})
		}
	}

	return highlights, nil
}

// findWordBoundaries finds the start and end of a word at the given position.
func (s *Server) findWordBoundaries(line string, pos int) (start, end int) {
	if pos < 0 || pos >= len(line) {
		return -1, -1
	}

	// Find start (word characters only)
	start = pos
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	// Find end
	end = pos
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return -1, -1
	}

	return start, end
}

// isWordChar returns true if the character is a word character.
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
}

// Handle implements the jsonrpc2.Handler interface.
func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	s.conn = conn

	var result interface{}
	var err error

	switch req.Method {
	case "initialize":
		var params lsp.InitializeParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.Initialize(ctx, &params)

	case "initialized":
		result, err = nil, nil

	case "shutdown":
		err = s.Shutdown(ctx)
		result = nil

	case "exit":
		err = s.Exit(ctx)
		result = nil

	case "textDocument/didOpen":
		var params lsp.DidOpenTextDocumentParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		err = s.DidOpen(ctx, &params)
		result = nil

	case "textDocument/didChange":
		var params lsp.DidChangeTextDocumentParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		err = s.DidChange(ctx, &params)
		result = nil

	case "textDocument/didClose":
		var params lsp.DidCloseTextDocumentParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		err = s.DidClose(ctx, &params)
		result = nil

	case "textDocument/documentSymbol":
		var params lsp.DocumentSymbolParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.DocumentSymbol(ctx, &params)

	case "textDocument/completion":
		var params lsp.CompletionParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.Completion(ctx, &params)

	case "textDocument/hover":
		var params lsp.TextDocumentPositionParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.Hover(ctx, &params)

	case "textDocument/definition":
		var params lsp.TextDocumentPositionParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.Definition(ctx, &params)

	case "textDocument/formatting":
		var params lsp.DocumentFormattingParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.Formatting(ctx, &params)

	case "textDocument/codeAction":
		var params lsp.CodeActionParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.CodeAction(ctx, &params)

	case "textDocument/references":
		var params lsp.ReferenceParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.References(ctx, &params)

	case "textDocument/rename":
		var params lsp.RenameParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.Rename(ctx, &params)

	case "textDocument/documentHighlight":
		var params lsp.TextDocumentPositionParams
		if unmarshalErr := json.Unmarshal(*req.Params, &params); unmarshalErr != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeParseError, Message: unmarshalErr.Error()})
			return
		}
		result, err = s.DocumentHighlight(ctx, &params)

	default:
		result, err = nil, nil
	}

	if req.ID.Num > 0 || req.ID.Str != "" {
		if err != nil {
			conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
		} else {
			conn.Reply(ctx, req.ID, result)
		}
	}
}
