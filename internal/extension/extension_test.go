// Package extension provides tests for the extension system.
package extension

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
)

func TestParseMacroAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPos  []string
		wantNamed map[string]string
	}{
		{
			name:     "empty",
			input:    "[]",
			wantPos:  []string{},
			wantNamed: map[string]string{},
		},
		{
			name:     "positional only",
			input:    "[val1,val2,val3]",
			wantPos:  []string{"val1", "val2", "val3"},
			wantNamed: map[string]string{},
		},
		{
			name:     "named only",
			input:    "[key1=val1,key2=val2]",
			wantPos:  []string{},
			wantNamed: map[string]string{"key1": "val1", "key2": "val2"},
		},
		{
			name:     "mixed",
			input:    "[pos1,key1=val1,pos2]",
			wantPos:  []string{"pos1", "pos2"},
			wantNamed: map[string]string{"key1": "val1"},
		},
		{
			name:     "with spaces",
			input:    "[val1, key2 = val2 , val3]",
			wantPos:  []string{"val1", "val3"},
			wantNamed: map[string]string{"key2": "val2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := ParseMacroAttributes(tt.input)

			// Check positional
			if len(attrs.Positional) != len(tt.wantPos) {
				t.Errorf("Positional count = %v, want %v", len(attrs.Positional), len(tt.wantPos))
			}
			for i, want := range tt.wantPos {
				if i < len(attrs.Positional) && attrs.Positional[i] != want {
					t.Errorf("Positional[%d] = %v, want %v", i, attrs.Positional[i], want)
				}
			}

			// Check named
			if len(attrs.Named) != len(tt.wantNamed) {
				t.Errorf("Named count = %v, want %v", len(attrs.Named), len(tt.wantNamed))
			}
			for k, wantVal := range tt.wantNamed {
				if gotVal, ok := attrs.Named[k]; !ok || gotVal != wantVal {
					t.Errorf("Named[%s] = %v, want %v", k, gotVal, wantVal)
				}
			}
		})
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.blockMacros == nil {
		t.Error("blockMacros map not initialized")
	}
	if registry.inlineMacros == nil {
		t.Error("inlineMacros map not initialized")
	}
	if registry.treeProcessors == nil {
		t.Error("treeProcessors slice not initialized")
	}
}

func TestRegistry_BlockMacro(t *testing.T) {
	registry := NewRegistry()

	// Register a test macro
	testMacro := &TestBlockMacro{name: "test"}
	registry.RegisterBlockMacro("test", testMacro)

	// Retrieve it
	processor, ok := registry.GetBlockMacro("test")
	if !ok {
		t.Fatal("GetBlockMacro() returned false for registered macro")
	}
	if processor != testMacro {
		t.Error("GetBlockMacro() returned wrong processor")
	}

	// Case insensitive
	processor, ok = registry.GetBlockMacro("TEST")
	if !ok {
		t.Fatal("GetBlockMacro() should be case insensitive")
	}

	// Unknown macro
	_, ok = registry.GetBlockMacro("unknown")
	if ok {
		t.Error("GetBlockMacro() returned true for unknown macro")
	}
}

func TestRegistry_InlineMacro(t *testing.T) {
	registry := NewRegistry()

	// Register a test macro
	testMacro := &TestInlineMacro{name: "test"}
	registry.RegisterInlineMacro("test", testMacro)

	// Retrieve it
	processor, ok := registry.GetInlineMacro("test")
	if !ok {
		t.Fatal("GetInlineMacro() returned false for registered macro")
	}
	if processor != testMacro {
		t.Error("GetInlineMacro() returned wrong processor")
	}

	// Case insensitive
	processor, ok = registry.GetInlineMacro("TEST")
	if !ok {
		t.Fatal("GetInlineMacro() should be case insensitive")
	}

	// Unknown macro
	_, ok = registry.GetInlineMacro("unknown")
	if ok {
		t.Error("GetInlineMacro() returned true for unknown macro")
	}
}

