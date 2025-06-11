package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AryaLabsHQ/agentree/internal/config"
	"github.com/AryaLabsHQ/agentree/internal/detector"
	"github.com/AryaLabsHQ/agentree/internal/env"
	"github.com/AryaLabsHQ/agentree/internal/git"
	"github.com/AryaLabsHQ/agentree/internal/scripts"
	"github.com/AryaLabsHQ/agentree/internal/tui"
	"github.com/spf13/cobra"
)

// Create command flags
var (
	branch        string
	base          string
	push          bool
	pr            bool
	dest          string
	copyEnv       bool
	runSetup      bool
	interactive   bool
	customScripts []string
	verbose       bool
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new worktree",
	Long: `Create a new Git worktree with an isolated branch.

By default, branches are prefixed with 'agent/' unless they already contain a slash.
The worktree is created in a sibling directory named <repo>-worktrees.`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Define flags
	createCmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch name (required)")
	createCmd.Flags().StringVarP(&base, "from", "f", "", "Base branch to fork from (default: current branch)")
	createCmd.Flags().BoolVarP(&push, "push", "p", false, "Push branch to origin after creation")
	createCmd.Flags().BoolVarP(&pr, "pr", "r", false, "Create GitHub PR after push (implies -p)")
	createCmd.Flags().StringVarP(&dest, "dest", "d", "", "Custom destination directory")
	createCmd.Flags().BoolVarP(&copyEnv, "env", "e", true, "Copy .env and .dev.vars files")
	createCmd.Flags().BoolVarP(&runSetup, "setup", "s", true, "Run setup scripts (auto-detect or from config)")
	createCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive wizard to guide through setup")
	createCmd.Flags().StringArrayVarP(&customScripts, "script", "S", nil, "Custom post-create script (can be used multiple times)")
	createCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed environment discovery process")

	// Make branch required unless in interactive mode
	_ = createCmd.MarkFlagRequired("branch")
}

