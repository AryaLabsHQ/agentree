// Package env handles environment file operations
package env

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EnvFileCopier handles intelligent copying of environment files
type EnvFileCopier struct {
	srcDir         string
	destDir        string
	parser         *GitignoreParser
	customPatterns []string
	verbose        bool
}

// NewEnvFileCopier creates a new environment file copier
func NewEnvFileCopier(srcDir, destDir string) *EnvFileCopier {
	return &EnvFileCopier{
		srcDir:  srcDir,
		destDir: destDir,
		parser:  NewGitignoreParser(srcDir),
	}
}

// SetVerbose enables verbose logging
func (c *EnvFileCopier) SetVerbose(verbose bool) {
	c.verbose = verbose
	if c.parser != nil {
		c.parser.SetVerbose(verbose)
	}
}

// AddCustomPatterns adds custom patterns to search for
func (c *EnvFileCopier) AddCustomPatterns(patterns []string) {
	c.customPatterns = append(c.customPatterns, patterns...)
}

// DiscoverFiles finds all environment files to copy
func (c *EnvFileCopier) DiscoverFiles() ([]string, error) {
	fileMap := make(map[string]bool)
	
	if c.verbose {
		fmt.Println("🔍 Starting environment file discovery...")
	}
	
	// 1. Find files from .gitignore patterns
	ignoredFiles, err := c.parser.FindIgnoredEnvFiles()
	if err != nil {
		// Don't fail if we can't parse .gitignore, just continue
		fmt.Fprintf(os.Stderr, "Warning: couldn't parse .gitignore files: %v\n", err)
		ignoredFiles = []string{}
	}
	
	if c.verbose && len(ignoredFiles) > 0 {
		fmt.Printf("📄 Found %d files from .gitignore patterns:\n", len(ignoredFiles))
		for _, file := range ignoredFiles {
			fmt.Printf("   - %s\n", file)
		}
	}
	
	for _, file := range ignoredFiles {
		fileMap[file] = true
	}
	
	// 2. Add AI tool configuration files
	aiConfigs := GetDefaultAIConfigPatterns()
	if c.verbose {
		fmt.Printf("🤖 Checking AI tool configuration patterns:\n")
		for _, pattern := range aiConfigs {
			fmt.Printf("   - %s\n", pattern)
		}
	}
	
	for _, pattern := range aiConfigs {
		matches, err := c.findFilesMatchingPattern(pattern)
		if err != nil {
			continue
		}
		if c.verbose && len(matches) > 0 {
			fmt.Printf("   ✓ Found %d matches for %s\n", len(matches), pattern)
		}
		for _, match := range matches {
			fileMap[match] = true
		}
	}
	
	// 3. Add custom patterns if provided
	if c.verbose && len(c.customPatterns) > 0 {
		fmt.Printf("🔧 Checking custom patterns:\n")
		for _, pattern := range c.customPatterns {
			fmt.Printf("   - %s\n", pattern)
		}
	}
	
	for _, pattern := range c.customPatterns {
		matches, err := c.findFilesMatchingPattern(pattern)
		if err != nil {
			continue
		}
		if c.verbose && len(matches) > 0 {
			fmt.Printf("   ✓ Found %d matches for %s\n", len(matches), pattern)
		}
		for _, match := range matches {
			fileMap[match] = true
		}
	}
	
	// 4. Add legacy default files for backward compatibility
	legacyFiles := []string{".env", ".dev.vars"}
	if c.verbose {
		fmt.Printf("📦 Checking legacy files for backward compatibility:\n")
	}
	for _, file := range legacyFiles {
		if c.fileExists(file) {
			fileMap[file] = true
			if c.verbose {
				fmt.Printf("   ✓ Found %s\n", file)
			}
		} else if c.verbose {
			fmt.Printf("   ✗ Not found: %s\n", file)
		}
	}
	
	// Convert to sorted slice
	var files []string
	for file := range fileMap {
		files = append(files, file)
	}
	sort.Strings(files)
	
	if c.verbose {
		fmt.Printf("\n📋 Total files discovered: %d\n", len(files))
		if len(files) == 0 {
			fmt.Println("   ⚠️  No environment files found!")
			fmt.Println("   💡 Make sure:")
			fmt.Println("      - Environment files exist in the repository")
			fmt.Println("      - They are listed in .gitignore")
			fmt.Println("      - Or use custom patterns with --include flag")
		}
	}
	
	return files, nil
}

// CopyFiles copies the discovered files to the destination
func (c *EnvFileCopier) CopyFiles(files []string) ([]string, error) {
	var copiedFiles []string
	
	for _, file := range files {
		srcPath := filepath.Join(c.srcDir, file)
		destPath := filepath.Join(c.destDir, file)
		
		// Create destination directory if needed
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return copiedFiles, fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}
		
		// Copy the file
		if err := copyFile(srcPath, destPath); err != nil {
			// Log warning but continue with other files
			fmt.Fprintf(os.Stderr, "Warning: failed to copy %s: %v\n", file, err)
			continue
		}
		
		copiedFiles = append(copiedFiles, file)
	}
	
	return copiedFiles, nil
}

// CopyAllDiscoveredFiles is a convenience method that discovers and copies files
func (c *EnvFileCopier) CopyAllDiscoveredFiles() ([]string, error) {
	files, err := c.DiscoverFiles()
	if err != nil {
		return nil, err
	}
	
	return c.CopyFiles(files)
}

// findFilesMatchingPattern finds files matching a glob pattern
func (c *EnvFileCopier) findFilesMatchingPattern(pattern string) ([]string, error) {
	var matches []string
	
	// Handle recursive patterns
	if strings.Contains(pattern, "**/") {
		basePattern := strings.TrimPrefix(pattern, "**/")
		err := filepath.Walk(c.srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			
			if !info.IsDir() {
				relPath, err := filepath.Rel(c.srcDir, path)
				if err != nil {
					return nil
				}
				
				fileName := filepath.Base(path)
				if matched, _ := filepath.Match(basePattern, fileName); matched {
					matches = append(matches, relPath)
				}
			}
			
			return nil
		})
		return matches, err
	}
	
	// Handle non-recursive patterns
	absPattern := filepath.Join(c.srcDir, pattern)
	files, err := filepath.Glob(absPattern)
	if err != nil {
		return nil, err
	}
	
	// Convert to relative paths
	for _, file := range files {
		relPath, err := filepath.Rel(c.srcDir, file)
		if err != nil {
			continue
		}
		matches = append(matches, relPath)
	}
	
	return matches, nil
}

// fileExists checks if a file exists relative to srcDir
func (c *EnvFileCopier) fileExists(relPath string) bool {
	absPath := filepath.Join(c.srcDir, relPath)
	_, err := os.Stat(absPath)
	return err == nil
}

// CopyEnvFilesEnhanced is the new enhanced version that uses gitignore patterns
func CopyEnvFilesEnhanced(srcDir, destDir string, customPatterns []string) ([]string, error) {
	copier := NewEnvFileCopier(srcDir, destDir)
	if len(customPatterns) > 0 {
		copier.AddCustomPatterns(customPatterns)
	}
	
	return copier.CopyAllDiscoveredFiles()
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