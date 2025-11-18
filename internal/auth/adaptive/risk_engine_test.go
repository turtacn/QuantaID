package adaptive

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/config"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"go.uber.org/zap"
)

func newTestRiskEngine() *RiskEngine {
	cfg := config.RiskConfig{
		Thresholds: config.RiskThresholds{
			Low:    0.3,
			Medium: 0.7,
		},
		Weights: config.RiskWeights{
			IPReputation: 0.5,
			GeoReputation: 0.2,
			DeviceChange: 0.3,
			GeoVelocity:  0.1,
		},
	}
	return NewRiskEngine(cfg, zap.NewNop())
}

func TestEvaluate_LowRiskScenario(t *testing.T) {
	engine := newTestRiskEngine()
	ac := auth.AuthContext{
		IPAddress:     "8.8.8.8", // Known good IP
		Timestamp:     time.Now(),
		IsKnownDevice: true,
	}

	score, level, err := engine.Evaluate(context.Background(), ac)

	assert.NoError(t, err)
	assert.Equal(t, auth.RiskLevelLow, level)
	assert.Less(t, float64(score), 0.3)
}

func TestEvaluate_MediumRiskScenario(t *testing.T) {
	engine := newTestRiskEngine()
	ac := auth.AuthContext{
		IPAddress:     "192.168.1.100", // Neutral IP
		Timestamp:     time.Now(),
		IsKnownDevice: false,
	}

	score, level, err := engine.Evaluate(context.Background(), ac)

	assert.NoError(t, err)
	assert.Equal(t, auth.RiskLevelMedium, level)
	assert.GreaterOrEqual(t, float64(score), 0.3)
	assert.LessOrEqual(t, float64(score), 0.7)
}

func TestEvaluate_HighRiskScenario(t *testing.T) {
	engine := newTestRiskEngine()
	ac := auth.AuthContext{
		IPAddress:     "1.2.3.4", // Known bad IP
		Timestamp:     time.Now(),
		IsKnownDevice: false,
	}

	score, level, err := engine.Evaluate(context.Background(), ac)

	assert.NoError(t, err)
	assert.Equal(t, auth.RiskLevelHigh, level)
	assert.Greater(t, float64(score), 0.7)
}
