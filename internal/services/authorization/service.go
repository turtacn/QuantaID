package authorization

import (
	"context"
	"github.com/turtacn/QuantaID/internal/services/audit"
	"github.com/turtacn/QuantaID/internal/domain/policy"
)

// Service provides a simplified interface for authorization checks,
// delegating the core logic to a configured policy evaluator.
type Service struct {
	evaluator    Evaluator
	auditService *audit.Service
}

// NewService creates a new authorization service.
func NewService(e Evaluator, auditService *audit.Service) *Service {
	return &Service{evaluator: e, auditService: auditService}
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
	decision, err := s.evaluator.Evaluate(ctx, evalCtx)
	if err != nil {
		return policy.DecisionDeny, err
	}

	// TODO: Extract IP and TraceID from context
	ip := "not_implemented"
	traceID := "not_implemented"

	details := map[string]any{
		"subject":     evalCtx.Subject,
		"action":      evalCtx.Action,
		"resource":    evalCtx.Resource,
		"environment": evalCtx.Environment,
	}

	s.auditService.RecordPolicyDecision(ctx, evalCtx.Subject.UserID, ip, evalCtx.Resource.ID, traceID, string(decision), details)

	return decision, nil
}
