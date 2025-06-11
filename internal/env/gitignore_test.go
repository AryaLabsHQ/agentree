package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitignoreParser_FindGitignoreFiles(t *testing.T) {
	// Create test directory structure
	tmpDir, err := os.MkdirTemp("", "agentree-gitignore-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	// Create multiple .gitignore files
	gitignoreFiles := []string{
		".gitignore",
		"packages/.gitignore",
		"packages/app/.gitignore",
		"src/.gitignore",
	}
	
	for _, file := range gitignoreFiles {
		filePath := filepath.Join(tmpDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte("# test gitignore"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Create .git directory to test it's skipped
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	gitIgnoreInGit := filepath.Join(gitDir, ".gitignore")
	if err := os.WriteFile(gitIgnoreInGit, []byte("# should be skipped"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test finding gitignore files
	parser := NewGitignoreParser(tmpDir)
	files, err := parser.findGitignoreFiles()
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that all expected files were found
	if len(files) != len(gitignoreFiles) {
		t.Errorf("Expected %d .gitignore files, found %d", len(gitignoreFiles), len(files))
	}
	
	// Check that .git directory was skipped
	for _, f := range files {
		if filepath.Dir(f) == gitDir {
			t.Error(".gitignore in .git directory should have been skipped")
		}
	}
}

func TestGitignoreParser_ParseGitignoreFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agentree-parse-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	gitignoreContent := `# Comment line
.env
.env.local

# Another comment
*.secret
/config/local.json
!important.env

# Directories
node_modules/
dist/
`
	
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	parser := NewGitignoreParser(tmpDir)
	patterns, err := parser.parseGitignoreFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	
	expectedPatterns := []string{
		".env",
		".env.local",
		"*.secret",
		"/config/local.json",
		"!important.env",
		"node_modules/",
		"dist/",
	}
	
	if len(patterns) != len(expectedPatterns) {
		t.Errorf("Expected %d patterns, got %d", len(expectedPatterns), len(patterns))
		t.Logf("Patterns: %v", patterns)
	}
	
	// Check each pattern
	patternMap := make(map[string]bool)
	for _, p := range patterns {
		patternMap[p] = true
	}
	
	for _, expected := range expectedPatterns {
		if !patternMap[expected] {
			t.Errorf("Expected pattern %s not found", expected)
		}
	}
}

func TestGitignoreParser_FilterEnvironmentPatterns(t *testing.T) {
	parser := NewGitignoreParser("/tmp")
	
	patterns := []string{
		".env",
		".env.local",
		"*.secret",
		"config/local.json",
		"node_modules/",
		"dist/",
		".claude/settings.local.json",
		"credentials.txt",
		"build/",
		"*.log",
		"settings.json",
		".cursorrules",
	}
	
	filtered := parser.filterEnvironmentPatterns(patterns)
	
	// Should include patterns with environment keywords
	expectedIncluded := []string{
		".env",
		".env.local",
		"*.secret",
		"config/local.json",
		".claude/settings.local.json",
		"credentials.txt",
		"settings.json",
		".cursorrules",
	}
	
	// Should exclude patterns without environment keywords
	expectedExcluded := []string{
		"node_modules/",
		"dist/",
		"build/",
		"*.log",
	}
	
	filteredMap := make(map[string]bool)
	for _, p := range filtered {
		filteredMap[p] = true
	}
	
	for _, expected := range expectedIncluded {
		if !filteredMap[expected] {
			t.Errorf("Pattern %s should have been included", expected)
		}
	}
	
	for _, excluded := range expectedExcluded {
		if filteredMap[excluded] {
			t.Errorf("Pattern %s should have been excluded", excluded)
		}
	}
}

func TestMatchesGitignorePattern(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		// Simple patterns
		{".env", ".env", true},
		{".env.local", ".env.local", true},
		{"src/.env", ".env", true},
		{"packages/app/.env", ".env", true},
		
		// Wildcard patterns
		{".env.local", "*.local", true},
		{"config.local", "*.local", true},
		{"test.txt", "*.local", false},
		
		// Path patterns
		{"config/local.json", "/config/local.json", true},
		{"src/config/local.json", "/config/local.json", false},
		
		// Recursive patterns
		{"packages/app/.env", "**/.env", true},
		{"deep/nested/path/.env", "**/.env", true},
		{".env", "**/.env", true},
		
		// Negation patterns (should return false)
		{".env", "!.env", false},
		
		// Directory patterns (should return false for files)
		{"node_modules", "node_modules/", false},
	}
	
	for _, tt := range tests {
		result := matchesGitignorePattern(tt.path, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchesGitignorePattern(%q, %q) = %v, want %v", 
				tt.path, tt.pattern, result, tt.expected)
		}
	}
}

func TestGitignoreParser_FindIgnoredEnvFiles_Integration(t *testing.T) {
	// Create a realistic test repository
	tmpDir, err := os.MkdirTemp("", "agentree-integration-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	// Create .gitignore
	gitignoreContent := `.env
.env.*
*.local
.dev.vars
.claude/
config/secrets.json
**/.env
`
	
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create matching files
	envFiles := []string{
		".env",
		".env.local",
		".env.production",
		"config.local",
		".dev.vars",
		".claude/settings.local.json",
		"config/secrets.json",
		"packages/api/.env",
		"apps/web/.env.local",
	}
	
	// Create non-matching files
	otherFiles := []string{
		"README.md",
		"package.json",
		"src/index.js",
	}
	
	// Create all files
	allFiles := append(envFiles, otherFiles...)
	for _, file := range allFiles {
		filePath := filepath.Join(tmpDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Test the integration
	parser := NewGitignoreParser(tmpDir)
	foundFiles, err := parser.FindIgnoredEnvFiles()
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that environment files were found
	foundMap := make(map[string]bool)
	for _, f := range foundFiles {
		foundMap[f] = true
		t.Logf("Found: %s", f)
	}
	
	// Most env files should be found (some might be filtered by keywords)
	if len(foundFiles) < 5 {
		t.Errorf("Expected at least 5 env files, found %d", len(foundFiles))
	}
	
	// Check that non-env files were not found
	for _, file := range otherFiles {
		if foundMap[file] {
			t.Errorf("Non-environment file %s should not have been found", file)
		}
	}
}