package apikey

import (
	"time"
)

// APIKey represents an application's access key for the platform.
// It is used for M2M authentication and rate limiting.
type APIKey struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	AppID       string     `json:"app_id" gorm:"index"`
	KeyID       string     `json:"key_id" gorm:"uniqueIndex"` // Public unique identifier for the key (part of the token)
	KeyHash     string     `json:"-"`
	Prefix      string     `json:"prefix" gorm:"index"` // Used for display and namespacing
	Scopes      []string   `json:"scopes" gorm:"serializer:json"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Revoked     bool       `json:"revoked"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RateLimitPolicy defines the rate limits for an application.
// If no policy exists for an AppID, a default policy is applied.
type RateLimitPolicy struct {
	AppID  string `json:"app_id"`
	Limit  int    `json:"limit"`  // Requests count
	Window int    `json:"window"` // In seconds
}
