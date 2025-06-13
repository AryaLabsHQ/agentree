package cmd

import (
	"fmt"
	"os"

	"github.com/AryaLabsHQ/agentree/internal/multiplex"
	"github.com/spf13/cobra"
)

var (
	// Flags
	multiplexAll    bool
	multiplexConfig string
	autoStart       bool
	tokenLimit      int64
)

// multiplexCmd represents the multiplex command
var multiplexCmd = &cobra.Command{
	Use:     "multiplex [worktrees...]",
	Aliases: []string{"mx"},
	Short:   "Run multiple Claude Code instances in a multiplexed TUI",
	Long: `Launch and manage multiple Claude Code instances across different worktrees.

The multiplexer provides a terminal UI for:
- Running Claude Code in multiple worktrees simultaneously
- Monitoring token usage and costs
- Switching between instances
- Coordinating work across branches

Examples:
  # Launch multiplexer for specific worktrees
  agentree multiplex feat-auth feat-ui

  # Launch all worktrees
  agentree multiplex --all

  # Use custom configuration
  agentree multiplex --config multiplex.yml`,
	Args: cobra.ArbitraryArgs,
	RunE: runMultiplex,
}

func init() {
	rootCmd.AddCommand(multiplexCmd)

	// Flags
	multiplexCmd.Flags().BoolVarP(&multiplexAll, "all", "a", false, "Launch all worktrees")
	multiplexCmd.Flags().StringVarP(&multiplexConfig, "config", "c", "", "Configuration file path")
	multiplexCmd.Flags().BoolVar(&autoStart, "auto-start", false, "Automatically start all instances")
	multiplexCmd.Flags().Int64Var(&tokenLimit, "token-limit", 50000, "Total token limit across all instances")

	// Hidden debug flags
	multiplexCmd.Flags().Bool("debug", false, "Enable debug logging")
	multiplexCmd.Flags().MarkHidden("debug")
}

func runMultiplex(cmd *cobra.Command, args []string) error {
	// Create configuration
	config := &multiplex.Config{
		AutoStart:  autoStart,
		TokenLimit: tokenLimit,
	}

	// Load config file if specified
	if multiplexConfig != "" {
		loadedConfig, err := multiplex.LoadConfig(multiplexConfig)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		config = loadedConfig
	}

	// Determine which worktrees to use
	var worktrees []string
	if multiplexAll {
		discovered, err := multiplex.DiscoverWorktrees()
		if err != nil {
			return fmt.Errorf("failed to discover worktrees: %w", err)
		}
		worktrees = discovered
	} else if len(args) > 0 {
		worktrees = args
	} else {
		// No worktrees specified
		return fmt.Errorf("no worktrees specified. Use --all or provide worktree names")
	}

	// Create and run multiplexer
	m, err := multiplex.New(config, worktrees)
	if err != nil {
		return fmt.Errorf("failed to create multiplexer: %w", err)
	}

	// Run the multiplexer (blocks until exit)
	if err := m.Run(); err != nil {
		return fmt.Errorf("multiplexer error: %w", err)
	}

	fmt.Fprintln(os.Stderr, "\n✨ Multiplexer session ended")
	return nil
}