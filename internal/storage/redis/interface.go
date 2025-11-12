package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientInterface defines the interface for Redis operations.
type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Close() error
	Client() *redis.Client
	HealthCheck(ctx context.Context) error
}