// For backward compatibility, also make flags available at root level
func init() {
	// Copy all flags to root command for backward compatibility
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch name")
	rootCmd.Flags().StringVarP(&base, "from", "f", "", "Base branch to fork from")
	rootCmd.Flags().BoolVarP(&push, "push", "p", false, "Push branch to origin")
	rootCmd.Flags().BoolVarP(&pr, "pr", "r", false, "Create GitHub PR (implies -p)")
	rootCmd.Flags().StringVarP(&dest, "dest", "d", "", "Custom destination directory")
	rootCmd.Flags().BoolVarP(&copyEnv, "env", "e", true, "Copy .env and .dev.vars files")
	rootCmd.Flags().BoolVarP(&runSetup, "setup", "s", true, "Run setup scripts")
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive wizard")
	rootCmd.Flags().StringArrayVarP(&customScripts, "script", "S", nil, "Custom post-create script")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed discovery process")

	// If root command is called with flags, run create
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Check if any create flags are set
		if branch != "" || interactive || cmd.Flags().Changed("branch") {
			return runCreate(cmd, args)
		}
		// Otherwise show help
		return cmd.Help()
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Create repository instance
	repo, err := git.NewRepository()
	if err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return err
	}

	// Handle interactive mode
	if interactive {
		branches, err := repo.ListBranches()
		if err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error listing branches: %v", err)))
			return err
		}

		currentBranch, err := repo.CurrentBranch()
		if err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error getting current branch: %v", err)))
			return err
		}

		// Get default directory for display
		worktreeDir := repo.GetDefaultWorktreeDir()
		defaultDest := filepath.Join(worktreeDir, "agent-{branch-name}")

		options, err := tui.RunWizard(branches, currentBranch, defaultDest)
		if err != nil {
			if err == tui.ErrWizardCancelled {
				// User cancelled - exit gracefully without error message
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: %v", err)))
			return err
		}

		// Apply wizard options
		branch = options.Branch
		base = options.BaseBranch
		copyEnv = options.CopyEnv
		runSetup = options.RunSetup
		push = options.Push
		pr = options.CreatePR
		if options.CustomDest != "" {
			dest = options.CustomDest
		}
	}

	// Validate branch
	if branch == "" {
		fmt.Fprintln(os.Stderr, errorStyle.Render("Error: -b/--branch is required"))
		return fmt.Errorf("branch name required")
	}

	// If -r is set, also set -p
	if pr {
		push = true
	}

	// Add agent/ prefix if no slash
	if !strings.Contains(branch, "/") {
		branch = "agent/" + branch
	}

	// Set default base branch if not specified
	if base == "" {
		base, err = repo.CurrentBranch()
		if err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error getting current branch: %v", err)))
			return err
		}
	}

	// Determine destination directory
	if dest == "" {
		worktreeDir := repo.GetDefaultWorktreeDir()
		if err := os.MkdirAll(worktreeDir, 0755); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error creating worktree directory: %v", err)))
			return err
		}

		sanitized := strings.ReplaceAll(branch, "/", "-")
		dest = filepath.Join(worktreeDir, sanitized)
	}

	// Check if destination already exists
	if _, err := os.Stat(dest); err == nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: Destination %s already exists", dest)))
		return fmt.Errorf("destination exists")
	}

	// Create the worktree
	fmt.Println(infoStyle.Render("Creating worktree..."))
	if err := repo.CreateWorktree(branch, base, dest); err != nil {
		fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return err
	}

	fmt.Println(successStyle.Render("✅ Worktree ready:"))
	fmt.Printf("    %s %s\n", labelStyle.Render("path"), dest)
	fmt.Printf("    %s %s (from %s)\n", labelStyle.Render("branch"), branch, base)

	// Copy environment files if requested
	if copyEnv {
		// Load configurations and merge them
		projectConfig, _ := config.LoadProjectConfig(repo.Root)
		globalConfig, _ := config.LoadGlobalConfig()
		mergedConfig := config.MergeConfig(globalConfig, projectConfig)
		
		// Check if env copying is enabled in config
		if !mergedConfig.EnvConfig.Enabled {
			fmt.Println(infoStyle.Render("Environment file copying disabled by configuration"))
		} else {
			// Use the enhanced copier with configuration
			copier := env.NewEnvFileCopier(repo.Root, dest)
			copier.SetVerbose(verbose)
			
			// Add custom patterns from config
			if len(mergedConfig.EnvConfig.IncludePatterns) > 0 {
				copier.AddCustomPatterns(mergedConfig.EnvConfig.IncludePatterns)
			}
			
			// Discover files based on .gitignore and patterns
			fmt.Println(infoStyle.Render("Discovering environment files..."))
			files, err := copier.DiscoverFiles()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Error discovering files: %v\n", err)
				files = []string{} // Continue with empty list instead of failing
			}
			
			if len(files) > 0 {
				// Filter out excluded patterns
				var filteredFiles []string
				for _, file := range files {
					excluded := false
					for _, pattern := range mergedConfig.EnvConfig.ExcludePatterns {
						if matched, _ := filepath.Match(pattern, file); matched {
							excluded = true
							break
						}
					}
					if !excluded {
						filteredFiles = append(filteredFiles, file)
					}
				}
				
				// Copy the filtered files
				if len(filteredFiles) > 0 {
					copiedFiles, err := copier.CopyFiles(filteredFiles)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Some files couldn't be copied: %v\n", err)
					}
					for _, file := range copiedFiles {
						fmt.Printf("📋 Copied %s\n", file)
					}
				} else {
					fmt.Println(infoStyle.Render("No environment files found to copy"))
				}
			} else {
				fmt.Println(infoStyle.Render("No environment files found to copy"))
			}
		}
	}

	// Run post-create scripts if requested
	if runSetup || len(customScripts) > 0 {
		projectConfig, _ := config.LoadProjectConfig(repo.Root)
		globalConfig, _ := config.LoadGlobalConfig()

		detectedScripts := detector.DetectSetupCommands(dest)

		var globalOverride string
		if len(detectedScripts) > 0 {
			if strings.Contains(detectedScripts[0], "pnpm") && globalConfig.PnpmSetup != "" {
				globalOverride = globalConfig.PnpmSetup
			} else if strings.Contains(detectedScripts[0], "npm") && globalConfig.NpmSetup != "" {
				globalOverride = globalConfig.NpmSetup
			} else if strings.Contains(detectedScripts[0], "yarn") && globalConfig.YarnSetup != "" {
				globalOverride = globalConfig.YarnSetup
			} else if globalConfig.DefaultSetup != "" {
				globalOverride = globalConfig.DefaultSetup
			}
		}

		scriptsToRun := scripts.DetermineScripts(
			customScripts,
			projectConfig.PostCreateScripts,
			detectedScripts,
			globalOverride,
		)

		runner := scripts.NewRunner(dest)
		if err := runner.RunScripts(scriptsToRun); err != nil {
			// Log error but don't fail the command
			fmt.Fprintf(os.Stderr, "Warning: Some post-create scripts failed: %v\n", err)
		}
	}

	// Push to origin if requested
	if push {
		fmt.Println(infoStyle.Render("Pushing to origin..."))
		pushCmd := exec.Command("git", "push", "-u", "origin", branch)
		pushCmd.Dir = dest
		if output, err := pushCmd.CombinedOutput(); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error pushing: %s", output)))
			return err
		}
		fmt.Println(successStyle.Render("✓ Pushed to origin"))
	}

	// Create PR if requested
	if pr {
		fmt.Println(infoStyle.Render("Creating GitHub PR..."))
		prCmd := exec.Command("gh", "pr", "create", "--fill", "--web")
		prCmd.Dir = dest
		if output, err := prCmd.CombinedOutput(); err != nil {
			if _, lookupErr := exec.LookPath("gh"); lookupErr != nil {
				fmt.Fprintln(os.Stderr, errorStyle.Render("⚠️  gh CLI not found; skipping PR creation"))
			} else {
				fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("Error creating PR: %s", string(output))))
			}
		}
	}

	return nil
}
