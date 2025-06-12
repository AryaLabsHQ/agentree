// Package config handles agentree configuration from .agentreerc and global config
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds agentree configuration
type Config struct {
	PostCreateScripts []string
	// Package manager specific overrides
	PnpmSetup    string
	NpmSetup     string
	YarnSetup    string
	DefaultSetup string
	
	// Environment file configuration
	EnvConfig EnvConfig
}

// EnvConfig holds environment file copying configuration
type EnvConfig struct {
	// Whether to copy environment files (default: true)
	Enabled bool
	// Additional patterns to include beyond gitignore
	IncludePatterns []string
	// Patterns to exclude even if found in gitignore
	ExcludePatterns []string
	// Whether to search recursively in monorepos (default: true)
	Recursive bool
	// Whether to use gitignore as source of truth (default: true)
	UseGitignore bool
	// Custom environment file patterns (overrides defaults if set)
	CustomPatterns []string
	// Whether to sync environment files back on remove (default: false)
	SyncBackOnRemove bool
	// Patterns for files to sync back (if empty, uses IncludePatterns)
	SyncPatterns []string
}

// LoadProjectConfig loads configuration from .agentreerc in the project root
func LoadProjectConfig(projectRoot string) (*Config, error) {
	cfg := &Config{
		// Set defaults for env config
		EnvConfig: EnvConfig{
			Enabled:      true,
			Recursive:    true,
			UseGitignore: true,
		},
	}
	agentreercPath := filepath.Join(projectRoot, ".agentreerc")

	file, err := os.Open(agentreercPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // No config file is okay
		}
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't fail since we already read the file
			fmt.Fprintf(os.Stderr, "Warning: failed to close config file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	inPostCreateScripts := false
	inEnvIncludePatterns := false
	inEnvExcludePatterns := false
	inEnvSyncPatterns := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle array endings
		if strings.Contains(line, ")") {
			if inPostCreateScripts {
				inPostCreateScripts = false
			} else if inEnvIncludePatterns {
				inEnvIncludePatterns = false
			} else if inEnvExcludePatterns {
				inEnvExcludePatterns = false
			} else if inEnvSyncPatterns {
				inEnvSyncPatterns = false
			}
			continue
		}

		// Look for POST_CREATE_SCRIPTS array
		if strings.Contains(line, "POST_CREATE_SCRIPTS=(") {
			inPostCreateScripts = true
			continue
		}

		// Look for ENV_INCLUDE_PATTERNS array
		if strings.Contains(line, "ENV_INCLUDE_PATTERNS=(") {
			inEnvIncludePatterns = true
			continue
		}

		// Look for ENV_EXCLUDE_PATTERNS array
		if strings.Contains(line, "ENV_EXCLUDE_PATTERNS=(") {
			inEnvExcludePatterns = true
			continue
		}

		// Look for ENV_SYNC_PATTERNS array
		if strings.Contains(line, "ENV_SYNC_PATTERNS=(") {
			inEnvSyncPatterns = true
			continue
		}

		// Handle array contents
		if inPostCreateScripts {
			// Extract script from quotes
			script := strings.Trim(line, ` "',`)
			if script != "" {
				cfg.PostCreateScripts = append(cfg.PostCreateScripts, script)
			}
		} else if inEnvIncludePatterns {
			pattern := strings.Trim(line, ` "',`)
			if pattern != "" {
				cfg.EnvConfig.IncludePatterns = append(cfg.EnvConfig.IncludePatterns, pattern)
			}
		} else if inEnvExcludePatterns {
			pattern := strings.Trim(line, ` "',`)
			if pattern != "" {
				cfg.EnvConfig.ExcludePatterns = append(cfg.EnvConfig.ExcludePatterns, pattern)
			}
		} else if inEnvSyncPatterns {
			pattern := strings.Trim(line, ` "',`)
			if pattern != "" {
				cfg.EnvConfig.SyncPatterns = append(cfg.EnvConfig.SyncPatterns, pattern)
			}
		} else {
			// Handle key=value pairs for env config
			if strings.HasPrefix(line, "ENV_") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					
					switch key {
					case "ENV_COPY_ENABLED":
						cfg.EnvConfig.Enabled = value == "true" || value == "1"
					case "ENV_RECURSIVE":
						cfg.EnvConfig.Recursive = value == "true" || value == "1"
					case "ENV_USE_GITIGNORE":
						cfg.EnvConfig.UseGitignore = value == "true" || value == "1"
					case "ENV_SYNC_BACK_ON_REMOVE":
						cfg.EnvConfig.SyncBackOnRemove = value == "true" || value == "1"
					}
				}
			}
		}
	}

	return cfg, scanner.Err()
}

