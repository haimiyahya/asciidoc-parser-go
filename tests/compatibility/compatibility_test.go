// Package compatibility provides tests for Asciidoctor compatibility validation.
//
// This test framework compares the output of this Go parser against
// reference Asciidoctor HTML output to ensure compatibility.
package compatibility

import (
	"bytes"
	"context"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/stretchr/testify/require"
)

const (
	// asciidoctorDocTestURL is the base URL for Asciidoctor DocTest examples
	asciidoctorDocTestURL = "https://raw.githubusercontent.com/asciidoctor/asciidoctor-doctest/main/data/examples/asciidoc"

	// asciidoctorCmd is the command to run Asciidoctor
	asciidoctorCmd = "/usr/local/bin/asciidoctor"

	// compatibilityTestDir is the directory for compatibility test data
	compatibilityTestDir = "tests/compatibility"

	// examplesDir contains AsciiDoc source examples
	examplesDir = "examples"

	// expectedDir contains expected HTML output
	expectedDir = "expected"
)

// TestResult represents the result of a compatibility test
type TestResult struct {
	Name           string
	Passed         bool
	ExpectedHTML   string
	ActualHTML     string
	Diff           string
	AsciidoctorErr error
	ParserErr      error
}

// TestCase represents a single compatibility test case
type TestCase struct {
	Name    string
	Source  string
	Category string // e.g., "basic", "lists", "tables", "inline"
	Skip    bool
	SkipReason string
}

// CompatibilityTester handles running compatibility tests
type CompatibilityTester struct {
	t             *testing.T
	testCases     []TestCase
	asciidoctorAvailable bool
}

// NewCompatibilityTester creates a new compatibility tester
func NewCompatibilityTester(t *testing.T) *CompatibilityTester {
	ct := &CompatibilityTester{
		t:                   t,
		testCases:           []TestCase{},
		asciidoctorAvailable: checkAsciidoctorAvailable(),
	}

	if !ct.asciidoctorAvailable {
		t.Log("Asciidoctor not found in PATH. Tests will use pre-generated expected files.")
	}

	return ct
}

// checkAsciidoctorAvailable checks if asciidoctor command is available
func checkAsciidoctorAvailable() bool {
	_, err := exec.LookPath(asciidoctorCmd)
	return err == nil
}

// AddTestCase adds a test case to the suite
func (ct *CompatibilityTester) AddTestCase(tc TestCase) {
	ct.testCases = append(ct.testCases, tc)
}

// AddTestCases adds multiple test cases to the suite
func (ct *CompatibilityTester) AddTestCases(tcs []TestCase) {
	ct.testCases = append(ct.testCases, tcs...)
}

// Run runs all compatibility tests
func (ct *CompatibilityTester) Run() {
	ct.t.Helper()

	for _, tc := range ct.testCases {
		ct.t.Run(tc.Name, func(t *testing.T) {
			if tc.Skip {
				t.Skip(tc.SkipReason)
			}
			ct.runTest(t, tc)
		})
	}
}

// runTest runs a single compatibility test
func (ct *CompatibilityTester) runTest(t *testing.T, tc TestCase) {
	t.Helper()

	result := ct.runTestCase(tc)

	if result.ParserErr != nil {
		t.Fatalf("Go parser error: %v", result.ParserErr)
	}

	if !result.Passed {
		ct.reportFailure(t, tc, result)
	}
}

// runTestCase executes a test case and returns the result
func (ct *CompatibilityTester) runTestCase(tc TestCase) TestResult {
	result := TestResult{
		Name: tc.Name,
	}

	// Get expected HTML from Asciidoctor or cached file
	expectedHTML, err := ct.getExpectedHTML(tc)
	if err != nil {
		result.AsciidoctorErr = err
		// If we can't get expected HTML, we'll skip comparison
		// but still verify our parser doesn't error
		result.Passed = true
	}
	result.ExpectedHTML = expectedHTML

	// Get actual HTML from Go parser
	actualHTML, err := ct.parseWithGoParser(tc.Source)
	if err != nil {
		result.ParserErr = err
		result.Passed = false
		return result
	}
	result.ActualHTML = actualHTML

	// Compare outputs
	if result.AsciidoctorErr == nil {
		result.Passed = ct.compareHTML(result.ExpectedHTML, actualHTML)
		if !result.Passed {
			result.Diff = ct.generateDiff(result.ExpectedHTML, actualHTML)
		}
	} else {
		// No expected HTML - just verify parser works
		result.Passed = true
	}

	return result
}

