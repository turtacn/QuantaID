package sinks

import (
	"context"

	"github.com/turtacn/QuantaID/pkg/audit/events"
)

// Sink defines the interface for an audit log destination.
type Sink interface {
	// WriteBatch writes a slice of audit events to the sink.
	WriteBatch(ctx context.Context, events []*events.AuditEvent) error

	// WriteSync writes a single audit event immediately to the sink.
	WriteSync(ctx context.Context, event *events.AuditEvent) error

	// Close cleans up any resources used by the sink.
	Close() error
}
