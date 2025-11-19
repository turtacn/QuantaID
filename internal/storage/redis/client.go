package redis

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	client  *redis.Client
	cfg     *RedisConfig
	metrics *Metrics
	stopCh  chan struct{} // Channel to signal goroutines to stop
}

// RedisConfig holds the configuration for connecting to a Redis server.
type RedisConfig struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Password    string        `yaml:"password"`
	DB          int           `yaml:"db"`
	PoolSize    int           `yaml:"pool_size"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	// Advanced options
	MinIdleConns    int           `yaml:"min_idle_conns"`
	MaxConnAge      time.Duration `yaml:"max_conn_age"`
	PoolTimeout     time.Duration `yaml:"pool_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	HealthCheck     HealthCheckConfig `yaml:"health_check"`
	Retry           RetryConfig       `yaml:"retry"`
}

// HealthCheckConfig defines the health check parameters.
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
}

// RetryConfig defines the exponential backoff parameters for reconnections.
type RetryConfig struct {
	MaxAttempts    int           `yaml:"max_attempts"`
	InitialBackoff time.Duration `yaml:"initial_backoff"`
	MaxBackoff     time.Duration `yaml:"max_backoff"`
}

// NewRedisClient creates a new Redis client and establishes a connection.
// It performs a health check and can be configured with advanced features
// like connection pooling, health checks, and exponential backoff.
func NewRedisClient(cfg *RedisConfig, metrics *Metrics) (RedisClientInterface, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		MinIdleConns:    cfg.MinIdleConns,
		ConnMaxLifetime: cfg.MaxConnAge,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	})

	rc := &redisClient{
		client: rdb,
		cfg:    cfg,
		metrics: metrics,
		stopCh: make(chan struct{}),
	}

	// Add a hook to instrument all commands
	rdb.AddHook(rc.newMetricsHook())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rc.reconnectWithBackoff(ctx); err != nil {
		return nil, fmt.Errorf("initial redis connection failed: %w", err)
	}

	if err := rc.warmUp(ctx); err != nil {
		// Log this as a warning, as the client can still function
	}

	if cfg.HealthCheck.Enabled {
		go rc.startHealthCheck()
	}

	go rc.startPoolStatsCollector()

	return rc, nil
}

// newMetricsHook creates a new redis.Hook for instrumenting commands.
func (rc *redisClient) newMetricsHook() redis.Hook {
	return &metricsHook{metrics: rc.metrics}
}

type metricsHook struct {
	metrics *Metrics
}

func (h *metricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *metricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(start)

		h.metrics.Commands.WithLabelValues(cmd.Name()).Inc()
		h.metrics.CommandLatency.WithLabelValues(cmd.Name()).Observe(duration.Seconds())

		if err != nil && err != redis.Nil {
			h.metrics.Errors.WithLabelValues(cmd.Name()).Inc()
		}

		return err
	}
}

func (h *metricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}

// startPoolStatsCollector periodically collects and updates pool statistics.
func (rc *redisClient) startPoolStatsCollector() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := rc.client.PoolStats()
			rc.metrics.PoolTotalConns.Set(float64(stats.TotalConns))
			rc.metrics.PoolIdleConns.Set(float64(stats.IdleConns))
			rc.metrics.PoolHits.Add(float64(stats.Hits))
			rc.metrics.PoolMisses.Add(float64(stats.Misses))
			rc.metrics.PoolTimeouts.Add(float64(stats.Timeouts))
			rc.metrics.PoolStaleConns.Add(float64(stats.StaleConns))
		case <-rc.stopCh:
			return
		}
	}
}


// reconnectWithBackoff attempts to connect to Redis with exponential backoff.
func (rc *redisClient) reconnectWithBackoff(ctx context.Context) error {
	backoff := rc.cfg.Retry.InitialBackoff
	if backoff == 0 {
		backoff = 1 * time.Second
	}

	maxBackoff := rc.cfg.Retry.MaxBackoff
	if maxBackoff == 0 {
		maxBackoff = 10 * time.Second
	}

	maxAttempts := rc.cfg.Retry.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 3
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := rc.client.Ping(ctx).Err(); err == nil {
			return nil // Connection successful
		}
		time.Sleep(backoff)
		backoff *= 2 // Increase backoff
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	return fmt.Errorf("max retries exceeded (%d)", maxAttempts)
}

// warmUp pre-heats the connection pool by creating MinIdleConns.
func (rc *redisClient) warmUp(ctx context.Context) error {
	if rc.cfg.MinIdleConns <= 0 {
		return nil // No pre-heating required
	}

	// In go-redis v9, the pool is warmed up automatically based on MinIdleConns.
	// We just need to check if the connections are healthy.
	for i := 0; i < rc.cfg.MinIdleConns; i++ {
		err := rc.client.Ping(ctx).Err()
		if err != nil {
			return fmt.Errorf("failed to warm up connection pool: could not ping connection %d: %w", i+1, err)
		}
	}

	return nil
}


// startHealthCheck periodically checks the Redis connection health.
func (rc *redisClient) startHealthCheck() {
	ticker := time.NewTicker(rc.cfg.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if err := rc.HealthCheck(ctx); err != nil {
				// Consider logging this error. For now, we attempt to reconnect.
				_ = rc.reconnectWithBackoff(context.Background())
			}
			cancel()
		case <-rc.stopCh:
			return
		}
	}
}


// HealthCheck performs a health check on the Redis connection.
func (rc *redisClient) HealthCheck(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Close gracefully closes the Redis connection and stops background goroutines.
func (rc *redisClient) Close() error {
	close(rc.stopCh) // Signal health check goroutine to stop
	return rc.client.Close()
}

// Client returns the underlying go-redis client.
func (rc *redisClient) Client() *redis.Client {
	return rc.client
}

// Set stores a value in Redis with an expiration.
func (rc *redisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// SetNX stores a value in Redis with an expiration if the key does not exist.
func (rc *redisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return rc.client.SetNX(ctx, key, value, expiration)
}

// Get retrieves a value from Redis.
func (rc *redisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Del deletes a value from Redis.
func (rc *redisClient) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// SAdd adds one or more members to a set.
func (rc *redisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return rc.client.SAdd(ctx, key, members...).Err()
}

// SCard gets the number of members in a set.
func (rc *redisClient) SCard(ctx context.Context, key string) (int64, error) {
	return rc.client.SCard(ctx, key).Result()
}

// SRem removes one or more members from a set.
func (rc *redisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return rc.client.SRem(ctx, key, members...).Err()
}

// SMembers returns all members of the set value stored at key.
func (rc *redisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return rc.client.SMembers(ctx, key).Result()
}

// ZAdd adds one or more members to a sorted set, or updates its score if it already exists.
func (rc *redisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return rc.client.ZAdd(ctx, key, members...).Err()
}

// ZCard gets the number of members in a sorted set.
func (rc *redisClient) ZCard(ctx context.Context, key string) (int64, error) {
	return rc.client.ZCard(ctx, key).Result()
}

// ZRemRangeByRank removes all members in a sorted set within the given rank range.
func (rc *redisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) {
	return rc.client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZRem removes one or more members from a sorted set.
func (rc *redisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return rc.client.ZRem(ctx, key, members...).Result()
}

// ZRange returns the specified range of elements in the sorted set stored at key.
func (rc *redisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.ZRange(ctx, key, start, stop).Result()
}

// SetEx sets a key with an expiration.
func (rc *redisClient) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return rc.client.SetEx(ctx, key, value, expiration)
}

// Exists checks if a key exists.
func (rc *redisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return rc.client.Exists(ctx, keys...).Result()
}
