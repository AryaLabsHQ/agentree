# Multiplexer Implementation Guide

## Project Structure

```
agentree/
├── cmd/
│   └── multiplex.go              # Main command entry point
├── internal/
│   └── multiplex/
│       ├── multiplexer.go        # Core multiplexer logic
│       ├── process.go            # Process management
│       ├── events.go             # Event system
│       ├── config.go             # Configuration
│       ├── ui/
│       │   ├── controller.go     # UI controller
│       │   ├── layout.go         # Layout management
│       │   ├── sidebar.go        # Sidebar component
│       │   ├── output.go         # Output display
│       │   ├── statusbar.go      # Status bar
│       │   └── theme.go          # Theming
│       └── terminal/
│           ├── vt.go             # Virtual terminal
│           ├── parser.go         # ANSI parser
│           └── buffer.go         # Scrollback buffer
```

## Key Implementation Details

### 1. Process Management

```go
// internal/multiplex/process.go

type ProcessManager struct {
    instances map[string]*Instance
    mu        sync.RWMutex
    events    chan<- Event
}

type Instance struct {
    ID          string
    Worktree    *git.WorktreeInfo
    Cmd         *exec.Cmd
    PTY         *os.File
    VT          *terminal.VirtualTerminal
    State       InstanceState
    TokenUsage  *TokenTracker
    Buffer      *terminal.Buffer
    
    // Channels
    outputChan  chan []byte
    errorChan   chan error
    done        chan struct{}
}

func (pm *ProcessManager) StartInstance(id string) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    instance := pm.instances[id]
    if instance.State == StateRunning {
        return fmt.Errorf("instance already running")
    }
    
    // Create command
    cmd := exec.Command("claude", "code")
    cmd.Dir = instance.Worktree.Path
    cmd.Env = append(os.Environ(), 
        fmt.Sprintf("AGENTREE_INSTANCE_ID=%s", id),
        fmt.Sprintf("AGENTREE_WORKTREE=%s", instance.Worktree.Name),
    )
    
    // Create PTY
    pty, tty, err := pty.Open()
    if err != nil {
        return fmt.Errorf("failed to create pty: %w", err)
    }
    
    cmd.Stdin = tty
    cmd.Stdout = tty
    cmd.Stderr = tty
    
    // Start process
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start process: %w", err)
    }
    
    instance.Cmd = cmd
    instance.PTY = pty
    instance.State = StateRunning
    
    // Start output reader
    go pm.readOutput(instance)
    
    // Send event
    pm.events <- &ProcessEvent{
        InstanceID: id,
        Type:      ProcessStarted,
    }
    
    return nil
}
```

### 2. Event System

```go
// internal/multiplex/events.go

type EventDispatcher struct {
    events     chan Event
    handlers   map[EventType][]EventHandler
    mu         sync.RWMutex
    ctx        context.Context
    wg         sync.WaitGroup
}

type EventHandler func(Event) error

func NewEventDispatcher(ctx context.Context) *EventDispatcher {
    ed := &EventDispatcher{
        events:   make(chan Event, 100),
        handlers: make(map[EventType][]EventHandler),
        ctx:      ctx,
    }
    
    ed.wg.Add(1)
    go ed.run()
    
    return ed
}

func (ed *EventDispatcher) run() {
    defer ed.wg.Done()
    
    for {
        select {
        case <-ed.ctx.Done():
            return
        case event := <-ed.events:
            ed.dispatch(event)
        }
    }
}

func (ed *EventDispatcher) dispatch(event Event) {
    ed.mu.RLock()
    handlers := ed.handlers[event.Type()]
    ed.mu.RUnlock()
    
    for _, handler := range handlers {
        if err := handler(event); err != nil {
            // Log error but continue
            log.Printf("Event handler error: %v", err)
        }
    }
}

// Event types
const (
    EventProcessStarted EventType = iota
    EventProcessStopped
    EventProcessOutput
    EventProcessError
    EventUIResize
    EventUIFocus
    EventUIInput
    EventTokenUpdate
)
```

### 3. Terminal Emulation

```go
// internal/multiplex/terminal/vt.go

type VirtualTerminal struct {
    screen     *Screen
    parser     *ANSIParser
    scrollback *Buffer
    
    // Cursor
    cursorX    int
    cursorY    int
    
    // Attributes
    attrs      CellAttrs
    
    // Callbacks
    onUpdate   func(x, y, w, h int)
}

func NewVirtualTerminal(width, height int) *VirtualTerminal {
    vt := &VirtualTerminal{
        screen:     NewScreen(width, height),
        parser:     NewANSIParser(),
        scrollback: NewBuffer(10000), // 10k lines
    }
    
    vt.parser.OnSequence = vt.handleSequence
    vt.parser.OnText = vt.handleText
    
    return vt
}

func (vt *VirtualTerminal) Write(data []byte) (int, error) {
    vt.parser.Parse(data)
    return len(data), nil
}

func (vt *VirtualTerminal) handleText(text []byte) {
    for _, ch := range string(text) {
        vt.putChar(ch)
    }
}

func (vt *VirtualTerminal) handleSequence(seq *EscapeSequence) {
    switch seq.Type {
    case SeqCursorUp:
        vt.cursorY = max(0, vt.cursorY-seq.Params[0])
    case SeqCursorDown:
        vt.cursorY = min(vt.screen.Height-1, vt.cursorY+seq.Params[0])
    case SeqEraseLine:
        vt.eraseLine(vt.cursorY)
    case SeqSetGraphics:
        vt.updateAttrs(seq.Params)
    // ... more sequences
    }
}
```

