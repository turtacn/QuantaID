package postgresql

import (
	"context"
	"time"

	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/pkg/audit/events"
	"gorm.io/gorm"
)

type PostgresAuditLogRepository struct {
	db *gorm.DB
}

func NewPostgresAuditLogRepository(db *gorm.DB) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

func (r *PostgresAuditLogRepository) WriteSync(ctx context.Context, event *events.AuditEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *PostgresAuditLogRepository) WriteBatch(ctx context.Context, evs []*events.AuditEvent) error {
	return r.db.WithContext(ctx).Create(&evs).Error
}

func (r *PostgresAuditLogRepository) Query(ctx context.Context, filter audit.QueryFilter) ([]*events.AuditEvent, error) {
	var results []*events.AuditEvent
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

	err := query.Find(&results).Error
	return results, err
}

func (r *PostgresAuditLogRepository) DeleteBefore(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", cutoff).Delete(&events.AuditEvent{}).Error
}

func (r *PostgresAuditLogRepository) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
