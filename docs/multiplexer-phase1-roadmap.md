# Phase 1: Core Infrastructure - Detailed Roadmap

## Week 2-3 Sprint Plan

### Sprint Overview
**Goal**: Build foundational multiplexer that can launch and display a single Claude Code instance
**Duration**: 2 weeks
**Deliverable**: Working prototype with basic TUI

## Week 2: Foundation

### Day 1-2: Project Setup & Command Structure
```bash
# New files to create:
cmd/multiplex.go              # Main command
internal/multiplex/
├── multiplexer.go           # Core logic
├── process.go               # Process management
└── config.go                # Configuration
```

**Tasks**:
1. Create `multiplex` command with basic flags
2. Set up command aliases (`mx`)
3. Add help text and examples
4. Create configuration structure
5. Write command tests

**Code Skeleton**:
```go
// cmd/multiplex.go
var multiplexCmd = &cobra.Command{
    Use:     "multiplex [worktrees...]",
    Aliases: []string{"mx"},
    Short:   "Run multiple Claude Code instances",
    RunE:    runMultiplex,
}

func init() {
    rootCmd.AddCommand(multiplexCmd)
    multiplexCmd.Flags().BoolP("all", "a", false, "Launch all worktrees")
    multiplexCmd.Flags().String("config", "", "Config file path")
}
```

### Day 3-4: Basic Process Management
**Goal**: Launch and manage a single process

**Tasks**:
1. Implement `ProcessManager` struct
2. Create process launching with PTY
3. Handle process lifecycle
4. Capture output
5. Clean shutdown

**Key Functions**:
```go
type ProcessManager struct {
    process *Instance
    pty     *os.File
}

func (pm *ProcessManager) Start(worktree string) error
func (pm *ProcessManager) Stop() error
func (pm *ProcessManager) ReadOutput() ([]byte, error)
```

### Day 5: Event System Foundation
**Goal**: Basic event handling

**Tasks**:
1. Define event types
2. Create event dispatcher
3. Implement event handlers
4. Add logging

**Event Types**:
- ProcessStarted
- ProcessOutput
- ProcessError
- ProcessExit

## Week 3: UI Implementation

### Day 6-7: TUI Framework Setup
**Goal**: Basic tcell application

**Tasks**:
1. Initialize tcell screen
2. Create main event loop
3. Handle keyboard input
4. Implement clean exit
5. Add resize handling

**Basic Structure**:
```go
type UI struct {
    screen tcell.Screen
    quit   chan struct{}
}

func (ui *UI) Run() error {
    for {
        select {
        case ev := <-ui.screen.PollEvent():
            // Handle tcell events
        case <-ui.quit:
            return nil
        }
    }
}
```

### Day 8-9: Output Display
**Goal**: Show process output in TUI

**Tasks**:
1. Create output view component
2. Implement scrolling text
3. Add basic ANSI color support
4. Handle line wrapping
5. Buffer management

**Challenges**:
- Efficient rendering
- Memory management
- Handling large outputs

### Day 10: Integration & Testing
**Goal**: Connect all components

**Tasks**:
1. Wire up process manager to UI
2. Connect event system
3. Add configuration loading
4. Create integration tests
5. Manual testing

**Test Scenarios**:
- Launch process successfully
- Handle process crash
- Keyboard navigation
- Clean shutdown
- Configuration loading

## Technical Decisions to Make

### 1. Terminal Emulation Approach
**Options**:
- **A**: Use simple output capture (no ANSI)
- **B**: Basic ANSI parsing (colors only)
- **C**: Full terminal emulation

**Recommendation**: Start with B, plan for C

### 2. Process Communication
**Options**:
- **A**: PTY only
- **B**: PTY + pipes
- **C**: PTY + IPC

**Recommendation**: Start with A

### 3. Configuration Format
**Options**:
- **A**: Extend .agentreerc
- **B**: Separate multiplex.yml
- **C**: JSON configuration

**Recommendation**: B for flexibility

## Dependencies to Add

```go
// go.mod additions
require (
    github.com/gdamore/tcell/v2 v2.6.0
    github.com/creack/pty v1.1.18
)
```

## File Structure After Phase 1

```
agentree/
├── cmd/
│   ├── multiplex.go         # New
│   └── ...
├── internal/
│   ├── multiplex/          # New
│   │   ├── multiplexer.go
│   │   ├── process.go
│   │   ├── events.go
│   │   ├── config.go
│   │   └── ui.go
│   └── ...
├── docs/
│   ├── multiplexer-*.md    # Documentation
│   └── ...
└── examples/
    └── multiplex.yml       # Example config
```

## Definition of Done

### Phase 1 is complete when:
1. ✅ Can launch Claude Code in a worktree
2. ✅ Shows output in TUI
3. ✅ Handles basic keyboard input (q to quit)
4. ✅ Graceful shutdown
5. ✅ Configuration loading works
6. ✅ Basic tests pass
7. ✅ Documentation updated

## Risks & Mitigations

### Risk 1: PTY Complexity
**Issue**: PTY handling is OS-specific
**Mitigation**: 
- Use well-tested `creack/pty` library
- Start with Linux/macOS, add Windows later
- Have fallback to simple exec

### Risk 2: TUI Framework Issues
**Issue**: tcell might be too low-level
**Mitigation**:
- Consider bubbletea as alternative
- Build abstraction layer
- Keep UI simple initially

### Risk 3: Claude Code Integration
**Issue**: Unknown how Claude Code behaves in PTY
**Mitigation**:
- Test with simple commands first
- Build mock for testing
- Have debug mode with raw output

## Daily Checklist

### Each day:
- [ ] Write failing tests first
- [ ] Implement minimal code to pass
- [ ] Refactor for clarity
- [ ] Update documentation
- [ ] Commit working code
- [ ] Note any blockers

## Success Metrics

### By end of Phase 1:
- **Lines of Code**: ~1500-2000
- **Test Coverage**: >70%
- **Features**: 5/5 complete
- **Documentation**: Updated
- **Demo**: Working prototype

## Next Phase Preview

Phase 2 will add:
- Multiple process support
- Process switching UI
- Better error handling
- Instance state management

But first, let's nail the single-instance case!