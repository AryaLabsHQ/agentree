package ui

import (
	"fmt"
	"strings"

	"github.com/AryaLabsHQ/agentree/internal/multiplex"
	"github.com/gdamore/tcell/v2"
)

// Component is the base interface for UI components
type Component interface {
	Draw(screen tcell.Screen, x, y int)
	Resize(width, height int)
}

// Sidebar shows the list of instances
type Sidebar struct {
	width     int
	height    int
	instances []*InstanceView
	focused   int
}

// NewSidebar creates a new sidebar
func NewSidebar(width, height int) *Sidebar {
	return &Sidebar{
		width:   width,
		height:  height,
		focused: -1,
	}
}

// Draw renders the sidebar
func (s *Sidebar) Draw(screen tcell.Screen, x, y int) {
	// Draw border
	s.drawBorder(screen, x, y)
	
	// Draw title
	title := " Instances "
	titleX := x + (s.width-len(title))/2
	s.drawText(screen, titleX, y, title, tcell.StyleDefault.Bold(true))
	
	// Draw instances
	for i, instance := range s.instances {
		if i >= s.height-3 { // Leave room for border and title
			break
		}
		
		lineY := y + i + 2
		style := tcell.StyleDefault
		
		// Highlight focused instance
		if i == s.focused {
			style = style.Reverse(true)
		}
		
		// Set color based on state
		switch instance.State {
		case multiplex.StateRunning:
			style = style.Foreground(tcell.ColorGreen)
		case multiplex.StateThinking:
			style = style.Foreground(tcell.ColorYellow)
		case multiplex.StateStopped:
			style = style.Foreground(tcell.ColorRed)
		case multiplex.StateCrashed:
			style = style.Foreground(tcell.ColorRed).Bold(true)
		}
		
		// Format line
		statusChar := s.getStatusChar(instance.State)
		line := fmt.Sprintf(" %s %s", statusChar, s.truncate(instance.Worktree, s.width-4))
		
		// Draw line
		s.drawLine(screen, x, lineY, line, s.width, style)
	}
}

// Resize updates the sidebar dimensions
func (s *Sidebar) Resize(width, height int) {
	s.width = width
	s.height = height
}

// UpdateInstances updates the instance list
func (s *Sidebar) UpdateInstances(instances []*InstanceView) {
	s.instances = instances
}

// SetFocused sets the focused instance
func (s *Sidebar) SetFocused(index int) {
	s.focused = index
}

// Helper methods

func (s *Sidebar) drawBorder(screen tcell.Screen, x, y int) {
	style := tcell.StyleDefault
	
	// Top border
	screen.SetContent(x, y, '┌', nil, style)
	for i := 1; i < s.width-1; i++ {
		screen.SetContent(x+i, y, '─', nil, style)
	}
	screen.SetContent(x+s.width-1, y, '┐', nil, style)
	
	// Side borders
	for i := 1; i < s.height-1; i++ {
		screen.SetContent(x, y+i, '│', nil, style)
		screen.SetContent(x+s.width-1, y+i, '│', nil, style)
	}
	
	// Bottom border
	screen.SetContent(x, y+s.height-1, '└', nil, style)
	for i := 1; i < s.width-1; i++ {
		screen.SetContent(x+i, y+s.height-1, '─', nil, style)
	}
	screen.SetContent(x+s.width-1, y+s.height-1, '┘', nil, style)
}

func (s *Sidebar) drawText(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		screen.SetContent(x+i, y, ch, nil, style)
	}
}

func (s *Sidebar) drawLine(screen tcell.Screen, x, y int, text string, width int, style tcell.Style) {
	// Clear line first
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}
	
	// Draw text
	for i, ch := range text {
		if i >= width {
			break
		}
		screen.SetContent(x+i, y, ch, nil, style)
	}
}

func (s *Sidebar) getStatusChar(state multiplex.InstanceState) string {
	switch state {
	case multiplex.StateIdle:
		return "○"
	case multiplex.StateStarting:
		return "◐"
	case multiplex.StateRunning:
		return "●"
	case multiplex.StateThinking:
		return "◍"
	case multiplex.StateStopping:
		return "◑"
	case multiplex.StateStopped:
		return "◯"
	case multiplex.StateCrashed:
		return "✗"
	default:
		return "?"
	}
}

func (s *Sidebar) truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-1] + "…"
}

// MainView displays the output of the selected instance
type MainView struct {
	width   int
	height  int
	content []string
	scroll  int
}

// NewMainView creates a new main view
func NewMainView(width, height int) *MainView {
	return &MainView{
		width:  width,
		height: height,
	}
}

// Draw renders the main view
func (m *MainView) Draw(screen tcell.Screen, x, y int) {
	// Clear area
	style := tcell.StyleDefault
	for row := 0; row < m.height; row++ {
		for col := 0; col < m.width; col++ {
			screen.SetContent(x+col, y+row, ' ', nil, style)
		}
	}
	
	// Draw content
	visibleLines := m.height
	startLine := m.scroll
	
	for i := 0; i < visibleLines && startLine+i < len(m.content); i++ {
		line := m.content[startLine+i]
		
		// Draw line (truncate if needed)
		for j, ch := range line {
			if j >= m.width {
				break
			}
			screen.SetContent(x+j, y+i, ch, nil, style)
		}
	}
}

// Resize updates the main view dimensions
func (m *MainView) Resize(width, height int) {
	m.width = width
	m.height = height
}

// SetContent updates the content
func (m *MainView) SetContent(content []string) {
	m.content = content
	// Auto-scroll to bottom
	if len(content) > m.height {
		m.scroll = len(content) - m.height
	} else {
		m.scroll = 0
	}
}

// ScrollUp scrolls the view up
func (m *MainView) ScrollUp(lines int) {
	m.scroll -= lines
	if m.scroll < 0 {
		m.scroll = 0
	}
}

// ScrollDown scrolls the view down
func (m *MainView) ScrollDown(lines int) {
	m.scroll += lines
	maxScroll := len(m.content) - m.height
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}

// StatusBar shows system status
type StatusBar struct {
	width  int
	status string
}

// NewStatusBar creates a new status bar
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		width: width,
	}
}

// Draw renders the status bar
func (s *StatusBar) Draw(screen tcell.Screen, x, y int) {
	style := tcell.StyleDefault.Background(tcell.ColorDarkGray)
	
	// Clear status bar
	for i := 0; i < s.width; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}
	
	// Draw status
	for i, ch := range s.status {
		if i >= s.width {
			break
		}
		screen.SetContent(x+i, y, ch, nil, style)
	}
	
	// Draw time on the right
	time := fmt.Sprintf(" %s ", strings.ToUpper(fmt.Sprintf("%d:%02d", 0, 0))) // Placeholder
	timeX := x + s.width - len(time)
	for i, ch := range time {
		screen.SetContent(timeX+i, y, ch, nil, style)
	}
}

// Resize updates the status bar width
func (s *StatusBar) Resize(width int) {
	s.width = width
}

// SetStatus updates the status message
func (s *StatusBar) SetStatus(status string) {
	s.status = status
}