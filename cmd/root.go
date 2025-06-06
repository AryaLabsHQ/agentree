// Package cmd contains all CLI commands for agentree
package cmd

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// version is set by goreleaser at build time
	version = "dev"

	// Style definitions
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Italic(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "agentree",
	Short: "Create and manage isolated Git worktrees for AI coding agents",
	Long: `agentree is a tool for creating and managing isolated Git worktrees.
	
It simplifies working with multiple AI coding agents by creating isolated
branches and directories for concurrent work without conflicts.`,
	Version: version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Disable Cobra's default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
