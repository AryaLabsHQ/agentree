package multiplex

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CoordinationType represents different coordination strategies
type CoordinationType int

const (
	// No coordination - instances work independently
	CoordinationNone CoordinationType = iota
	
	// Shared context - instances share context and state
	CoordinationSharedContext
	
	// Task distribution - coordinator assigns tasks to workers
	CoordinationTaskDistribution
	
	// Pipeline - instances work in sequence
	CoordinationPipeline
)

// Coordinator manages multi-instance coordination
type Coordinator struct {
	coordType CoordinationType
	instances map[string]*Instance
	
	// Shared state
	sharedContext *SharedContext
	taskQueue     *TaskQueue
	
	// Communication
	messages      chan *CoordinationMessage
	
	mu sync.RWMutex
}

// SharedContext holds shared state between instances
type SharedContext struct {
	// Current task information
	CurrentTask   string
	TaskProgress  map[string]float64
	
	// Shared knowledge
	Discoveries   []Discovery
	Decisions     []Decision
	
	// Files being worked on
	FileOwnership map[string]string // file -> instance ID
	
	mu sync.RWMutex
}

// Discovery represents something learned by an instance
type Discovery struct {
	InstanceID  string
	Timestamp   time.Time
	Type        string // "bug", "pattern", "dependency", etc.
	Description string
	Details     map[string]interface{}
}

// Decision represents a decision made by an instance
type Decision struct {
	InstanceID  string
	Timestamp   time.Time
	Type        string // "architecture", "approach", etc.
	Decision    string
	Rationale   string
}

// TaskQueue manages task distribution
type TaskQueue struct {
	tasks     []*Task
	assigned  map[string]*Task // instance ID -> task
	completed []*Task
	mu        sync.Mutex
}

// Task represents a unit of work
type Task struct {
	ID          string
	Type        string // "feature", "bug", "refactor", etc.
	Description string
	Priority    int
	AssignedTo  string
	Status      TaskStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	DependsOn   []string // Task IDs
}

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskAssigned
	TaskInProgress
	TaskCompleted
	TaskFailed
)

// CoordinationMessage represents inter-instance communication
type CoordinationMessage struct {
	From      string
	To        string // Empty for broadcast
	Type      string
	Payload   json.RawMessage
	Timestamp time.Time
}

// NewCoordinator creates a new coordinator
func NewCoordinator(coordType CoordinationType) *Coordinator {
	return &Coordinator{
		coordType:     coordType,
		instances:     make(map[string]*Instance),
		sharedContext: NewSharedContext(),
		taskQueue:     NewTaskQueue(),
		messages:      make(chan *CoordinationMessage, 100),
	}
}

// RegisterInstance registers an instance with the coordinator
func (c *Coordinator) RegisterInstance(instance *Instance) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.instances[instance.ID] = instance
}

// SendMessage sends a coordination message
func (c *Coordinator) SendMessage(msg *CoordinationMessage) {
	select {
	case c.messages <- msg:
	default:
		// Message queue full, drop message
		// In production, we'd want better handling
	}
}

// ProcessMessages processes coordination messages
func (c *Coordinator) ProcessMessages() {
	for msg := range c.messages {
		c.handleMessage(msg)
	}
}

// handleMessage processes a single coordination message
func (c *Coordinator) handleMessage(msg *CoordinationMessage) {
	switch msg.Type {
	case "discovery":
		var discovery Discovery
		if err := json.Unmarshal(msg.Payload, &discovery); err == nil {
			c.sharedContext.AddDiscovery(discovery)
		}
		
	case "decision":
		var decision Decision
		if err := json.Unmarshal(msg.Payload, &decision); err == nil {
			c.sharedContext.AddDecision(decision)
		}
		
	case "task_complete":
		var taskID string
		if err := json.Unmarshal(msg.Payload, &taskID); err == nil {
			c.taskQueue.CompleteTask(taskID)
		}
		
	case "request_task":
		if task := c.taskQueue.GetNextTask(); task != nil {
			c.taskQueue.AssignTask(task.ID, msg.From)
			// Send task assignment back
			payload, _ := json.Marshal(task)
			c.SendMessage(&CoordinationMessage{
				From:    "coordinator",
				To:      msg.From,
				Type:    "task_assignment",
				Payload: payload,
			})
		}
	}
}

// GetSharedContext returns the shared context
func (c *Coordinator) GetSharedContext() *SharedContext {
	return c.sharedContext
}

// GetTaskQueue returns the task queue
func (c *Coordinator) GetTaskQueue() *TaskQueue {
	return c.taskQueue
}

// NewSharedContext creates a new shared context
func NewSharedContext() *SharedContext {
	return &SharedContext{
		TaskProgress:  make(map[string]float64),
		Discoveries:   make([]Discovery, 0),
		Decisions:     make([]Decision, 0),
		FileOwnership: make(map[string]string),
	}
}

// AddDiscovery adds a discovery to the shared context
func (sc *SharedContext) AddDiscovery(discovery Discovery) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.Discoveries = append(sc.Discoveries, discovery)
}

