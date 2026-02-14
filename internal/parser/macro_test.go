// Package parser provides tests for block macro parsing.
package parser

import (
	"strings"
	"testing"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImageMacro(t *testing.T) {
	source := `image::path/to/image.png[]`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, doc)

	assert.Len(t, doc.Blocks, 1)

	macro, ok := doc.Blocks[0].(*ast.MacroNode)
	require.True(t, ok, "First block should be a MacroNode")
	assert.Equal(t, ast.TypeMacro, macro.Type())
	assert.Equal(t, "image", macro.Target)
	assert.Equal(t, "path/to/image.png", macro.Path)
}

func TestParseVideoMacro(t *testing.T) {
	source := `video::path/to/video.mp4[]`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	macro, ok := doc.Blocks[0].(*ast.MacroNode)
	require.True(t, ok)
	assert.Equal(t, "video", macro.Target)
	assert.Equal(t, "path/to/video.mp4", macro.Path)
}

func TestParseAudioMacro(t *testing.T) {
	source := `audio::path/to/audio.mp3[]`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	macro, ok := doc.Blocks[0].(*ast.MacroNode)
	require.True(t, ok)
	assert.Equal(t, "audio", macro.Target)
}

func TestParseMacroWithAttributes(t *testing.T) {
	source := `image::image.png[width=200,height=100]`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	macro, ok := doc.Blocks[0].(*ast.MacroNode)
	require.True(t, ok)
	assert.Equal(t, "image", macro.Target)
	assert.Equal(t, "image.png", macro.Path)
	assert.NotEmpty(t, macro.Attributes)
}

func TestParseMacroWithParagraphAfter(t *testing.T) {
	source := `image::photo.jpg

This is a regular paragraph.`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	require.Len(t, doc.Blocks, 2)

	macro, ok := doc.Blocks[0].(*ast.MacroNode)
	require.True(t, ok)
	assert.Equal(t, "image", macro.Target)

	para, ok := doc.Blocks[1].(*ast.NodeParagraph)
	require.True(t, ok)
	assert.Equal(t, "This is a regular paragraph.", para.Text)
}

func TestParseMixedMacros(t *testing.T) {
	source := `image::img1.png

Some text

video::vid1.mp4`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	// Should have: image macro, paragraph, video macro
	require.Len(t, doc.Blocks, 3)

	macro1, ok := doc.Blocks[0].(*ast.MacroNode)
	require.True(t, ok)
	assert.Equal(t, "image", macro1.Target)

	_, ok = doc.Blocks[1].(*ast.NodeParagraph)
	require.True(t, ok)

	macro2, ok := doc.Blocks[2].(*ast.MacroNode)
	require.True(t, ok)
	assert.Equal(t, "video", macro2.Target)
}

func TestImageMacroHTML5Conversion(t *testing.T) {
	source := `image::path/to/image.png`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf strings.Builder
	conv := converter.NewHTML5Converter()
	err = conv.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<img src="path/to/image.png"`)
}

func TestVideoMacroHTML5Conversion(t *testing.T) {
	source := `video::path/to/video.mp4`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf strings.Builder
	conv := converter.NewHTML5Converter()
	err = conv.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<video src="path/to/video.mp4">`)
	assert.Contains(t, output, `</video>`)
}

func TestAudioMacroHTML5Conversion(t *testing.T) {
	source := `audio::path/to/audio.mp3`

	p, err := NewParserFromString(source)
	require.NoError(t, err)

	doc, err := p.Parse()
	require.NoError(t, err)

	var buf strings.Builder
	conv := converter.NewHTML5Converter()
	err = conv.Convert(doc, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<audio src="path/to/audio.mp3">`)
	assert.Contains(t, output, `</audio>`)
}
