// Package main provides the CLI for the asciidoc parser.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/haimiyahya/asciidoc-parser-go/internal/converter"
	"github.com/haimiyahya/asciidoc-parser-go/internal/parser"
	"github.com/haimiyahya/asciidoc-parser-go/internal/processor"
	flag "github.com/spf13/pflag"
)

const (
	version = "0.1.0"
)

func main() {
	// Define command-line flags
	var (
		outputFile      string
		backend         string
		attributeDefs   []string
		baseDir         string
		noHeaderFooter  bool
		verbose         bool
		showVersion     bool
		help            bool
	)

	// Define flags
	flag.StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	flag.StringVarP(&backend, "backend", "b", "html5", "Backend format (html5)")
	flag.StringSliceVarP(&attributeDefs, "attribute", "a", nil, "Define attribute (can be used multiple times)")
	flag.StringVarP(&baseDir, "base-dir", "D", "", "Base directory for includes (default: input file directory)")
	flag.BoolVar(&noHeaderFooter, "no-header-footer", false, "Suppress document header and footer")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	flag.BoolVarP(&showVersion, "version", "V", false, "Show version information")
	flag.BoolVarP(&help, "help", "h", false, "Show this help message")

	flag.Parse()

	// Handle help
	if help {
		printHelp()
		os.Exit(0)
	}

	// Handle version
	if showVersion {
		fmt.Printf("asciidoc-parser-go version %s\n", version)
		os.Exit(0)
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
		// Read from file
		f, err := os.Open(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open input file: %v\n", err)
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

	// Parse the document
	p, err := parser.NewParserFromReader(input,
		parser.WithIncludeProcessor(processor.NewIncludeProcessor(baseDir)),
		parser.WithBaseDir(baseDir))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create parser: %v\n", err)
		os.Exit(1)
	}

	doc, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse document: %v\n", err)
		os.Exit(1)
	}

	// Apply command-line attribute definitions
	for _, attrDef := range attributeDefs {
		parts := strings.SplitN(attrDef, "=", 2)
		if len(parts) == 1 {
			// Attribute without value, set to empty string
			doc.Attributes[parts[0]] = ""
		} else {
			// Attribute with value
			doc.Attributes[parts[0]] = parts[1]
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Parsed %s: %d blocks, %d attributes\n",
			inputFileName, len(doc.Blocks), len(doc.Attributes))
	}

	// Convert to output format
	var output string
	switch strings.ToLower(backend) {
	case "html5", "html":
		htmlConverter := converter.NewHTML5Converter()
		if noHeaderFooter {
			htmlConverter.WithoutHeaderFooter()
		}
		var buf strings.Builder
		if err := htmlConverter.Convert(doc, &buf); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to convert document: %v\n", err)
			os.Exit(1)
		}
		output = buf.String()
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported backend '%s'\n", backend)
		fmt.Fprintf(os.Stderr, "Supported backends: html5\n")
		os.Exit(1)
	}

	// Write output
	var outputFileWriter io.Writer
	if outputFile == "" || outputFile == "-" {
		outputFileWriter = os.Stdout
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		outputFileWriter = f
	}

	if _, err := fmt.Fprint(outputFileWriter, output); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to write output: %v\n", err)
		os.Exit(1)
	}
}

// printHelp prints the help message
func printHelp() {
	fmt.Printf(`asciidoc-parser-go - AsciiDoc Parser in Go
Version %s

USAGE:
    asciidoc [OPTIONS] [INPUT_FILE]

ARGUMENTS:
    INPUT_FILE    Input AsciiDoc file (default: stdin)
                  Use "-" to explicitly read from stdin

OPTIONS:
    -o, --output <file>           Output file (default: stdout)
                                  Use "-" to explicitly write to stdout
    -b, --backend <format>        Backend format (default: html5)
                                  Supported: html5
    -a, --attribute <name=value>  Define attribute (can be used multiple times)
    -D, --base-dir <dir>          Base directory for includes
                                  (default: input file directory)
    --no-header-footer            Suppress document header and footer
    -v, --verbose                 Enable verbose output
    -V, --version                 Show version information
    -h, --help                    Show this help message

EXAMPLES:
    # Convert file to HTML
    asciidoc document.adoc

    # Convert file to HTML with output file
    asciidoc -o output.html document.adoc

    # Read from stdin, write to stdout
    asciidoc < document.adoc > output.html

    # Define attributes
    asciidoc -a version=1.0 -a author="John Doe" document.adoc

    # Suppress header/footer for embedding
    asciidoc --no-header-footer document.adoc

    # Process with custom base directory for includes
    asciidoc -D /path/to/includes document.adoc

`, version)
}

// checkWriter is a helper to check if a writer is a file
func checkWriter(w io.Writer) (*os.File, bool) {
	if f, ok := w.(*os.File); ok {
		return f, true
	}
	return nil, false
}

// This is a stub for future extension loading
func loadExtension(name string) error {
	return errors.New("extensions not yet implemented")
}
