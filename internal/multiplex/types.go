package multiplex

import "time"

// InstanceState represents the current state of an instance
type InstanceState int

const (
	StateIdle InstanceState = iota
	StateStarting
	StateRunning
	StateThinking
	StateStopping
	StateStopped
	StateCrashed
)

func (s InstanceState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateThinking:
		return "thinking"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateCrashed:
		return "crashed"
	default:
		return "unknown"
	}
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens   int64
	OutputTokens  int64
	TotalTokens   int64
	EstimatedCost float64
	LastUpdated   time.Time
}