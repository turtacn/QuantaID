package postgresql

import (
	"context"
	"errors"

	"github.com/turtacn/QuantaID/internal/domain/apikey"
	"github.com/turtacn/QuantaID/pkg/types"
	"gorm.io/gorm"
)

// APIKeyRepository implements the apikey.Repository interface using GORM.
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new APIKeyRepository.
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create persists a new API key.
func (r *APIKeyRepository) Create(ctx context.Context, key *apikey.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetByID retrieves an API key by its unique ID.
func (r *APIKeyRepository) GetByID(ctx context.Context, id string) (*apikey.APIKey, error) {
	var key apikey.APIKey
	if err := r.db.WithContext(ctx).First(&key, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	return &key, nil
}

// GetByPrefix retrieves API keys matching a given prefix.
func (r *APIKeyRepository) GetByPrefix(ctx context.Context, prefix string) ([]*apikey.APIKey, error) {
	var keys []*apikey.APIKey
	if err := r.db.WithContext(ctx).Where("prefix = ?", prefix).Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// GetByKeyID retrieves an API key by its public KeyID.
func (r *APIKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*apikey.APIKey, error) {
	var key apikey.APIKey
	if err := r.db.WithContext(ctx).First(&key, "key_id = ?", keyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.ErrNotFound // Using generic Not Found for security
		}
		return nil, err
	}
	return &key, nil
}

// Update modifies an existing API key.
func (r *APIKeyRepository) Update(ctx context.Context, key *apikey.APIKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

// Delete removes an API key.
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&apikey.APIKey{}, "id = ?", id).Error
}

// ListByAppID returns all API keys for a specific application.
func (r *APIKeyRepository) ListByAppID(ctx context.Context, appID string) ([]*apikey.APIKey, error) {
	var keys []*apikey.APIKey
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// GetRateLimitPolicy retrieves the rate limit policy for a specific application.
func (r *APIKeyRepository) GetRateLimitPolicy(ctx context.Context, appID string) (*apikey.RateLimitPolicy, error) {
	var policy apikey.RateLimitPolicy
	if err := r.db.WithContext(ctx).First(&policy, "app_id = ?", appID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no policy found, implying default
		}
		return nil, err
	}
	return &policy, nil
}
