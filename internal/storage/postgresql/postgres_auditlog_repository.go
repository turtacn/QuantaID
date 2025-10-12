package postgresql

import (
	"context"

	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresAuditLogRepository provides a GORM-based implementation of the auditlog-related repositories.
type PostgresAuditLogRepository struct {
	db *gorm.DB
}

// NewPostgresAuditLogRepository creates a new PostgreSQL auditlog repository.
func NewPostgresAuditLogRepository(db *gorm.DB) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

// --- AuditLogRepository Implementation ---

// CreateLogEntry adds a new audit log entry to the database.
func (r *PostgresAuditLogRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// GetLogsForUser retrieves a paginated list of audit logs for a specific user.
func (r *PostgresAuditLogRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	var logs []*types.AuditLog
	err := r.db.WithContext(ctx).Where("actor_id = ?", userID).Offset(pq.Offset).Limit(pq.PageSize).Find(&logs).Error
	return logs, err
}

// GetLogsByAction retrieves a paginated list of audit logs for a specific action.
func (r *PostgresAuditLogRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	var logs []*types.AuditLog
	err := r.db.WithContext(ctx).Where("action = ?", action).Offset(pq.Offset).Limit(pq.PageSize).Find(&logs).Error
	return logs, err
}