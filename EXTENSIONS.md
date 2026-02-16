# Extension System

The AsciiDoc parser now includes a powerful extension system that allows you to add custom block macros, inline macros, and tree processors.

## Overview

The extension system is located in `internal/extension/` and provides:

- **Block Macro Extensions**: Create custom block-level content
- **Inline Macro Extensions**: Create custom inline content
- **Tree Processor Extensions**: Modify the AST after parsing
- **Block Processor Extensions**: Transform blocks matching specific patterns

## Quick Start

```go
package main

import (
    "os"
    "github.com/haimiyahya/asciidoc-parser-go/internal/extension"
    "github.com/haimiyahya/asciidoc-parser-go/internal/parser"
    "github.com/haimiyahya/asciidoc-parser-go/internal/converter"
)

func main() {
    // Create extension registry
    registry := extension.NewRegistry()

    // Register custom extensions
    registry.RegisterBlockMacro("todo", &TodoBlockMacro{})
    registry.RegisterInlineMacro("badge", &BadgeInlineMacro{})
    registry.RegisterTreeProcessor(&TOCTreeProcessor{})

    // Parse with extensions
    src := `= Document Title

todo::Fix this bug[priority=high]

This is a badge:New[warning].`

    reader, _ := parser.NewReader(src)
    p := parser.NewParser(reader, parser.WithExtensionRegistry(registry))
    doc, _ := p.Parse()

    // Convert with extensions
    conv := converter.NewHTML5Converter().WithExtensionRegistry(registry)
    conv.Convert(doc, os.Stdout)
}
```

## Block Macros

Block macros create block-level content with the syntax `macro::target[attrs]`.

```go
type MyBlockMacro struct {
    *extension.BaseBlockMacro
}

func (m *MyBlockMacro) Process(target string, attrs map[string]string, content []string, pos ast.Position) (ast.Node, error) {
    return &ast.NodeBlock{
        Kind:  ast.TypeStyledBlock,
        Style: "custom",
        Lines: []string{"Custom: " + target},
        Pos:   pos,
    }, nil
}

// Register it
registry.RegisterBlockMacro("custom", &MyBlockMacro{
    BaseBlockMacro: extension.NewBaseBlockMacro("custom", false),
})
```

## Inline Macros

Inline macros create inline content with the syntax `macro:target[attrs]`.

```go
type MyInlineMacro struct {
    *extension.BaseInlineMacro
}

func (m *MyInlineMacro) Process(target string, attrs map[string]string) (string, error) {
    return fmt.Sprintf(`<span class="custom">%s</span>`, target), nil
}

// Register it
registry.RegisterInlineMacro("custom", &MyInlineMacro{
    BaseInlineMacro: extension.NewBaseInlineMacro("custom", "custom"),
})
```

## Tree Processors

Tree processors traverse and modify the entire AST after parsing.

```go
type MyTreeProcessor struct {
    *extension.BaseTreeProcessor
}

func (p *MyTreeProcessor) Process(document *ast.NodeDocument) error {
    // Process all sections, blocks, etc.
    for _, block := range document.Blocks {
        if section, ok := block.(*ast.NodeSection); ok {
            // Modify sections
            section.Title = strings.ToUpper(section.Title)
        }
    }
    return nil
}

func (p *MyTreeProcessor) Priority() int {
    return 100 // Lower = earlier execution
}
```

## Example Extensions

The package includes several example extensions:

### Block Macros

- `todo::` - Creates TODO blocks with priority levels
- `info::` - Creates information callout blocks
- `chart::` - Creates chart blocks

### Inline Macros

- `badge:label[]` - Creates badge/label elements
- `version:ver[]` - Creates version tags
- `label:text[color=red]` - Creates styled labels
- `footnote:id[text]` - Creates footnote references

### Tree Processors

- `SectionNumberer` - Numbers sections hierarchically
- `TOCTreeProcessor` - Generates table of contents

### Block Processors

- `NoteBlockProcessor` - Converts note-style blocks to admonitions

## Bundled Extensions

To register all example extensions:

```go
extension.RegisterBundledExtensions(registry)
```

## Files

- `internal/extension/extension.go` - Core extension system interfaces
- `internal/extension/examples.go` - Example extensions
- `internal/extension/extension_test.go` - Extension system tests

## Integration Points

### Parser Integration

```go
parser.WithExtensionRegistry(registry)
```

### Converter Integration

```go
converter.NewHTML5Converter().WithExtensionRegistry(registry)
```

## Testing

Run extension tests:

```bash
go test ./internal/extension -v
```

All tests pass:
- TestParseMacroAttributes
- TestNewRegistry
- TestRegistry_BlockMacro
- TestRegistry_InlineMacro
- TestRegistry_TreeProcessor
- TestRegistry_BlockProcessor
- TestAttributesManager
- TestJoinContent
- TestParseKeyValuePairs
