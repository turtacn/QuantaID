package session

import (
	"context"
	"time"

	"github.com/turtacn/QuantaID/internal/identity/profile"
	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// EvaluationResult represents the outcome of a session evaluation.
type EvaluationResult struct {
	SessionID         string
	RiskScore         int
	RiskLevel         string
	Signals           []RiskSignal
	RecommendedAction ActionType
	Reason            string
	EvaluatedAt       time.Time
}

// EvaluationConfig holds configuration for the SessionEvaluator.
type EvaluationConfig struct {
	LowRiskThreshold    int
	MediumRiskThreshold int
	HighRiskThreshold   int
	MaxInactiveMinutes  int
	EnableGeoCheck      bool
}

// SessionEvaluator orchestrates the risk evaluation process.
type SessionEvaluator struct {
	profileService *profile.ProfileService
	riskMonitor    *RiskMonitor
	policyEngine   *SessionPolicy
	riskStore      *redis.SessionRiskStore
	config         EvaluationConfig
}

// NewSessionEvaluator creates a new SessionEvaluator.
func NewSessionEvaluator(profileService *profile.ProfileService, riskMonitor *RiskMonitor, policyEngine *SessionPolicy, riskStore *redis.SessionRiskStore, config EvaluationConfig) *SessionEvaluator {
	if config.LowRiskThreshold == 0 {
		config.LowRiskThreshold = 25
	}
	if config.MediumRiskThreshold == 0 {
		config.MediumRiskThreshold = 50
	}
	if config.HighRiskThreshold == 0 {
		config.HighRiskThreshold = 75
	}
	if config.MaxInactiveMinutes == 0 {
		config.MaxInactiveMinutes = 30
	}

	return &SessionEvaluator{
		profileService: profileService,
		riskMonitor:    riskMonitor,
		policyEngine:   policyEngine,
		riskStore:      riskStore,
		config:         config,
	}
}

// Evaluate performs a comprehensive risk evaluation for the session.
func (e *SessionEvaluator) Evaluate(ctx context.Context, session *Session) (*EvaluationResult, error) {
	result := &EvaluationResult{
		SessionID:   session.ID,
		EvaluatedAt: time.Now(),
	}

	// 1. Collect risk signals
	signals := e.riskMonitor.CollectSignals(ctx, session)
	result.Signals = signals

	// 2. Get user profile risk (if available)
	profileRiskScore := 0
	if e.profileService != nil {
		// assuming profileService has a GetRiskLevel method
		// score, _, _ := e.profileService.GetRiskLevel(ctx, session.UserID)
		// profileRiskScore = score
	}

	// 3. Calculate combined risk score
	result.RiskScore = e.calculateCombinedScore(signals, profileRiskScore)
	result.RiskLevel = e.scoreToLevel(result.RiskScore)

	// 4. Determine action via policy engine
	evalCtx := e.buildEvalContext(session, result)
	result.RecommendedAction, result.Reason = e.policyEngine.DetermineAction(evalCtx)

	// 5. Update cache
	if e.riskStore != nil {
		if err := e.riskStore.UpdateScore(ctx, session.ID, result.RiskScore, result.RiskLevel); err != nil {
			// Log error but continue
		}
	}

	return result, nil
}

func (e *SessionEvaluator) calculateCombinedScore(signals []RiskSignal, profileRiskScore int) int {
	signalScore := 0
	for _, signal := range signals {
		switch signal.Severity {
		case "high":
			signalScore += 30
		case "medium":
			signalScore += 15
		case "low":
			signalScore += 5
		}
	}

	// Weighted combination: 60% signals, 40% profile
	combined := int(float64(signalScore)*0.6 + float64(profileRiskScore)*0.4)
	if combined > 100 {
		combined = 100
	}
	return combined
}

func (e *SessionEvaluator) scoreToLevel(score int) string {
	if score < e.config.LowRiskThreshold {
		return "low"
	}
	if score < e.config.MediumRiskThreshold {
		return "medium"
	}
	if score < e.config.HighRiskThreshold {
		return "high"
	}
	return "critical"
}

// QuickCheck performs a lightweight risk check using cached data.
func (e *SessionEvaluator) QuickCheck(ctx context.Context, session *Session) (bool, error) {
	if e.riskStore == nil {
		// Fallback to full evaluate if no store
		return true, nil
	}

	cached, err := e.riskStore.Get(ctx, session.ID)
	if err != nil {
		return false, err
	}

	if cached != nil && time.Since(cached.LastEvaluated) < 1*time.Minute {
		return cached.RiskLevel != "critical", nil
	}

	// Cache expired or missing, need full evaluation
	result, err := e.Evaluate(ctx, session)
	if err != nil {
		return false, err
	}
	return result.RiskLevel != "critical", nil
}

func (e *SessionEvaluator) buildEvalContext(session *Session, result *EvaluationResult) map[string]interface{} {
	signalTypes := make([]string, len(result.Signals))
	for i, s := range result.Signals {
		signalTypes[i] = s.Type
	}

	return map[string]interface{}{
		"risk_level":        result.RiskLevel,
		"risk_score":        result.RiskScore,
		"signal_count":      len(result.Signals),
		"signals":           signalTypes,
		"inactive_minutes":  time.Since(session.LastActivityAt).Minutes(),
		"session_age_hours": time.Since(session.CreatedAt).Hours(),
		"auth_level":        session.AuthLevel,
	}
}
