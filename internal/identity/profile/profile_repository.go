package profile

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ProfileRepository defines the interface for profile data access
type ProfileRepository interface {
	Create(ctx context.Context, profile *UserProfile) error
	GetByUserID(ctx context.Context, userID string) (*UserProfile, error)
	Update(ctx context.Context, profile *UserProfile) error
	UpdateRisk(ctx context.Context, userID string, risk RiskIndicators, score int, level string) error
	UpdateBehavior(ctx context.Context, userID string, behavior BehaviorMetrics) error
	UpdateTags(ctx context.Context, userID string, autoTags, manualTags StringSlice) error
	UpdateQuality(ctx context.Context, userID string, score int, details QualityDetails) error
	FindByRiskLevel(ctx context.Context, tenantID, level string, limit int) ([]*UserProfile, error)
	FindByTag(ctx context.Context, tenantID, tag string) ([]*UserProfile, error)
	Delete(ctx context.Context, userID string) error
}

// PostgresProfileRepository implements ProfileRepository for PostgreSQL
type PostgresProfileRepository struct {
	db *gorm.DB
}

// NewPostgresProfileRepository creates a new PostgresProfileRepository
func NewPostgresProfileRepository(db *gorm.DB) ProfileRepository {
	return &PostgresProfileRepository{db: db}
}

// Create creates a new user profile
func (r *PostgresProfileRepository) Create(ctx context.Context, profile *UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// GetByUserID retrieves a profile by user ID
func (r *PostgresProfileRepository) GetByUserID(ctx context.Context, userID string) (*UserProfile, error) {
	var profile UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Return nil if not found, let caller handle initialization
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// Update updates a user profile
func (r *PostgresProfileRepository) Update(ctx context.Context, profile *UserProfile) error {
	profile.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(profile).Error
}

// UpdateRisk updates risk indicators and score
func (r *PostgresProfileRepository) UpdateRisk(ctx context.Context, userID string, risk RiskIndicators, score int, level string) error {
	updates := map[string]interface{}{
		"risk":                risk,
		"risk_score":          score,
		"risk_level":          level,
		"last_risk_update_at": time.Now(),
		"updated_at":          time.Now(),
	}
	return r.db.WithContext(ctx).Model(&UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

// UpdateBehavior updates behavior metrics
func (r *PostgresProfileRepository) UpdateBehavior(ctx context.Context, userID string, behavior BehaviorMetrics) error {
	updates := map[string]interface{}{
		"behavior":   behavior,
		"updated_at": time.Now(),
	}
	return r.db.WithContext(ctx).Model(&UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

// UpdateTags updates auto and manual tags
func (r *PostgresProfileRepository) UpdateTags(ctx context.Context, userID string, autoTags, manualTags StringSlice) error {
	updates := map[string]interface{}{
		"auto_tags":   autoTags,
		"manual_tags": manualTags,
		"updated_at":  time.Now(),
	}
	return r.db.WithContext(ctx).Model(&UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

// UpdateQuality updates quality score and details
func (r *PostgresProfileRepository) UpdateQuality(ctx context.Context, userID string, score int, details QualityDetails) error {
	updates := map[string]interface{}{
		"quality_score":   score,
		"quality_details": details,
		"updated_at":      time.Now(),
	}
	return r.db.WithContext(ctx).Model(&UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

// FindByRiskLevel finds profiles by risk level for a tenant
func (r *PostgresProfileRepository) FindByRiskLevel(ctx context.Context, tenantID, level string, limit int) ([]*UserProfile, error) {
	var profiles []*UserProfile
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND risk_level = ?", tenantID, level).
		Limit(limit).
		Find(&profiles).Error
	return profiles, err
}

// FindByTag finds profiles by tag (auto or manual)
func (r *PostgresProfileRepository) FindByTag(ctx context.Context, tenantID, tag string) ([]*UserProfile, error) {
	var profiles []*UserProfile
	// PostgreSQL JSONB operator ? checks if key/element exists. GORM requires escaping ? as ??
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (auto_tags ?? ? OR manual_tags ?? ?)", tenantID, tag, tag).
		Find(&profiles).Error
	return profiles, err
}

// Delete deletes a user profile
func (r *PostgresProfileRepository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserProfile{}).Error
}
