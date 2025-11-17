package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/turtacn/QuantaID/internal/audit"
)

// AuditRepository implements the audit.AuditRepository interface for PostgreSQL.
type AuditRepository struct {
	db *pg.DB
}

// NewAuditRepository creates a new PostgreSQL audit repository.
func NewAuditRepository(db *pg.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// pgAuditEvent is a private struct that maps directly to the database schema.
// We use this to avoid adding pg-specific tags to the domain model.
type pgAuditEvent struct {
	tableName struct{} `pg:"audit_logs"`

	ID          string                 `pg:"id,pk"`
	Timestamp   time.Time              `pg:"timestamp,notnull"`
	EventType   audit.EventType        `pg:"event_type,notnull"`
	ActorID     string                 `pg:"actor_id"`
	ActorType   string                 `pg:"actor_type"`
	ActorName   string                 `pg:"actor_name"`
	TargetID    string                 `pg:"target_id"`
	TargetType  string                 `pg:"target_type"`
	TargetName  string                 `pg:"target_name"`
	Action      string                 `pg:"action,notnull"`
	Result      audit.Result           `pg:"result,notnull"`
	Metadata    map[string]interface{} `pg:"metadata,jsonb"`
	IPAddress   string                 `pg:"ip_address"`
	UserAgent   string                 `pg:"user_agent"`
	CreatedAt   time.Time              `pg:"created_at,default:now()"`
}

// toPGEvent converts a domain AuditEvent to its PostgreSQL representation.
func toPGEvent(e *audit.AuditEvent) *pgAuditEvent {
	return &pgAuditEvent{
		ID:         e.ID,
		Timestamp:  e.Timestamp,
		EventType:  e.EventType,
		ActorID:    e.Actor.ID,
		ActorType:  e.Actor.Type,
		ActorName:  e.Actor.Name,
		TargetID:   e.Target.ID,
		TargetType: e.Target.Type,
		TargetName: e.Target.Name,
		Action:     e.Action,
		Result:     e.Result,
		Metadata:   e.Metadata,
		IPAddress:  e.IPAddress,
		UserAgent:  e.UserAgent,
	}
}


// WriteBatch writes a slice of audit events to the database in a single transaction.
func (r *AuditRepository) WriteBatch(ctx context.Context, events []*audit.AuditEvent) error {
	if len(events) == 0 {
		return nil
	}

	pgEvents := make([]*pgAuditEvent, len(events))
	for i, e := range events {
		pgEvents[i] = toPGEvent(e)
	}

	_, err := r.db.ModelContext(ctx, &pgEvents).Insert()
	if err != nil {
		return fmt.Errorf("failed to insert audit log batch: %w", err)
	}

	return nil
}

// WriteSync writes a single audit event to the database immediately.
func (r *AuditRepository) WriteSync(ctx context.Context, event *audit.AuditEvent) error {
	pgEvent := toPGEvent(event)
	_, err := r.db.ModelContext(ctx, pgEvent).Insert()
	return err
}

// fromPGEvent converts a PostgreSQL event representation back to the domain model.
func fromPGEvent(e *pgAuditEvent) *audit.AuditEvent {
	return &audit.AuditEvent{
		ID:        e.ID,
		Timestamp: e.Timestamp,
		EventType: e.EventType,
		Actor: audit.Actor{
			ID:   e.ActorID,
			Type: e.ActorType,
			Name: e.ActorName,
		},
		Target: audit.Target{
			ID:   e.TargetID,
			Type: e.TargetType,
			Name: e.TargetName,
		},
		Action:    e.Action,
		Result:    e.Result,
		Metadata:  e.Metadata,
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
	}
}


// Query retrieves audit logs from the database based on the provided filters.
func (r *AuditRepository) Query(ctx context.Context, filter audit.QueryFilter) ([]*audit.AuditEvent, error) {
	var pgEvents []*pgAuditEvent
	query := r.db.ModelContext(ctx, &pgEvents).Order("timestamp DESC")

	if !filter.StartTimestamp.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartTimestamp)
	}
	if !filter.EndTimestamp.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndTimestamp)
	}
	if len(filter.EventTypes) > 0 {
		query = query.Where("event_type IN (?)", pg.In(filter.EventTypes))
	}
	if filter.ActorID != "" {
		query = query.Where("actor_id = ?", filter.ActorID)
	}
	if filter.TargetID != "" {
		query = query.Where("target_id = ?", filter.TargetID)
	}

	err := query.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return []*audit.AuditEvent{}, nil
		}
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}

	domainEvents := make([]*audit.AuditEvent, len(pgEvents))
	for i, e := range pgEvents {
		domainEvents[i] = fromPGEvent(e)
	}

	return domainEvents, nil
}
