package e2e_test

import (
	"context"
	"net"
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
	// Add internal storage to simulate persistence
	data    map[string]map[string]string
	geoData map[string]map[string]*github_redis.GeoLocation // key -> member -> location
}

// Helper to initialize the map
func NewMockRedisE2E() *MockRedisClientE2E {
	return &MockRedisClientE2E{
		data:    make(map[string]map[string]string),
		geoData: make(map[string]map[string]*github_redis.GeoLocation),
	}
}


func (m *MockRedisClientE2E) HMSet(ctx context.Context, key string, values ...interface{}) *github_redis.BoolCmd {
	if m.data[key] == nil {
		m.data[key] = make(map[string]string)
	}
	for i := 0; i < len(values); i += 2 {
		m.data[key][values[i].(string)] = values[i+1].(string)
	}
	return github_redis.NewBoolResult(true, nil)
}

func (m *MockRedisClientE2E) HGetAll(ctx context.Context, key string) *github_redis.MapStringStringCmd {
	if data, ok := m.data[key]; ok {
		return github_redis.NewMapStringStringResult(data, nil)
	}
	return github_redis.NewMapStringStringResult(map[string]string{}, nil)
}

func (m *MockRedisClientE2E) Expire(ctx context.Context, key string, expiration time.Duration) *github_redis.BoolCmd {
	return github_redis.NewBoolResult(true, nil)
}

func (m *MockRedisClientE2E) SIsMember(ctx context.Context, key string, member interface{}) *github_redis.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*github_redis.BoolCmd)
}
func (m *MockRedisClientE2E) GeoAdd(ctx context.Context, key string, geoLocation ...*github_redis.GeoLocation) (int64, error) {
	if m.geoData[key] == nil {
		m.geoData[key] = make(map[string]*github_redis.GeoLocation)
	}
	for _, gl := range geoLocation {
		m.geoData[key][gl.Name] = gl
	}
	return int64(len(geoLocation)), nil
}

func (m *MockRedisClientE2E) GeoPos(ctx context.Context, key string, members ...string) ([]*github_redis.GeoPos, error) {
	var positions []*github_redis.GeoPos
	if memberMap, ok := m.geoData[key]; ok {
		for _, member := range members {
			if loc, exists := memberMap[member]; exists {
				positions = append(positions, &github_redis.GeoPos{
					Longitude: loc.Longitude,
					Latitude:  loc.Latitude,
				})
			} else {
				positions = append(positions, nil)
			}
		}
	} else {
		for range members {
			positions = append(positions, nil)
		}
	}
	return positions, nil
}

func (m *MockRedisClientE2E) Client() *github_redis.Client { return nil }
func (m *MockRedisClientE2E) Close() error                 { return nil }
func (m *MockRedisClientE2E) HealthCheck(ctx context.Context) error { return nil }
func (m *MockRedisClientE2E) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.data[key] == nil {
		m.data[key] = make(map[string]string)
	}
	m.data[key]["value"] = value.(string)
	return nil
}
func (m *MockRedisClientE2E) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *github_redis.BoolCmd { return nil }
func (m *MockRedisClientE2E) Get(ctx context.Context, key string) (string, error) {
	if data, ok := m.data[key]; ok {
		return data["value"], nil
	}
	return "", github_redis.Nil
}
func (m *MockRedisClientE2E) MGet(ctx context.Context, keys ...string) ([]interface{}, error) { return nil, nil }
func (m *MockRedisClientE2E) Del(ctx context.Context, keys ...string) error { return nil }
func (m *MockRedisClientE2E) SAdd(ctx context.Context, key string, members ...interface{}) error { return nil }
func (m *MockRedisClientE2E) SCard(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) SRem(ctx context.Context, key string, members ...interface{}) error { return nil }
func (m *MockRedisClientE2E) SMembers(ctx context.Context, key string) ([]string, error) { return nil, nil }
func (m *MockRedisClientE2E) ZAdd(ctx context.Context, key string, members ...github_redis.Z) error {
	if m.data[key] == nil {
		m.data[key] = make(map[string]string)
	}
	for _, member := range members {
		m.data[key][member.Member.(string)] = ""
	}
	return nil
}
func (m *MockRedisClientE2E) ZCard(ctx context.Context, key string) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) { return 0, nil }
func (m *MockRedisClientE2E) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if data, ok := m.data[key]; ok {
		// a bit of a hack for this test, we know we only have one item
		for k := range data {
			return []string{k}, nil
		}
	}
	return nil, nil
}
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
	mockRedis := NewMockRedisE2E()
	geoManager := redis.NewGeoManager(mockRedis)
	mockGeoIP := &MockGeoIP{}
	logger := zap.NewNop()

	cfg := config.RiskConfig{
		Thresholds: config.RiskThresholds{Low: 0.3, Medium: 0.7, High: 0.9},
		Weights: config.RiskWeights{
			GeoVelocity:  0.5,
			IPReputation: 0.2,
			DeviceChange: 0.3,
		},
	}

	engine := adaptive.NewRiskEngine(cfg, mockRedis, geoManager, mockGeoIP, logger)
	ctx := context.Background()

	mockRedis.On("SIsMember", ctx, mock.Anything, mock.Anything).Return(github_redis.NewBoolResult(false, nil))

	// Step 1: First login (Beijing)
	ac1 := auth.AuthContext{
		UserID:    "user_e2e",
		IPAddress: "1.2.3.4",
		Timestamp: time.Now(),
	}

	_, level1, err := engine.Evaluate(ctx, ac1)
	assert.NoError(t, err)
	assert.Equal(t, auth.RiskLevelMedium, level1)

	// Step 2: Second login (Shanghai) - IMPOSSIBLE TRAVEL
	ac2 := auth.AuthContext{
		UserID:    "user_e2e",
		IPAddress: "5.6.7.8",
		Timestamp: time.Now().Add(10 * time.Minute), // 10 minutes later, impossible for 1200km
	}

	score2, level2, err := engine.Evaluate(ctx, ac2)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, float64(score2), 0.7)
	assert.Equal(t, auth.RiskLevelHigh, level2)
}
