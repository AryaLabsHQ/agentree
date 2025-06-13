package multiplex

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/AryaLabsHQ/agentree/internal/multiplex/terminal"
	"github.com/creack/pty"
)


// Instance represents a single Claude Code process
type Instance struct {
	ID       string
	Worktree string
	Cmd      *exec.Cmd
	PTY      *os.File
	VT       *terminal.VirtualTerminal
	State    InstanceState
	
	// Token tracking
	TokenUsage  *TokenTracker
	
	// Buffers and channels
	outputChan chan []byte
	errorChan  chan error
	done       chan struct{}
	
	// Timestamps
	StartedAt   time.Time
	LastActive  time.Time
	
	// Mutex for thread-safe access
	mu sync.RWMutex
}

// ProcessManager manages multiple Claude Code instances
type ProcessManager struct {
	instances map[string]*Instance
	events    chan<- Event
	mu        sync.RWMutex
}

// NewProcessManager creates a new process manager
func NewProcessManager(events chan<- Event) (*ProcessManager, error) {
	return &ProcessManager{
		instances: make(map[string]*Instance),
		events:    events,
	}, nil
}

// Run starts the process manager main loop
func (pm *ProcessManager) Run(ctx context.Context) error {
	// Monitor instances
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			pm.checkInstances()
		}
	}
}

// AddInstance adds a new instance to manage
func (pm *ProcessManager) AddInstance(instance *Instance) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Initialize instance fields
	instance.outputChan = make(chan []byte, 100)
	instance.errorChan = make(chan error, 10)
	instance.done = make(chan struct{})
	instance.TokenUsage = NewTokenTracker()
	
	pm.instances[instance.ID] = instance
	
	// Send event
	pm.events <- NewProcessStateEvent(instance.ID, StateIdle, instance.State)
}

// GetInstances returns all instances
func (pm *ProcessManager) GetInstances() []*Instance {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	instances := make([]*Instance, 0, len(pm.instances))
	for _, instance := range pm.instances {
		instances = append(instances, instance)
	}
	return instances
}

// StartInstance starts a Claude Code instance
func (pm *ProcessManager) StartInstance(id string) error {
	pm.mu.Lock()
	instance, exists := pm.instances[id]
	pm.mu.Unlock()
	
	if !exists {
		return fmt.Errorf("instance %s not found", id)
	}
	
	instance.mu.Lock()
	defer instance.mu.Unlock()
	
	if instance.State == StateRunning {
		return fmt.Errorf("instance already running")
	}
	
	// Update state
	oldState := instance.State
	instance.State = StateStarting
	instance.StartedAt = time.Now()
	
	// Send state change event
	pm.events <- NewProcessStateEvent(id, oldState, StateStarting)
	
	// Create command
	cmd := exec.Command("claude", "code")
	cmd.Dir = instance.Worktree
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AGENTREE_INSTANCE_ID=%s", id),
		fmt.Sprintf("AGENTREE_WORKTREE=%s", instance.Worktree),
	)
	
	// Create PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		instance.State = StateCrashed
		pm.events <- &ProcessErrorEvent{
			BaseEvent:  BaseEvent{EventType: EventProcessError, Time: time.Now()},
			InstanceID: id,
			Error:      err,
		}
		return fmt.Errorf("failed to start PTY: %w", err)
	}
	
	instance.Cmd = cmd
	instance.PTY = ptmx
	instance.State = StateRunning
	
	// Create virtual terminal
	instance.VT = terminal.NewVirtualTerminal(80, 24) // Default size, will be updated
	
	// Start output reader
	go pm.readOutput(instance)
	
	// Send running event
	pm.events <- NewProcessStateEvent(id, StateStarting, StateRunning)
	
	return nil
}

// StopInstance stops a Claude Code instance
func (pm *ProcessManager) StopInstance(id string) error {
	pm.mu.RLock()
	instance, exists := pm.instances[id]
	pm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("instance %s not found", id)
	}
	
	instance.mu.Lock()
	defer instance.mu.Unlock()
	
	if instance.State != StateRunning {
		return fmt.Errorf("instance not running")
	}
	
	// Update state
	instance.State = StateStopping
	
	// Close PTY and kill process
	if instance.PTY != nil {
		instance.PTY.Close()
	}
	
	if instance.Cmd != nil && instance.Cmd.Process != nil {
		instance.Cmd.Process.Kill()
		instance.Cmd.Wait()
	}
	
	instance.State = StateStopped
	close(instance.done)
	
	// Send event
	pm.events <- NewProcessStateEvent(id, StateRunning, StateStopped)
	
	return nil
}

// StopAll stops all running instances
func (pm *ProcessManager) StopAll() {
	pm.mu.RLock()
	ids := make([]string, 0, len(pm.instances))
	for id := range pm.instances {
		ids = append(ids, id)
	}
	pm.mu.RUnlock()
	
	for _, id := range ids {
		pm.StopInstance(id)
	}
}

// readOutput reads output from an instance
func (pm *ProcessManager) readOutput(instance *Instance) {
	scanner := bufio.NewScanner(instance.PTY)
	
	for scanner.Scan() {
		data := scanner.Bytes()
		
		// Feed to virtual terminal
		instance.VT.Write(data)
		
		// Parse for token usage
		// TODO: Implement token parsing
		
		// Send output event
		pm.events <- NewProcessOutputEvent(instance.ID, data)
		
		// Update last active
		instance.mu.Lock()
		instance.LastActive = time.Now()
		instance.mu.Unlock()
	}
	
	// Handle completion
	instance.mu.Lock()
	if instance.State == StateRunning {
		instance.State = StateCrashed
		pm.events <- NewProcessStateEvent(instance.ID, StateRunning, StateCrashed)
	}
	instance.mu.Unlock()
}

// checkInstances monitors instance health
func (pm *ProcessManager) checkInstances() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	for _, instance := range pm.instances {
		instance.mu.RLock()
		
		// Check for hung instances
		if instance.State == StateRunning {
			if time.Since(instance.LastActive) > 5*time.Minute {
				// TODO: Send warning event
			}
		}
		
		instance.mu.RUnlock()
	}
}