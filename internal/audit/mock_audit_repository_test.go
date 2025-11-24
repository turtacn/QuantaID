package audit

import (
	"context"
	"sync"
	"time"

	"github.com/turtacn/QuantaID/pkg/audit/events"
)

// mockAuditRepository is a mock implementation of AuditRepository for testing.
type mockAuditRepository struct {
	mu          sync.Mutex
	syncWrites  []*events.AuditEvent
	batchWrites [][]*events.AuditEvent
}

func (m *mockAuditRepository) WriteBatch(ctx context.Context, batch []*events.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batchWrites = append(m.batchWrites, batch)
	return nil
}

func (m *mockAuditRepository) WriteSync(ctx context.Context, event *events.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.syncWrites = append(m.syncWrites, event)
	return nil
}

func (m *mockAuditRepository) Query(ctx context.Context, filter QueryFilter) ([]*events.AuditEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var results []*events.AuditEvent

	allWrites := m.syncWrites
	for _, batch := range m.batchWrites {
		allWrites = append(allWrites, batch...)
	}

	for _, event := range allWrites {
		match := true
		if len(filter.EventTypes) > 0 {
			match = false
			for _, et := range filter.EventTypes {
				if event.EventType == et {
					match = true
					break
				}
			}
		}
		if !filter.StartTimestamp.IsZero() && event.Timestamp.Before(filter.StartTimestamp) {
			match = false
		}

		if match {
			results = append(results, event)
		}
	}

	return results, nil
}

func (m *mockAuditRepository) DeleteBefore(ctx context.Context, cutoff time.Time) error {
	return nil
}
