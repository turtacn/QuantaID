package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRedisClient_HealthCheck(t *testing.T) {
	db, mock := redismock.NewClientMock()
	rc := &redisClient{client: db}

	// Mock a successful ping
	mock.ExpectPing().SetVal("PONG")
	err := rc.HealthCheck(context.Background())
	assert.NoError(t, err)

	// Mock a failed ping
	mock.ExpectPing().SetErr(assert.AnError)
	err = rc.HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestRedisClient_ReconnectWithBackoff(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &RedisConfig{
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	}
	reg := prometheus.NewRegistry()
	metrics := NewMetrics("test", reg)
	rc := &redisClient{client: db, cfg: cfg, metrics: metrics}

	// Mock failed pings, then a successful one
	mock.ExpectPing().SetErr(assert.AnError)
	mock.ExpectPing().SetErr(assert.AnError)
	mock.ExpectPing().SetVal("PONG")

	err := rc.reconnectWithBackoff(context.Background())
	assert.NoError(t, err)
}
