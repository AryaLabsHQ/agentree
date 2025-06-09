// Package tui provides terminal user interface components using Bubbletea
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardOptions holds all the options that can be configured
type WizardOptions struct {
	Branch       string
	BaseBranch   string
	CopyEnv      bool
	RunSetup     bool
	Push         bool
	CreatePR     bool
	CustomDest   string
}

// wizardModel holds the state for our interactive wizard
type wizardModel struct {
	steps        []string
	currentStep  int
	options      WizardOptions
	branches     []string
	textInput    textinput.Model
	cursor       int
	filtered     []string
	err          error
	quitting     bool
	completed    bool
	defaultDir   string
}

// Style definitions for wizard
var (
	wizardTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("147"))

	wizardPromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	wizardSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	wizardCursorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	wizardHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	wizardCompleteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))
)

// NewWizard creates a new wizard model
func NewWizard(branches []string, currentBranch string, defaultDir string) wizardModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	// Reorder branches to put current branch first
	orderedBranches := make([]string, 0, len(branches))
	
	// Add current branch first if it exists in the list
	foundCurrent := false
	for _, b := range branches {
		if b == currentBranch {
			orderedBranches = append(orderedBranches, b)
			foundCurrent = true
			break
		}
	}
	
	// Add all other branches
	for _, b := range branches {
		if b != currentBranch {
			orderedBranches = append(orderedBranches, b)
		}
	}
	
	// If current branch wasn't in the list, add it first
	if !foundCurrent && currentBranch != "" {
		orderedBranches = append([]string{currentBranch}, orderedBranches...)
	}

	return wizardModel{
		steps: []string{
			"baseBranch",
			"branchName",
			"copyEnv",
			"runSetup",
			"push",
			"createPR",
			"customDest",
			"review",
		},
		currentStep: 0,
		options: WizardOptions{
			BaseBranch: currentBranch,
			CopyEnv:    true,  // Default to true
			RunSetup:   true,  // Default to true
		},
		branches:   orderedBranches,
		textInput:  ti,
		filtered:   orderedBranches,
		cursor:     0, // Current branch is now at position 0
		defaultDir: defaultDir,
	}
}

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle y/n questions first to prevent text input from processing these keys
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch m.steps[m.currentStep] {
		case "copyEnv", "runSetup", "push", "createPR":
			switch keyMsg.String() {
			case "y", "Y":
				switch m.steps[m.currentStep] {
				case "copyEnv":
					m.options.CopyEnv = true
					m.currentStep++
				case "runSetup":
					m.options.RunSetup = true
					m.currentStep++
				case "push":
					m.options.Push = true
					m.currentStep++
				case "createPR":
					m.options.CreatePR = true
					m.currentStep++
					// Clear text input for next step
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Leave empty for default or enter custom path"
				}
				return m, nil
			case "n", "N":
				switch m.steps[m.currentStep] {
				case "copyEnv":
					m.options.CopyEnv = false
					m.currentStep++
				case "runSetup":
					m.options.RunSetup = false
					m.currentStep++
				case "push":
					m.options.Push = false
					m.options.CreatePR = false // Can't create PR without push
					m.currentStep += 2 // Skip PR question
				case "createPR":
					m.options.CreatePR = false
					m.currentStep++
					// Clear text input for next step
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Leave empty for default or enter custom path"
				}
				return m, nil
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyLeft:
			// Allow going back to previous step
			if m.currentStep > 0 {
				m.currentStep--
				// Reset text input for certain steps
				switch m.steps[m.currentStep] {
				case "branchName":
					m.textInput.SetValue(m.options.Branch)
				case "customDest":
					m.textInput.SetValue(m.options.CustomDest)
				}
			}

		case tea.KeyUp:
			if m.steps[m.currentStep] == "baseBranch" && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.steps[m.currentStep] == "baseBranch" && m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case tea.KeyEnter:
			switch m.steps[m.currentStep] {
			case "baseBranch":
				if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
					m.options.BaseBranch = m.filtered[m.cursor]
					m.currentStep++
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter new branch name (e.g., feature-x)"
				}

			case "branchName":
				if m.textInput.Value() != "" {
					m.options.Branch = m.textInput.Value()
					m.currentStep++
				}

			case "copyEnv":
				// Toggle is handled by y/n keys
				m.currentStep++

			case "runSetup":
				// Toggle is handled by y/n keys
				m.currentStep++

			case "push":
				// Toggle is handled by y/n keys
				m.currentStep++

			case "createPR":
				m.currentStep++

			case "customDest":
				// Optional - can be empty
				if m.textInput.Value() != "" {
					m.options.CustomDest = m.textInput.Value()
				}
				m.currentStep++

			case "review":
				m.completed = true
				m.quitting = true
				return m, tea.Quit
			}

			// Reset for next step
			if m.currentStep < len(m.steps) {
				switch m.steps[m.currentStep] {
				case "customDest":
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Leave empty for default or enter custom path"
				}
			}

		default:
			// Default case - no special handling needed
		}
	}

	// Update text input for relevant steps only
	switch m.steps[m.currentStep] {
	case "baseBranch", "branchName", "customDest":
		m.textInput, cmd = m.textInput.Update(msg)

		// Filter branches for base branch selection
		if m.steps[m.currentStep] == "baseBranch" {
			m.filterBranches()
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
	}

	return m, cmd
}

