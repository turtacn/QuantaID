package profile

import (
	"github.com/turtacn/QuantaID/pkg/utils"
)

// QualityScorer assesses the quality and completeness of user profiles
type QualityScorer struct {
	weights utils.QualityWeightsConfig
}

// NewQualityScorer creates a new QualityScorer
func NewQualityScorer(weights utils.QualityWeightsConfig) *QualityScorer {
	return &QualityScorer{weights: weights}
}

// CalculateScore calculates the quality score based on details
func (s *QualityScorer) CalculateScore(details QualityDetails) int {
	score := 0

	if details.HasEmail {
		score += s.weights.Email
	}
	if details.EmailVerified {
		score += s.weights.EmailVerified
	}
	if details.HasPhone {
		score += s.weights.Phone
	}
	if details.PhoneVerified {
		score += s.weights.PhoneVerified
	}
	if details.HasMFA {
		score += s.weights.MFA
	}
	if details.HasRecoveryEmail {
		score += s.weights.RecoveryEmail
	}

	// Add profile completion bonus
	score += int(details.ProfileComplete * float64(s.weights.ProfileComplete))

	if score > 100 {
		score = 100
	}
	return score
}

// GetImprovement suggestions returns a list of suggestions to improve profile quality
func (s *QualityScorer) GetImprovementSuggestions(details QualityDetails) []string {
	var suggestions []string

	if !details.HasEmail {
		suggestions = append(suggestions, "Add an email address")
	} else if !details.EmailVerified {
		suggestions = append(suggestions, "Verify your email address")
	}

	if !details.HasPhone {
		suggestions = append(suggestions, "Add a phone number")
	} else if !details.PhoneVerified {
		suggestions = append(suggestions, "Verify your phone number")
	}

	if !details.HasMFA {
		suggestions = append(suggestions, "Enable Multi-Factor Authentication (MFA)")
	}

	if !details.HasRecoveryEmail {
		suggestions = append(suggestions, "Add a recovery email address")
	}

	return suggestions
}
