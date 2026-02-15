// Package reader provides a line-oriented reader for AsciiDoc source input.
//
// This package implements the foundational "Reader" component from Asciidoctor's
// architecture, designed to mimic how humans visually parse text:
//   - Line-oriented, top-to-bottom scanning
//   - Minimal backtracking with lookahead capabilities
//   - Context-aware line tracking (file, line number, position)
//
// Reference: https://github.com/asciidoctor/asciidoctor/blob/main/lib/asciidoctor/reader.rb
package reader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var (
	// ErrNoMoreLines is returned when reading past the end of the source.
	ErrNoMoreLines = errors.New("no more lines")
	// ErrEmptySource is returned when the reader is initialized with empty content.
	ErrEmptySource = errors.New("empty source")
)

// Position represents a location in the source document.
// Similar to Asciidoctor's Cursor class.
type Position struct {
	File   string // File path or name
	Dir    string // Directory of the file
	Path   string // Path relative to base dir
	Lineno int    // 1-based line number
}

// String returns a human-readable position in the format "path:line".
func (p Position) String() string {
	if p.Path != "" {
		return fmt.Sprintf("%s:%d", p.Path, p.Lineno)
	}
	return fmt.Sprintf("line %d", p.Lineno)
}

// Clone creates a copy of the Position.
func (p Position) Clone() Position {
	return Position{
		File:   p.File,
		Dir:    p.Dir,
		Path:   p.Path,
		Lineno: p.Lineno,
	}
}

// Line represents a single line of AsciiDoc source with metadata.
// This corresponds to a processed line in Asciidoctor's Reader.
type Line struct {
	// Content is the actual line content (normalized: UTF-8, trailing whitespace stripped)
	Content string
	// Lineno is the 1-based line number in the source
	Lineno int
	// SourceLocation is the file/position information for error reporting
	SourceLocation Position
}

// String returns the line content.
func (l Line) String() string {
	return l.Content
}

// IsEmpty returns true if the line contains only whitespace.
func (l Line) IsEmpty() bool {
	return len(strings.TrimSpace(l.Content)) == 0
}

// Reader is a line-oriented reader for AsciiDoc source.
//
// It maintains a stack of lines (reversed for efficient pop operations)
// and tracks position information for error reporting. The design mirrors
// Asciidoctor's Reader class:
//   - Lines are stored in reverse order (so we can pop from end efficiently)
//   - Line numbers are 1-based (matching human convention and editors)
//   - Peek looks ahead without consuming
//   - Read/NextLine consumes and advances
//
// Example usage:
//
//	r := reader.NewReader("= Document Title\n\nSome content")
//	for r.HasMoreLines() {
//	    line, _ := r.NextLine()
//	    fmt.Println(line)
//	}
type Reader struct {
	// sourceLines is the original, normalized source lines
	sourceLines []string

	// lines is the internal stack of remaining lines (in reverse order)
	lines []string

	// lineno is the 1-based line number of the next line to be read
	lineno int

	// lookAhead counts how many lines have been peeked but not consumed
	lookAhead int

	// Position tracking for error reporting
	file string
	dir  string
	path string

	// mark stores a saved position for rollback
	mark *savedState

	// processLines controls whether processLine is called during peek
	processLines bool

	// unterminated tracks whether an unterminated block was detected
	unterminated bool
}

// savedState stores the state at a mark point.
type savedState struct {
	lines     []string
	lineno    int
	file      string
	dir       string
	path      string
	lookAhead int
}

// ReaderOption configures a Reader.
type ReaderOption func(*Reader)

// WithFile sets the file path metadata for the reader.
func WithFile(file string) ReaderOption {
	return func(r *Reader) {
		r.file = file
		// Split file into dir and basename
		idx := strings.LastIndex(file, "/")
		if idx == -1 {
			idx = strings.LastIndex(file, "\\")
		}
		if idx >= 0 {
			r.dir = file[:idx]
			r.path = file[idx+1:]
		} else {
			r.dir = "."
			r.path = file
		}
	}
}

