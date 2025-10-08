package policy

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

// Service encapsulates the business logic for managing and evaluating authorization policies.
// It provides the core functionality for making access control decisions.
type Service struct {
	repo   Repository
	logger utils.Logger
}

// NewService creates a new policy service instance.
//
// Parameters:
//   - repo: The repository for accessing policy data.
//   - logger: The logger for service-level messages.
//
// Returns:
//   A new policy service instance.
func NewService(repo Repository, logger utils.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Evaluate determines if a subject has permission to perform an action on a resource
// based on the configured policies. It follows a "deny overrides" model.
//
// Parameters:
//   - ctx: The context for the request.
//   - evalCtx: The context containing the subject, action, resource, and environmental data for evaluation.
//
// Returns:
//   A PolicyDecision object indicating whether the action is allowed, along with the reasoning.
//   Returns an error if policy retrieval fails.
func (s *Service) Evaluate(ctx context.Context, evalCtx *types.PolicyEvaluationContext) (*types.PolicyDecision, error) {
	subjects := s.getSubjectsFromContext(evalCtx)

	var relevantPolicies []*types.Policy
	for _, subject := range subjects {
		policies, err := s.repo.FindPoliciesForSubject(ctx, subject)
		if err != nil {
			s.logger.Error(ctx, "Failed to find policies for subject", zap.Error(err), zap.String("subject", subject))
			return nil, types.ErrInternal.WithCause(err)
		}
		relevantPolicies = append(relevantPolicies, policies...)
	}

	decision := &types.PolicyDecision{
		Allowed:         false,
		MatchingPolicies: []string{},
	}

	isAllowed := false
	for _, policy := range relevantPolicies {
		if s.policyMatches(policy, evalCtx) {
			if policy.Effect == types.EffectDeny {
				decision.Allowed = false
				decision.Reason = "Explicitly denied by policy " + policy.ID
				decision.MatchingPolicies = append(decision.MatchingPolicies, policy.ID)
				return decision, nil
			}
			if policy.Effect == types.EffectAllow {
				isAllowed = true
				decision.MatchingPolicies = append(decision.MatchingPolicies, policy.ID)
			}
		}
	}

	decision.Allowed = isAllowed
	if !isAllowed {
		decision.Reason = "No allowing policy found"
	}

	return decision, nil
}

// policyMatches checks if a given policy applies to the evaluation context.
// Currently, it checks for matching actions and resources.
// TODO: Implement condition evaluation for full ABAC support.
func (s *Service) policyMatches(policy *types.Policy, evalCtx *types.PolicyEvaluationContext) bool {
	if !s.matchesAction(policy.Actions, evalCtx.Action) {
		return false
	}
	if !s.matchesResource(policy.Resources, evalCtx.Resource["id"].(string)) {
		return false
	}
	return true
}

// matchesAction checks if the requested action is covered by the policy's actions.
// It supports wildcards ('*').
func (s *Service) matchesAction(policyActions []string, requestAction string) bool {
	return slices.Contains(policyActions, requestAction) || slices.Contains(policyActions, "*")
}

// matchesResource checks if the requested resource is covered by the policy's resources.
// It supports wildcards ('*').
func (s *Service) matchesResource(policyResources []string, requestResource string) bool {
	return slices.Contains(policyResources, requestResource) || slices.Contains(policyResources, "*")
}

// getSubjectsFromContext extracts all relevant subject identifiers from the evaluation context.
// This includes the user's ID and the IDs of all groups they belong to.
func (s *Service) getSubjectsFromContext(evalCtx *types.PolicyEvaluationContext) []string {
	var subjects []string
	if userID, ok := evalCtx.Subject["id"].(string); ok {
		subjects = append(subjects, "user:"+userID)
	}
	if groups, ok := evalCtx.Subject["groups"].([]string); ok {
		for _, groupID := range groups {
			subjects = append(subjects, "group:"+groupID)
		}
	}
	return subjects
}

// CreatePolicy provides a straightforward way to add a new policy via the service layer.
//
// Parameters:
//   - ctx: The context for the request.
//   - policy: The policy object to be created.
//
// Returns:
//   An error if the creation fails.
func (s *Service) CreatePolicy(ctx context.Context, policy *types.Policy) error {
	return s.repo.CreatePolicy(ctx, policy)
}
