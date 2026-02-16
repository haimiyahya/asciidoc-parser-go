// Package lsp provides Language Server Protocol implementation for AsciiDoc.
package lsp

import (
	"context"
	"io"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

// Streamer wraps stdio for JSON-RPC communication.
type Streamer struct {
	in  io.Reader
	out io.Writer
}

// Close implements io.Closer.
func (s *Streamer) Close() error {
	return nil
}

// NewStreamer creates a new stdio streamer.
func NewStreamer() *Streamer {
	return &Streamer{
		in:  os.Stdin,
		out: os.Stdout,
	}
}

// Read implements io.Reader.
func (s *Streamer) Read(p []byte) (int, error) {
	return s.in.Read(p)
}

// Write implements io.Writer.
func (s *Streamer) Write(p []byte) (int, error) {
	return s.out.Write(p)
}

// Run starts the LSP server with stdio transport.
func (s *Server) Run() error {
	stream := jsonrpc2.NewBufferedStream(NewStreamer(), jsonrpc2.PlainObjectCodec{})
	conn := jsonrpc2.NewConn(context.Background(), stream, s)

	// Wait for connection to close
	<-conn.DisconnectNotify()
	return nil
}