// WithPath sets the path relative to base directory.
func WithPath(path string) ReaderOption {
	return func(r *Reader) {
		r.path = path
	}
}

// WithDir sets the directory for resolving relative paths.
func WithDir(dir string) ReaderOption {
	return func(r *Reader) {
		r.dir = dir
	}
}

// WithStartingLine sets the initial line number (default: 1).
func WithStartingLine(lineno int) ReaderOption {
	return func(r *Reader) {
		r.lineno = lineno
	}
}

// WithProcessLines controls whether lines are processed on first visit.
func WithProcessLines(process bool) ReaderOption {
	return func(r *Reader) {
		r.processLines = process
	}
}

// NewReader creates a new Reader from the given source string.
//
// The source is normalized (UTF-8 validated, trailing whitespace stripped)
// and split into lines.
func NewReader(source string, opts ...ReaderOption) (*Reader, error) {
	lines := prepareLines(source)
	if len(lines) == 0 {
		return nil, ErrEmptySource
	}

	r := &Reader{
		sourceLines:  make([]string, len(lines)),
		lines:        make([]string, 0, len(lines)),
		lineno:       1,
		lookAhead:    0,
		processLines: true,
		dir:          ".",
		path:         "",
	}
	copy(r.sourceLines, lines)

	// Reverse lines so we can efficiently pop from the end
	for i := len(lines) - 1; i >= 0; i-- {
		r.lines = append(r.lines, lines[i])
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// NewReaderFromLines creates a Reader from an already-split slice of lines.
// The lines are assumed to be already normalized.
func NewReaderFromLines(lines []string, opts ...ReaderOption) (*Reader, error) {
	if len(lines) == 0 {
		return nil, ErrEmptySource
	}

	r := &Reader{
		sourceLines:  make([]string, len(lines)),
		lines:        make([]string, 0, len(lines)),
		lineno:       1,
		lookAhead:    0,
		processLines: true,
		dir:          ".",
		path:         "",
	}
	copy(r.sourceLines, lines)

	// Reverse lines
	for i := len(lines) - 1; i >= 0; i-- {
		r.lines = append(r.lines, lines[i])
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// NewReaderFromReader creates a Reader from an io.Reader.
func NewReaderFromReader(rdr io.Reader, opts ...ReaderOption) (*Reader, error) {
	content, err := io.ReadAll(rdr)
	if err != nil {
		return nil, fmt.Errorf("reading from io.Reader: %w", err)
	}
	return NewReader(string(content), opts...)
}

// prepareLines normalizes and splits source content into lines.
//
// Normalization:
//   - Validates UTF-8 encoding
//   - Strips trailing whitespace from each line
//   - Removes trailing empty line from final newline (matches Asciidoctor)
//   - Preserves intentional empty lines
//   - Returns empty slice for empty source
//
// This mirrors Asciidoctor's Reader#prepare_lines with normalize: true.
func prepareLines(source string) []string {
	// Handle empty source
	if source == "" {
		return []string{}
	}

	// Split on newlines, preserving trailing empty lines
	// strings.Split("a\nb\nc\n", "\n") returns ["a", "b", "c", ""]
	lines := strings.Split(source, "\n")

	// Strip trailing whitespace from each line
	// Asciidoctor uses .rstrip in Ruby
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = strings.TrimRight(line, " \t\r\n")
	}

	// Remove trailing empty line if source ends with newline
	// This matches Asciidoctor's behavior where "a\nb\nc\n" → ["a", "b", "c"]
	if len(result) > 0 && result[len(result)-1] == "" && strings.HasSuffix(source, "\n") {
		result = result[:len(result)-1]
	}

	return result
}

// HasMoreLines returns true if there are more lines to read.
// Returns false if the reader is empty or only has blank lines remaining.
func (r *Reader) HasMoreLines() bool {
	for _, line := range r.lines {
		if !isEmpty(line) {
			return true
		}
	}
	return false
}

// IsEmpty returns true if the reader has no more lines.
func (r *Reader) IsEmpty() bool {
	return len(r.lines) == 0
}

// EOF is an alias for IsEmpty for familiarity.
func (r *Reader) EOF() bool {
	return len(r.lines) == 0
}

// PeekLine returns the next line without consuming it.
//
// The line is marked as "visited" (lookAhead incremented) to avoid
// reprocessing on subsequent calls. This mirrors Asciidoctor's
// Reader#peek_line behavior.
//
// Returns empty string if no more lines are available.
func (r *Reader) PeekLine() string {
	if len(r.lines) == 0 {
		return ""
	}

	line := r.lines[len(r.lines)-1]

	// Mark line as visited if not already
	if r.lookAhead == 0 {
		line = r.processLine(line)
	}

	return line
}

// PeekLines returns the next n lines without consuming them.
// If n is 0 or negative, returns all remaining lines.
//
// The lines are marked as visited.
func (r *Reader) PeekLines(n int) []string {
	if len(r.lines) == 0 {
		return []string{}
	}

	if n <= 0 {
		n = len(r.lines)
	}
	if n > len(r.lines) {
		n = len(r.lines)
	}

	result := make([]string, 0, n)
	for i := 0; i < n; i++ {
		idx := len(r.lines) - 1 - i
		if idx < 0 {
			break
		}
		line := r.lines[idx]

		// Process line if not already visited
		if r.lookAhead <= i {
			line = r.processLine(line)
			r.lookAhead++
		}

		result = append(result, line)
	}

	return result
}

// NextLine consumes and returns the next line.
//
// If the line was previously peeked, it returns the cached value.
// Otherwise, it processes and returns the line.
//
// Returns empty string if no more lines are available.
func (r *Reader) NextLine() string {
	if len(r.lines) == 0 {
		return ""
	}

	// If line was already peeked, lookAhead > 0
	if r.lookAhead > 0 {
		r.lookAhead--
	}

	r.lineno++
	line := r.lines[len(r.lines)-1]
	r.lines = r.lines[:len(r.lines)-1]

	return line
}

// ReadLine is an alias for NextLine.
func (r *Reader) ReadLine() string {
	return r.NextLine()
}

// ReadLines consumes and returns all remaining lines.
func (r *Reader) ReadLines() []string {
	result := make([]string, 0, len(r.lines))
	for r.HasMoreLines() {
		result = append(result, r.NextLine())
	}
	return result
}

// ReadString consumes and returns all remaining lines joined as a string.
func (r *Reader) ReadString() string {
	return strings.Join(r.ReadLines(), "\n")
}

// Advance consumes the next line without returning it.
// Returns true if a line was consumed, false if at EOF.
func (r *Reader) Advance() bool {
	if len(r.lines) == 0 {
		return false
	}

	if r.lookAhead > 0 {
		r.lookAhead--
	}

	r.lineno++
	r.lines = r.lines[:len(r.lines)-1]
	return true
}

// SkipBlankLines consumes consecutive blank (empty/whitespace-only) lines.
// Returns the number of lines skipped.
func (r *Reader) SkipBlankLines() int {
	skipped := 0
	for r.HasMoreLines() {
		nextLine := r.PeekLine()
		if !isEmpty(nextLine) {
			break
		}
		r.Advance()
		skipped++
	}
	return skipped
}

// SkipCommentLines consumes consecutive single-line comment lines.
// Returns the number of lines skipped.
func (r *Reader) SkipCommentLines() int {
	skipped := 0
	for r.HasMoreLines() {
		nextLine := r.PeekLine()
		if isEmpty(nextLine) {
			break
		}
		if !strings.HasPrefix(nextLine, "//") {
			break
		}
		// Don't skip block delimiters (///)
		if strings.HasPrefix(nextLine, "///") && !strings.HasPrefix(nextLine, "////") {
			// This might be a block delimiter, check if it's all slashes
			if isAllSlashes(nextLine) {
				break
			}
		}
			r.Advance()
		skipped++
	}
	return skipped
}

// UnshiftLine pushes a line back onto the front of the reader.
// The line is marked as already processed (lookAhead incremented).
//
// This is typically used to "put back" a line that was read but
// shouldn't have been consumed.
func (r *Reader) UnshiftLine(line string) {
	r.lines = append(r.lines, line)
	r.lineno--
	r.lookAhead++
}

// RestoreLine is an alias for UnshiftLine.
func (r *Reader) RestoreLine(line string) {
	r.UnshiftLine(line)
}

// UnshiftLines pushes multiple lines back onto the reader.
// The lines are pushed in order (first element will be read first).
func (r *Reader) UnshiftLines(lines []string) {
	for i := len(lines) - 1; i >= 0; i-- {
		r.lines = append(r.lines, lines[i])
		r.lineno--
		r.lookAhead++
	}
}

// ReplaceNextLine replaces the next line with the given string.
// The current next line is consumed first.
func (r *Reader) ReplaceNextLine(replacement string) {
	if r.HasMoreLines() {
		r.NextLine()
	}
	r.UnshiftLine(replacement)
}

// ReadLinesUntil reads lines until a termination condition is met.
//
// The options control behavior:
//   - Terminator: stop when this exact line is found
//   - BreakOnBlankLines: stop on empty/whitespace-only lines
//   - BreakOnListContinuation: stop on list continuation (+)
//   - SkipFirstLine: consume first line before starting
//   - PreserveLastLine: put the terminating line back on the stack
//   - ReadLastLine: include the terminating line in results
//
// Returns the collected lines (not including terminator unless ReadLastLine).
func (r *Reader) ReadLinesUntil(opts ...ReadUntilOption) []string {
	cfg := &readUntilConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	result := []string{}

	if cfg.SkipFirstLine && r.HasMoreLines() {
		r.NextLine()
	}

	for r.HasMoreLines() {
		line := r.PeekLine()

		// Check termination conditions
		shouldBreak := false
		includeLastLine := false

		if cfg.Terminator != "" && line == cfg.Terminator {
			if cfg.ReadLastLine {
				includeLastLine = true
			}
			shouldBreak = true
		} else if cfg.BreakOnBlankLines && isEmpty(line) {
			shouldBreak = true
		} else if cfg.BreakOnListContinuation && line == "+" && len(result) > 0 {
			if cfg.ReadLastLine {
				includeLastLine = true
			}
			shouldBreak = true
		} else if cfg.Cond != nil && cfg.Cond(line) {
			if cfg.ReadLastLine {
				includeLastLine = true
			}
			shouldBreak = true
		}

		if shouldBreak {
			if cfg.PreserveLastLine {
				// Don't consume the line that caused us to stop
			} else if includeLastLine {
				result = append(result, r.NextLine())
			}
			break
		}

		result = append(result, r.NextLine())
	}

	return result
}

// ReadUntilConfig options for ReadLinesUntil.
type ReadUntilOption func(*readUntilConfig)

type readUntilConfig struct {
	Terminator              string
	BreakOnBlankLines       bool
	BreakOnListContinuation bool
	SkipFirstLine           bool
	PreserveLastLine        bool
	ReadLastLine            bool
	Cond                    func(string) bool
}

// WithTerminator sets the terminator line string.
func WithTerminator(term string) ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.Terminator = term
	}
}

// WithBreakOnBlankLines enables stopping on blank lines.
func WithBreakOnBlankLines() ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.BreakOnBlankLines = true
	}
}

