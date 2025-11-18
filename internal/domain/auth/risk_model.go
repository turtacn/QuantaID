package auth

import (
	"context"
	"time"

	"github.com/turtacn/QuantaID/internal/config"
)

// AuthContext contains the contextual information for an authentication attempt.
type AuthContext struct {
	UserID            string
	IPAddress         string
	UserAgent         string
	AcceptLanguage    string
	Timestamp         time.Time
	DeviceFingerprint string
	IsKnownDevice     bool
}

// RiskEngine evaluates the risk of an authentication attempt.
type RiskEngine interface {
	Evaluate(ctx context.Context, ac AuthContext) (RiskScore, RiskLevel, error)
}

// RiskScore is a float64 representing the calculated risk, constrained to a range of 0.0 to 1.0.
type RiskScore float64

// RiskLevel represents the qualitative risk level determined from a RiskScore.
type RiskLevel string

const (
	// RiskLevelLow indicates a low-risk activity, typically requiring no additional verification.
	RiskLevelLow RiskLevel = "LOW"
	// RiskLevelMedium indicates a medium-risk activity, often triggering standard MFA.
	RiskLevelMedium RiskLevel = "MEDIUM"
	// RiskLevelHigh indicates a high-risk activity, which may require strong MFA or be blocked.
	RiskLevelHigh RiskLevel = "HIGH"
)

// RiskFactors holds the raw signals collected during an authentication attempt, used for risk calculation.
type RiskFactors struct {
	IPReputation   float64
	GeoReputation  float64
	DeviceTrusted  bool
	TimeWindow     time.Time
	LoginHistory   []time.Time
	IsKnownDevice  bool
	GeoVelocity    float64
	UserAgent      string
	IPAddress      string
	AcceptLanguage string
}

// ToScore calculates the final risk score based on the factor weights defined in the configuration.
// It ensures the final score is normalized between 0.0 and 1.0.
func (f *RiskFactors) ToScore(cfg config.RiskConfig) RiskScore {
	// This is a simplified calculation. A real implementation would involve more sophisticated scoring logic.
	score := f.IPReputation*cfg.Weights.IPReputation +
		f.GeoReputation*cfg.Weights.GeoReputation +
		f.GeoVelocity*cfg.Weights.GeoVelocity

	if !f.IsKnownDevice {
		score += cfg.Weights.DeviceChange
	}

	// Normalize score to be within 0.0 - 1.0
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return RiskScore(score)
}

// Level determines the RiskLevel based on the score and the configured thresholds.
func (s RiskScore) Level(thresholds config.RiskThresholds) RiskLevel {
	if float64(s) <= thresholds.Low {
		return RiskLevelLow
	}
	if float64(s) <= thresholds.Medium {
		return RiskLevelMedium
	}
	return RiskLevelHigh
}
