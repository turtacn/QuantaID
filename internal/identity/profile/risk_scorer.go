package profile

import (
	"context"
	"math"
	"time"

	"github.com/turtacn/QuantaID/pkg/utils"
)

// IPReputationChecker defines the interface for checking IP reputation
type IPReputationChecker interface {
	GetReputation(ctx context.Context, ip string) (int, error) // Returns score 0-100 (100 is best)
}

// RiskScorer calculates risk scores based on risk indicators
type RiskScorer struct {
	config    utils.RiskScorerConfig
	ipChecker IPReputationChecker
}

// NewRiskScorer creates a new RiskScorer
func NewRiskScorer(config utils.RiskScorerConfig, ipChecker IPReputationChecker) *RiskScorer {
	return &RiskScorer{
		config:    config,
		ipChecker: ipChecker,
	}
}

// CalculateScore calculates the risk score and determines the risk level
func (s *RiskScorer) CalculateScore(indicators RiskIndicators) (int, string) {
	score := 0.0
	score += float64(indicators.AnomalyCount) * s.config.AnomalyWeight
	score += float64(indicators.GeoJumpCount) * s.config.GeoJumpWeight
	score += float64(indicators.FailedMFACount) * s.config.FailedMFAWeight
	score += float64(indicators.SuspiciousIPCount) * s.config.SuspiciousIPWeight
	score += float64(indicators.NewDeviceCount30d) * s.config.NewDeviceWeight

	// Ensure config has non-zero weights if not set (fallback)
	// Or trust caller. In test, we provided a config struct with some zeros.
	// But in `TestRiskScorer_HighRisk_MultipleAnomalies`:
	// AnomalyWeight: 15, GeoJumpWeight: 20, FailedMFAWeight: 10
	// 5 * 15 + 3 * 20 + 2 * 10 = 75 + 60 + 20 = 155 -> capped at 100

	// Time decay
	if indicators.LastAnomalyAt != nil {
		daysSince := time.Since(*indicators.LastAnomalyAt).Hours() / 24.0
		if daysSince > 0 && s.config.DecayDays > 0 {
			decayFactor := math.Pow(s.config.DecayRate, daysSince/float64(s.config.DecayDays))
			score *= decayFactor
		}
	}

	finalScore := int(score)
	if finalScore > 100 {
		finalScore = 100
	}
	if finalScore < 0 {
		finalScore = 0
	}

	return finalScore, s.ScoreToLevel(finalScore)
}

// ScoreToLevel converts a numeric score to a risk level string
func (s *RiskScorer) ScoreToLevel(score int) string {
	switch {
	case score < 25:
		return "low"
	case score < 50:
		return "medium"
	case score < 75:
		return "high"
	default:
		return "critical"
	}
}

// AnomalyEvent represents an anomaly detected in the system
type AnomalyEvent struct {
	Type      string
	Timestamp time.Time
	Details   map[string]interface{}
}

// UpdateFromEvent updates risk indicators based on a new anomaly event
func (s *RiskScorer) UpdateFromEvent(ctx context.Context, profile *UserProfile, event AnomalyEvent) RiskIndicators {
	indicators := profile.Risk

	switch event.Type {
	case "geo_jump":
		indicators.GeoJumpCount++
	case "fingerprint_change":
		indicators.AnomalyCount++
	case "suspicious_ip":
		indicators.SuspiciousIPCount++
	case "mfa_failure":
		indicators.FailedMFACount++
	case "new_device":
		indicators.NewDeviceCount30d++
	default:
		indicators.AnomalyCount++
	}

	indicators.LastAnomalyAt = &event.Timestamp
	return indicators
}

// CheckIP checks the reputation of an IP address
func (s *RiskScorer) CheckIP(ctx context.Context, ip string) (bool, error) {
	if s.ipChecker == nil {
		return false, nil // Assume safe if no checker configured
	}
	score, err := s.ipChecker.GetReputation(ctx, ip)
	if err != nil {
		return false, err
	}
	// Assuming score < 30 is suspicious
	return score < 30, nil
}