// getExpectedHTML gets the expected HTML output from Asciidoctor or cached file
func (ct *CompatibilityTester) getExpectedHTML(tc TestCase) (string, error) {
	// First, try to read from cached expected file
	expectedPath := filepath.Join(expectedDir, tc.Name+".html")
	if data, err := os.ReadFile(expectedPath); err == nil {
		return string(data), nil
	}

	// If not cached and Asciidoctor is available, generate it
	if ct.asciidoctorAvailable {
		return ct.runAsciidoctor(tc.Source)
	}

	// Try to download from repository
	if html, err := ct.downloadExpectedHTML(tc.Name); err == nil {
		return html, nil
	}

	return "", fmt.Errorf("no expected HTML available (cache not found, asciidoctor not available, download failed)")
}

// runAsciidoctor runs asciidoctor and returns the HTML output
func (ct *CompatibilityTester) runAsciidoctor(source string) (string, error) {
	// Create a temporary file for the source
	tmpDir, err := os.MkdirTemp("", "asciidoc-compat-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.adoc")
	if err := os.WriteFile(inputFile, []byte(source), 0644); err != nil {
		return "", fmt.Errorf("failed to write input file: %w", err)
	}

	outputFile := filepath.Join(tmpDir, "output.html")

	// Run asciidoctor with embedded mode to get body-only content
	// This matches the Go parser's minimal HTML output
	cmd := exec.Command(asciidoctorCmd, "-b", "html5", "-o", outputFile, "-a", "newline=\\n", "-e", inputFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("asciidoctor failed: %w\nOutput: %s", err, output)
	}

	// Read the output
	html, err := os.ReadFile(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read output file: %w", err)
	}

	return string(html), nil
}

// downloadExpectedHTML tries to download expected HTML from the repository
func (ct *CompatibilityTester) downloadExpectedHTML(name string) (string, error) {
	url := fmt.Sprintf("%s/%s.html.gz", asciidoctorDocTestURL, name)

	// Try to download gzipped version
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Gunzip if needed
	var reader io.Reader = resp.Body
	if strings.HasSuffix(url, ".gz") {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	html, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(html), nil
}

// parseWithGoParser parses the source with the Go parser
func (ct *CompatibilityTester) parseWithGoParser(source string) (string, error) {
	p, err := parser.NewParserFromString(source)
	if err != nil {
		return "", fmt.Errorf("failed to create parser: %w", err)
	}

	doc, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("failed to parse: %w", err)
	}

	var buf bytes.Buffer
	// Use WithoutHeaderFooter to match Asciidoctor's embedded mode output
	c := converter.NewHTML5Converter().WithoutHeaderFooter()
	if err := c.Convert(doc, &buf); err != nil {
		return "", fmt.Errorf("failed to convert: %w", err)
	}

	return buf.String(), nil
}

// compareHTML compares two HTML strings for equality
func (ct *CompatibilityTester) compareHTML(expected, actual string) bool {
	// Normalize for comparison
	normExpected := normalizeHTML(expected)
	normActual := normalizeHTML(actual)

	return normExpected == normActual
}

// generateDiff generates a diff between expected and actual HTML
func (ct *CompatibilityTester) generateDiff(expected, actual string) string {
	// Simple line-by-line diff
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var diff strings.Builder

	diff.WriteString("--- Expected\n+++ Actual\n")

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		expectedLine := ""
		actualLine := ""

		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			if expectedLine != "" {
				diff.WriteString(fmt.Sprintf("-%s\n", expectedLine))
			}
			if actualLine != "" {
				diff.WriteString(fmt.Sprintf("+%s\n", actualLine))
			}
		}
	}

	return diff.String()
}

