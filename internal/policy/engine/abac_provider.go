package engine

import (
	"context"
	"fmt"
	"time"
)

// SimpleABACProvider is a basic implementation of the ABACProvider interface.
type SimpleABACProvider struct {
	// In a real-world scenario, this would be a more complex rule engine.
}

// NewSimpleABACProvider creates a new SimpleABACProvider.
func NewSimpleABACProvider() *SimpleABACProvider {
	return &SimpleABACProvider{}
}

// Evaluate evaluates the context against a set of rules.
// This is a placeholder for a real rule engine.
func (p *SimpleABACProvider) Evaluate(ctx context.Context, requestContext map[string]interface{}) (bool, error) {
	if rule, ok := requestContext["rule"]; ok {
		switch rule {
		case "resource.owner_id == subject.id":
			return p.checkOwnership(requestContext)
		case "time.hour >= 9 && time.hour <= 18":
			return p.checkWorkingHours(requestContext)
		// Add more rule cases here
		default:
			return false, fmt.Errorf("unknown rule: %s", rule)
		}
	}

	// If no specific rule is present, default to allowing the action
	// as RBAC has already cleared it.
	return true, nil
}

// checkOwnership is a simple example of an ABAC rule.
func (p *SimpleABACProvider) checkOwnership(requestContext map[string]interface{}) (bool, error) {
	resourceOwnerID, okOwner := requestContext["resource.owner_id"].(string)
	subjectID, okSubject := requestContext["subject.id"].(string)

	if !okOwner || !okSubject {
		// If the necessary attributes are not present, deny access.
		return false, nil
	}

	return resourceOwnerID == subjectID, nil
}

func (p *SimpleABACProvider) checkWorkingHours(requestContext map[string]interface{}) (bool, error) {
	hour := time.Now().Hour()
	return hour >= 9 && hour <= 18, nil
}
