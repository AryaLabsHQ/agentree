package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyEnvFiles(t *testing.T) {
	tests := []struct {
		name         string
		sourceFiles  map[string]string
		expectedCopy []string
		wantErr      bool
	}{
		{
			name: "copy both env files",
			sourceFiles: map[string]string{
				".env":      "API_KEY=secret123",
				".dev.vars": "DEV_MODE=true",
			},
			expectedCopy: []string{".env", ".dev.vars"},
			wantErr:      false,
		},
		{
			name: "only .env exists",
			sourceFiles: map[string]string{
				".env": "API_KEY=secret123",
			},
			expectedCopy: []string{".env"},
			wantErr:      false,
		},
		{
			name: "only .dev.vars exists",
			sourceFiles: map[string]string{
				".dev.vars": "DEV_MODE=true",
			},
			expectedCopy: []string{".dev.vars"},
			wantErr:      false,
		},
		{
			name:         "no env files",
			sourceFiles:  map[string]string{},
			expectedCopy: []string{},
			wantErr:      false,
		},
		{
			name: "other files are ignored",
			sourceFiles: map[string]string{
				".env":        "API_KEY=secret123",
				"config.json": `{"key": "value"}`,
				".gitignore":  "node_modules/",
			},
			expectedCopy: []string{".env"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create source and destination directories
			srcDir := t.TempDir()
			destDir := t.TempDir()

			// Create source files
			for filename, content := range tt.sourceFiles {
				path := filepath.Join(srcDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create source file %s: %v", filename, err)
				}
			}

			// Run the copy operation
			copied, err := CopyEnvFiles(srcDir, destDir)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyEnvFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check copied files
			if len(copied) != len(tt.expectedCopy) {
				t.Errorf("Expected %d files copied, got %d", len(tt.expectedCopy), len(copied))
				t.Errorf("Expected: %v, Got: %v", tt.expectedCopy, copied)
				return
			}

			// Verify files were actually copied
			for _, filename := range tt.expectedCopy {
				destPath := filepath.Join(destDir, filename)
				srcPath := filepath.Join(srcDir, filename)

				// Check file exists
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Errorf("Expected file %s to be copied, but it doesn't exist", filename)
					continue
				}

				// Check content matches
				srcContent, _ := os.ReadFile(srcPath)
				destContent, _ := os.ReadFile(destPath)
				if string(srcContent) != string(destContent) {
					t.Errorf("Content mismatch for %s", filename)
				}

				// Check permissions match
				srcInfo, _ := os.Stat(srcPath)
				destInfo, _ := os.Stat(destPath)
				if srcInfo.Mode() != destInfo.Mode() {
					t.Errorf("Permission mismatch for %s: src=%v, dest=%v", 
						filename, srcInfo.Mode(), destInfo.Mode())
				}
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Test the internal copyFile function
	srcDir := t.TempDir()
	destDir := t.TempDir()

	// Create source file with specific permissions
	srcPath := filepath.Join(srcDir, "test.txt")
	content := "test content"
	if err := os.WriteFile(srcPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	destPath := filepath.Join(destDir, "test.txt")

	// Copy the file
	if err := copyFile(srcPath, destPath); err != nil {
		t.Fatalf("copyFile() failed: %v", err)
	}

	// Verify content
	destContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}
	if string(destContent) != content {
		t.Errorf("Content mismatch: expected %q, got %q", content, string(destContent))
	}

	// Verify permissions
	srcInfo, _ := os.Stat(srcPath)
	destInfo, _ := os.Stat(destPath)
	if srcInfo.Mode() != destInfo.Mode() {
		t.Errorf("Permission mismatch: src=%v, dest=%v", srcInfo.Mode(), destInfo.Mode())
	}
}

func TestCopyFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, string)
		wantErr bool
	}{
		{
			name: "source file doesn't exist",
			setup: func() (string, string) {
				return "/nonexistent/file", t.TempDir() + "/dest"
			},
			wantErr: true,
		},
		{
			name: "destination directory doesn't exist",
			setup: func() (string, string) {
				srcDir := t.TempDir()
				srcPath := filepath.Join(srcDir, "test.txt")
				if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return srcPath, "/nonexistent/dir/dest.txt"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dest := tt.setup()
			err := copyFile(src, dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}