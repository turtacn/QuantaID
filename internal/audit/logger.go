package audit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditLogger provides a high-performance, asynchronous pipeline for recording audit events.
type AuditLogger struct {
	repo         AuditRepository
	buffer       chan *AuditEvent
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	logger       *zap.Logger
	batchSize    int
	flushInterval time.Duration
}

// NewAuditLogger creates and starts a new AuditLogger.
func NewAuditLogger(repo AuditRepository, logger *zap.Logger, batchSize int, flushInterval time.Duration, bufferSize int) *AuditLogger {
	ctx, cancel := context.WithCancel(context.Background())

	al := &AuditLogger{
		repo:         repo,
		buffer:       make(chan *AuditEvent, bufferSize),
		cancel:       cancel,
		logger:       logger.Named("audit-logger"),
		batchSize:    batchSize,
		flushInterval: flushInterval,
	}

	al.wg.Add(1)
	go al.flushLoop(ctx)

	return al
}

// Record submits an audit event to the asynchronous logging buffer.
func (al *AuditLogger) Record(ctx context.Context, event *AuditEvent) {
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
		if err := al.repo.WriteSync(context.Background(), event); err != nil {
			al.logger.Error("Failed to perform synchronous audit write.", zap.Error(err), zap.String("event_id", event.ID))
		}
	}
}

// flushLoop is the core worker goroutine that batches and writes events.
func (al *AuditLogger) flushLoop(ctx context.Context) {
	defer al.wg.Done()

	batch := make([]*AuditEvent, 0, al.batchSize)
	ticker := time.NewTicker(al.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event := <-al.buffer:
			batch = append(batch, event)
			if len(batch) >= al.batchSize {
				al.flushBatch(context.Background(), batch)
				batch = make([]*AuditEvent, 0, al.batchSize) // Reset batch
			}
		case <-ticker.C:
			if len(batch) > 0 {
				al.flushBatch(context.Background(), batch)
				batch = make([]*AuditEvent, 0, al.batchSize) // Reset batch
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
			return
		}
	}
}

// flushBatch writes a batch of events to the repository.
func (al *AuditLogger) flushBatch(ctx context.Context, batch []*AuditEvent) {
	if err := al.repo.WriteBatch(ctx, batch); err != nil {
		al.logger.Error("Failed to flush audit event batch.", zap.Error(err), zap.Int("batch_size", len(batch)))
		// In a real-world scenario, you might add a retry mechanism or write to a dead-letter queue.
	}
}

// Shutdown gracefully stops the logger, ensuring all buffered events are flushed.
func (al *AuditLogger) Shutdown() {
	al.cancel()
	al.wg.Wait()
}

func (al *AuditLogger) GetRepo() AuditRepository {
	return al.repo
}