func TestRegistry_TreeProcessor(t *testing.T) {
	registry := NewRegistry()

	// Register test processors
	p1 := &TestTreeProcessor{priority: 10}
	p2 := &TestTreeProcessor{priority: 5}
	registry.RegisterTreeProcessor(p1)
	registry.RegisterTreeProcessor(p2)

	// Retrieve them
	processors := registry.GetTreeProcessors()
	if len(processors) != 2 {
		t.Fatalf("GetTreeProcessors() returned %d processors, want 2", len(processors))
	}

	// Note: GetTreeProcessors returns them in registration order
	// Sorting happens in runTreeProcessors()
	if processors[0].Priority() != 10 {
		t.Errorf("First processor priority = %d, want 10", processors[0].Priority())
	}
	if processors[1].Priority() != 5 {
		t.Errorf("Second processor priority = %d, want 5", processors[1].Priority())
	}
}

func TestRegistry_BlockProcessor(t *testing.T) {
	registry := NewRegistry()

	// Register test processor
	processor := &TestBlockProcessor{priority: 1}
	registry.RegisterBlockProcessor(processor)

	// Retrieve it
	processors := registry.GetBlockProcessors()
	if len(processors) != 1 {
		t.Fatalf("GetBlockProcessors() returned %d processors, want 1", len(processors))
	}
	if processors[0] != processor {
		t.Error("GetBlockProcessors() returned wrong processor")
	}
}

func TestAttributesManager(t *testing.T) {
	doc := &ast.NodeDocument{
		Attributes: map[string]string{
			"key1": "val1",
			"key2": "val2",
		},
	}

	manager := NewAttributesManager(doc)

	// Get existing attribute
	val, ok := manager.Get("key1")
	if !ok || val != "val1" {
		t.Errorf("Get(key1) = %v, %v; want val1, true", val, ok)
	}

	// Get non-existent attribute
	_, ok = manager.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) returned true")
	}

	// Set attribute
	manager.Set("key3", "val3")
	val, ok = manager.Get("key3")
	if !ok || val != "val3" {
		t.Errorf("After Set(key3), Get() = %v, %v; want val3, true", val, ok)
	}

	// Delete attribute
	manager.Delete("key1")
	_, ok = manager.Get("key1")
	if ok {
		t.Error("After Delete(key1), Get() returned true")
	}

	// GetAll
	attrs := manager.GetAll()
	if len(attrs) != 2 { // key2 and key3
		t.Errorf("GetAll() returned %d attributes, want 2", len(attrs))
	}
}

func TestJoinContent(t *testing.T) {
	content := []string{"line1", "line2", "line3"}
	result := JoinContent(content)
	expected := "line1\nline2\nline3"
	if result != expected {
		t.Errorf("JoinContent() = %q, want %q", result, expected)
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	content := []string{
		"key1=value1",
		"key2=value2",
		"// comment",
		"key3=value3",
	}

	result := ParseKeyValuePairs(content)

	if len(result) != 3 {
		t.Fatalf("ParseKeyValuePairs() returned %d pairs, want 3", len(result))
	}
	if result["key1"] != "value1" {
		t.Errorf("key1 = %v, want value1", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("key2 = %v, want value2", result["key2"])
	}
	if result["key3"] != "value3" {
		t.Errorf("key3 = %v, want value3", result["key3"])
	}
}

// =============================================================================
// Test Doubles
// =============================================================================

type TestBlockMacro struct {
	*BaseBlockMacro
	name string
}

func (m *TestBlockMacro) Name() string {
	return m.name
}

func (m *TestBlockMacro) Process(target string, attrs map[string]string, content []string, pos ast.Position) (ast.Node, error) {
	return &ast.NodeBlock{
		Kind:  ast.TypeBlock,
		Lines: []string{"test: " + target},
		Pos:   pos,
	}, nil
}

type TestInlineMacro struct {
	*BaseInlineMacro
	name string
}

func (m *TestInlineMacro) Name() string {
	return m.name
}

func (m *TestInlineMacro) Process(target string, attrs map[string]string) (string, error) {
	return "<span class=\"test\">" + target + "</span>", nil
}

type TestTreeProcessor struct {
	*BaseTreeProcessor
	priority int
	called   bool
}

func (p *TestTreeProcessor) Process(document *ast.NodeDocument) error {
	p.called = true
	return nil
}

func (p *TestTreeProcessor) Priority() int {
	return p.priority
}

type TestBlockProcessor struct {
	*BaseTreeProcessor
	priority int
}

func (p *TestBlockProcessor) Match(block *ast.NodeBlock) bool {
	return block.Style == "test"
}

func (p *TestBlockProcessor) Process(block *ast.NodeBlock) (ast.Node, error) {
	return block, nil
}

func (p *TestBlockProcessor) Priority() int {
	return p.priority
}
