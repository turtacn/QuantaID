package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockAuditRepository is a mock implementation of AuditRepository for testing.
type mockAuditRepository struct {
	mu         sync.Mutex
	syncWrites []*AuditEvent
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


func TestAuditLogger_RecordEvent(t *testing.T) {
	repo := &mockAuditRepository{}
	logger := NewAuditLogger(repo, zap.NewNop(), 10, 1*time.Second, 100)
	defer logger.Shutdown()

	event := &AuditEvent{EventType: EventLoginSuccess, Action: "user login"}
	logger.Record(context.Background(), event)

	// Wait for the flush interval to trigger
	time.Sleep(1100 * time.Millisecond)

	require.Len(t, repo.batchWrites, 1)
	require.Len(t, repo.batchWrites[0], 1)

	recordedEvent := repo.batchWrites[0][0]
	assert.NotEmpty(t, recordedEvent.ID)
	assert.False(t, recordedEvent.Timestamp.IsZero())
	assert.Equal(t, EventLoginSuccess, recordedEvent.EventType)
}

func TestAuditLogger_Batching(t *testing.T) {
	repo := &mockAuditRepository{}
	logger := NewAuditLogger(repo, zap.NewNop(), 5, 1*time.Second, 100)
	defer logger.Shutdown()

	// Record 4 events, less than the batch size
	for i := 0; i < 4; i++ {
		logger.Record(context.Background(), &AuditEvent{Action: "test"})
	}

	time.Sleep(200 * time.Millisecond)
	// No batch should have been written yet
	assert.Empty(t, repo.batchWrites)

	// Record one more to hit the batch size
	logger.Record(context.Background(), &AuditEvent{Action: "test"})

	// Give a moment for the batch to be processed
	time.Sleep(200 * time.Millisecond)

	require.Len(t, repo.batchWrites, 1)
	assert.Len(t, repo.batchWrites[0], 5)
}


func TestAuditLogger_FlushInterval(t *testing.T) {
	repo := &mockAuditRepository{}
	logger := NewAuditLogger(repo, zap.NewNop(), 10, 500*time.Millisecond, 100)
	defer logger.Shutdown()

	logger.Record(context.Background(), &AuditEvent{Action: "test"})
	time.Sleep(200 * time.Millisecond)
	assert.Empty(t, repo.batchWrites) // Not flushed yet

	time.Sleep(400 * time.Millisecond) // Pass the flush interval
	require.Len(t, repo.batchWrites, 1)
	assert.Len(t, repo.batchWrites[0], 1)
}


func TestAuditLogger_BufferOverflow(t *testing.T) {
	repo := &mockAuditRepository{}
	// Small buffer size to test overflow
	logger := NewAuditLogger(repo, zap.NewNop(), 10, 1*time.Second, 2)
	defer logger.Shutdown()

	// Fill the buffer
	logger.Record(context.Background(), &AuditEvent{Action: "test1"})
	logger.Record(context.Background(), &AuditEvent{Action: "test2"})

	// This one should overflow and trigger a sync write
	logger.Record(context.Background(), &AuditEvent{Action: "sync_test"})

	time.Sleep(100 * time.Millisecond)

	assert.Len(t, repo.syncWrites, 1)
	assert.Equal(t, "sync_test", repo.syncWrites[0].Action)
}

func TestAuditLogger_Shutdown(t *testing.T) {
	repo := &mockAuditRepository{}
	logger := NewAuditLogger(repo, zap.NewNop(), 10, 5*time.Second, 100)

	logger.Record(context.Background(), &AuditEvent{Action: "test1"})
	logger.Record(context.Background(), &AuditEvent{Action: "test2"})

	// Shutdown before the flush interval
	logger.Shutdown()

	// Check that the remaining events in the buffer were flushed
	require.Len(t, repo.batchWrites, 1)
	assert.Len(t, repo.batchWrites[0], 2)
}
