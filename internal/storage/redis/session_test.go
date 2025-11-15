package redis

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

func TestSessionManager_RotateSession(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := zap.NewNop()
	rc := &redisClient{client: db}
	uuidGen := &MockUUIDGenerator{uuid: "new-session-id"}
	clock := &MockClock{}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	sm := NewSessionManager(rc, SessionConfig{
		DefaultTTL: 24 * time.Hour,
	}, logger, uuidGen, clock, metrics)

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/", nil)
	oldSessionID := "old-session-id"

	// Mock the old session
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	clock.SetNow(now)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("User-Agent", "curl/7.64.1")
	oldSession := &types.UserSession{
		ID:                oldSessionID,
		UserID:            "user-123",
		ExpiresAt:         clock.Now().Add(24 * time.Hour),
		IPAddress:         "192.0.2.1:1234",
		UserAgent:         "curl/7.64.1",
		DeviceFingerprint: "47DEQpj8HBSa+/TImW+5JCeuQeRKm5NMpJWZG3hSuFU=",
	}
	oldSessionJSON, _ := json.Marshal(oldSession)
	mock.ExpectGet("session:" + oldSessionID).SetVal(string(oldSessionJSON))

	newSessionID := "new-session-id"
	newSession := &types.UserSession{
		ID:            newSessionID,
		UserID:        "user-123",
		CreatedAt:     now,
		ExpiresAt:     now.Add(24 * time.Hour),
		LastRotatedAt: now,
	}
	newSessionJSON, _ := json.Marshal(newSession)

	// Expect the new session to be created
	mock.ExpectSetNX("session:"+newSessionID, newSessionJSON, 24*time.Hour).SetVal(true)
	mock.ExpectZAdd("user_sessions:user-123", redis.Z{Score: float64(now.UnixNano()), Member: newSessionID}).SetVal(1)

	// Expect the old session to be put in the grace period
	mock.ExpectSet("session:"+oldSessionID, newSessionID, 5*time.Minute).SetVal("OK")

	// Expect the old session to be removed from the user's session set
	mock.ExpectZRemRangeByRank("user_sessions:user-123", 0, -1).SetVal(1)

	rotatedSession, err := sm.RotateSession(ctx, oldSessionID, req)
	assert.NoError(t, err)
	assert.NotNil(t, rotatedSession)
}

func TestSessionManager_DeviceBinding(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := zap.NewNop()
	rc := &redisClient{client: db}
	uuidGen := &MockUUIDGenerator{uuid: "new-session-id"}
	clock := &MockClock{}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	sm := NewSessionManager(rc, SessionConfig{
		DefaultTTL:          24 * time.Hour,
		EnableDeviceBinding: true,
	}, logger, uuidGen, clock, metrics)

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", "curl/7.64.1")
	sessionID := "session-id"

	// Mock the session with a different device fingerprint
	clock.SetNow(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	session := &types.UserSession{
		ID:        sessionID,
		UserID:    "user-123",
		ExpiresAt: clock.Now().Add(24 * time.Hour),
	}
	sessionJSON, _ := json.Marshal(session)
	mock.ExpectGet("session:" + sessionID).SetVal(string(sessionJSON))

	_, err := sm.GetSession(ctx, sessionID, req)
	assert.Error(t, err)
	assert.Equal(t, types.ErrDeviceMismatch, err)
}

func TestSessionManager_EnforceMaxSessions(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := zap.NewNop()
	rc := &redisClient{client: db}
	uuidGen := &MockUUIDGenerator{uuid: "new-session-id"}
	clock := &MockClock{}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	sm := NewSessionManager(rc, SessionConfig{
		DefaultTTL:         24 * time.Hour,
		MaxSessionsPerUser: 1,
	}, logger, uuidGen, clock, metrics)

	ctx := context.Background()
	userID := "user-123"

	clock.SetNow(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	// Mock the user having one session already
	mock.ExpectZCard("user_sessions:" + userID).SetVal(1)

	// Expect the old session to be evicted
	mock.ExpectZRemRangeByRank("user_sessions:"+userID, 0, 0).SetVal(1)

	err := sm.enforceMaxSessions(ctx, userID)
	assert.NoError(t, err)
}
