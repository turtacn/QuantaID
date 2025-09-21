package redis

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"sync"
	"time"
)

type tokenWithValue struct {
	userID    string
	expiresAt time.Time
}

type denyListEntry struct {
	expiresAt time.Time
}

// InMemoryTokenRepository is an in-memory implementation of the TokenRepository.
type InMemoryTokenRepository struct {
	mu            sync.RWMutex
	refreshTokens map[string]tokenWithValue
	denyList      map[string]denyListEntry
}

// NewInMemoryTokenRepository creates a new in-memory token repository.
func NewInMemoryTokenRepository() *InMemoryTokenRepository {
	return &InMemoryTokenRepository{
		refreshTokens: make(map[string]tokenWithValue),
		denyList:      make(map[string]denyListEntry),
	}
}

func (r *InMemoryTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshTokens[token] = tokenWithValue{
		userID:    userID,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

func (r *InMemoryTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, exists := r.refreshTokens[token]
	if !exists {
		return "", types.ErrNotFound.WithDetails(map[string]string{"reason": "token not found"})
	}
	if time.Now().After(val.expiresAt) {
		return "", types.ErrNotFound.WithDetails(map[string]string{"reason": "token expired"})
	}
	return val.userID, nil
}

func (r *InMemoryTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.refreshTokens, token)
	return nil
}

func (r *InMemoryTokenRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.denyList[jti] = denyListEntry{
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

func (r *InMemoryTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, exists := r.denyList[jti]
	if !exists {
		return false, nil
	}
	if time.Now().After(entry.expiresAt) {
		return false, nil
	}
	return true, nil
}

//Personal.AI order the ending
