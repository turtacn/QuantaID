package adaptive

import (
	"github.com/turtacn/QuantaID/internal/domain/auth"
)

// PolicyDecision represents the action to be taken based on a risk assessment.
type PolicyDecision string

const (
	// PolicyDecisionAllow allows the authentication attempt to proceed without further checks.
	PolicyDecisionAllow PolicyDecision = "ALLOW"
	// PolicyDecisionRequireMFA requires the user to complete a Multi-Factor Authentication challenge.
	PolicyDecisionRequireMFA PolicyDecision = "REQUIRE_MFA"
	// PolicyDecisionDeny denies the authentication attempt outright.
	PolicyDecisionDeny PolicyDecision = "DENY"
)

// PolicyEngine determines the appropriate action based on the risk level.
type PolicyEngine struct {
	// In a real implementation, this would hold tenant-specific policies.
}

// NewPolicyEngine creates a new policy engine.
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{}
}

// Decide makes a policy decision based on the provided risk level.
func (p *PolicyEngine) Decide(level auth.RiskLevel, ac auth.AuthContext) PolicyDecision {
	switch level {
	case auth.RiskLevelLow:
		return PolicyDecisionAllow
	case auth.RiskLevelMedium:
		return PolicyDecisionRequireMFA
	case auth.RiskLevelHigh:
		// Depending on policy, this could be REQUIRE_STRONG_MFA or DENY.
		// For now, we'll map it to REQUIRE_MFA.
		return PolicyDecisionRequireMFA
	default:
		return PolicyDecisionDeny
	}
}