// AddDecision adds a decision to the shared context
func (sc *SharedContext) AddDecision(decision Decision) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	sc.Decisions = append(sc.Decisions, decision)
}

// ClaimFile claims ownership of a file
func (sc *SharedContext) ClaimFile(filename, instanceID string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if owner, exists := sc.FileOwnership[filename]; exists && owner != instanceID {
		return false // Already owned by another instance
	}
	
	sc.FileOwnership[filename] = instanceID
	return true
}

// ReleaseFile releases ownership of a file
func (sc *SharedContext) ReleaseFile(filename, instanceID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if sc.FileOwnership[filename] == instanceID {
		delete(sc.FileOwnership, filename)
	}
}

// GetFileOwner returns the owner of a file
func (sc *SharedContext) GetFileOwner(filename string) string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	return sc.FileOwnership[filename]
}

// NewTaskQueue creates a new task queue
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks:     make([]*Task, 0),
		assigned:  make(map[string]*Task),
		completed: make([]*Task, 0),
	}
}

// AddTask adds a task to the queue
func (tq *TaskQueue) AddTask(task *Task) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	task.Status = TaskPending
	task.CreatedAt = time.Now()
	tq.tasks = append(tq.tasks, task)
}

// GetNextTask returns the next available task
func (tq *TaskQueue) GetNextTask() *Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	for _, task := range tq.tasks {
		if task.Status == TaskPending {
			// Check dependencies
			if tq.dependenciesMet(task) {
				return task
			}
		}
	}
	
	return nil
}

// AssignTask assigns a task to an instance
func (tq *TaskQueue) AssignTask(taskID, instanceID string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	for _, task := range tq.tasks {
		if task.ID == taskID {
			task.AssignedTo = instanceID
			task.Status = TaskAssigned
			now := time.Now()
			task.StartedAt = &now
			tq.assigned[instanceID] = task
			break
		}
	}
}

// CompleteTask marks a task as completed
func (tq *TaskQueue) CompleteTask(taskID string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	for i, task := range tq.tasks {
		if task.ID == taskID {
			task.Status = TaskCompleted
			now := time.Now()
			task.CompletedAt = &now
			
			// Move to completed
			tq.completed = append(tq.completed, task)
			tq.tasks = append(tq.tasks[:i], tq.tasks[i+1:]...)
			
			// Remove from assigned
			if task.AssignedTo != "" {
				delete(tq.assigned, task.AssignedTo)
			}
			break
		}
	}
}

// dependenciesMet checks if all dependencies of a task are completed
func (tq *TaskQueue) dependenciesMet(task *Task) bool {
	for _, depID := range task.DependsOn {
		found := false
		for _, completed := range tq.completed {
			if completed.ID == depID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GetTaskStatus returns the status of all tasks
func (tq *TaskQueue) GetTaskStatus() map[string]interface{} {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	return map[string]interface{}{
		"pending":   len(tq.tasks),
		"assigned":  len(tq.assigned),
		"completed": len(tq.completed),
	}
}

// InstanceRole defines the role of an instance in coordination
type InstanceRole int

const (
	RoleWorker InstanceRole = iota
	RoleCoordinator
	RoleReviewer
	RoleSpecialist
)

// AssignRoles assigns roles to instances based on configuration
func (c *Coordinator) AssignRoles(config map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for instanceID, roleStr := range config {
		if instance, exists := c.instances[instanceID]; exists {
			// Parse role from string
			switch roleStr {
			case "coordinator":
				// Set as coordinator
			case "reviewer":
				// Set as reviewer
			case "specialist":
				// Set as specialist
			default:
				// Default to worker
			}
			_ = instance // Use instance
		}
	}
}

// BroadcastDiscovery broadcasts a discovery to all instances
func (c *Coordinator) BroadcastDiscovery(discovery Discovery) {
	payload, err := json.Marshal(discovery)
	if err != nil {
		return
	}
	
	c.SendMessage(&CoordinationMessage{
		From:      discovery.InstanceID,
		To:        "", // Broadcast
		Type:      "discovery",
		Payload:   payload,
		Timestamp: time.Now(),
	})
}

// RequestTask requests a new task for an instance
func (c *Coordinator) RequestTask(instanceID string) {
	c.SendMessage(&CoordinationMessage{
		From:      instanceID,
		To:        "coordinator",
		Type:      "request_task",
		Timestamp: time.Now(),
	})
}

// GetCoordinationSummary returns a summary of coordination state
func (c *Coordinator) GetCoordinationSummary() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	sc := c.sharedContext
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	return map[string]interface{}{
		"type":          c.coordType,
		"instances":     len(c.instances),
		"discoveries":   len(sc.Discoveries),
		"decisions":     len(sc.Decisions),
		"file_locks":    len(sc.FileOwnership),
		"task_status":   c.taskQueue.GetTaskStatus(),
		"message_queue": len(c.messages),
	}
}