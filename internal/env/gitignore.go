// Package env handles environment file operations
package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GitignoreParser parses .gitignore files to find environment and local files
type GitignoreParser struct {
	root string
}

// NewGitignoreParser creates a new parser for the given repository root
func NewGitignoreParser(root string) *GitignoreParser {
	return &GitignoreParser{root: root}
}

// FindIgnoredEnvFiles discovers environment files based on .gitignore patterns
func (p *GitignoreParser) FindIgnoredEnvFiles() ([]string, error) {
	// Find all .gitignore files in the repository
	gitignoreFiles, err := p.findGitignoreFiles()
	if err != nil {
		return nil, err
	}
	
	// Parse all .gitignore files to get patterns
	patterns := make([]string, 0)
	for _, gitignorePath := range gitignoreFiles {
		filePatterns, err := p.parseGitignoreFile(gitignorePath)
		if err != nil {
			continue // Skip files we can't read
		}
		patterns = append(patterns, filePatterns...)
	}
	
	// Filter patterns that likely represent environment/config files
	envPatterns := p.filterEnvironmentPatterns(patterns)
	
	// Find actual files matching these patterns
	matchedFiles, err := p.findMatchingFiles(envPatterns)
	if err != nil {
		return nil, err
	}
	
	return matchedFiles, nil
}

// findGitignoreFiles recursively finds all .gitignore files
func (p *GitignoreParser) findGitignoreFiles() ([]string, error) {
	var gitignores []string
	
	err := filepath.Walk(p.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}
		
		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		
		if info.Name() == ".gitignore" {
			gitignores = append(gitignores, path)
		}
		
		return nil
	})
	
	return gitignores, err
}

// parseGitignoreFile reads a .gitignore file and returns its patterns
func (p *GitignoreParser) parseGitignoreFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't fail since we already read the file
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", err)
		}
	}()
	
	var patterns []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		patterns = append(patterns, line)
	}
	
	return patterns, scanner.Err()
}

// filterEnvironmentPatterns filters patterns that likely represent env/config files
func (p *GitignoreParser) filterEnvironmentPatterns(patterns []string) []string {
	var envPatterns []string
	
	// Keywords that suggest environment/configuration files
	keywords := []string{
		".env",
		".vars",
		"local",
		"secret",
		"config",
		"settings",
		"credentials",
		".claude",
		".cursor",
		"copilot",
	}
	
	for _, pattern := range patterns {
		// Remove leading slash if present
		pattern = strings.TrimPrefix(pattern, "/")
		
		// Check if pattern contains any environment-related keywords
		lowerPattern := strings.ToLower(pattern)
		for _, keyword := range keywords {
			if strings.Contains(lowerPattern, keyword) {
				envPatterns = append(envPatterns, pattern)
				break
			}
		}
	}
	
	return envPatterns
}

// findMatchingFiles finds actual files that match the given patterns
func (p *GitignoreParser) findMatchingFiles(patterns []string) ([]string, error) {
	matchedFiles := make(map[string]bool)
	
	err := filepath.Walk(p.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}
		
		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Get relative path from root
		relPath, err := filepath.Rel(p.root, path)
		if err != nil {
			return nil
		}
		
		// Check if file matches any pattern
		for _, pattern := range patterns {
			if matchesGitignorePattern(relPath, pattern) {
				matchedFiles[relPath] = true
				break
			}
		}
		
		return nil
	})
	
	// Convert map to slice
	var files []string
	for file := range matchedFiles {
		files = append(files, file)
	}
	
	return files, err
}

// matchesGitignorePattern checks if a path matches a gitignore pattern
func matchesGitignorePattern(path, pattern string) bool {
	// Handle directory patterns
	if strings.HasSuffix(pattern, "/") {
		return false // We're only matching files
	}
	
	// Handle negation patterns
	if strings.HasPrefix(pattern, "!") {
		return false
	}
	
	// Handle patterns starting with /
	if strings.HasPrefix(pattern, "/") {
		pattern = strings.TrimPrefix(pattern, "/")
		// For absolute patterns, match the entire path from root
		matched, _ := filepath.Match(pattern, path)
		return matched
	}
	
	// Handle patterns with ** (recursive)
	if strings.Contains(pattern, "**") {
		// Simplified handling - just check if the filename matches the end pattern
		parts := strings.Split(pattern, "**/")
		if len(parts) > 1 {
			fileName := filepath.Base(path)
			matched, _ := filepath.Match(parts[len(parts)-1], fileName)
			return matched
		}
	}
	
	// Check if pattern matches the full path or just the filename
	matched, _ := filepath.Match(pattern, path)
	if matched {
		return true
	}
	
	// Also check against just the filename
	fileName := filepath.Base(path)
	matched, _ = filepath.Match(pattern, fileName)
	return matched
}

// GetDefaultAIConfigPatterns returns patterns for AI tool configurations
// These are added regardless of .gitignore content
func GetDefaultAIConfigPatterns() []string {
	return []string{
		".claude/settings.local.json",
		".cursorrules",
		".github/copilot/config.json",
		".aider.conf",
		".codeium/config.json",
	}
}