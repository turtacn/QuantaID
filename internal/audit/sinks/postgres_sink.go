package sinks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/turtacn/QuantaID/pkg/audit/events"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// PostgresSink implements the Sink interface for PostgreSQL using GORM.
type PostgresSink struct {
	db *gorm.DB
}

// NewPostgresSink creates a new PostgresSink instance.
func NewPostgresSink(db *gorm.DB) *PostgresSink {
	return &PostgresSink{db: db}
}

// WriteBatch writes a slice of audit events to the database.
func (s *PostgresSink) WriteBatch(ctx context.Context, batch []*events.AuditEvent) error {
	if len(batch) == 0 {
		return nil
	}

	// Convert domain events to GORM models if necessary.
	// Currently assuming AuditEvent is compatible or GORM can handle it with JSON serialization.
	// Since the original PostgresAuditLogRepository used *audit.AuditEvent, we assume it works,
	// but we must ensure JSON fields are handled.

	// We use a dedicated model struct to ensure correct GORM behavior, specifically for JSONB.
	dbEvents := make([]*AuditLogEntry, len(batch))
	for i, e := range batch {
		dbEvents[i] = toAuditLogEntry(e)
	}

	return s.db.WithContext(ctx).Create(&dbEvents).Error
}

// WriteSync writes a single audit event immediately to the database.
func (s *PostgresSink) WriteSync(ctx context.Context, event *events.AuditEvent) error {
	dbEvent := toAuditLogEntry(event)
	return s.db.WithContext(ctx).Create(dbEvent).Error
}

// Close closes the sink (no-op for PostgresSink as DB connection is managed externally).
func (s *PostgresSink) Close() error {
	return nil
}

// --- GORM Model ---

// AuditLogEntry is the GORM model for the audit log table.
type AuditLogEntry struct {
	ID        string          `gorm:"primaryKey;type:varchar(36)"`
	Timestamp time.Time       `gorm:"index"`
	EventType string          `gorm:"index;type:varchar(255)"`
	Actor     json.RawMessage `gorm:"type:jsonb"`
	Target    json.RawMessage `gorm:"type:jsonb"`
	Action    string          `gorm:"type:varchar(255)"`
	Result    string          `gorm:"type:varchar(50)"`
	Metadata  json.RawMessage `gorm:"type:jsonb"`
	IPAddress string          `gorm:"type:varchar(45)"`
	UserAgent string          `gorm:"type:text"`
	TraceID   string          `gorm:"type:varchar(64)"`
	Category  string          `gorm:"type:varchar(100)"`
	UserID    string          `gorm:"index;type:varchar(36)"` // Derived for easier query if needed
	Details   json.RawMessage `gorm:"type:jsonb"`
}

// TableName overrides the table name.
func (AuditLogEntry) TableName() string {
	return "audit_logs"
}

// toAuditLogEntry converts a domain AuditEvent to the GORM model.
func toAuditLogEntry(e *events.AuditEvent) *AuditLogEntry {
	actorBytes, _ := json.Marshal(e.Actor)
	targetBytes, _ := json.Marshal(e.Target)
	metadataBytes, _ := json.Marshal(e.Metadata)
	detailsBytes, _ := json.Marshal(e.Details)

	return &AuditLogEntry{
		ID:        e.ID,
		Timestamp: e.Timestamp,
		EventType: string(e.EventType),
		Actor:     actorBytes,
		Target:    targetBytes,
		Action:    e.Action,
		Result:    string(e.Result),
		Metadata:  metadataBytes,
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
		TraceID:   e.TraceID,
		Category:  e.Category,
		UserID:    e.UserID,
		Details:   detailsBytes,
	}
}

// Ensure AuditLogEntry implements schema.Tabler if needed (already does via TableName)
var _ schema.Tabler = (*AuditLogEntry)(nil)
