// Package extension provides a system for extending the AsciiDoc parser
// with custom block macros, inline macros, and tree processors.
//
// # Extension Types
//
// ## Block Macro Extensions
// Block macros create custom block-level content. Example:
//
//	tweet::jdoe[]
//
// ## Inline Macro Extensions
// Inline macros create custom inline content. Example:
//
//	Button:btn:[Click Me]
//
// ## Tree Processor Extensions
// Tree processors traverse and modify the entire AST after parsing.
// They can be used for table of contents generation, custom processing, etc.
//
// # Usage
//
//	registry := extension.NewRegistry()
//	registry.RegisterBlockMacro("note", &NoteBlockMacro{})
//	registry.RegisterInlineMacro("btn", &ButtonInlineMacro{})
//	registry.RegisterTreeProcessor(&TOCTreeProcessor{})
//
//	parser := parser.NewParser()
//	parser.SetExtensionRegistry(registry)
package extension

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// Registry manages all registered extensions.
type Registry struct {
	// blockMacros are registered by macro name (e.g., "note", "tweet")
	blockMacros map[string]BlockMacroProcessor

	// inlineMacros are registered by macro name (e.g., "btn", "kbd")
	inlineMacros map[string]InlineMacroProcessor

	// treeProcessors are called in order to process the entire AST
	treeProcessors []TreeProcessor

	// blockProcessors are called for blocks matching a pattern
	blockProcessors []BlockProcessor

	// preprocessors are called before parsing
	preprocessors []Preprocessor

	// postprocessors are called after conversion
	postprocessors []Postprocessor
}

// NewRegistry creates a new extension registry.
func NewRegistry() *Registry {
	return &Registry{
		blockMacros:     make(map[string]BlockMacroProcessor),
		inlineMacros:    make(map[string]InlineMacroProcessor),
		treeProcessors:  make([]TreeProcessor, 0),
		blockProcessors: make([]BlockProcessor, 0),
		preprocessors:   make([]Preprocessor, 0),
		postprocessors:  make([]Postprocessor, 0),
	}
}

// RegisterBlockMacro registers a block macro processor.
func (r *Registry) RegisterBlockMacro(name string, processor BlockMacroProcessor) {
	r.blockMacros[strings.ToLower(name)] = processor
}

// RegisterInlineMacro registers an inline macro processor.
func (r *Registry) RegisterInlineMacro(name string, processor InlineMacroProcessor) {
	r.inlineMacros[strings.ToLower(name)] = processor
}

// RegisterTreeProcessor registers a tree processor.
func (r *Registry) RegisterTreeProcessor(processor TreeProcessor) {
	r.treeProcessors = append(r.treeProcessors, processor)
}

// RegisterBlockProcessor registers a block processor.
func (r *Registry) RegisterBlockProcessor(processor BlockProcessor) {
	r.blockProcessors = append(r.blockProcessors, processor)
}

// RegisterPreprocessor registers a preprocessor.
func (r *Registry) RegisterPreprocessor(processor Preprocessor) {
	r.preprocessors = append(r.preprocessors, processor)
}

// RegisterPostprocessor registers a postprocessor.
func (r *Registry) RegisterPostprocessor(processor Postprocessor) {
	r.postprocessors = append(r.postprocessors, processor)
}

// GetBlockMacro returns a block macro processor by name.
func (r *Registry) GetBlockMacro(name string) (BlockMacroProcessor, bool) {
	processor, ok := r.blockMacros[strings.ToLower(name)]
	return processor, ok
}

// GetInlineMacro returns an inline macro processor by name.
func (r *Registry) GetInlineMacro(name string) (InlineMacroProcessor, bool) {
	processor, ok := r.inlineMacros[strings.ToLower(name)]
	return processor, ok
}

// GetTreeProcessors returns all registered tree processors.
func (r *Registry) GetTreeProcessors() []TreeProcessor {
	return r.treeProcessors
}

// GetBlockProcessors returns all registered block processors.
func (r *Registry) GetBlockProcessors() []BlockProcessor {
	return r.blockProcessors
}

// GetPreprocessors returns all registered preprocessors.
func (r *Registry) GetPreprocessors() []Preprocessor {
	return r.preprocessors
}

// GetPostprocessors returns all registered postprocessors.
func (r *Registry) GetPostprocessors() []Postprocessor {
	return r.postprocessors
}

// =============================================================================
// Block Macro Extension
// =============================================================================

// BlockMacroProcessor defines the interface for custom block macros.
//
// Block macros create block-level content in the document.
// Syntax: macro::target[attrlist]
//
// Example extensions could include:
//   - tweet::username[] - embed a tweet
//   - video::url[] - embed a video
//   - diagram::file.txt[] - include a diagram
type BlockMacroProcessor interface {
	// Name returns the macro name.
	Name() string

	// Process processes the block macro and returns the AST node to insert.
	// The target is the part after :: (e.g., "username" in tweet::username[])
	// The attrs map contains any attributes in the square brackets.
	Process(target string, attrs map[string]string, content []string, pos ast.Position) (ast.Node, error)

	// HasContent returns true if the macro expects content lines.
	// If true, lines following the macro (until a blank line) will be captured.
	HasContent() bool
}

