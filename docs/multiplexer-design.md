# Agentree Multiplexer Design Document

## Overview

The Agentree Multiplexer is a terminal-based UI system for managing multiple Claude Code instances across different git worktrees. Inspired by SST's mosaic implementation and turborepo's task orchestration, it provides a unified interface for coordinating AI agents working on different aspects of a project simultaneously.

## Goals

1. **Parallel AI Development**: Enable multiple Claude Code instances to work on different features/branches simultaneously
2. **Resource Efficiency**: Monitor and manage token usage across instances
3. **Coordination**: Facilitate communication and context sharing between agents
4. **Developer Experience**: Provide an intuitive TUI for managing multiple AI coding sessions

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Multiplexer Core                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐   │
│  │Process      │  │Event         │  │UI                 │   │
│  │Manager      │  │Dispatcher    │  │Controller         │   │
│  └─────────────┘  └──────────────┘  └───────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐   │
│  │Claude       │  │Claude        │  │Claude             │   │
│  │Instance 1   │  │Instance 2    │  │Instance N         │   │
│  │(VT + PTY)   │  │(VT + PTY)    │  │(VT + PTY)         │   │
│  └─────────────┘  └──────────────┘  └───────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### 1. Multiplexer Core (`cmd/multiplex.go`)
- Main entry point and command setup
- Initialize TUI framework (tcell)
- Coordinate between all components
- Handle graceful shutdown

#### 2. Process Manager (`internal/multiplex/process.go`)
- Spawn and manage Claude Code processes
- Track process state (running, stopped, crashed)
- Handle process lifecycle (start, stop, restart)
- Resource monitoring (CPU, memory, token usage)

#### 3. Event Dispatcher (`internal/multiplex/events.go`)
- Central event bus for all components
- Event types:
  - Process events (started, stopped, output, error)
  - UI events (resize, focus change, user input)
  - Coordination events (sync request, context share)
- Async event handling with channels

#### 4. UI Controller (`internal/multiplex/ui/`)
- Layout management (sidebar, main area, footer)
- Render loop and screen updates
- Input handling (keyboard, mouse)
- Theme and styling

#### 5. Virtual Terminal (`internal/multiplex/terminal/`)
- Terminal emulator per Claude instance
- ANSI escape sequence parsing
- Scrollback buffer management
- Text selection and copy support

### Data Structures

```go
// Instance represents a single Claude Code process
type Instance struct {
    ID          string
    Worktree    string
    Process     *os.Process
    VT          *VirtualTerminal
    State       InstanceState
    TokenUsage  TokenStats
    StartedAt   time.Time
    LastActive  time.Time
}

// InstanceState represents the current state
type InstanceState int

const (
    StateIdle InstanceState = iota
    StateRunning
    StateThinking
    StateStopped
    StateCrashed
)

// TokenStats tracks token usage
type TokenStats struct {
    InputTokens  int64
    OutputTokens int64
    TotalCost    float64
    LastUpdated  time.Time
}

// Event represents a system event
type Event interface {
    Type() EventType
    Timestamp() time.Time
}

// ProcessEvent for process-related events
type ProcessEvent struct {
    InstanceID string
    Type       ProcessEventType
    Data       interface{}
    Time       time.Time
}
```

## User Interface Design

### Layout

```
┌─ agentree multiplexer ─────────────────────────────────────────┐
│ Worktrees          │ agent/feat-authentication                 │
│ ──────────────     │ ╭─ Claude ──────────────────────────────╮ │
│ ● auth       12.3k │ │ I'll implement the login system...    │ │
│ ○ ui          5.2k │ ╰───────────────────────────────────────╯ │
│ ○ api         3.1k │                                           │
│ ○ main        1.0k │ > npm run test:auth                       │
│                    │   ✓ auth.test.js (5 tests passed)         │
│ [Inactive]         │   ✓ login.test.js (3 tests passed)        │
│ ─────────────      │                                           │
│ ✗ old-feat    dead │ Token usage: 12.3k/50k (24.6%)            │
│                    │ Cost: $0.48 | Time: 12m 34s               │
└────────────────────┴───────────────────────────────────────────┘
 j/k:nav ↵:focus/start TAB:next q:quit r:restart x:kill ?:help
```

### UI Components

#### Sidebar (20-25 chars wide)
- List of all worktrees
- Visual indicators:
  - `●` Active/running
  - `○` Idle/ready
  - `✗` Dead/crashed
- Token usage indicator (abbreviated)
- Grouped by status (active first)

#### Main Area
- Selected instance output
- Terminal emulation with full ANSI support
- Scrollback buffer (10k lines)
- Text selection with mouse
- Search functionality (Ctrl+F)

