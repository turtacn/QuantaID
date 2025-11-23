package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientInterface defines the interface for a Redis client.
type RedisClientInterface interface {
	Client() *redis.Client
	Close() error
	HealthCheck(ctx context.Context) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) (string, error)
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	Del(ctx context.Context, keys ...string) error
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SCard(ctx context.Context, key string) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	ZAdd(ctx context.Context, key string, members ...redis.Z) error
	ZCard(ctx context.Context, key string) (int64, error)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Exists(ctx context.Context, keys ...string) (int64, error)
	SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd

	// Added for GeoManager
	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

// UUIDGenerator generates a new UUID.
type UUIDGenerator interface {
	New() string
}
