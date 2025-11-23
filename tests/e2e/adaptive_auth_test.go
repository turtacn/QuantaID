package e2e_test

import (
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/oschwald/geoip2-golang"
	github_redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/auth/adaptive"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	"go.uber.org/zap"
)

// Simplified Mock Redis Client for E2E
type MockRedisClientE2E struct {
	mock.Mock
}

func (m *MockRedisClientE2E) HMSet(ctx context.Context, key string, values ...interface{}) *github_redis.BoolCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*github_redis.BoolCmd)
}

func (m *MockRedisClientE2E) HGetAll(ctx context.Context, key string) *github_redis.MapStringStringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*github_redis.MapStringStringCmd)
}

func (m *MockRedisClientE2E) Expire(ctx context.Context, key string, expiration time.Duration) *github_redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*github_redis.BoolCmd)
}

func (m *MockRedisClientE2E) SIsMember(ctx context.Context, key string, member interface{}) *github_redis.BoolCmd {
	return github_redis.NewBoolResult(false, nil)
}

func (m *MockRedisClientE2E) Client() *github_redis.Client { return nil }
func (m *MockRedisClientE2E) Close() error                 { return nil }
func (m *MockRedisClientE2E) HealthCheck(ctx context.Context) error { return nil }
func (m *MockRedisClientE2E) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error { return nil }
func (m *MockRedisClientE2E) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *github_redis.BoolCmd { return nil }
func (m *MockRedisClientE2E) Get(ctx context.Context, key string) (string, error) { return "", nil }
func (m *MockRedisClientE2E) MGet(ctx context.Context, keys ...string) ([]interface{}, error) { return nil, nil }
func (m *MockRedisClientE2E) Del(ctx context.Context, keys ...string) error { return nil }
func (m *MockRedisClientE2E) SAdd(ctx context.Context, key string, members ...interface{}) error { return nil }
func (m *MockRedisClientE2E) SCard(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) SRem(ctx context.Context, key string, members ...interface{}) error { return nil }
func (m *MockRedisClientE2E) SMembers(ctx context.Context, key string) ([]string, error) { return nil, nil }
func (m *MockRedisClientE2E) ZAdd(ctx context.Context, key string, members ...github_redis.Z) error { return nil }
func (m *MockRedisClientE2E) ZCard(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) { return nil, nil }
func (m *MockRedisClientE2E) SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *github_redis.StatusCmd { return nil }
func (m *MockRedisClientE2E) Exists(ctx context.Context, keys ...string) (int64, error) { return 0, nil }


// Mock GeoIP
type MockGeoIP struct{}

func (m *MockGeoIP) City(ip net.IP) (*geoip2.City, error) {
	city := &geoip2.City{}
	if strings.HasPrefix(ip.String(), "1.2.3") {
		// Use direct assignment to avoid struct literal type mismatch
		city.Location.Latitude = 39.9
		city.Location.Longitude = 116.4
		return city, nil
	}
	if strings.HasPrefix(ip.String(), "5.6.7") {
		city.Location.Latitude = 31.2
		city.Location.Longitude = 121.5
		return city, nil
	}
	return nil, nil
}
func (m *MockGeoIP) Close() error { return nil }

func TestRiskEngine_ImpossibleTravel_Flow(t *testing.T) {
	mockRedis := new(MockRedisClientE2E)
	geoManager := redis.NewGeoManager(mockRedis)
	mockGeoIP := &MockGeoIP{}
	logger := zap.NewNop()

	cfg := config.RiskConfig{
		Thresholds: config.RiskThresholds{Low: 0.3, Medium: 0.7, High: 0.9},
		Weights: config.RiskWeights{
			GeoVelocity:  0.8,
			IPReputation: 0.1,
			DeviceChange: 0.1,
		},
	}

	engine := adaptive.NewRiskEngine(cfg, mockRedis, geoManager, mockGeoIP, logger)
	ctx := context.Background()

	// Step 1: First login (Beijing)
	// Expect HGetAll (empty), HMSet (save), Expire
	cmdEmpty := github_redis.NewMapStringStringCmd(ctx)
	// Return empty map
	cmdEmpty.SetVal(map[string]string{})

	mockRedis.On("HGetAll", ctx, "user:geo:user_e2e").Return(cmdEmpty).Once()
	mockRedis.On("HMSet", ctx, "user:geo:user_e2e", mock.Anything).Return(github_redis.NewBoolResult(true, nil)).Once()
	mockRedis.On("Expire", ctx, "user:geo:user_e2e", mock.Anything).Return(github_redis.NewBoolResult(true, nil)).Once()

	ac1 := auth.AuthContext{
		UserID:    "user_e2e",
		IPAddress: "1.2.3.4",
		Timestamp: time.Now(),
	}

	_, level1, err := engine.Evaluate(ctx, ac1)
	assert.NoError(t, err)
	assert.Equal(t, auth.RiskLevelLow, level1)

	// Step 2: Second login (Shanghai) - IMPOSSIBLE TRAVEL
	// Expect HGetAll (return Beijing), HMSet (save Shanghai), Expire

	// We cheat a bit by defining the return value of HGetAll for the second call
	cmdBeijing := github_redis.NewMapStringStringCmd(ctx)
	cmdBeijing.SetVal(map[string]string{
		"lat":       "39.9",
		"lon":       "116.4",
		"timestamp": strconv.FormatInt(ac1.Timestamp.Unix(), 10),
	})

	mockRedis.On("HGetAll", ctx, "user:geo:user_e2e").Return(cmdBeijing).Once()
	mockRedis.On("HMSet", ctx, "user:geo:user_e2e", mock.Anything).Return(github_redis.NewBoolResult(true, nil)).Once()
	mockRedis.On("Expire", ctx, "user:geo:user_e2e", mock.Anything).Return(github_redis.NewBoolResult(true, nil)).Once()

	ac2 := auth.AuthContext{
		UserID:    "user_e2e",
		IPAddress: "5.6.7.8",
		Timestamp: time.Now().Add(10 * time.Minute), // 10 minutes later, impossible for 1200km
	}

	score2, level2, err := engine.Evaluate(ctx, ac2)
	assert.NoError(t, err)

	// Risk Calculation:
	// Travel: 1.0 (impossible) * 0.8 = 0.8
	// IP: 0.4 (default neutral) * 0.1 = 0.04
	// Device: 1.0 (unknown) * 0.1 = 0.1
	// Total ~= 0.94 -> High Risk

	assert.GreaterOrEqual(t, float64(score2), 0.7)
	assert.Equal(t, auth.RiskLevelHigh, level2)
}
