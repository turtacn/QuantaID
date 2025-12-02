package apikey

import (
	"context"
)

// Repository defines the interface for managing API keys and rate limit policies.
type Repository interface {
	// Create persists a new API key.
	Create(ctx context.Context, key *APIKey) error

	// GetByID retrieves an API key by its unique ID.
	GetByID(ctx context.Context, id string) (*APIKey, error)

	// GetByPrefix retrieves API keys matching a given prefix.
	// Since prefixes might not be unique globally (though highly likely with random generation),
	// this returns a list.
	GetByPrefix(ctx context.Context, prefix string) ([]*APIKey, error)

	// GetByKeyID retrieves an API key by its public KeyID.
	GetByKeyID(ctx context.Context, keyID string) (*APIKey, error)

	// Update modifies an existing API key.
	Update(ctx context.Context, key *APIKey) error

	// Delete removes an API key (or marks it as deleted).
	Delete(ctx context.Context, id string) error

	// ListByAppID returns all API keys for a specific application.
	ListByAppID(ctx context.Context, appID string) ([]*APIKey, error)

	// GetRateLimitPolicy retrieves the rate limit policy for a specific application.
	// Returns nil if no specific policy is found (implies default).
	GetRateLimitPolicy(ctx context.Context, appID string) (*RateLimitPolicy, error)
}