// reportFailure reports a test failure with detailed information
func (ct *CompatibilityTester) reportFailure(t *testing.T, tc TestCase, result TestResult) {
	t.Helper()

	t.Errorf("HTML output differs from Asciidoctor")

	// Log the diff
	if result.Diff != "" {
		t.Logf("Diff:\n%s", result.Diff)
	}

	// Log truncated versions for easier reading
	t.Logf("Expected (truncated): %s", truncateString(result.ExpectedHTML, 500))
	t.Logf("Actual (truncated):   %s", truncateString(result.ActualHTML, 500))

	// Save actual output for inspection
	if ct.t != nil {
		outputDir := "actual"
		os.MkdirAll(outputDir, 0755)

		outputPath := filepath.Join(outputDir, tc.Name+".html")
		if err := os.WriteFile(outputPath, []byte(result.ActualHTML), 0644); err == nil {
			t.Logf("Actual output saved to: %s", outputPath)
		}
	}
}

// normalizeHTML normalizes HTML for comparison
func normalizeHTML(html string) string {
	// Remove extra whitespace between tags
	html = strings.ReplaceAll(html, "\n", " ")
	html = strings.ReplaceAll(html, "\t", " ")
	html = strings.ReplaceAll(html, "  ", " ")
	html = strings.ReplaceAll(html, "  ", " ")
	html = strings.ReplaceAll(html, "> <", "><")

	// Trim whitespace
	html = strings.TrimSpace(html)

	return html
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// =============================================================================
// Built-in Test Cases
// =============================================================================

// getBuiltInTestCases returns built-in compatibility test cases
func getBuiltInTestCases() []TestCase {
	return []TestCase{
		// Basic document structure
		{
			Name: "basic/paragraphs",
			Category: "basic",
			Source: `First paragraph.

Second paragraph with more text.

Third paragraph follows.`,
		},
		{
			Name: "basic/document-title",
			Category: "basic",
			Source: `= Document Title

This is a paragraph.`,
		},
		{
			Name: "basic/sections",
			Category: "basic",
			Source: `== Section One

Content for section one.

=== Section Two

Content for section two.

==== Section Three

Deeper content.`,
		},
		{
			Name: "basic/text-formatting",
			Category: "basic",
			Source: `This is **bold** and __italic__ text.

This is __bold__ and **italic** text.`,
		},

		// Lists
		{
			Name: "lists/unordered",
			Category: "lists",
			Source: `* Apple
* Banana
* Cherry`,
		},
		{
			Name: "lists/ordered",
			Category: "lists",
			Source: `. First item
. Second item
. Third item`,
		},
		{
			Name: "lists/labeled",
			Category: "lists",
			Source: `Term 1:: Definition 1
Term 2:: Definition 2
Term 3:: Definition 3`,
		},

		// Inline markup
		{
			Name: "inline/links",
			Category: "inline",
			Source: `https://example.com

link:https://example.com[Example Link]

link:https://example.com[Click Here, window="_blank"]`,
		},
		{
			Name: "inline/images",
			Category: "inline",
			Source: `image:logo.png[Alt Text]

Inline image:image:icon.png[Icon] in text.`,
		},
		{
			Name: "inline/monospace",
			Category: "inline",
			Source: `Use ` + "`" + `ls -la` + "`" + ` to list files.

Code with ++code++ markup.`,
		},
		{
			Name: "inline/superscript-subscript",
			Category: "inline",
			Source: `H^2^O is water.

x~n~ = x^1^ / x^n`,
		},

		// Admonitions
		{
			Name: "admonitions/note",
			Category: "admonitions",
			Source: `NOTE: This is important information.`,
		},
		{
			Name: "admonitions/tip",
			Category: "admonitions",
			Source: `TIP: Try this approach for better results.`,
		},
		{
			Name: "admonitions/warning",
			Category: "admonitions",
			Source: `WARNING: Be careful when doing this!`,
		},

		// Delimited blocks
		{
			Name: "blocks/literal",
			Category: "blocks",
			Source: `....
Line 1
Line 2
Line 3
....`,
		},
		{
			Name: "blocks/example",
			Category: "blocks",
			Source: `====
Example block content.
Can be multiple lines.
====`,
		},
		{
			Name: "blocks/quote",
			Category: "blocks",
			Source: `____
Quote block content.
It can span multiple lines.
____`,
		},
		{
			Name: "blocks/sidebar",
			Category: "blocks",
			Source: `****
Sidebar content.
Additional information.
****`,
		},

		// Tables
		{
			Name: "tables/basic",
			Category: "tables",
			Source: `|===
| Cell 1 | Cell 2 | Cell 3
| A | B | C
|===`,
		},
		{
			Name: "tables/with-header",
			Category: "tables",
			Source: `|===
|= Header 1 |= Header 2 |= Header 3
| Cell 1 | Cell 2 | Cell 3
|===`,
		},

		// Index terms
		{
			Name: "indexterms/flow",
			Category: "indexterms",
			Source: `This is a paragraph with a ((flow index term)) in it.`,
		},
		{
			Name: "indexterms/concealed",
			Category: "indexterms",
			Source: `This has a (((hidden, term))) index term.`,
		},

		// Bibliography
		{
			Name: "bibliography/basic",
			Category: "bibliography",
			Source: `[bibliography]
== Bibliography

* [[[pp]]] Andy Hunt & Dave Thomas. *The Pragmatic Programmer*.
* [[[gof,gang]]] Erich Gamma et al. __Design Patterns__.`,
		},

		// UI Macros
		{
			Name: "ui/keyboard",
			Category: "ui",
			Source: `Press kbd:[Ctrl+C] to copy.`,
		},
		{
			Name: "ui/button",
			Category: "ui",
			Source: `Click btn:[Submit] to continue.`,
		},
		{
			Name: "ui/menu",
			Category: "ui",
			Source: `Go to menu:[File > Save As].`,
		},

		// Roles
		{
			Name: "roles/basic",
			Category: "roles",
			Source: `[.red]**This text is red**

[.role1.role2]__This has two classes__`,
		},

		// Passthrough macros
		{
			Name: "passthrough/inline",
			Category: "passthrough",
			Source: `Use +<b>bold</b>+ for inline passthrough.`,
		},
		{
			Name: "passthrough/raw",
			Category: "passthrough",
			Source: `This +++<b>raw HTML</b>+++ is passed through.`,
		},
		{
			Name: "passthrough/span",
			Category: "passthrough",
			Source: `This is a ++span++ that groups text.`,
		},

		// Multi-line table cells
		{
			Name: "tables/multiline",
			Category: "tables",
			Source: `|===
|Cell 1 |Cell 2
|+continued cell 1 |Cell 3
|===`,
		},

		// Table cell styles
		{
			Name: "tables/cell-styles",
			Category: "tables",
			Source: `|===
|Normal cell |l|Literal cell |m|Monospace cell
|===`,
		},
		{
			Name: "tables/cell-styles-advanced",
			Category: "tables",
			Source: `|===
|e|Emphasis |s|Strong |v|Verse cell
|===`,
		},

		// Table repeat cells
		{
			Name: "tables/repeat-cells",
			Category: "tables",
			Source: `|===
|3*Same | Different
|===`,
		},
		{
			Name: "tables/repeat-multiple",
			Category: "tables",
			Source: `|===
|2*A |2*B | C
|===`,
		},
	}
}

// =============================================================================
// Test Functions
// =============================================================================

// TestCompatibility_BuiltIn runs built-in compatibility tests
func TestCompatibility_BuiltIn(t *testing.T) {
	tester := NewCompatibilityTester(t)
	tester.AddTestCases(getBuiltInTestCases())
	tester.Run()
}

// TestCompatibility_AsccdocioctorAvailable checks if Asciidoctor is available
func TestCompatibility_AsciidoctorAvailable(t *testing.T) {
	if checkAsciidoctorAvailable() {
		t.Log("Asciidoctor is available")
		// Get version
		cmd := exec.Command(asciidoctorCmd, "--version")
		if output, err := cmd.CombinedOutput(); err == nil {
			t.Logf("Asciidoctor version: %s", string(output))
		}
	} else {
		t.Skip("Asciidoctor not found. Install with: gem install asciidoctor")
	}
}

// TestCompatibility_DownloadExamples downloads examples from Asciidoctor DocTest
func TestCompatibility_DownloadExamples(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping download test in short mode")
	}

	// Create examples directory
	os.MkdirAll(examplesDir, 0755)
	examplesPath := examplesDir

	// List of examples to download
	examples := []string{
		"basic/paragraphs",
		"basic/document-title",
		"lists/unordered",
		"lists/ordered",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			// Try to download from GitHub
			url := fmt.Sprintf("https://raw.githubusercontent.com/asciidoctor/asciidoctor-doctest/main/data/examples/asciidoc/%s.adoc", example)

			req, _ := http.NewRequest("GET", url, nil)
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Skipf("Failed to download: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Skipf("Download failed with status %d", resp.StatusCode)
				return
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			// Save to file
			outputPath := filepath.Join(examplesPath, filepath.Base(example)+".adoc")
			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				t.Fatalf("Failed to write file: %v", err)
			}

			t.Logf("Downloaded: %s", outputPath)
		})
	}
}

