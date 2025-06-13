// Package terminal provides virtual terminal emulation for the multiplexer
package terminal

import (
	"bytes"
	"sync"
)

// Cell represents a single character cell in the terminal
type Cell struct {
	Rune       rune
	Foreground Color
	Background Color
	Attributes Attribute
}

// Color represents a terminal color
type Color uint32

// Standard colors
const (
	ColorDefault Color = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// Attribute represents text attributes
type Attribute uint16

// Text attributes
const (
	AttrBold Attribute = 1 << iota
	AttrUnderline
	AttrReverse
	AttrBlink
	AttrDim
	AttrItalic
	AttrStrikethrough
)

// VirtualTerminal represents a virtual terminal that interprets ANSI escape sequences
type VirtualTerminal struct {
	// Dimensions
	width  int
	height int
	
	// Screen buffer
	cells [][]Cell
	
	// Cursor position
	cursorX int
	cursorY int
	
	// Scrollback buffer
	scrollback [][]Cell
	maxScrollback int
	
	// State
	mu sync.RWMutex
	
	// Parser state
	parser *ANSIParser
	
	// Current attributes
	currentFg   Color
	currentBg   Color
	currentAttr Attribute
}

// NewVirtualTerminal creates a new virtual terminal
func NewVirtualTerminal(width, height int) *VirtualTerminal {
	vt := &VirtualTerminal{
		width:         width,
		height:        height,
		cells:         make([][]Cell, height),
		scrollback:    make([][]Cell, 0),
		maxScrollback: 10000,
		currentFg:     ColorDefault,
		currentBg:     ColorDefault,
	}
	
	// Initialize cells
	for i := 0; i < height; i++ {
		vt.cells[i] = make([]Cell, width)
		for j := 0; j < width; j++ {
			vt.cells[i][j] = Cell{Rune: ' '}
		}
	}
	
	// Create parser
	vt.parser = NewANSIParser(vt)
	
	return vt
}

// Resize changes the terminal dimensions
func (vt *VirtualTerminal) Resize(width, height int) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	
	// Create new cell buffer
	newCells := make([][]Cell, height)
	for i := 0; i < height; i++ {
		newCells[i] = make([]Cell, width)
		for j := 0; j < width; j++ {
			newCells[i][j] = Cell{Rune: ' '}
		}
	}
	
	// Copy existing cells
	copyHeight := height
	if vt.height < height {
		copyHeight = vt.height
	}
	copyWidth := width
	if vt.width < width {
		copyWidth = vt.width
	}
	
	for i := 0; i < copyHeight; i++ {
		for j := 0; j < copyWidth; j++ {
			newCells[i][j] = vt.cells[i][j]
		}
	}
	
	vt.width = width
	vt.height = height
	vt.cells = newCells
	
	// Adjust cursor position
	if vt.cursorX >= width {
		vt.cursorX = width - 1
	}
	if vt.cursorY >= height {
		vt.cursorY = height - 1
	}
}

// Write processes input data and updates the terminal state
func (vt *VirtualTerminal) Write(data []byte) (int, error) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	
	// Feed data to parser
	vt.parser.Parse(data)
	
	return len(data), nil
}

// GetScreen returns the current screen content
func (vt *VirtualTerminal) GetScreen() [][]Cell {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	
	// Deep copy cells
	screen := make([][]Cell, vt.height)
	for i := 0; i < vt.height; i++ {
		screen[i] = make([]Cell, vt.width)
		copy(screen[i], vt.cells[i])
	}
	
	return screen
}

// GetLine returns a specific line as a string
func (vt *VirtualTerminal) GetLine(y int) string {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	
	if y < 0 || y >= vt.height {
		return ""
	}
	
	var buf bytes.Buffer
	for x := 0; x < vt.width; x++ {
		buf.WriteRune(vt.cells[y][x].Rune)
	}
	
	return buf.String()
}

