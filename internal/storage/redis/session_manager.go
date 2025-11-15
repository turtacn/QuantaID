package redis

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// SessionManager provides a robust, centralized mechanism for handling user sessions.
type SessionManager struct {
	client        RedisClientInterface
	config        SessionConfig
	logger        *zap.Logger
	uuidGenerator UUIDGenerator
	clock         Clock
	metrics       *Metrics
}

// SessionConfig holds the configuration for session management.
type SessionConfig struct {
	DefaultTTL          time.Duration `yaml:"default_ttl"`
	EnableRotation      bool          `yaml:"enable_rotation"`
	RotationInterval    time.Duration `yaml:"rotation_interval"`
	EnableDeviceBinding bool          `yaml:"enable_device_binding"`
	MaxSessionsPerUser  int           `yaml:"max_sessions_per_user"`
}

// NewSessionManager creates a new session manager.
func NewSessionManager(client RedisClientInterface, config SessionConfig, logger *zap.Logger, uuidGenerator UUIDGenerator, clock Clock, metrics *Metrics) *SessionManager {
	return &SessionManager{
		client:        client,
		config:        config,
		logger:        logger,
		uuidGenerator: uuidGenerator,
		clock:         clock,
		metrics:       metrics,
	}
}

// computeDeviceFingerprint creates a basic device fingerprint from request headers.
func computeDeviceFingerprint(r *http.Request) string {
	ua := r.UserAgent()
	hash := sha256.Sum256([]byte(ua))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// CreateSession creates a new session for a user.
func (sm *SessionManager) CreateSession(ctx context.Context, userID string, r *http.Request) (*types.UserSession, error) {
	if sm.config.MaxSessionsPerUser > 0 {
		if err := sm.enforceMaxSessions(ctx, userID); err != nil {
			sm.logger.Warn("Failed to enforce max sessions", zap.String("userID", userID), zap.Error(err))
			// Continue to create session even if cleanup fails
		}
	}

	sessionID := sm.uuidGenerator.New()

	now := sm.clock.Now()
	session := &types.UserSession{
		ID:            sessionID,
		UserID:        userID,
		CreatedAt:     now,
		ExpiresAt:     now.Add(sm.config.DefaultTTL),
		LastRotatedAt: now,
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	data, err := json.Marshal(session)
	if err != nil {
		sm.metrics.Errors.WithLabelValues("marshal").Inc()
		return nil, fmt.Errorf("could not marshal session: %w", err)
	}

	start := sm.clock.Now()
	ok, err := sm.client.SetNX(ctx, sessionKey, data, sm.config.DefaultTTL).Result()
	sm.metrics.Commands.WithLabelValues("setnx").Inc()
	sm.metrics.CommandLatency.WithLabelValues("setnx").Observe(time.Since(start).Seconds())
	if err != nil {
		sm.metrics.Errors.WithLabelValues("setnx").Inc()
		return nil, fmt.Errorf("could not store session in redis: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("session already exists")
	}

	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	start = sm.clock.Now()
	if err := sm.client.ZAdd(ctx, userSessionsKey, redis.Z{Score: float64(now.UnixNano()), Member: sessionID}); err != nil {
		sm.metrics.Commands.WithLabelValues("zadd").Inc()
		sm.metrics.CommandLatency.WithLabelValues("zadd").Observe(time.Since(start).Seconds())
		sm.metrics.Errors.WithLabelValues("zadd").Inc()
		// Attempt to clean up the session key if adding to the set fails
		_ = sm.client.Del(ctx, sessionKey)
		return nil, fmt.Errorf("could not add session to user set: %w", err)
	}
	sm.metrics.Commands.WithLabelValues("zadd").Inc()
	sm.metrics.CommandLatency.WithLabelValues("zadd").Observe(time.Since(start).Seconds())

	sm.logger.Info("Session created", zap.String("userID", userID), zap.String("sessionID", sessionID))
	return session, nil
}

// GetSession retrieves and validates a session.
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string, r *http.Request) (*types.UserSession, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	start := sm.clock.Now()
	data, err := sm.client.Get(ctx, key)
	sm.metrics.Commands.WithLabelValues("get").Inc()
	sm.metrics.CommandLatency.WithLabelValues("get").Observe(time.Since(start).Seconds())
	if err != nil {
		sm.metrics.Errors.WithLabelValues("get").Inc()
		// Consider specific error types, e.g., redis.Nil
		return nil, types.ErrNotFound
	}

	var session types.UserSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		sm.metrics.Errors.WithLabelValues("unmarshal").Inc()
		return nil, fmt.Errorf("could not unmarshal session: %w", err)
	}

	if sm.clock.Now().After(session.ExpiresAt) {
		// Proactively delete expired session
		sm.DeleteSession(ctx, session.UserID, session.ID)
		return nil, types.ErrSessionExpired
	}

	if sm.config.EnableDeviceBinding && session.DeviceFingerprint != computeDeviceFingerprint(r) {
		sm.logger.Warn("Device fingerprint mismatch", zap.String("sessionID", sessionID))
		return nil, types.ErrDeviceMismatch
	}

	return &session, nil
}

// DeleteSession removes a user's session.
func (sm *SessionManager) DeleteSession(ctx context.Context, userID, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)

	start := sm.clock.Now()
	if err := sm.client.Del(ctx, sessionKey); err != nil {
		sm.metrics.Commands.WithLabelValues("del").Inc()
		sm.metrics.CommandLatency.WithLabelValues("del").Observe(time.Since(start).Seconds())
		sm.metrics.Errors.WithLabelValues("del").Inc()
		return fmt.Errorf("could not delete session key: %w", err)
	}
	sm.metrics.Commands.WithLabelValues("del").Inc()
	sm.metrics.CommandLatency.WithLabelValues("del").Observe(time.Since(start).Seconds())

	start = sm.clock.Now()
	if _, err := sm.client.ZRem(ctx, userSessionsKey, sessionID); err != nil {
		sm.metrics.Commands.WithLabelValues("zrem").Inc()
		sm.metrics.CommandLatency.WithLabelValues("zrem").Observe(time.Since(start).Seconds())
		sm.metrics.Errors.WithLabelValues("zrem").Inc()
		return fmt.Errorf("could not remove session from user set: %w", err)
	}
	sm.metrics.Commands.WithLabelValues("zrem").Inc()
	sm.metrics.CommandLatency.WithLabelValues("zrem").Observe(time.Since(start).Seconds())

	sm.logger.Info("Session deleted", zap.String("userID", userID), zap.String("sessionID", sessionID))
	return nil
}

