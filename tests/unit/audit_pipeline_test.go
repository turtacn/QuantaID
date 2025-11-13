package unit

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/turtacn/QuantaID/internal/audit"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// mockSink is a simple in-memory sink for testing the pipeline.
type mockSink struct {
	mu     sync.Mutex
	events []*audit.AuditEvent
	err    error // Optional error to simulate sink failure
}

func (s *mockSink) Write(ctx context.Context, event *audit.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.events = append(s.events, event)
	return nil
}

func (s *mockSink) Close() error { return nil }

func (s *mockSink) getEvents() []*audit.AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.events
}

func TestPipeline_EmitFanout(t *testing.T) {
	sink1 := &mockSink{}
	sink2 := &mockSink{}
	sink3 := &mockSink{err: assert.AnError} // This one will fail

	logger, _ := zap.NewDevelopment()
	pipeline := audit.NewPipeline(logger, sink1, sink2, sink3)

	event := &audit.AuditEvent{ID: "test-event"}
	pipeline.Emit(context.Background(), event)

	// Check that the event was written to the successful sinks
	assert.Len(t, sink1.getEvents(), 1)
	assert.Equal(t, "test-event", sink1.getEvents()[0].ID)
	assert.Len(t, sink2.getEvents(), 1)
	assert.Equal(t, "test-event", sink2.getEvents()[0].ID)

	// Check that the failing sink did not receive the event
	assert.Len(t, sink3.getEvents(), 0)
}

func TestFileSink_WriteJSONLine(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "audit.log")

	sink, err := audit.NewFileSink(logPath)
	require.NoError(t, err)
	defer sink.Close()

	event := &audit.AuditEvent{
		ID:        "evt-123",
		Timestamp: time.Now().UTC(),
		Category:  "auth",
		Action:    "login_test",
	}

	err = sink.Write(context.Background(), event)
	require.NoError(t, err)

	// Read the file and verify the content
	data, err := os.ReadFile(logPath)
	require.NoError(t, err)

	var readEvent audit.AuditEvent
	err = json.Unmarshal(data, &readEvent)
	require.NoError(t, err)

	assert.Equal(t, event.ID, readEvent.ID)
	assert.Equal(t, event.Category, readEvent.Category)
	assert.Equal(t, event.Action, readEvent.Action)
}
