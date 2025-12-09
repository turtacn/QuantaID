package profile

import (
	"context"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Publish(event interface{}) error
}

// ProfileService orchestrates profile operations
type ProfileService struct {
	profileRepo    ProfileRepository
	profileBuilder *ProfileBuilder
	riskScorer     *RiskScorer
	tagManager     *TagManager
	qualityScorer  *QualityScorer
	eventBus       EventPublisher
}

// NewProfileService creates a new ProfileService
func NewProfileService(
	profileRepo ProfileRepository,
	profileBuilder *ProfileBuilder,
	riskScorer *RiskScorer,
	tagManager *TagManager,
	qualityScorer *QualityScorer,
	eventBus EventPublisher,
) *ProfileService {
	return &ProfileService{
		profileRepo:    profileRepo,
		profileBuilder: profileBuilder,
		riskScorer:     riskScorer,
		tagManager:     tagManager,
		qualityScorer:  qualityScorer,
		eventBus:       eventBus,
	}
}

// GetProfile retrieves or builds a user profile
func (s *ProfileService) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return s.profileBuilder.BuildOrUpdate(ctx, userID)
	}
	return profile, nil
}

// GetRiskLevel retrieves the risk score and level for a user
func (s *ProfileService) GetRiskLevel(ctx context.Context, userID string) (int, string, error) {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return 0, "", err
	}
	return profile.RiskScore, profile.RiskLevel, nil
}

// RefreshProfile forces a rebuild and re-evaluation of the profile
func (s *ProfileService) RefreshProfile(ctx context.Context, userID string) (*UserProfile, error) {
	profile, err := s.profileBuilder.BuildOrUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Re-calculate Risk
	score, level := s.riskScorer.CalculateScore(profile.Risk)
	profile.RiskScore = score
	profile.RiskLevel = level
	if err := s.profileRepo.UpdateRisk(ctx, userID, profile.Risk, score, level); err != nil {
		return nil, err
	}

	// Re-evaluate Tags
	profile.AutoTags = s.tagManager.EvaluateAutoTags(profile)
	if err := s.profileRepo.UpdateTags(ctx, userID, profile.AutoTags, profile.ManualTags); err != nil {
		return nil, err
	}

	// Re-calculate Quality
	profile.QualityScore = s.qualityScorer.CalculateScore(profile.QualityDetails)
	if err := s.profileRepo.UpdateQuality(ctx, userID, profile.QualityScore, profile.QualityDetails); err != nil {
		return nil, err
	}

	return profile, nil
}

// HandleAnomalyEvent handles incoming anomaly events
func (s *ProfileService) HandleAnomalyEvent(ctx context.Context, userID string, event AnomalyEvent) error {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return err
	}

	profile.Risk = s.riskScorer.UpdateFromEvent(ctx, profile, event)
	score, level := s.riskScorer.CalculateScore(profile.Risk)

	if err := s.profileRepo.UpdateRisk(ctx, userID, profile.Risk, score, level); err != nil {
		return err
	}

	// Trigger alert if critical
	if level == "critical" && s.eventBus != nil {
		// s.eventBus.Publish(CriticalRiskEvent{UserID: userID, Score: score})
	}

	return nil
}

// GetQualityScore retrieves quality score and improvement suggestions
func (s *ProfileService) GetQualityScore(ctx context.Context, userID string) (int, []string, error) {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return 0, nil, err
	}
	suggestions := s.qualityScorer.GetImprovementSuggestions(profile.QualityDetails)
	return profile.QualityScore, suggestions, nil
}

// AddTag adds a manual tag
func (s *ProfileService) AddTag(ctx context.Context, userID, tag string) error {
	return s.tagManager.AddManualTag(ctx, userID, tag)
}

// RemoveTag removes a manual tag
func (s *ProfileService) RemoveTag(ctx context.Context, userID, tag string) error {
	return s.tagManager.RemoveManualTag(ctx, userID, tag)
}

// GetUserTags retrieves all tags for a user
func (s *ProfileService) GetUserTags(ctx context.Context, userID string) ([]string, []string, error) {
	profile, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return profile.AutoTags, profile.ManualTags, nil
}
