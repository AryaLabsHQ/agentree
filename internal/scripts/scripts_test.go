package scripts

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetermineScripts(t *testing.T) {
	tests := []struct {
		name            string
		customScripts   []string
		projectScripts  []string
		detectedScripts []string
		globalOverride  string
		expected        []string
	}{
		{
			name:            "custom scripts take precedence",
			customScripts:   []string{"echo custom1", "echo custom2"},
			projectScripts:  []string{"echo project"},
			detectedScripts: []string{"echo detected"},
			globalOverride:  "echo global",
			expected:        []string{"echo custom1", "echo custom2"},
		},
		{
			name:            "project scripts over detected",
			customScripts:   []string{},
			projectScripts:  []string{"npm install", "npm test"},
			detectedScripts: []string{"npm install"},
			globalOverride:  "npm ci",
			expected:        []string{"npm install", "npm test"},
		},
		{
			name:            "global override over detected",
			customScripts:   []string{},
			projectScripts:  []string{},
			detectedScripts: []string{"pnpm install"},
			globalOverride:  "pnpm install --frozen-lockfile && pnpm build",
			expected:        []string{"pnpm install --frozen-lockfile", "pnpm build"},
		},
		{
			name:            "detected scripts as fallback",
			customScripts:   []string{},
			projectScripts:  []string{},
			detectedScripts: []string{"go mod download", "go build"},
			globalOverride:  "",
			expected:        []string{"go mod download", "go build"},
		},
		{
			name:            "empty when nothing provided",
			customScripts:   []string{},
			projectScripts:  []string{},
			detectedScripts: []string{},
			globalOverride:  "",
			expected:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineScripts(
				tt.customScripts,
				tt.projectScripts,
				tt.detectedScripts,
				tt.globalOverride,
			)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d scripts, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", result)
				return
			}

			for i, script := range result {
				if script != tt.expected[i] {
					t.Errorf("Script %d: expected %q, got %q", i, tt.expected[i], script)
				}
			}
		})
	}
}

func TestParseScriptString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single command",
			input:    "npm install",
			expected: []string{"npm install"},
		},
		{
			name:     "multiple commands with &&",
			input:    "npm install && npm test && npm build",
			expected: []string{"npm install", "npm test", "npm build"},
		},
		{
			name:     "commands with extra spaces",
			input:    "  npm install   &&   npm test  ",
			expected: []string{"npm install", "npm test"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only spaces",
			input:    "   &&   ",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseScriptString(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d commands, got %d", len(tt.expected), len(result))
				t.Errorf("Expected: %v", tt.expected)
				t.Errorf("Got: %v", result)
				return
			}

			for i, cmd := range result {
				if cmd != tt.expected[i] {
					t.Errorf("Command %d: expected %q, got %q", i, tt.expected[i], cmd)
				}
			}
		})
	}
}

func TestRunScripts(t *testing.T) {
	tests := []struct {
		name    string
		scripts []string
		files   map[string]string // files to create before running
		wantErr bool
	}{
		{
			name: "successful scripts",
			scripts: []string{
				"echo 'Hello World' > output.txt",
				"echo 'Second line' >> output.txt",
			},
			wantErr: false,
		},
		{
			name: "script with error continues",
			scripts: []string{
				"echo 'First' > first.txt",
				"false", // This will fail
				"echo 'Third' > third.txt",
			},
			wantErr: false, // RunScripts doesn't return error, just logs
		},
		{
			name:    "empty scripts",
			scripts: []string{},
			wantErr: false,
		},
		{
			name: "script reading file",
			scripts: []string{
				"cat input.txt > output.txt",
			},
			files: map[string]string{
				"input.txt": "test content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			// Create any required files
			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Capture output (optional, for debugging)
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			defer func() {
				os.Stdout = oldStdout
				os.Stderr = oldStderr
			}()

			// Run scripts
			runner := NewRunner(tmpDir)
			err := runner.RunScripts(tt.scripts)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunScripts() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify files were created (for successful scripts)
			if tt.name == "successful scripts" {
				outputPath := filepath.Join(tmpDir, "output.txt")
				if content, err := os.ReadFile(outputPath); err != nil {
					t.Errorf("Expected output.txt to be created")
				} else {
					expected := "Hello World\nSecond line\n"
					if string(content) != expected {
						t.Errorf("output.txt content mismatch: got %q, want %q", string(content), expected)
					}
				}
			}
		})
	}
}

func TestRunScript(t *testing.T) {
	tmpDir := t.TempDir()
	runner := NewRunner(tmpDir)

	// Test successful command
	err := runner.runScript("echo 'test' > test.txt")
	if err != nil {
		t.Errorf("Expected successful script to pass: %v", err)
	}

	// Verify file was created
	testFile := filepath.Join(tmpDir, "test.txt")
	if content, err := os.ReadFile(testFile); err != nil {
		t.Errorf("Expected test.txt to be created")
	} else if strings.TrimSpace(string(content)) != "test" {
		t.Errorf("Expected 'test', got %q", string(content))
	}

	// Test failing command
	err = runner.runScript("exit 1")
	if err == nil {
		t.Errorf("Expected failing script to return error")
	}
}

// TestScriptOutput verifies that script output is properly displayed
func TestScriptOutput(t *testing.T) {
	// This test captures stdout to verify output formatting
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tmpDir := t.TempDir()
	runner := NewRunner(tmpDir)
	runner.RunScripts([]string{"echo 'Hello from script'"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check for expected output patterns
	if !strings.Contains(output, "Running post-create scripts") {
		t.Error("Expected header message")
	}
	if !strings.Contains(output, "echo 'Hello from script'") {
		t.Error("Expected script command to be shown")
	}
	if !strings.Contains(output, "Success") {
		t.Error("Expected success message")
	}
}