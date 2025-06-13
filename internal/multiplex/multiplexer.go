// Package multiplex provides a terminal multiplexer for managing multiple Claude Code instances
package multiplex

import (
	"context"
	"fmt"
	"sync"
)

// UIController interface defines the UI controller contract
type UIController interface {
	Run(ctx context.Context) error
	Stop()
	AddInstance(id, worktree string)
	UpdateInstanceState(id string, state InstanceState)
	AddOutput(id string, data []byte)
	UpdateTokenUsage(id string, usage TokenUsage)
}

// Multiplexer manages multiple Claude Code instances in a TUI
type Multiplexer struct {
	config    *Config
	worktrees []string

	// Core components
	processManager *ProcessManager
	eventDispatcher *EventDispatcher
	uiController   UIController
	
	// State
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	// Channels
	errors chan error
}

// New creates a new multiplexer instance
func New(config *Config, worktrees []string) (*Multiplexer, error) {
	if len(worktrees) == 0 {
		return nil, fmt.Errorf("no worktrees provided")
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Multiplexer{
		config:    config,
		worktrees: worktrees,
		ctx:       ctx,
		cancel:    cancel,
		errors:    make(chan error, 10),
	}

	// Initialize components
	if err := m.initialize(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return m, nil
}

// SetUIController sets the UI controller
func (m *Multiplexer) SetUIController(ui UIController) {
	m.uiController = ui
}

// GetEventChannel returns the event channel for sending events
func (m *Multiplexer) GetEventChannel() chan<- Event {
	return m.eventDispatcher.Events()
}

// Run starts the multiplexer and blocks until exit
func (m *Multiplexer) Run() error {
	if m.uiController == nil {
		return fmt.Errorf("UI controller not set")
	}
	
	// Start all components
	m.start()

	// Wait for completion or error
	select {
	case err := <-m.errors:
		m.shutdown()
		return err
	case <-m.ctx.Done():
		m.shutdown()
		return nil
	}
}

// initialize sets up all components
func (m *Multiplexer) initialize() error {
	// Create event dispatcher
	m.eventDispatcher = NewEventDispatcher(m.ctx)

	// Create process manager
	var err error
	m.processManager, err = NewProcessManager(m.eventDispatcher.Events())
	if err != nil {
		return fmt.Errorf("failed to create process manager: %w", err)
	}

	// Initialize instances for each worktree
	for i, worktree := range m.worktrees {
		instanceID := fmt.Sprintf("instance-%d", i)
		instance := &Instance{
			ID:       instanceID,
			Worktree: worktree,
			State:    StateIdle,
		}
		m.processManager.AddInstance(instance)
	}

	// UI controller will be set externally to avoid import cycle
	// The cmd package will create the UI controller and set it

	// Register event handlers
	m.registerEventHandlers()

	return nil
}

// start begins all component goroutines
func (m *Multiplexer) start() {
	// Start event dispatcher
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.eventDispatcher.Run()
	}()

	// Start process manager
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.processManager.Run(m.ctx); err != nil {
			m.errors <- fmt.Errorf("process manager error: %w", err)
		}
	}()
	
	// Initialize UI with instances
	for _, instance := range m.processManager.GetInstances() {
		m.uiController.AddInstance(instance.ID, instance.Worktree)
	}

	// Start UI controller (this blocks)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.uiController.Run(m.ctx); err != nil {
			m.errors <- fmt.Errorf("UI error: %w", err)
		}
	}()

	// Auto-start instances if configured
	if m.config.AutoStart {
		for _, instance := range m.processManager.GetInstances() {
			m.eventDispatcher.Send(NewProcessControlEvent(instance.ID, ControlStart))
		}
	}
}

// shutdown gracefully stops all components
func (m *Multiplexer) shutdown() {
	// Cancel context to signal shutdown
	m.cancel()

	// Stop UI first (releases terminal)
	if m.uiController != nil {
		m.uiController.Stop()
	}

	// Stop process manager
	if m.processManager != nil {
		m.processManager.StopAll()
	}

	// Wait for all goroutines
	m.wg.Wait()

	// Close channels
	close(m.errors)
}

// registerEventHandlers sets up event routing
func (m *Multiplexer) registerEventHandlers() {
	// UI events -> Process Manager
	m.eventDispatcher.Register(EventUIControl, func(e Event) error {
		if ctrlEvent, ok := e.(*UIControlEvent); ok {
			switch ctrlEvent.Action {
			case "start":
				return m.processManager.StartInstance(ctrlEvent.InstanceID)
			case "stop":
				return m.processManager.StopInstance(ctrlEvent.InstanceID)
			case "restart":
				if err := m.processManager.StopInstance(ctrlEvent.InstanceID); err != nil {
					return err
				}
				return m.processManager.StartInstance(ctrlEvent.InstanceID)
			}
		}
		return nil
	})

	// Process events -> UI updates
	m.eventDispatcher.Register(EventProcessStateChange, func(e Event) error {
		if stateEvent, ok := e.(*ProcessStateEvent); ok {
			m.uiController.UpdateInstanceState(stateEvent.InstanceID, stateEvent.NewState)
		}
		return nil
	})
	
	// Process output -> UI
	m.eventDispatcher.Register(EventProcessOutput, func(e Event) error {
		if outputEvent, ok := e.(*ProcessOutputEvent); ok {
			m.uiController.AddOutput(outputEvent.InstanceID, outputEvent.Data)
		}
		return nil
	})
	
	// Token updates -> UI
	m.eventDispatcher.Register(EventTokenUpdate, func(e Event) error {
		if tokenEvent, ok := e.(*TokenUpdateEvent); ok {
			usage := TokenUsage{
				InputTokens:  tokenEvent.InputTokens,
				OutputTokens: tokenEvent.OutputTokens,
			}
			m.uiController.UpdateTokenUsage(tokenEvent.InstanceID, usage)
		}
		return nil
	})

	// Handle quit event
	m.eventDispatcher.Register(EventQuit, func(e Event) error {
		m.cancel() // Trigger shutdown
		return nil
	})
}