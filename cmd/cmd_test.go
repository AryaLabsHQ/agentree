package cmd

import (
	"testing"
	"github.com/spf13/cobra"
)

func TestCommandStructure(t *testing.T) {
	// Test that commands are properly structured
	tests := []struct {
		name        string
		commandName string
		hasFlags    []string
	}{
		{
			name:        "root command exists",
			commandName: "hatch",
			hasFlags:    []string{"branch", "env", "push"},
		},
		{
			name:        "create command exists", 
			commandName: "create",
			hasFlags:    []string{"branch", "from", "push", "env"},
		},
		{
			name:        "remove command exists",
			commandName: "rm",
			hasFlags:    []string{"yes", "delete-branch"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd *cobra.Command
			
			switch tt.commandName {
			case "hatch":
				cmd = rootCmd
			case "create":
				cmd = createCmd
			case "rm":
				cmd = removeCmd
			}
			
			if cmd == nil {
				t.Errorf("Command %s not found", tt.commandName)
				return
			}
			
			// Check flags exist
			for _, flagName := range tt.hasFlags {
				flag := cmd.Flags().Lookup(flagName)
				if flag == nil {
					t.Errorf("Command %s missing flag %s", tt.commandName, flagName)
				}
			}
		})
	}
}

func TestVersionOutput(t *testing.T) {
	// Test that version is set
	if version == "" {
		t.Error("Version should not be empty")
	}
	
	// Test root command has version
	if rootCmd.Version != version {
		t.Errorf("Root command version = %v, want %v", rootCmd.Version, version)
	}
}