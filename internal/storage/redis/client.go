package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client to provide a consistent interface for Redis operations.
type RedisClient struct {
	client *redis.Client
	cfg    *RedisConfig
}

// RedisConfig holds the configuration for connecting to a Redis server.
type RedisConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Password    string        `yaml:"password"`
	DB          int           `yaml:"db"`
	PoolSize    int           `yaml:"pool_size"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
}

// NewRedisClient creates a new Redis client and establishes a connection.
// It performs a health check to ensure the connection is valid.
func NewRedisClient(cfg *RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisClient{client: rdb, cfg: cfg}, nil
}

// HealthCheck performs a health check on the Redis connection.
func (rc *RedisClient) HealthCheck(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close gracefully closes the Redis connection.
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}
