package authorization

import (
	"context"
	"github.com/turtacn/QuantaID/internal/domain/policy"
)

// Service provides a simplified interface for authorization checks,
// delegating the core logic to a configured policy evaluator.
type Service struct {
	evaluator Evaluator
}

// NewService creates a new authorization service.
//
// Parameters:
//   - e: The policy evaluator that will make the authorization decisions.
//
// Returns:
//   A new instance of the authorization service.
func NewService(e Evaluator) *Service {
	return &Service{evaluator: e}
}

// Authorize evaluates the given context and determines if the requested
// action is permitted. It acts as a thin wrapper around the evaluator.
//
// Parameters:
//   - ctx: The context for the request.
//   - evalCtx: The context containing all information for the evaluation.
//
// Returns:
//   A decision (Allow/Deny) and an error if the evaluation fails.
func (s *Service) Authorize(ctx context.Context, evalCtx policy.EvaluationContext) (policy.Decision, error) {
	return s.evaluator.Evaluate(ctx, evalCtx)
}
