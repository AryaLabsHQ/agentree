// Package tui provides terminal user interface components using Bubbletea
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Bubbletea follows The Elm Architecture:
// 1. Model - holds the state
// 2. Update - handles messages and updates state
// 3. View - renders the UI based on state

// branchSelectorModel holds the state for our branch selector
type branchSelectorModel struct {
	textInput    textinput.Model
	branches     []string
	filtered     []string
	cursor       int
	selected     string
	choosing     bool
	err          error
}

// Style definitions
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("147"))
		
	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)
		
	cursorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))
)

// NewBranchSelector creates a new branch selector model
func NewBranchSelector(branches []string) branchSelectorModel {
	ti := textinput.New()
	ti.Placeholder = "Type to filter branches or enter new branch name"
	ti.Focus() // Focus the text input immediately
	ti.CharLimit = 100
	ti.Width = 50

	return branchSelectorModel{
		textInput: ti,
		branches:  branches,
		filtered:  branches,
		choosing:  true,
	}
}

// Init is called when the program starts
// It returns an initial command to run (we don't need one)
func (m branchSelectorModel) Init() tea.Cmd {
	// Return a command that tells the text input to blink
	return textinput.Blink
}

// Update handles messages (keyboard input, etc.) and updates the model
func (m branchSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// User wants to quit
			m.choosing = false
			return m, tea.Quit
			
		case tea.KeyUp:
			// Move cursor up
			if m.cursor > 0 {
				m.cursor--
			}
			
		case tea.KeyDown:
			// Move cursor down
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			
		case tea.KeyEnter:
			// Select the branch or create new one
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
			} else if m.textInput.Value() != "" {
				// Use the typed value as a new branch
				m.selected = m.textInput.Value()
			}
			m.choosing = false
			return m, tea.Quit
		}
	}

	// Update the text input
	m.textInput, cmd = m.textInput.Update(msg)
	
	// Filter branches based on input
	m.filterBranches()
	
	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	return m, cmd
}

// filterBranches filters the branch list based on text input
func (m *branchSelectorModel) filterBranches() {
	search := strings.ToLower(m.textInput.Value())
	if search == "" {
		m.filtered = m.branches
		return
	}

	filtered := make([]string, 0)
	for _, branch := range m.branches {
		if strings.Contains(strings.ToLower(branch), search) {
			filtered = append(filtered, branch)
		}
	}
	m.filtered = filtered
}

// View renders the UI
func (m branchSelectorModel) View() string {
	if !m.choosing {
		return ""
	}

	var s strings.Builder
	
	// Title
	s.WriteString(titleStyle.Render("Select a branch or create a new one:") + "\n\n")
	
	// Text input
	s.WriteString(m.textInput.View() + "\n\n")
	
	// Branch list
	if len(m.filtered) > 0 {
		s.WriteString("Existing branches:\n")
		for i, branch := range m.filtered {
			cursor := "  "
			style := lipgloss.NewStyle()
			
			if i == m.cursor {
				cursor = cursorStyle.Render("→ ")
				style = selectedStyle
			}
			
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(branch)))
		}
	} else if m.textInput.Value() != "" {
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("No matching branches. Press Enter to create: ") +
			selectedStyle.Render(m.textInput.Value()) + "\n")
	}
	
	// Help
	s.WriteString("\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: navigate • enter: select • esc: cancel"))

	return s.String()
}

// GetSelected returns the selected branch name
func (m branchSelectorModel) GetSelected() string {
	return m.selected
}

// RunBranchSelector runs the interactive branch selector and returns the selected branch
func RunBranchSelector(branches []string) (string, error) {
	p := tea.NewProgram(NewBranchSelector(branches))
	
	// Run returns the final model and any error
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	
	// Type assert to get our model back
	if m, ok := finalModel.(branchSelectorModel); ok {
		if m.selected == "" {
			return "", fmt.Errorf("no branch selected")
		}
		return m.selected, nil
	}
	
	return "", fmt.Errorf("unexpected model type")
}