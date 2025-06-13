package cmd

import (
	"testing"
)

// TestMultiplexCommand tests the multiplex command structure
func TestMultiplexCommand(t *testing.T) {
	// Test that the command is properly initialized
	if multiplexCmd == nil {
		t.Fatal("multiplex command is nil")
	}

	// Test command properties
	if multiplexCmd.Use != "multiplex [worktrees...]" {
		t.Errorf("unexpected command use: %s", multiplexCmd.Use)
	}

	// Test aliases
	if len(multiplexCmd.Aliases) != 1 || multiplexCmd.Aliases[0] != "mx" {
		t.Errorf("unexpected aliases: %v", multiplexCmd.Aliases)
	}

	// Test that RunE is set
	if multiplexCmd.RunE == nil {
		t.Fatal("multiplex command RunE is nil")
	}
}

// TestMultiplexFlags tests the multiplex command flags
func TestMultiplexFlags(t *testing.T) {
	// Test that flags are properly defined
	flags := multiplexCmd.Flags()

	// Check for expected flags
	expectedFlags := []string{"all", "config", "auto-start", "token-limit"}
	for _, flagName := range expectedFlags {
		if flag := flags.Lookup(flagName); flag == nil {
			t.Errorf("flag %s not found", flagName)
		}
	}

	// Check hidden debug flag
	if flag := flags.Lookup("debug"); flag == nil {
		t.Error("debug flag not found")
	} else if !flag.Hidden {
		t.Error("debug flag should be hidden")
	}
}