package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// PostgresMFARepository provides a GORM-based implementation of the MFA repository.
type PostgresMFARepository struct {
	db *gorm.DB
}

// NewPostgresMFARepository creates a new PostgreSQL MFA repository.
func NewPostgresMFARepository(db *gorm.DB) *PostgresMFARepository {
	return &PostgresMFARepository{db: db}
}

// CreateUserMFAConfig adds a new MFA configuration for a user.
func (r *PostgresMFARepository) CreateUserMFAConfig(ctx context.Context, config *types.UserMFAConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetUserMFAConfig retrieves an MFA configuration for a user and a specific method.
func (r *PostgresMFARepository) GetUserMFAConfig(ctx context.Context, userID uuid.UUID, method string) (*types.UserMFAConfig, error) {
	var config types.UserMFAConfig
	err := r.db.WithContext(ctx).First(&config, "user_id = ? AND method = ?", userID, method).Error
	return &config, err
}

// GetUserMFAConfigs retrieves all MFA configurations for a user.
func (r *PostgresMFARepository) GetUserMFAConfigs(ctx context.Context, userID uuid.UUID) ([]*types.UserMFAConfig, error) {
	var configs []*types.UserMFAConfig
	err := r.db.WithContext(ctx).Find(&configs, "user_id = ?", userID).Error
	return configs, err
}


// UpdateUserMFAConfig updates an existing MFA configuration for a user.
func (r *PostgresMFARepository) UpdateUserMFAConfig(ctx context.Context, config *types.UserMFAConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// DeleteUserMFAConfig removes an MFA configuration for a user.
func (r *PostgresMFARepository) DeleteUserMFAConfig(ctx context.Context, userID uuid.UUID, method string) error {
	return r.db.WithContext(ctx).Delete(&types.UserMFAConfig{}, "user_id = ? AND method = ?", userID, method).Error
}

// CreateMFAVerificationLog adds a new MFA verification log entry.
func (r *PostgresMFARepository) CreateMFAVerificationLog(ctx context.Context, log *types.MFAVerificationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
