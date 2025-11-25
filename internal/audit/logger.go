package audit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/internal/audit/sinks"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"go.uber.org/zap"
)

// AuditLogger provides a high-performance, asynchronous pipeline for recording audit events.
type AuditLogger struct {
	sinks        []sinks.Sink
	buffer       chan *events.AuditEvent
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	logger       *zap.Logger
	batchSize    int
	flushInterval time.Duration
	// Keep a primary repo if needed for non-sink operations, but effectively we only use sinks now.
	// We expose GetRepo for legacy support if needed, but it's better to pass repo directly where needed.
	repo AuditRepository
}

// GetRepo returns the primary audit repository if available.
func (al *AuditLogger) GetRepo() AuditRepository {
	return al.repo
}

// NewAuditLogger creates and starts a new AuditLogger.
func NewAuditLogger(logger *zap.Logger, batchSize int, flushInterval time.Duration, bufferSize int, sinks ...sinks.Sink) *AuditLogger {
	ctx, cancel := context.WithCancel(context.Background())

	var repo AuditRepository
	// Try to find a sink that implements AuditRepository to use as the primary repo
	for _, s := range sinks {
		if r, ok := s.(AuditRepository); ok {
			repo = r
			break
		}
	}

	al := &AuditLogger{
		sinks:         sinks,
		buffer:        make(chan *events.AuditEvent, bufferSize),
		cancel:        cancel,
		logger:        logger.Named("audit-logger"),
		batchSize:     batchSize,
		flushInterval: flushInterval,
		repo:          repo,
	}

	al.wg.Add(1)
	go al.flushLoop(ctx)

	return al
}

// Record submits an audit event to the asynchronous logging buffer.
func (al *AuditLogger) Record(ctx context.Context, event *events.AuditEvent) {
	// Enrich event with server-generated data
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	select {
	case al.buffer <- event:
		// Event successfully buffered
	default:
		// Buffer is full, escalate to synchronous write as a fallback.
		al.logger.Warn("Audit buffer is full. Falling back to synchronous write.", zap.String("event_id", event.ID))
		al.writeSync(context.Background(), event)
	}
}

// writeSync writes to all sinks synchronously
func (al *AuditLogger) writeSync(ctx context.Context, event *events.AuditEvent) {
	for _, sink := range al.sinks {
		if err := sink.WriteSync(ctx, event); err != nil {
			al.logger.Error("Failed to perform synchronous audit write.", zap.Error(err), zap.String("event_id", event.ID))
		}
	}
}

// flushLoop is the core worker goroutine that batches and writes events.
func (al *AuditLogger) flushLoop(ctx context.Context) {
	defer al.wg.Done()

	batch := make([]*events.AuditEvent, 0, al.batchSize)
	ticker := time.NewTicker(al.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-al.buffer:
			batch = append(batch, event)
			if len(batch) >= al.batchSize {
				al.flushBatch(context.Background(), batch)
				batch = make([]*events.AuditEvent, 0, al.batchSize) // Reset batch
			}
		case <-ticker.C:
			if len(batch) > 0 {
				al.flushBatch(context.Background(), batch)
				batch = make([]*events.AuditEvent, 0, al.batchSize) // Reset batch
			}
		case <-ctx.Done():
			// Drain the buffer on shutdown
			close(al.buffer)
			for event := range al.buffer {
				batch = append(batch, event)
			}
			if len(batch) > 0 {
				al.flushBatch(context.Background(), batch)
			}
			// Close sinks
			for _, sink := range al.sinks {
				_ = sink.Close()
			}
			return
		}
	}
}

// flushBatch writes a batch of events to the repository.
func (al *AuditLogger) flushBatch(ctx context.Context, batch []*events.AuditEvent) {
	for _, sink := range al.sinks {
		if err := sink.WriteBatch(ctx, batch); err != nil {
			al.logger.Error("Failed to flush audit event batch to sink.", zap.Error(err), zap.Int("batch_size", len(batch)))
		}
	}
}

// Shutdown gracefully stops the logger, ensuring all buffered events are flushed.
func (al *AuditLogger) Shutdown() {
	al.cancel()
	al.wg.Wait()
}
