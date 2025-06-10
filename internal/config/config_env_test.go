package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadProjectConfig_EnvConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "agentree-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	// Create .agentreerc with environment configuration
	agentreercContent := `# Test agentreerc with env config

# Environment configuration
ENV_COPY_ENABLED=true
ENV_RECURSIVE=false
ENV_USE_GITIGNORE=true

# Include patterns
ENV_INCLUDE_PATTERNS=(
  "*.env"
  "*.secrets"
  "config/*.local"
)

# Exclude patterns
ENV_EXCLUDE_PATTERNS=(
  "*.test.env"
  "temp/*"
)

# Post create scripts
POST_CREATE_SCRIPTS=(
  "npm install"
  "npm run setup"
)
`
	
	agentreercPath := filepath.Join(tmpDir, ".agentreerc")
	if err := os.WriteFile(agentreercPath, []byte(agentreercContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Load config
	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	
	// Test basic env config
	if !cfg.EnvConfig.Enabled {
		t.Error("EnvConfig.Enabled should be true")
	}
	
	if cfg.EnvConfig.Recursive {
		t.Error("EnvConfig.Recursive should be false")
	}
	
	if !cfg.EnvConfig.UseGitignore {
		t.Error("EnvConfig.UseGitignore should be true")
	}
	
	// Test include patterns
	expectedInclude := []string{"*.env", "*.secrets", "config/*.local"}
	if !reflect.DeepEqual(cfg.EnvConfig.IncludePatterns, expectedInclude) {
		t.Errorf("IncludePatterns mismatch. Got: %v, Want: %v", 
			cfg.EnvConfig.IncludePatterns, expectedInclude)
	}
	
	// Test exclude patterns
	expectedExclude := []string{"*.test.env", "temp/*"}
	if !reflect.DeepEqual(cfg.EnvConfig.ExcludePatterns, expectedExclude) {
		t.Errorf("ExcludePatterns mismatch. Got: %v, Want: %v", 
			cfg.EnvConfig.ExcludePatterns, expectedExclude)
	}
	
	// Test post create scripts
	expectedScripts := []string{"npm install", "npm run setup"}
	if !reflect.DeepEqual(cfg.PostCreateScripts, expectedScripts) {
		t.Errorf("PostCreateScripts mismatch. Got: %v, Want: %v", 
			cfg.PostCreateScripts, expectedScripts)
	}
}

func TestLoadGlobalConfig_EnvConfig(t *testing.T) {
	// Create temporary home directory
	tmpHome, err := os.MkdirTemp("", "agentree-home-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpHome)
	}()
	
	// Set HOME environment variable temporarily
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpHome)
	defer func() {
		_ = os.Setenv("HOME", oldHome)
	}()
	
	// Create config directory
	configDir := filepath.Join(tmpHome, ".config", "agentree")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create global config
	globalConfigContent := `# Global agentree config

# Package manager setups
PNPM_SETUP=pnpm install --frozen-lockfile
NPM_SETUP=npm ci
YARN_SETUP=yarn install --frozen-lockfile

# Environment configuration
ENV_COPY_ENABLED=false
ENV_RECURSIVE=true
ENV_USE_GITIGNORE=false
ENV_INCLUDE_PATTERNS=.env.global,.secrets
ENV_EXCLUDE_PATTERNS=*.temp,*.bak

# Default post create
DEFAULT_POST_CREATE=echo "Worktree created"
`
	
	configPath := filepath.Join(configDir, "config")
	if err := os.WriteFile(configPath, []byte(globalConfigContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Load config
	cfg, err := LoadGlobalConfig()
	if err != nil {
		t.Fatal(err)
	}
	
	// Test package manager setups
	if cfg.PnpmSetup != "pnpm install --frozen-lockfile" {
		t.Errorf("PnpmSetup mismatch. Got: %s", cfg.PnpmSetup)
	}
	
	// Test env config
	if cfg.EnvConfig.Enabled {
		t.Error("EnvConfig.Enabled should be false")
	}
	
	if !cfg.EnvConfig.Recursive {
		t.Error("EnvConfig.Recursive should be true")
	}
	
	if cfg.EnvConfig.UseGitignore {
		t.Error("EnvConfig.UseGitignore should be false")
	}
	
	// Test comma-separated patterns
	expectedInclude := []string{".env.global", ".secrets"}
	if !reflect.DeepEqual(cfg.EnvConfig.IncludePatterns, expectedInclude) {
		t.Errorf("IncludePatterns mismatch. Got: %v, Want: %v", 
			cfg.EnvConfig.IncludePatterns, expectedInclude)
	}
	
	expectedExclude := []string{"*.temp", "*.bak"}
	if !reflect.DeepEqual(cfg.EnvConfig.ExcludePatterns, expectedExclude) {
		t.Errorf("ExcludePatterns mismatch. Got: %v, Want: %v", 
			cfg.EnvConfig.ExcludePatterns, expectedExclude)
	}
}

func TestMergeConfig(t *testing.T) {
	// Create test configs
	globalCfg := &Config{
		PnpmSetup:    "pnpm install",
		NpmSetup:     "npm install",
		DefaultSetup: "echo global",
		EnvConfig: EnvConfig{
			Enabled:         false,
			Recursive:       false,
			UseGitignore:    false,
			IncludePatterns: []string{"global.env"},
			ExcludePatterns: []string{"global.exclude"},
			CustomPatterns:  []string{"global.custom"},
		},
	}
	
	projectCfg := &Config{
		PostCreateScripts: []string{"npm test"},
		PnpmSetup:        "pnpm install --no-frozen-lockfile",
		EnvConfig: EnvConfig{
			Enabled:         true,
			Recursive:       true,
			UseGitignore:    true,
			IncludePatterns: []string{"project.env"},
			ExcludePatterns: []string{"project.exclude"},
			CustomPatterns:  []string{"project.custom"},
		},
	}
	
	// Test merge
	merged := MergeConfig(globalCfg, projectCfg)
	
	// Project config should override global
	if !merged.EnvConfig.Enabled {
		t.Error("EnvConfig.Enabled should be true (from project)")
	}
	
	if !merged.EnvConfig.Recursive {
		t.Error("EnvConfig.Recursive should be true (from project)")
	}
	
	if !merged.EnvConfig.UseGitignore {
		t.Error("EnvConfig.UseGitignore should be true (from project)")
	}
	
	// Package manager setup should be overridden
	if merged.PnpmSetup != "pnpm install --no-frozen-lockfile" {
		t.Errorf("PnpmSetup should be from project config. Got: %s", merged.PnpmSetup)
	}
	
	// npm setup should come from global (not overridden)
	if merged.NpmSetup != "npm install" {
		t.Errorf("NpmSetup should be from global config. Got: %s", merged.NpmSetup)
	}
	
	// Patterns should be combined (both global and project)
	expectedInclude := []string{"global.env", "project.env"}
	if !reflect.DeepEqual(merged.EnvConfig.IncludePatterns, expectedInclude) {
		t.Errorf("IncludePatterns should combine both. Got: %v, Want: %v", 
			merged.EnvConfig.IncludePatterns, expectedInclude)
	}
	
	expectedExclude := []string{"global.exclude", "project.exclude"}
	if !reflect.DeepEqual(merged.EnvConfig.ExcludePatterns, expectedExclude) {
		t.Errorf("ExcludePatterns should combine both. Got: %v, Want: %v", 
			merged.EnvConfig.ExcludePatterns, expectedExclude)
	}
	
	// Custom patterns from project should replace global
	expectedCustom := []string{"project.custom"}
	if !reflect.DeepEqual(merged.EnvConfig.CustomPatterns, expectedCustom) {
		t.Errorf("CustomPatterns should be from project only. Got: %v, Want: %v", 
			merged.EnvConfig.CustomPatterns, expectedCustom)
	}
	
	// Post create scripts should come from project
	if !reflect.DeepEqual(merged.PostCreateScripts, []string{"npm test"}) {
		t.Errorf("PostCreateScripts mismatch. Got: %v", merged.PostCreateScripts)
	}
}

func TestMergeConfig_Defaults(t *testing.T) {
	// Test with nil configs
	merged := MergeConfig(nil, nil)
	
	// Should have defaults
	if !merged.EnvConfig.Enabled {
		t.Error("Default EnvConfig.Enabled should be true")
	}
	
	if !merged.EnvConfig.Recursive {
		t.Error("Default EnvConfig.Recursive should be true")
	}
	
	if !merged.EnvConfig.UseGitignore {
		t.Error("Default EnvConfig.UseGitignore should be true")
	}
	
	// Test with only global config
	globalCfg := &Config{
		EnvConfig: EnvConfig{
			Enabled: false,
		},
	}
	
	merged = MergeConfig(globalCfg, nil)
	if merged.EnvConfig.Enabled {
		t.Error("Global config should override default")
	}
}