// Package config handles hatch configuration from .hatchrc and global config
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds hatch configuration
type Config struct {
	PostCreateScripts []string
	// Package manager specific overrides
	PnpmSetup    string
	NpmSetup     string
	YarnSetup    string
	DefaultSetup string
}

// LoadProjectConfig loads configuration from .hatchrc in the project root
func LoadProjectConfig(projectRoot string) (*Config, error) {
	cfg := &Config{}
	hatchrcPath := filepath.Join(projectRoot, ".hatchrc")
	
	file, err := os.Open(hatchrcPath)
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
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Look for POST_CREATE_SCRIPTS array
		if strings.Contains(line, "POST_CREATE_SCRIPTS=(") {
			inPostCreateScripts = true
			continue
		}
		
		if inPostCreateScripts {
			if strings.Contains(line, ")") {
				inPostCreateScripts = false
				continue
			}
			// Extract script from quotes
			script := strings.Trim(line, ` "',`)
			if script != "" {
				cfg.PostCreateScripts = append(cfg.PostCreateScripts, script)
			}
		}
	}
	
	return cfg, scanner.Err()
}

// LoadGlobalConfig loads configuration from ~/.config/hatch/config
func LoadGlobalConfig() (*Config, error) {
	cfg := &Config{}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil // Ignore if can't get home dir
	}
	
	configPath := filepath.Join(homeDir, ".config", "hatch", "config")
	
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
		}
	}
	
	return cfg, scanner.Err()
}