package adaptive

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/storage/redis/mock"
	"go.uber.org/zap"
)

func TestRiskEngine_Evaluate(t *testing.T) {
	// Arrange
	redisClient := &mock.RedisClient{}
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
	engine := NewRiskEngine(cfg, redisClient, logger)
	ac := auth.AuthContext{
		UserID:            "user-123",
		IPAddress:         "1.2.3.4",
		DeviceFingerprint: "fingerprint-123",
		Timestamp:         time.Now(),
	}

	redisClient.On("SIsMember", context.Background(), "user:user-123:devices", "fingerprint-123").Return(redis.NewBoolResult(false, nil))
	redisClient.On("Get", context.Background(), "user:user-123:failed_logins").Return(redis.NewStringResult("", redis.Nil))

	// Act
	score, level, err := engine.Evaluate(context.Background(), ac)

	// Assert
	assert.NoError(t, err)
	assert.Greater(t, float64(score), 0.0)
	assert.Equal(t, auth.RiskLevelHigh, level)
}
