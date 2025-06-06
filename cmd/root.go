// Package cmd contains all CLI commands for hatch
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/charmbracelet/lipgloss"
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
	Use:   "hatch",
	Short: "Create and manage isolated Git worktrees for agentic workflows",
	Long: `hatch is a tool for creating and managing isolated Git worktrees.
	
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