package audit

import (
	"context"
	"sync"
	"time"
)

// mockAuditRepository is a mock implementation of AuditRepository for testing.
type mockAuditRepository struct {
	mu          sync.Mutex
	syncWrites  []*AuditEvent
	batchWrites [][]*AuditEvent
}

func (m *mockAuditRepository) WriteBatch(ctx context.Context, events []*AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batchWrites = append(m.batchWrites, events)
	return nil
}

func (m *mockAuditRepository) WriteSync(ctx context.Context, event *AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.syncWrites = append(m.syncWrites, event)
	return nil
}

func (m *mockAuditRepository) Query(ctx context.Context, filter QueryFilter) ([]*AuditEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var results []*AuditEvent

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
