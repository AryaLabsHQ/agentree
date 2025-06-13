// +build integration

package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AryaLabsHQ/agentree/internal/env"
)

func TestRemoveWithSync(t *testing.T) {
	// This is an integration test that would require a real git repo
	// For now, we'll test the sync functionality in isolation
	
	// Create temporary directories
	tempDir := t.TempDir()
	mainDir := filepath.Join(tempDir, "main")
	worktreeDir := filepath.Join(tempDir, "worktree")
	
	// Create directories
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main dir: %v", err)
	}
	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		t.Fatalf("Failed to create worktree dir: %v", err)
	}
	
	// Create test files
	mainEnvFile := filepath.Join(mainDir, ".env")
	worktreeEnvFile := filepath.Join(worktreeDir, ".env")
	
	// Create original file in main
	if err := os.WriteFile(mainEnvFile, []byte("ORIGINAL=true\n"), 0644); err != nil {
		t.Fatalf("Failed to create main .env: %v", err)
	}
	
	// Sleep to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	
	// Create modified file in worktree
	if err := os.WriteFile(worktreeEnvFile, []byte("MODIFIED=true\n"), 0644); err != nil {
		t.Fatalf("Failed to create worktree .env: %v", err)
	}
	
	// Create syncer and sync
	syncer := env.NewEnvFileSyncer(worktreeDir, mainDir)
	syncedFiles, err := syncer.SyncModifiedFiles()
	if err != nil {
		t.Fatalf("SyncModifiedFiles failed: %v", err)
	}
	
	// Verify sync happened
	if len(syncedFiles) != 1 {
		t.Errorf("Expected 1 synced file, got %d", len(syncedFiles))
	}
	
	// Verify content
	content, err := os.ReadFile(mainEnvFile)
	if err != nil {
		t.Fatalf("Failed to read synced file: %v", err)
	}
	
	if string(content) != "MODIFIED=true\n" {
		t.Errorf("Expected synced content 'MODIFIED=true\\n', got '%s'", string(content))
	}
}