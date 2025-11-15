package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/pkg/types"
)

// RedisSessionRepository provides a Redis-backed implementation of the auth.SessionRepository interface.
// It leverages the SessionManager for all session-related operations.
type RedisSessionRepository struct {
	client       RedisClientInterface
	sessionManager *SessionManager
}

// NewRedisSessionRepository creates a new Redis session repository.
func NewRedisSessionRepository(client RedisClientInterface, sessionManager *SessionManager) *RedisSessionRepository {
	return &RedisSessionRepository{
		client:       client,
		sessionManager: sessionManager,
	}
}

// CreateSession stores a new user session in Redis. Note that the core logic is now in SessionManager.
// This function now expects an *http.Request to generate a device fingerprint.
func (r *RedisSessionRepository) CreateSession(ctx context.Context, userID string, req *http.Request) (*types.UserSession, error) {
	return r.sessionManager.CreateSession(ctx, userID, req)
}

// GetSession retrieves a session by its ID from Redis.
// This function now expects an *http.Request to validate the device fingerprint.
func (r *RedisSessionRepository) GetSession(ctx context.Context, sessionID string, req *http.Request) (*types.UserSession, error) {
	return r.sessionManager.GetSession(ctx, sessionID, req)
}

// DeleteSession removes a session from Redis by its ID.
func (r *RedisSessionRepository) DeleteSession(ctx context.Context, userID, sessionID string) error {
	return r.sessionManager.DeleteSession(ctx, userID, sessionID)
}

// GetUserSessions retrieves all active sessions for a specific user from Redis.
// This implementation is now efficient, using a Redis sorted set to track user sessions.
func (r *RedisSessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*types.UserSession, error) {
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	sessionIDs, err := r.client.ZRange(ctx, userSessionsKey, 0, -1)
	if err != nil {
		if err == redis.Nil {
			return []*types.UserSession{}, nil
		}
		return nil, fmt.Errorf("could not retrieve user sessions: %w", err)
	}

	var userSessions []*types.UserSession
	for _, sessionID := range sessionIDs {
		key := fmt.Sprintf("session:%s", sessionID)
		data, err := r.client.Get(ctx, key)
		if err != nil {
			// Session might have expired, which is acceptable.
			continue
		}

		var session types.UserSession
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			// Log this error but continue, as other sessions might be valid.
			continue
		}
		userSessions = append(userSessions, &session)
	}

	return userSessions, nil
}
