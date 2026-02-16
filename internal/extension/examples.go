// Package extension provides example extensions for the AsciiDoc parser.
package extension

import (
	"fmt"
	"html"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

// =============================================================================
// Example Block Macros
// =============================================================================

// TODOBlockMacro creates a TODO block with optional priority.
//
// Usage:
//   todo::Implement this feature[priority=high]
type TODOBlockMacro struct {
	*BaseBlockMacro
}

// NewTODOBlockMacro creates a new TODO block macro.
func NewTODOBlockMacro() *TODOBlockMacro {
	return &TODOBlockMacro{
		BaseBlockMacro: NewBaseBlockMacro("todo", false),
	}
}

// Process processes the TODO macro and returns a styled block node.
func (t *TODOBlockMacro) Process(target string, attrs map[string]string, content []string, pos ast.Position) (ast.Node, error) {
	// Get priority from attributes (default: medium)
	priority := "medium"
	if p, ok := attrs["priority"]; ok {
		priority = p
	}

	// Build the TODO block content
	title := fmt.Sprintf("TODO: %s", target)
	if priority != "" {
		title = fmt.Sprintf("TODO [%s]: %s", strings.ToUpper(priority), target)
	}

	return &ast.NodeBlock{
		Kind:     ast.TypeStyledBlock,
		Style:    "todo",
		Lines:    []string{title},
		Attributes: map[string]string{
			"priority": priority,
			"title":    target,
		},
		Pos: pos,
	}, nil
}

// InfoBlockMacro creates an information callout block.
//
// Usage:
//   info::Note something important[]
//   info::[title=Remember]
//   This is multi-line content.
type InfoBlockMacro struct {
	*BaseBlockMacro
}

// NewInfoBlockMacro creates a new Info block macro.
func NewInfoBlockMacro() *InfoBlockMacro {
	return &InfoBlockMacro{
		BaseBlockMacro: NewBaseBlockMacro("info", true),
	}
}

// Process processes the Info macro.
func (i *InfoBlockMacro) Process(target string, attrs map[string]string, content []string, pos ast.Position) (ast.Node, error) {
	title := target
	if t, ok := attrs["title"]; ok {
		title = t
	}

	lines := []string{title}
	lines = append(lines, content...)

	return &ast.NodeBlock{
		Kind:       ast.TypeAdmonition,
		Style:      "info",
		Lines:      lines,
		Attributes: attrs,
		Pos:        pos,
	}, nil
}

// ChartBlockMacro creates a simple chart/block diagram.
//
// Usage:
//   chart::sales[]
//   Q1, 100
//   Q2, 150
//   Q3, 200
type ChartBlockMacro struct {
	*BaseBlockMacro
}

// NewChartBlockMacro creates a new Chart block macro.
func NewChartBlockMacro() *ChartBlockMacro {
	return &ChartBlockMacro{
		BaseBlockMacro: NewBaseBlockMacro("chart", true),
	}
}

// Process processes the Chart macro.
func (c *ChartBlockMacro) Process(target string, attrs map[string]string, content []string, pos ast.Position) (ast.Node, error) {
	// Chart type from target or attributes
	chartType := target
	if chartType == "" {
		chartType = "bar"
		if t, ok := attrs["type"]; ok {
			chartType = t
		}
	}

	// Store chart data in attributes
	chartData := strings.Join(content, "\n")

	return &ast.NodeBlock{
		Kind:  ast.TypeStyledBlock,
		Style: "chart",
		Lines: []string{fmt.Sprintf("Chart: %s", chartType)},
		Attributes: map[string]string{
			"chart-type": chartType,
			"chart-data": chartData,
		},
		Pos: pos,
	}, nil
}

// =============================================================================
// Example Inline Macros
// =============================================================================

// BadgeInlineMacro creates a badge/label element.
//
// Usage:
//   badge:New[]
//   badge:Beta[color=warning]
type BadgeInlineMacro struct {
	*BaseInlineMacro
}

// NewBadgeInlineMacro creates a new Badge inline macro.
func NewBadgeInlineMacro() *BadgeInlineMacro {
	return &BadgeInlineMacro{
		BaseInlineMacro: NewBaseInlineMacro("badge", "badge"),
	}
}

// Process processes the Badge macro.
func (b *BadgeInlineMacro) Process(target string, attrs map[string]string) (string, error) {
	color := "primary"
	if c, ok := attrs["color"]; ok {
		color = c
	}

	// Map color names to CSS classes
	colorClass := map[string]string{
		"primary":   "badge-primary",
		"secondary": "badge-secondary",
		"success":   "badge-success",
		"warning":   "badge-warning",
		"danger":    "badge-danger",
		"info":      "badge-info",
		"light":     "badge-light",
		"dark":      "badge-dark",
	}

	className := colorClass[color]
	if className == "" {
		className = "badge-" + color
	}

	return fmt.Sprintf(`<span class="badge %s">%s</span>`, className, html.EscapeString(target)), nil
}

// VersionInlineMacro creates a version tag/badge.
//
// Usage:
//   version:1.0.0[]
//   version:2.0.0[status=stable]
type VersionInlineMacro struct {
	*BaseInlineMacro
}

// NewVersionInlineMacro creates a new Version inline macro.
func NewVersionInlineMacro() *VersionInlineMacro {
	return &VersionInlineMacro{
		BaseInlineMacro: NewBaseInlineMacro("version", "version"),
	}
}

// Process processes the Version macro.
func (v *VersionInlineMacro) Process(target string, attrs map[string]string) (string, error) {
	status := ""
	if s, ok := attrs["status"]; ok {
		status = fmt.Sprintf(` <span class="version-status">%s</span>`, html.EscapeString(s))
	}

	return fmt.Sprintf(`<span class="version">v%s%s</span>`, html.EscapeString(target), status), nil
}

// LabelInlineMacro creates a styled label.
//
// Usage:
//   label:Important[]
//   label:Deprecated[color=red]
type LabelInlineMacro struct {
	*BaseInlineMacro
}

// NewLabelInlineMacro creates a new Label inline macro.
func NewLabelInlineMacro() *LabelInlineMacro {
	return &LabelInlineMacro{
		BaseInlineMacro: NewBaseInlineMacro("label", "label"),
	}
}

// Process processes the Label macro.
func (l *LabelInlineMacro) Process(target string, attrs map[string]string) (string, error) {
	color := "default"
	if c, ok := attrs["color"]; ok {
		color = c
	}

	return fmt.Sprintf(`<span class="label label-%s">%s</span>`, html.EscapeString(color), html.EscapeString(target)), nil
}

// FootnoteRefInlineMacro creates a footnote reference.
//
// Usage:
//   footnote:note1[Text of footnote]
type FootnoteRefInlineMacro struct {
	*BaseInlineMacro
	footnoteID int
}

// NewFootnoteRefInlineMacro creates a new FootnoteRef inline macro.
func NewFootnoteRefInlineMacro() *FootnoteRefInlineMacro {
	return &FootnoteRefInlineMacro{
		BaseInlineMacro: NewBaseInlineMacro("footnote", "footnote-ref"),
	}
}

// Process processes the FootnoteRef macro.
func (f *FootnoteRefInlineMacro) Process(target string, attrs map[string]string) (string, error) {
	f.footnoteID++

	text := target
	if t, ok := attrs["text"]; ok {
		text = t
	}

	return fmt.Sprintf(`<sup class="footnote-ref" id="fr%d">[%d]</sup> <span class="footnote-text">%s</span>`,
		f.footnoteID, f.footnoteID, html.EscapeString(text)), nil
}

// =============================================================================
// Example Tree Processors
// =============================================================================

// TOCTreeProcessor generates a table of contents.
type TOCTreeProcessor struct {
	*BaseTreeProcessor
	// Levels indicates how many levels to include (0 = all)
	Levels int
	// ID is the ID for the TOC div
	ID string
	// Title is the title for the TOC (empty = no title)
	Title string
}

// NewTOCTreeProcessor creates a new TOC tree processor.
func NewTOCTreeProcessor(levels int) *TOCTreeProcessor {
	return &TOCTreeProcessor{
		BaseTreeProcessor: NewBaseTreeProcessor(100),
		Levels:            levels,
		ID:                "toc",
		Title:             "Table of Contents",
	}
}

// Process generates a table of contents and prepends it to the document.
func (t *TOCTreeProcessor) Process(document *ast.NodeDocument) error {
	if document.Blocks == nil {
		return nil
	}

	// Collect sections
	sections := t.collectSections(document.Blocks, 1)
	if len(sections) == 0 {
		return nil
	}

	// Create TOC block
	tocLines := t.generateTOCLines(sections)

	tocBlock := &ast.NodeBlock{
		Kind:       ast.TypeStyledBlock,
		Style:      "toc",
		Lines:      tocLines,
		Attributes: map[string]string{
			"toc-id":    t.ID,
			"toc-title": t.Title,
			"toc-level": fmt.Sprintf("%d", t.Levels),
		},
		Pos: ast.Position{Line: 0},
	}

	// Prepend TOC to document blocks
	document.Blocks = append([]ast.Node{tocBlock}, document.Blocks...)

	return nil
}

// SectionInfo holds information about a section for TOC generation.
type SectionInfo struct {
	Level   int
	ID      string
	Title   string
	Number  string
	Parent  *SectionInfo
	Children []*SectionInfo
}

// collectSections recursively collects section information.
func (t *TOCTreeProcessor) collectSections(blocks []ast.Node, level int) []*SectionInfo {
	var sections []*SectionInfo

	for _, block := range blocks {
		switch n := block.(type) {
		case *ast.NodeSection:
			// Check if we should include this level
			if t.Levels > 0 && level > t.Levels {
				continue
			}

			info := &SectionInfo{
				Level:  level,
				ID:     n.ID,
				Title:  n.Title,
				Number: "", // Will be filled by SectionNumberer if present
			}

			// Recursively collect subsections from Children
			for _, child := range n.Children {
				if subsection, ok := child.(*ast.NodeSection); ok {
					subInfo := &SectionInfo{
						Level: level + 1,
						ID:    subsection.ID,
						Title: subsection.Title,
						Parent: info,
					}
					info.Children = append(info.Children, subInfo)
				}
			}

			sections = append(sections, info)

		case *ast.NodeBlock:
			// NodeBlock doesn't have nested blocks, skip
			_ = n
		}
	}

	return sections
}

// generateTOCLines generates TOC HTML lines.
func (t *TOCTreeProcessor) generateTOCLines(sections []*SectionInfo) []string {
	var lines []string

	if t.Title != "" {
		lines = append(lines, fmt.Sprintf("== %s", t.Title))
	}

	for _, section := range sections {
		indent := strings.Repeat("*", section.Level)
		line := fmt.Sprintf("%s xref:%s[%s]", indent, section.ID, section.Title)
		lines = append(lines, line)
	}

	return lines
}

// SectionNumbererTreeProcessor numbers sections hierarchically.
type SectionNumbererTreeProcessor struct {
	*BaseTreeProcessor
}

// NewSectionNumbererTreeProcessor creates a new section numberer.
func NewSectionNumbererTreeProcessor() *SectionNumbererTreeProcessor {
	return &SectionNumbererTreeProcessor{
		BaseTreeProcessor: NewBaseTreeProcessor(50),
	}
}

// Process numbers all sections in the document.
func (s *SectionNumbererTreeProcessor) Process(document *ast.NodeDocument) error {
	counter := make(map[int]int) // level -> count
	return s.numberBlocks(document.Blocks, counter, 1)
}

// numberBlocks recursively numbers sections.
func (s *SectionNumbererTreeProcessor) numberBlocks(blocks []ast.Node, counter map[int]int, level int) error {
	for _, block := range blocks {
		if section, ok := block.(*ast.NodeSection); ok {
			counter[level]++
			number := s.formatNumber(counter, level)

			// Store the number in attributes since NodeSection doesn't have a Number field
			if section.Attributes == nil {
				section.Attributes = make(map[string]string)
			}
			section.Attributes["number"] = number

			// Reset counters for deeper levels
			for l := level + 1; l <= len(counter); l++ {
				counter[l] = 0
			}
		}
	}
	return nil
}

// formatNumber formats a section number from the counter.
func (s *SectionNumbererTreeProcessor) formatNumber(counter map[int]int, level int) string {
	var parts []string
	for l := 1; l <= level; l++ {
		if counter[l] > 0 {
			parts = append(parts, fmt.Sprintf("%d", counter[l]))
		}
	}
	return strings.Join(parts, ".")
}

// =============================================================================
// Example Block Processor
// =============================================================================

// NoteBlockProcessor converts note-style blocks to admonitions.
//
// Matches blocks with style="note" or custom note types.
type NoteBlockProcessor struct {
	*BaseTreeProcessor
	noteTypes map[string]string
}

// NewNoteBlockProcessor creates a new note block processor.
func NewNoteBlockProcessor() *NoteBlockProcessor {
	return &NoteBlockProcessor{
		BaseTreeProcessor: NewBaseTreeProcessor(75),
		noteTypes: map[string]string{
			"note":      "NOTE",
			"tip":       "TIP",
			"warning":   "WARNING",
			"important": "IMPORTANT",
			"caution":   "CAUTION",
		},
	}
}

// Match checks if this processor should handle the block.
func (n *NoteBlockProcessor) Match(block *ast.NodeBlock) bool {
	if block.Style == "" {
		return false
	}
	_, ok := n.noteTypes[block.Style]
	return ok
}

// Process converts the note block to an admonition.
func (n *NoteBlockProcessor) Process(block *ast.NodeBlock) (ast.Node, error) {
	noteType := n.noteTypes[block.Style]
	if noteType == "" {
		return block, nil
	}

	// Add the type as the first line
	lines := append([]string{noteType}, block.Lines...)

	return &ast.NodeBlock{
		Kind:       ast.TypeAdmonition,
		Style:      block.Style,
		Lines:      lines,
		Attributes: block.Attributes,
		Pos:        block.Pos,
	}, nil
}

// Priority returns the processor priority.
func (n *NoteBlockProcessor) Priority() int {
	return n.BaseTreeProcessor.Priority()
}

// =============================================================================
// Bundled Extensions
// =============================================================================

// RegisterBundledExtensions registers all built-in example extensions.
func RegisterBundledExtensions(registry *Registry) {
	// Block macros
	registry.RegisterBlockMacro("todo", NewTODOBlockMacro())
	registry.RegisterBlockMacro("info", NewInfoBlockMacro())
	registry.RegisterBlockMacro("chart", NewChartBlockMacro())

	// Inline macros
	registry.RegisterInlineMacro("badge", NewBadgeInlineMacro())
	registry.RegisterInlineMacro("version", NewVersionInlineMacro())
	registry.RegisterInlineMacro("label", NewLabelInlineMacro())
	registry.RegisterInlineMacro("footnote", NewFootnoteRefInlineMacro())

	// Tree processors
	registry.RegisterTreeProcessor(NewSectionNumbererTreeProcessor())
	registry.RegisterBlockProcessor(NewNoteBlockProcessor())
}
