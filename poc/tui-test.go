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
		log.Fatalf("Failed to create screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %v", err)
	}
	defer screen.Fini()

	// Define styles
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	headerStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true)
	highlightStyle := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
	
	screen.SetStyle(defStyle)

	// Test content
	menuItems := []string{
		"Item 1: Process One",
		"Item 2: Process Two", 
		"Item 3: Process Three",
		"Item 4: Process Four",
	}
	selectedIndex := 0

	// Main loop
	for {
		screen.Clear()
		width, height := screen.Size()

		// Draw header
		drawText(screen, 0, 0, width, "=== TUI Test - Press 'q' to quit, ↑↓ to navigate ===", headerStyle)

		// Draw menu items
		for i, item := range menuItems {
			y := i + 2
			style := defStyle
			prefix := "  "
			if i == selectedIndex {
				style = highlightStyle
				prefix = "► "
			}
			drawText(screen, 0, y, width, prefix+item, style)
		}

		// Draw footer
		footer := fmt.Sprintf("Selected: %d | Width: %d | Height: %d", selectedIndex, width, height)
		drawText(screen, 0, height-1, width, footer, headerStyle)

		screen.Show()

		// Handle events
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return
			case tcell.KeyUp:
				if selectedIndex > 0 {
					selectedIndex--
				}
			case tcell.KeyDown:
				if selectedIndex < len(menuItems)-1 {
					selectedIndex++
				}
			case tcell.KeyRune:
				if ev.Rune() == 'q' || ev.Rune() == 'Q' {
					return
				}
			}
		case *tcell.EventResize:
			screen.Sync()
		}
	}
}

func drawText(s tcell.Screen, x, y, maxWidth int, text string, style tcell.Style) {
	for i, r := range text {
		if x+i >= maxWidth {
			break
		}
		s.SetContent(x+i, y, r, nil, style)
	}
	// Fill the rest of the line with the style
	for i := len(text); x+i < maxWidth; i++ {
		s.SetContent(x+i, y, ' ', nil, style)
	}
}