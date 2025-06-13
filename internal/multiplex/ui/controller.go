// Package ui provides the user interface for the multiplexer
package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AryaLabsHQ/agentree/internal/multiplex"
	"github.com/gdamore/tcell/v2"
)

// Controller manages the multiplexer UI
type Controller struct {
	screen tcell.Screen
	events chan<- multiplex.Event
	
	// UI components
	sidebar  *Sidebar
	mainView *MainView
	statusBar *StatusBar
	
	// State
	instances      []*InstanceView
	focusedIndex   int
	mu             sync.RWMutex
	
	// Layout
	width          int
	height         int
	sidebarWidth   int
	
	// Input handling
	shortcuts      map[string]func()
	lastKeyTime    time.Time
}

// InstanceView represents a single instance in the UI
type InstanceView struct {
	ID         string
	Worktree   string
	State      multiplex.InstanceState
	TokenUsage multiplex.TokenUsage
	Buffer     []string
	ScrollPos  int
	LastUpdate time.Time
}

// NewController creates a new UI controller
func NewController(screen tcell.Screen, events chan<- multiplex.Event) *Controller {
	c := &Controller{
		screen:       screen,
		events:       events,
		instances:    make([]*InstanceView, 0),
		sidebarWidth: 25,
		shortcuts:    make(map[string]func()),
	}
	
	// Get initial dimensions
	c.width, c.height = screen.Size()
	
	// Create components
	c.sidebar = NewSidebar(c.sidebarWidth, c.height)
	c.mainView = NewMainView(c.width-c.sidebarWidth, c.height)
	c.statusBar = NewStatusBar(c.width)
	
	// Setup shortcuts
	c.setupShortcuts()
	
	return c
}

// Run starts the UI controller
func (c *Controller) Run(ctx context.Context) error {
	// Clear screen
	c.screen.Clear()
	
	// Initial draw
	c.draw()
	
	// Event loop
	for {
		select {
		case <-ctx.Done():
			return nil
			
		default:
			// Poll for events with timeout
			ev := c.screen.PollEvent()
			if ev == nil {
				continue
			}
			
			switch event := ev.(type) {
			case *tcell.EventResize:
				c.handleResize(event)
				
			case *tcell.EventKey:
				if !c.handleKey(event) {
					return nil // Quit requested
				}
				
			case *tcell.EventMouse:
				c.handleMouse(event)
			}
		}
	}
}

// Stop stops the UI controller
func (c *Controller) Stop() {
	c.screen.Fini()
}

// AddInstance adds an instance to the UI
func (c *Controller) AddInstance(id, worktree string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	instance := &InstanceView{
		ID:       id,
		Worktree: worktree,
		State:    multiplex.StateIdle,
		Buffer:   make([]string, 0),
	}
	
	c.instances = append(c.instances, instance)
	
	// Update sidebar
	c.sidebar.UpdateInstances(c.instances)
	
	// Redraw
	c.draw()
}

// UpdateInstanceState updates the state of an instance
func (c *Controller) UpdateInstanceState(id string, state multiplex.InstanceState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, instance := range c.instances {
		if instance.ID == id {
			instance.State = state
			instance.LastUpdate = time.Now()
			break
		}
	}
	
	// Update sidebar
	c.sidebar.UpdateInstances(c.instances)
	
	// Redraw
	c.draw()
}

// AddOutput adds output to an instance buffer
func (c *Controller) AddOutput(id string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, instance := range c.instances {
		if instance.ID == id {
			// Parse output into lines
			// TODO: Handle ANSI codes properly
			lines := string(data) // Simplified for stub
			instance.Buffer = append(instance.Buffer, lines)
			
			// Limit buffer size
			if len(instance.Buffer) > 10000 {
				instance.Buffer = instance.Buffer[len(instance.Buffer)-10000:]
			}
			
			instance.LastUpdate = time.Now()
			break
		}
	}
	
	// Update main view if this is the focused instance
	if c.focusedIndex < len(c.instances) && c.instances[c.focusedIndex].ID == id {
		c.mainView.SetContent(c.instances[c.focusedIndex].Buffer)
	}
	
	// Redraw
	c.draw()
}

// UpdateTokenUsage updates token usage for an instance
func (c *Controller) UpdateTokenUsage(id string, usage multiplex.TokenUsage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for _, instance := range c.instances {
		if instance.ID == id {
			instance.TokenUsage = usage
			break
		}
	}
	
	// Update status bar
	c.updateStatusBar()
	
	// Redraw
	c.draw()
}

// draw renders the entire UI
func (c *Controller) draw() {
	c.screen.Clear()
	
	// Draw components
	c.sidebar.Draw(c.screen, 0, 0)
	c.mainView.Draw(c.screen, c.sidebarWidth, 0)
	c.statusBar.Draw(c.screen, 0, c.height-1)
	
	c.screen.Show()
}

// handleResize handles terminal resize events
func (c *Controller) handleResize(ev *tcell.EventResize) {
	c.width, c.height = ev.Size()
	
	// Resize components
	c.sidebar.Resize(c.sidebarWidth, c.height-1)
	c.mainView.Resize(c.width-c.sidebarWidth, c.height-1)
	c.statusBar.Resize(c.width)
	
	// Send resize event
	c.events <- multiplex.NewUIResizeEvent(c.width, c.height)
	
	// Redraw
	c.draw()
}

