# Agentree Multiplexer POC

This directory contains proof-of-concept implementations for the agentree multiplexer.

## Files

- `pty-test-simple.go` - Basic PTY test to validate process management
- `tui-test.go` - Basic TUI test using tcell framework
- `multiplex-poc.go` - Combined PTY + TUI proof of concept

## Running the POC

**Note**: These programs require a real terminal to run properly. They won't work in environments without TTY support.

### 1. PTY Test
```bash
go run pty-test-simple.go
```
This validates that we can:
- Launch processes in a PTY
- Capture colored output
- Handle process lifecycle

### 2. TUI Test
```bash
go run tui-test.go
```
This validates that we can:
- Create a terminal UI with tcell
- Handle keyboard input
- Draw and update the screen

### 3. Multiplexer POC
```bash
go run multiplex-poc.go
```
This is the main POC that combines PTY and TUI to show:
- Process output in a TUI
- Scrolling through output
- Real-time updates
- Basic interaction

## Controls (Multiplexer POC)

- `q` - Quit
- `↑/↓` - Scroll line by line
- `PgUp/PgDn` - Scroll by page
- `Home/End` - Go to top/bottom
- `c` - Clear output

## Building

```bash
go build -o multiplex-poc multiplex-poc.go
./multiplex-poc
```

## Test with Claude Code

To test with actual Claude Code:
1. Modify `multiplex-poc.go` line ~100 to:
   ```go
   m.cmd = exec.Command("claude", "code")
   m.cmd.Dir = "/path/to/your/worktree"
   ```
2. Run the POC in a worktree directory

## Findings

### ✅ Working
- PTY process management
- ANSI color preservation
- TUI framework (tcell)
- Output buffering and scrolling
- Real-time output capture

### ⚠️ Considerations
- Need proper ANSI parsing for color support in TUI
- Memory management for large outputs
- Process cleanup on exit
- Cross-platform compatibility (currently macOS/Linux)

### 📝 Next Steps
1. Implement proper ANSI sequence parsing
2. Add multiple process support
3. Improve UI layout (sidebar, status bar)
4. Add configuration support
5. Test with real Claude Code instances