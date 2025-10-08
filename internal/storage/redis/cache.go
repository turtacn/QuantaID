package redis

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
	"time"
)

// tokenWithValue stores the user ID and expiration for a refresh token.
type tokenWithValue struct {
	userID    string
	expiresAt time.Time
}

// denyListEntry stores the expiration time for a denied JWT ID (jti).
type denyListEntry struct {
	expiresAt time.Time
}

// InMemoryTokenRepository provides an in-memory implementation of the auth.TokenRepository interface.
// NOTE: Despite the package name 'redis', this is an IN-MEMORY implementation,
// likely used for testing or simple, non-persistent deployments. It uses maps
// with a mutex for thread-safe operations.
type InMemoryTokenRepository struct {
	mu            sync.RWMutex
	refreshTokens map[string]tokenWithValue
	denyList      map[string]denyListEntry
}

// NewInMemoryTokenRepository creates a new, empty in-memory token repository.
func NewInMemoryTokenRepository() *InMemoryTokenRepository {
	return &InMemoryTokenRepository{
		refreshTokens: make(map[string]tokenWithValue),
		denyList:      make(map[string]denyListEntry),
	}
}

// StoreRefreshToken saves a refresh token to the in-memory store with a specified duration.
func (r *InMemoryTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshTokens[token] = tokenWithValue{
		userID:    userID,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

// GetRefreshTokenUserID retrieves the user ID associated with a refresh token from the in-memory store.
// It returns an error if the token is not found or has expired.
func (r *InMemoryTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, exists := r.refreshTokens[token]
	if !exists {
		return "", types.ErrNotFound.WithDetails(map[string]string{"reason": "token not found"})
	}
	if time.Now().After(val.expiresAt) {
		// Clean up expired token
		go func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			delete(r.refreshTokens, token)
		}()
		return "", types.ErrNotFound.WithDetails(map[string]string{"reason": "token expired"})
	}
	return val.userID, nil
}

// DeleteRefreshToken removes a refresh token from the in-memory store.
func (r *InMemoryTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.refreshTokens, token)
	return nil
}

// AddToDenyList adds a JWT ID (jti) to the in-memory deny list with a specified duration.
// This is used to prevent the reuse of logged-out tokens.
func (r *InMemoryTokenRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.denyList[jti] = denyListEntry{
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

// IsInDenyList checks if a JWT ID (jti) exists in the in-memory deny list and has not expired.
func (r *InMemoryTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, exists := r.denyList[jti]
	if !exists {
		return false, nil
	}
	if time.Now().After(entry.expiresAt) {
		// Clean up expired entry
		go func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			delete(r.denyList, jti)
		}()
		return false, nil
	}
	return true, nil
}
