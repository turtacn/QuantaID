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

// Evaluate performs the policy evaluation using a hybrid logic:
// 1. RBAC Check: Baseline permissions.
// 2. OPA Check: Can override RBAC deny (allow) or enforce explicit deny (deny).
// Logic: (RBAC_Allow || OPA_Allow) && !OPA_Deny
func (e *HybridEvaluator) Evaluate(ctx context.Context, req EvaluationRequest) (bool, error) {
	// 1. RBAC Check
	rbacAllowed, err := e.rbac.IsAllowed(ctx, req.SubjectID, req.Action, req.Resource)
	if err != nil {
		return false, err
	}

	// 2. OPA Check
	var opaAllowed, opaDenied bool
	if e.opa != nil {
		opaAllowed, opaDenied, err = e.opa.Evaluate(ctx, req)
		if err != nil {
			// Fail Close: If OPA fails, we deny access
			return false, fmt.Errorf("OPA evaluation failed: %w", err)
		}
	}

	// 3. Decision Logic

	// If OPA explicitly denies, then access is forbidden regardless of RBAC
	if opaDenied {
		return false, nil
	}

	// If OPA explicitly allows, then access is granted (overrides RBAC deny)
	if opaAllowed {
		return true, nil
	}

	// Otherwise, fall back to RBAC decision
	return rbacAllowed, nil
}

// RBACProvider is the interface for the RBAC component of the policy engine.
type RBACProvider interface {
	IsAllowed(ctx context.Context, subjectID, action, resource string) (bool, error)
}

// ABACProvider is the interface for the ABAC component of the policy engine.
type ABACProvider interface {
	Evaluate(ctx context.Context, requestContext map[string]interface{}) (bool, error)
}
