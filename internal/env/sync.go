// Package env handles environment file operations
package env

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EnvFileSyncer handles syncing environment files back to the main worktree
type EnvFileSyncer struct {
	worktreeDir string
	mainDir     string
	parser      *GitignoreParser
	verbose     bool
	patterns    []string
}

// NewEnvFileSyncer creates a new environment file syncer
func NewEnvFileSyncer(worktreeDir, mainDir string) *EnvFileSyncer {
	return &EnvFileSyncer{
		worktreeDir: worktreeDir,
		mainDir:     mainDir,
		parser:      NewGitignoreParser(worktreeDir),
	}
}

// SetVerbose enables verbose logging
func (s *EnvFileSyncer) SetVerbose(verbose bool) {
	s.verbose = verbose
	if s.parser != nil {
		s.parser.SetVerbose(verbose)
	}
}

// SetPatterns sets custom patterns for syncing (if different from discovery patterns)
func (s *EnvFileSyncer) SetPatterns(patterns []string) {
	s.patterns = patterns
}

// SyncModifiedFiles syncs environment files that have been modified in the worktree
func (s *EnvFileSyncer) SyncModifiedFiles() ([]string, error) {
	// Discover files to check for syncing
	copier := NewEnvFileCopier(s.worktreeDir, s.mainDir)
	copier.SetVerbose(s.verbose)
	
	// Use custom patterns if set, otherwise use default discovery
	if len(s.patterns) > 0 {
		copier.AddCustomPatterns(s.patterns)
	}
	
	files, err := copier.DiscoverFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to discover files: %w", err)
	}
	
	if s.verbose {
		fmt.Printf("🔍 Checking %d files for modifications...\n", len(files))
	}
	
	var syncedFiles []string
	for _, file := range files {
		worktreePath := filepath.Join(s.worktreeDir, file)
		mainPath := filepath.Join(s.mainDir, file)
		
		// Check if file exists in worktree
		worktreeInfo, err := os.Stat(worktreePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // File doesn't exist in worktree, skip
			}
			return syncedFiles, fmt.Errorf("failed to stat %s: %w", file, err)
		}
		
		// Check if file exists in main
		mainInfo, err := os.Stat(mainPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File exists in worktree but not in main - this is a new file
				if s.verbose {
					fmt.Printf("⚠️  Skipping %s (new file, not in main worktree)\n", file)
				}
				continue
			}
			return syncedFiles, fmt.Errorf("failed to stat %s in main: %w", file, err)
		}
		
		// Compare modification times
		if worktreeInfo.ModTime().After(mainInfo.ModTime()) {
			if s.verbose {
				fmt.Printf("📝 File %s modified (worktree: %s, main: %s)\n", 
					file, 
					worktreeInfo.ModTime().Format(time.RFC3339),
					mainInfo.ModTime().Format(time.RFC3339))
			}
			
			// Create backup of main file
			backupPath := mainPath + ".backup"
			if err := s.createBackup(mainPath, backupPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create backup for %s: %v\n", file, err)
				// Continue anyway - backup is nice to have but not critical
			}
			
			// Sync the file
			if err := copyFile(worktreePath, mainPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to sync %s: %v\n", file, err)
				continue
			}
			
			syncedFiles = append(syncedFiles, file)
		}
	}
	
	return syncedFiles, nil
}

// createBackup creates a backup of a file
func (s *EnvFileSyncer) createBackup(src, dst string) error {
	// Ensure backup directory exists
	backupDir := filepath.Dir(dst)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	return copyFile(src, dst)
}

// CompareFiles checks if two files have different content
func (s *EnvFileSyncer) CompareFiles(file1, file2 string) (bool, error) {
	content1, err := os.ReadFile(file1)
	if err != nil {
		return false, err
	}
	
	content2, err := os.ReadFile(file2)
	if err != nil {
		return false, err
	}
	
	return string(content1) != string(content2), nil
}