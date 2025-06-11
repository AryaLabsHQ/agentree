// Package git provides functions for interacting with Git repositories.
// In Go, package names should be short, lowercase, and descriptive.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository with methods for worktree operations.
// In Go, we often create types (structs) to group related functionality.
type Repository struct {
	// Root is the absolute path to the repository root
	Root string
	// RepoName is the name of the repository (e.g., "agentree")
	RepoName string
}

// NewRepository creates a new Repository instance by finding the Git root.
// Functions that create new instances often start with "New" in Go.
func NewRepository() (*Repository, error) {
	// exec.Command runs external commands (like subprocess in Python)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	
	// Run the command and capture output
	output, err := cmd.Output()
	if err != nil {
		// In Go, we return errors instead of throwing exceptions
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}
	
	// strings.TrimSpace removes leading/trailing whitespace (including newlines)
	root := strings.TrimSpace(string(output))
	
	// filepath.Base gets the last element of a path (like basename in shell)
	repoName := filepath.Base(root)
	
	repo := &Repository{
		Root:     root,
		RepoName: repoName,
	}
	
	// Fetch updates from remote
	if err := repo.Fetch(); err != nil {
		// Don't fail if fetch fails, just warn
		fmt.Fprintf(os.Stderr, "Warning: git fetch failed: %v\n", err)
	}
	
	return repo, nil
}

// Fetch runs git fetch to update remote refs
func (r *Repository) Fetch() error {
	cmd := exec.Command("git", "fetch", "--prune")
	cmd.Dir = r.Root
	return cmd.Run()
}

// CreateWorktree creates a new worktree for the given branch.
// Methods in Go are functions with a receiver (r *Repository)
func (r *Repository) CreateWorktree(branch, base, dest string) error {
	// First, let's check if the branch already exists
	checkCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err := checkCmd.Run(); err == nil {
		return fmt.Errorf("branch %s already exists", branch)
	}
	
	// Create the branch
	createCmd := exec.Command("git", "branch", branch, base)
	createCmd.Dir = r.Root
	if output, err := createCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch: %s", output)
	}
	
	// Add the worktree
	addCmd := exec.Command("git", "worktree", "add", dest, branch)
	addCmd.Dir = r.Root
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add worktree: %s", output)
	}
	
	return nil
}

// GetDefaultWorktreeDir returns the default directory for worktrees
func (r *Repository) GetDefaultWorktreeDir() string {
	parent := filepath.Dir(r.Root)
	return filepath.Join(parent, r.RepoName+"-worktrees")
}

// CurrentBranch returns the current branch name or HEAD commit
func (r *Repository) CurrentBranch() (string, error) {
	// Try to get symbolic ref first
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Dir = r.Root
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output)), nil
	}
	
	// Fall back to commit hash
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = r.Root
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// ListBranches returns a list of all local branches
func (r *Repository) ListBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	cmd.Dir = r.Root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	
	// Split output into lines and filter empty ones
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			branches = append(branches, line)
		}
	}
	
	return branches, nil
}

// WorktreeInfo represents information about a worktree
type WorktreeInfo struct {
	Path   string
	Branch string
}

// FindWorktree finds a worktree by branch name or path
func (r *Repository) FindWorktree(target string) (*WorktreeInfo, error) {
	// Get list of worktrees
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = r.Root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	var currentPath string
	var info *WorktreeInfo
	
	// Check if target is a directory
	if stat, err := os.Stat(target); err == nil && stat.IsDir() {
		// Convert to absolute path and resolve symlinks for comparison
		if absPath, err := filepath.Abs(target); err == nil {
			if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
				target = resolved
			} else {
				target = absPath
			}
		}
	}
	
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		switch parts[0] {
		case "worktree":
			currentPath = parts[1]
			// Resolve symlinks for comparison
			resolvedPath, _ := filepath.EvalSymlinks(currentPath)
			if currentPath == target || resolvedPath == target {
				info = &WorktreeInfo{Path: currentPath}
			}
		case "branch":
			branch := strings.TrimPrefix(parts[1], "refs/heads/")
			if currentPath != "" {
				if branch == target {
					info = &WorktreeInfo{Path: currentPath, Branch: branch}
				} else if info != nil && info.Path == currentPath {
					info.Branch = branch
				}
			}
		}
		
		if info != nil && info.Branch != "" {
			return info, nil
		}
	}
	
	return nil, fmt.Errorf("worktree not found for %s", target)
}

// RemoveWorktree removes a worktree
func (r *Repository) RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove worktree: %s", output)
	}
	
	return nil
}

// DeleteBranch deletes a local branch
func (r *Repository) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = r.Root
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete branch: %s", output)
	}
	
	return nil
}

// ListWorktrees returns a list of all worktree paths
func (r *Repository) ListWorktrees() ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = r.Root
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	var worktrees []string
	
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == "worktree" {
			worktrees = append(worktrees, parts[1])
		}
	}
	
	return worktrees, nil
}