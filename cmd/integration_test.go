//go:build integration
// +build integration

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This file contains integration tests that test the actual agentree binary
// Run with: go test -tags=integration ./cmd

func TestAgentreeBinaryIntegration(t *testing.T) {
	// Build the agentree binary
	binary := filepath.Join(t.TempDir(), "agentree")
	buildCmd := exec.Command("go", "build", "-o", binary, "../cmd/agentree")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build agentree binary: %v", err)
	}

	// Create test repo
	repoDir := t.TempDir()
	setupGitRepo(t, repoDir)

	// Change to repo directory
	oldWd, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(oldWd)

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "create worktree",
			args:    []string{"-b", "test-integration"},
			wantErr: false,
			contains: []string{
				"Worktree ready",
				"agent/test-integration",
			},
		},
		{
			name:    "create with env copy",
			args:    []string{"-b", "test-env", "-e"},
			wantErr: false,
			contains: []string{
				"Worktree ready",
				"Copied .env",
			},
		},
		{
			name:    "show help",
			args:    []string{},
			wantErr: false,
			contains: []string{
				"agentree is a tool for creating and managing isolated Git worktrees",
			},
		},
		{
			name:    "remove worktree",
			args:    []string{"rm", "-y", "agent/test-integration"},
			wantErr: false,
			contains: []string{
				"Removed worktree",
			},
		},
	}

	// Create test env file
	os.WriteFile(".env", []byte("TEST=123"), 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			output, err := cmd.CombinedOutput()

			if (err != nil) != tt.wantErr {
				t.Errorf("agentree %v: error = %v, wantErr %v\nOutput: %s",
					tt.args, err, tt.wantErr, output)
			}

			outputStr := string(output)
			for _, want := range tt.contains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Expected output to contain %q, got:\n%s", want, outputStr)
				}
			}
		})
	}
}

func setupGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run %v: %v", args, err)
		}
	}

	// Create initial commit
	readme := filepath.Join(dir, "README.md")
	os.WriteFile(readme, []byte("# Test"), 0644)

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial")
	cmd.Dir = dir
	cmd.Run()
}
