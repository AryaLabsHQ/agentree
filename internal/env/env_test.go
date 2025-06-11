package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvFileCopier_DiscoverFiles(t *testing.T) {
	// Create a test directory structure
	tmpDir, err := os.MkdirTemp("", "agentree-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	destDir, err := os.MkdirTemp("", "agentree-dest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(destDir)
	}()
	
	// Create test files
	testFiles := []string{
		".env",
		".env.local",
		".env.development",
		".env.production.local",
		".dev.vars",
		".claude/settings.local.json",
		"packages/app/.env",
		"packages/api/.env.local",
		"src/config.json", // Should not be copied
	}
	
	// Create .gitignore with patterns
	gitignoreContent := `
# Environment files
.env
.env.*
.env.local
.env.*.local
.dev.vars

# AI tool configs
.claude/settings.local.json

# Nested env files
**/.env
**/.env.local
`
	
	// Create directories and files
	for _, file := range testFiles {
		filePath := filepath.Join(tmpDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Create .gitignore
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Test discovery
	copier := NewEnvFileCopier(tmpDir, destDir)
	files, err := copier.DiscoverFiles()
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that all expected files were discovered
	expectedFiles := []string{
		".env",
		".env.local",
		".env.development",
		".env.production.local",
		".dev.vars",
		".claude/settings.local.json",
		"packages/app/.env",
		"packages/api/.env.local",
	}
	
	if len(files) < len(expectedFiles) {
		t.Errorf("Expected at least %d files, got %d", len(expectedFiles), len(files))
		t.Logf("Files found: %v", files)
	}
	
	// Check that each expected file was found
	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}
	
	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("Expected file %s was not discovered", expected)
		}
	}
	
	// Check that src/config.json was NOT discovered
	if fileMap["src/config.json"] {
		t.Error("src/config.json should not have been discovered")
	}
}

func TestEnvFileCopier_CopyFiles(t *testing.T) {
	// Create source and destination directories
	srcDir, err := os.MkdirTemp("", "agentree-src-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(srcDir)
	}()
	
	destDir, err := os.MkdirTemp("", "agentree-dest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(destDir)
	}()
	
	// Create test files with specific content
	testFiles := map[string]string{
		".env":                        "API_KEY=secret123",
		".env.local":                  "LOCAL_VAR=local_value",
		"packages/app/.env":           "APP_ENV=development",
		".claude/settings.local.json": `{"theme": "dark"}`,
	}
	
	for file, content := range testFiles {
		filePath := filepath.Join(srcDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Copy files
	copier := NewEnvFileCopier(srcDir, destDir)
	filesToCopy := []string{".env", ".env.local", "packages/app/.env", ".claude/settings.local.json"}
	copiedFiles, err := copier.CopyFiles(filesToCopy)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify all files were copied
	if len(copiedFiles) != len(filesToCopy) {
		t.Errorf("Expected %d files to be copied, got %d", len(filesToCopy), len(copiedFiles))
	}
	
	// Verify content of copied files
	for file, expectedContent := range testFiles {
		destPath := filepath.Join(destDir, file)
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", file, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("File %s content mismatch. Expected: %s, Got: %s", file, expectedContent, string(content))
		}
	}
	
	// Verify directory structure was preserved
	if _, err := os.Stat(filepath.Join(destDir, "packages/app/.env")); os.IsNotExist(err) {
		t.Error("Nested directory structure was not preserved")
	}
	if _, err := os.Stat(filepath.Join(destDir, ".claude/settings.local.json")); os.IsNotExist(err) {
		t.Error(".claude directory structure was not preserved")
	}
}

func TestEnvFileCopier_CustomPatterns(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "agentree-src-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(srcDir)
	}()
	
	destDir, err := os.MkdirTemp("", "agentree-dest-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(destDir)
	}()
	
	// Create custom config files
	customFiles := []string{
		"custom.config",
		"app.secrets",
		"nested/custom.config",
	}
	
	for _, file := range customFiles {
		filePath := filepath.Join(srcDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte("custom content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Test with custom patterns
	copier := NewEnvFileCopier(srcDir, destDir)
	copier.AddCustomPatterns([]string{"*.config", "*.secrets", "**/custom.config"})
	
	files, err := copier.DiscoverFiles()
	if err != nil {
		t.Fatal(err)
	}
	
	// Check that custom files were discovered
	fileMap := make(map[string]bool)
	for _, f := range files {
		fileMap[f] = true
	}
	
	for _, expected := range customFiles {
		if !fileMap[expected] {
			t.Errorf("Expected custom file %s was not discovered", expected)
		}
	}
}