// BlockMacroContext provides context for block macro processing.
type BlockMacroContext struct {
	// Name is the macro name.
	Name string

	// Target is the target of the macro (after ::).
	Target string

	// Attributes are the attributes from the square brackets.
	Attributes map[string]string

	// Content are any content lines following the macro.
	Content []string

	// Position is the source position.
	Position ast.Position

	// Document is the document being processed (read-only access).
	Document *ast.NodeDocument

	// AttributesManager provides access to document attributes.
	AttributesManager *AttributesManager
}

// =============================================================================
// Inline Macro Extension
// =============================================================================

// InlineMacroProcessor defines the interface for custom inline macros.
//
// Inline macros create inline content within paragraphs and other blocks.
// Syntax: macro:target[attrlist] or macro::target[attrlist]
//
// Example extensions could include:
//   - btn:label[] - a button element
//   - kbd:keys[] - keyboard shortcuts
//   - menu:path[] - UI menu paths
type InlineMacroProcessor interface {
	// Name returns the macro name.
	Name() string

	// Process processes the inline macro and returns the HTML string to insert.
	// The target is the part after : (e.g., "label" in btn:label[])
	// The attrs map contains any attributes in the square brackets.
	Process(target string, attrs map[string]string) (string, error)

	// Type returns the inline node type for this macro.
	// This helps the converter apply special handling if needed.
	Type() string
}

// InlineMacroContext provides context for inline macro processing.
type InlineMacroContext struct {
	// Name is the macro name.
	Name string

	// Target is the target of the macro.
	Target string

	// Attributes are the attributes from the square brackets.
	Attributes map[string]string

	// Text is the full text of the macro.
	Text string

	// Document is the document being processed (read-only access).
	Document *ast.NodeDocument

	// AttributesManager provides access to document attributes.
	AttributesManager *AttributesManager
}

// =============================================================================
// Tree Processor Extension
// =============================================================================

// TreeProcessor defines the interface for AST tree processors.
//
// Tree processors are called after parsing but before conversion.
// They can traverse and modify the entire AST.
//
// Common uses:
//   - Generate table of contents
//   - Number sections
//   - Add cross-reference links
//   - Process custom blocks globally
type TreeProcessor interface {
	// Process processes the document AST and can modify it.
	Process(document *ast.NodeDocument) error

	// Priority returns the priority (lower = earlier execution).
	// Tree processors are executed in priority order.
	Priority() int
}

// =============================================================================
// Block Processor Extension
// =============================================================================

// BlockProcessor defines the interface for processing blocks by pattern.
//
// Block processors can match blocks by name, style, or other attributes
// and transform them into custom AST nodes.
type BlockProcessor interface {
	// Match returns true if this processor should handle the given block.
	Match(block *ast.NodeBlock) bool

	// Process processes the block and returns a new AST node.
	// If the processor wants to keep the block as-is, it can return the original block.
	Process(block *ast.NodeBlock) (ast.Node, error)

	// Priority returns the priority (lower = earlier execution).
	Priority() int
}

// =============================================================================
// Preprocessor Extension
// =============================================================================

// Preprocessor defines the interface for preprocessing source lines.
//
// Preprocessors are called before parsing and can modify the source lines.
type Preprocessor interface {
	// Process processes the source lines before parsing.
	// It can add, remove, or modify lines.
	Process(lines []string) ([]string, error)

	// Priority returns the priority (lower = earlier execution).
	Priority() int
}

// =============================================================================
// Postprocessor Extension
// =============================================================================

// Postprocessor defines the interface for postprocessing output.
//
// Postprocessors are called after conversion and can modify the output.
type Postprocessor interface {
	// Process processes the converted output.
	Process(output string) (string, error)

	// Priority returns the priority (lower = earlier execution).
	Priority() int
}

// =============================================================================
// Attributes Manager
// =============================================================================

// AttributesManager provides access to document attributes for extensions.
type AttributesManager struct {
	document *ast.NodeDocument
}

// NewAttributesManager creates a new attributes manager.
func NewAttributesManager(doc *ast.NodeDocument) *AttributesManager {
	return &AttributesManager{document: doc}
}

// Get returns an attribute value.
func (am *AttributesManager) Get(name string) (string, bool) {
	if am.document.Attributes == nil {
		return "", false
	}
	val, ok := am.document.Attributes[name]
	return val, ok
}

// Set sets an attribute value.
func (am *AttributesManager) Set(name, value string) {
	if am.document.Attributes == nil {
		am.document.Attributes = make(map[string]string)
	}
	am.document.Attributes[name] = value
}

// Delete removes an attribute.
func (am *AttributesManager) Delete(name string) {
	if am.document.Attributes != nil {
		delete(am.document.Attributes, name)
	}
}

