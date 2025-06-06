package detector

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectSetupCommands tests the auto-detection of package managers
func TestDetectSetupCommands(t *testing.T) {
	// Table-driven tests are a Go idiom for testing multiple scenarios
	tests := []struct {
		name     string
		files    map[string]string // files to create
		expected []string         // expected commands
	}{
		{
			name: "pnpm project with build script",
			files: map[string]string{
				"pnpm-lock.yaml": "",
				"package.json": `{
					"scripts": {
						"build": "tsc"
					}
				}`,
			},
			expected: []string{"pnpm install", "pnpm build"},
		},
		{
			name: "npm project without build script",
			files: map[string]string{
				"package-lock.json": "",
				"package.json":      `{"scripts": {}}`,
			},
			expected: []string{"npm install"},
		},
		{
			name: "yarn project with build",
			files: map[string]string{
				"yarn.lock": "",
				"package.json": `{
					"scripts": {
						"build": "webpack"
					}
				}`,
			},
			expected: []string{"yarn install", "yarn build"},
		},
		{
			name: "cargo project",
			files: map[string]string{
				"Cargo.lock": "",
			},
			expected: []string{"cargo build"},
		},
		{
			name: "go project",
			files: map[string]string{
				"go.mod": "",
			},
			expected: []string{"go mod download"},
		},
		{
			name: "python project",
			files: map[string]string{
				"requirements.txt": "",
			},
			expected: []string{"pip install -r requirements.txt"},
		},
		{
			name: "ruby project",
			files: map[string]string{
				"Gemfile.lock": "",
			},
			expected: []string{"bundle install"},
		},
		{
			name:     "no recognized files",
			files:    map[string]string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		// t.Run creates a sub-test for each test case
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tmpDir := t.TempDir() // Automatically cleaned up after test
			
			// Create test files
			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}
			
			// Run the detection
			result := DetectSetupCommands(tmpDir)
			
			// Compare results
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

// TestHasNpmScript tests the npm script detection
func TestHasNpmScript(t *testing.T) {
	tests := []struct {
		name        string
		packageJSON string
		script      string
		expected    bool
	}{
		{
			name: "has build script",
			packageJSON: `{
				"scripts": {
					"build": "webpack",
					"test": "jest"
				}
			}`,
			script:   "build",
			expected: true,
		},
		{
			name: "missing build script",
			packageJSON: `{
				"scripts": {
					"test": "jest"
				}
			}`,
			script:   "build",
			expected: false,
		},
		{
			name:        "no scripts section",
			packageJSON: `{}`,
			script:      "build",
			expected:    false,
		},
		{
			name:        "invalid json",
			packageJSON: `{invalid}`,
			script:      "build",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create package.json
			if err := os.WriteFile(
				filepath.Join(tmpDir, "package.json"),
				[]byte(tt.packageJSON),
				0644,
			); err != nil {
				t.Fatalf("Failed to create package.json: %v", err)
			}
			
			result := hasNpmScript(tmpDir, tt.script)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}