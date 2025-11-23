package strategies

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/storage/redis"
	goredis "github.com/redis/go-redis/v9"
)

// MockGeoManager is a mock for redis.GeoManager
// Since GeoManager is a struct, we should ideally interface it or mock the underlying Redis client.
// However, ImpossibleTravelStrategy depends on *redis.GeoManager struct.
// For this unit test, we can mock the behavior by mocking the underlying RedisClientInterface used by GeoManager.
// BUT, GeoManager logic is now complex (GeoAdd + ZAdd).
// It's better to verify the CALCULATION logic in ImpossibleTravelStrategy by mocking the return of GetLastLoginLocation.
// Since GetLastLoginLocation is a method on the struct, we can't easily mock it unless we make an interface for GeoManager.
//
// Refactoring Strategy to use an interface would be best practice.
// But based on the task, we are testing the Strategy.
// Let's create a "Testable" version or use the MockRedisClient to drive the GeoManager.

func TestSpeedCalculation(t *testing.T) {
	// Setup
	mockRedis := new(redis.MockRedisClient)
	geoManager := redis.NewGeoManager(mockRedis)
	strategy := NewImpossibleTravelStrategy(geoManager)

	userID := "user123"
	now := time.Now()

	// Case 1: Beijing -> Shanghai (approx 1200km) in 1 hour
	// Beijing: 39.9042, 116.4074
	// Shanghai: 31.2304, 121.4737
	// Expected: High Risk (> 800km/h)

	// We need to mock what GeoManager calls.
	// GeoManager.GetLastLoginLocation calls:
	// 1. ZRange(timelineKey, -1, -1) -> ["timestamp"]
	// 2. GeoPos(geoKey, "timestamp") -> [pos]

	lastTime := now.Add(-1 * time.Hour)
	lastTsStr := fmt.Sprintf("%d", lastTime.Unix())

	// Mock ZRange
	mockRedis.On("ZRange", context.Background(), "user:geo:timeline:"+userID, int64(-1), int64(-1)).
		Return([]string{lastTsStr}, nil)

	// Mock GeoPos
	mockRedis.On("GeoPos", context.Background(), "user:geo:"+userID, []string{lastTsStr}).
		Return([]*goredis.GeoPos{
			{Latitude: 39.9042, Longitude: 116.4074},
		}, nil)

	risk, err := strategy.CalculateRisk(context.Background(), userID, 31.2304, 121.4737) // Shanghai
	assert.NoError(t, err)
	assert.Equal(t, 1.0, risk, "Speed 1200km/h should be high risk")

	// Case 2: Beijing -> Shanghai in 5 hours
	// Speed: 240 km/h -> Low Risk

	// Reset mocks or use new ones? The mock object records calls.
	// Easier to make a new setup for cleaner test.
}

func TestSpeedCalculation_LowRisk(t *testing.T) {
	mockRedis := new(redis.MockRedisClient)
	geoManager := redis.NewGeoManager(mockRedis)
	strategy := NewImpossibleTravelStrategy(geoManager)

	userID := "user123"
	now := time.Now()

	lastTime := now.Add(-5 * time.Hour)
	lastTsStr := fmt.Sprintf("%d", lastTime.Unix())

	mockRedis.On("ZRange", context.Background(), "user:geo:timeline:"+userID, int64(-1), int64(-1)).
		Return([]string{lastTsStr}, nil)

	mockRedis.On("GeoPos", context.Background(), "user:geo:"+userID, []string{lastTsStr}).
		Return([]*goredis.GeoPos{
			{Latitude: 39.9042, Longitude: 116.4074},
		}, nil)

	risk, err := strategy.CalculateRisk(context.Background(), userID, 31.2304, 121.4737) // Shanghai
	assert.NoError(t, err)
	assert.Equal(t, 0.0, risk, "Speed 240km/h should be low risk")
}

func TestCalculateRisk_NoHistory(t *testing.T) {
	mockRedis := new(redis.MockRedisClient)
	geoManager := redis.NewGeoManager(mockRedis)
	strategy := NewImpossibleTravelStrategy(geoManager)

	userID := "newuser"

	mockRedis.On("ZRange", context.Background(), "user:geo:timeline:"+userID, int64(-1), int64(-1)).
		Return([]string{}, nil) // No history found

	risk, err := strategy.CalculateRisk(context.Background(), userID, 31.2304, 121.4737)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, risk, "No history should result in 0 risk")
}
