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

// CreateFactor adds a new MFA factor for a user.
func (r *PostgresMFARepository) CreateFactor(ctx context.Context, factor *types.MFAFactor) error {
	return r.db.WithContext(ctx).Create(factor).Error
}

// GetFactor retrieves an MFA factor by its ID.
func (r *PostgresMFARepository) GetFactor(ctx context.Context, factorID uuid.UUID) (*types.MFAFactor, error) {
	var factor types.MFAFactor
	err := r.db.WithContext(ctx).First(&factor, "id = ?", factorID).Error
	return &factor, err
}

// GetUserFactors retrieves all MFA factors for a user.
func (r *PostgresMFARepository) GetUserFactors(ctx context.Context, userID uuid.UUID) ([]*types.MFAFactor, error) {
	var factors []*types.MFAFactor
	err := r.db.WithContext(ctx).Find(&factors, "user_id = ?", userID).Error
	return factors, err
}

// GetUserFactorsByType retrieves all MFA factors of a specific type for a user.
func (r *PostgresMFARepository) GetUserFactorsByType(ctx context.Context, userID uuid.UUID, factorType string) ([]*types.MFAFactor, error) {
	var factors []*types.MFAFactor
	err := r.db.WithContext(ctx).Find(&factors, "user_id = ? AND type = ?", userID, factorType).Error
	return factors, err
}

// UpdateFactor updates an existing MFA factor.
func (r *PostgresMFARepository) UpdateFactor(ctx context.Context, factor *types.MFAFactor) error {
	return r.db.WithContext(ctx).Save(factor).Error
}

// DeleteFactor removes an MFA factor.
func (r *PostgresMFARepository) DeleteFactor(ctx context.Context, factorID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&types.MFAFactor{}, "id = ?", factorID).Error
}

// CreateVerificationLog adds a new MFA verification log entry.
func (r *PostgresMFARepository) CreateVerificationLog(ctx context.Context, log *types.MFAVerificationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
