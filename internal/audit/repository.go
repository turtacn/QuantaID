package audit

import (
	"context"
	"time"

	"github.com/turtacn/QuantaID/pkg/audit/events"
)

// AuditRepository defines the interface for storing and retrieving audit events.
type AuditRepository interface {
	// WriteBatch writes a slice of audit events to the persistent storage.
	// This method is designed for high-throughput, asynchronous logging.
	WriteBatch(ctx context.Context, events []*events.AuditEvent) error

	// WriteSync writes a single audit event immediately to persistent storage.
	// This is a fallback for critical events or when the async buffer is full.
	WriteSync(ctx context.Context, event *events.AuditEvent) error

	// Query retrieves audit events based on a set of filters and pagination options.
	// The implementation should handle complex queries on indexed fields.
	Query(ctx context.Context, filter QueryFilter) ([]*events.AuditEvent, error)

	// DeleteBefore deletes all audit events created before the specified time.
	DeleteBefore(ctx context.Context, cutoff time.Time) error
}

// QueryFilter defines the criteria for querying audit logs.
type QueryFilter struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
	EventTypes     []events.EventType
	ActorID        string
	TargetID       string
}
