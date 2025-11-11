package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/pkg/types"
)

// RedisSessionRepository provides a Redis-backed implementation of the auth.SessionRepository interface.
type RedisSessionRepository struct {
	client *redis.Client
}

// NewRedisSessionRepository creates a new Redis session repository.
func NewRedisSessionRepository(client *redis.Client) *RedisSessionRepository {
	return &RedisSessionRepository{
		client: client,
	}
}

// CreateSession stores a new user session in Redis with a specified duration.
func (r *RedisSessionRepository) CreateSession(ctx context.Context, session *types.UserSession) error {
	key := fmt.Sprintf("session:%s", session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("session already expired")
	}

	return r.client.SetEx(ctx, key, data, ttl).Err()
}

// GetSession retrieves a session by its ID from Redis.
func (r *RedisSessionRepository) GetSession(ctx context.Context, sessionID string) (*types.UserSession, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var session types.UserSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &session, nil
}

// DeleteSession removes a session from Redis by its ID.
func (r *RedisSessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Del(ctx, key).Err()
}

// GetUserSessions retrieves all active sessions for a specific user from Redis.
// This implementation uses SCAN to avoid blocking the server with KEYS.
func (r *RedisSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) {
	var userSessions []*types.UserSession
	iter := r.client.Scan(ctx, 0, "session:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			// Log the error but continue processing other keys
			continue
		}

		var session types.UserSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		if session.UserID == userID {
			userSessions = append(userSessions, &session)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("redis scan: %w", err)
	}

	return userSessions, nil
}
