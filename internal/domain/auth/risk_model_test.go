package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/config"
)

func TestRiskScore_LevelsMapping(t *testing.T) {
	thresholds := config.RiskThresholds{
		Low:    0.3,
		Medium: 0.7,
		High:   1.0,
	}

	testCases := []struct {
		name     string
		score    RiskScore
		expected RiskLevel
	}{
		{"LowRiskScore", 0.2, RiskLevelLow},
		{"MediumRiskScore", 0.5, RiskLevelMedium},
		{"HighRiskScore", 0.8, RiskLevelHigh},
		{"BoundaryLowToMedium", 0.3, RiskLevelLow},
		{"BoundaryMediumToHigh", 0.7, RiskLevelMedium},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := tc.score.Level(thresholds)
			assert.Equal(t, tc.expected, level)
		})
	}
}

func TestRiskFactors_ToScore(t *testing.T) {
	cfg := config.RiskConfig{
		Weights: config.RiskWeights{
			IPReputation: 0.4,
			GeoReputation: 0.3,
			DeviceChange: 0.2,
			GeoVelocity:  0.1,
		},
	}

	testCases := []struct {
		name     string
		factors  RiskFactors
		expected RiskScore
	}{
		{
			name: "KnownDeviceLowRisk",
			factors: RiskFactors{
				IPReputation:  0.1,
				GeoReputation: 0.1,
				IsKnownDevice: true,
				GeoVelocity:   0.1,
			},
			expected: 0.08,
		},
		{
			name: "NewDeviceMediumRisk",
			factors: RiskFactors{
				IPReputation:  0.5,
				GeoReputation: 0.5,
				IsKnownDevice: false,
				GeoVelocity:   0.5,
			},
			expected: 0.6,
		},
		{
			name: "HighRiskScoreCappedAtOne",
			factors: RiskFactors{
				IPReputation:  1.0,
				GeoReputation: 1.0,
				IsKnownDevice: false,
				GeoVelocity:   1.0,
			},
			expected: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := tc.factors.ToScore(cfg)
			assert.InDelta(t, float64(tc.expected), float64(score), 0.01)
		})
	}
}