// WithBreakOnListContinuation enables stopping on list continuation (+).
func WithBreakOnListContinuation() ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.BreakOnListContinuation = true
	}
}

// WithSkipFirstLine consumes the first line before starting.
func WithSkipFirstLine() ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.SkipFirstLine = true
	}
}

// WithPreserveLastLine keeps the terminating line on the stack.
func WithPreserveLastLine() ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.PreserveLastLine = true
	}
}

// WithReadLastLine includes the terminating line in results.
func WithReadLastLine() ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.ReadLastLine = true
	}
}

// WithCondition sets a custom predicate function for termination.
func WithCondition(cond func(string) bool) ReadUntilOption {
	return func(cfg *readUntilConfig) {
		cfg.Cond = cond
	}
}

// GetPosition returns the current position in the source.
func (r *Reader) GetPosition() Position {
	return Position{
		File:   r.file,
		Dir:    r.dir,
		Path:   r.path,
		Lineno: r.lineno,
	}
}

// GetLineno returns the current 1-based line number.
func (r *Reader) GetLineno() int {
	return r.lineno
}

// GetLines returns a copy of remaining lines (in original order).
func (r *Reader) GetLines() []string {
	result := make([]string, len(r.lines))
	for i := 0; i < len(r.lines); i++ {
		result[i] = r.lines[len(r.lines)-1-i]
	}
	return result
}

