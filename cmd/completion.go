package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for agentree.

This command generates shell completion scripts that enable tab completion
for agentree commands, flags, branch names, worktree names, and more.

Examples:
  # Bash (add to ~/.bashrc or ~/.bash_profile):
  source <(agentree completion bash)
  
  # Zsh (add to ~/.zshrc):
  source <(agentree completion zsh)
  # Or for oh-my-zsh, add to ~/.oh-my-zsh/completions/:
  agentree completion zsh > ~/.oh-my-zsh/completions/_agentree
  
  # Fish (add to ~/.config/fish/completions/):
  agentree completion fish > ~/.config/fish/completions/agentree.fish
  
  # PowerShell (add to your PowerShell profile):
  agentree completion powershell | Out-String | Invoke-Expression`,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}