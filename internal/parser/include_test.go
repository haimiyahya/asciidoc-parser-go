// Package parser provides include directive tests.
package parser

import (
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncludeDirective verifies basic include directive functionality.
func TestIncludeDirective(t *testing.T) {
	t.Skip("Include processing not yet integrated with parser - TODO: integrate IncludeProcessor")

	src := `= Document Title

== Introduction

This document includes external content.

include::chapter-a.adoc[]

== Conclusion

Conclusion here.
`

	p, err := NewParserFromString(src)
	require.NoError(t, err)

	doc, _ := p.Parse()

	// Should have more blocks after processing include
	assert.Greater(t, len(doc.Blocks), 3, "document should have blocks from include directive")
}

// TestIncludeProcessor_ParseInclude tests the include directive parsing.
func TestIncludeProcessor_ParseInclude(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantPath  string
		wantAttrs map[string]string
		wantErr   bool
	}{
		{
			name:     "simple include",
			line:     "include::chapter1.adoc[]",
			wantPath: "chapter1.adoc",
			wantAttrs: map[string]string{},
			wantErr:  false,
		},
		{
			name:     "include with tags",
			line:     "include::chapter1.adoc[tags=feature]",
			wantPath: "chapter1.adoc",
			wantAttrs: map[string]string{"tags": "feature"},
			wantErr:  false,
		},
		{
			name:     "include with lines",
			line:     "include::chapter1.adoc[lines=1..10]",
			wantPath: "chapter1.adoc",
			wantAttrs: map[string]string{"lines": "1..10"},
			wantErr:  false,
		},
		{
			name:    "not an include directive",
			line:    "This is just text",
			wantErr: true,
		},
		{
			name:    "missing closing bracket",
			line:    "include::chapter1.adoc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directive, err := processor.ParseInclude(tt.line, 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPath, directive.Path)
				assert.Equal(t, tt.wantAttrs, directive.Attributes)
			}
		})
	}
}

// TestIncludeProcessor_NewIncludeProcessor tests the include processor creation.
func TestIncludeProcessor_NewIncludeProcessor(t *testing.T) {
	ip := processor.NewIncludeProcessor("/tmp")

	assert.NotNil(t, ip)
}

// TestIncludeProcessor_MaxDepth tests the max depth setting.
func TestIncludeProcessor_MaxDepth(t *testing.T) {
	ip := processor.NewIncludeProcessor("/tmp")
	ip.SetMaxDepth(50)

	// No assertion - just ensuring it compiles and runs
	assert.NotNil(t, ip)
}
