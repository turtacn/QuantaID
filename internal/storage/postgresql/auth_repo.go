package postgresql

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresAuthRepository provides a GORM-backed implementation of the auth-related repositories,
// specifically the IdentityProviderRepository and AuditLogRepository.
type PostgresAuthRepository struct {
	db *gorm.DB
}

// NewPostgresAuthRepository creates a new, GORM-backed auth repository.
func NewPostgresAuthRepository(db *gorm.DB) (auth.IdentityProviderRepository, auth.AuditLogRepository) {
	return &PostgresAuthRepository{db: db}, &PostgresAuthRepository{db: db}
}

// --- IdentityProviderRepository Implementation ---

// CreateProvider adds a new identity provider to the database.
func (r *PostgresAuthRepository) CreateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	if err := r.db.WithContext(ctx).Create(provider).Error; err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return types.ErrConflict.WithDetails(map[string]string{"field": pqErr.Constraint})
		}
		return err
	}
	return nil
}

// GetProviderByID retrieves an identity provider by its ID from the database.
func (r *PostgresAuthRepository) GetProviderByID(ctx context.Context, id string) (*types.IdentityProvider, error) {
	var provider types.IdentityProvider
	if err := r.db.WithContext(ctx).First(&provider, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrNotFound.WithDetails(map[string]string{"id": id})
		}
		return nil, err
	}
	return &provider, nil
}

// GetProviderByName searches for an identity provider by its name in the database.
func (r *PostgresAuthRepository) GetProviderByName(ctx context.Context, name string) (*types.IdentityProvider, error) {
	var provider types.IdentityProvider
	if err := r.db.WithContext(ctx).First(&provider, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrNotFound.WithDetails(map[string]string{"name": name})
		}
		return nil, err
	}
	return &provider, nil
}

// ListProviders returns all identity providers from the database.
// Note: In a real-world scenario, this should be paginated.
func (r *PostgresAuthRepository) ListProviders(ctx context.Context) ([]*types.IdentityProvider, error) {
	var providers []*types.IdentityProvider
	if err := r.db.WithContext(ctx).Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// UpdateProvider updates an existing identity provider in the database.
func (r *PostgresAuthRepository) UpdateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// DeleteProvider removes an identity provider from the database by its ID.
func (r *PostgresAuthRepository) DeleteProvider(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.IdentityProvider{}, "id = ?", id).Error
}

// --- AuditLogRepository Implementation ---

// CreateLogEntry adds a new audit log entry to the database.
func (r *PostgresAuthRepository) CreateLogEntry(ctx context.Context, entry *types.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// GetLogsForUser retrieves a paginated list of audit logs for a specific user from the database.
func (r *PostgresAuthRepository) GetLogsForUser(ctx context.Context, userID string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	var logs []*types.AuditLog
	err := r.db.WithContext(ctx).
		Where("actor_id = ?", userID).
		Order("timestamp desc").
		Offset(pq.Offset).
		Limit(pq.PageSize).
		Find(&logs).Error
	return logs, err
}

// GetLogsByAction retrieves a paginated list of audit logs for a specific action from the database.
func (r *PostgresAuthRepository) GetLogsByAction(ctx context.Context, action string, pq types.PaginationQuery) ([]*types.AuditLog, error) {
	var logs []*types.AuditLog
	err := r.db.WithContext(ctx).
		Where("action = ?", action).
		Order("timestamp desc").
		Offset(pq.Offset).
		Limit(pq.PageSize).
		Find(&logs).Error
	return logs, err
}
