package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AryaLabsHQ/agentree/internal/config"
	"github.com/AryaLabsHQ/agentree/internal/env"
	"github.com/AryaLabsHQ/agentree/internal/git"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "rm [branch|path]",
	Short: "Remove a worktree",
	Long: `Remove a Git worktree by branch name or path.

You can specify either:
- A branch name: agentree rm agent/feature-x
- A worktree path: agentree rm ../myrepo-worktrees/agent-feature-x

Use -y to skip confirmation and force removal of dirty worktrees.
Use -R to also delete the local branch after removing the worktree.
Use -S to sync environment files back to the main worktree before removal.`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// Provide both branch and worktree completions
		branches, _ := getBranchCompletions(cmd, args, toComplete)
		worktrees, _ := getWorktreeCompletions(cmd, args, toComplete)
		return append(branches, worktrees...), cobra.ShellCompDirectiveNoFileComp
	},
}

var (
	force        bool
	deleteBranch bool
	syncEnv      bool
	rmVerbose    bool
)

func init() {
	rootCmd.AddCommand(removeCmd)

	// Also add 'remove' as an alias
	removeAlias := *removeCmd
	removeAlias.Use = "remove [branch|path]"
	rootCmd.AddCommand(&removeAlias)

	// Define flags
	removeCmd.Flags().BoolVarP(&force, "yes", "y", false, "Force removal without confirmation")
	removeCmd.Flags().BoolVarP(&deleteBranch, "delete-branch", "R", false, "Also delete the local branch")
	removeCmd.Flags().BoolVarP(&syncEnv, "sync-env", "S", false, "Sync environment files back to main worktree")
	removeCmd.Flags().BoolVarP(&rmVerbose, "verbose", "v", false, "Show detailed sync process")
	
	// Copy flags to alias
	removeAlias.Flags().BoolVarP(&force, "yes", "y", false, "Force removal without confirmation")
	removeAlias.Flags().BoolVarP(&deleteBranch, "delete-branch", "R", false, "Also delete the local branch")
	removeAlias.Flags().BoolVarP(&syncEnv, "sync-env", "S", false, "Sync environment files back to main worktree")
	removeAlias.Flags().BoolVarP(&rmVerbose, "verbose", "v", false, "Show detailed sync process")
}

func runRemove(cmd *cobra.Command, args []string) error {
	target := args[0]

	// Create repository instance
	repo, err := git.NewRepository()
	if err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return err
	}

	// Find the worktree
	info, err := repo.FindWorktree(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return err
	}

	// Check if we should sync environment files
	shouldSync := syncEnv
	if !shouldSync {
		// Check configuration for auto-sync setting
		projectConfig, _ := config.LoadProjectConfig(repo.Root)
		globalConfig, _ := config.LoadGlobalConfig()
		mergedConfig := config.MergeConfig(globalConfig, projectConfig)
		shouldSync = mergedConfig.EnvConfig.SyncBackOnRemove
	}
	
	// Sync environment files if requested
	if shouldSync {
		fmt.Println(infoStyle.Render("Syncing environment files back to main worktree..."))
		
		// Get main worktree path
		mainPath, err := repo.GetMainWorktreePath()
		if err != nil {
			fmt.Fprintln(os.Stderr, warningStyle.Render(fmt.Sprintf("Warning: Could not get main worktree path: %v", err)))
			fmt.Fprintln(os.Stderr, warningStyle.Render("Skipping environment file sync"))
			// Continue with remove operation
			goto skipSync
		}
		
		// Resolve the worktree path to absolute
		worktreePath := info.Path
		if !filepath.IsAbs(worktreePath) {
			worktreePath, _ = filepath.Abs(worktreePath)
		}
		
		// Don't sync if we're removing the main worktree
		if worktreePath == mainPath {
			fmt.Println(infoStyle.Render("Skipping sync: cannot sync main worktree to itself"))
		} else {
			// Create syncer
			syncer := env.NewEnvFileSyncer(worktreePath, mainPath)
			syncer.SetVerbose(rmVerbose)
			
			// Load sync patterns from config if available
			projectConfig, _ := config.LoadProjectConfig(repo.Root)
			globalConfig, _ := config.LoadGlobalConfig()
			mergedConfig := config.MergeConfig(globalConfig, projectConfig)
			
			if len(mergedConfig.EnvConfig.SyncPatterns) > 0 {
				syncer.SetPatterns(mergedConfig.EnvConfig.SyncPatterns)
			}
			
			// Perform sync
			syncedFiles, err := syncer.SyncModifiedFiles()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Error during sync: %v\n", err)
			} else if len(syncedFiles) > 0 {
				fmt.Println(successStyle.Render(fmt.Sprintf("✅ Synced %d files:", len(syncedFiles))))
				for _, file := range syncedFiles {
					fmt.Printf("    📋 %s\n", file)
				}
			} else {
				fmt.Println(infoStyle.Render("No modified environment files to sync"))
			}
		}
	}
skipSync:

	// Confirm if not forced
	if !force {
		fmt.Printf("Remove worktree at %s? [y/N] ", info.Path)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// Default to "no" on error
			response = "n"
		}
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Remove the worktree
	if err := repo.RemoveWorktree(info.Path, force); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return err
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("✅ Removed worktree %s", info.Path)))

	// Delete branch if requested
	if deleteBranch && info.Branch != "" {
		if err := repo.DeleteBranch(info.Branch); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Warning: %v", err)))
		} else {
			fmt.Println(successStyle.Render(fmt.Sprintf("🗑️  Deleted branch %s", info.Branch)))
		}
	}

	return nil
}
