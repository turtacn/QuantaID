package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/turtacn/QuantaID/pkg/types"
)

// RedisSessionRepository provides a Redis-backed implementation of the auth.SessionRepository interface.
type RedisSessionRepository struct {
	client RedisClientInterface
}

// NewRedisSessionRepository creates a new Redis session repository.
func NewRedisSessionRepository(client RedisClientInterface) *RedisSessionRepository {
	return &RedisSessionRepository{
		client: client,
	}
}

// CreateSession stores a new user session in Redis.
func (r *RedisSessionRepository) CreateSession(ctx context.Context, session *types.UserSession, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl)
}

// GetSession retrieves a session by its ID from Redis.
func (r *RedisSessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	var session types.UserSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return &session, nil
}

// DeleteSession removes a session from Redis by its ID.
func (r *RedisSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Del(ctx, key)
}

// GetUserSessions retrieves all active sessions for a specific user from Redis.
func (r *RedisSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) {
	// This is a placeholder implementation. A real implementation would need to
	// efficiently query sessions by user ID, perhaps using a secondary index.
	return nil, fmt.Errorf("not implemented")
}
