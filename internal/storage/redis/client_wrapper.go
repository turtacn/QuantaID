package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientWrapper adapts the concrete *redis.Client to the RedisClientInterface,
// allowing the real client to be used in tests or application code that expects the interface.
type RedisClientWrapper struct {
	client *redis.Client
}

// Client returns the underlying go-redis client.
func (w *RedisClientWrapper) Client() *redis.Client {
	return w.client
}

// Set stores a value in Redis with an expiration.
func (w *RedisClientWrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return w.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value from Redis.
func (w *RedisClientWrapper) Get(ctx context.Context, key string) (string, error) {
	return w.client.Get(ctx, key).Result()
}

// MGet retrieves multiple values from Redis.
func (w *RedisClientWrapper) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return w.client.MGet(ctx, keys...).Result()
}

// Del deletes a value from Redis.
func (w *RedisClientWrapper) Del(ctx context.Context, keys ...string) error {
	return w.client.Del(ctx, keys...).Err()
}

// SAdd adds one or more members to a set.
func (w *RedisClientWrapper) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return w.client.SAdd(ctx, key, members...).Err()
}

// SCard gets the number of members in a set.
func (w *RedisClientWrapper) SCard(ctx context.Context, key string) (int64, error) {
	return w.client.SCard(ctx, key).Result()
}

// SRem removes one or more members from a set.
func (w *RedisClientWrapper) SRem(ctx context.Context, key string, members ...interface{}) error {
	return w.client.SRem(ctx, key, members...).Err()
}

// SMembers returns all members of the set value stored at key.
func (w *RedisClientWrapper) SMembers(ctx context.Context, key string) ([]string, error) {
	return w.client.SMembers(ctx, key).Result()
}

// ZAdd adds one or more members to a sorted set, or updates its score if it already exists.
func (w *RedisClientWrapper) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return w.client.ZAdd(ctx, key, members...).Err()
}

// ZCard gets the number of members in a sorted set.
func (w *RedisClientWrapper) ZCard(ctx context.Context, key string) (int64, error) {
	return w.client.ZCard(ctx, key).Result()
}

// ZRemRangeByRank removes all members in a sorted set within the given rank range.
func (w *RedisClientWrapper) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return w.client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZRem removes one or more members from a sorted set.
func (w *RedisClientWrapper) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return w.client.ZRem(ctx, key, members...).Result()
}

// ZRange returns the specified range of elements in the sorted set stored at key.
func (w *RedisClientWrapper) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return w.client.ZRange(ctx, key, start, stop).Result()
}

// SetEx sets a key with an expiration.
func (w *RedisClientWrapper) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return w.client.SetEx(ctx, key, value, expiration)
}

// Exists checks if a key exists.
func (w *RedisClientWrapper) Exists(ctx context.Context, keys ...string) (int64, error) {
	return w.client.Exists(ctx, keys...).Result()
}

// Close closes the Redis client.
func (w *RedisClientWrapper) Close() error {
	return w.client.Close()
}

// HealthCheck checks the health of the Redis client.
func (w *RedisClientWrapper) HealthCheck(ctx context.Context) error {
	return w.client.Ping(ctx).Err()
}

// SetNX sets a key if it does not already exist.
func (w *RedisClientWrapper) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return w.client.SetNX(ctx, key, value, expiration)
}