func (m *wizardModel) filterBranches() {
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

func (m wizardModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(wizardTitleStyle.Render("🌳 agentree Setup Wizard") + "\n\n")

	// Progress indicator
	progress := fmt.Sprintf("Step %d of %d", m.currentStep+1, len(m.steps)-1)
	s.WriteString(wizardHelpStyle.Render(progress) + "\n\n")

	// Current step
	switch m.steps[m.currentStep] {
	case "baseBranch":
		s.WriteString(wizardPromptStyle.Render("Select base branch:") + "\n\n")
		s.WriteString(m.textInput.View() + "\n\n")

		if len(m.filtered) > 0 {
			for i, branch := range m.filtered {
				cursor := "  "
				style := lipgloss.NewStyle()
				branchDisplay := branch

				// Add (current) indicator for the current branch
				if branch == m.options.BaseBranch && i == 0 && m.textInput.Value() == "" {
					branchDisplay = branch + " (current)"
				}

				if i == m.cursor {
					cursor = wizardCursorStyle.Render("→ ")
					style = wizardSelectedStyle
				}

				s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(branchDisplay)))
			}
		}

	case "branchName":
		s.WriteString(wizardPromptStyle.Render("Enter new branch name:") + "\n\n")
		s.WriteString(m.textInput.View() + "\n")
		s.WriteString(wizardHelpStyle.Render("\nThis will be prefixed with 'agent/' if it doesn't contain a slash"))

	case "copyEnv":
		s.WriteString(wizardPromptStyle.Render("Copy .env and .dev.vars files?") + "\n\n")
		s.WriteString(fmt.Sprintf("Current: %s\n\n", m.boolToYesNo(m.options.CopyEnv)))
		s.WriteString(wizardHelpStyle.Render("Press y/n to choose"))

	case "runSetup":
		s.WriteString(wizardPromptStyle.Render("Run setup commands (npm/pnpm install, etc)?") + "\n\n")
		s.WriteString(fmt.Sprintf("Current: %s\n\n", m.boolToYesNo(m.options.RunSetup)))
		s.WriteString(wizardHelpStyle.Render("Press y/n to choose"))

	case "push":
		s.WriteString(wizardPromptStyle.Render("Push to GitHub after creation?") + "\n\n")
		s.WriteString(fmt.Sprintf("Current: %s\n\n", m.boolToYesNo(m.options.Push)))
		s.WriteString(wizardHelpStyle.Render("Press y/n to choose"))

	case "createPR":
		s.WriteString(wizardPromptStyle.Render("Create a GitHub PR?") + "\n\n")
		s.WriteString(fmt.Sprintf("Current: %s\n\n", m.boolToYesNo(m.options.CreatePR)))
		s.WriteString(wizardHelpStyle.Render("Press y/n to choose"))

	case "customDest":
		s.WriteString(wizardPromptStyle.Render("Custom destination directory (optional):") + "\n\n")
		s.WriteString(m.textInput.View() + "\n")
		s.WriteString(wizardHelpStyle.Render(fmt.Sprintf("\nPress Enter to use default: %s", m.defaultDir)))

	case "review":
		s.WriteString(wizardPromptStyle.Render("Review your choices:") + "\n\n")
		s.WriteString(fmt.Sprintf("• Base branch: %s\n", wizardSelectedStyle.Render(m.options.BaseBranch)))
		s.WriteString(fmt.Sprintf("• New branch: %s\n", wizardSelectedStyle.Render(m.options.Branch)))
		s.WriteString(fmt.Sprintf("• Copy env files: %s\n", m.boolToYesNo(m.options.CopyEnv)))
		s.WriteString(fmt.Sprintf("• Run setup: %s\n", m.boolToYesNo(m.options.RunSetup)))
		s.WriteString(fmt.Sprintf("• Push to GitHub: %s\n", m.boolToYesNo(m.options.Push)))
		if m.options.Push {
			s.WriteString(fmt.Sprintf("• Create PR: %s\n", m.boolToYesNo(m.options.CreatePR)))
		}
		if m.options.CustomDest != "" {
			s.WriteString(fmt.Sprintf("• Custom destination: %s\n", wizardSelectedStyle.Render(m.options.CustomDest)))
		}
		s.WriteString("\n" + wizardCompleteStyle.Render("Press Enter to create worktree"))
	}

	// Help
	if m.currentStep > 0 {
		s.WriteString("\n\n" + wizardHelpStyle.Render("←: back • ctrl+c: cancel"))
	} else {
		s.WriteString("\n\n" + wizardHelpStyle.Render("ctrl+c: cancel"))
	}

	return s.String()
}

func (m wizardModel) boolToYesNo(b bool) string {
	if b {
		return wizardCompleteStyle.Render("Yes")
	}
	return wizardHelpStyle.Render("No")
}

// GetOptions returns the configured options
func (m wizardModel) GetOptions() WizardOptions {
	return m.options
}

// ErrWizardCancelled is returned when the user cancels the wizard
var ErrWizardCancelled = fmt.Errorf("cancelled")

// RunWizard runs the interactive wizard and returns the configured options
func RunWizard(branches []string, currentBranch string, defaultDir string) (WizardOptions, error) {
	p := tea.NewProgram(NewWizard(branches, currentBranch, defaultDir))

	finalModel, err := p.Run()
	if err != nil {
		return WizardOptions{}, err
	}

	if m, ok := finalModel.(wizardModel); ok {
		if !m.completed {
			return WizardOptions{}, ErrWizardCancelled
		}
		return m.options, nil
	}

	return WizardOptions{}, fmt.Errorf("unexpected model type")
}