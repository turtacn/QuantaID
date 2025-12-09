package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/turtacn/QuantaID/internal/identity/profile"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// === RiskScorer Tests ===

func TestRiskScorer_HighRisk_MultipleAnomalies(t *testing.T) {
	now := time.Now()
	indicators := profile.RiskIndicators{
		AnomalyCount:   5,
		GeoJumpCount:   3,
		FailedMFACount: 2,
		LastAnomalyAt:  &now,
	}
	config := utils.RiskScorerConfig{
		AnomalyWeight:   15,
		GeoJumpWeight:   20,
		FailedMFAWeight: 10,
	}
	scorer := profile.NewRiskScorer(config, nil)
	score, level := scorer.CalculateScore(indicators)

	assert.GreaterOrEqual(t, score, 75)
	assert.Contains(t, []string{"high", "critical"}, level)
}

func TestRiskScorer_LowRisk_NormalBehavior(t *testing.T) {
	indicators := profile.RiskIndicators{} // All zero
	config := utils.RiskScorerConfig{
		AnomalyWeight: 15,
	}
	scorer := profile.NewRiskScorer(config, nil)
	score, level := scorer.CalculateScore(indicators)

	assert.Less(t, score, 25)
	assert.Equal(t, "low", level)
}

func TestRiskScorer_DecayOverTime(t *testing.T) {
	oldAnomaly := time.Now().AddDate(0, 0, -60) // 60 days ago
	indicators := profile.RiskIndicators{
		AnomalyCount:  5,
		LastAnomalyAt: &oldAnomaly,
	}
	config := utils.RiskScorerConfig{
		AnomalyWeight: 15,
		DecayDays:     30,
		DecayRate:     0.5,
	}
	scorer := profile.NewRiskScorer(config, nil)
	score, _ := scorer.CalculateScore(indicators)

	// Base score: 5 * 15 = 75
	// Decay: 60 days / 30 days = 2 cycles
	// Factor: 0.5 ^ 2 = 0.25
	// Expected: 75 * 0.25 = 18.75 -> 18
	assert.Less(t, score, 40) // Should be significantly lower than 75
}

// === TagManager Tests ===

func TestTagManager_AutoTag_FrequentTraveler(t *testing.T) {
	userProfile := &profile.UserProfile{
		Behavior: profile.BehaviorMetrics{UniqueLocations: 8},
	}
	manager := profile.NewTagManager([]map[string]interface{}{
		{
			"tag": profile.TagFrequentTraveler,
			"condition": map[string]interface{}{
				"field":    "behavior.unique_locations",
				"operator": ">=",
				"value":    5,
			},
		},
	}, nil)

	tags := manager.EvaluateAutoTags(userProfile)
	assert.Contains(t, tags, profile.TagFrequentTraveler)
}

func TestTagManager_AutoTag_HighValueUser(t *testing.T) {
	userProfile := &profile.UserProfile{
		Behavior: profile.BehaviorMetrics{LoginFrequency: 15.0},
	}
	manager := profile.NewTagManager([]map[string]interface{}{
		{
			"tag": profile.TagHighValueUser,
			"condition": map[string]interface{}{
				"field":    "behavior.login_frequency",
				"operator": ">=",
				"value":    10.0,
			},
		},
	}, nil)

	tags := manager.EvaluateAutoTags(userProfile)
	assert.Contains(t, tags, profile.TagHighValueUser)
}

// MockProfileRepository for manual tag tests
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) Create(ctx context.Context, p *profile.UserProfile) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockProfileRepository) GetByUserID(ctx context.Context, userID string) (*profile.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.UserProfile), args.Error(1)
}
func (m *MockProfileRepository) Update(ctx context.Context, p *profile.UserProfile) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockProfileRepository) UpdateRisk(ctx context.Context, userID string, risk profile.RiskIndicators, score int, level string) error {
	return m.Called(ctx, userID, risk, score, level).Error(0)
}
func (m *MockProfileRepository) UpdateBehavior(ctx context.Context, userID string, behavior profile.BehaviorMetrics) error {
	return m.Called(ctx, userID, behavior).Error(0)
}
func (m *MockProfileRepository) UpdateTags(ctx context.Context, userID string, autoTags, manualTags profile.StringSlice) error {
	return m.Called(ctx, userID, autoTags, manualTags).Error(0)
}
func (m *MockProfileRepository) UpdateQuality(ctx context.Context, userID string, score int, details profile.QualityDetails) error {
	return m.Called(ctx, userID, score, details).Error(0)
}
func (m *MockProfileRepository) FindByRiskLevel(ctx context.Context, tenantID, level string, limit int) ([]*profile.UserProfile, error) {
	return nil, nil
}
func (m *MockProfileRepository) FindByTag(ctx context.Context, tenantID, tag string) ([]*profile.UserProfile, error) {
	return nil, nil
}
func (m *MockProfileRepository) Delete(ctx context.Context, userID string) error {
	return nil
}

func TestTagManager_ManualTag_CRUD(t *testing.T) {
	mockRepo := &MockProfileRepository{}
	userProfile := &profile.UserProfile{ManualTags: []string{}}

	mockRepo.On("GetByUserID", mock.Anything, "user1").Return(userProfile, nil)
	mockRepo.On("UpdateTags", mock.Anything, "user1", mock.Anything, mock.MatchedBy(func(tags profile.StringSlice) bool {
		return len(tags) == 1 && tags[0] == "vip"
	})).Return(nil)

	manager := profile.NewTagManager(nil, mockRepo)
	err := manager.AddManualTag(context.Background(), "user1", "vip")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// === QualityScorer Tests ===

func TestQualityScorer_Complete_Profile(t *testing.T) {
	details := profile.QualityDetails{
		HasEmail:         true,
		EmailVerified:    true,
		HasPhone:         true,
		PhoneVerified:    true,
		HasMFA:           true,
		HasRecoveryEmail: true,
		ProfileComplete:  1.0,
	}
	weights := utils.QualityWeightsConfig{
		Email:           15,
		EmailVerified:   10,
		Phone:           15,
		PhoneVerified:   10,
		MFA:             20,
		RecoveryEmail:   10,
		ProfileComplete: 20,
	}
	scorer := profile.NewQualityScorer(weights)
	score := scorer.CalculateScore(details)
	assert.Equal(t, 100, score)
}

func TestQualityScorer_Missing_Email_Penalty(t *testing.T) {
	details := profile.QualityDetails{
		HasEmail:      false,
		HasPhone:      true,
		PhoneVerified: true,
		HasMFA:        true,
	}
	weights := utils.QualityWeightsConfig{
		Email: 15,
		Phone: 15,
		MFA:   20,
	}
	scorer := profile.NewQualityScorer(weights)
	score := scorer.CalculateScore(details)

	// Max possible minus email related points
	// Here only counting explicit points: Phone(15)+PhoneVerified(default 0 in test setup if not set)+MFA(20) = 35
	// But let's check against what we expect from setup.
	// We didn't set all weights in struct literal, so zeros apply.
	// Let's set them all for clarity.
	weights = utils.QualityWeightsConfig{
		Email:           15,
		EmailVerified:   10,
		Phone:           15,
		PhoneVerified:   10,
		MFA:             20,
		RecoveryEmail:   10,
		ProfileComplete: 20,
	}
	scorer = profile.NewQualityScorer(weights)

	score = scorer.CalculateScore(details)

	// Expected: Phone(15) + PhoneVerified(10) + MFA(20) = 45
	assert.Equal(t, 45, score)
}
