package policy

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

// Service encapsulates the business logic for policy evaluation.
type Service struct {
	repo   Repository
	logger utils.Logger
}

// NewService creates a new policy service.
func NewService(repo Repository, logger utils.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Evaluate determines if a subject has permission to perform an action on a resource.
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

func (s *Service) policyMatches(policy *types.Policy, evalCtx *types.PolicyEvaluationContext) bool {
	if !s.matchesAction(policy.Actions, evalCtx.Action) {
		return false
	}
	if !s.matchesResource(policy.Resources, evalCtx.Resource["id"].(string)) {
		return false
	}
	return true
}

func (s *Service) matchesAction(policyActions []string, requestAction string) bool {
	return slices.Contains(policyActions, requestAction) || slices.Contains(policyActions, "*")
}

func (s *Service) matchesResource(policyResources []string, requestResource string) bool {
	return slices.Contains(policyResources, requestResource) || slices.Contains(policyResources, "*")
}

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

func (s *Service) CreatePolicy(ctx context.Context, policy *types.Policy) error {
	return s.repo.CreatePolicy(ctx, policy)
}

//Personal.AI order the ending
