package memory

import (
	"context"
	"sync"

	"github.com/turtacn/QuantaID/pkg/audit/events"
)

// InMemorySink provides an in-memory implementation of the audit sink.
type InMemorySink struct {
	mu     sync.RWMutex
	events []*events.AuditEvent
}

// NewInMemorySink creates a new in-memory audit sink.
func NewInMemorySink() *InMemorySink {
	return &InMemorySink{
		events: make([]*events.AuditEvent, 0),
	}
}

// Write appends the event to the in-memory slice.
func (s *InMemorySink) Write(ctx context.Context, event *events.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

// Close is a no-op for InMemorySink.
func (s *InMemorySink) Close() error {
	return nil
}

// GetEvents returns all the events that have been written to the sink.
func (s *InMemorySink) GetEvents() []*events.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events
}
