package env

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnvFileSyncer_SyncModifiedFiles(t *testing.T) {
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
	
	// Create test files in main
	mainEnvFile := filepath.Join(mainDir, ".env")
	if err := os.WriteFile(mainEnvFile, []byte("ORIGINAL=true"), 0644); err != nil {
		t.Fatalf("Failed to create main .env: %v", err)
	}
	
	// Sleep to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	
	// Create modified file in worktree
	worktreeEnvFile := filepath.Join(worktreeDir, ".env")
	if err := os.WriteFile(worktreeEnvFile, []byte("MODIFIED=true"), 0644); err != nil {
		t.Fatalf("Failed to create worktree .env: %v", err)
	}
	
	// Create syncer
	syncer := NewEnvFileSyncer(worktreeDir, mainDir)
	
	// Perform sync
	syncedFiles, err := syncer.SyncModifiedFiles()
	if err != nil {
		t.Fatalf("SyncModifiedFiles failed: %v", err)
	}
	
	// Check results
	if len(syncedFiles) != 1 {
		t.Errorf("Expected 1 synced file, got %d", len(syncedFiles))
	}
	
	// Verify content was synced
	content, err := os.ReadFile(mainEnvFile)
	if err != nil {
		t.Fatalf("Failed to read synced file: %v", err)
	}
	
	if string(content) != "MODIFIED=true" {
		t.Errorf("Expected synced content 'MODIFIED=true', got '%s'", string(content))
	}
	
	// Verify backup was created
	backupFile := mainEnvFile + ".backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
	
	// Verify backup content
	backupContent, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	
	if string(backupContent) != "ORIGINAL=true" {
		t.Errorf("Expected backup content 'ORIGINAL=true', got '%s'", string(backupContent))
	}
}

func TestEnvFileSyncer_SkipUnmodifiedFiles(t *testing.T) {
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
	
	// Create identical files with same timestamp
	content := []byte("SAME=true")
	envFile := ".env"
	
	mainFile := filepath.Join(mainDir, envFile)
	worktreeFile := filepath.Join(worktreeDir, envFile)
	
	if err := os.WriteFile(mainFile, content, 0644); err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}
	if err := os.WriteFile(worktreeFile, content, 0644); err != nil {
		t.Fatalf("Failed to create worktree file: %v", err)
	}
	
	// Set same modification time
	modTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(mainFile, modTime, modTime); err != nil {
		t.Fatalf("Failed to set main file time: %v", err)
	}
	if err := os.Chtimes(worktreeFile, modTime, modTime); err != nil {
		t.Fatalf("Failed to set worktree file time: %v", err)
	}
	
	// Create syncer
	syncer := NewEnvFileSyncer(worktreeDir, mainDir)
	
	// Perform sync
	syncedFiles, err := syncer.SyncModifiedFiles()
	if err != nil {
		t.Fatalf("SyncModifiedFiles failed: %v", err)
	}
	
	// Should not sync unmodified files
	if len(syncedFiles) != 0 {
		t.Errorf("Expected 0 synced files, got %d", len(syncedFiles))
	}
}

func TestEnvFileSyncer_SkipNewFiles(t *testing.T) {
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
	
	// Create file only in worktree (new file)
	worktreeFile := filepath.Join(worktreeDir, ".env.local")
	if err := os.WriteFile(worktreeFile, []byte("NEW=true"), 0644); err != nil {
		t.Fatalf("Failed to create worktree file: %v", err)
	}
	
	// Create syncer
	syncer := NewEnvFileSyncer(worktreeDir, mainDir)
	
	// Perform sync
	syncedFiles, err := syncer.SyncModifiedFiles()
	if err != nil {
		t.Fatalf("SyncModifiedFiles failed: %v", err)
	}
	
	// Should not sync new files
	if len(syncedFiles) != 0 {
		t.Errorf("Expected 0 synced files, got %d", len(syncedFiles))
	}
	
	// Verify file was not created in main
	mainFile := filepath.Join(mainDir, ".env.local")
	if _, err := os.Stat(mainFile); !os.IsNotExist(err) {
		t.Error("New file should not have been created in main")
	}
}