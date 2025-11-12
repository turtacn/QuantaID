package integration

import (
	"context"
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
	client, err := redis.NewRedisClient(cfg)
	assert.NoError(t, err)
	defer client.Close()

	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
}