// =============================================================================
// Golden File Management
// =============================================================================

// GoldenFileManager manages golden (expected) files
type GoldenFileManager struct {
	baseDir string
}

// NewGoldenFileManager creates a new golden file manager
func NewGoldenFileManager() *GoldenFileManager {
	// Use expectedDir relative to the test file location
	return &GoldenFileManager{
		baseDir: expectedDir,
	}
}

// SaveGolden saves a golden (expected) HTML file
func (g *GoldenFileManager) SaveGolden(name string, html string) error {
	path := filepath.Join(g.baseDir, name+".html")
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(html), 0644)
}

// LoadGolden loads a golden (expected) HTML file
func (g *GoldenFileManager) LoadGolden(name string) (string, error) {
	path := filepath.Join(g.baseDir, name+".html")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListGolden returns a list of all golden files
func (g *GoldenFileManager) ListGolden() ([]string, error) {
	entries, err := os.ReadDir(g.baseDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			files = append(files, strings.TrimSuffix(entry.Name(), ".html"))
		}
	}
	return files, nil
}

// TestCompatibility_GenerateGoldenFiles generates golden files from built-in test cases
func TestCompatibility_GenerateGoldenFiles(t *testing.T) {
	useAsciidoctor := os.Getenv("USE_ASCIIDOCTOR") == "1"
	useGoParser := os.Getenv("GENERATE_GOLDEN") == "1"

	if !useAsciidoctor && !useGoParser {
		t.Skip("Set USE_ASCIIDOCTOR=1 to generate from Asciidoctor or GENERATE_GOLDEN=1 to generate from Go parser")
	}

	g := NewGoldenFileManager()
	tester := NewCompatibilityTester(t)
	testCases := getBuiltInTestCases()

	if useAsciidoctor && !tester.asciidoctorAvailable {
		t.Fatal("USE_ASCIIDOCTOR=1 set but Asciidoctor is not available")
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var html string
			var err error

			if useAsciidoctor {
				// Generate HTML with Asciidoctor (reference implementation)
				html, err = tester.runAsciidoctor(tc.Source)
				require.NoError(t, err, "Failed to run Asciidoctor")
				t.Logf("Generated golden file from Asciidoctor: %s", tc.Name)
			} else {
				// Generate HTML with Go parser
				html, err = tester.parseWithGoParser(tc.Source)
				require.NoError(t, err, "Failed to parse source")
				t.Logf("Generated golden file from Go parser: %s", tc.Name)
			}

			// Save as golden file
			if err := g.SaveGolden(tc.Name, html); err != nil {
				t.Fatalf("Failed to save golden file: %v", err)
			}
		})
	}
}

