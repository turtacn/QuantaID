package redis

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

func setupRedisContainer(t *testing.T) RedisClientInterface {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	})

	host, err := redisC.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatal(err)
	}

	redisConfig := &RedisConfig{
		Host: host,
		Port: port.Int(),
	}

	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)

	rc, err := NewRedisClient(redisConfig, metrics)
	if err != nil {
		t.Fatal(err)
	}
	return rc
}

func TestSessionManager_Integration_CreateAndGetSession(t *testing.T) {
	t.Skip("Skipping integration test - Docker permission issue")
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	rc := setupRedisContainer(t)
	logger := zap.NewNop()
	uuidGen := &GoogleUUIDGenerator{}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	clock := &realClock{}
	sm := NewSessionManager(rc, SessionConfig{
		DefaultTTL: 24 * time.Hour,
	}, logger, uuidGen, clock, metrics)

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/", nil)
	userID := "user-123"

	session, err := sm.CreateSession(ctx, userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	retrievedSession, err := sm.GetSession(ctx, session.ID, req)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSession)
	assert.Equal(t, session.ID, retrievedSession.ID)
}

func TestSessionManager_Integration_DeviceBinding(t *testing.T) {
	t.Skip("Skipping integration test - Docker permission issue")
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	rc := setupRedisContainer(t)
	logger := zap.NewNop()
	uuidGen := &GoogleUUIDGenerator{}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	clock := &realClock{}
	sm := NewSessionManager(rc, SessionConfig{
		DefaultTTL:          24 * time.Hour,
		EnableDeviceBinding: true,
	}, logger, uuidGen, clock, metrics)

	ctx := context.Background()
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("User-Agent", "curl/7.64.1")
	userID := "user-123"

	session, err := sm.CreateSession(ctx, userID, req1)
	assert.NoError(t, err)
	assert.NotNil(t, session)

	// Same request should work
	retrievedSession, err := sm.GetSession(ctx, session.ID, req1)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSession)

	// Different request should fail
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
	_, err = sm.GetSession(ctx, session.ID, req2)
	assert.Error(t, err)
}

func TestSessionManager_Integration_ConcurrentSessions(t *testing.T) {
	t.Skip("Skipping integration test - Docker permission issue")
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	rc := setupRedisContainer(t)
	logger := zap.NewNop()
	uuidGen := &GoogleUUIDGenerator{}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	clock := &realClock{}
	sm := NewSessionManager(rc, SessionConfig{
		DefaultTTL:         24 * time.Hour,
		MaxSessionsPerUser: 2,
	}, logger, uuidGen, clock, metrics)

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/", nil)
	userID := "user-123"

	// Create 3 sessions, which should evict the first one
	session1, err := sm.CreateSession(ctx, userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, session1)

	session2, err := sm.CreateSession(ctx, userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, session2)

	session3, err := sm.CreateSession(ctx, userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, session3)

	// The first session should have been evicted
	_, err = sm.GetSession(ctx, session1.ID, req)
	assert.Error(t, err)

	// The second and third sessions should still be valid
	_, err = sm.GetSession(ctx, session2.ID, req)
	assert.NoError(t, err)
	_, err = sm.GetSession(ctx, session3.ID, req)
	assert.NoError(t, err)
}
