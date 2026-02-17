// Package main provides the CLI for the asciidoc parser.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/haimiyahya/asciidoc-parser-go/internal/ast"
	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/haimiyahya/asciidoc-parser-go/internal/processor"
	flag "github.com/spf13/pflag"
)

const (
	version = "0.2.0"
)

// Verbosity levels
const (
	verbosityQuiet   = 0
	verbosityNormal  = 1
	verbosityVerbose = 2
)

func main() {
	// Define command-line flags
	var (
		outputFile       string
		backend          string
		docType          string
		attributeDefs    []string
		baseDir          string
		sourceDir        string
		destDir          string
		noHeaderFooter   bool
		embedded         bool
		sectionNumbers   bool
		quiet            bool
		verbose          bool
		warnings         bool
		timings          bool
		showVersion      bool
		help             bool
		// PDF-specific options
		pdfPageSize       string
		pdfTOC            bool
		pdfCoverPage      bool
		pdfNoPageNumbers  bool
		pdfHeaderTemplate string
		pdfFooterTemplate string
	)

	// Define flags - compatible with Asciidoctor CLI
	flag.StringVarP(&outputFile, "out-file", "o", "", "Output file (default: based on input file; use - for STDOUT)")
	flag.StringVarP(&backend, "backend", "b", "html5", "Set backend output format: [html5, pdf, docbook, manpage, epub] (default: html5)")
	flag.StringVarP(&docType, "doctype", "d", "article", "Document type: [article, book, manpage, inline] (default: article)")
	flag.StringSliceVarP(&attributeDefs, "attribute", "a", nil, "Define a document attribute: name, name!, or name=value (may be specified multiple times)")
	flag.StringVarP(&baseDir, "base-dir", "B", "", "Base directory containing the document and resources (default: directory of input file)")
	flag.StringVarP(&sourceDir, "source-dir", "R", "", "Source root directory (used for calculating path in destination directory)")
	flag.StringVarP(&destDir, "destination-dir", "D", "", "Destination output directory (default: directory of input file)")
	flag.BoolVarP(&embedded, "embedded", "e", false, "Suppress enclosing document structure and output an embedded document")
	flag.BoolVarP(&noHeaderFooter, "no-header-footer", "s", false, "Suppress enclosing document structure (same as -e)")
	flag.BoolVarP(&sectionNumbers, "section-numbers", "n", false, "Auto-number section titles in the HTML backend")
	flag.BoolVarP(&quiet, "quiet", "q", false, "Silence application log messages (default: false)")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output (show all log messages)")
	flag.BoolVarP(&warnings, "warnings", "w", false, "Turn on script warnings")
	flag.BoolVarP(&timings, "timings", "t", false, "Print timings report")
	flag.BoolVarP(&showVersion, "version", "V", false, "Display version information")
	flag.BoolVarP(&help, "help", "h", false, "Show this help message")

	// PDF-specific flags
	flag.StringVar(&pdfPageSize, "pdf-page-size", "letter", "PDF page size: [letter, a4] (default: letter)")
	flag.BoolVar(&pdfTOC, "pdf-toc", false, "Generate table of contents in PDF")
	flag.BoolVar(&pdfCoverPage, "pdf-cover-page", false, "Generate cover page in PDF")
	flag.BoolVar(&pdfNoPageNumbers, "pdf-no-page-numbers", false, "Disable page numbers in PDF")
	flag.StringVar(&pdfHeaderTemplate, "pdf-header", "", "PDF header template")
	flag.StringVar(&pdfFooterTemplate, "pdf-footer", "", "PDF footer template")

	flag.Parse()

	// Handle help
	if help {
		printHelp()
		os.Exit(0)
	}

	// Handle version
	if showVersion {
		printVersion()
		os.Exit(0)
	}

	// Determine verbosity level
	verbosity := verbosityNormal
	if quiet {
		verbosity = verbosityQuiet
	} else if verbose {
		verbosity = verbosityVerbose
	}

	// Handle both -e and -s flags for embedded mode (they do the same thing)
	embeddedMode := embedded || noHeaderFooter

	// Start timing if requested
	var startTime time.Time
	if timings {
		startTime = time.Now()
	}

	// Get input file or stdin
	args := flag.Args()
	var inputPath string
	if len(args) > 0 {
		inputPath = args[0]
	}

	// Determine input source
	var input io.Reader
	var inputFileName string

	if inputPath == "" || inputPath == "-" {
		// Read from stdin
		input = os.Stdin
		inputFileName = "<stdin>"
	} else {
		// Check if input file exists and is readable
		if err := checkInputFile(inputPath); err != nil {
			printError("input file %s: %v", inputPath, err)
			os.Exit(1)
		}
		// Read from file
		f, err := os.Open(inputPath)
		if err != nil {
			printError("failed to open input file: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
		inputFileName = inputPath
	}

	// Determine base directory
	if baseDir == "" && inputPath != "" && inputPath != "-" {
		baseDir = filepath.Dir(inputPath)
	}
	if baseDir == "" {
		baseDir = "."
	}

	// Validate base directory
	if baseDir != "" {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			printError("base directory does not exist: %s", baseDir)
			os.Exit(1)
		}
	}

	// Parse the document with error recovery
	p, err := parser.NewParserFromReader(input,
		parser.WithIncludeProcessor(processor.NewIncludeProcessor(baseDir)),
		parser.WithBaseDir(baseDir))
	if err != nil {
		printError("failed to create parser: %v", err)
		os.Exit(1)
	}

	doc, err := p.Parse()
	if err != nil {
		printError("failed to parse document: %v", err)
		os.Exit(1)
	}

	// Handle empty document gracefully
	if doc == nil {
		if verbosity > verbosityQuiet {
			printWarning("document is empty")
		}
		// Create minimal document for output
		doc = &ast.NodeDocument{
			Attributes: make(map[string]string),
			Blocks:     []ast.Node{},
		}
	}

	// Apply command-line attribute definitions
	for _, attrDef := range attributeDefs {
		parts := strings.SplitN(attrDef, "=", 2)
		attrName := strings.TrimSpace(parts[0])
		if attrName == "" {
			printWarning("skipping empty attribute name")
			continue
		}
		if len(parts) == 1 {
			// Attribute without value or with ! suffix
			if strings.HasSuffix(attrName, "!") {
				// Attribute unset - remove it
				doc.Attributes[strings.TrimSuffix(attrName, "!")] = ""
			} else {
				// Set attribute to empty string
				doc.Attributes[attrName] = ""
			}
		} else {
			// Attribute with value
			doc.Attributes[attrName] = parts[1]
		}
	}

	// Set doctype attribute if specified
	if docType != "" && docType != "article" {
		doc.Attributes["doctype"] = docType
	}

	// Set section numbers if requested
	if sectionNumbers {
		doc.Attributes["sectnums"] = ""
	}

	if verbosity >= verbosityVerbose {
		fmt.Fprintf(os.Stderr, "asciidoc-parser-go: parsed %s: %d blocks, %d attributes\n",
			inputFileName, len(doc.Blocks), len(doc.Attributes))
	}

	// Convert to output format
	var output []byte
	var outputExt string

	switch strings.ToLower(backend) {
	case "html5", "html":
		htmlConverter := converter.NewHTML5Converter()
		if embeddedMode {
			htmlConverter.WithoutHeaderFooter()
		}
		var buf strings.Builder
		if err := htmlConverter.Convert(doc, &buf); err != nil {
			printError("failed to convert document: %v", err)
			os.Exit(1)
		}
		output = []byte(buf.String())
		outputExt = ".html"
	case "pdf":
		pdfConverter := converter.NewPDFConverter()

		// Apply PDF-specific options
		switch strings.ToLower(pdfPageSize) {
		case "a4":
			pdfConverter.WithA4()
		case "letter":
			pdfConverter.WithLetter()
		}

		if pdfTOC {
			pdfConverter.WithTOC(true)
		}

		if pdfCoverPage {
			pdfConverter.WithCoverPage(true)
		}

		if pdfNoPageNumbers {
			pdfConverter.WithPageNumbers(false)
		}

		if pdfHeaderTemplate != "" {
			pdfConverter.WithHeaderTemplate(pdfHeaderTemplate)
		}

		if pdfFooterTemplate != "" {
			pdfConverter.WithFooterTemplate(pdfFooterTemplate)
		}

		// Extract metadata from document attributes
		if title, ok := doc.Attributes["title"]; ok {
			pdfConverter.WithPDFTitle(title)
		}
		if author, ok := doc.Attributes["author"]; ok {
			pdfConverter.WithPDFAuthor(author)
		}
		if keywords, ok := doc.Attributes["keywords"]; ok {
			pdfConverter.WithPDFKeywords(keywords)
		}
		if subject, ok := doc.Attributes["subject"]; ok {
			pdfConverter.WithPDFSubject(subject)
		}

		var buf bytes.Buffer
		if err := pdfConverter.Convert(doc, &buf); err != nil {
			printError("failed to convert document: %v", err)
			os.Exit(1)
		}
		output = buf.Bytes()
		outputExt = ".pdf"
	case "docbook", "db":
		docbookConverter := converter.NewDocBookConverter()
		if docType != "" {
			docbookConverter.WithDoctype(docType)
		}
		var buf bytes.Buffer
		if err := docbookConverter.Convert(doc, &buf); err != nil {
			printError("failed to convert document: %v", err)
			os.Exit(1)
		}
		output = buf.Bytes()
		outputExt = ".xml"
	case "manpage", "man":
		// Extract manual name from input filename
		manualName := ""
		if inputPath != "" && inputPath != "-" {
			manualName = strings.ToUpper(filepath.Base(inputPath))
			manualName = strings.TrimSuffix(manualName, ".adoc")
			manualName = strings.TrimSuffix(manualName, ".asciidoc")
		}

		// Extract manual section from attributes or default to 1
		manualSection := "1"
		if sec, ok := doc.Attributes["mansection"]; ok {
			manualSection = sec
		} else if sec, ok := doc.Attributes["man-manual-section"]; ok {
			manualSection = sec
		}

		manConverter := converter.NewManPageConverter()
		if manualName != "" {
			manConverter.WithManualName(manualName)
		}
		if manualSection != "" {
			manConverter.WithSection(manualSection)
		}
		var buf bytes.Buffer
		if err := manConverter.Convert(doc, &buf); err != nil {
			printError("failed to convert document: %v", err)
			os.Exit(1)
		}
		output = buf.Bytes()
		outputExt = "." + manualSection
	case "epub":
		// Extract title and author from document
		epubConverter := converter.NewEPUBConverter()
		if doc.Header != nil {
			if doc.Header.Title != "" {
				epubConverter.WithTitle(doc.Header.Title)
			}
			if doc.Header.Author != "" {
				epubConverter.WithAuthor(doc.Header.Author)
			}
		}
		// Override with attributes if present
		if title, ok := doc.Attributes["title"]; ok {
			epubConverter.WithTitle(title)
		}
		if author, ok := doc.Attributes["author"]; ok {
			epubConverter.WithAuthor(author)
		}
		if lang, ok := doc.Attributes["lang"]; ok {
			epubConverter.WithLanguage(lang)
		}
		if publisher, ok := doc.Attributes["publisher"]; ok {
			epubConverter.WithPublisher(publisher)
		}
		var buf bytes.Buffer
		if err := epubConverter.Convert(doc, &buf); err != nil {
			printError("failed to convert document: %v", err)
			os.Exit(1)
		}
		output = buf.Bytes()
		outputExt = ".epub"
	default:
		printError("unsupported backend '%s' (supported: html5, pdf, docbook, manpage, epub)", backend)
		os.Exit(1)
	}

	// Determine output file path
	outputFilePath := outputFile
	if outputFilePath == "" && inputPath != "" && inputPath != "-" {
		// Auto-generate output filename based on input
		baseName := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
		outputFilePath = baseName + outputExt
	}

	// Handle destination directory
	if destDir != "" && outputFilePath != "" && outputFilePath != "-" {
		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(destDir, 0755); err != nil {
			printError("failed to create destination directory: %v", err)
			os.Exit(1)
		}
		// Place output in destination directory
		baseName := filepath.Base(outputFilePath)
		outputFilePath = filepath.Join(destDir, baseName)
	}

	// Write output
	if outputFilePath == "" || outputFilePath == "-" {
		// Write to stdout
		if _, err := fmt.Fprint(os.Stdout, output); err != nil {
			printError("failed to write output: %v", err)
			os.Exit(1)
		}
	} else {
		// Write to file
		if err := os.WriteFile(outputFilePath, []byte(output), 0644); err != nil {
			printError("failed to write output file: %v", err)
			os.Exit(1)
		}
		if verbosity >= verbosityNormal {
			fmt.Fprintf(os.Stderr, "asciidoc-parser-go: wrote %s\n", outputFilePath)
		}
	}

	// Print timings if requested
	if timings {
		duration := time.Since(startTime)
		fmt.Fprintf(os.Stderr, "asciidoc-parser-go: timings: parse: %s, convert: %s, total: %s\n",
			duration, duration, duration)
	}
}

// printHelp prints the help message
func printHelp() {
	fmt.Printf(`asciidoc-parser-go - AsciiDoc to HTML5/PDF/DocBook/Man Page/EPUB Converter
Version %s

Usage: asciidoctor [OPTIONS] FILE
Convert the AsciiDoc input FILE to HTML5, PDF, DocBook, Man Page, or EPUB output.
Unless specified otherwise, output is written to a file whose name is derived
from the input file. Application log messages are printed to STDERR.

Arguments:
  FILE                    Path to AsciiDoc file to convert.
                          Use '-' to read from STDIN.

Options:
  -b, --backend BACKEND           Set backend output format: [html5, pdf, docbook, manpage, epub] (default: html5)
  -d, --doctype DOCTYPE           Document type: [article, book, manpage, inline] (default: article)
  -e, --embedded                  Suppress enclosing document structure (same as -s)
  -o, --out-file FILE             Output file (default: based on input file; use - for STDOUT)
  -s, --no-header-footer          Suppress enclosing document structure (same as -e)
  -n, --section-numbers           Auto-number section titles
  -a, --attribute name[=value]    Define a document attribute (may be specified multiple times)
                                  Format: name, name! (unset), or name=value
  -B, --base-dir DIR              Base directory containing the document and resources
                                  (default: directory of input file)
  -D, --destination-dir DIR       Destination output directory
                                  (default: directory of input file)
  -R, --source-dir DIR            Source root directory
  -q, --quiet                     Silence application log messages
  -v, --verbose                   Enable verbose output (show all log messages)
  -w, --warnings                  Turn on script warnings
  -t, --timings                   Print timings report
  -V, --version                   Display version information
  -h, --help                      Show this help message

PDF Options:
  --pdf-page-size SIZE            PDF page size: [letter, a4] (default: letter)
  --pdf-toc                       Generate table of contents in PDF
  --pdf-cover-page                Generate cover page in PDF
  --pdf-no-page-numbers           Disable page numbers in PDF
  --pdf-header TEMPLATE           PDF header template
  --pdf-footer TEMPLATE           PDF footer template

Examples:
  # Convert file to HTML
  asciidoctor document.adoc

  # Convert file to PDF
  asciidoctor -b pdf document.adoc

  # Convert file to PDF with TOC and cover page
  asciidoctor -b pdf --pdf-toc --pdf-cover-page document.adoc

  # Convert file to PDF with A4 page size
  asciidoctor -b pdf --pdf-page-size a4 document.adoc

  # Convert file to DocBook
  asciidoctor -b docbook document.adoc

  # Convert file to Man Page
  asciidoctor -b manpage mycommand.adoc

  # Convert file to EPUB
  asciidoctor -b epub document.adoc

  # Convert with custom output file
  asciidoctor -o output.html document.adoc

  # Convert with embedded output (no header/footer)
  asciidoctor -s document.adoc
  asciidoctor --embedded document.adoc

  # Define attributes
  asciidoctor -a version=1.0 -a author="John Doe" document.adoc

  # Enable section numbering
  asciidoctor -n document.adoc

  # Convert with destination directory
  asciidoctor -D ./output document.adoc

  # Read from stdin, write to stdout
  asciidoctor < document.adoc > output.html

`, version)
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("asciidoc-parser-go %s [https://github.com/haimiyahya/asciidoc-parser-go]\n", version)
	fmt.Printf("Runtime Environment (go version %s)\n", "1.22")
}

// printError prints an error message with the program prefix
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "asciidoc-parser-go: FAILED: "+format+"\n", args...)
}

// printWarning prints a warning message with the program prefix
func printWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "asciidoc-parser-go: WARNING: "+format+"\n", args...)
}

// checkInputFile checks if the input file exists and is readable
func checkInputFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("access denied: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file (is a %s)", info.Mode().Type().String())
	}

	// Check if file is readable
	if info.Mode().Perm()&0400 == 0 {
		return fmt.Errorf("file is not readable")
	}

	return nil
}

// This is a stub for future extension loading
func loadExtension(name string) error {
	return errors.New("extensions not yet implemented")
}
