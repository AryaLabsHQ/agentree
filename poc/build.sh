#!/bin/bash

echo "Building Agentree Multiplexer POC..."

# Build all POC programs
echo "Building PTY test..."
go build -o pty-test pty-test-simple.go

echo "Building TUI test..."
go build -o tui-test tui-test.go

echo "Building Multiplexer POC..."
go build -o multiplex-poc multiplex-poc.go

echo "Build complete!"
echo ""
echo "To run the programs:"
echo "  ./pty-test        - Test PTY functionality"
echo "  ./tui-test        - Test TUI framework (requires terminal)"
echo "  ./multiplex-poc   - Full POC (requires terminal)"
echo ""
echo "Note: TUI programs must be run in a real terminal, not in an IDE console."