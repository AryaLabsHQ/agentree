package multiplex

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds multiplexer configuration
type Config struct {
	// General settings
	AutoStart    bool              `yaml:"auto_start"`
	TokenLimit   int64             `yaml:"token_limit"`
	Theme        string            `yaml:"theme"`
	LogLevel     string            `yaml:"log_level"`
	
	// Instance configurations
	Instances    []InstanceConfig  `yaml:"instances"`
	
	// UI settings
	UI           UIConfig          `yaml:"ui"`
	
	// Shortcuts
	Shortcuts    map[string]string `yaml:"shortcuts"`
	
	// Coordination settings
	Coordination CoordinationConfig `yaml:"coordination"`
}

// InstanceConfig configures a specific instance
type InstanceConfig struct {
	Worktree    string   `yaml:"worktree"`
	AutoStart   bool     `yaml:"auto_start"`
	TokenLimit  int64    `yaml:"token_limit"`
	Environment []string `yaml:"environment"`
	Role        string   `yaml:"role"` // "coordinator", "developer", etc.
}

// UIConfig holds UI-specific settings
type UIConfig struct {
	SidebarWidth   int    `yaml:"sidebar_width"`
	ShowTokenUsage bool   `yaml:"show_token_usage"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	ScrollbackSize int    `yaml:"scrollback_size"`
	Theme          string `yaml:"theme"`
}

// CoordinationConfig holds multi-instance coordination settings
type CoordinationConfig struct {
	SharedContext bool   `yaml:"shared_context"`
	SyncInterval  string `yaml:"sync_interval"`
	EnableIPC     bool   `yaml:"enable_ipc"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		AutoStart:  false,
		TokenLimit: 100000,
		Theme:      "dark",
		LogLevel:   "info",
		
		UI: UIConfig{
			SidebarWidth:   25,
			ShowTokenUsage: true,
			ShowTimestamps: true,
			ScrollbackSize: 10000,
			Theme:          "dark",
		},
		
		Shortcuts: map[string]string{
			"quit":         "q",
			"start":        "s",
			"stop":         "x",
			"restart":      "r",
			"next":         "j",
			"prev":         "k",
			"focus":        "Enter",
			"scroll_up":    "Up",
			"scroll_down":  "Down",
			"page_up":      "PgUp",
			"page_down":    "PgDn",
			"home":         "Home",
			"end":          "End",
			"clear":        "c",
			"help":         "?",
		},
		
		Coordination: CoordinationConfig{
			SharedContext: true,
			SyncInterval:  "30s",
			EnableIPC:     false,
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	// Validate
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return config, nil
}

// LoadConfigFromDefault tries to load from default locations
func LoadConfigFromDefault() (*Config, error) {
	// Try project-specific config first
	projectConfig := ".agentree-multiplex.yml"
	if _, err := os.Stat(projectConfig); err == nil {
		return LoadConfig(projectConfig)
	}
	
	// Try user config
	home, err := os.UserHomeDir()
	if err == nil {
		userConfig := filepath.Join(home, ".config", "agentree", "multiplex.yml")
		if _, err := os.Stat(userConfig); err == nil {
			return LoadConfig(userConfig)
		}
	}
	
	// Return defaults
	return DefaultConfig(), nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.TokenLimit <= 0 {
		return fmt.Errorf("token_limit must be positive")
	}
	
	if c.UI.SidebarWidth < 10 || c.UI.SidebarWidth > 50 {
		return fmt.Errorf("sidebar_width must be between 10 and 50")
	}
	
	if c.UI.ScrollbackSize < 100 {
		return fmt.Errorf("scrollback_size must be at least 100")
	}
	
	// Validate theme
	validThemes := map[string]bool{"dark": true, "light": true, "nord": true}
	if !validThemes[c.Theme] {
		return fmt.Errorf("invalid theme: %s", c.Theme)
	}
	
	return nil
}

// GetInstanceConfig returns config for a specific instance
func (c *Config) GetInstanceConfig(worktree string) *InstanceConfig {
	for _, ic := range c.Instances {
		if ic.Worktree == worktree {
			return &ic
		}
	}
	
	// Return default instance config
	return &InstanceConfig{
		Worktree:   worktree,
		AutoStart:  c.AutoStart,
		TokenLimit: c.TokenLimit / int64(len(c.Instances)+1), // Divide limit
	}
}