#### Status Bar
- Detailed token usage for selected instance
- Cost tracking
- Execution time
- Error indicators

#### Footer
- Context-sensitive hotkeys
- Mode indicator (normal, search, select)

### Keyboard Shortcuts

**Navigation**
- `j/k` or `↑/↓`: Navigate instances
- `TAB`: Cycle through active instances
- `Enter`: Focus instance (or start if stopped)
- `Space`: Toggle instance details

**Instance Control**
- `s`: Start/stop instance
- `r`: Restart instance
- `x`: Kill instance (with confirmation)
- `X`: Kill all instances

**Output Control**
- `f`: Follow output (auto-scroll)
- `Ctrl+F`: Search in output
- `Ctrl+C`: Copy selection
- `PageUp/PageDown`: Scroll output

**Multiplexer Control**
- `q`: Quit (with confirmation if instances running)
- `?`: Show help
- `:`: Command mode

### Mouse Support
- Click sidebar item to select
- Click and drag in output to select text
- Scroll wheel for output scrolling
- Double-click to select word
- Triple-click to select line

## Features

### 1. Process Management

**Auto-start Behavior**
- Option to auto-start all instances on launch
- Or manual start per instance
- Remember last state on restart

**Resource Limits**
- Token usage limits per instance
- Total token budget across all instances
- Automatic pause when approaching limits

**Health Monitoring**
- Detect crashed instances
- Auto-restart with backoff
- Error reporting

### 2. Coordination Features

**Shared Context**
- Shared CLAUDE.md across instances
- Project-wide knowledge base
- Cross-instance file awareness

**Task Distribution**
- Assign tasks to specific instances
- Load balancing based on token usage
- Priority queuing

**Inter-agent Communication**
- Message passing between instances
- Shared clipboard
- Sync points for coordinated changes

### 3. Output Management

**Filtering**
- Filter by log level
- Search across all instances
- Regex support

**Recording**
- Save session transcripts
- Export token usage reports
- Replay sessions

### 4. Integration Features

**Git Integration**
- Show current branch per instance
- Commit status indicators
- Auto-switch on branch change

**File System Watching**
- Detect file changes
- Show which instance modified files
- Conflict detection

## Implementation Plan

### Phase 1: Core Infrastructure (Week 1-2)
1. Set up basic TUI with tcell
2. Implement process manager
3. Create event system
4. Basic UI layout (sidebar + main area)

### Phase 2: Terminal Emulation (Week 2-3)
1. Integrate terminal emulator
2. PTY management
3. Output capture and buffering
4. Basic scrollback support

### Phase 3: Process Control (Week 3-4)
1. Start/stop/restart functionality
2. State management
3. Error handling
4. Basic token tracking

### Phase 4: Advanced Features (Week 4-5)
1. Inter-instance communication
2. Shared context management
3. Search and filtering
4. Mouse support

### Phase 5: Polish (Week 5-6)
1. Performance optimization
2. Configuration system
3. Help system
4. Testing and documentation

## Technical Decisions

### Dependencies
- `github.com/gdamore/tcell/v2`: Terminal UI framework
- `github.com/creack/pty`: Pseudo-terminal support
- Custom terminal emulator (adapted from tcell-term)

### Configuration
```yaml
# ~/.config/agentree/multiplex.yml
defaults:
  auto_start: false
  token_limit: 50000
  theme: dark
  
instances:
  - worktree: main
    auto_start: true
    token_limit: 10000
    
shortcuts:
  quit: "q"
  restart: "r"
  kill: "x"
```

### State Persistence
- Save UI state (selected instance, scroll positions)
- Persist token usage across restarts
- Session replay capability

## Security Considerations

1. **Process Isolation**: Each Claude instance runs in separate process
2. **Token Management**: Secure storage of API tokens
3. **File Access**: Respect gitignore and file permissions
4. **Output Sanitization**: Prevent terminal escape injection

## Performance Considerations

1. **Lazy Loading**: Only render visible output
2. **Buffering**: Efficient scrollback implementation
3. **Event Batching**: Batch UI updates
4. **Resource Limits**: CPU/memory limits per instance

## Future Enhancements

1. **Web UI**: Browser-based alternative to TUI
2. **Remote Instances**: Distribute instances across machines
3. **Plugins**: Extensibility for custom workflows
4. **AI Coordination**: Higher-level task orchestration
5. **Visual Diff**: See changes across instances in real-time

## Conclusion

The Agentree Multiplexer will transform how developers work with AI coding assistants by enabling parallel development workflows. By providing a robust, intuitive interface for managing multiple Claude Code instances, developers can leverage AI more effectively for complex, multi-faceted projects.