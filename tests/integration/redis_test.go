package integration

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

func TestRedisClient(t *testing.T) {
	// This test requires a running Redis instance
	t.Skip("Skipping Redis integration test")

	cfg := &redis.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}
	metrics := redis.NewMetrics("test", prometheus.NewRegistry())
	client, err := redis.NewRedisClient(cfg, metrics)
	assert.NoError(t, err)
	defer client.Close()

	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
}
