package authorization

import (
	"context"
	"github.com/turtacn/QuantaID/internal/policy/engine"
	"github.com/turtacn/QuantaID/internal/domain/policy"
)

// EvaluatorAdapter adapts the new policy engine Evaluator to the service layer Evaluator interface.
type EvaluatorAdapter struct {
	engineEvaluator engine.Evaluator
}

// NewEvaluatorAdapter creates a new adapter.
func NewEvaluatorAdapter(e engine.Evaluator) *EvaluatorAdapter {
	return &EvaluatorAdapter{engineEvaluator: e}
}

// Evaluate converts the service evaluation context to the engine request and delegates.
func (a *EvaluatorAdapter) Evaluate(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error) {
	req := engine.EvaluationRequest{
		SubjectID: evalCtx.Subject.UserID,
		Action:    string(evalCtx.Action),
		Resource:  evalCtx.Resource.ID,
		Context: map[string]interface{}{
			"resource_type": evalCtx.Resource.Type,
			"resource":      evalCtx.Resource,
			"environment":   evalCtx.Environment,
			"subject":       evalCtx.Subject,
			"roles":         evalCtx.Subject.Groups, // Mapping roles/groups to "roles" for OPA
		},
	}

	allowed, err := a.engineEvaluator.Evaluate(ctx, req)
	if err != nil {
		return policy.DecisionDeny, err
	}

	if allowed {
		return policy.DecisionAllow, nil
	}
	return policy.DecisionDeny, nil
}
