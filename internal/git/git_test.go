package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	
	// Create temp directory
	tmpDir := t.TempDir()
	
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}
	
	// Configure git user (required for commits)
	exec.Command("git", "config", "user.name", "Test User").Dir = tmpDir
	exec.Command("git", "config", "user.email", "test@example.com").Dir = tmpDir
	
	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	cmd.Run()
	
	// Return cleanup function
	cleanup := func() {
		// Cleanup is handled by t.TempDir()
	}
	
	return tmpDir, cleanup
}

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		wantErr bool
	}{
		{
			name: "valid git repository",
			setup: func() (string, func()) {
				tmpDir, cleanup := setupTestRepo(t)
				// Change to the repo directory
				originalWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				return originalWd, func() {
					os.Chdir(originalWd)
					cleanup()
				}
			},
			wantErr: false,
		},
		{
			name: "not a git repository",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				oldWd, _ := os.Getwd()
				os.Chdir(tmpDir)
				return oldWd, func() {
					os.Chdir(oldWd)
				}
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setup()
			defer cleanup()
			
			repo, err := NewRepository()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !tt.wantErr && repo == nil {
				t.Error("Expected repository object, got nil")
			}
		})
	}
}

func TestGetDefaultWorktreeDir(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	// Resolve symlinks for comparison
	tmpDirResolved, _ := filepath.EvalSymlinks(tmpDir)
	expected := filepath.Join(filepath.Dir(tmpDirResolved), filepath.Base(tmpDirResolved)+"-worktrees")
	result := repo.GetDefaultWorktreeDir()
	
	// Also resolve result path
	resultResolved, _ := filepath.EvalSymlinks(result)
	expectedResolved, _ := filepath.EvalSymlinks(expected)
	
	if resultResolved != expectedResolved {
		t.Errorf("GetDefaultWorktreeDir() = %v, want %v", result, expected)
	}
}

func TestCurrentBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	// Test on main/master branch
	branch, err := repo.CurrentBranch()
	if err != nil {
		t.Errorf("CurrentBranch() error = %v", err)
	}
	
	// Git might use 'main' or 'master' as default
	if branch != "main" && branch != "master" {
		t.Errorf("CurrentBranch() = %v, want main or master", branch)
	}
	
	// Test detached HEAD
	cmd := exec.Command("git", "checkout", "HEAD~0")
	cmd.Dir = tmpDir
	cmd.Run()
	
	branch, err = repo.CurrentBranch()
	if err != nil {
		t.Errorf("CurrentBranch() on detached HEAD error = %v", err)
	}
	
	// Should return short commit hash
	if len(branch) < 6 || len(branch) > 8 {
		t.Errorf("Expected short commit hash, got %v", branch)
	}
}

func TestCreateWorktree(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	tests := []struct {
		name    string
		branch  string
		base    string
		dest    string
		wantErr bool
	}{
		{
			name:    "create new worktree",
			branch:  "feature/test",
			base:    "HEAD",
			dest:    filepath.Join(t.TempDir(), "test-worktree"),
			wantErr: false,
		},
		{
			name:    "branch already exists",
			branch:  "feature/test", // Same as above
			base:    "HEAD",
			dest:    filepath.Join(t.TempDir(), "test-worktree2"),
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateWorktree(tt.branch, tt.base, tt.dest)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWorktree() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			// Verify worktree was created
			if !tt.wantErr {
				if _, err := os.Stat(tt.dest); os.IsNotExist(err) {
					t.Errorf("Expected worktree directory to exist at %s", tt.dest)
				}
			}
		})
	}
}

func TestListBranches(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	// Create some test branches
	exec.Command("git", "branch", "test-branch-1").Dir = tmpDir
	exec.Command("git", "branch", "test-branch-2").Dir = tmpDir
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	branches, err := repo.ListBranches()
	if err != nil {
		t.Fatalf("ListBranches() error = %v", err)
	}
	
	// Should have at least the main/master branch
	if len(branches) < 1 {
		t.Error("Expected at least one branch")
	}
	
	// Check if branches contain expected names
	branchMap := make(map[string]bool)
	for _, b := range branches {
		branchMap[b] = true
	}
	
	if !branchMap["main"] && !branchMap["master"] {
		t.Error("Expected main or master branch")
	}
}

func TestFindWorktree(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	// Create a test worktree
	worktreePath := filepath.Join(t.TempDir(), "test-worktree")
	err = repo.CreateWorktree("test/branch", "HEAD", worktreePath)
	if err != nil {
		t.Fatalf("Failed to create test worktree: %v", err)
	}
	
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{
			name:    "find by branch name",
			target:  "test/branch",
			wantErr: false,
		},
		{
			name:    "find by path",
			target:  worktreePath,
			wantErr: false,
		},
		{
			name:    "non-existent branch",
			target:  "non-existent",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := repo.FindWorktree(tt.target)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("FindWorktree() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !tt.wantErr && info != nil {
				// Resolve symlinks for path comparison
				expectedPath, _ := filepath.EvalSymlinks(worktreePath)
				actualPath, _ := filepath.EvalSymlinks(info.Path)
				
				if actualPath != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, actualPath)
				}
				if info.Branch != "test/branch" {
					t.Errorf("Expected branch test/branch, got %s", info.Branch)
				}
			}
		})
	}
}

func TestRemoveWorktree(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	// Create a test worktree
	worktreePath := filepath.Join(t.TempDir(), "test-worktree")
	err = repo.CreateWorktree("test/remove", "HEAD", worktreePath)
	if err != nil {
		t.Fatalf("Failed to create test worktree: %v", err)
	}
	
	// Test removing worktree
	err = repo.RemoveWorktree(worktreePath, false)
	if err != nil {
		t.Errorf("RemoveWorktree() error = %v", err)
	}
	
	// Verify worktree was removed
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("Expected worktree directory to be removed")
	}
}

func TestDeleteBranch(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	// Create a test branch
	cmd := exec.Command("git", "branch", "test-delete")
	cmd.Dir = tmpDir
	cmd.Run()
	
	repo, err := NewRepository()
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	// Delete the branch
	err = repo.DeleteBranch("test-delete")
	if err != nil {
		t.Errorf("DeleteBranch() error = %v", err)
	}
	
	// Verify branch was deleted
	branches, _ := repo.ListBranches()
	for _, b := range branches {
		if b == "test-delete" {
			t.Error("Expected branch to be deleted")
		}
	}
}