// enforceMaxSessions ensures the user does not exceed the maximum number of concurrent sessions.
func (sm *SessionManager) enforceMaxSessions(ctx context.Context, userID string) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	start := sm.clock.Now()
	count, err := sm.client.ZCard(ctx, userSessionsKey)
	sm.metrics.Commands.WithLabelValues("zcard").Inc()
	sm.metrics.CommandLatency.WithLabelValues("zcard").Observe(time.Since(start).Seconds())
	if err != nil {
		sm.metrics.Errors.WithLabelValues("zcard").Inc()
		return fmt.Errorf("could not get user session count: %w", err)
	}

	if count < int64(sm.config.MaxSessionsPerUser) {
		return nil
	}

	sm.logger.Info("Max session limit reached, evicting oldest session", zap.String("userID", userID))

	// Evict the oldest session(s) to make room for the new one.
	numToEvict := count - int64(sm.config.MaxSessionsPerUser) + 1
	if numToEvict <= 0 {
		return nil
	}

	start = sm.clock.Now()
	sessionsToRemove, err := sm.client.ZRemRangeByRank(ctx, userSessionsKey, 0, numToEvict-1)
	sm.metrics.Commands.WithLabelValues("zremrangebyrank").Inc()
	sm.metrics.CommandLatency.WithLabelValues("zremrangebyrank").Observe(time.Since(start).Seconds())
	if err != nil {
		sm.metrics.Errors.WithLabelValues("zremrangebyrank").Inc()
		return fmt.Errorf("could not retrieve sessions for eviction: %w", err)
	}

	if sessionsToRemove > 0 {
		sm.logger.Info("Evicted oldest sessions", zap.Int64("count", sessionsToRemove), zap.String("userID", userID))
	}

	return nil
}

// RotateSession generates a new session ID for an existing session (session fixation prevention).
func (sm *SessionManager) RotateSession(ctx context.Context, oldSessionID string, r *http.Request) (*types.UserSession, error) {
	oldSession, err := sm.GetSession(ctx, oldSessionID, r)
	if err != nil {
		return nil, err
	}

	// Create the new session first.
	newSession, err := sm.CreateSession(ctx, oldSession.UserID, r)
	if err != nil {
		return nil, fmt.Errorf("could not create new session during rotation: %w", err)
	}
	newSession.DeviceFingerprint = oldSession.DeviceFingerprint
	newSession.IPAddress = oldSession.IPAddress
	newSession.UserAgent = oldSession.UserAgent

	// Update the old session key to point to the new session ID for a grace period.
	// This handles race conditions where a request with the old session ID arrives
	// shortly after rotation.
	gracePeriod := 5 * time.Minute
	start := sm.clock.Now()
	if err := sm.client.Set(ctx, fmt.Sprintf("session:%s", oldSessionID), newSession.ID, gracePeriod); err != nil {
		sm.metrics.Commands.WithLabelValues("set").Inc()
		sm.metrics.CommandLatency.WithLabelValues("set").Observe(time.Since(start).Seconds())
		sm.metrics.Errors.WithLabelValues("set").Inc()
		// If this fails, the rotation is still successful, but the grace period won't be available.
		sm.logger.Warn("Failed to set grace period for old session", zap.Error(err))
	}
	sm.metrics.Commands.WithLabelValues("set").Inc()
	sm.metrics.CommandLatency.WithLabelValues("set").Observe(time.Since(start).Seconds())


	// Remove the old session ID from the user's set of active sessions.
	userSessionsKey := fmt.Sprintf("user_sessions:%s", oldSession.UserID)
	start = sm.clock.Now()
	if _, err := sm.client.ZRem(ctx, userSessionsKey, oldSessionID); err != nil {
		sm.metrics.Commands.WithLabelValues("zrem").Inc()
		sm.metrics.CommandLatency.WithLabelValues("zrem").Observe(time.Since(start).Seconds())
		sm.metrics.Errors.WithLabelValues("zrem").Inc()
		sm.logger.Warn("Failed to remove old session from user set during rotation", zap.Error(err))
	}
	sm.metrics.Commands.WithLabelValues("zrem").Inc()
	sm.metrics.CommandLatency.WithLabelValues("zrem").Observe(time.Since(start).Seconds())


	sm.logger.Info("Session rotated", zap.String("oldSessionID", oldSessionID), zap.String("newSessionID", newSession.ID))

	return newSession, nil
}
