package cmd

import (
	"fmt"
	"os"

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
Use -R to also delete the local branch after removing the worktree.`,
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
	
	// Copy flags to alias
	removeAlias.Flags().BoolVarP(&force, "yes", "y", false, "Force removal without confirmation")
	removeAlias.Flags().BoolVarP(&deleteBranch, "delete-branch", "R", false, "Also delete the local branch")
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
