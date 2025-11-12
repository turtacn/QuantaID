package cache

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/internal/metrics"
	"github.com/turtacn/QuantaID/pkg/types"
)

// ErrNotFound is returned when an item is not found in the cache.
var ErrNotFound = errors.New("item not found in cache")

// UserRepository defines the interface for user data access, used for cache fallbacks.
type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*types.User, error)
}

// RedisCache provides a cache implementation using Redis.
type RedisCache struct {
	client   *redis.Client
	userRepo UserRepository
}

// NewRedisCache creates a new RedisCache.
func NewRedisCache(client *redis.Client, userRepo UserRepository) *RedisCache {
	return &RedisCache{
		client:   client,
		userRepo: userRepo,
	}
}

// GetUser retrieves a user from the cache. If the user is not in the cache, it falls back to the database.
func (rc *RedisCache) GetUser(ctx context.Context, userID string) (*types.User, error) {
	// 1. Try to get from cache
	cacheKey := "user:" + userID
	cached, err := rc.client.Get(ctx, cacheKey).Result()
	if err == nil {
		var user types.User
		if json.Unmarshal([]byte(cached), &user) == nil {
			metrics.CacheHitsTotal.Inc()
			return &user, nil
		}
	}
	if err != redis.Nil {
		// Log the error if it's not a cache miss
	}

	// 2. Cache miss, get from database
	metrics.CacheMissesTotal.Inc()
	user, err := rc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. Write to cache with random TTL
	data, err := json.Marshal(user)
	if err != nil {
		return user, nil // Return user, but log the marshalling error
	}

	// Random TTL to prevent cache avalanche (base TTL ± 10%)
	baseTTL := 5 * time.Minute
	randomOffset := time.Duration(rand.Intn(60)-30) * time.Second // +/- 30 seconds
	randomizedTTL := baseTTL + randomOffset
	rc.client.Set(ctx, cacheKey, data, randomizedTTL)

	return user, nil
}

// SetSession caches a session with a randomized TTL.
func (rc *RedisCache) SetSession(ctx context.Context, sessionID string, sessionData interface{}) error {
	cacheKey := "session:" + sessionID
	data, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	baseTTL := 30 * time.Minute
	randomOffset := time.Duration(rand.Intn(3*60)-90) * time.Second // +/- 1.5 minutes
	randomizedTTL := baseTTL + randomOffset
	return rc.client.Set(ctx, cacheKey, data, randomizedTTL).Err()
}

// GetSession retrieves a session from the cache.
func (rc *RedisCache) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	cacheKey := "session:" + sessionID
	cached, err := rc.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}
		return err
	}
	return json.Unmarshal([]byte(cached), dest)
}

// SetAccessToken caches an OAuth access token with a randomized TTL.
func (rc *RedisCache) SetAccessToken(ctx context.Context, accessToken string, tokenData interface{}, expiration time.Duration) error {
	cacheKey := "token:" + accessToken
	data, err := json.Marshal(tokenData)
	if err != nil {
		return err
	}
	randomOffset := time.Duration(rand.Intn(int(expiration.Seconds()/10)) - int(expiration.Seconds()/20)) * time.Second
	randomizedTTL := expiration + randomOffset
	return rc.client.Set(ctx, cacheKey, data, randomizedTTL).Err()
}

// GetAccessToken retrieves an access token from the cache.
func (rc *RedisCache) GetAccessToken(ctx context.Context, accessToken string, dest interface{}) error {
	cacheKey := "token:" + accessToken
	cached, err := rc.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}
		return err
	}
	return json.Unmarshal([]byte(cached), dest)
}
