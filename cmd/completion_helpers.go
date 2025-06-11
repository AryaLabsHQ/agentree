package cmd

import (
	"path/filepath"
	"strings"

	"github.com/AryaLabsHQ/agentree/internal/git"
	"github.com/spf13/cobra"
)

// getBranchCompletions returns all available git branches for completion
func getBranchCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	repo, err := git.NewRepository()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	branches, err := repo.ListBranches()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Filter branches that start with the toComplete string
	var completions []string
	for _, branch := range branches {
		if strings.HasPrefix(branch, toComplete) {
			completions = append(completions, branch)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// getWorktreeCompletions returns all existing worktrees for completion
func getWorktreeCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	repo, err := git.NewRepository()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Extract just the directory names for easier completion
	var completions []string
	for _, worktree := range worktrees {
		// Use the base name of the path for completion
		name := filepath.Base(worktree)
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Commented out for future use when agent types and config commands are implemented

// // getAgentTypeCompletions returns available agent types for completion
// func getAgentTypeCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 	// These are placeholder agent types - will be expanded in the future
// 	agentTypes := []string{
// 		"claude",
// 		"cursor",
// 		"copilot",
// 		"generic",
// 	}

// 	var completions []string
// 	for _, agent := range agentTypes {
// 		if strings.HasPrefix(agent, toComplete) {
// 			completions = append(completions, agent)
// 		}
// 	}

// 	return completions, cobra.ShellCompDirectiveNoFileComp
// }

// // getConfigKeyCompletions returns available configuration keys for completion
// func getConfigKeyCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 	// Configuration keys that can be set
// 	configKeys := []string{
// 		"defaultSetup",
// 		"npmSetup",
// 		"pnpmSetup",
// 		"yarnSetup",
// 		"postCreateScripts",
// 		"env.enabled",
// 		"env.includePatterns",
// 		"env.excludePatterns",
// 	}

// 	var completions []string
// 	for _, key := range configKeys {
// 		if strings.HasPrefix(key, toComplete) {
// 			completions = append(completions, key)
// 		}
// 	}

// 	return completions, cobra.ShellCompDirectiveNoFileComp
// }