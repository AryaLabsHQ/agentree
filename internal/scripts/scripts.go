// Package scripts handles execution of post-create scripts
package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	scriptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("cyan"))
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("green"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red"))
)

// Runner executes scripts in a directory
type Runner struct {
	Dir string
}

// NewRunner creates a new script runner for the given directory
func NewRunner(dir string) *Runner {
	return &Runner{Dir: dir}
}

// RunScripts executes a list of scripts in the runner's directory
func (r *Runner) RunScripts(scripts []string) error {
	if len(scripts) == 0 {
		return nil
	}

	fmt.Println("🚀 Running post-create scripts...")

	for _, script := range scripts {
		fmt.Printf("   → %s\n", scriptStyle.Render(script))

		// Execute the script
		if err := r.runScript(script); err != nil {
			fmt.Printf("   %s Failed: %v\n", errorStyle.Render("✗"), err)
			// Continue with other scripts even if one fails
		} else {
			fmt.Printf("   %s Success\n", successStyle.Render("✓"))
		}
	}

	return nil
}

// runScript executes a single script
func (r *Runner) runScript(script string) error {
	// Use sh -c to run the script, allowing for complex commands
	cmd := exec.Command("sh", "-c", script)
	cmd.Dir = r.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// DetermineScripts determines which scripts to run based on configuration and detection
func DetermineScripts(
	customScripts []string,
	projectScripts []string,
	detectedScripts []string,
	globalOverride string,
) []string {
	// Priority order:
	// 1. Custom scripts from -S flag
	if len(customScripts) > 0 {
		return customScripts
	}

	// 2. Project-specific scripts from .agentreerc
	if len(projectScripts) > 0 {
		return projectScripts
	}

	// 3. Global override for detected package manager
	if len(detectedScripts) > 0 && globalOverride != "" {
		// If we have a global override for this package manager, use it
		return parseScriptString(globalOverride)
	}

	// 4. Auto-detected scripts
	return detectedScripts
}

// parseScriptString splits a script string by && into individual commands
func parseScriptString(script string) []string {
	parts := strings.Split(script, "&&")
	scripts := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			scripts = append(scripts, trimmed)
		}
	}

	return scripts
}
