package device

import (
	"time"

	"github.com/turtacn/QuantaID/internal/storage/postgresql/models"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// TrustLevel defines the level of trust for a device
type TrustLevel string

const (
	TrustLevelLow      TrustLevel = "low"
	TrustLevelMedium   TrustLevel = "medium"
	TrustLevelHigh     TrustLevel = "high"
	TrustLevelVerified TrustLevel = "verified"
)

// TrustScorer calculates trust scores for devices
type TrustScorer struct {
	config utils.DeviceTrustConfig
}

// NewTrustScorer creates a new TrustScorer
func NewTrustScorer(config utils.DeviceTrustConfig) *TrustScorer {
	// Set defaults if 0 (though config loading should handle this, defensive coding)
	if config.BaseScore == 0 { config.BaseScore = 20 }
	if config.AgeBonus == 0 { config.AgeBonus = 1 }
	if config.MaxAgeBonus == 0 { config.MaxAgeBonus = 30 }
	if config.BoundBonus == 0 { config.BoundBonus = 20 }
	if config.FrequencyBonus == 0 { config.FrequencyBonus = 10 }
	if config.VerifiedBonus == 0 { config.VerifiedBonus = 20 }

	return &TrustScorer{config: config}
}

// CalculateScore calculates the trust score for a device
func (s *TrustScorer) CalculateScore(device *models.Device) int {
	score := s.config.BaseScore

	// Age Bonus
	if !device.CreatedAt.IsZero() {
		daysOld := int(time.Since(device.CreatedAt).Hours() / 24)
		ageBonus := daysOld * s.config.AgeBonus
		if ageBonus > s.config.MaxAgeBonus {
			ageBonus = s.config.MaxAgeBonus
		}
		score += ageBonus
	}

	// Bound Bonus
	if device.UserID != "" {
		score += s.config.BoundBonus
	}

	// Verified Bonus (Bound for > 7 days)
	if device.BoundAt != nil && !device.BoundAt.IsZero() {
		daysBound := int(time.Since(*device.BoundAt).Hours() / 24)
		if daysBound > 7 {
			score += s.config.VerifiedBonus
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	// Ensure non-negative
	if score < 0 {
		score = 0
	}

	return score
}

// GetTrustLevel returns the trust level based on the score
func (s *TrustScorer) GetTrustLevel(score int) TrustLevel {
	switch {
	case score < 30:
		return TrustLevelLow
	case score < 60:
		return TrustLevelMedium
	case score < 80:
		return TrustLevelHigh
	default:
		return TrustLevelVerified
	}
}
