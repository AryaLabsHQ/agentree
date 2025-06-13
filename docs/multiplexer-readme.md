# Agentree Multiplexer

The Agentree Multiplexer allows you to run and manage multiple Claude Code instances across different Git worktrees simultaneously.

## Overview

The multiplexer provides a terminal UI (TUI) for:
- Running Claude Code in multiple worktrees at once
- Monitoring token usage and costs across instances
- Switching between instances to view their output
- Coordinating work across branches
- Managing instance lifecycle (start/stop/restart)

## Usage

### Basic Usage

```bash
# Launch multiplexer for specific worktrees
agentree multiplex feat-auth feat-ui

# Launch all worktrees
agentree multiplex --all

# Use custom configuration
agentree multiplex --config multiplex.yml
```

### Keyboard Shortcuts

- `q` - Quit
- `s` - Start instance
- `x` - Stop instance
- `r` - Restart instance
- `c` - Clear output
- `j`/`↓` - Navigate down
- `k`/`↑` - Navigate up
- `Enter` - Focus instance
- `?` - Show help

## Architecture

The multiplexer consists of several key components:

### Process Manager
- Manages Claude Code process lifecycle
- Uses PTY (pseudo-terminal) for proper terminal emulation
- Tracks instance state (idle, running, thinking, etc.)
- Handles process output and error streams

### UI Controller
- Built with tcell for cross-platform terminal UI
- Three-panel layout: sidebar, main view, status bar
- Real-time updates of instance states
- Responsive to terminal resize

### Token Tracker
- Parses Claude output for token usage information
- Tracks cumulative usage per instance
- Estimates costs based on token consumption
- Enforces token limits

### Event System
- Event-driven architecture for loose coupling
- Process events (start, stop, output, errors)
- UI events (resize, focus, input)
- System events (token updates, config changes)

## Configuration

The multiplexer can be configured via YAML files:

```yaml
# .agentree-multiplex.yml
auto_start: false
token_limit: 100000
theme: dark

instances:
  - worktree: feat-auth
    auto_start: true
    token_limit: 50000
    role: developer
  
  - worktree: feat-ui
    auto_start: true
    token_limit: 50000
    role: developer

ui:
  sidebar_width: 25
  show_token_usage: true
  show_timestamps: true
  scrollback_size: 10000

coordination:
  shared_context: true
  sync_interval: 30s
  enable_ipc: false
```

## Implementation Status

This is currently a **stub implementation** that demonstrates the architecture and API design. The following components are stubbed:

- ✅ Command structure (`cmd/multiplex.go`)
- ✅ Core multiplexer orchestration (`internal/multiplex/multiplexer.go`)
- ✅ Process management (`internal/multiplex/process.go`)
- ✅ Event system (`internal/multiplex/events.go`)
- ✅ Configuration management (`internal/multiplex/config.go`)
- ✅ Terminal emulation (`internal/multiplex/terminal/vt.go`)
- ✅ UI components (`internal/multiplex/ui/controller.go`, `components.go`)
- ✅ Token tracking (`internal/multiplex/token.go`)
- ✅ Worktree discovery (`internal/multiplex/discovery.go`)

## Next Steps

To complete the implementation:

1. **Add dependencies**:
   ```bash
   go get github.com/gdamore/tcell/v2
   go get github.com/creack/pty
   go get gopkg.in/yaml.v3
   ```

2. **Implement ANSI parser** in the virtual terminal

3. **Add token parsing logic** for actual Claude output formats

4. **Implement process coordination** features

5. **Add tests** for all components

6. **Create integration tests** with real Claude instances

## Design Principles

1. **Modular Architecture**: Each component has a single responsibility
2. **Event-Driven**: Loose coupling through event system
3. **Testable**: Interfaces and dependency injection for easy testing
4. **Extensible**: Easy to add new features and UI components
5. **Performance**: Efficient handling of terminal output and UI updates

## Future Enhancements

- **Multi-instance coordination**: Share context between instances
- **Task distribution**: Automatically assign tasks to instances
- **Performance monitoring**: CPU/memory usage per instance
- **Session recording**: Record and replay sessions
- **Remote instances**: Support for remote Claude instances
- **Plugin system**: Extend functionality with plugins