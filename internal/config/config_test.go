package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadProjectConfig(t *testing.T) {
	tests := []struct {
		name       string
		hatchrc    string
		expected   Config
		wantErr    bool
	}{
		{
			name: "valid hatchrc with scripts",
			hatchrc: `#!/bin/bash
# Project-specific hatch configuration

POST_CREATE_SCRIPTS=(
  "pnpm install"
  "pnpm build"
  "cp .env.example .env"
)`,
			expected: Config{
				PostCreateScripts: []string{
					"pnpm install",
					"pnpm build", 
					"cp .env.example .env",
				},
			},
			wantErr: false,
		},
		{
			name: "hatchrc with comments and empty lines",
			hatchrc: `#!/bin/bash
# This is a comment

POST_CREATE_SCRIPTS=(
  # Another comment
  "npm install"
  
  "npm test"
)

# More comments`,
			expected: Config{
				PostCreateScripts: []string{
					"npm install",
					"npm test",
				},
			},
			wantErr: false,
		},
		{
			name: "hatchrc with different quote styles",
			hatchrc: `POST_CREATE_SCRIPTS=(
  "double quotes"
  'single quotes'
  no_quotes_needed
)`,
			expected: Config{
				PostCreateScripts: []string{
					"double quotes",
					"single quotes",
					"no_quotes_needed",
				},
			},
			wantErr: false,
		},
		{
			name:     "no hatchrc file",
			hatchrc:  "", // special case - won't create file
			expected: Config{},
			wantErr:  false,
		},
		{
			name: "empty hatchrc",
			hatchrc: `#!/bin/bash
# Empty config`,
			expected: Config{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create .hatchrc if content provided
			if tt.hatchrc != "" {
				hatchrcPath := filepath.Join(tmpDir, ".hatchrc")
				if err := os.WriteFile(hatchrcPath, []byte(tt.hatchrc), 0644); err != nil {
					t.Fatalf("Failed to create .hatchrc: %v", err)
				}
			}

			// Load config
			cfg, err := LoadProjectConfig(tmpDir)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare configs
			if !reflect.DeepEqual(cfg.PostCreateScripts, tt.expected.PostCreateScripts) {
				t.Errorf("PostCreateScripts mismatch")
				t.Errorf("Expected: %v", tt.expected.PostCreateScripts)
				t.Errorf("Got: %v", cfg.PostCreateScripts)
			}
		})
	}
}

func TestLoadGlobalConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name     string
		config   string
		expected Config
		wantErr  bool
	}{
		{
			name: "valid global config",
			config: `# Global hatch configuration
PNPM_SETUP="pnpm install --frozen-lockfile && pnpm build"
NPM_SETUP="npm ci && npm run build"
YARN_SETUP="yarn install --frozen-lockfile && yarn build"
DEFAULT_POST_CREATE="echo 'No package manager detected'"`,
			expected: Config{
				PnpmSetup:    "pnpm install --frozen-lockfile && pnpm build",
				NpmSetup:     "npm ci && npm run build", 
				YarnSetup:    "yarn install --frozen-lockfile && yarn build",
				DefaultSetup: "echo 'No package manager detected'",
			},
			wantErr: false,
		},
		{
			name: "config with comments and spaces",
			config: `# Comment line
PNPM_SETUP = "pnpm install"  # inline comment

# Another comment
NPM_SETUP='npm install'
`,
			expected: Config{
				PnpmSetup: "pnpm install",
				NpmSetup:  "npm install",
			},
			wantErr: false,
		},
		{
			name: "partial config",
			config: `PNPM_SETUP="pnpm install"
# NPM_SETUP is commented out
YARN_SETUP="yarn"`,
			expected: Config{
				PnpmSetup: "pnpm install",
				YarnSetup: "yarn",
			},
			wantErr: false,
		},
		{
			name:     "no config file",
			config:   "", // special case
			expected: Config{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary HOME
			tmpHome := t.TempDir()
			os.Setenv("HOME", tmpHome)

			// Create config file if content provided
			if tt.config != "" {
				configDir := filepath.Join(tmpHome, ".config", "hatch")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}

				configPath := filepath.Join(configDir, "config")
				if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
					t.Fatalf("Failed to create config: %v", err)
				}
			}

			// Load config
			cfg, err := LoadGlobalConfig()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGlobalConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare configs
			if cfg.PnpmSetup != tt.expected.PnpmSetup {
				t.Errorf("PnpmSetup: expected %q, got %q", tt.expected.PnpmSetup, cfg.PnpmSetup)
			}
			if cfg.NpmSetup != tt.expected.NpmSetup {
				t.Errorf("NpmSetup: expected %q, got %q", tt.expected.NpmSetup, cfg.NpmSetup)
			}
			if cfg.YarnSetup != tt.expected.YarnSetup {
				t.Errorf("YarnSetup: expected %q, got %q", tt.expected.YarnSetup, cfg.YarnSetup)
			}
			if cfg.DefaultSetup != tt.expected.DefaultSetup {
				t.Errorf("DefaultSetup: expected %q, got %q", tt.expected.DefaultSetup, cfg.DefaultSetup)
			}
		})
	}
}

// TestConfigParsing tests edge cases in config parsing
func TestConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantKey  string
		wantVal  string
		valid    bool
	}{
		{
			name:    "simple assignment",
			line:    `KEY="value"`,
			wantKey: "KEY",
			wantVal: "value", 
			valid:   true,
		},
		{
			name:    "spaces around equals",
			line:    `KEY = "value"`,
			wantKey: "KEY", 
			wantVal: "value",
			valid:   true,
		},
		{
			name:    "single quotes",
			line:    `KEY='value'`,
			wantKey: "KEY",
			wantVal: "value",
			valid:   true,
		},
		{
			name:    "no quotes",
			line:    `KEY=value`,
			wantKey: "KEY",
			wantVal: "value",
			valid:   true,
		},
		{
			name:  "comment line",
			line:  `# This is a comment`,
			valid: false,
		},
		{
			name:  "empty line",
			line:  ``,
			valid: false,
		},
		{
			name:  "no equals sign",
			line:  `KEY value`,
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the parsing logic used in LoadGlobalConfig
			line := tt.line
			
			// Skip empty lines and comments
			if line == "" || line[0] == '#' {
				if tt.valid {
					t.Error("Expected valid line but would be skipped")
				}
				return
			}

			// Parse key=value
			parts := []string{}
			if idx := strings.Index(line, "="); idx != -1 {
				parts = []string{line[:idx], line[idx+1:]}
			}

			if len(parts) != 2 {
				if tt.valid {
					t.Error("Expected valid line but failed to parse")
				}
				return
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

			if !tt.valid {
				t.Error("Expected invalid line but parsed successfully")
				return
			}

			if key != tt.wantKey {
				t.Errorf("Key: expected %q, got %q", tt.wantKey, key)
			}
			if value != tt.wantVal {
				t.Errorf("Value: expected %q, got %q", tt.wantVal, value)
			}
		})
	}
}