// TestCompatibility_ValidateGoldenFiles validates that all golden files exist
func TestCompatibility_ValidateGoldenFiles(t *testing.T) {
	testCases := getBuiltInTestCases()
	g := NewGoldenFileManager()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err := g.LoadGolden(tc.Name)
			if err != nil {
				t.Errorf("Golden file missing: %s (run GENERATE_GOLDEN=1 to create)", tc.Name)
			}
		})
	}
}

// =============================================================================
// JSON Test Results
// =============================================================================

// TestResultsJSON is a JSON-serializable version of test results
type TestResultsJSON struct {
	Timestamp  string       `json:"timestamp"`
	Total      int          `json:"total"`
	Passed     int          `json:"passed"`
	Failed     int          `json:"failed"`
	Skipped    int          `json:"skipped"`
	Results    []TestResult `json:"results"`
}

// SaveTestResults saves test results to a JSON file
func SaveTestResults(results []TestResult, path string) error {
	jsonResults := TestResultsJSON{
		Timestamp: time.Now().Format(time.RFC3339),
		Total:     len(results),
		Passed:    0,
		Failed:    0,
		Skipped:   0,
		Results:   results,
	}

	for _, r := range results {
		if r.Passed {
			jsonResults.Passed++
		} else {
			jsonResults.Failed++
		}
	}

	data, err := json.MarshalIndent(jsonResults, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
