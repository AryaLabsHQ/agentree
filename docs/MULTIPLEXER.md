# Agentree Multiplexer

The Agentree Multiplexer enables you to run multiple Claude Code instances simultaneously across different Git worktrees, providing a powerful interface for parallel AI-assisted development.

## Features

- 🖥️ **Terminal UI**: Full-featured TUI with sidebar, output view, and status bar
- 🚀 **Parallel Execution**: Run multiple Claude instances at once
- 📊 **Token Tracking**: Monitor token usage and costs in real-time
- 🎮 **Process Control**: Start, stop, and restart instances on demand
- ⚙️ **Configurable**: YAML-based configuration with per-instance settings
- 🤝 **Coordination**: Built-in support for multi-instance collaboration

## Quick Start

```bash
# Launch multiplexer for specific worktrees
agentree multiplex feat-auth feat-ui

# Launch all worktrees
agentree multiplex --all

# Use configuration file
agentree multiplex --config multiplex.yml
```

## User Interface

```
┌─────────────────────────┬───────────────────────────────────────────────────────┐
│ Instances              │ Claude Output                                       │
│ ● feat-auth [running]  │                                                     │
│ ○ feat-ui              │ Human: implement user authentication                │
│ ◍ feat-api [thinking]  │                                                     │
│                        │ Assistant: I'll help you implement user            │
│ States:                │ authentication. Let me analyze your current         │
│ ○ idle                 │ setup and create a secure implementation...        │
│ ◐ starting             │                                                     │
│ ● running              │ First, I'll check your existing auth structure:    │
│ ◍ thinking             │                                                     │
│ ◑ stopping             │ [Reading src/auth/index.js...]                     │
│ ◯ stopped              │                                                     │
│ ✗ crashed              │                                                     │
└─────────────────────────┴───────────────────────────────────────────────────────┘
 Tokens: 5,234/100,000 | Cost: ~$0.0421 | Instances: 3/3 | Press ? for help   
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` | Quit multiplexer |
| `s` | Start selected instance |
| `x` | Stop selected instance |
| `r` | Restart selected instance |
| `c` | Clear output |
| `j`/`↓` | Navigate down |
| `k`/`↑` | Navigate up |
| `Enter` | Focus on instance |
| `?` | Show help |

## Configuration

Create `.agentree-multiplex.yml` in your project root:

```yaml
# Global settings
auto_start: false
token_limit: 100000
theme: dark

# Per-instance configuration
instances:
  - worktree: feat-ui
    auto_start: true
    token_limit: 30000
    role: developer
    environment:
      - CLAUDE_FOCUS=frontend
  
  - worktree: feat-api
    auto_start: true
    token_limit: 40000
    role: developer
    environment:
      - CLAUDE_FOCUS=backend

# UI settings
ui:
  sidebar_width: 25
  show_token_usage: true
  show_timestamps: false
  scrollback_size: 10000
```

## Architecture

### Components

1. **Process Manager**: Handles Claude process lifecycle using PTY
2. **Event System**: Event-driven architecture for loose coupling
3. **UI Controller**: Terminal UI management with tcell
4. **Token Tracker**: Parses and tracks token usage from output
5. **Configuration**: YAML-based settings with validation

### Process States

- **Idle**: Process not started
- **Starting**: Process launching
- **Running**: Active and accepting input
- **Thinking**: Processing a request
- **Stopping**: Shutting down
- **Stopped**: Cleanly terminated
- **Crashed**: Unexpected termination

## Token Tracking

The multiplexer automatically parses token usage from Claude output:

- Tracks input and output tokens separately
- Estimates costs based on current pricing
- Enforces token limits per instance
- Shows real-time usage in status bar

## Examples

See the `examples/` directory for configuration examples:

- `multiplex-basic.yml`: Simple configuration with comments
- `multiplex-advanced.yml`: Per-instance settings and roles
- `multiplex-team.yml`: Team collaboration setup
- `multiplex-minimal.yml`: Minimal configuration

## Demo

Run the interactive demo:

```bash
./examples/demo-multiplex.sh
```

This creates test worktrees and launches the multiplexer with mock Claude instances.

## Development

### Testing

```bash
# Run unit tests
go test ./internal/multiplex

# Run with mock Claude
MOCK_CLAUDE_PATH=./mock-claude agentree multiplex feat-test
```

### Building

```bash
# Build with multiplexer support
make build

# Build mock Claude for testing
go build -o mock-claude ./cmd/mock-claude
```

## Troubleshooting

### Common Issues

1. **"claude executable not found"**
   - Ensure Claude CLI is installed and in PATH
   - Use `MOCK_CLAUDE_PATH` for testing

2. **"failed to create screen"**
   - Requires a proper terminal environment
   - Won't work in non-interactive shells

3. **High token usage**
   - Set token limits in configuration
   - Monitor usage in status bar
   - Stop instances when not needed

### Debug Mode

Enable debug logging:

```yaml
log_level: debug
```

Or use the hidden flag:

```bash
agentree multiplex --debug
```

## Future Enhancements

- [ ] Session recording and replay
- [ ] Remote instance support
- [ ] Advanced coordination modes
- [ ] Token usage analytics
- [ ] Plugin system
- [ ] Web UI option

## Contributing

The multiplexer is under active development. Contributions welcome!

1. Test with real Claude instances
2. Report bugs and feature requests
3. Submit PRs for enhancements

See `docs/multiplexer-design.md` for architecture details.