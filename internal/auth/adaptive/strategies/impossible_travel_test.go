package strategies_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	github_redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/adaptive/strategies"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// MockRedisClient is a mock for RedisClientInterface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Client() *github_redis.Client { return nil }
func (m *MockRedisClient) Close() error                 { return nil }
func (m *MockRedisClient) HealthCheck(ctx context.Context) error { return nil }
func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error { return nil }
func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *github_redis.BoolCmd { return nil }
func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) { return "", nil }
func (m *MockRedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) { return nil, nil }
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error { return nil }
func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error { return nil }
func (m *MockRedisClient) SCard(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) error { return nil }
func (m *MockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) { return nil, nil }
func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...github_redis.Z) error { return nil }
func (m *MockRedisClient) ZCard(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *MockRedisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) { return 0, nil }
func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) { return 0, nil }
func (m *MockRedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) { return nil, nil }
func (m *MockRedisClient) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *github_redis.StatusCmd { return nil }
func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) { return 0, nil }
func (m *MockRedisClient) SIsMember(ctx context.Context, key string, member interface{}) *github_redis.BoolCmd { return nil }

// Relevant mocks for GeoManager
func (m *MockRedisClient) HMSet(ctx context.Context, key string, values ...interface{}) *github_redis.BoolCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*github_redis.BoolCmd)
}

func (m *MockRedisClient) HGetAll(ctx context.Context, key string) *github_redis.MapStringStringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*github_redis.MapStringStringCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *github_redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*github_redis.BoolCmd)
}

func TestImpossibleTravel_CalculateRisk(t *testing.T) {
	// Beijing
	lat1, lon1 := 39.9042, 116.4074
	// Shanghai
	lat2, lon2 := 31.2304, 121.4737

	// We'll mock the internal behavior of GeoManager by mocking the Redis client it uses.
	mockRedis := new(MockRedisClient)
	geoManager := redis.NewGeoManager(mockRedis)
	strategy := strategies.NewImpossibleTravelStrategy(geoManager)

	// Scenario 1: Teleport (Impossible)
	// Same time, different location
	ctx := context.Background()
	timestamp := time.Now()

	// Setup mock return for GetLastLoginGeo
	cmd := github_redis.NewMapStringStringCmd(ctx)
	cmd.SetVal(map[string]string{
		"lat": strconv.FormatFloat(lat1, 'f', 6, 64),
		"lon": strconv.FormatFloat(lon1, 'f', 6, 64),
		"timestamp": strconv.FormatInt(timestamp.Unix(), 10),
	})

	mockRedis.On("HGetAll", ctx, "user:geo:user1").Return(cmd)

	// User at Shanghai now, was at Beijing same time
	risk, err := strategy.CalculateRisk(ctx, "user1", lat2, lon2)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, risk) // Should be high risk (teleport)

	// Scenario 2: Reasonable Travel
	// 5 hours later
	timestamp2 := timestamp.Add(-5 * time.Hour) // Last login was 5 hours ago

	cmd2 := github_redis.NewMapStringStringCmd(ctx)
	cmd2.SetVal(map[string]string{
		"lat": strconv.FormatFloat(lat1, 'f', 6, 64),
		"lon": strconv.FormatFloat(lon1, 'f', 6, 64),
		"timestamp": strconv.FormatInt(timestamp2.Unix(), 10),
	})

	mockRedis.On("HGetAll", ctx, "user:geo:user2").Return(cmd2)

	risk2, err2 := strategy.CalculateRisk(ctx, "user2", lat2, lon2)
	assert.NoError(t, err2)
	assert.Equal(t, 0.0, risk2) // Should be low risk
}
