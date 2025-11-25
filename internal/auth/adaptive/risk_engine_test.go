package adaptive

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	testify_mock "github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	storage_redis "github.com/turtacn/QuantaID/internal/storage/redis"
	"github.com/turtacn/QuantaID/internal/storage/redis/mock"
	"go.uber.org/zap"
)

func TestRiskEngine_Evaluate(t *testing.T) {
	// Arrange
	redisClient := &mock.RedisClient{}
	geoManager := storage_redis.NewGeoManager(redisClient)
	logger := zap.NewNop()
	cfg := config.RiskConfig{
		Thresholds: config.RiskThresholds{
			Low:    0.3,
			Medium: 0.7,
			High:   1.0,
		},
		Weights: config.RiskWeights{
			IPReputation: 0.5,
			DeviceChange: 0.3,
		},
	}
	engine := NewRiskEngine(cfg, redisClient, geoManager, nil, logger)
	ac := auth.AuthContext{
		UserID:            "user-123",
		IPAddress:         "1.2.3.4",
		DeviceFingerprint: "fingerprint-123",
		Timestamp:         time.Now(),
	}

	redisClient.On("SIsMember", context.Background(), "user:user-123:devices", "fingerprint-123").Return(redis.NewBoolResult(false, nil))
	redisClient.On("Get", context.Background(), "user:user-123:failed_logins").Return(redis.NewStringResult("", redis.Nil))
	// Mock geo calls
	redisClient.On("ZRange", testify_mock.Anything, testify_mock.Anything, testify_mock.Anything, testify_mock.Anything).Return([]string{}, nil)
	redisClient.On("GeoAdd", testify_mock.Anything, testify_mock.Anything, testify_mock.Anything).Return(int64(1), nil)
	redisClient.On("ZAdd", testify_mock.Anything, testify_mock.Anything, testify_mock.Anything).Return(nil)
	redisClient.On("Expire", testify_mock.Anything, testify_mock.Anything, testify_mock.Anything).Return(redis.NewBoolResult(true, nil))

	// Act
	score, level, err := engine.Evaluate(context.Background(), ac)

	// Assert
	assert.NoError(t, err)
	assert.Greater(t, float64(score), 0.0)
	assert.Equal(t, auth.RiskLevelHigh, level)
}
