package auth

import (
	"context"
	"github.com/turtacn/QuantaID/internal/services/audit"
	"github.com/turtacn/QuantaID/internal/domain/auth"
)

// RiskEngine evaluates the risk of a login attempt.
type RiskEngine interface {
	Assess(ctx context.Context, loginCtx auth.LoginContext) (*auth.RiskAssessment, error)
}

// SimpleRiskEngine provides a basic, rule-based implementation of the RiskEngine.
type SimpleRiskEngine struct {
	cfg          SimpleRiskConfig
	auditService *audit.Service
}

// SimpleRiskConfig holds the configuration for the SimpleRiskEngine.
type SimpleRiskConfig struct {
	NewDeviceScore   float64 `yaml:"new_device_score"`
	GeoVelocityScore float64 `yaml:"geo_velocity_score"`
	UnusualTimeScore float64 `yaml:"unusual_time_score"`
	BlockThreshold   float64 `yaml:"block_threshold"`
	MfaThreshold     float64 `yaml:"mfa_threshold"`
}

// NewSimpleRiskEngine creates a new SimpleRiskEngine.
func NewSimpleRiskEngine(cfg SimpleRiskConfig, auditService *audit.Service) *SimpleRiskEngine {
	return &SimpleRiskEngine{cfg: cfg, auditService: auditService}
}

// Assess evaluates the login context against a set of rules to determine a risk score.
func (e *SimpleRiskEngine) Assess(ctx context.Context, loginCtx auth.LoginContext) (*auth.RiskAssessment, error) {
	score := 0.0
	var factors []auth.RiskFactor

	if loginCtx.LastLoginIP != "" && loginCtx.LastLoginIP != loginCtx.CurrentIP {
		score += e.cfg.NewDeviceScore
		factors = append(factors, auth.RiskFactorNewDevice)
	}

	if loginCtx.LastLoginCountry != "" && loginCtx.LastLoginCountry != loginCtx.CurrentCountry {
		score += e.cfg.GeoVelocityScore
		factors = append(factors, auth.RiskFactorGeoVelocity)
	}

	hour := loginCtx.Now.Hour()
	if hour < 7 || hour > 22 {
		score += e.cfg.UnusualTimeScore
		factors = append(factors, auth.RiskFactorUnusualTime)
	}

	decision := auth.RiskDecisionAllow
	if score >= e.cfg.BlockThreshold {
		decision = auth.RiskDecisionDeny
	} else if score >= e.cfg.MfaThreshold {
		decision = auth.RiskDecisionRequireMFA
		// TODO: Extract TraceID from context
		traceID := "not_implemented"
		var factorStrs []string
		for _, f := range factors {
			factorStrs = append(factorStrs, string(f))
		}
		e.auditService.RecordHighRiskLogin(ctx, loginCtx.UserID, loginCtx.CurrentIP, traceID, score, factorStrs, nil)
	}

	return &auth.RiskAssessment{
		Score:    auth.RiskScore(score),
		Factors:  factors,
		Decision: decision,
	}, nil
}
