package multiplex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AryaLabsHQ/agentree/internal/git"
)

// DiscoverWorktrees finds all worktrees in the current repository
func DiscoverWorktrees() ([]string, error) {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Find git repository root
	repo, err := git.NewRepository()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Get all worktrees
	worktrees, err := repo.ListWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// All paths are valid worktrees
	return worktrees, nil
}

// DiscoverWorktreesByPattern finds worktrees matching a pattern
func DiscoverWorktreesByPattern(pattern string) ([]string, error) {
	all, err := DiscoverWorktrees()
	if err != nil {
		return nil, err
	}

	var matched []string
	for _, wt := range all {
		name := filepath.Base(wt)
		if matchesPattern(name, pattern) {
			matched = append(matched, wt)
		}
	}

	return matched, nil
}

// GetWorktreeInfo returns detailed information about a worktree
type WorktreeInfo struct {
	Path       string
	Name       string
	Branch     string
	HEAD       string
	IsDetached bool
	HasChanges bool
}

// GetWorktreeInfo retrieves information about a specific worktree
func GetWorktreeInfo(path string) (*WorktreeInfo, error) {
	// Change to the worktree directory
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(oldDir)
	
	if err := os.Chdir(path); err != nil {
		return nil, fmt.Errorf("failed to change to worktree directory: %w", err)
	}

	repo, err := git.NewRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current branch
	branch, err := repo.CurrentBranch()
	if err != nil {
		branch = "detached"
	}

	// Get HEAD commit
	// TODO: Add GetHEAD method to git.Repository
	head := "unknown"

	// Check for changes
	// TODO: Add HasChanges method to git.Repository
	hasChanges := false

	info := &WorktreeInfo{
		Path:       path,
		Name:       filepath.Base(path),
		Branch:     branch,
		HEAD:       head,
		IsDetached: branch == "detached",
		HasChanges: hasChanges,
	}

	return info, nil
}

// FindClaudeExecutable finds the claude executable in PATH
func FindClaudeExecutable() (string, error) {
	// Check common locations first
	commonPaths := []string{
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		"/opt/homebrew/bin/claude",
		fmt.Sprintf("%s/.local/bin/claude", os.Getenv("HOME")),
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Search in PATH
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	for _, dir := range paths {
		path := filepath.Join(dir, "claude")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("claude executable not found in PATH")
}

// ValidateWorktrees checks if all worktrees are valid
func ValidateWorktrees(paths []string) error {
	for _, path := range paths {
		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("worktree %s does not exist: %w", path, err)
		}

		if !info.IsDir() {
			return fmt.Errorf("worktree %s is not a directory", path)
		}

		// Check if it's a git repository
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			return fmt.Errorf("worktree %s is not a git repository", path)
		}
	}

	return nil
}

// matchesPattern checks if a name matches a simple glob pattern
func matchesPattern(name, pattern string) bool {
	// Simple pattern matching (can be enhanced)
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// Simple prefix/suffix matching
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(name, suffix)
		}
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(name, prefix)
		}
	}

	return name == pattern
}