// GetCursorPosition returns the current cursor position
func (vt *VirtualTerminal) GetCursorPosition() (x, y int) {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	return vt.cursorX, vt.cursorY
}

// Clear clears the terminal screen
func (vt *VirtualTerminal) Clear() {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	
	for y := 0; y < vt.height; y++ {
		for x := 0; x < vt.width; x++ {
			vt.cells[y][x] = Cell{Rune: ' '}
		}
	}
	
	vt.cursorX = 0
	vt.cursorY = 0
}

// PutRune places a rune at the current cursor position
func (vt *VirtualTerminal) PutRune(r rune) {
	if vt.cursorY >= vt.height {
		vt.scroll()
		vt.cursorY = vt.height - 1
	}
	
	if vt.cursorX < vt.width {
		vt.cells[vt.cursorY][vt.cursorX] = Cell{
			Rune:       r,
			Foreground: vt.currentFg,
			Background: vt.currentBg,
			Attributes: vt.currentAttr,
		}
		vt.cursorX++
	}
	
	// Handle line wrap
	if vt.cursorX >= vt.width {
		vt.cursorX = 0
		vt.cursorY++
	}
}

// MoveCursor moves the cursor to a specific position
func (vt *VirtualTerminal) MoveCursor(x, y int) {
	if x < 0 {
		x = 0
	}
	if x >= vt.width {
		x = vt.width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= vt.height {
		y = vt.height - 1
	}
	
	vt.cursorX = x
	vt.cursorY = y
}

// NewLine moves cursor to the beginning of the next line
func (vt *VirtualTerminal) NewLine() {
	vt.cursorX = 0
	vt.cursorY++
	
	if vt.cursorY >= vt.height {
		vt.scroll()
		vt.cursorY = vt.height - 1
	}
}

// CarriageReturn moves cursor to the beginning of the current line
func (vt *VirtualTerminal) CarriageReturn() {
	vt.cursorX = 0
}

// SetForeground sets the current foreground color
func (vt *VirtualTerminal) SetForeground(color Color) {
	vt.currentFg = color
}

// SetBackground sets the current background color
func (vt *VirtualTerminal) SetBackground(color Color) {
	vt.currentBg = color
}

// SetAttribute sets text attributes
func (vt *VirtualTerminal) SetAttribute(attr Attribute) {
	vt.currentAttr = attr
}

// ResetAttributes resets all attributes to default
func (vt *VirtualTerminal) ResetAttributes() {
	vt.currentFg = ColorDefault
	vt.currentBg = ColorDefault
	vt.currentAttr = 0
}

// scroll scrolls the terminal up by one line
func (vt *VirtualTerminal) scroll() {
	// Add top line to scrollback
	if len(vt.scrollback) >= vt.maxScrollback {
		vt.scrollback = vt.scrollback[1:]
	}
	vt.scrollback = append(vt.scrollback, vt.cells[0])
	
	// Shift lines up
	for i := 0; i < vt.height-1; i++ {
		vt.cells[i] = vt.cells[i+1]
	}
	
	// Clear bottom line
	vt.cells[vt.height-1] = make([]Cell, vt.width)
	for x := 0; x < vt.width; x++ {
		vt.cells[vt.height-1][x] = Cell{Rune: ' '}
	}
}

// GetScrollback returns the scrollback buffer
func (vt *VirtualTerminal) GetScrollback() [][]Cell {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	
	// Deep copy scrollback
	scrollback := make([][]Cell, len(vt.scrollback))
	for i, line := range vt.scrollback {
		scrollback[i] = make([]Cell, len(line))
		copy(scrollback[i], line)
	}
	
	return scrollback
}

// String returns the terminal content as a string
func (vt *VirtualTerminal) String() string {
	vt.mu.RLock()
	defer vt.mu.RUnlock()
	
	var buf bytes.Buffer
	for y := 0; y < vt.height; y++ {
		if y > 0 {
			buf.WriteRune('\n')
		}
		for x := 0; x < vt.width; x++ {
			buf.WriteRune(vt.cells[y][x].Rune)
		}
	}
	
	return buf.String()
}

// ANSIParser handles ANSI escape sequence parsing
type ANSIParser struct {
	vt    *VirtualTerminal
	state parseState
	buf   []byte
}

type parseState int

const (
	stateNormal parseState = iota
	stateEscape
	stateCSI
)

// NewANSIParser creates a new ANSI parser
func NewANSIParser(vt *VirtualTerminal) *ANSIParser {
	return &ANSIParser{
		vt:    vt,
		state: stateNormal,
		buf:   make([]byte, 0, 256),
	}
}

// Parse processes input data
func (p *ANSIParser) Parse(data []byte) {
	for _, b := range data {
		switch p.state {
		case stateNormal:
			if b == 0x1B { // ESC
				p.state = stateEscape
			} else if b == '\n' {
				p.vt.NewLine()
			} else if b == '\r' {
				p.vt.CarriageReturn()
			} else if b == '\t' {
				// Handle tab
				tabStop := ((p.vt.cursorX / 8) + 1) * 8
				if tabStop < p.vt.width {
					p.vt.cursorX = tabStop
				}
			} else if b >= 0x20 && b < 0x7F {
				// Printable ASCII
				p.vt.PutRune(rune(b))
			} else if b >= 0x80 {
				// UTF-8 handling would go here
				// For now, just put the byte
				p.vt.PutRune(rune(b))
			}
			
		case stateEscape:
			if b == '[' {
				p.state = stateCSI
				p.buf = p.buf[:0]
			} else {
				// Other escape sequences
				p.state = stateNormal
			}
			
		case stateCSI:
			if b >= 0x40 && b <= 0x7E {
				// CSI sequence complete
				p.handleCSI(b)
				p.state = stateNormal
			} else {
				// Accumulate parameters
				p.buf = append(p.buf, b)
			}
		}
	}
}

// handleCSI processes CSI sequences
func (p *ANSIParser) handleCSI(cmd byte) {
	// Parse parameters
	params := p.parseParams()
	
	switch cmd {
	case 'H', 'f': // Cursor position
		y := 0
		x := 0
		if len(params) > 0 && params[0] > 0 {
			y = params[0] - 1
		}
		if len(params) > 1 && params[1] > 0 {
			x = params[1] - 1
		}
		p.vt.MoveCursor(x, y)
		
	case 'J': // Clear screen
		if len(params) == 0 || params[0] == 2 {
			p.vt.Clear()
		}
		
	case 'K': // Clear line
		// Clear from cursor to end of line
		for x := p.vt.cursorX; x < p.vt.width; x++ {
			p.vt.cells[p.vt.cursorY][x] = Cell{Rune: ' '}
		}
		
	case 'm': // Set graphics mode
		if len(params) == 0 {
			p.vt.ResetAttributes()
		} else {
			for _, param := range params {
				switch param {
				case 0:
					p.vt.ResetAttributes()
				case 1:
					p.vt.SetAttribute(p.vt.currentAttr | AttrBold)
				case 4:
					p.vt.SetAttribute(p.vt.currentAttr | AttrUnderline)
				case 7:
					p.vt.SetAttribute(p.vt.currentAttr | AttrReverse)
				// Add more color handling as needed
				}
			}
		}
	}
}

// parseParams parses CSI parameters
func (p *ANSIParser) parseParams() []int {
	if len(p.buf) == 0 {
		return nil
	}
	
	var params []int
	var current int
	var hasDigit bool
	
	for _, b := range p.buf {
		if b >= '0' && b <= '9' {
			current = current*10 + int(b-'0')
			hasDigit = true
		} else if b == ';' {
			if hasDigit {
				params = append(params, current)
			} else {
				params = append(params, 0)
			}
			current = 0
			hasDigit = false
		}
	}
	
	if hasDigit {
		params = append(params, current)
	}
	
	return params
}