// GetAll returns all attributes.
func (am *AttributesManager) GetAll() map[string]string {
	if am.document.Attributes == nil {
		return make(map[string]string)
	}
	result := make(map[string]string, len(am.document.Attributes))
	for k, v := range am.document.Attributes {
		result[k] = v
	}
	return result
}

// =============================================================================
// Helper Types
// =============================================================================

// MacroAttributes represents parsed macro attributes.
type MacroAttributes struct {
	// Positional arguments in order.
	Positional []string

	// Named attributes.
	Named map[string]string

	// All raw attributes (before parsing).
	Raw map[string]string
}

// ParseMacroAttributes parses attributes from a block/inline macro.
//
// Input: "[attr1,val1,attr2=val2,key=val3]"
// Returns: MacroAttributes with Positional=[attr1,val1], Named={attr2:val2, key:val3}
func ParseMacroAttributes(attrString string) *MacroAttributes {
	attrs := &MacroAttributes{
		Positional: make([]string, 0),
		Named:      make(map[string]string),
		Raw:        make(map[string]string),
	}

	if attrString == "" {
		return attrs
	}

	// Remove surrounding brackets if present
	attrString = strings.TrimSpace(attrString)
	if strings.HasPrefix(attrString, "[") && strings.HasSuffix(attrString, "]") {
		attrString = attrString[1 : len(attrString)-1]
	}

	if attrString == "" {
		return attrs
	}

	// Split by comma
	parts := strings.Split(attrString, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a key=value pair
		if eqIdx := strings.Index(part, "="); eqIdx > 0 && eqIdx < len(part)-1 {
			key := strings.TrimSpace(part[:eqIdx])
			val := strings.TrimSpace(part[eqIdx+1:])
			attrs.Named[key] = val
			attrs.Raw[key] = val
		} else {
			// Positional argument
			attrs.Positional = append(attrs.Positional, part)
			attrs.Raw[part] = part
		}
	}

	return attrs
}

// =============================================================================
// Convenience Base Types
// =============================================================================

// BaseBlockMacro provides a base implementation for block macros.
type BaseBlockMacro struct {
	macroName  string
	hasContent bool
}

// NewBaseBlockMacro creates a new base block macro.
func NewBaseBlockMacro(name string, hasContent bool) *BaseBlockMacro {
	return &BaseBlockMacro{
		macroName:  name,
		hasContent: hasContent,
	}
}

// Name returns the macro name.
func (b *BaseBlockMacro) Name() string {
	return b.macroName
}

// HasContent returns whether the macro expects content lines.
func (b *BaseBlockMacro) HasContent() bool {
	return b.hasContent
}

// BaseInlineMacro provides a base implementation for inline macros.
type BaseInlineMacro struct {
	macroName string
	nodeType  string
}

// NewBaseInlineMacro creates a new base inline macro.
func NewBaseInlineMacro(name, nodeType string) *BaseInlineMacro {
	return &BaseInlineMacro{
		macroName: name,
		nodeType:  nodeType,
	}
}

// Name returns the macro name.
func (b *BaseInlineMacro) Name() string {
	return b.macroName
}

// Type returns the inline node type.
func (b *BaseInlineMacro) Type() string {
	return b.nodeType
}

// BaseTreeProcessor provides a base implementation for tree processors.
type BaseTreeProcessor struct {
	priority int
}

// NewBaseTreeProcessor creates a new base tree processor.
func NewBaseTreeProcessor(priority int) *BaseTreeProcessor {
	return &BaseTreeProcessor{priority: priority}
}

// Priority returns the priority.
func (b *BaseTreeProcessor) Priority() int {
	return b.priority
}

// =============================================================================
// Macro Name Pattern Matching
// =============================================================================

// MacroNamePattern represents a pattern for matching macro names.
type MacroNamePattern struct {
	pattern *regexp.Regexp
}

// NewMacroNamePattern creates a new macro name pattern.
func NewMacroNamePattern(pattern string) (*MacroNamePattern, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid macro name pattern: %w", err)
	}
	return &MacroNamePattern{pattern: re}, nil
}

// Match checks if a macro name matches the pattern.
func (p *MacroNamePattern) Match(name string) bool {
	return p.pattern.MatchString(name)
}

// =============================================================================
// Content Helpers
// =============================================================================

// ParseKeyValuePairs parses content lines as key=value pairs.
// Used by macros that accept structured content.
func ParseKeyValuePairs(content []string) map[string]string {
	result := make(map[string]string)
	for _, line := range content {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if eqIdx := strings.Index(line, "="); eqIdx > 0 {
			key := strings.TrimSpace(line[:eqIdx])
			val := strings.TrimSpace(line[eqIdx+1:])
			result[key] = val
		}
	}
	return result
}

// JoinContent joins content lines with newlines.
func JoinContent(content []string) string {
	return strings.Join(content, "\n")
}
