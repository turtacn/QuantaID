package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/pkg/types"
)

const (
	tokenPrefix = "token:"
)

// RedisTokenRepository provides a Redis-backed implementation of the auth.TokenRepository interface.
type RedisTokenRepository struct {
	client RedisClientInterface
}

// NewRedisTokenRepository creates a new Redis token repository.
func NewRedisTokenRepository(client RedisClientInterface) *RedisTokenRepository {
	return &RedisTokenRepository{
		client: client,
	}
}

// SaveToken stores a token in Redis.
func (r *RedisTokenRepository) SaveToken(ctx context.Context, token *types.Token) error {
	key := tokenPrefix + token.AccessToken
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("could not marshal token: %w", err)
	}

	expiration := time.Duration(token.ExpiresIn) * time.Second
	if expiration <= 0 {
		expiration = time.Hour
	}

	return r.client.Set(ctx, key, data, expiration)
}

// FetchToken retrieves a token by its access token from Redis.
func (r *RedisTokenRepository) FetchToken(ctx context.Context, accessToken string) (*types.Token, error) {
	key := tokenPrefix + accessToken
	data, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, types.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch token from redis: %w", err)
	}

	var token types.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("could not unmarshal token: %w", err)
	}
	return &token, nil
}

// DeleteToken removes a token from Redis by its access token.
func (r *RedisTokenRepository) DeleteToken(ctx context.Context, accessToken string) error {
	key := tokenPrefix + accessToken
	return r.client.Del(ctx, key)
}

// StoreRefreshToken saves a refresh token to Redis with a specified duration.
func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, token string, userID string, duration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.client.SetEx(ctx, key, userID, duration).Err()
}

// GetRefreshTokenUserID retrieves the user ID associated with a refresh token from Redis.
func (r *RedisTokenRepository) GetRefreshTokenUserID(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	userID, err := r.client.Get(ctx, key)
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
	return r.client.Del(ctx, key)
}

// AddToDenyList adds a JWT ID (jti) to the deny list in Redis with a specified duration.
func (r *RedisTokenRepository) AddToDenyList(ctx context.Context, jti string, duration time.Duration) error {
	key := fmt.Sprintf("deny_list:%s", jti)
	return r.client.SetEx(ctx, key, "", duration).Err()
}

// IsInDenyList checks if a JWT ID (jti) exists in the deny list in Redis.
func (r *RedisTokenRepository) IsInDenyList(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("deny_list:%s", jti)
	val, err := r.client.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return val == 1, nil
}