### 4. UI Implementation

```go
// internal/multiplex/ui/controller.go

type UIController struct {
    screen      tcell.Screen
    layout      *Layout
    events      <-chan Event
    instances   []*InstanceView
    selected    int
    
    // Components
    sidebar     *Sidebar
    output      *OutputView
    statusbar   *StatusBar
}

func (ui *UIController) Run() error {
    // Main UI loop
    for {
        select {
        case ev := <-ui.events:
            ui.handleEvent(ev)
            
        case tcellEv := <-ui.screen.PollEvent():
            switch ev := tcellEv.(type) {
            case *tcell.EventKey:
                ui.handleKey(ev)
            case *tcell.EventMouse:
                ui.handleMouse(ev)
            case *tcell.EventResize:
                ui.resize()
            }
        }
        
        ui.draw()
    }
}

func (ui *UIController) handleKey(ev *tcell.EventKey) {
    switch ev.Key() {
    case tcell.KeyRune:
        switch ev.Rune() {
        case 'j':
            ui.selectNext()
        case 'k':
            ui.selectPrev()
        case 'q':
            ui.quit()
        case 'r':
            ui.restartSelected()
        }
    case tcell.KeyEnter:
        ui.focusSelected()
    case tcell.KeyTab:
        ui.cycleActive()
    }
}

func (ui *UIController) draw() {
    ui.screen.Clear()
    
    // Draw components
    ui.sidebar.Draw(ui.screen)
    ui.output.Draw(ui.screen)
    ui.statusbar.Draw(ui.screen)
    
    ui.screen.Show()
}
```

### 5. Token Tracking

```go
// internal/multiplex/token.go

type TokenTracker struct {
    mu           sync.RWMutex
    inputTokens  int64
    outputTokens int64
    requests     int64
    startTime    time.Time
    
    // Rate limiting
    limit        int64
    window       time.Duration
    buckets      []TokenBucket
}

func (tt *TokenTracker) AddUsage(input, output int64) {
    tt.mu.Lock()
    defer tt.mu.Unlock()
    
    tt.inputTokens += input
    tt.outputTokens += output
    tt.requests++
    
    // Add to current bucket
    now := time.Now()
    bucket := tt.getCurrentBucket(now)
    bucket.Input += input
    bucket.Output += output
    
    // Check limits
    if tt.isOverLimit() {
        // Send pause event
    }
}

func (tt *TokenTracker) GetStats() TokenStats {
    tt.mu.RLock()
    defer tt.mu.RUnlock()
    
    return TokenStats{
        InputTokens:  tt.inputTokens,
        OutputTokens: tt.outputTokens,
        TotalCost:   tt.calculateCost(),
        Duration:    time.Since(tt.startTime),
        RequestCount: tt.requests,
    }
}
```

## Integration Points

### 1. Claude Code Integration

Monitor Claude Code output for token usage:
```go
// Parse Claude output for token info
func parseTokenUsage(output string) (input, output int64) {
    // Look for patterns like:
    // "Tokens used: 1,234 (input: 500, output: 734)"
    re := regexp.MustCompile(`input:\s*(\d+).*output:\s*(\d+)`)
    matches := re.FindStringSubmatch(output)
    if len(matches) >= 3 {
        input, _ = strconv.ParseInt(matches[1], 10, 64)
        output, _ = strconv.ParseInt(matches[2], 10, 64)
    }
    return
}
```

### 2. Git Worktree Integration

```go
// Discover worktrees
func discoverWorktrees() ([]*git.WorktreeInfo, error) {
    repo, err := git.OpenRepository(".")
    if err != nil {
        return nil, err
    }
    
    worktrees, err := repo.ListWorktrees()
    if err != nil {
        return nil, err
    }
    
    return worktrees, nil
}
```

### 3. Configuration

```go
// Load multiplex config
type MultiplexConfig struct {
    AutoStart    bool                    `yaml:"auto_start"`
    TokenLimit   int64                   `yaml:"token_limit"`
    Theme        string                  `yaml:"theme"`
    Instances    []InstanceConfig        `yaml:"instances"`
    Shortcuts    map[string]string       `yaml:"shortcuts"`
}

type InstanceConfig struct {
    Worktree    string `yaml:"worktree"`
    AutoStart   bool   `yaml:"auto_start"`
    TokenLimit  int64  `yaml:"token_limit"`
    Environment []string `yaml:"environment"`
}
```

## Testing Strategy

### Unit Tests
- Event dispatcher
- Token tracking
- Terminal emulation
- Process lifecycle

### Integration Tests
- Full multiplexer flow
- Multi-instance coordination
- Error recovery
- Resource limits

### Manual Testing
- UI responsiveness
- Terminal compatibility
- Performance with many instances
- Edge cases (crashes, hangs)

## Performance Optimizations

1. **Lazy Rendering**: Only update changed regions
2. **Buffer Pooling**: Reuse buffers for output
3. **Event Batching**: Batch multiple updates
4. **Goroutine Management**: Limit concurrent operations

## Error Handling

1. **Process Crashes**: Auto-restart with backoff
2. **PTY Errors**: Graceful degradation
3. **UI Errors**: Fallback to simple mode
4. **Resource Exhaustion**: Pause instances

## Security

1. **Process Isolation**: Separate processes with limited permissions
2. **Input Sanitization**: Prevent escape sequence injection
3. **Token Security**: Never log or display full tokens
4. **File Access**: Respect project boundaries