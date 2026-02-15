// Package processor provides include directive processing for AsciiDoc.
package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// IncludeProcessor handles include::[] directives in AsciiDoc documents.
type IncludeProcessor struct {
	// baseDir is the base directory for resolving relative paths.
	baseDir string

	// maxDepth is the maximum include depth to prevent infinite loops.
	maxDepth int

	// includeStack tracks currently processed files for circular reference detection.
	includeStack map[string]bool

	// tagFilter specifies which tagged content to include.
	tagFilter map[string]bool

	// includeDepth tracks current nesting depth of includes.
	includeDepth int
}

// NewIncludeProcessor creates a new include processor.
func NewIncludeProcessor(baseDir string) *IncludeProcessor {
	return &IncludeProcessor{
		baseDir:   baseDir,
		maxDepth:   100,
		includeStack: make(map[string]bool),
		tagFilter:   make(map[string]bool),
		includeDepth: 0,
	}
}

// IncludeDirective represents a parsed include::[] directive.
type IncludeDirective struct {
	Path      string
	Attributes map[string]string
	LineNumber int
}

// ParseInclude parses an include::[] directive string.
func ParseInclude(line string, lineNumber int) (*IncludeDirective, error) {
	// Format: include::path[] or include::path[attrs]
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "include::") {
		return nil, fmt.Errorf("not an include directive")
	}

	rest := strings.TrimPrefix(trimmed, "include::")
	if rest == "" {
		return nil, fmt.Errorf("empty include directive")
	}

	// Find the opening bracket [
	openBracket := strings.Index(rest, "[")
	if openBracket == -1 {
		return nil, fmt.Errorf("missing opening bracket")
	}

	// Extract path (everything before [)
	pathPart := rest[:openBracket]

	// Find the closing bracket ]
	closeBracket := strings.Index(rest[openBracket:], "]")
	if closeBracket == -1 {
		return nil, fmt.Errorf("missing closing bracket")
	}

	// Extract attributes (everything between [ and ])
	attrsPart := rest[openBracket+1 : openBracket+closeBracket]

	directive := &IncludeDirective{
		Path:       strings.TrimSpace(pathPart),
		Attributes: parseIncludeAttributes(attrsPart),
		LineNumber: lineNumber,
	}

	return directive, nil
}

// parseIncludeAttributes parses attributes from include directive.
func parseIncludeAttributes(attrs string) map[string]string {
	result := make(map[string]string)

	if attrs == "" {
		return result
	}

	// Parse key=value pairs separated by commas
	pairs := strings.Split(attrs, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split by first =
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		result[key] = value
	}

	return result
}

// Process processes an include directive and returns included content.
func (p *IncludeProcessor) Process(directive *IncludeDirective) (string, error) {
	// Check for circular references
	fullPath := p.resolvePath(directive.Path)
	if fullPath == "" {
		return "", fmt.Errorf("empty include path")
	}

	if _, exists := p.includeStack[fullPath]; exists {
		return "", fmt.Errorf("circular include detected: %s", fullPath)
	}

	// Check include depth
	if p.includeDepth >= p.maxDepth {
		return "", fmt.Errorf("maximum include depth exceeded")
	}

	// Read file content
	content, err := p.readFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read include file: %w", err)
	}

	// Add to include stack
	p.includeStack[fullPath] = true
	defer func() {
		delete(p.includeStack, fullPath)
	}()

	// Filter by tags if specified
	if tags, ok := directive.Attributes["tags"]; ok {
		content = p.filterByTags(content, tags)
	}

	// Filter by line range if specified
	if lines, ok := directive.Attributes["lines"]; ok {
		content = p.filterByLineRange(content, lines)
	}

	return content, nil
}

// resolvePath resolves an include path relative to base directory.
func (p *IncludeProcessor) resolvePath(includePath string) string {
	// Clean up the path
	cleanPath := strings.TrimSpace(includePath)

	// If absolute, use as is
	if filepath.IsAbs(cleanPath) {
		return cleanPath
	}

	// Resolve relative to base directory
	return filepath.Join(p.baseDir, cleanPath)
}

// readFile reads content from a file.
func (p *IncludeProcessor) readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// filterByTags filters content by tag markers.
// Tags are marked with // tag::name in AsciiDoc.
func (p *IncludeProcessor) filterByTags(content string, tagsSpec string) string {
	// Parse tag list: tags=tag1,tag2
	tagList := strings.Split(tagsSpec, ",")
	includeTags := make(map[string]bool)
	excludeTags := make(map[string]bool)

	for _, tag := range tagList {
		tag = strings.TrimSpace(tag)
		if strings.HasPrefix(tag, "!") {
			excludeTags[strings.TrimPrefix(tag, "!")] = true
		} else {
			includeTags[tag] = true
		}
	}

	var result strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for tag marker: // tag::name
		if strings.HasPrefix(trimmed, "// tag::") {
			tagName := strings.TrimPrefix(trimmed, "// tag::")
			tagName = strings.TrimSpace(tagName)
			if includeTags[tagName] {
				result.WriteString(line)
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

// filterByLineRange filters content by line range specification.
// Format: lines=1..5 or lines=1,3,5
func (p *IncludeProcessor) filterByLineRange(content string, lineSpec string) string {
	lines := strings.Split(content, "\n")

	// Parse line specification
	ranges := p.parseLineRanges(lineSpec)

	var result strings.Builder
	for i, line := range lines {
		// Lines are 1-indexed in spec
		lineNum := i + 1
		if p.inLineRanges(lineNum, ranges) {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// parseLineRanges parses a line range specification.
func (p *IncludeProcessor) parseLineRanges(spec string) []lineRange {
	var ranges []lineRange

	// Split by comma for multiple ranges
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for range (1..5)
		if strings.Contains(part, "..") {
			nums := strings.Split(part, "..")
			if len(nums) == 2 {
				start, err1 := strconv.Atoi(nums[0])
				end, err2 := strconv.Atoi(nums[1])
				if err1 == nil && err2 == nil {
					ranges = append(ranges, lineRange{start: start, end: end})
				}
			}
		} else {
			// Single number
			num, err := strconv.Atoi(part)
			if err == nil {
				ranges = append(ranges, lineRange{start: num, end: num})
			}
		}
	}

	return ranges
}

// inLineRanges checks if a line number is within any of the ranges.
func (p *IncludeProcessor) inLineRanges(lineNum int, ranges []lineRange) bool {
	for _, r := range ranges {
		if lineNum >= r.start && lineNum <= r.end {
			return true
		}
	}
	return false
}

// lineRange represents a range of line numbers.
type lineRange struct {
	start int
	end   int
}

// SetMaxDepth sets the maximum include depth.
func (p *IncludeProcessor) SetMaxDepth(depth int) {
	p.maxDepth = depth
}

// SetTagFilter sets which tags to include.
func (p *IncludeProcessor) SetTagFilter(tags map[string]bool) {
	p.tagFilter = tags
}

// Reset clears the include processor state.
func (p *IncludeProcessor) Reset() {
	p.includeStack = make(map[string]bool)
	p.tagFilter = make(map[string]bool)
	p.includeDepth = 0
}
