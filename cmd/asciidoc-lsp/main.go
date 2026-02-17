// Command asciidoc-lsp runs the AsciiDoc Language Server.
//
// Usage:
//
//	asciidoc-lsp
//
// The server communicates via stdin/stdout using the Language Server Protocol.
// This command is typically launched by an editor or IDE, not run manually.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/haimiyahya/asciidoc-parser-go/internal/lsp"
)

var (
	version = "0.2.0"
	commit  = "unknown"
	date    = "unknown"

	// Command-line flags
	showVersion bool
)

func main() {
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("asciidoc-language-server %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Log to file for debugging
	logFile, err := os.OpenFile(os.TempDir()+"/asciidoc-lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Println("Starting AsciiDoc Language Server")

	server := lsp.NewServer()

	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
