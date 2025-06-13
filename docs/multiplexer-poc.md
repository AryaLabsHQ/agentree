# Multiplexer Proof of Concept Plan

## Objective
Build a minimal proof-of-concept in 2-3 days to validate technical approach and identify potential blockers.

## POC Scope

### What it WILL do:
- Launch a process in a PTY
- Display output in a basic TUI
- Handle keyboard input (q to quit)
- Demonstrate tcell framework usage

### What it WON'T do:
- Multiple processes
- Fancy UI layout
- Configuration
- Error handling
- Terminal emulation

## Implementation Steps

### Step 1: Basic PTY Process (Day 1 Morning)
**File**: `poc/pty-test.go`

```go
package main

import (
    "io"
    "log"
    "os"
    "os/exec"
    
    "github.com/creack/pty"
)

func main() {
    // Launch simple command in PTY
    cmd := exec.Command("bash", "-c", "while true; do date; sleep 1; done")
    
    ptmx, err := pty.Start(cmd)
    if err != nil {
        log.Fatal(err)
    }
    defer ptmx.Close()
    
    // Copy PTY output to stdout
    go io.Copy(os.Stdout, ptmx)
    
    // Wait for process
    cmd.Wait()
}
```

**Validation**:
- Can launch process in PTY ✓
- Can capture output ✓
- Handles ANSI codes ✓

### Step 2: Basic TUI (Day 1 Afternoon)
**File**: `poc/tui-test.go`

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/gdamore/tcell/v2"
)

func main() {
    // Initialize screen
    screen, err := tcell.NewScreen()
    if err != nil {
        log.Fatal(err)
    }
    
    if err := screen.Init(); err != nil {
        log.Fatal(err)
    }
    defer screen.Fini()
    
    // Main loop
    for {
        screen.Clear()
        drawText(screen, 1, 1, "Multiplexer POC - Press 'q' to quit")
        screen.Show()
        
        ev := screen.PollEvent()
        switch ev := ev.(type) {
        case *tcell.EventKey:
            if ev.Rune() == 'q' {
                return
            }
        case *tcell.EventResize:
            screen.Sync()
        }
    }
}

func drawText(s tcell.Screen, x, y int, text string) {
    style := tcell.StyleDefault
    for i, r := range text {
        s.SetContent(x+i, y, r, nil, style)
    }
}
```

**Validation**:
- TUI framework works ✓
- Can handle input ✓
- Can draw text ✓

### Step 3: Combine PTY + TUI (Day 2)
**File**: `poc/multiplex-poc.go`

```go
package main

import (
    "bufio"
    "fmt"
    "os/exec"
    "strings"
    "sync"
    
    "github.com/creack/pty"
    "github.com/gdamore/tcell/v2"
)

type MultiplexPOC struct {
    screen    tcell.Screen
    ptyFile   *os.File
    cmd       *exec.Cmd
    output    []string
    outputMu  sync.Mutex
    scrollPos int
}

func (m *MultiplexPOC) Run() error {
    // Start process
    if err := m.startProcess(); err != nil {
        return err
    }
    
    // Start output reader
    go m.readOutput()
    
    // Main UI loop
    for {
        m.draw()
        
        ev := m.screen.PollEvent()
        switch ev := ev.(type) {
        case *tcell.EventKey:
            switch ev.Key() {
            case tcell.KeyRune:
                if ev.Rune() == 'q' {
                    return nil
                }
            case tcell.KeyUp:
                m.scroll(-1)
            case tcell.KeyDown:
                m.scroll(1)
            }
        case *tcell.EventResize:
            m.screen.Sync()
        }
    }
}

func (m *MultiplexPOC) startProcess() error {
    m.cmd = exec.Command("bash", "-c", `
        echo "Starting process..."
        for i in {1..100}; do
            echo "Line $i: $(date)"
            sleep 0.5
        done
    `)
    
    var err error
    m.ptyFile, err = pty.Start(m.cmd)
    return err
}

func (m *MultiplexPOC) readOutput() {
    scanner := bufio.NewScanner(m.ptyFile)
    for scanner.Scan() {
        m.outputMu.Lock()
        m.output = append(m.output, scanner.Text())
        m.outputMu.Unlock()
    }
}

func (m *MultiplexPOC) draw() {
    m.screen.Clear()
    
    // Header
    m.drawText(0, 0, "=== Multiplexer POC === (q: quit, ↑↓: scroll)")
    
    // Output area
    m.outputMu.Lock()
    _, height := m.screen.Size()
    startLine := m.scrollPos
    endLine := startLine + height - 3
    
    for i := startLine; i < endLine && i < len(m.output); i++ {
        m.drawText(0, i-startLine+2, m.output[i])
    }
    m.outputMu.Unlock()
    
    m.screen.Show()
}
```

**Validation**:
- PTY + TUI integration ✓
- Output buffering ✓
- Basic scrolling ✓

### Step 4: Claude Code Test (Day 2 Afternoon)
Replace test command with actual Claude Code:

```go
m.cmd = exec.Command("claude", "code")
m.cmd.Dir = "/path/to/worktree"
```

**Test scenarios**:
1. Launch Claude Code successfully
2. Capture colored output
3. Handle interactive prompts
4. Clean shutdown

## Evaluation Criteria

### Technical Feasibility
- [ ] PTY works correctly on target OS
- [ ] TUI responsive with process output
- [ ] Memory usage acceptable
- [ ] CPU usage acceptable

### Implementation Complexity
- [ ] Code is maintainable
- [ ] Abstractions are clear
- [ ] Error handling is manageable
- [ ] Testing is possible

### Performance
- [ ] UI remains responsive
- [ ] Output doesn't lag
- [ ] Scrolling is smooth
- [ ] Memory doesn't leak

## Decision Points

### 1. Continue with tcell?
**Pros**: Low-level control, efficient
**Cons**: More code needed
**Alternative**: Bubbletea (higher level)

### 2. Terminal emulation approach?
**Current**: Raw output capture
**Alternative**: Full VT100 emulation
**Decision**: Based on Claude Code output

### 3. Process management strategy?
**Current**: Simple PTY
**Alternative**: More complex IPC
**Decision**: Based on coordination needs

## Risk Assessment

### High Risk
- Terminal emulation complexity
- Claude Code compatibility
- Cross-platform support

### Medium Risk
- Performance with large outputs
- Memory management
- Error recovery

### Low Risk
- Basic TUI functionality
- Process launching
- Configuration

## Next Steps After POC

### If Successful:
1. Begin Phase 1 implementation
2. Set up proper project structure
3. Add comprehensive tests
4. Start documentation

### If Issues Found:
1. Evaluate alternatives
2. Adjust architecture
3. Simplify scope
4. Update timeline

## POC Deliverables

1. **Working code** in `poc/` directory
2. **Test results** document
3. **Performance metrics**
4. **Decision recommendations**
5. **Updated risk assessment**

## Timeline

### Day 1
- Morning: PTY process test
- Afternoon: Basic TUI test
- Evening: Initial integration

### Day 2
- Morning: Complete integration
- Afternoon: Claude Code testing
- Evening: Performance testing

### Day 3
- Morning: Edge case testing
- Afternoon: Documentation
- Evening: Decision meeting

This POC will give us confidence in the technical approach and identify any major blockers before committing to the full implementation.