// GetString returns remaining lines joined as a string.
func (r *Reader) GetString() string {
	return strings.Join(r.GetLines(), "\n")
}

// GetSource returns the original source lines joined as a string.
func (r *Reader) GetSource() string {
	return strings.Join(r.sourceLines, "\n")
}

// Dir returns the directory path for the reader.
func (r *Reader) Dir() string {
	return r.dir
}

// SetDir sets the directory path for the reader.
func (r *Reader) SetDir(dir string) {
	r.dir = dir
}

// InjectLines adds new lines at the current position in the reader.
//
// The lines are inserted such that they will be read before any remaining
// lines in the reader. This is used for include directive processing.
func (r *Reader) InjectLines(newLines []string) {
	// Reverse the new lines and append to the end
	// Since reader pops from the end, this makes injected lines read first
	for i := len(newLines) - 1; i >= 0; i-- {
		line := newLines[i]
		// Strip trailing whitespace
		line = strings.TrimRight(line, " \t")
		// Add to the end of current lines
		r.lines = append(r.lines, line)
	}
}

// Mark saves the current reader state for potential rollback.
func (r *Reader) Mark() {
	r.mark = &savedState{
		lines:     make([]string, len(r.lines)),
		lineno:    r.lineno,
		file:      r.file,
		dir:       r.dir,
		path:      r.path,
		lookAhead: r.lookAhead,
	}
	copy(r.mark.lines, r.lines)
}