// handleKey handles keyboard input
func (c *Controller) handleKey(ev *tcell.EventKey) bool {
	c.lastKeyTime = time.Now()
	
	// Check shortcuts
	key := ev.Key()
	mod := ev.Modifiers()
	
	// Handle special keys
	switch key {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		c.events <- multiplex.NewQuitEvent()
		return false
		
	case tcell.KeyUp:
		c.navigateUp()
		return true
		
	case tcell.KeyDown:
		c.navigateDown()
		return true
		
	case tcell.KeyEnter:
		c.focusCurrent()
		return true
		
	case tcell.KeyRune:
		rune := ev.Rune()
		
		// Check character shortcuts
		switch rune {
		case 'q', 'Q':
			if mod&tcell.ModCtrl == 0 {
				c.events <- multiplex.NewQuitEvent()
				return false
			}
			
		case 's', 'S':
			c.startCurrent()
			
		case 'x', 'X':
			c.stopCurrent()
			
		case 'r', 'R':
			c.restartCurrent()
			
		case 'c', 'C':
			c.clearCurrent()
			
		case '?':
			c.showHelp()
			
		case 'j':
			c.navigateDown()
			
		case 'k':
			c.navigateUp()
		}
	}
	
	return true
}

// handleMouse handles mouse input
func (c *Controller) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	
	// Check if click is in sidebar
	if x < c.sidebarWidth {
		// Calculate which instance was clicked
		instanceIndex := y - 2 // Account for header
		if instanceIndex >= 0 && instanceIndex < len(c.instances) {
			c.focusedIndex = instanceIndex
			c.sidebar.SetFocused(c.focusedIndex)
			c.focusCurrent()
			c.draw()
		}
	}
}

// Navigation methods

func (c *Controller) navigateUp() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.focusedIndex > 0 {
		c.focusedIndex--
		c.sidebar.SetFocused(c.focusedIndex)
		c.draw()
	}
}

func (c *Controller) navigateDown() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.focusedIndex < len(c.instances)-1 {
		c.focusedIndex++
		c.sidebar.SetFocused(c.focusedIndex)
		c.draw()
	}
}

func (c *Controller) focusCurrent() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.focusedIndex < len(c.instances) {
		instance := c.instances[c.focusedIndex]
		c.mainView.SetContent(instance.Buffer)
		c.events <- &multiplex.UIFocusEvent{
			BaseEvent:  multiplex.BaseEvent{EventType: multiplex.EventUIFocus, Time: time.Now()},
			InstanceID: instance.ID,
		}
	}
}

// Control methods

func (c *Controller) startCurrent() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.focusedIndex < len(c.instances) {
		instance := c.instances[c.focusedIndex]
		c.events <- &multiplex.UIControlEvent{
			BaseEvent:  multiplex.BaseEvent{EventType: multiplex.EventUIControl, Time: time.Now()},
			InstanceID: instance.ID,
			Action:     "start",
		}
	}
}

func (c *Controller) stopCurrent() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.focusedIndex < len(c.instances) {
		instance := c.instances[c.focusedIndex]
		c.events <- &multiplex.UIControlEvent{
			BaseEvent:  multiplex.BaseEvent{EventType: multiplex.EventUIControl, Time: time.Now()},
			InstanceID: instance.ID,
			Action:     "stop",
		}
	}
}

func (c *Controller) restartCurrent() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.focusedIndex < len(c.instances) {
		instance := c.instances[c.focusedIndex]
		c.events <- &multiplex.UIControlEvent{
			BaseEvent:  multiplex.BaseEvent{EventType: multiplex.EventUIControl, Time: time.Now()},
			InstanceID: instance.ID,
			Action:     "restart",
		}
	}
}

func (c *Controller) clearCurrent() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.focusedIndex < len(c.instances) {
		instance := c.instances[c.focusedIndex]
		instance.Buffer = make([]string, 0)
		c.mainView.SetContent(instance.Buffer)
		c.draw()
	}
}

func (c *Controller) showHelp() {
	// TODO: Implement help overlay
	helpText := []string{
		"Keyboard Shortcuts:",
		"",
		"  q     - Quit",
		"  s     - Start instance",
		"  x     - Stop instance",
		"  r     - Restart instance",
		"  c     - Clear output",
		"  j/↓   - Navigate down",
		"  k/↑   - Navigate up",
		"  Enter - Focus instance",
		"  ?     - Show this help",
	}
	
	c.mainView.SetContent(helpText)
	c.draw()
}

// setupShortcuts configures keyboard shortcuts
func (c *Controller) setupShortcuts() {
	// Shortcuts are handled directly in handleKey for now
}

// updateStatusBar updates the status bar content
func (c *Controller) updateStatusBar() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Calculate total token usage
	var totalInput, totalOutput int64
	for _, instance := range c.instances {
		totalInput += instance.TokenUsage.InputTokens
		totalOutput += instance.TokenUsage.OutputTokens
	}
	
	// Update status
	status := fmt.Sprintf("Tokens: %d/%d | Instances: %d | Press ? for help",
		totalInput+totalOutput, 100000, len(c.instances))
	
	c.statusBar.SetStatus(status)
}