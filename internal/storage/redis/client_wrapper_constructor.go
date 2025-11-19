package redis

import "github.com/redis/go-redis/v9"

// NewRedisClientWrapper creates a new RedisClientWrapper.
func NewRedisClientWrapper(client *redis.Client) *RedisClientWrapper {
	return &RedisClientWrapper{client: client}
}
