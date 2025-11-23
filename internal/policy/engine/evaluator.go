package engine

import (
	"context"
	"fmt"
)

// EvaluationRequest represents the input for a policy evaluation.
type EvaluationRequest struct {
	SubjectID  string
	Action     string
	Resource   string
	Context    map[string]interface{}
}

// Evaluator is the interface for the policy engine.
type Evaluator interface {
	Evaluate(ctx context.Context, req EvaluationRequest) (bool, error)
}

// HybridEvaluator implements the Evaluator interface with a hybrid RBAC/ABAC/OPA model.
type HybridEvaluator struct {
	rbac RBACProvider
	abac ABACProvider
	opa  *OPAProvider
}

// NewHybridEvaluator creates a new HybridEvaluator.
func NewHybridEvaluator(rbac RBACProvider, abac ABACProvider, opa *OPAProvider) *HybridEvaluator {
	return &HybridEvaluator{
		rbac: rbac,
		abac: abac,
		opa:  opa,
	}
}

// Evaluate performs the policy evaluation.
// 1. It first checks for a definitive "allow" from the RBAC provider.
// 2. If RBAC allows, it then checks for any "deny" rules from the ABAC provider.
// 3. Finally, if OPA is configured, it checks against OPA policies.
func (e *HybridEvaluator) Evaluate(ctx context.Context, req EvaluationRequest) (bool, error) {
	// RBAC check (fast path)
	allowedByRBAC, err := e.rbac.IsAllowed(ctx, req.SubjectID, req.Action, req.Resource)
	if err != nil {
		return false, err
	}
	if !allowedByRBAC {
		return false, nil // Denied by RBAC
	}

	// ABAC check (slower path, only if RBAC allows)
	// Here you could load specific ABAC policies related to the resource/action
	// For now, we'll just pass the context to a generic ABAC provider
	abacDecision, err := e.abac.Evaluate(ctx, req.Context)
	if err != nil {
		return false, err
	}
	if !abacDecision {
		return false, nil // Denied by ABAC
	}

	// OPA check (Fine-grained external policy)
	if e.opa != nil {
		opaDecision, err := e.opa.Evaluate(ctx, req)
		if err != nil {
			// Fail Close: If OPA fails, we deny access
			return false, fmt.Errorf("OPA evaluation failed: %w", err)
		}
		if !opaDecision {
			return false, nil // Denied by OPA
		}
	}

	return true, nil
}

// RBACProvider is the interface for the RBAC component of the policy engine.
type RBACProvider interface {
	IsAllowed(ctx context.Context, subjectID, action, resource string) (bool, error)
}

// ABACProvider is the interface for the ABAC component of the policy engine.
type ABACProvider interface {
	Evaluate(ctx context.Context, requestContext map[string]interface{}) (bool, error)
}
