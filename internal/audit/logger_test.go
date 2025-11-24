package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"go.uber.org/zap"
)

type mockSink struct {
	mu     sync.Mutex
	events []*events.AuditEvent
	closed bool
}

func (m *mockSink) WriteBatch(ctx context.Context, events []*events.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulate slow write
	time.Sleep(10 * time.Millisecond)

	m.events = append(m.events, events...)
	return nil
}

func (m *mockSink) WriteSync(ctx context.Context, event *events.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockSink) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func TestAsyncAudit(t *testing.T) {
	mock := &mockSink{}
	logger := NewAuditLogger(zap.NewNop(), 10, 100*time.Millisecond, 100, mock)

	// Send events
	for i := 0; i < 50; i++ {
		logger.Record(context.Background(), &events.AuditEvent{
			EventType: events.EventLoginSuccess,
			Action:    "test",
		})
	}

	// Should return immediately (async) - verified by test execution speed

	// Wait for flush
	time.Sleep(200 * time.Millisecond)

	logger.Shutdown()

	mock.mu.Lock()
	defer mock.mu.Unlock()

	assert.Len(t, mock.events, 50)
	assert.True(t, mock.closed)
}

func TestAsyncNonBlocking(t *testing.T) {
	// A mock sink that blocks for a while
	slowSink := &mockSink{}
	// Make the sink very slow? Actually WriteBatch already sleeps 10ms.
	// If we send 100 events, serial processing would take 1s.
	// Async logger buffers them and returns immediately.

	logger := NewAuditLogger(zap.NewNop(), 10, 1*time.Second, 1000, slowSink)

	start := time.Now()
	for i := 0; i < 100; i++ {
		logger.Record(context.Background(), &events.AuditEvent{
			EventType: events.EventLoginSuccess,
		})
	}
	duration := time.Since(start)

	// Recording 100 events should be instantaneous, much less than 100 * 10ms = 1s
	assert.Less(t, duration, 100*time.Millisecond)

	logger.Shutdown()
}
