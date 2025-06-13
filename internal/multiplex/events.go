package multiplex

import (
	"context"
	"log"
	"sync"
	"time"
)

// EventType represents different types of events in the system
type EventType int

const (
	// Process events
	EventProcessStarted EventType = iota
	EventProcessStopped
	EventProcessOutput
	EventProcessError
	EventProcessStateChange
	
	// UI events
	EventUIResize
	EventUIFocus
	EventUIInput
	EventUIControl
	
	// System events
	EventTokenUpdate
	EventConfigChange
	EventQuit
)

// Event is the base interface for all events
type Event interface {
	Type() EventType
	Timestamp() time.Time
}

// BaseEvent provides common event fields
type BaseEvent struct {
	EventType EventType
	Time      time.Time
}

func (e BaseEvent) Type() EventType     { return e.EventType }
func (e BaseEvent) Timestamp() time.Time { return e.Time }

// Process Events

// ProcessStateEvent indicates a state change
type ProcessStateEvent struct {
	BaseEvent
	InstanceID string
	OldState   InstanceState
	NewState   InstanceState
}

// ProcessOutputEvent contains output data
type ProcessOutputEvent struct {
	BaseEvent
	InstanceID string
	Data       []byte
}

// ProcessErrorEvent indicates an error
type ProcessErrorEvent struct {
	BaseEvent
	InstanceID string
	Error      error
}

// UI Events

// UIControlEvent represents user control actions
type UIControlEvent struct {
	BaseEvent
	InstanceID string
	Action     string // "start", "stop", "restart", etc.
}

// UIResizeEvent indicates terminal resize
type UIResizeEvent struct {
	BaseEvent
	Width  int
	Height int
}

// UIFocusEvent indicates focus change
type UIFocusEvent struct {
	BaseEvent
	InstanceID string
}

// System Events

// TokenUpdateEvent contains token usage update
type TokenUpdateEvent struct {
	BaseEvent
	InstanceID   string
	InputTokens  int64
	OutputTokens int64
}

// Control Events

type ProcessControl int

const (
	ControlStart ProcessControl = iota
	ControlStop
	ControlRestart
)

// ProcessControlEvent requests process control action
type ProcessControlEvent struct {
	BaseEvent
	InstanceID string
	Control    ProcessControl
}

// EventHandler processes events
type EventHandler func(Event) error

// EventDispatcher manages event routing
type EventDispatcher struct {
	events   chan Event
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
	ctx      context.Context
	wg       sync.WaitGroup
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(ctx context.Context) *EventDispatcher {
	return &EventDispatcher{
		events:   make(chan Event, 100),
		handlers: make(map[EventType][]EventHandler),
		ctx:      ctx,
	}
}

// Events returns the event channel for sending
func (ed *EventDispatcher) Events() chan<- Event {
	return ed.events
}

// Run starts the event dispatcher
func (ed *EventDispatcher) Run() {
	ed.wg.Add(1)
	defer ed.wg.Done()
	
	for {
		select {
		case <-ed.ctx.Done():
			return
		case event := <-ed.events:
			ed.dispatch(event)
		}
	}
}

// Send sends an event
func (ed *EventDispatcher) Send(event Event) {
	select {
	case ed.events <- event:
	case <-ed.ctx.Done():
	}
}

// Register adds a handler for an event type
func (ed *EventDispatcher) Register(eventType EventType, handler EventHandler) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	
	ed.handlers[eventType] = append(ed.handlers[eventType], handler)
}

// dispatch routes an event to its handlers
func (ed *EventDispatcher) dispatch(event Event) {
	ed.mu.RLock()
	handlers := ed.handlers[event.Type()]
	ed.mu.RUnlock()
	
	for _, handler := range handlers {
		// Run handlers in goroutines to prevent blocking
		go func(h EventHandler) {
			if err := h(event); err != nil {
				log.Printf("Event handler error for %v: %v", event.Type(), err)
			}
		}(handler)
	}
}

// Wait waits for all events to be processed
func (ed *EventDispatcher) Wait() {
	ed.wg.Wait()
}

// Helper constructors

func NewProcessStateEvent(instanceID string, oldState, newState InstanceState) *ProcessStateEvent {
	return &ProcessStateEvent{
		BaseEvent:  BaseEvent{EventType: EventProcessStateChange, Time: time.Now()},
		InstanceID: instanceID,
		OldState:   oldState,
		NewState:   newState,
	}
}

func NewProcessOutputEvent(instanceID string, data []byte) *ProcessOutputEvent {
	return &ProcessOutputEvent{
		BaseEvent:  BaseEvent{EventType: EventProcessOutput, Time: time.Now()},
		InstanceID: instanceID,
		Data:       data,
	}
}

func NewProcessControlEvent(instanceID string, control ProcessControl) *ProcessControlEvent {
	return &ProcessControlEvent{
		BaseEvent:  BaseEvent{EventType: EventUIControl, Time: time.Now()},
		InstanceID: instanceID,
		Control:    control,
	}
}

func NewUIResizeEvent(width, height int) *UIResizeEvent {
	return &UIResizeEvent{
		BaseEvent: BaseEvent{EventType: EventUIResize, Time: time.Now()},
		Width:     width,
		Height:    height,
	}
}

func NewQuitEvent() Event {
	return BaseEvent{EventType: EventQuit, Time: time.Now()}
}