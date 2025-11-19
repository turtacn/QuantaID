package authorization

import (
	"context"

	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/pkg/types"
	"golang.org/x/exp/slices"
)

// Evaluator defines the interface for the policy evaluation engine.
type Evaluator interface {
	Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error)
}

// DefaultEvaluator is the default implementation of the Evaluator interface.
// It uses a policy repository to fetch and evaluate policies.
type DefaultEvaluator struct {
	repo policy.PolicyRepository
}

// NewDefaultEvaluator creates a new DefaultEvaluator.
func NewDefaultEvaluator(repo policy.PolicyRepository) *DefaultEvaluator {
	return &DefaultEvaluator{repo: repo}
}

// Evaluate checks the evaluation context against the policies in the repository.
func (e *DefaultEvaluator) Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error) {
	// For simplicity, we fetch all policies for the subject.
	// A more optimized approach might involve more specific queries.
	policies, err := e.repo.FindPoliciesForSubject(ctx, "user:"+evalCtx.Subject.UserID)
	if err != nil {
		return policy.DecisionDeny, err
	}

	// Deny overrides allow
	for _, p := range policies {
		if e.matches(p, evalCtx) {
			if p.Effect == types.EffectDeny {
				return policy.DecisionDeny, nil
			}
		}
	}

	for _, p := range policies {
		if e.matches(p, evalCtx) {
			if p.Effect == types.EffectAllow {
				return policy.DecisionAllow, nil
			}
		}
	}

	// Default deny if no rule matches
	return policy.DecisionDeny, nil
}

func (e *DefaultEvaluator) matches(p *types.Policy, evalCtx policy.EvaluationContext) bool {
	return e.matchesAction(p, evalCtx) && e.matchesSubject(p, evalCtx)
}

func (e *DefaultEvaluator) matchesAction(p *types.Policy, evalCtx policy.EvaluationContext) bool {
	if len(p.Actions) == 0 {
		return true
	}
	action := string(evalCtx.Action)
	return slices.Contains(p.Actions, "*") || slices.Contains(p.Actions, action)
}

func (e *DefaultEvaluator) matchesSubject(p *types.Policy, evalCtx policy.EvaluationContext) bool {
	if len(p.Subjects) == 0 {
		return true
	}

	if slices.Contains(p.Subjects, "*") {
		return true
	}

	for _, subject := range p.Subjects {
		if subject == "user:"+evalCtx.Subject.UserID {
			return true
		}
		for _, group := range evalCtx.Subject.Groups {
			if subject == "group:"+group {
				return true
			}
		}
	}
	return false
}

