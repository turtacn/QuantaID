package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/pkg/types"
)

// RedisTokenRepository provides a Redis-backed implementation of the auth.TokenRepository interface.
type RedisTokenRepository struct {
	client *redis.Client
}

// NewRedisTokenRepository creates a new Redis token repository.
func NewRedisTokenRepository(client *redis.Client) *RedisTokenRepository {
	return &RedisTokenRepository{
		client: client,
	}
}

// StoreRefreshToken saves a refresh token to Redis with a specified duration.
func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.SetEx(ctx, key, userID, duration).Err()
}

// GetRefreshTokenUserID retrieves the user ID associated with a refresh token from Redis.
func (r *RedisTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	userID, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", types.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("redis get: %w", err)
	}
	return userID, nil
}

// DeleteRefreshToken removes a refresh token from Redis.
func (r *RedisTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.Del(ctx, key).Err()
}

// AddToDenyList adds a JWT ID (jti) to the deny list in Redis with a specified duration.
func (r *RedisTokenRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error {
	key := fmt.Sprintf("deny_list:%s", jti)
	return r.client.SetEx(ctx, key, "", duration).Err()
}

// IsInDenyList checks if a JWT ID (jti) exists in the deny list in Redis.
func (r *RedisTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("deny_list:%s", jti)
	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return val == 1, nil
}
