package authorization

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/auth"
	"github.com/turtacn/QuantaID/internal/domain/policy"
)

type PolicyEngine struct {
	evaluator Evaluator
}

func NewPolicyEngine(evaluator Evaluator) *PolicyEngine {
	return &PolicyEngine{evaluator: evaluator}
}

func (p *PolicyEngine) Decide(level auth.RiskLevel, ac auth.AuthContext) string {
	evalCtx := policy.EvaluationContext{
		Subject: policy.Subject{
			UserID: ac.UserID,
		},
		Action: "login",
		Environment: policy.Environment{
			IP: ac.IPAddress,
		},
	}
	decision, err := p.evaluator.Evaluate(context.Background(), evalCtx)
	if err != nil {
		return "DENY"
	}
	if decision == policy.DecisionAllow {
		if level == auth.RiskLevelHigh {
			return "REQUIRE_MFA"
		}
		return "ALLOW"
	}
	return "DENY"
}
