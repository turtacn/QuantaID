package postgresql

import (
	"context"

	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresAuthRepository provides a GORM-based implementation of the auth-related repositories.
type PostgresAuthRepository struct {
	db *gorm.DB
}

// NewPostgresAuthRepository creates a new PostgreSQL auth repository.
func NewPostgresAuthRepository(db *gorm.DB) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}

// --- IdentityProviderRepository Implementation ---

// CreateProvider adds a new identity provider to the database.
func (r *PostgresAuthRepository) CreateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

// GetProviderByID retrieves an identity provider by its ID from the database.
func (r *PostgresAuthRepository) GetProviderByID(ctx context.Context, id string) (*types.IdentityProvider, error) {
	var provider types.IdentityProvider
	err := r.db.WithContext(ctx).First(&provider, "id = ?", id).Error
	return &provider, err
}

// GetProviderByName searches for an identity provider by its name in the database.
func (r *PostgresAuthRepository) GetProviderByName(ctx context.Context, name string) (*types.IdentityProvider, error) {
	var provider types.IdentityProvider
	err := r.db.WithContext(ctx).First(&provider, "name = ?", name).Error
	return &provider, err
}

// ListProviders returns all identity providers from the database.
func (r *PostgresAuthRepository) ListProviders(ctx context.Context) ([]*types.IdentityProvider, error) {
	var providers []*types.IdentityProvider
	err := r.db.WithContext(ctx).Find(&providers).Error
	return providers, err
}

// UpdateProvider updates an existing identity provider in the database.
func (r *PostgresAuthRepository) UpdateProvider(ctx context.Context, provider *types.IdentityProvider) error {
	return r.db.WithContext(ctx).Save(provider).Error
}

// DeleteProvider removes an identity provider from the database.
func (r *PostgresAuthRepository) DeleteProvider(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&types.IdentityProvider{}, "id = ?", id).Error
}