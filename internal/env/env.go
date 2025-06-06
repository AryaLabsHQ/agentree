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
	defer srcFile.Close() // defer ensures the file is closed when function returns
	
	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
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