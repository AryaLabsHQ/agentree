package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	
	// Verify backup was created (with timestamp pattern)
	dir := filepath.Dir(mainEnvFile)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	
	var backupFile string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".env.backup.") {
			backupFile = filepath.Join(dir, entry.Name())
			break
		}
	}
	
	if backupFile == "" {
		t.Error("Backup file was not created")
	} else {
		// Verify backup content
		backupContent, err := os.ReadFile(backupFile)
		if err != nil {
			t.Fatalf("Failed to read backup file: %v", err)
		}
		
		if string(backupContent) != "ORIGINAL=true" {
			t.Errorf("Expected backup content 'ORIGINAL=true', got '%s'", string(backupContent))
		}
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

func TestEnvFileSyncer_CleanupOldBackups(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create a test file
	testFile := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(testFile, []byte("TEST=true"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create old backup files
	oldTime := time.Now().Add(-8 * 24 * time.Hour) // 8 days ago
	oldTimestamp := oldTime.Format("20060102-150405")
	oldBackup := filepath.Join(tempDir, fmt.Sprintf(".env.backup.%s", oldTimestamp))
	if err := os.WriteFile(oldBackup, []byte("OLD=true"), 0644); err != nil {
		t.Fatalf("Failed to create old backup: %v", err)
	}
	
	// Create recent backup files
	recentTime := time.Now().Add(-2 * 24 * time.Hour) // 2 days ago
	recentTimestamp := recentTime.Format("20060102-150405")
	recentBackup := filepath.Join(tempDir, fmt.Sprintf(".env.backup.%s", recentTimestamp))
	if err := os.WriteFile(recentBackup, []byte("RECENT=true"), 0644); err != nil {
		t.Fatalf("Failed to create recent backup: %v", err)
	}
	
	// Create syncer
	syncer := NewEnvFileSyncer(tempDir, tempDir)
	
	// Run cleanup
	err := syncer.cleanupOldBackups(testFile)
	if err != nil {
		t.Fatalf("cleanupOldBackups failed: %v", err)
	}
	
	// Verify old backup was removed
	if _, err := os.Stat(oldBackup); !os.IsNotExist(err) {
		t.Error("Old backup should have been removed")
	}
	
	// Verify recent backup still exists
	if _, err := os.Stat(recentBackup); os.IsNotExist(err) {
		t.Error("Recent backup should not have been removed")
	}
}