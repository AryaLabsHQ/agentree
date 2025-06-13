package multiplex_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AryaLabsHQ/agentree/internal/multiplex"
)

// TestMultiplexerCreation tests creating a new multiplexer
func TestMultiplexerCreation(t *testing.T) {
	config := multiplex.DefaultConfig()
	worktrees := []string{"/tmp/test-worktree"}

	m, err := multiplex.New(config, worktrees)
	if err != nil {
		t.Fatalf("Failed to create multiplexer: %v", err)
	}

	if m == nil {
		t.Fatal("Multiplexer is nil")
	}
}

// TestProcessManager tests process management functionality
func TestProcessManager(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events := make(chan multiplex.Event, 100)
	pm, err := multiplex.NewProcessManager(events)
	if err != nil {
		t.Fatalf("Failed to create process manager: %v", err)
	}

	// Add a test instance
	instance := &multiplex.Instance{
		ID:       "test-instance",
		Worktree: "/tmp/test-worktree",
		State:    multiplex.StateIdle,
	}
	pm.AddInstance(instance)

	// Check instance was added
	instances := pm.GetInstances()
	if len(instances) != 1 {
		t.Errorf("Expected 1 instance, got %d", len(instances))
	}

	// Test getting specific instance
	got, err := pm.GetInstance("test-instance")
	if err != nil {
		t.Errorf("Failed to get instance: %v", err)
	}
	if got.ID != "test-instance" {
		t.Errorf("Wrong instance returned: %s", got.ID)
	}

	// Run process manager
	go pm.Run(ctx)

	// Wait for context to cancel
	<-ctx.Done()
}

// TestEventDispatcher tests event routing
func TestEventDispatcher(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ed := multiplex.NewEventDispatcher(ctx)

	// Register handler
	received := make(chan multiplex.Event, 1)
	ed.Register(multiplex.EventProcessStateChange, func(e multiplex.Event) error {
		received <- e
		return nil
	})

	// Start dispatcher
	go ed.Run()
	
	// Give dispatcher time to start
	time.Sleep(10 * time.Millisecond)

	// Send event
	event := multiplex.NewProcessStateEvent("test", multiplex.StateIdle, multiplex.StateStarting)
	ed.Send(event)

	// Wait for event
	select {
	case e := <-received:
		if e.Type() != multiplex.EventProcessStateChange {
			t.Errorf("Wrong event type: %v", e.Type())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event not received")
	}
}

// TestTokenTracker tests token parsing and tracking
func TestTokenTracker(t *testing.T) {
	tt := multiplex.NewTokenTracker()

	// Test parsing various formats
	tests := []struct {
		name   string
		output []byte
		wantInput int64
		wantOutput int64
	}{
		{
			name: "Claude format",
			output: []byte("Tokens: Input: 100, Output: 200"),
			wantInput: 100,
			wantOutput: 200,
		},
		{
			name: "JSON format",
			output: []byte(`{"usage": {"input_tokens": 50, "output_tokens": 150}}`),
			wantInput: 50,
			wantOutput: 150,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tt.Reset()
			tt.ParseOutput(test.output)
			usage := tt.GetUsage()

			if usage.InputTokens != test.wantInput {
				t.Errorf("Input tokens: got %d, want %d", usage.InputTokens, test.wantInput)
			}
			if usage.OutputTokens != test.wantOutput {
				t.Errorf("Output tokens: got %d, want %d", usage.OutputTokens, test.wantOutput)
			}
		})
	}
}

// TestMockClaudeIntegration tests with the mock Claude CLI
func TestMockClaudeIntegration(t *testing.T) {
	// Skip if mock-claude not available
	mockPath := "./mock-claude"
	if _, err := os.Stat(mockPath); os.IsNotExist(err) {
		t.Skip("mock-claude binary not found")
	}

	// Make mock-claude absolute path
	absPath, err := filepath.Abs(mockPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	events := make(chan multiplex.Event, 100)
	pm, err := multiplex.NewProcessManager(events)
	if err != nil {
		t.Fatalf("Failed to create process manager: %v", err)
	}

	// Create test instance with mock Claude
	instance := &multiplex.Instance{
		ID:       "test-mock",
		Worktree: ".", // Current directory
		State:    multiplex.StateIdle,
	}
	pm.AddInstance(instance)

	// Override command to use mock-claude
	t.Setenv("MOCK_CLAUDE_PATH", absPath)

	// Start process manager
	go pm.Run(ctx)

	// Collect events
	var stateChanges []multiplex.InstanceState
	go func() {
		for {
			select {
			case event := <-events:
				if stateEvent, ok := event.(*multiplex.ProcessStateEvent); ok {
					stateChanges = append(stateChanges, stateEvent.NewState)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start instance
	err = pm.StartInstance("test-mock")
	if err != nil {
		t.Errorf("Failed to start instance: %v", err)
	}

	// Give it time to start
	time.Sleep(500 * time.Millisecond)

	// Send input
	err = pm.SendInput("test-mock", "hello\n")
	if err != nil {
		t.Errorf("Failed to send input: %v", err)
	}

	// Wait for response
	time.Sleep(2 * time.Second)

	// Stop instance
	err = pm.StopInstance("test-mock")
	if err != nil {
		t.Errorf("Failed to stop instance: %v", err)
	}

	// Check state transitions
	if len(stateChanges) < 2 {
		t.Errorf("Expected at least 2 state changes, got %d", len(stateChanges))
	}
}

// TestConfiguration tests configuration loading and validation
func TestConfiguration(t *testing.T) {
	// Test default config
	config := multiplex.DefaultConfig()
	if err := config.Validate(); err != nil {
		t.Errorf("Default config validation failed: %v", err)
	}

	// Test invalid configs
	invalidConfigs := []struct {
		name   string
		modify func(*multiplex.Config)
	}{
		{
			name: "negative token limit",
			modify: func(c *multiplex.Config) {
				c.TokenLimit = -1
			},
		},
		{
			name: "invalid theme",
			modify: func(c *multiplex.Config) {
				c.Theme = "invalid"
			},
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			config := multiplex.DefaultConfig()
			tc.modify(config)
			if err := config.Validate(); err == nil {
				t.Error("Expected validation error")
			}
		})
	}
}