// Unmark clears the saved mark state.
func (r *Reader) Unmark() {
	r.mark = nil
}

// Restore restores the reader to the previously marked state.
// Returns true if there was a mark to restore, false otherwise.
func (r *Reader) Restore() bool {
	if r.mark == nil {
		return false
	}

	r.lines = make([]string, len(r.mark.lines))
	copy(r.lines, r.mark.lines)
	r.lineno = r.mark.lineno
	r.file = r.mark.file
	r.dir = r.mark.dir
	r.path = r.mark.path
	r.lookAhead = r.mark.lookAhead

	r.mark = nil
	return true
}

// Cursor returns the current cursor/position information.
// In Asciidoctor, this creates a new Cursor object.
func (r *Reader) Cursor() Position {
	return Position{
		File:   r.file,
		Dir:    r.dir,
		Path:   r.path,
		Lineno: r.lineno,
	}
}

// CursorAtMark returns the position at the last mark.
func (r *Reader) CursorAtMark() Position {
	if r.mark != nil {
		return Position{
			File:   r.mark.file,
			Dir:    r.mark.dir,
			Path:   r.mark.path,
			Lineno: r.mark.lineno,
		}
	}
	return r.Cursor()
}

// LineInfo returns a human-readable position string.
func (r *Reader) LineInfo() string {
	pos := r.GetPosition()
	return pos.String()
}

// SetUnterminated marks that an unterminated block was detected.
func (r *Reader) SetUnterminated() {
	r.unterminated = true
}

// IsUnterminated returns whether an unterminated block was detected.
func (r *Reader) IsUnterminated() bool {
	return r.unterminated
}

// processLine is called the first time a line is visited.
//
// In the base Reader, this just marks the line as visited.
// Subclasses (like PreprocessorReader) override this to handle
// preprocessor directives like includes and conditionals.
func (r *Reader) processLine(line string) string {
	if r.processLines {
		r.lookAhead++
	}
	return line
}

