// Package env handles environment file operations
package env

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyEnvFiles copies environment files from source to destination
func CopyEnvFiles(srcDir, destDir string) ([]string, error) {
	envFiles := []string{".env", ".dev.vars"}
	copiedFiles := []string{}
	
	for _, file := range envFiles {
		srcPath := filepath.Join(srcDir, file)
		destPath := filepath.Join(destDir, file)
		
		// Check if source file exists
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue // Skip if file doesn't exist
		}
		
		// Copy the file
		if err := copyFile(srcPath, destPath); err != nil {
			return copiedFiles, fmt.Errorf("failed to copy %s: %w", file, err)
		}
		
		copiedFiles = append(copiedFiles, file)
	}
	
	return copiedFiles, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			// Log error but don't fail since we already read the file
			fmt.Fprintf(os.Stderr, "Warning: failed to close source file: %v\n", err)
		}
	}()
	
	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			// Log error but don't fail since we already wrote the file
			fmt.Fprintf(os.Stderr, "Warning: failed to close destination file: %v\n", err)
		}
	}()
	
	// Copy the contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	
	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, srcInfo.Mode())
}