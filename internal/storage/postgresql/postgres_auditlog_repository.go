package postgresql

import (
	"context"
	"time"

	"github.com/turtacn/QuantaID/internal/audit"
	"gorm.io/gorm"
)

type PostgresAuditLogRepository struct {
	db *gorm.DB
}

func NewPostgresAuditLogRepository(db *gorm.DB) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

func (r *PostgresAuditLogRepository) WriteSync(ctx context.Context, event *audit.AuditEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *PostgresAuditLogRepository) WriteBatch(ctx context.Context, events []*audit.AuditEvent) error {
	return r.db.WithContext(ctx).Create(&events).Error
}

func (r *PostgresAuditLogRepository) Query(ctx context.Context, filter audit.QueryFilter) ([]*audit.AuditEvent, error) {
	var events []*audit.AuditEvent
	query := r.db.WithContext(ctx)

	if !filter.StartTimestamp.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartTimestamp)
	}
	if !filter.EndTimestamp.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndTimestamp)
	}
	if len(filter.EventTypes) > 0 {
		query = query.Where("event_type IN ?", filter.EventTypes)
	}
	if filter.ActorID != "" {
		query = query.Where("actor ->> 'id' = ?", filter.ActorID)
	}
	if filter.TargetID != "" {
		query = query.Where("target ->> 'id' = ?", filter.TargetID)
	}

	err := query.Find(&events).Error
	return events, err
}

func (r *PostgresAuditLogRepository) DeleteBefore(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", cutoff).Delete(&audit.AuditEvent{}).Error
}