// Helper functions

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func isAllSlashes(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return false
	}
	for _, r := range trimmed {
		if r != '/' {
			return false
		}
	}
	return true
}

// Scanner wraps bufio.Scanner for convenient line-by-line reading.
type Scanner struct {
	scanner *bufio.Scanner
	lineno  int
	file    string
	path    string
}

// NewScanner creates a new Scanner from an io.Reader.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		scanner: bufio.NewScanner(r),
		lineno:  0,
	}
}

// NewScannerFile creates a Scanner with file metadata.
func NewScannerFile(r io.Reader, file, path string) *Scanner {
	return &Scanner{
		scanner: bufio.NewScanner(r),
		lineno:  0,
		file:    file,
		path:    path,
	}
}

// Scan advances to the next line.
func (s *Scanner) Scan() bool {
	ok := s.scanner.Scan()
	if ok {
		s.lineno++
	}
	return ok
}

// Bytes returns the current line as bytes.
func (s *Scanner) Bytes() []byte {
	return s.scanner.Bytes()
}

// Text returns the current line as a string.
func (s *Scanner) Text() string {
	return s.scanner.Text()
}

// Err returns any error encountered during scanning.
func (s *Scanner) Err() error {
	return s.scanner.Err()
}

// Lineno returns the current 1-based line number.
func (s *Scanner) Lineno() int {
	return s.lineno
}

// File returns the file path.
func (s *Scanner) File() string {
	return s.file
}

// Position returns the current Position.
func (s *Scanner) Position() Position {
	return Position{
		File:   s.file,
		Path:   s.path,
		Lineno: s.lineno,
	}
}

// ReadAsciidocLines reads all lines from a Scanner into a normalized slice.
func ReadAsciidocLines(s *bufio.Scanner) ([]string, error) {
	var lines []string
	lineno := 0

	for s.Scan() {
		lineno++
		line := strings.TrimRight(s.Text(), " \t\r\n")

		// Validate UTF-8 incrementally
		if !utf8.ValidString(line) {
			return nil, fmt.Errorf("invalid UTF-8 at line %d", lineno)
		}

		lines = append(lines, line)
	}

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scanning: %w", err)
	}

	return lines, nil
}

// SkipFrontMatter skips YAML front matter if present.
// Returns the skipped lines (or nil if none found).
func SkipFrontMatter(lines []string) ([]string, []string) {
	if len(lines) == 0 {
		return nil, lines
	}

	// Check for front matter delimiter
	delim := lines[0]
	if delim != "---" && delim != "+++" {
		return nil, lines
	}

	if len(lines) < 2 {
		return nil, lines
	}

	// Find closing delimiter
	endIdx := 1
	for endIdx < len(lines) && lines[endIdx] != delim {
		endIdx++
	}

	if endIdx >= len(lines) {
		// Unterminated front matter - treat as normal content
		return nil, lines
	}

	// Extract and skip front matter
	frontMatter := lines[1:endIdx]
	remaining := lines[endIdx+1:]

	return frontMatter, remaining
}

// CountLeadingEmptyLines counts consecutive empty lines from the start.
func CountLeadingEmptyLines(lines []string) int {
	count := 0
	for _, line := range lines {
		if isEmpty(line) {
			count++
		} else {
			break
		}
	}
	return count
}

// CountTrailingEmptyLines counts consecutive empty lines from the end.
func CountTrailingEmptyLines(lines []string) int {
	count := 0
	for i := len(lines) - 1; i >= 0; i-- {
		if isEmpty(lines[i]) {
			count++
		} else {
			break
		}
	}
	return count
}

// TrimLeadingEmptyLines removes leading empty lines.
func TrimLeadingEmptyLines(lines []string) []string {
	count := CountLeadingEmptyLines(lines)
	if count == 0 {
		return lines
	}
	return lines[count:]
}

// TrimTrailingEmptyLines removes trailing empty lines.
func TrimTrailingEmptyLines(lines []string) []string {
	count := CountTrailingEmptyLines(lines)
	if count == 0 {
		return lines
	}
	return lines[:len(lines)-count]
}
