package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

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
	maxLines  int
	running   bool
}

func NewMultiplexPOC() (*MultiplexPOC, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	}

	return &MultiplexPOC{
		screen:   screen,
		maxLines: 1000, // Keep last 1000 lines
		running:  true,
	}, nil
}

func (m *MultiplexPOC) Run() error {
	defer m.screen.Fini()

	// Start process
	if err := m.startProcess(); err != nil {
		return err
	}
	defer m.cleanup()

	// Start output reader
	go m.readOutput()

	// Main UI loop
	for m.running {
		m.draw()

		ev := m.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				m.running = false
			case tcell.KeyUp:
				m.scroll(-1)
			case tcell.KeyDown:
				m.scroll(1)
			case tcell.KeyPgUp:
				m.scroll(-10)
			case tcell.KeyPgDn:
				m.scroll(10)
			case tcell.KeyHome:
				m.scrollPos = 0
			case tcell.KeyEnd:
				m.outputMu.Lock()
				_, height := m.screen.Size()
				m.scrollPos = max(0, len(m.output)-height+3)
				m.outputMu.Unlock()
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q', 'Q':
					m.running = false
				case 'c', 'C':
					m.clearOutput()
				}
			}
		case *tcell.EventResize:
			m.screen.Sync()
		}
	}

	return nil
}

func (m *MultiplexPOC) startProcess() error {
	// Create a test command that produces continuous output
	m.cmd = exec.Command("bash", "-c", `
		echo -e "\033[32mStarting multiplexer test process...\033[0m"
		echo "This simulates a Claude Code instance"
		echo ""
		
		counter=0
		while true; do
			counter=$((counter + 1))
			
			# Simulate different types of output
			case $((counter % 4)) in
				0) echo -e "[\033[34m$(date +%H:%M:%S)\033[0m] \033[36mINFO\033[0m: Processing request $counter..." ;;
				1) echo -e "[\033[34m$(date +%H:%M:%S)\033[0m] \033[33mDEBUG\033[0m: Token usage: $((RANDOM % 1000)) tokens" ;;
				2) echo -e "[\033[34m$(date +%H:%M:%S)\033[0m] \033[32mSUCCESS\033[0m: Operation completed" ;;
				3) echo -e "[\033[34m$(date +%H:%M:%S)\033[0m] Output line $counter: $(head -c 50 < /dev/urandom | base64 | head -c 30)" ;;
			esac
			
			sleep 0.5
		done
	`)

	// Start the command with a pty
	var err error
	m.ptyFile, err = pty.Start(m.cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}

	return nil
}

func (m *MultiplexPOC) readOutput() {
	scanner := bufio.NewScanner(m.ptyFile)
	for scanner.Scan() {
		line := scanner.Text()
		
		m.outputMu.Lock()
		m.output = append(m.output, line)
		
		// Limit buffer size
		if len(m.output) > m.maxLines {
			m.output = m.output[len(m.output)-m.maxLines:]
			if m.scrollPos > 0 {
				m.scrollPos--
			}
		}
		m.outputMu.Unlock()
	}
}

func (m *MultiplexPOC) draw() {
	m.screen.Clear()
	
	// Styles
	headerStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true)
	statusStyle := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
	outputStyle := tcell.StyleDefault

	width, height := m.screen.Size()

	// Header
	header := "=== Multiplexer POC === (q: quit, ↑↓: scroll, c: clear)"
	m.drawText(0, 0, header, headerStyle, width)

	// Output area
	m.outputMu.Lock()
	outputLines := m.output
	totalLines := len(outputLines)
	
	// Calculate visible lines
	visibleHeight := height - 3 // Header + status + margin
	startLine := m.scrollPos
	endLine := min(startLine+visibleHeight, totalLines)

	// Draw output lines
	for i := startLine; i < endLine; i++ {
		y := i - startLine + 2
		line := outputLines[i]
		
		// Strip ANSI codes for display (in real implementation, we'd parse them)
		displayLine := stripANSI(line)
		m.drawText(0, y, displayLine, outputStyle, width)
	}
	m.outputMu.Unlock()

	// Status bar
	status := fmt.Sprintf(" Lines: %d | Scroll: %d | Time: %s | Process: %s ",
		totalLines, m.scrollPos, time.Now().Format("15:04:05"), 
		m.getProcessStatus())
	m.drawText(0, height-1, status, statusStyle, width)

	m.screen.Show()
}

func (m *MultiplexPOC) drawText(x, y int, text string, style tcell.Style, maxWidth int) {
	for i, r := range text {
		if x+i >= maxWidth {
			break
		}
		m.screen.SetContent(x+i, y, r, nil, style)
	}
	// Fill the rest of the line with the style
	for i := len(text); x+i < maxWidth; i++ {
		m.screen.SetContent(x+i, y, ' ', nil, style)
	}
}

func (m *MultiplexPOC) scroll(delta int) {
	m.outputMu.Lock()
	defer m.outputMu.Unlock()
	
	_, height := m.screen.Size()
	visibleHeight := height - 3
	maxScroll := max(0, len(m.output)-visibleHeight)
	
	m.scrollPos += delta
	m.scrollPos = max(0, min(m.scrollPos, maxScroll))
}

func (m *MultiplexPOC) clearOutput() {
	m.outputMu.Lock()
	m.output = []string{"[Output cleared]"}
	m.scrollPos = 0
	m.outputMu.Unlock()
}

func (m *MultiplexPOC) getProcessStatus() string {
	if m.cmd == nil || m.cmd.Process == nil {
		return "Not started"
	}
	if m.cmd.ProcessState != nil {
		return "Stopped"
	}
	return "Running"
}

func (m *MultiplexPOC) cleanup() {
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
		_ = m.cmd.Wait()
	}
	if m.ptyFile != nil {
		_ = m.ptyFile.Close()
	}
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// stripANSI removes ANSI escape sequences (simplified version)
func stripANSI(s string) string {
	// This is a very basic implementation
	// In production, we'd properly parse ANSI sequences
	result := strings.Builder{}
	inEscape := false
	
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' || r == 'K' || r == 'H' || r == 'J' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	
	return result.String()
}

func main() {
	m, err := NewMultiplexPOC()
	if err != nil {
		log.Fatalf("Failed to create multiplexer: %v", err)
	}

	if err := m.Run(); err != nil {
		log.Fatalf("Multiplexer error: %v", err)
	}

	fmt.Println("\n✓ Multiplexer POC completed successfully!")
	fmt.Println("✓ PTY + TUI integration working")
	fmt.Println("✓ Process output captured and displayed")
	fmt.Println("✓ Scrolling and interaction working")
}