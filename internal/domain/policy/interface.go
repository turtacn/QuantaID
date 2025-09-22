package policy

import (
	"context"
	"github.com/turtacn/QuantaID/pkg/types"
)

// IService defines the interface for the policy domain service.
type IService interface {
	Evaluate(ctx context.Context, evalCtx *types.PolicyEvaluationContext) (*types.PolicyDecision, error)
	CreatePolicy(ctx context.Context, policy *types.Policy) error
}

//Personal.AI order the ending