// LoadGlobalConfig loads configuration from ~/.config/agentree/config
func LoadGlobalConfig() (*Config, error) {
	cfg := &Config{
		// Set defaults for env config
		EnvConfig: EnvConfig{
			Enabled:      true,
			Recursive:    true,
			UseGitignore: true,
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil // Ignore if can't get home dir
	}

	configPath := filepath.Join(homeDir, ".config", "agentree", "config")

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // No config file is okay
		}
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't fail since we already read the file
			fmt.Fprintf(os.Stderr, "Warning: failed to close config file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if they match
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		switch key {
		case "PNPM_SETUP":
			cfg.PnpmSetup = value
		case "NPM_SETUP":
			cfg.NpmSetup = value
		case "YARN_SETUP":
			cfg.YarnSetup = value
		case "DEFAULT_POST_CREATE":
			cfg.DefaultSetup = value
		case "ENV_COPY_ENABLED":
			cfg.EnvConfig.Enabled = value == "true" || value == "1"
		case "ENV_RECURSIVE":
			cfg.EnvConfig.Recursive = value == "true" || value == "1"
		case "ENV_USE_GITIGNORE":
			cfg.EnvConfig.UseGitignore = value == "true" || value == "1"
		case "ENV_INCLUDE_PATTERNS":
			// Support comma-separated patterns in global config
			patterns := strings.Split(value, ",")
			for _, pattern := range patterns {
				pattern = strings.TrimSpace(pattern)
				if pattern != "" {
					cfg.EnvConfig.IncludePatterns = append(cfg.EnvConfig.IncludePatterns, pattern)
				}
			}
		case "ENV_EXCLUDE_PATTERNS":
			// Support comma-separated patterns in global config
			patterns := strings.Split(value, ",")
			for _, pattern := range patterns {
				pattern = strings.TrimSpace(pattern)
				if pattern != "" {
					cfg.EnvConfig.ExcludePatterns = append(cfg.EnvConfig.ExcludePatterns, pattern)
				}
			}
		case "ENV_SYNC_BACK_ON_REMOVE":
			cfg.EnvConfig.SyncBackOnRemove = value == "true" || value == "1"
		case "ENV_SYNC_PATTERNS":
			// Support comma-separated patterns in global config
			patterns := strings.Split(value, ",")
			for _, pattern := range patterns {
				pattern = strings.TrimSpace(pattern)
				if pattern != "" {
					cfg.EnvConfig.SyncPatterns = append(cfg.EnvConfig.SyncPatterns, pattern)
				}
			}
		}
	}

	return cfg, scanner.Err()
}

// MergeConfig merges configurations with proper precedence:
// CLI flags > project config > global config > defaults
func MergeConfig(globalCfg, projectCfg *Config) *Config {
	merged := &Config{
		// Start with defaults
		EnvConfig: EnvConfig{
			Enabled:      true,
			Recursive:    true,
			UseGitignore: true,
		},
	}
	
	// Apply global config
	if globalCfg != nil {
		if globalCfg.PnpmSetup != "" {
			merged.PnpmSetup = globalCfg.PnpmSetup
		}
		if globalCfg.NpmSetup != "" {
			merged.NpmSetup = globalCfg.NpmSetup
		}
		if globalCfg.YarnSetup != "" {
			merged.YarnSetup = globalCfg.YarnSetup
		}
		if globalCfg.DefaultSetup != "" {
			merged.DefaultSetup = globalCfg.DefaultSetup
		}
		
		// Merge env config
		merged.EnvConfig.Enabled = globalCfg.EnvConfig.Enabled
		merged.EnvConfig.Recursive = globalCfg.EnvConfig.Recursive
		merged.EnvConfig.UseGitignore = globalCfg.EnvConfig.UseGitignore
		merged.EnvConfig.SyncBackOnRemove = globalCfg.EnvConfig.SyncBackOnRemove
		merged.EnvConfig.IncludePatterns = append(merged.EnvConfig.IncludePatterns, globalCfg.EnvConfig.IncludePatterns...)
		merged.EnvConfig.ExcludePatterns = append(merged.EnvConfig.ExcludePatterns, globalCfg.EnvConfig.ExcludePatterns...)
		merged.EnvConfig.CustomPatterns = append(merged.EnvConfig.CustomPatterns, globalCfg.EnvConfig.CustomPatterns...)
		merged.EnvConfig.SyncPatterns = append(merged.EnvConfig.SyncPatterns, globalCfg.EnvConfig.SyncPatterns...)
	}
	
	// Apply project config (overrides global)
	if projectCfg != nil {
		merged.PostCreateScripts = projectCfg.PostCreateScripts
		
		if projectCfg.PnpmSetup != "" {
			merged.PnpmSetup = projectCfg.PnpmSetup
		}
		if projectCfg.NpmSetup != "" {
			merged.NpmSetup = projectCfg.NpmSetup
		}
		if projectCfg.YarnSetup != "" {
			merged.YarnSetup = projectCfg.YarnSetup
		}
		if projectCfg.DefaultSetup != "" {
			merged.DefaultSetup = projectCfg.DefaultSetup
		}
		
		// Project env config overrides global
		merged.EnvConfig.Enabled = projectCfg.EnvConfig.Enabled
		merged.EnvConfig.Recursive = projectCfg.EnvConfig.Recursive
		merged.EnvConfig.UseGitignore = projectCfg.EnvConfig.UseGitignore
		merged.EnvConfig.SyncBackOnRemove = projectCfg.EnvConfig.SyncBackOnRemove
		
		// Append patterns (don't replace, allow both to contribute)
		merged.EnvConfig.IncludePatterns = append(merged.EnvConfig.IncludePatterns, projectCfg.EnvConfig.IncludePatterns...)
		merged.EnvConfig.ExcludePatterns = append(merged.EnvConfig.ExcludePatterns, projectCfg.EnvConfig.ExcludePatterns...)
		merged.EnvConfig.SyncPatterns = append(merged.EnvConfig.SyncPatterns, projectCfg.EnvConfig.SyncPatterns...)
		
		// Custom patterns from project replace global ones
		if len(projectCfg.EnvConfig.CustomPatterns) > 0 {
			merged.EnvConfig.CustomPatterns = projectCfg.EnvConfig.CustomPatterns
		}
	}
	
	return merged
}
